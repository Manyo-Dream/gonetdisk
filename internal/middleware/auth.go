package middleware
import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/manyodream/gonetdisk/internal/repository"
	"github.com/manyodream/gonetdisk/internal/util"
)
func AuthMiddleware(jwtManager *util.JWTManager, userRepo *repository.UserRepo) gin.HandlerFunc {
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

		userID, err := strconv.ParseUint(claims.RegisteredClaims.Subject, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			ctx.Abort()
			return
		}

		user, err := userRepo.GetUserByID(userID)
		if err != nil || user.Status != 0 {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "用户已被禁用"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.RegisteredClaims.Subject)
		ctx.Set("username", claims.Username)
		ctx.Set("email", claims.Email)
		ctx.Next()
	}
}
func GetUsername(ctx *gin.Context) (string, bool) {
	username, exists := ctx.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}
func GetEmail(ctx *gin.Context) (string, bool) {
	email, exists := ctx.Get("email")
	if !exists {
		return "", false
	}
	v, ok := email.(string)
	return v, ok
}
func GetUserID(ctx *gin.Context) (uint64, bool) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return 0, false
	}
	vStr, ok := userID.(string)
	v, err := strconv.ParseUint(vStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, ok
}
