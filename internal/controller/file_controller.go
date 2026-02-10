package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/service"
)

type FileController struct {
	FileService *service.FileService
}

func NewFileController(fileService *service.FileService) *FileController {
	return &FileController{FileService: fileService}
}

func (c *FileController) UploadFile(ctx *gin.Context) {
	// 获取请求参数 ParentID
	var req dto.FileUploadRequest

	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误:" + err.Error(),
		})
		return
	}

	// 获取文件流
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "获取文件失败:" + err.Error()})
		return
	}

	// 上传文件
	resp, err := c.FileService.UploadPhyFile(fileHeader)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "上传文件失败"})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
