package controller

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

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

	fileName := filemata.FileName
	escapedName := url.PathEscape(fileName)
	disposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=utf-8''%s", escapedName, escapedName)

	ctx.Header("Content-Disposition", disposition)
	ctx.Header("X-Content-Type-Options", "nosniff")

	ctx.DataFromReader(http.StatusOK, filemata.FileSize, contentType, file, nil)
}

func (c *FileController) ReturnFileList(ctx *gin.Context) {
	var req dto.FileListRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
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

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 5
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	resp, err := c.FileService.GetUserFileList(
		userID,
		req.ParentID,
		req.Page,
		req.PageSize,
		req.SortBy,
		req.OrderBy,
	)
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"error": "获取文件列表失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": resp,
	})
}

func (c *FileController) MoveFileToTrash(ctx *gin.Context) {
	// 获取用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户",
		})
		return
	}

	// 获取用户文件ID
	userFileIDStr := ctx.Param("userfile_id")
	if userFileIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "用户文件ID为空",
		})
		return
	}

	userFileID, err := strconv.Atoi(userFileIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误" + err.Error(),
		})
		return
	}

	resp, err := c.FileService.MoveFileToTrash(userID, uint64(userFileID))
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"error": "移动到回收站失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": resp,
	})
}

func (c *FileController) RestoreFile(ctx *gin.Context) {
	// 获取用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户",
		})
		return
	}

	// 获取用户文件ID
	userFileIDStr := ctx.Param("userfile_id")
	if userFileIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "用户文件ID为空",
		})
		return
	}

	userFileID, err := strconv.Atoi(userFileIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误" + err.Error(),
		})
		return
	}

	resp, err := c.FileService.RestoreFile(userID, uint64(userFileID))
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"error": "移动到回收站失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": resp,
	})
}

func (c *FileController) ReturnTrashList(ctx *gin.Context) {
	var req dto.TrashListRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
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

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 5
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	resp, err := c.FileService.GetTrashList(
		userID,
		req.Page,
		req.PageSize,
	)
	if err != nil {
		ctx.JSON(statusFromErr(err), gin.H{
			"error": "获取文件列表失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": resp,
	})
}
