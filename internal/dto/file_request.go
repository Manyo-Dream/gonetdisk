package dto

type FileUploadRequest struct {
	ParentID uint64 `form:"parent_id"`
}

type FileDownloadRequest struct {
	UserFileID uint64 `uri:"userfile_id" binding:"required,min=1"`
}
