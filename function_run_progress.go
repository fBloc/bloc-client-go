package bloc_client

import (
	"context"
	"encoding/json"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const FuncRunProgressReportPath = "/report_progress"

type HighReadableFunctionRunProgress struct {
	Progress          float32 `json:"progress"`
	Msg               string  `json:"msg"`
	ProcessStageIndex int     `json:"process_stage_index"`
}

type progressReportHttpReq struct {
	FunctionRunRecordID string                          `json:"function_run_record_id"`
	FuncRunProgress     HighReadableFunctionRunProgress `json:"high_readable_run_progress"`
}

func (bC *BlocClient) ReportFuncRunProgress(
	ctx context.Context,
	funcRunRecordID string,
	progress float32, msg string, index int,
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
	if index > 0 {
		p.ProcessStageIndex = index
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
		string(TraceID): GetTraceIDFromContext(ctx)}
	err = http_util.PostJson(
		bC.configBuilder.ServerConf.String()+FuncRunProgressReportPath,
		header, body, &resp)
	return err
}
