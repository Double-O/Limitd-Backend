package middleware

import (
	"net/http"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(client *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customErr := utils.IsAccessTokenValid(ctx, client)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
