package bloc_client

import (
	"github.com/fBloc/bloc-client-go/internal/http_util"
)

const fetchObjectStorageDataByKeyFromServerPath = "get_byte_value_by_key"

type ServerObjectStorageHttpResp struct {
	StatusCode int    `json:"status_code"`
	Data       []byte `json:"data"`
}

func (bC *BlocClient) FetchObjectStorageDataByKeyFromServer(
	key string,
) ([]byte, error) {
	var resp ServerObjectStorageHttpResp
	err := http_util.Get(
		bC.GenReqServerPath(fetchObjectStorageDataByKeyFromServerPath, key),
		http_util.BlankHeader, &resp)
	return resp.Data, err
}
