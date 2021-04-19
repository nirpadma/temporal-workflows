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
	env.OnActivity(CheckMediaStatusActivity, mock.Anything).Return(NotObtainable, nil)
	fileID := uuid.New()
	outputfileName := "mediaprocessing_" + fileID
	env.ExecuteWorkflow(MediaProcessingWorkflow, outputfileName)
	
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}


// Test the `success` status but without proceeding with the subsequent activities
func (s *UnitTestSuite) Test_MediaProcessingWorkflow_Error() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	})

	env.RegisterActivity(CheckMediaStatusActivity)
	env.RegisterActivity(GetMediaURLsActivity)
	env.RegisterActivity(DownloadFilesActivity)
	env.RegisterActivity(EncodeFileActivity)
	env.RegisterActivity(MergeFilesActivity)

	env.OnActivity(CheckMediaStatusActivity, mock.Anything).Return(Success, nil)
	env.OnActivity(GetMediaURLsActivity, mock.Anything).Return([]string{"url1", "url2"}, nil)
	env.OnActivity(DownloadFilesActivity, mock.Anything, []string{"url1", "url2"}).Return([]string{"download1", "download2"}, nil)
	env.OnActivity(EncodeFileActivity, mock.Anything, "download1").Return("encode1", nil)
	env.OnActivity(EncodeFileActivity, mock.Anything, "download2").Return("encode2", nil)
	env.OnActivity(MergeFilesActivity, mock.Anything, []string{"encode1", "encode2"}, mock.Anything).Return("output.mp4", nil)

	fileID := uuid.New()
	outputfileName := "mediaprocessing_" + fileID
	env.ExecuteWorkflow(MediaProcessingWorkflow, outputfileName)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
