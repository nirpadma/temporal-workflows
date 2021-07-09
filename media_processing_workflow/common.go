package media_processing_workflow

const (
	// media URL statuses
	Success       = "success"
	Pending       = "pending"
	NotObtainable = "not_obtainable"

	// vendor API statuses
	VendorAPIMediaStatusTemplate = "http://localhost:8220/mediastatus/%s"
	VendorAPIMediaURLsTemplate   = "http://localhost:8220/mediaurls/%s"

	// encoding output type
	EncodedOutputFileType = "mp4"

	// upload file name attribute
	FileNameAttribute = "uploadfile"
	FileUploadEndpoint = "http://localhost:9220/uploadmedia"
)

// MediaURLs is the struct for the json response of /mediaurls endpoint
type MediaURLs struct {
	DeviceId string `json:"deviceId"`
	Links []string `json:"urls"`
}

// MediaStatus is the struct for the json response of /mediastatus endpoint
type MediaStatus struct {
	DeviceId string `json:"deviceId"`
	Status string `json:"status"`
}
