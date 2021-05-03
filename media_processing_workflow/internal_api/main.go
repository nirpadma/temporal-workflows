package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nirpadma/temporal-workflows/media_processing_workflow"
)


const MAX_UPLOAD_SIZE = 500 * 1024 * 1024 // 500 MB


func uploadMediaHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Only POST Method is permitted", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too big.", http.StatusBadRequest)
		return
	}

    multipartFile, multipartFileHeader, err := r.FormFile(media_processing_workflow.FileNameAttribute)
    if err != nil {
        fmt.Println("Error Retrieving the File")
        fmt.Println(err)
		http.Error(w, "Error Retrieving the File.", http.StatusInternalServerError)
        return
    }
    defer multipartFile.Close()
    fmt.Printf("Uploaded File Name: %+v\n", multipartFileHeader.Filename)
    fmt.Printf("File Size bytes: %+v\n", multipartFileHeader.Size)
    fmt.Printf("MIME Header: %+v\n", multipartFileHeader.Header)

	// The uploaded files are stored within the `uploadedfiles` directory inside the internal_api
    tmpFile, err := ioutil.TempFile("uploadedfiles", "video-*.mp4")
    if err != nil {
        fmt.Println(err)
    }
    defer tmpFile.Close()

    fileBytes, err := ioutil.ReadAll(multipartFile)
    if err != nil {
		http.Error(w, "Error reading the contents of the uploaded file.", http.StatusInternalServerError)
        fmt.Println(err)
    }

	_, err = tmpFile.Write(fileBytes)
	if err != nil {
		http.Error(w, "Error writing the contents of the uploaded file.", http.StatusInternalServerError)
	}

	fmt.Fprintf(w, "File Uploaded.\n")
}


func main() {
	fmt.Println("Starting internal API server...")
	http.HandleFunc("/uploadmedia", uploadMediaHandler)
	_ = http.ListenAndServe(":9220", nil)
}
