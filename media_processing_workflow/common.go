package media_processing_workflow

const (
	Success = "success"
	Pending = "pending"
)

// MediaURLs is the struct for the json response of /mediaurls endpoint
type MediaURLs struct {
	Links []string `json:"urls"`
}
