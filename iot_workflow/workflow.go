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
func IOTWorkflow(ctx workflow.Context, outputFileName string) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 3 * time.Minute,
		HeartbeatTimeout:    3 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval: time.Second,
			// retry every second. In real-world settings, it may be more apropriate to set an actual value
			BackoffCoefficient: 1.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i < workflowMaxAttempts; i++ {
		err = processIOTWorkflow(ctx, outputFileName)
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

func processIOTWorkflow(ctx workflow.Context, outputFileName string) (err error) {

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    3 * time.Minute,
		HeartbeatTimeout:       3 * time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	logger.Info("starting CheckMediaStatusActivity")

	err = workflow.ExecuteActivity(ctx1, CheckMediaStatusActivity).Get(ctx1, nil)
	if err != nil {
		logger.Error("CheckMediaStatusActivity failed", "Error", err)
		return err
	}

	var mediaURLs []string
	err = workflow.ExecuteActivity(ctx1, GetMediaURLsActivity).Get(ctx1, &mediaURLs)
	if err != nil {
		logger.Error("GetMediaURLsActivity failed", "Error", err)
		return err
	}

	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 8 * time.Minute,
		HeartbeatTimeout: 5 * time.Minute,
	}
	// Use the session context for the activities to schedule on the same host
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
