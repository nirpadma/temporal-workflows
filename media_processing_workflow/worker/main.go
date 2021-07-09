package main

import (
	"log"

	"github.com/nirpadma/temporal-workflows/media_processing_workflow"
	"github.com/xfrr/goffmpeg/transcoder"
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
	w := worker.New(c, "mediaprocessing", workerOptions)

	transcoder := new(transcoder.Transcoder)
	activity := media_processing_workflow.Activities{
		VendorAPIMediaStatusTemplate: media_processing_workflow.VendorAPIMediaStatusTemplate,
		VendorAPIMediaURLsTemplate:   media_processing_workflow.VendorAPIMediaURLsTemplate,
		Transcoder:                   transcoder,
		OutputFileType:               media_processing_workflow.EncodedOutputFileType,
		FileUploadEndpoint:           media_processing_workflow.FileUploadEndpoint,
	}

	w.RegisterWorkflow(media_processing_workflow.MediaProcessingWorkflow)
	w.RegisterActivity(&activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
