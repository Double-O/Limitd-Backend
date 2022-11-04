package utils

import "time"

var ValidFroms = map[string]bool{
	"Google": true,
}

const AT_EXPIRATION_TIME_NANO_SECOND = time.Second * 35
const RT_EXPIRATION_TIME_NANO_SECOND = time.Second * 60
const RT_EXPIRATION_TIME_COOKIE_SECOND = 7 * 24 * 60 * 60 //7 days
