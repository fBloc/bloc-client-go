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
	Level LogLevel  `json:"level"`
	Data  string    `json:"data"`
	Time  time.Time `json:"time"`
}

type Logger struct {
	name string
	data []*msg
	sync.Mutex
}

func (logger *Logger) IsZero() bool {
	if logger == nil {
		return true
	}
	return logger.name == ""
}

func NewLogger(name, server string) *Logger {
	serverUrl = server
	l := &Logger{
		name: name,
	}
	go l.upload()
	return l
}

func (
	logger *Logger,
) Infof(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: Info,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *Logger,
) Warningf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: Warning,
		Data:  fmt.Sprintf(format, a...),
	})
}

func (
	logger *Logger,
) Errorf(format string, a ...interface{}) {
	logger.Lock()
	defer logger.Unlock()

	logger.data = append(logger.data, &msg{
		Time:  time.Now(),
		Level: Error,
		Data:  fmt.Sprintf(format, a...),
	})
}

type HttpReq struct {
	Name    string `json:"name"`
	LogData []*msg `json:"log_data"`
}

type HttpResp struct {
	Code int         `json:"status_code"`
	Msg  string      `json:"status_msg"`
	Data interface{} `json:"data"`
}

func (logger *Logger) ForceUpload() {
	if len(logger.data) <= 0 {
		return
	}

	// TODO 要不要panic？
	logger.Lock()
	httpReqData := HttpReq{
		Name:    logger.name,
		LogData: logger.data,
	}
	httpReqByte, err := json.Marshal(httpReqData)
	logger.data = logger.data[:0]
	logger.Unlock()
	if err != nil {
		panic(err)
	}

	var resp HttpResp
	err = http_util.PostJson(
		path.Join(serverUrl, logSubPath), http_util.BlankHeader,
		httpReqByte, &resp)
	if err != nil {
		panic(err)
	}
}

func (logger *Logger) upload() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if len(logger.data) <= 0 {
			continue
		}

		logger.Lock()
		httpReqData := HttpReq{
			Name:    logger.name,
			LogData: logger.data,
		}
		httpReqByte, err := json.Marshal(httpReqData)
		logger.data = logger.data[:0]
		logger.Unlock()
		if err != nil {
			panic(err)
		}

		var resp HttpResp
		err = http_util.PostJson(
			path.Join(serverUrl, logSubPath), http_util.BlankHeader,
			httpReqByte, &resp)
		if err != nil {
			panic(err)
		}
	}
}
