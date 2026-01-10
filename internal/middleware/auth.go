package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/util"
)

func AuthMiddleware(jwtManager *util.JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token格式"})
			ctx.Abort()
			return
		}

		claims, err := jwtManager.VerifyToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			ctx.Abort()
			return
		}

		ctx.Set("username", claims.Username)
		ctx.Set("email", claims.Email)

		ctx.Next()
	}
}
