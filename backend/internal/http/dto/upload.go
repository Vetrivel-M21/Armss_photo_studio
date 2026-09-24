package dto

type UploadResponse struct {
	ID          string `json:"id"`
	ContentType string `json:"contentType"`
	SizeBytes   int    `json:"sizeBytes"`
}
