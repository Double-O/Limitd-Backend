package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func LogMessage(level zerolog.Level, fileName string, functionName string, message string) {
	log.WithLevel(level).
		Str("file", fileName).
		Str("function", functionName).
		Msg(message)
}
