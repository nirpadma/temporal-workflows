package iot_workflow

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"go.temporal.io/sdk/activity"
)

type Activities struct {
}

// Create a temporary file and download the media file at the provided fileURL into the temp file
// return the path to the temp file.
func (a *Activities) DownloadFileActivity(ctx context.Context, fileURL string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file...", "fileURL", fileURL)

	tmpFile, err := ioutil.TempFile("dir", "videoFile")
	if err != nil {
		logger.Error("Err creating temp file")
		return "", err
	}

	filePath := tmpFile.Name()

	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("Err creating file to save to", "filePath", filePath)
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(fileURL)
	if err != nil {
		logger.Error("http error downloading file", "fileURL", fileURL)
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Error("Error copying downloaded file to filepath")
		return "", err
	}

	return file.Name(), nil
}

// Encode the downloaded file into the expected output
func (a *Activities) EncodeFileActivity(ctx context.Context, fileName string) (string, error) {
	return "", nil
}

// Do cleanup activity of the original downloaded file and the encoded file
func (a *Activities) DeleteFileActivity(ctx context.Context, fileName string) error {
	return nil
}

// Upload the encoded file to an API
func (a *Activities) UploadFileActivity(ctx context.Context, fileName string) error {
	return nil
}
