package iot_workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"go.temporal.io/sdk/activity"
)

const vendorAPImediaStatus = "http://localhost:8220/mediastatus"
const vendorAPImediaURLs = "http://localhost:8220/mediaurls"

// CheckMediaStatusActivity ...
func CheckMediaStatusActivity(ctx context.Context) (bool, error) {
	logger := activity.GetLogger(ctx)
	resp, err := http.Get(vendorAPImediaStatus)
	if err != nil {
		logger.Error("http err calling vendor API for status", "endpoint", vendorAPImediaStatus)
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", vendorAPImediaStatus)
		return false, err
	}
	status := string(bodyBytes)
	if status == "success" {
		return true, nil
	}
	return false, errors.New("media not ready to download")
}

// GetMediaURLsActivity ...
func GetMediaURLsActivity(ctx context.Context) ([]string, error) {
	logger := activity.GetLogger(ctx)
	resp, err := http.Get(vendorAPImediaURLs)
	if err != nil {
		logger.Error("http err calling vendor API for status", "endpoint", vendorAPImediaStatus)
		return []string{}, err
	}
	defer resp.Body.Close()

	// logger.Info("bodyBytes: ", "resp.Body", resp.Body)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", vendorAPImediaStatus)
		return []string{}, err
	}
	var urls MediaURLs
	json.Unmarshal(bodyBytes, &urls)

	return urls.Links, nil
}

// Create a temporary file and download the media file at the provided fileURL into the temp file
// return the path to the temp file.
func DownloadFileActivity(ctx context.Context, fileURL string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file...", "fileURL", fileURL)

	tmpFile, err := ioutil.TempFile("", "videoFile")
	if err != nil {
		logger.Error(fmt.Sprintf("Err creating temp file %s", err.Error()))
		return "", err
	}

	filePath := tmpFile.Name()

	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("Err creating file to save to", "filePath", filePath)
		return "", err
	}
	defer file.Close()

	logger.Info(fmt.Sprintf("created file with name %s", file.Name()))

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
func EncodeFileActivity(ctx context.Context, fileName string) (string, error) {
	return "", nil
}

// Upload the encoded file to an API
func UploadFileActivity(ctx context.Context, fileName string) error {
	return nil
}

// Do cleanup activity of the specified file names
func DeleteFilesActivity(ctx context.Context, fileNames []string) error {
	return nil
}
