# bloc-client-go
The go language client SDK for [bloc](https://github.com/fBloc/bloc).

You can develop bloc's function node in go language based on this SDK.

First make sure you already have a knowledge of [bloc](https://github.com/fBloc/bloc) and already have deployed a local test bloc environment(see [deploy a local bloc environment](https://github.com/fBloc/bloc/blob/main/docs/guide/deploy_local_environment_guide.md)).

## Develop bloc function node tutorial
Let's write a simple bloc function node which receive some integers and do a designated mathematical calculation to these integers.

### prepare
- create a directory for code
	```shell
	mkdir bloc_go_tryout && cd bloc_go_tryout
	```
- init 
	```shell
	go mod init bloc_go_tryout
	```
- install package
	```shell
	go get github.com/fBloc/bloc-client-go
	```
- create a folder to hold bloc_function_nodes
	```shell
	mkdir bloc_node
	```

### write bloc function node
1. create a struct which stand for the function node:
	```go
	// math_calcu.go
	type MathCalcu struct {
	}
	```

	then we are going to implement the [BlocFunctionNodeInterface](https://github.com/fBloc/bloc-client-go/blob/main/function_interface.go#L10).

2. implement IptConfig() which defined function node's input params:
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

3. implement OptConfig() which defined function node's opt:
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

4. implement AllProgressMilestones() which define the highly readable describe milestones of the function node's run:
	AllProgressMilestones() is designed for long run function, during it is running, it can report it's current running milestone for the user in frontend to get the information. If your function is quick run, maybe no need to set it and just return blank string array.
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

4. implement Run() which do the real work:
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
- By now, we finished write the code of a bloc function node, next we will write unit test code to this function node

### write unit test
- write a simple unit test in `math_calcu_test.go`:
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

- under `bloc_node` directory, run command `go test -v .`, you will see the PASS. which means your function run meet your expectation.
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


### report to the server
After make sure your function runs well, you can deploy it to report to bloc.

I suppose you have already deployed a local bloc environment. if not, follow [guide](https://github.com/fBloc/bloc/blob/main/docs/guide/deploy_local_environment_guide.md) to deploy it.

- During `bloc_go_tryout` directory and make a `main.go` file with content:
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
- now run it:
	```shell
	go run main.go
	```
- After suc run, this client's all function node are registered to the bloc-server, which can be see and operate in the frontend, and this client will receive bloc-server's trigger function to run msg and do the execute. If you are first to the bloc web, you may check this [brief introduction to bloc web](#todo)

### total code 
you can find the demo code [here](https://github.com/fBloc/bloc-client-go/tree/main/examples/bloc_go_tryout).

## Other references
- you can find more bloc function node examples [here](https://github.com/fBloc/bloc-function-demo-go)
