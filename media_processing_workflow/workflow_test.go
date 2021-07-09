package media_processing_workflow

import (
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

// Test the `not_obtainable` status
func (s *UnitTestSuite) Test_MediaProcessingWorkflow_NotObtainable() {
	env := s.NewTestWorkflowEnvironment()
	var a *Activities
	env.OnActivity(a.CheckMediaStatusActivity, mock.Anything, mock.Anything).Return(NotObtainable, nil)
	fileID := uuid.New()
	outputfileName := "mediaprocessing_" + fileID
	env.ExecuteWorkflow(MediaProcessingWorkflow, outputfileName, "deviceId")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// Test the `success` status and downstream activities
func (s *UnitTestSuite) Test_MediaProcessingWorkflow_NoError() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	})
	var a *Activities

	env.RegisterActivity(a.CheckMediaStatusActivity)
	env.RegisterActivity(a.GetMediaURLsActivity)
	env.RegisterActivity(a.DownloadFilesActivity)
	env.RegisterActivity(a.EncodeFileActivity)
	env.RegisterActivity(a.MergeFilesActivity)

	env.OnActivity(a.CheckMediaStatusActivity, mock.Anything, mock.Anything).Return(Success, nil)
	env.OnActivity(a.GetMediaURLsActivity, mock.Anything, mock.Anything).Return([]string{"url1", "url2"}, nil)
	env.OnActivity(a.DownloadFilesActivity, mock.Anything, []string{"url1", "url2"}).Return([]string{"download1", "download2"}, nil)
	env.OnActivity(a.EncodeFileActivity, mock.Anything, "download1").Return("encode1", nil)
	env.OnActivity(a.EncodeFileActivity, mock.Anything, "download2").Return("encode2", nil)
	env.OnActivity(a.MergeFilesActivity, mock.Anything, []string{"encode1", "encode2"}, mock.Anything).Return("output.mp4", nil)
	env.OnActivity(a.UploadFileActivity, mock.Anything, "output.mp4", mock.Anything).Return(true, nil)

	fileID := uuid.New()
	outputfileName := "mediaprocessing_" + fileID
	env.ExecuteWorkflow(MediaProcessingWorkflow, outputfileName, "deviceId")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
