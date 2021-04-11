package main

import (
	"log"

	"github.com/nirpadma/temporal-workflows/iot_workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "iotprocessing", worker.Options{})

	w.RegisterWorkflow(iot_workflow.IOTWorkflow)
	w.RegisterActivity(iot_workflow.CheckMediaStatusActivity)
	w.RegisterActivity(iot_workflow.GetMediaURLsActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
