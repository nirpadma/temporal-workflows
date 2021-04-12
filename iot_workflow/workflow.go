package iot_workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const workflowMaxAttempts = 3

// IOTWorkflow workflow definition
// NOTE: The initial structure for this workflow was taken from https://github.com/temporalio/samples-go
func IOTWorkflow(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		HeartbeatTimeout:    time.Second * 2,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			// retry every second. In real-world settings, it may be more apropriate to set an actual value
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i < workflowMaxAttempts; i++ {
		err = processIOTWorkflow(ctx)
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

func processIOTWorkflow(ctx workflow.Context) (err error) {

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    2 * time.Minute,
		HeartbeatTimeout:       time.Second * 10,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("starting CheckMediaStatusActivity")

	var isMediaReadyToDownload bool
	err = workflow.ExecuteActivity(ctx1, CheckMediaStatusActivity).Get(ctx1, &isMediaReadyToDownload)
	if err != nil {
		logger.Error("CheckMediaStatusActivity failed", "Error", err)
		return err
	}

	logger.Info("Media is ready to download", "isReady", isMediaReadyToDownload)

	var mediaURLs []string
	err = workflow.ExecuteActivity(ctx1, GetMediaURLsActivity).Get(ctx1, &mediaURLs)
	if err != nil {
		logger.Error("GetMediaURLsActivity failed", "Error", err)
		return err
	}

	logger.Info("mediaURLs: ", "mediaURLs", mediaURLs)

	so := &workflow.SessionOptions{
		CreationTimeout:  3 * time.Minute,
		ExecutionTimeout: 3 * time.Minute,
		HeartbeatTimeout: 3 * time.Minute,
	}
	// Use the session context for the activities to schedule on the same host
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	downloadedfileNames := []string{}
	for i, mediaFileURL := range mediaURLs {
		logger.Info(fmt.Sprintf("i=%d", i))
		logger.Info("downloading file", "mediaFileURL", mediaFileURL)
		var downloadedFileName string
		err = workflow.ExecuteActivity(sessionCtx, DownloadFileActivity, mediaFileURL).Get(sessionCtx, &downloadedFileName)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Downloaded the following file: %s", downloadedFileName))
		downloadedfileNames = append(downloadedfileNames, downloadedFileName)
	}

	// var encodedFilePath string
	// err = workflow.ExecuteActivity(sessionCtx, EncodeFileActivity, downloadedName).Get(sessionCtx, &encodedFilePath)
	// if err != nil {
	// 	return err
	// }

	// var uploadedFilePath string
	// err = workflow.ExecuteActivity(sessionCtx, UploadFileActivity, downloadedName).Get(sessionCtx, &uploadedFilePath)
	// if err != nil {
	// 	return err
	// }

	// var deleteSuccessful bool
	// err = workflow.ExecuteActivity(sessionCtx, DeleteFilesActivity, []string{downloadedName, encodedFilePath}).Get(sessionCtx, &deleteSuccessful)
	// if err != nil {
	// 	return err
	// }

	// if !deleteSuccessful {
	// 	fmt.Println(fmt.Sprintf("Error deleting %s or %s", downloadedName, encodedFilePath))
	// }

	return nil
}
