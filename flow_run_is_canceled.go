package bloc_client

import "github.com/fBloc/bloc-client-go/internal/http_util"

const FlowRunIsCanceledPath = "/check_flowRun_is_canceled_by_flowRunID/"

type FlowRunIsCanceledHttpResp struct {
	StatusCode int `json:"status_code"`
	Data       struct {
		Canceled bool `json:"canceled"`
	} `json:"data"`
	Canceled bool `json:"canceled"`
}

func (bC *blocClient) FlowRunIsCanceled(
	flowRunRecordID string,
) (bool, error) {
	var resp FlowRunIsCanceledHttpResp
	err := http_util.Get(
		bC.GenReqServerPath(FlowRunIsCanceledPath, flowRunRecordID),
		http_util.BlankHeader,
		&resp)
	if err != nil {
		return false, err
	}
	return resp.Data.Canceled, nil
}
