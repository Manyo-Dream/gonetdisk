package router

import (
	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/controller"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/service"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, jwtManager *util.JWTManager) *gin.Engine {
	r := gin.Default()

	userRepo := repository.NewUserRepo(db)
	userService := service.NewUserService(userRepo, jwtManager)
	userController := controller.NewUserController(userService)
	fileRepo := repository.NewFileRepo(db)
	fileService := service.NewFileService(userRepo, fileRepo, jwtManager)
	fileController := controller.NewFileController(fileService)

	v1 := r.Group("/api/v1")
	{
		user := v1.Group("/user")
		{
			user.POST("/register", userController.Register)
			user.POST("/login", userController.Login)
			user.GET("/info", userController.GetUserInfo)
			user.PUT("/info", userController.UpdateUserInfo)
		}
		fileRepo := v1.Group("/file")
		{
			fileRepo.POST("/upload", fileController.UploadFile)
		}
	}
	return r
}
