package router

import (
	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/controller"
	"github.com/manyodream/gonetdisk/internal/middleware"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/service"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, jwtManager *util.JWTManager) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	userRepo := repository.NewUserRepo(db)
	userService := service.NewUserService(userRepo, jwtManager)
	userController := controller.NewUserController(userService)

	fileRepo := repository.NewFileRepo(db)
	fileService := service.NewFileService(userRepo, fileRepo, jwtManager)
	fileController := controller.NewFileController(fileService)

	folderService := service.NewFolderService(userRepo, fileRepo, jwtManager)
	folderController := controller.NewFolderController(folderService)

	v1 := r.Group("/api/v1")
	{
		userRepo := v1.Group("/user")
		{
			userRepo.POST("/register", userController.Register)
			userRepo.POST("/login", userController.Login)
		}
		userRepo.Use(middleware.AuthMiddleware(jwtManager))
		{
			userRepo.GET("/info", userController.GetUserInfo)
			userRepo.PUT("/info", userController.UpdateUserInfo)
		}

		fileRepo := v1.Group("/file")
		fileRepo.Use(middleware.AuthMiddleware(jwtManager))
		{
			fileRepo.POST("/upload", fileController.UploadFile)
			fileRepo.GET("/download/:userfile_id", fileController.DownloadFile)
		}

		folderRepo := v1.Group("/folder")
		folderRepo.Use(middleware.AuthMiddleware(jwtManager))
		{
			folderRepo.POST("/create", folderController.CreateFolder)
		}
	}
	return r
}
