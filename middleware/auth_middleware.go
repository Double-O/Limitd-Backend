package middleware

import (
	"net/http"

	"github.com/Double-O/Limitd-Backend/service"
	"github.com/dgrijalva/jwt-go"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(
	userService service.UserService,
	client *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// check acess token validity
		customErr := utils.IsAccessTokenValid(ctx, client)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			ctx.Abort()
			return
		}

		// get the actual access token
		accessToken, customErr := utils.GetAccessToken(ctx)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			ctx.Abort()
			return
		}

		// get the claims(including token_uuid, user uuid)
		accessTokenClaims := accessToken.Claims.(jwt.MapClaims)
		userUuid := accessTokenClaims["uuid"].(string)

		// get the user and set it from gin context
		user, customErr := userService.FindUserByUUID(userUuid)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			ctx.Abort()
			return
		}
		ctx.Set(utils.USER, user)

		// get the accessUuid and set it from gin context
		accessUuid := accessTokenClaims[utils.TOKEN_UUID].(string)
		ctx.Set(utils.ACCESS_TOKEN_UUID, accessUuid)

		ctx.Next()
	}
}
