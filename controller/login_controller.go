package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Double-O/Limitd-Backend/utils"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/rs/zerolog"

	"github.com/Double-O/Limitd-Backend/domain/request_entity"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func HandleLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginInput request_entity.LoginRequest
		ctx.BindJSON(&loginInput)
		googleClientID := os.Getenv("GOOGLE_CLIENT_ID")

		//validate the google token
		payload, err := idtoken.Validate(context.Background(), loginInput.Token, googleClientID)
		if err != nil {
			errorMessage := fmt.Sprintf(utils.InvalidGoogleTokenMsg, loginInput.Token, googleClientID)
			logger.LogMessage(zerolog.ErrorLevel, "controller.login_controller", "Login", errorMessage)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": errorMessage,
			})
		}

		// check if token values and request body value matches or not
		if ok, err := validateClaim(payload.Claims, loginInput); !ok {
			logger.LogMessage(zerolog.ErrorLevel, "controller.login_controller", "Login", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		// check if 3rd party issuer value is valid
		if !utils.ValidFroms[loginInput.From] {
			errorMessage := fmt.Sprintf(utils.InvalidThirdPartyIssuerMsg, loginInput.Token)
			logger.LogMessage(zerolog.ErrorLevel, "controller.login_controller", "Login", errorMessage)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"message": errorMessage,
			})
			return
		}

		// check if the user exists
		// if not, create the user

		// issue him a refresh and jwt token

		ctx.JSON(http.StatusCreated, "")
	}

}

func validateClaim(claims map[string]interface{}, loginInput request_entity.LoginRequest) (bool, error) {
	if claims["email"] != loginInput.Email || claims["given_name"] != loginInput.FirstName || claims["family_name"] != loginInput.LastName {
		errorMessage := fmt.Sprintf(utils.MismatchTokenAndLoginReq, claims, loginInput)
		logger.LogMessage(zerolog.ErrorLevel, "controller.login_controller", "validateClaim", errorMessage)
		return false, errors.New(errorMessage)
	}
	return true, nil
}
