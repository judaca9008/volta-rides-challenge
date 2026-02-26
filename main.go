package main

import (
	"voltarides/smart-router/cmd/httpServer"
	"voltarides/smart-router/common/constants"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func main() {
	// Start DataDog tracer (Yuno standard)
	tracer.Start(
		tracer.WithService(constants.MicroserviceName),
	)
	defer tracer.Stop()

	// Initialize and start server
	server := httpServer.EchoServer{}
	server.RunServer()
}
