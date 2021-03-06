package media_processing_workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const sessionMaxAttempts = 3

// MediaProcessingWorkflow defines a workflow that queries an API, downloads media files, encodes, and combines media.
// NOTE: The initial structure for this workflow was inspired by https://github.com/temporalio/samples-go
func MediaProcessingWorkflow(ctx workflow.Context, deviceId string, outputFileName string) (err error) {

	logger := workflow.GetLogger(ctx)
	// use an exponential retry policy for activities where "real world" delays may occur
	expAO := workflow.ActivityOptions{
		StartToCloseTimeout:    1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, expAO)

	var a *Activities
	var status string
	err = workflow.ExecuteActivity(ctx, a.CheckMediaStatusActivity, deviceId).Get(ctx, &status)
	if err != nil {
		logger.Error("CheckMediaStatusActivity failed", "Error", err)
		return err
	}

	// End the workflow early if the media is never obtainable
	if status == NotObtainable {
		logger.Info("Media not obtainable; finishing workflow")
		// any clean-up activities would go here.
		return nil
	}

	uniformAO := workflow.ActivityOptions{
		StartToCloseTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			BackoffCoefficient: 1.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, uniformAO)

	var mediaURLs []string
	err = workflow.ExecuteActivity(ctx, a.GetMediaURLsActivity, deviceId).Get(ctx, &mediaURLs)
	if err != nil {
		logger.Error("GetMediaURLsActivity failed", "Error", err)
		return err
	}

	for i := 1; i <= sessionMaxAttempts; i++ {
		err = processMediaFiles(ctx, mediaURLs, outputFileName)
		if err == nil {
			break
		} 
		logger.Error("processMediaFiles errored. Retrying...")
	}

	if err != nil {
		logger.Error("Processing Media in Session Failed.", "Error", err.Error())
	} else {
		logger.Info("Processing Media in Session Succeeded.")
	}
	return err
}

func processMediaFiles(ctx workflow.Context, mediaFilesOfInterest []string, outputFileName string) (err error) {
	// Create and use the session API for the activities that need to be scheduled on the same host
	so := &workflow.SessionOptions{
		CreationTimeout:  3 * time.Minute,
		ExecutionTimeout: 3 * time.Minute,
	}

	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	logger := workflow.GetLogger(sessionCtx)

	var a *Activities

	downloadedfileNames := []string{}
	err = workflow.ExecuteActivity(sessionCtx, a.DownloadFilesActivity, mediaFilesOfInterest).Get(sessionCtx, &downloadedfileNames)
	if err != nil {
		return err
	}

	encodedfileNames := []string{}
	for _, downloadedFile := range downloadedfileNames {
		logger.Info("encoding file", "file", downloadedFile)
		var encodedFileName string
		err = workflow.ExecuteActivity(sessionCtx, a.EncodeFileActivity, downloadedFile).Get(sessionCtx, &encodedFileName)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Encoded the following file: %s", encodedFileName))
		encodedfileNames = append(encodedfileNames, encodedFileName)
	}

	var mergedFile string
	err = workflow.ExecuteActivity(sessionCtx, a.MergeFilesActivity, encodedfileNames, outputFileName).Get(sessionCtx, &mergedFile)
	if err != nil {
		return err
	}

	var uploadSuccess bool
	err = workflow.ExecuteActivity(sessionCtx, a.UploadFileActivity, mergedFile).Get(sessionCtx, &uploadSuccess)
	if err != nil {
		return err
	}

	return nil
}
