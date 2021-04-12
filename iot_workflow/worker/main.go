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

	// these worker options are necessary for the worker to participate in a session
	// see https://docs.temporal.io/docs/go-sessions/ for more details
	workerOptions := worker.Options{
		EnableSessionWorker: true,
	}
	w := worker.New(c, "iotprocessing", workerOptions)

	w.RegisterWorkflow(iot_workflow.IOTWorkflow)
	w.RegisterActivity(iot_workflow.CheckMediaStatusActivity)
	w.RegisterActivity(iot_workflow.GetMediaURLsActivity)
	w.RegisterActivity(iot_workflow.DownloadFileActivity)
	w.RegisterActivity(iot_workflow.DownloadFilesActivity)
	w.RegisterActivity(iot_workflow.EncodeFileActivity)
	w.RegisterActivity(iot_workflow.MergeFilesActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
