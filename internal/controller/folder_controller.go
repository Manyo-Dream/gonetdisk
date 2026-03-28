package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/middleware"
	"github.com/manyodream/gonetdisk/internal/service"
)

type FolderController struct {
	FolderServicer *service.FolderService
}

func NewFolderController(folderServicer *service.FolderService) *FolderController {
	return &FolderController{FolderServicer: folderServicer}
}

func (fdc *FolderController) CreateFolder(ctx *gin.Context) {
	var req dto.FolderRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误:" + err.Error(),
		})
		return
	}

	email, ok := middleware.GetEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户"})
		return
	}

	resp, err := fdc.FolderServicer.CreateFolder(email, req.FolderName, req.ParentID)
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"error": "创建文件夹失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
