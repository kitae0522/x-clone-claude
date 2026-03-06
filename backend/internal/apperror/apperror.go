package apperror

import "fmt"

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

func BadRequest(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 400, Message: fmt.Sprintf(msg, args...)}
}

func Forbidden(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 403, Message: fmt.Sprintf(msg, args...)}
}

func Unauthorized(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 401, Message: fmt.Sprintf(msg, args...)}
}

func NotFound(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 404, Message: fmt.Sprintf(msg, args...)}
}

func Conflict(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 409, Message: fmt.Sprintf(msg, args...)}
}

func Internal(msg string, args ...interface{}) *AppError {
	return &AppError{Code: 500, Message: fmt.Sprintf(msg, args...)}
}
