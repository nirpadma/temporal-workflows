package media_processing_workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/xfrr/goffmpeg/transcoder"
	"go.temporal.io/sdk/activity"
)

const vendorAPImediaStatus = "http://localhost:8220/mediastatus"
const vendorAPImediaURLs = "http://localhost:8220/mediaurls"

// CheckMediaStatusActivity ...
func CheckMediaStatusActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	resp, err := http.Get(vendorAPImediaStatus)
	if err != nil {
		logger.Error("http err calling vendor API for status", "endpoint", vendorAPImediaStatus)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", vendorAPImediaStatus)
		return "", err
	}
	status := string(bodyBytes)
	switch status {
		case Success :
			return Success, nil
		case Pending :
			return Pending, errors.New("media still pending")
		case NotObtainable :
			return NotObtainable, nil
		default:
			return NotObtainable, nil
	}
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", vendorAPImediaStatus)
		return []string{}, err
	}
	var urls MediaURLs
	json.Unmarshal(bodyBytes, &urls)

	return urls.Links, nil
}

// Create a temporary files and download the media files at the provided fileURL into the temp files
// return an array containing paths to the temp files.
func DownloadFilesActivity(ctx context.Context, fileURLs []string) ([]string, error) {
	logger := activity.GetLogger(ctx)
	downloadedFiles := []string{}
	for _, fileURL := range fileURLs {

		logger.Info("Downloading file...", "fileURL", fileURL)

		tmpFile, err := ioutil.TempFile("", "videoFile")
		if err != nil {
			logger.Error(fmt.Sprintf("Err creating temp file %s", err.Error()))
			return downloadedFiles, err
		}

		filePath := tmpFile.Name()

		file, err := os.Create(filePath)
		if err != nil {
			logger.Error("Err creating file to save to", "filePath", filePath)
			return downloadedFiles, err
		}
		defer file.Close()

		logger.Info(fmt.Sprintf("created file with name %s", file.Name()))

		// For potentially long running activites, record a heartbeat
		activity.RecordHeartbeat(ctx, "")

		resp, err := http.Get(fileURL)
		if err != nil {
			logger.Error("http error downloading file", "fileURL", fileURL)
			return downloadedFiles, err
		}
		defer resp.Body.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			logger.Error("Error copying downloaded file to filepath")
			return downloadedFiles, err
		}
		// For potentially long running activites, record a heartbeat
		activity.RecordHeartbeat(ctx, "")

		logger.Info(fmt.Sprintf("saved file with name %s", file.Name()))
		downloadedFiles = append(downloadedFiles, file.Name())
	}
	return downloadedFiles, nil
}

// EncodeFileActivity encodes the downloaded file into the expected output
func EncodeFileActivity(ctx context.Context, fileName string) (string, error) {
	logger := activity.GetLogger(ctx)
	tmpFile, err := ioutil.TempFile("", "encodedFile")
	if err != nil {
		logger.Error(fmt.Sprintf("Err creating temp file %s", err.Error()))
		return "", err
	}
	outputFilePath := fmt.Sprintf("%s.mp4", tmpFile.Name())

	transcoder := new(transcoder.Transcoder)
	err = transcoder.Initialize(fileName, outputFilePath)

	if err != nil {
		logger.Error(fmt.Sprintf("Err initializing ffmpeg transcoder %s", err.Error()))
		return "", err
	}
	// Start transcoder with the `true` flag to show the progress
	done := transcoder.Run(true)

	progress := transcoder.Output()

	// print out transcoding progress
	for msg := range progress {
		// For potentially long running activites, record a heartbeat
		activity.RecordHeartbeat(ctx, "")
		fmt.Println(msg)
	}

	err = <-done
	if err != nil {
		logger.Error(fmt.Sprintf("Err in transcoding %s", err.Error()))
		return "", err
	}

	return outputFilePath, nil
}

func writefileNamesToFile(path string, fileNames []string) error {
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer file.Close()

	for _, f := range fileNames {
		_, err = file.WriteString(fmt.Sprintf("file '%s'\n", f))
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}

	err = file.Sync()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func createTempFile(prefix string) (string, error) {
	tmpFile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

// MergeFilesActivity combines the media files into one file based on the ordering in the input array
func MergeFilesActivity(ctx context.Context, fileNames []string, outputFileName string) (string, error) {
	logger := activity.GetLogger(ctx)
	filesToMerge, err := createTempFile("filesToMerge")
	err = writefileNamesToFile(filesToMerge, fileNames)
	if err != nil {
		return "", err
	}

	//Use ffmpeg concatenate instructions from here: https://trac.ffmpeg.org/wiki/Concatenate

	ffMpegCommand := "ffmpeg"
	arg0 := "-f"
	arg1 := "concat"
	arg2 := "-safe"
	arg3 := "0"
	arg4 := "-i"
	arg5 := filesToMerge
	arg6 := "-c"
	arg7 := "copy"
	arg8 := outputFileName

	cmd := exec.Command(ffMpegCommand, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	stdout, err := cmd.Output()
	if err != nil {
		logger.Error("error executing command")
		logger.Error(err.Error())
		return "", err
	}
	logger.Info(string(stdout))
	return outputFileName, nil
}
