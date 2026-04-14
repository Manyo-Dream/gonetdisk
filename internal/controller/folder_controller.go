package controller

import (
	"net/http"
	"strconv"

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

func (fdc *FolderController) MoveFolderToTrash(ctx *gin.Context) {
	// 获取用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户",
		})
		return
	}

	// 获取用户文件夹ID
	userFolderIDStr := ctx.Param("userfolder_id")
	if userFolderIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "用户文件ID为空",
		})
		return
	}

	userFolderID, err := strconv.Atoi(userFolderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误" + err.Error(),
		})
		return
	}

	resp, err := fdc.FolderServicer.MoveFolderToTrash(userID, uint64(userFolderID))
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

func (fdc *FolderController) RestoreFolder(ctx *gin.Context) {
		// 获取用户ID
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "未认证用户",
		})
		return
	}

	// 获取用户文件夹ID
	userFolderIDStr := ctx.Param("userfolder_id")
	if userFolderIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "用户文件ID为空",
		})
		return
	}

	userFolderID, err := strconv.Atoi(userFolderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误" + err.Error(),
		})
		return
	}

	resp, err := fdc.FolderServicer.RestoreFolder(userID, uint64(userFolderID))
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