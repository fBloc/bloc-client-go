package bloc_client

import (
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/fBloc/bloc-client-go/internal/http_util"
)

var serverUrl string

const logSubPath = "report_log"

type LogLevel = string

const (
	Info    LogLevel = "info"
	Warning LogLevel = "Warning"
	Error   LogLevel = "error"
)

type msg struct {
	Level  LogLevel          `json:"level"`
	TagMap map[string]string `json:"tag_map"`
	Data   string            `json:"data"`
	Time   time.Time         `json:"time"`
}

type Logger struct {
	name                string
	functionRunRecordID string
	sync.Mutex
}

func (logger *Logger) IsZero() bool {
	if logger == nil {
		return true
	}
	return logger.name == ""
}

func NewLogger(name, server, functionRunRecordID string) *Logger {
	serverUrl = server
	l := &Logger{
		name:                name,
		functionRunRecordID: functionRunRecordID}
	return l
}

func (
	logger *Logger,
) Infof(
	format string, a ...interface{},
) {
	go uploadMsg(&msg{
		Time:  time.Now(),
		Level: Info,
		Data:  fmt.Sprintf(format, a...),
		TagMap: map[string]string{
			"function_run_record_id": logger.functionRunRecordID},
	})
}

func (
	logger *Logger,
) Warningf(
	format string, a ...interface{},
) {
	go uploadMsg(&msg{
		Time:  time.Now(),
		Level: Warning,
		Data:  fmt.Sprintf(format, a...),
		TagMap: map[string]string{
			"function_run_record_id": logger.functionRunRecordID},
	})
}

func (
	logger *Logger,
) Errorf(
	format string, a ...interface{},
) {
	go uploadMsg(&msg{
		Time:  time.Now(),
		Level: Error,
		Data:  fmt.Sprintf(format, a...),
		TagMap: map[string]string{
			"function_run_record_id": logger.functionRunRecordID},
	})
}

type HttpReq struct {
	LogData []*msg `json:"logs"`
}

type HttpResp struct {
	Code int         `json:"status_code"`
	Msg  string      `json:"status_msg"`
	Data interface{} `json:"data"`
}

func uploadMsg(logMsg *msg) {
	httpReqData := HttpReq{LogData: []*msg{logMsg}}
	httpReqByte, err := json.Marshal(httpReqData)
	// log error would not do future handle as log is not crucial
	if err != nil {
		return
	}

	var resp HttpResp
	http_util.PostJson(
		path.Join(serverUrl, logSubPath),
		http_util.BlankHeader, httpReqByte, &resp)
}
