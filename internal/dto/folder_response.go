package dto

type FolderResponse struct {
	FolderName string `json:"folder_name" form:"folder_name" binding:"required"`
	ParentID   uint64 `json:"parent_id" form:"parent_id"`
}
