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

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, jwtManager)
	userController := controller.NewUserController(userService)

	v1 := r.Group("/api/v1")
	{
		user := v1.Group("/user")
		{
			user.POST("/register", userController.Register)
			user.POST("/login", userController.Login)
		}
	}
	return r
}
