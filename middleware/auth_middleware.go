package middleware

import (
	"net/http"

	"github.com/Double-O/Limitd-Backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		customErr := utils.IsTokenValid(c)
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
