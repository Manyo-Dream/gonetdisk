package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/dto"
	"github.com/manyodream/gonetdisk/internal/service"
)

type UserController struct {
	UserService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{UserService: userService}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误:" + err.Error(),
		})
		return
	}

	if err := c.UserService.Register(req.Email, req.Username, req.Password); err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "用户注册失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "用户注册成功",
	})
}

func (c *UserController) Login(ctx *gin.Context) {
	var req dto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误:" + err.Error(),
		})
		return
	}

	resp, err := c.UserService.Login(req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "用户登录失败:" + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}