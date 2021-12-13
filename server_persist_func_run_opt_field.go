package bloc_client

import (
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const serverFuncRunOptPersistToObjectStoragePath = "persist_certain_function_run_opt_field"

type FuncRunOptPersistToObjectStorageHttpReq struct {
	FunctionRunRecordID string      `json:"function_run_record_id"`
	OptKey              string      `json:"opt_key"`
	Data                interface{} `json:"data"`
}

type FuncOptFieldServerPersisResp struct {
	ObjectStorageKey string `json:"object_storage_key"`
	Brief            string `json:"brief"`
}

type FuncRunOptPersistToObjectStorageHttpResp struct {
	StatusCode int                          `json:"status_code"`
	Data       FuncOptFieldServerPersisResp `json:"data"`
}

func (bC *BlocClient) PersistFunctionRunOptFieldToServer(
	funcRunRecordID string, OptFieldKey string,
	OptFieldValue interface{},
) (*FuncOptFieldServerPersisResp, error) {
	req := FuncRunOptPersistToObjectStorageHttpReq{
		FunctionRunRecordID: funcRunRecordID,
		OptKey:              OptFieldKey, Data: OptFieldValue}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var resp FuncRunOptPersistToObjectStorageHttpResp
	err = http_util.PostJson(
		bC.GenReqServerPath(serverFuncRunOptPersistToObjectStoragePath),
		http_util.BlankHeader, reqBody, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data, nil
}
