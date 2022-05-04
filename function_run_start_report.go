package bloc_client

import (
	"context"
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const FuncRunStartHttpPath = "function_run_start"

type FuncRunStartHttpReq struct {
	FunctionRunRecordID string `json:"function_run_record_id"`
}

func newFuncRunStartHttpReq(
	functionRunRecordID string,
) *FuncRunStartHttpReq {
	return &FuncRunStartHttpReq{
		FunctionRunRecordID: functionRunRecordID,
	}
}

func (bC *blocClient) ReportFuncRunStart(
	ctx context.Context,
	functionRunRecordID string,
) error {
	funcRunStartReq := newFuncRunStartHttpReq(functionRunRecordID)
	body, err := json.Marshal(*funcRunStartReq)
	if err != nil {
		return err
	}

	var resp interface{}
	header := map[string]string{
		string(TraceID): GetTraceIDFromContext(ctx),
		string(SpanID):  GetSpanIDFromContext(ctx)}
	err = http_util.PostJson(
		bC.GenReqServerPath(FuncRunStartHttpPath),
		header, body, &resp)
	return err
}
