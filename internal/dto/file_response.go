package dto

type FileUploadResponse struct {
	UserFileID  uint64 `json:"userfile_id"`
	FileName    string `json:"file_name"`
	FileExt     string `json:"file_ext"`
	FIleSize    int64  `json:"file_size"`
	ParentID    uint64 `json:"parent_id"`
}

type FileDownloadMeta struct {
	FileName    string `json:"file_name"`
	StorageType string `json:"storage_type"`
	FileExt     string `json:"file_ext"`
	FileSize    int64  `json:"file_size"`
}
