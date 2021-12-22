package bloc_client

import (
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const registerFuncPath = "register_functions"

type HttpReqFunction struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	GroupName     string   `json:"group_name"`
	Description   string   `json:"description"`
	Ipts          []*Ipt   `json:"ipts"`
	Opts          []*Opt   `json:"opts"`
	ProcessStages []string `json:"process_stages"`
}

type HttpRespFunction struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GroupName string `json:"group_name"`
	ErrorMsg  string `json:"error_msg"`
}

type RegisterFuncResp struct {
	StatusCode int `json:"status_code"`
	Data       struct {
		GroupNameMapFunctions map[string][]*HttpRespFunction `json:"groupName_map_functions"`
	} `json:"data"`
}

type RegisterFuncReq struct {
	Who                   string                        `json:"who"`
	GroupNameMapFunctions map[string][]*HttpReqFunction `json:"groupName_map_functions"`
}
type GroupNameMapFunctions map[string][]*HttpReqFunction

func (bC *BlocClient) RegisterFunctionsToServer() {
	req := RegisterFuncReq{
		Who:                   bC.Name,
		GroupNameMapFunctions: make(map[string][]*HttpReqFunction)}

	for _, funcGroup := range bC.FunctionGroups {
		groupName := funcGroup.Name
		req.GroupNameMapFunctions[groupName] = make(
			[]*HttpReqFunction, len(funcGroup.Functions))
		for i, function := range funcGroup.Functions {
			req.GroupNameMapFunctions[groupName][i] = &HttpReqFunction{
				Name:          function.Name,
				GroupName:     function.GroupName,
				Description:   function.Description,
				Ipts:          function.Ipts,
				Opts:          function.Opts,
				ProcessStages: function.ProcessStages,
			}
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	var resp RegisterFuncResp
	err = http_util.PostJson(
		bC.GenReqServerPath(registerFuncPath),
		http_util.BlankHeader, body, &resp)
	if err != nil {
		panic(err)
	}

	for _, funcGroup := range bC.FunctionGroups {
		groupName := funcGroup.Name
		respFunctions := resp.Data.GroupNameMapFunctions[groupName]
		nameMapRespFunc := make(map[string]*HttpRespFunction, len(respFunctions))
		for _, f := range respFunctions {
			if f.ErrorMsg != "" {
				panic(f.ErrorMsg)
			}
			nameMapRespFunc[f.Name] = f
		}

		for _, function := range funcGroup.Functions {
			function.ID = nameMapRespFunc[function.Name].ID
		}
	}
}
