package apperrors

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status, Err: err}
}

var (
	ErrNotFound       = New("NOT_FOUND", "resource not found", 404)
	ErrUnauthorized   = New("UNAUTHORIZED", "unauthorized", 401)
	ErrForbidden      = New("FORBIDDEN", "forbidden", 403)
	ErrBadRequest     = New("BAD_REQUEST", "invalid request", 400)
	ErrConflict       = New("CONFLICT", "resource conflict", 409)
	ErrInternal       = New("INTERNAL_ERROR", "internal server error", 500)
	ErrExternalAPI    = New("EXTERNAL_API_ERROR", "external service error", 502)
	ErrValidation     = New("VALIDATION_ERROR", "validation failed", 422)
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
