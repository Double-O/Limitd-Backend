package utils

import "time"

var ValidFroms = map[string]bool{
	"Google": true,
}

const AT_EXPIRATION_TIME_NANO_SECOND = time.Second * 35
const RT_EXPIRATION_TIME_NANO_SECOND = time.Second * 90
const RT_EXPIRATION_TIME_COOKIE_SECOND = 60 * 60 //1 hour

const ACCESS = "access"
const REFRESH = "refresh"
const ACCESS_SECRET = "ACCESS_SECRET"
const REFRESH_SECRET = "REFRESH_SECRET"

const REFRESH_TOKEN = "Refresh_Token"
const ACCESS_TOKEN_UUID = "Access_Token_UUID"
const TOKEN_UUID = "token_uuid"

const USER = "user"
