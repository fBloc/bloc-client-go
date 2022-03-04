package bloc_client

import (
	"context"
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const FuncRunFinishedHttpPath = "function_run_finished"

type FuncRunFinishedHttpReq struct {
	FunctionRunRecordID       string            `json:"function_run_record_id"`
	Suc                       bool              `json:"suc"`
	Canceled                  bool              `json:"canceled"`
	TimeoutCanceled           bool              `json:"timeout_canceled"`
	InterceptBelowFunctionRun bool              `json:"intercept_below_function_run"`
	ErrorMsg                  string            `json:"error_msg"`
	Description               string            `json:"description"`
	OptKeyMapBriefData        map[string]string `json:"optKey_map_briefData"`
	OptKeyMapObjectStorageKey map[string]string `json:"optKey_map_objectStorageKey"`
}

func newFuncRunFinishedHttpReqFromFuncOpt(
	functionRunRecordID string, opt FunctionRunOpt,
) *FuncRunFinishedHttpReq {
	return &FuncRunFinishedHttpReq{
		FunctionRunRecordID:       functionRunRecordID,
		Suc:                       opt.Suc,
		Canceled:                  opt.Canceled,
		TimeoutCanceled:           opt.TimeoutCanceled,
		ErrorMsg:                  opt.ErrorMsg,
		Description:               opt.Description,
		OptKeyMapBriefData:        opt.Brief,
		OptKeyMapObjectStorageKey: opt.KeyMapObjectStorageKey,
	}
}

func (bC *blocClient) ReportFuncRunFinished(
	ctx context.Context,
	functionRunRecordID string, opt FunctionRunOpt,
) error {
	funcRunFinishedReq := newFuncRunFinishedHttpReqFromFuncOpt(
		functionRunRecordID, opt)
	body, err := json.Marshal(*funcRunFinishedReq)
	if err != nil {
		return err
	}

	var resp interface{}
	header := map[string]string{
		string(TraceID): GetTraceIDFromContext(ctx),
		string(SpanID):  GetSpanIDFromContext(ctx)}
	err = http_util.PostJson(
		bC.GenReqServerPath(FuncRunFinishedHttpPath),
		header, body, &resp)
	return err
}
