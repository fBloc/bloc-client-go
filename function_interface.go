package bloc_client

import (
	"context"
)

type FunctionDeveloperImplementInterface interface {
	Run(
		context.Context,
		Ipts,
		chan HighReadableFunctionRunProgress,
		chan *FunctionRunOpt,
		*Logger,
	)
	IptConfig() Ipts
	OptConfig() []*Opt
	AllProcessStages() []string
}
