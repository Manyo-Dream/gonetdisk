package dto

import "github.com/manyodream/gonetdisk/internal/model"

type TrashListResponse struct {
	List     []model.UserFile `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type TrashDeleteResponse struct {
	Message string `json:"message"`
}

type TrashRestoreResponse struct {
	Message string `json:"message"`
}
