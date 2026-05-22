package apperr

import "errors"

const (
	CodeAlreadyExists = "ALREADY_EXISTS"
	CodeNotFound      = "NOT_FOUND"
	CodeInternalError = "INTERNAL_ERROR"
	CodeInvalidInput  = "INVALID_INPUT"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}