package router

import (
	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/configs"
	"github.com/manyodream/gonetdisk/internal/controller"
	"github.com/manyodream/gonetdisk/internal/middleware"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/service"
	"github.com/manyodream/gonetdisk/internal/util"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, jwtManager *util.JWTManager, config *configs.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	userRepo := repository.NewUserRepo(db)
	userService := service.NewUserService(userRepo, jwtManager)
	userController := controller.NewUserController(userService)

	fileRepo := repository.NewFileRepo(db)
	fileService := service.NewFileService(userRepo, fileRepo, jwtManager, config)
	fileController := controller.NewFileController(fileService)

	folderService := service.NewFolderService(userRepo, fileRepo, jwtManager)
	folderController := controller.NewFolderController(folderService)

	v1 := r.Group("/api/v1")
	{
		userHandler := v1.Group("/user")
		{
			userHandler.POST("/register", userController.Register)
			userHandler.POST("/login", userController.Login)
		}
		userHandler.Use(middleware.AuthMiddleware(jwtManager, userRepo))
		{
			userHandler.GET("/info", userController.GetUserInfo)
			userHandler.PUT("/info", userController.UpdateUserInfo)
		}

		fileHandler := v1.Group("/file")
		fileHandler.Use(middleware.AuthMiddleware(jwtManager, userRepo))
		{
			fileHandler.POST("/upload", fileController.UploadFile)
			fileHandler.GET("/download/:userfile_id", fileController.DownloadFile)
			fileHandler.DELETE("/delete/:userfile_id", fileController.MoveFileToTrash)
			fileHandler.GET("/list", fileController.ReturnFileList)
		}

		folderHandler := v1.Group("/folder")
		folderHandler.Use(middleware.AuthMiddleware(jwtManager, userRepo))
		{
			folderHandler.POST("/create", folderController.CreateFolder)
			folderHandler.DELETE("/delete/:userfolder_id", folderController.MoveFolderToTrash)
		}

		trashHandler := v1.Group("/trash")
		trashHandler.Use(middleware.AuthMiddleware(jwtManager, userRepo))
		{
			trashHandler.GET("/list", fileController.ReturnTrashList)
			trashHandler.POST("/file/:userfile_id", fileController.RestoreFile)
			trashHandler.POST("/folder/:userfolder_id", folderController.RestoreFolder)
		}
	}
	return r
}
