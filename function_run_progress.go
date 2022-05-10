package bloc_client

import (
	"context"
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const FuncRunProgressReportPath = "/report_progress"

type HighReadableFunctionRunProgress struct {
	Progress               float32 `json:"progress"`
	Msg                    string  `json:"msg"`
	ProgressMilestoneIndex *int    `json:"progress_milestone_index"`
}

type progressReportHttpReq struct {
	FunctionRunRecordID string                          `json:"function_run_record_id"`
	FuncRunProgress     HighReadableFunctionRunProgress `json:"high_readable_run_progress"`
}

func (bC *blocClient) ReportFuncRunProgress(
	ctx context.Context,
	funcRunRecordID string,
	progress float32,
	msg string,
	index *int,
) error {
	p := HighReadableFunctionRunProgress{}
	dataValid := false
	if progress > 0 {
		p.Progress = progress
		dataValid = true
	}
	if msg != "" {
		p.Msg = msg
		dataValid = true
	}
	if index != nil {
		p.ProgressMilestoneIndex = index
		dataValid = true
	}
	if !dataValid {
		return nil
	}

	body, err := json.Marshal(progressReportHttpReq{
		FunctionRunRecordID: funcRunRecordID,
		FuncRunProgress:     p})
	if err != nil {
		return err
	}

	var resp interface{}
	header := map[string]string{
		string(TraceID): GetTraceIDFromContext(ctx),
		string(SpanID):  GetSpanIDFromContext(ctx)}
	err = http_util.PostJson(
		bC.GenReqServerPath(FuncRunProgressReportPath),
		header, body, &resp)
	return err
}
