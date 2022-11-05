package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Double-O/Limitd-Backend/domain/response_entity"
	"github.com/dgrijalva/jwt-go"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/domain/custom_error"

	"github.com/Double-O/Limitd-Backend/service"

	"github.com/Double-O/Limitd-Backend/utils"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/rs/zerolog"

	"github.com/Double-O/Limitd-Backend/domain/request_entity"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

// we will assume no registration should be needed from user's perspective
// he will just simply login
// ReqBody  : request_entity.HandleLoginRequest
// RespBody : response_entity.HandleLoginResponse
// PathParam : None
func HandleLogin(
	userService service.UserService,
	redisClient *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginInput request_entity.HandleLoginRequest
		ctx.BindJSON(&loginInput)
		googleClientID := os.Getenv("GOOGLE_CLIENT_ID")

		//validate the google token
		payload, err := idtoken.Validate(context.Background(), loginInput.Token, googleClientID)
		if err != nil {
			errorMessage := fmt.Sprintf(utils.InvalidGoogleTokenMsg, loginInput.Token, googleClientID)
			customErr := custom_error.NewErrorFromMessage("InvalidGoogleTokenMsg", errorMessage)
			logger.LogMessage(zerolog.ErrorLevel, "controller.auth_controller", "Login", errorMessage)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		// check if token values and request body value matches or not
		if ok, customErr := validateGoogleClaim(payload.Claims, loginInput); !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		// check if 3rd party issuer value is valid
		if !utils.ValidFroms[loginInput.From] {
			errorMessage := fmt.Sprintf(utils.InvalidThirdPartyIssuerMsg, loginInput.Token)
			customErr := custom_error.NewErrorFromMessage("InvalidThirdPartyIssuerMsg", errorMessage)
			logger.LogMessage(zerolog.ErrorLevel, "controller.auth_controller", "Login", errorMessage)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		// check if the user exists
		user, customErr := userService.FindUserByEmail(loginInput.Email)
		// if not, create the user
		// TODO: if we implement soft delete, the logic here needs to be changed
		// because if user is soft deleted, it will return error hee and will try to create the user
		// but the db won't let us create as unique email constraint is present
		// currently assuming we will implement hard delete
		if customErr != nil {
			user, customErr = userService.CreateUser(&loginInput)
			if customErr != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"result": customErr,
				})
				return
			}
		}

		// issue the new/existing user a refresh and jwt token
		tokenDetails, customErr := utils.CreateToken(user.UUID)
		if customErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		customErr = utils.SaveTokenInRedis(context.Background(), redisClient, user.UUID, tokenDetails)
		if customErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		handleLonginResponse := response_entity.HandleLoginResponse{
			User:        response_entity.ConvertUserToUserResponse(user),
			AccessToken: tokenDetails.AccessToken,
		}

		// setting refresh token in the cookie
		// TODO: depending on env, we can use a secured cookie
		cookieDomain := os.Getenv("SHARED_COOKIE_DOMAIN")
		ctx.SetCookie(utils.REFRESH_TOKEN, tokenDetails.RefreshToken, utils.RT_EXPIRATION_TIME_COOKIE_SECOND, "/", cookieDomain, false, true)

		ctx.JSON(http.StatusCreated, gin.H{
			"result": handleLonginResponse,
		})
	}

}

func validateGoogleClaim(claims map[string]interface{}, loginInput request_entity.HandleLoginRequest) (bool, *custom_error.Error) {

	misMatchedField := ""
	misMatch := false
	misMatchValueInput := ""
	misMatchValueClaims := ""

	if claims["email"] != loginInput.Email {
		misMatch = true
		misMatchedField = "email"
		misMatchValueInput = loginInput.Email
		misMatchValueClaims = claims["email"].(string)
	} else if claims["given_name"] != loginInput.FirstName {
		misMatch = true
		misMatchedField = "first_name"
		misMatchValueInput = loginInput.FirstName
		misMatchValueClaims = claims["given_name"].(string)
	} else if claims["family_name"] != loginInput.LastName {
		misMatch = true
		misMatchedField = "last_name"
		misMatchValueInput = loginInput.LastName
		misMatchValueClaims = claims["family_name"].(string)
	}

	if misMatch {
		errorMessage := fmt.Sprintf(utils.MismatchTokenAndLoginReqMsg, misMatchedField, misMatchValueClaims, misMatchValueInput)
		customErr := custom_error.NewErrorFromMessage("MismatchTokenAndLoginReqMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "controller.auth_controller", "validateClaim", errorMessage)
		return false, customErr
	}

	return true, nil
}

// ReqBody  : None
// RespBody : response_entity.HandleLoginResponse
// PathParam : None
func HandleRefresh(
	userService service.UserService,
	redisClient *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//check refresh token validity
		customErr := utils.IsRefreshTokenValid(ctx, redisClient)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			return
		}

		// get the actual refresh token
		refreshToken, customErr := utils.GetRefreshTOken(ctx)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			return
		}

		// get the claims(including token_uuid, user uuid)
		refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
		refreshUuid := refreshTokenClaims[utils.TOKEN_UUID].(string)
		userUuid := refreshTokenClaims["uuid"].(string)

		//delete the refresh token from redis
		customErr = utils.DeleteToken(ctx, refreshUuid, redisClient)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			return
		}

		user, customErr := userService.FindUserByUUID(userUuid)
		if customErr != nil {
			ctx.JSON(http.StatusForbidden, gin.H{
				"result": customErr,
			})
			return
		}

		// TODO: Next part is also used in login request, should be in a separate function
		// to prevent copy of same logic

		// issue the new/existing user a refresh and jwt token
		tokenDetails, customErr := utils.CreateToken(user.UUID)
		if customErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		customErr = utils.SaveTokenInRedis(context.Background(), redisClient, user.UUID, tokenDetails)
		if customErr != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"result": customErr,
			})
			return
		}

		handleLonginResponse := response_entity.HandleLoginResponse{
			User:        response_entity.ConvertUserToUserResponse(user),
			AccessToken: tokenDetails.AccessToken,
		}

		// setting refresh token in the cookie
		// TODO: depending on env, we can use a secured cookie
		cookieDomain := os.Getenv("SHARED_COOKIE_DOMAIN")
		ctx.SetCookie(utils.REFRESH_TOKEN, tokenDetails.RefreshToken, utils.RT_EXPIRATION_TIME_COOKIE_SECOND, "/", cookieDomain, false, true)

		ctx.JSON(http.StatusCreated, gin.H{
			"result": handleLonginResponse,
		})

	}
}
