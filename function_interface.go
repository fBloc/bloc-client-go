package bloc_client

import (
	"context"
)

// BlocFunctionNodeInterface is the interface of a function_node in bloc,
// once you develop a function_node implement this interface and register it to bloc_server,
// your function_node can be used in bloc.
type BlocFunctionNodeInterface interface {
	// AllProgressMilestones defines your function's execute milestones.
	// when the function is running and if you report the function's current progress,
	// the frontend user can see it.
	// An example of function visit remote http address 4 weather:
	// 		["parse param suc", "start to call remote api", "finished"]
	// I personal recommend to set it especially when your function is a long run function
	AllProgressMilestones() []string

	// Ipts defines your function's input params
	IptConfig() Ipts

	// Opts defines your function's output data
	OptConfig() Opts

	// Run the logic of your code
	// Ipts param carry the value of your function's params in the same order and struct you defined in IptConfig(),
	// 		which means you can get the real input value from it
	// HighReadableFunctionRunProgress chan is used to report the function's current progress.
	// 		you can report the function's running msg/progress percent/milestone index to show in the frontend.
	// FunctionRunOpt chan is used to report the function's result.
	// Logger is log your function's log
	// 		this log is for developer, this log can also directely show in the frontend.
	// 		the difference between this log and the HighReadableFunctionRunProgress.Msg is that
	// 		HighReadableFunctionRunProgress.Msg is "log" for others but not the developer himself/herself.
	Run(
		context.Context,
		Ipts,
		chan HighReadableFunctionRunProgress,
		chan *FunctionRunOpt,
		*Logger,
	)
}
