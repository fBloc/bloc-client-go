# bloc-client-go
The go language client SDK for [bloc](https://github.com/fBloc/bloc).

You can develop bloc's function node in go language based on this SDK.

First make sure you already have a knowledge of [bloc](https://github.com/fBloc/bloc) and already have deployed a bloc-server instance.

## How to use
Let's write a simple function which receive some integers and do a designated mathematical calculation to these integers.

### prepare
create a go program package and initial it:
```shell
# first to your go path
$ mkdir bloc_go_tryout
$ cd bloc_go_tryout
$ go mod init bloc_go_tryout
```

get sdk:
```shell
$ go get github.com/fBloc/bloc-client-go 
```

create a folder to hold function:
```shell
# first to your go path
$ mkdir functions
```

### write the function
1. first create a struct which stand for the function node:
```go
type MathCalcu struct {
}
```

then the function node should implement the [interface](https://github.com/fBloc/bloc-client-go/blob/5b8cfb723afd6d7c81a57fbd45ab35a2ded52f05/function_interface.go#L7).

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
					ValueType:       bloc_client.IntValueType,     // input should be int type
					FormControlType: bloc_client.InputFormControl, // frontend should use input
					Hint:            "input integer numbers",      // hint for user
					AllowMulti:      true,                         // allow input multi
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
					FormControlType: bloc_client.SelectFormControl,
					Hint:            "+/-/*/%",
					SelectOptions: []bloc_client.SelectOption{
						{Label: "addition", Value: 1},
						{Label: "subtraction", Value: 2},
						{Label: "multiplication", Value: 3},
						{Label: "division", Value: 4},
					},
					AllowMulti: false,
				},
			},
		},
	}
}
```

3. implement OptConfig() which defined function node's opt:
```go
func (*MathCalcu) OptConfig() bloc_client.Opts {
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

This is designed 4 long run function, during it is running, it can report it's current running milestone for the user in frontend to get the information.

If your function is quick run. maybe no need to set it and just return blank.

```go
func (*MathCalcu) AllProgressMilestones() []string {
	return []string{
        "finished parsing ipt",
        "start do the calculation",
        "finished do the calculation",
    }
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
	// logger.Infof("start")

	progressReportChan <- bloc_client.HighReadableFunctionRunProgress{
		ProgressMilestoneIndex: 0, // AllProgressMilestones() index 0 - "start parsing ipt". which also will be represented in the frontend immediately.
	}

	numbersSlice, err := ipts.GetIntSliceValue(0, 0)
	if err != nil {
		blocOptChan <- &bloc_client.FunctionRunOpt{
			Suc:         false,
			Description: "parse ipt `numbers` failed",
		}
		return
	}
	if len(numbersSlice) <= 0 {
		blocOptChan <- &bloc_client.FunctionRunOpt{
			Suc: true,
			Detail: map[string]interface{}{ // detail should match the OptConfig()
				"result": 0,
			},
		}
		return
	}

	operator, err := ipts.GetIntValue(1, 0)
	if err != nil {
		blocOptChan <- &bloc_client.FunctionRunOpt{
			Suc:         false,
			Description: "parse ipt `arithmetic_operator` failed",
		}
		return
	}

	progressReportChan <- bloc_client.HighReadableFunctionRunProgress{ProgressMilestoneIndex: 1}

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
					Suc:         false,
					Description: "division not allowed zero as denominator",
				}
				return
			}
			ret /= i
		}
	default:
		blocOptChan <- &bloc_client.FunctionRunOpt{
			Suc:         false,
			Description: "not valid arithmetic_operator",
		}
		return
	}
	progressReportChan <- bloc_client.HighReadableFunctionRunProgress{ProgressMilestoneIndex: 2}

	blocOptChan <- &bloc_client.FunctionRunOpt{
		Suc:         true,
		Detail:      map[string]interface{}{"result": ret},
		Description: fmt.Sprintf("received %d number", len(numbersSlice)),
	}
}
```

ok, we finished write the function

### write test of the function
write a simple example. `function/math_calcu_test.go`:
```go
package function

import (
	"testing"

	bloc_client "github.com/fBloc/bloc-client-go"
)

func TestMathCalcu(t *testing.T) {
	blocClient := bloc_client.NewClient("test")
	funcRunOpt := blocClient.TestRunFunction(
		&MathCalcu{},
		[][]interface{}{
			{ // ipt 1 group, numbers
				[]interface{}{1, 2},
			},
			{ // ipt 2 group, arithmetic operator
				1,
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

During `function` directory, and run command `go test .`, you will see the PASS. which means your function run meet your expectation.

### report to the server
after make sure your function runs well, you can deploy it.

During `bloc_go_tryout` directory and make a `main.go` file with content:
```go
package main

import (
	"bloc-backend-examples/go/bloc_go_tryout/function"

	bloc_client "github.com/fBloc/bloc-client-go"
)

const appName = "tryout"

func main() {
	client := bloc_client.NewClient(appName)

	// config
	blocServerPort := 8080 // fake port
	client.GetConfigBuilder().SetRabbitConfig(
		"$user", "$password", []string{"host"}, "$vhost",
	).SetServer(
		"$blocServerIP", blocServerPort,
	).BuildUp()

	// register your functions
	sourceFunctionGroup := client.RegisterFunctionGroup("math") // give your function a group name
	sourceFunctionGroup.AddFunction(
		"calcu", // name your function node's name
		"receive numbers and do certain math operation to them", // the describe of your function node
		&function.MathCalcu{}, // your function implement
	)

	client.Run()
}
```

after replace the configs, input your bloc-server's address and rabbit address.

you can run it by:
```shell
$ go run main.go
```

after suc run, this client's all function node are registered to the bloc-server, which can be see and operate in the frontend, and this client will receive bloc-server's trigger function to run msg and run it.

demo code is [here](/examples/bloc_go_tryout).