package bloc_client

import (
	"context"

	"github.com/fBloc/bloc-client-go/internal/log"
)

type FunctionDeveloperImplementInterface interface {
	Run(
		context.Context,
		Ipts,
		chan HighReadableFunctionRunProgress,
		chan *FunctionRunOpt,
		*log.Logger,
	)
	IptConfig() Ipts
	OptConfig() []*Opt
	AllProcessStages() []string
}
