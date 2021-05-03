package media_processing_workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/xfrr/goffmpeg/transcoder"
	"go.temporal.io/sdk/activity"
)

type Activities struct {
	VendorAPIMediaStatus string
	VendorAPIMediaURLs   string
	Transcoder           *transcoder.Transcoder
	OutputFileType       string
	FileUploadEndpoint   string
}

/**
NOTE: Use these activities only as a general guide. For production settings, there may be modifications to be made.
For instance, in the encoding activity, we use a simple ffmpeg wrapper to do the encoding. In production settings,
there may need to be additional configurations, modifications, or settings that are necessary.
**/

// CheckMediaStatusActivity checks vendor API to determine whether the media is ready to be downloaded
func (a *Activities) CheckMediaStatusActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	resp, err := http.Get(a.VendorAPIMediaStatus)
	if err != nil {
		logger.Error("http err calling vendor API for status", "endpoint", a.VendorAPIMediaStatus)
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", a.VendorAPIMediaStatus)
		return "", err
	}
	status := string(bodyBytes)
	switch status {
	case Success:
		return Success, nil
	case Pending:
		return Pending, errors.New("media still pending")
	case NotObtainable:
		return NotObtainable, nil
	default:
		return NotObtainable, nil
	}
}

// GetMediaURLsActivity obtains the media URLs to be downloaded from the vendor
func (a *Activities) GetMediaURLsActivity(ctx context.Context) ([]string, error) {
	logger := activity.GetLogger(ctx)
	resp, err := http.Get(a.VendorAPIMediaURLs)
	if err != nil {
		logger.Error("http err calling vendor API for status", "endpoint", a.VendorAPIMediaURLs)
		return []string{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ioutil err reading mediastatus response", "endpoint", a.VendorAPIMediaURLs)
		return []string{}, err
	}
	var urls MediaURLs
	json.Unmarshal(bodyBytes, &urls)

	return urls.Links, nil
}

// DownloadFilesActivity creates temporary files and download the media files at the provided fileURLs into the temp files
// and return an array containing paths to the temp files.
// As a side effect, the activity records heartbeats of the activity execution to the Temporal service
func (a *Activities) DownloadFilesActivity(ctx context.Context, fileURLs []string) ([]string, error) {
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
// **NOTE:** In production settings, we'd want to update up this function to better
// handle specifics of the media encoding. This is a simple activity to illustrate
// an end-to-end example using Temporal.
func (a *Activities) EncodeFileActivity(ctx context.Context, fileName string) (string, error) {
	logger := activity.GetLogger(ctx)
	tmpFile, err := ioutil.TempFile("", "encodedFile")
	if err != nil {
		logger.Error(fmt.Sprintf("Err creating temp file %s", err.Error()))
		return "", err
	}
	outputFilePath := fmt.Sprintf("%s.%s", tmpFile.Name(), a.OutputFileType)

	err = a.Transcoder.Initialize(fileName, outputFilePath)

	if err != nil {
		logger.Error(fmt.Sprintf("Err initializing ffmpeg transcoder %s", err.Error()))
		return "", err
	}
	// Start transcoder with the `true` flag to show the progress
	done := a.Transcoder.Run(true)

	progress := a.Transcoder.Output()

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

func deleteTempFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return errors.New("unable to delete file")
	}
	return nil
}

// MergeFilesActivity combines the media files into one file based on the ordering in the input array
func (a *Activities) MergeFilesActivity(ctx context.Context, fileNames []string, outputFileName string) (string, error) {
	logger := activity.GetLogger(ctx)

	// Use ffmpeg to concatenate instructions from here: https://trac.ffmpeg.org/wiki/Concatenate
	// The recommended approach utilizes a file that includes a list of files to merge.
	fileContaingFilesToMerge, err := createTempFile("filesToMerge")
	err = writefileNamesToFile(fileContaingFilesToMerge, fileNames)
	if err != nil {
		return "", err
	}

	ffMpegCommand := "ffmpeg"
	arg0 := "-f"
	arg1 := "concat"
	arg2 := "-safe"
	arg3 := "0"
	arg4 := "-i"
	arg5 := fileContaingFilesToMerge
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

	err = deleteTempFile(fileContaingFilesToMerge)
	if err != nil {
		logger.Error(fmt.Sprintf("unable to delete file %s", fileContaingFilesToMerge))
	}

	return outputFileName, nil
}

// UploadFileActivity uploads the provided file to the internal API
func (a *Activities) UploadFileActivity(ctx context.Context, fileName string) (bool, error) {
	targetUrl := a.FileUploadEndpoint

	buffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(buffer)

	fileWriter, err := bodyWriter.CreateFormFile(FileNameAttribute, fileName)
	if err != nil {
		fmt.Println("error creating form file")
		return false, err
	}

	fh, err := os.Open(fileName)
	if err != nil {
		fmt.Println("error while opening file")
		return false, err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return false, err
	}

	formDataContentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, formDataContentType, buffer)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	fmt.Println(resp.Status)
	fmt.Println(fmt.Sprintf("Response body: %s", string(respBody)))

	// Delete File as a side effect; Ideally, move this into its own Activity. 
	if resp.StatusCode == int(http.StatusOK) {
		deleteTempFile(fileName)
	}

	return true, nil
}
