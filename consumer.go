package bloc_client

import (
	"fmt"
	"time"
)

func (bC *blocClient) Run() {
	// Periodic register function.
	// only panic when startup.
	err := bC.RegisterFunctionsToServer()
	if err != nil {
		panic(
			fmt.Sprintf(
				"failed to register functions to server: %v. Check whether server is alive",
				err),
		)
	}
	go func(b *blocClient) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			go bC.RegisterFunctionsToServer()
		}
	}(bC)

	// function consumer
	go bC.FunctionRunConsumer()

	forever := make(chan struct{})
	<-forever
}
