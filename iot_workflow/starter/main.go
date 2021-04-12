package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nirpadma/temporal-workflows/iot_workflow"
	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	// Since the client is a heavyweight object, only create once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create temporal client", err)
	}
	defer c.Close()

	fileID := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "iotprocessing_" + fileID,
		TaskQueue: "iotprocessing",
	}
	outputFileName := fmt.Sprintf("mergedFile_%s.mp4", fileID)

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, iot_workflow.IOTWorkflow, outputFileName)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
