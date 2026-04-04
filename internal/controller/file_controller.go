package controller

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/middleware"
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

	// 控制器从 token 取 email
	email, ok := middleware.GetEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未认证用户"})
		return
	}

	// 上传物理文件
	resp, err := c.FileService.UploadPhyFileAndBindFile(email, req.ParentID, fileHeader)
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{"error": "上传文件失败:" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (c *FileController) DownloadFile(ctx *gin.Context) {
	var req dto.FileDownloadRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误:" + err.Error(),
		})
		return
	}

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户",
		})
		return
	}

	filemata, file, err := c.FileService.DownloadUserFile(userID, req.UserFileID)
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"message": err.Error(),
		})
		return
	}
	defer file.Close()

	contentType := mime.TypeByExtension(filepath.Ext(filemata.FileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx.Header("X-Content-Type-Options", "nosniff")
	encodedFilename := c.buildContentDisposition(filemata.FileName)
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", encodedFilename))
	ctx.DataFromReader(http.StatusOK, filemata.FileSize, contentType, file, nil)
}

func (c *FileController) buildContentDisposition(filename string) string {
	// ASCII 回退：过滤非 ASCII 字符
	asciiName := strings.Map(func(r rune) rune {
		if r < 128 && r != '"' && r != '\\' && r != '\r' && r != '\n' {
			return r
		}
		return '_'
	}, filename)

	if asciiName == filename {
		return fmt.Sprintf("attachment; filename=\"%s\"", asciiName)
	}

	// 同时提供两种格式
	encoded := url.QueryEscape(filename)

	// filename*=UTF-8''<percent-encoded>
	return fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		asciiName, encoded)
}
