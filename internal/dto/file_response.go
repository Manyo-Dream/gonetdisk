package dto

import (
	"github.com/manyodream/gonetdisk/internal/model"
)

type FileUploadResponse struct {
	UserFileID uint64 `json:"userfile_id"`
	FileName   string `json:"file_name"`
	FileExt    string `json:"file_ext"`
	FIleSize   int64  `json:"file_size"`
	ParentID   uint64 `json:"parent_id"`
}

type FileDownloadMeta struct {
	FileName    string `json:"file_name"`
	StorageType string `json:"storage_type"`
	FileExt     string `json:"file_ext"`
	FileSize    int64  `json:"file_size"`
}

type FileListResponse struct {
	List     []model.UserFile `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
