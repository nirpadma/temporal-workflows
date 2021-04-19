package media_processing_workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const workflowMaxAttempts = 3

// MediaProcessingWorkflow defines a workflow that queries an API, downloads media files, encodes, and combines media.
// NOTE: The initial structure for this workflow came from https://github.com/temporalio/samples-go
func MediaProcessingWorkflow(ctx workflow.Context, outputFileName string) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		HeartbeatTimeout:    3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			// retry with a constant backoff coefficient (i.e constant time between retry intervals)
			// In real-world settings, it may be more apropriate to set a value > 1 
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= workflowMaxAttempts; i++ {
		err = processMediaWorkflow(ctx, outputFileName)
		if err == nil {
			break
		}
	}
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err.Error())
	} else {
		workflow.GetLogger(ctx).Info("Workflow completed.")
	}
	return err
}

func processMediaWorkflow(ctx workflow.Context, outputFileName string) (err error) {

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    3 * time.Minute,
		HeartbeatTimeout:       3 * time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	var status string
	err = workflow.ExecuteActivity(ctx1, CheckMediaStatusActivity).Get(ctx1, &status)
	if err != nil {
		logger.Error("CheckMediaStatusActivity failed", "Error", err)
		return err
	}

	if status == NotObtainable {
		logger.Info("Media not obtainable; finishing workflow")
		// any clean-up activities would go here.
		return nil
	}

	var mediaURLs []string
	err = workflow.ExecuteActivity(ctx1, GetMediaURLsActivity).Get(ctx1, &mediaURLs)
	if err != nil {
		logger.Error("GetMediaURLsActivity failed", "Error", err)
		return err
	}

	// Create and use the session API for the activities that need to be scheduled on the same host
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 8 * time.Minute,
		HeartbeatTimeout: 5 * time.Minute,
	}
	
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	downloadedfileNames := []string{}
	err = workflow.ExecuteActivity(sessionCtx, DownloadFilesActivity, mediaURLs).Get(sessionCtx, &downloadedfileNames)
	if err != nil {
		return err
	}

	encodedfileNames := []string{}
	for _, downloadedFile := range downloadedfileNames {
		logger.Info("encoding file", "file", downloadedFile)
		var encodedFileName string
		err = workflow.ExecuteActivity(sessionCtx, EncodeFileActivity, downloadedFile).Get(sessionCtx, &encodedFileName)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Encoded the following file: %s", encodedFileName))
		encodedfileNames = append(encodedfileNames, encodedFileName)
	}

	var mergedFile string
	err = workflow.ExecuteActivity(sessionCtx, MergeFilesActivity, encodedfileNames, outputFileName).Get(sessionCtx, &mergedFile)
	if err != nil {
		return err
	}

	return nil
}
