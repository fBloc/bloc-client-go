# bloc-client-go
[bloc](https://github.com/fBloc/bloc)的python client SDK

可以基于此SDK使用开发基于go语言的 `bloc function`

在继续之前，请确保你已经大概了解了[bloc](https://github.com/fBloc/bloc)是什么，并且已经通过[教程](https://github.com/fBloc/bloc/blob/main/docs/guide/deploy_local_environment_guide.md)在本地环境部署了bloc环境

## bloc function是什么？
bloc function就是开发者通过对应语言的bloc client SDK开发的一个函数，其在部署后便能够被用户在bloc前端看到，并可被编入工作流进行组合和自定义设置参数。待到对应工作流执行时到此函数时，就会按照用户在前端配置的参数进行执行并上报运行信息和运行结果到bloc。

## 开发bloc function教程
这里我们通过编写一个简单的 bloc function来进行说明。

其接收多个整数，并且接收一个特定的算法。然后对那些数字执行对应的算法

### 准备
- 创建目录
    ```shell
    mkdir bloc_go_tryout && cd bloc_go_tryout
    ```
- 初始化
    ```shell
	go mod init bloc_go_tryout
	```
- 安装依赖
    ```shell
	go get github.com/fBloc/bloc-client-go
	```
- 创建一个用于放bloc function的目录
	```shell
	mkdir bloc_node
	```

### 编写bloc function
1. 首先创建一个结构体、其表示了这个bloc function节点
    ```go
	// bloc_node/math_calcu.go
	type MathCalcu struct {
	}
	```
    而后此bloc function 需要实现 [BlocFunctionNodeInterface](https://github.com/fBloc/bloc-client-go/blob/main/function_interface.go#L10)

2. 实现 IptConfig() 方法 - 其描述了此function需要的入参:
   ```go
	func (*MathCalcu) IptConfig() bloc_client.Ipts {
		return bloc_client.Ipts{
			{
				Key:     "numbers",
				Display: "int numbers",
				Must:    true, // this ipt must be set
				Components: []*bloc_client.IptComponent{
					{
						ValueType:       bloc_client.IntValueType,     // input value should be int type
						FormControlType: bloc_client.InputFormControl, // frontend should use input
						Hint:            "input integer numbers",      // hint for user
						AllowMulti:      true,                         // multiple input is allowed
					},
				},
			},
			{
				Key:     "arithmetic_operator",
				Display: "choose arithmetic operators",
				Must:    true,
				Components: []*bloc_client.IptComponent{
					{
						ValueType:       bloc_client.IntValueType,
						FormControlType: bloc_client.SelectFormControl, // frontend should use select
						Hint:            "+/-/*/%",
						SelectOptions: []bloc_client.SelectOption{ // select options
							{Label: "addition", Value: 1},
							{Label: "subtraction", Value: 2},
							{Label: "multiplication", Value: 3},
							{Label: "division", Value: 4},
						},
						AllowMulti: false,  // only allow single select value
					},
				},
			},
		}
	}
	```

3. 实现 OptConfig() - 其描述了此function需要的出参:
    ```go
	func (*MathCalcu) OptConfig() bloc_client.Opts {
		// bloc_client.Opts: array type for a fixed order to show in the frontend which lead to a better user experience
		return bloc_client.Opts{
			{
				Key:         "result",
				Description: "arithmetic operation result",
				ValueType:   bloc_client.IntValueType,
				IsArray:     false,
			},
		}
	}
	```

4.  实现 AllProgressMilestones() - 其定义了此函数在运行进度过程中的高可读的里程碑事件

    此设计主要是为运行时间较长的函数准备的。当其正在运行时，其可以上报自己当前运行到的进度里程碑。从而关心进度的人能够在bloc前端看到此信息。如果你的bloc function很快就能运行完成，那就没必要设置这个了。

    ```go
	type progress int

	const (
		parsingParam progress = iota
		inCalculation
		finish
		maxProgress
	)

	func (p progress) String() string {
		switch p {
		case parsingParam:
			return "parsing ipt"
		case inCalculation:
			return "in calculation"
		case finish:
			return "finished"
		}
		return "unknown"
	}

	func (p progress) MilestoneIndex() *int {
		tmp := int(p)
		return &tmp
	}

	func (*MathCalcu) AllProgressMilestones() []string {
		tmp := make([]string, 0, maxProgress-1)
		for i := 0; i < int(maxProgress); i++ {
			tmp = append(tmp, progress(i).String())
		}
		return tmp
	}
	```


5. 实现 Run() - 真正的执行逻辑:
    ```go
	// Run do the real work
	func (*MathCalcu) Run(
		ctx context.Context,
		ipts bloc_client.Ipts,
		progressReportChan chan bloc_client.HighReadableFunctionRunProgress,
		blocOptChan chan *bloc_client.FunctionRunOpt,
		logger *bloc_client.Logger,
	) {
		// logger msg will be reported to bloc-server and can be represent in the frontend
		// which means during this function's running, the frontend can get the realtime log msg
		logger.Infof("start")

		progressReportChan <- bloc_client.HighReadableFunctionRunProgress{
			ProgressMilestoneIndex: parsingParam.MilestoneIndex(), // AllProgressMilestones() index 0 - "parsing ipt". which will be represented in the frontend immediately.
		}

		numbersSlice, err := ipts.GetIntSliceValue(0, 0)
		if err != nil {
			blocOptChan <- &bloc_client.FunctionRunOpt{
				Suc:                       false,                        // function run failed
				InterceptBelowFunctionRun: true,                         // intercept flow's below function run (you can think like raise panic in the flow)
				ErrorMsg:                  "parse ipt `numbers` failed", // error description
			}
			// Suc can be false and InterceptBelowFunctionRun can also be false
			// which means this function node's fail should not intercept it's below function node's running
			return
		}
		if len(numbersSlice) <= 0 {
			blocOptChan <- &bloc_client.FunctionRunOpt{
				Suc:                       true,
				InterceptBelowFunctionRun: false,
				Description:               "get no number to do calculation",
				Detail: map[string]interface{}{ // detail should match the OptConfig()
					"result": 0,
				},
			}
			return
		}

		operator, err := ipts.GetIntValue(1, 0)
		if err != nil {
			blocOptChan <- &bloc_client.FunctionRunOpt{
				Suc:                       false,
				InterceptBelowFunctionRun: true,
				ErrorMsg:                  "parse ipt `arithmetic_operator` failed",
			}
			return
		}

		progressReportChan <- bloc_client.HighReadableFunctionRunProgress{
			ProgressMilestoneIndex: inCalculation.MilestoneIndex(), // AllProgressMilestones() index 1 - "in calculation". which also will be represented in the frontend immediately.
		}

		ret := 0
		switch operator {
		case 1:
			for _, i := range numbersSlice {
				ret += i
			}
		case 2:
			for _, i := range numbersSlice {
				ret -= i
			}
		case 3:
			ret = numbersSlice[0]
			for _, i := range numbersSlice[1:] {
				ret *= i
			}
		case 4:
			ret = numbersSlice[0]
			for _, i := range numbersSlice[1:] {
				if i == 0 {
					blocOptChan <- &bloc_client.FunctionRunOpt{
						Suc:                       false,
						InterceptBelowFunctionRun: true,
						ErrorMsg:                  "division not allowed zero as denominator",
					}
					return
				}
				ret /= i
			}
		default:
			blocOptChan <- &bloc_client.FunctionRunOpt{
				Suc:                       false,
				InterceptBelowFunctionRun: true,
				ErrorMsg:                  "not valid arithmetic_operator",
			}
			return
		}
		progressReportChan <- bloc_client.HighReadableFunctionRunProgress{
			ProgressMilestoneIndex: finish.MilestoneIndex()}

		blocOptChan <- &bloc_client.FunctionRunOpt{
			Suc:                       true,
			InterceptBelowFunctionRun: false,
			Detail:                    map[string]interface{}{"result": ret},
			Description:               fmt.Sprintf("received %d number", len(numbersSlice)),
		}
	}
	```
- 到此，我们完成了bloc function的核心定义代码。接下来我们它写测试代码

### 编写单元测试代码
- 在`bloc_node/math_calcu_test.go`中编写简单的测试:
    ```go
	import (
		"testing"

		bloc_client "github.com/fBloc/bloc-client-go"
	)

	func TestMathCalcu(t *testing.T) {
		blocClient := bloc_client.NewTestClient()
		funcRunOpt := blocClient.TestRunFunction(
			&MathCalcu{},
			[][]interface{}{
				{ // ipt 1 group, numbers
					[]interface{}{1, 2},
				},
				{ // ipt 2 group, arithmetic operator
					1, // "+" operater
				},
			},
		)
		if !funcRunOpt.Suc {
			t.Errorf("TestMathCalcu failed wit error msg: %s", funcRunOpt.ErrorMsg)
		}
		if funcRunOpt.Detail["result"] != 3 {
			t.Errorf("TestMathCalcu failed, detail: %v", funcRunOpt.Detail)
		}
	}
	```
- 在 `bloc_node` 目录下, 执行命令 `go test -v .`, 你可以看到 `PASS`. 表示你开发的函数的运行满足你的预期
	```shell
	$ go test -v -count=1  .
	=== RUN   TestMathCalcu
	2022/05/20 12:11:56 reporting progress: {0  0x140000252e0}
	2022/05/20 12:11:56 reporting progress: {0  0x140000252e8}
	2022/05/20 12:11:56 reporting progress: {0  0x14000025338}
	2022/05/20 12:11:56 run finished with resp: &{Suc:true Canceled:false TimeoutCanceled:false InterceptBelowFunctionRun:false ErrorMsg: Description:received 2 number Detail:map[result:3] KeyMapObjectStorageKey:map[] Brief:map[]}
	--- PASS: TestMathCalcu (0.00s)
	PASS
	ok      github.com/fBloc/bloc-client-go/examples/bloc_go_tryout/bloc_node       0.109s
	```


### 向 bloc-server 汇报
当通过编写足够的测试确认你的函数运行没有问题了之后，你可以通过部署它来向bloc-server进行汇报

这里我假设你已经通过[教程](https://github.com/fBloc/bloc/blob/main/docs/guide/deploy_local_environment_guide.md)部署了一套本地bloc环境。如果是的话、请继续。

- 在 `bloc_go_tryout` 目录创建一个`main.go`文件，并写入内容:
    ```go
    package main

	import (
		bloc_client "github.com/fBloc/bloc-client-go"
		"github.com/fBloc/bloc-client-go/examples/bloc_go_tryout/bloc_node"
	)

	const appName = "tryout"

	func main() {
		client := bloc_client.NewClient(appName)

		// config
		client.GetConfigBuilder().SetRabbitConfig(
			"blocRabbit", "blocRabbitPasswd", []string{"127.0.0.1:5672"}, "", // bloc rabbitMQ address
		).SetServer(
			"127.0.0.1", 8080, // bloc-backend-server address
		).BuildUp()

		// register your functions
		sourceFunctionGroup := client.RegisterFunctionGroup("math") // give your function a group name
		sourceFunctionGroup.AddFunction(
			"calcu", // name your function node's name
			"receive numbers and do certain math operation to them", // the describe of your function node
			&bloc_node.MathCalcu{}, // your function implement
		)

		client.Run()
	}
    ```
- 运行:
    ```shell
    go run main.go
    ```

- 运行成功后，此bloc function就成功提交到bloc了。从而你可以在bloc web端看到与使用此function了（如果你刚刚接触bloc web端、不清楚有哪些操作，可见[教程](https://docs.blocapp.xyz/docs/category/web%E7%AB%AF%E5%8A%9F%E8%83%BD%E7%AE%80%E4%BB%8B)）

### 完整代码
[这里](https://github.com/fBloc/bloc-client-go/tree/main/examples/bloc_go_tryout)

## 其他资料
- 你可以在[这里](https://github.com/fBloc/bloc-function-demo-go) 找到更多更复杂的bloc function例子