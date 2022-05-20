package main

import (
	bloc_client "github.com/fBloc/bloc-client-go"
	"github.com/fBloc/bloc-client-go/examples/bloc_go_tryout/bloc_node"
)

const appName = "tryout"

func main() {
	client := bloc_client.NewClient(appName)

	// config
	client.GetConfigBuilder().SetRabbitConfig(
		"blocRabbit", "blocRabbitPasswd", []string{"127.0.0.1:5672"}, "", // bloc rabbitMQ address
	).SetServer(
		"127.0.0.1", 8080, // bloc-backend-server address
	).BuildUp()

	// register your functions
	sourceFunctionGroup := client.RegisterFunctionGroup("math") // give your function a group name
	sourceFunctionGroup.AddFunction(
		"calcu", // name your function node's name
		"receive numbers and do certain math operation to them", // the describe of your function node
		&bloc_node.MathCalcu{}, // your function implement
	)

	client.Run()
}
