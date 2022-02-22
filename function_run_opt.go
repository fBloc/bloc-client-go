package bloc_client

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

func CanceldBlocOpt() *FunctionRunOpt {
	return &FunctionRunOpt{Canceled: true}
}
