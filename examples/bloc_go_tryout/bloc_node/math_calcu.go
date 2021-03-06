package bloc_node

import (
	"context"
	"fmt"
	"time"

	bloc_client "github.com/fBloc/bloc-client-go"
)

func init() {
	var _ bloc_client.BlocFunctionNodeInterface = &MathCalcu{}
}

type MathCalcu struct {
}

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

// IptConfig define function node's ipt config
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
					AllowMulti: false, // only allow single select value
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
