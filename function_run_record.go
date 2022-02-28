package bloc_client

import (
	"time"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

type briefAndKey struct {
	ObjectStorageKey string `json:"object_storage_key"`
}

type FunctionRunRecord struct {
	ID                          string          `json:"id"`
	FunctionID                  string          `json:"function_id"`
	FlowRunRecordID             string          `json:"flow_run_record_id"`
	TraceID                     string          `json:"trace_id"`
	IptBriefAndObjectStoragekey [][]briefAndKey `json:"ipt"`
	Canceled                    bool            `json:"canceled"`
	ShouldBeCanceledAt          time.Time       `json:"should_be_canceled_at"`
}

type FuncRecordHttpResp struct {
	StatusCode        int               `json:"status_code"`
	FunctionRunRecord FunctionRunRecord `json:"data"`
}

const functionRecordPath = "get_function_run_record_by_id"

func (bC *BlocClient) GetFunctionRunRecordByID(
	funcRunRecordID string,
) (*FunctionRunRecord, error) {
	var resp FuncRecordHttpResp
	err := http_util.Get(
		bC.GenReqServerPath(functionRecordPath, funcRunRecordID),
		http_util.BlankHeader,
		&resp)
	if err != nil {
		return nil, err
	}
	return &resp.FunctionRunRecord, nil
}
