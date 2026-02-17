package dto

type FileUploadRequest struct {
	Email    string `form:"email"`
	ParentID uint64 `form:"parentID"`
}
