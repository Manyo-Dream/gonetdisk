package dto

type TrashListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}
