package function

import (
	"context"
	"fmt"
	"time"

	bloc_client "github.com/fBloc/bloc-client-go"
)

func init() {
	var _ bloc_client.FunctionDeveloperImplementInterface = &MathCalcu{}
}

type MathCalcu struct {
}

func (*MathCalcu) AllProgressMilestones() []string {
	return []string{
		"start parsing ipt",
		"start do the calculation",
		"finished do the calculation",
	}
}

// IptConfig define function node's ipt config
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

type SucInMathCalcuAnimals struct {
	ID         int       `json:"id"`
	InputTime  time.Time `json:"input_time"`
	OutputTime time.Time `json:"output_time"`
}

// OptConfig
func (*MathCalcu) OptConfig() []*bloc_client.Opt {
	return []*bloc_client.Opt{
		{
			Key:         "result",
			Description: "arithmetic operation result",
			ValueType:   bloc_client.IntValueType,
			IsArray:     false,
		},
	}
}

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
