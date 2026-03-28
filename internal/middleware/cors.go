package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	allowMethods  = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowHeaders  = "Origin, Content-Type, Content-Length, Accept, Authorization, X-Requested-With"
	exposeHeaders = "Content-Length, Content-Disposition"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Vary", "Origin")
		ctx.Header("Access-Control-Allow-Methods", allowMethods)
		ctx.Header("Access-Control-Allow-Headers", allowHeaders)
		ctx.Header("Access-Control-Expose-Headers", exposeHeaders)

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
