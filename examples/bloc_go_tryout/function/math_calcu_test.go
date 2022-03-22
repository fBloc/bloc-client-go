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
