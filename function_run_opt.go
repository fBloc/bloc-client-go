package bloc_client

import "fmt"

type FunctionRunOpt struct {
	Suc                       bool
	Canceled                  bool
	TimeoutCanceled           bool
	InterceptBelowFunctionRun bool // 拦截后续的运行
	ErrorMsg                  string
	Description               string
	Detail                    map[string]interface{}
	KeyMapObjectStorageKey    map[string]string
	Brief                     map[string]string
}

func NewFailedFunctionRunOpt(format string, a ...interface{}) *FunctionRunOpt {
	return &FunctionRunOpt{
		Suc:      false,
		ErrorMsg: fmt.Sprintf(format, a...)}
}

func CanceldBlocOpt() *FunctionRunOpt {
	return &FunctionRunOpt{Canceled: true}
}

func NewTimeoutCanceldFunctionRunOpt() *FunctionRunOpt {
	return &FunctionRunOpt{
		TimeoutCanceled: true, Canceled: true, InterceptBelowFunctionRun: true}
}
