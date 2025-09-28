package domain

const (
	Bucket = "static-file"
)

type ObjectUploadResp struct {
	Key string `json:"key"`
}

type ObjectDownloadResp struct {
	URL string `json:"url"`
}
