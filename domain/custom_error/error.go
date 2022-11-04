package custom_error

import (
	"errors"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func NewErrorFromError(code string, err error) *Error {
	return &Error{
		Code:    code,
		Message: err.Error(),
		Err:     err,
	}
}

func NewErrorFromMessage(code string, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     errors.New(message),
	}
}
