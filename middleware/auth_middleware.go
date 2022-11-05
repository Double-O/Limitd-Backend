package middleware

import (
	"net/http"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(client *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		customErr := utils.IsTokenValid(c, client)
		if customErr != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
