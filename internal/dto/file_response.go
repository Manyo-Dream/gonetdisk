package dto

type FileUploadResponse struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
}
