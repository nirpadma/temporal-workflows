package media_processing_workflow

const (
	// media URL statuses
	Success       = "success"
	Pending       = "pending"
	NotObtainable = "not_obtainable"

	// vendor API statuses
	VendorAPIMediaStatus = "http://localhost:8220/mediastatus"
	VendorAPIMediaURLs   = "http://localhost:8220/mediaurls"

	// encoding output type
	EncodedOutputFileType = "mp4"

	// upload file name attribute
	FileNameAttribute = "uploadfile"
	FileUploadEndpoint = "http://localhost:9220/uploadmedia"
)

// MediaURLs is the struct for the json response of /mediaurls endpoint
type MediaURLs struct {
	Links []string `json:"urls"`
}
