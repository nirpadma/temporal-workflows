package iot_workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// IOTWorkflow workflow definition
// NOTE: The initial template for this code was taken from https://github.com/temporalio/samples-go
func IOTWorkflow(ctx workflow.Context, fileURL string) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		HeartbeatTimeout:    time.Second * 2,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i < 5; i++ {
		err = processFile(ctx, fileURL)
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

func processFile(ctx workflow.Context, fileURL string) (err error) {
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: time.Minute,
	}
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	var downloadedName string
	var a *Activities
	err = workflow.ExecuteActivity(sessionCtx, a.DownloadFileActivity, fileURL).Get(sessionCtx, &downloadedName)
	if err != nil {
		return err
	}

	var encodedFilePath string
	err = workflow.ExecuteActivity(sessionCtx, a.EncodeFileActivity, downloadedName).Get(sessionCtx, &encodedFilePath)

	return nil
}
