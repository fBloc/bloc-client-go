package bloc_client

import (
	"context"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const FuncHeartBeat = "report_functionExecute_heartbeat"

func (bC *blocClient) ReportFuncExecuteHeartbeat(
	ctx context.Context, functionRunRecordID string,
) error {
	var resp interface{}
	header := map[string]string{
		string(TraceID): GetTraceIDFromContext(ctx),
		string(SpanID):  GetSpanIDFromContext(ctx)}
	err := http_util.Get(
		bC.GenReqServerPath(FuncHeartBeat, functionRunRecordID),
		header, &resp)
	return err
}
