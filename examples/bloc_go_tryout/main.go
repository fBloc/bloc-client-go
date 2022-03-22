package main

import (
	"bloc_go_tryout/function"

	bloc_client "github.com/fBloc/bloc-client-go"
)

const appName = "tryout"

func main() {
	client := bloc_client.NewClient(appName)

	// config
	blocServerPort := 8080 // fake port, change it
	client.GetConfigBuilder().SetRabbitConfig(
		"$user", "$password", []string{"host"}, "$vhost",
	).SetServer(
		"$blocServerIP", blocServerPort,
	).BuildUp()

	// register your functions
	sourceFunctionGroup := client.RegisterFunctionGroup("math") // give your function a group name
	sourceFunctionGroup.AddFunction(
		"calcu", // name your function node's name
		"receive numbers and do certain math operation to them", // the describe of your function node
		&function.MathCalcu{}, // your function implement
	)

	client.Run()
}
