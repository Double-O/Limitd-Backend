package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-redis/redis/v9"

	"github.com/Double-O/Limitd-Backend/domain/custom_error"

	"github.com/Double-O/Limitd-Backend/logger"
	"github.com/rs/zerolog"

	"github.com/google/uuid"

	"github.com/Double-O/Limitd-Backend/domain/entity"
	"github.com/dgrijalva/jwt-go"
)

func CreateToken(userUUID uuid.UUID) (*entity.TokenDetails, *custom_error.Error) {
	td := &entity.TokenDetails{}

	td.AtExpires = time.Now().Add(AT_EXPIRATION_TIME_NANO_SECOND).Unix()
	td.AccessUuid = uuid.New().String()

	td.RtExpires = time.Now().Add(RT_EXPIRATION_TIME_NANO_SECOND).Unix()
	td.RefreshUuid = uuid.New().String()

	var err error
	// Creating Access Token
	// ACESS_SECRET should be set from env
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["uuid"] = userUUID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		errorMessage := fmt.Sprintf(AccessTokenGenerationFailedMsg, err.Error())
		customErr := custom_error.NewErrorFromMessage("AccessTokenGenerationFailedMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "CreateToken", errorMessage)
		return nil, customErr
	}

	// Creating Refresh Token
	// Refresh secret should be set from env
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["uuid"] = userUUID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		errorMessage := fmt.Sprintf(RefreshTokenGenerationFailedMsg, err.Error())
		customErr := custom_error.NewErrorFromMessage("RefreshTokenGenerationFailedMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "CreateToken", errorMessage)
		return nil, customErr
	}
	return td, nil
}

func SaveTokenInRedis(ctx context.Context, redisClient *redis.Client, userUUID uuid.UUID, td *entity.TokenDetails) *custom_error.Error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	err := redisClient.Set(ctx, td.AccessUuid, userUUID.String(), at.Sub(now)).Err()
	if err != nil {
		customErr := custom_error.NewErrorFromError("RedisAccessTokenSaveError", err)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "SaveTokenInRedis", customErr.Message)
		return customErr
	}
	err = redisClient.Set(ctx, td.RefreshUuid, userUUID.String(), rt.Sub(now)).Err()
	if err != nil {
		customErr := custom_error.NewErrorFromError("RedisRefreshTokenSaveError", err)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "SaveTokenInRedis", customErr.Message)
		return customErr
	}
	return nil
}

func ExtractTokenFromRequest(ctx *gin.Context) string {
	bearToken := ctx.GetHeader("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyAccessToken(ctx *gin.Context) (*jwt.Token, *custom_error.Error) {
	tokenString := ExtractTokenFromRequest(ctx)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		errorMessage := fmt.Sprintf(FailedToParseAccessTokenMsg, err.Error())
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "VerifyAccessToken", errorMessage)
		return nil, custom_error.NewErrorFromMessage("FailedToParseAccessTokenMsg", errorMessage)
	}
	return token, nil
}

func IsTokenValid(ctx *gin.Context, client *redis.Client) *custom_error.Error {
	token, customErr := VerifyAccessToken(ctx)
	if customErr != nil {
		return customErr
	}
	if !token.Valid {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", InvalidAccessTokenMsg)
		customErr := custom_error.NewErrorFromMessage("InvalidAccessTokenMsg", InvalidAccessTokenMsg)
		return customErr
	}
	atClaims := token.Claims.(jwt.MapClaims)

	result, err := client.Get(ctx, atClaims["access_uuid"].(string)).Result()
	if err != nil {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", AccessTokenNotFoundMsg)
		customErr := custom_error.NewErrorFromMessage("AccessTokenNotFoundMsg", AccessTokenNotFoundMsg)
		return customErr
	}
	if result != atClaims["uuid"] {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", InvalidAccessTokenMsg)
		customErr := custom_error.NewErrorFromMessage("InvalidAccessTokenMsg", InvalidAccessTokenMsg)
		return customErr
	}

	return nil
}
