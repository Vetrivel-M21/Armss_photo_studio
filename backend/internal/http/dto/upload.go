package dto

type UploadResponse struct {
	ID          string `json:"id"`
	ObjectKey   string `json:"objectKey,omitempty"`
	ContentType string `json:"contentType"`
	SizeBytes   int    `json:"sizeBytes"`
}
