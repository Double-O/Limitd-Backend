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
	atClaims := jwt.MapClaims{}
	atClaims["type"] = ACCESS
	atClaims["authorized"] = true
	atClaims[TOKEN_UUID] = td.AccessUuid
	atClaims["uuid"] = userUUID
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv(ACCESS_SECRET)))

	if err != nil {
		errorMessage := fmt.Sprintf(AccessTokenGenerationFailedMsg, err.Error())
		customErr := custom_error.NewErrorFromMessage("AccessTokenGenerationFailedMsg", errorMessage)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "CreateToken", errorMessage)
		return nil, customErr
	}

	// Creating Refresh Token
	// Refresh secret should be set from env
	rtClaims := jwt.MapClaims{}
	rtClaims["type"] = REFRESH
	rtClaims[TOKEN_UUID] = td.RefreshUuid
	rtClaims["uuid"] = userUUID
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv(REFRESH_SECRET)))
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

func DeleteToken(ctx context.Context, givenUuid string, redisClient *redis.Client) *custom_error.Error {
	_, err := redisClient.Del(ctx, givenUuid).Result()
	if err != nil {
		customErr := custom_error.NewErrorFromError("RedisTokenDeleteError", err)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "DeleteToken", customErr.Message)
		return customErr
	}
	return nil
}

func VerifyToken(ctx *gin.Context, tokenString string, secret string, tokenType string) (*jwt.Token, *custom_error.Error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		errorMessage := fmt.Sprintf(FailedToParseTokenMsg, tokenType, err.Error())
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "VerifyToken", errorMessage)
		return nil, custom_error.NewErrorFromMessage("FailedToParseTokenMsg", errorMessage)
	}
	return token, nil
}

func IsTokenValid(
	ctx *gin.Context,
	redisClient *redis.Client,
	tokenString string,
	tokenType string,
	secret string) *custom_error.Error {

	if tokenType != ACCESS && tokenType != REFRESH {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", InvalidTypeOfTokenCallMsg)
		customErr := custom_error.NewErrorFromMessage("InvalidTypeOfTokenCallMsg", InvalidTypeOfTokenCallMsg)
		return customErr
	}

	token, customErr := VerifyToken(ctx, tokenString, secret, tokenType)
	if customErr != nil {
		return customErr
	}
	if !token.Valid {
		errorMessage := fmt.Sprintf(InvalidTokenMsg, tokenType)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", errorMessage)
		customErr := custom_error.NewErrorFromMessage("InvalidTokenMsg", errorMessage)
		return customErr
	}

	tokenClaims := token.Claims.(jwt.MapClaims)
	if tokenClaims["type"] != tokenType {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", InvalidTypeOfTokenCallMsg)
		customErr := custom_error.NewErrorFromMessage("InvalidTypeOfTokenCallMsg", InvalidTypeOfTokenCallMsg)
		return customErr
	}

	result, err := redisClient.Get(ctx, tokenClaims[TOKEN_UUID].(string)).Result()
	if err != nil {
		errorMessage := fmt.Sprintf(TokenUUIDNotFoundMsg, tokenType)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", errorMessage)
		customErr := custom_error.NewErrorFromMessage("TokenUUIDNotFoundMsg", errorMessage)
		return customErr
	}
	if result != tokenClaims["uuid"] {
		errorMessage := fmt.Sprintf(InvalidTokenMsg, tokenType)
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "IsTokenValid", errorMessage)
		customErr := custom_error.NewErrorFromMessage("InvalidTokenMsg", errorMessage)
		return customErr
	}

	return nil

}

func ExtractAccessTokenFromRequest(ctx *gin.Context) string {
	bearToken := ctx.GetHeader("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func ExtractRefreshTokenFromCookie(ctx *gin.Context) string {
	token, err := ctx.Cookie(REFRESH_TOKEN)
	if err != nil {
		logger.LogMessage(zerolog.ErrorLevel, "utils.token_utils", "ExtractRefreshTokenFromCookie", RefreshTokenNotFoundInCookieMsg)
		return ""
	}
	return token
}

func IsRefreshTokenValid(ctx *gin.Context, redisClient *redis.Client) *custom_error.Error {
	tokenString := ExtractRefreshTokenFromCookie(ctx)
	secret := os.Getenv(REFRESH_SECRET)
	return IsTokenValid(ctx, redisClient, tokenString, REFRESH, secret)
}

func IsAccessTokenValid(ctx *gin.Context, redisClient *redis.Client) *custom_error.Error {
	tokenString := ExtractAccessTokenFromRequest(ctx)
	secret := os.Getenv(ACCESS_SECRET)
	return IsTokenValid(ctx, redisClient, tokenString, ACCESS, secret)
}

func GetRefreshTOken(ctx *gin.Context) (*jwt.Token, *custom_error.Error) {
	refreshTokenString := ExtractRefreshTokenFromCookie(ctx)
	secret := os.Getenv(REFRESH_SECRET)
	return VerifyToken(ctx, refreshTokenString, secret, REFRESH)
}
