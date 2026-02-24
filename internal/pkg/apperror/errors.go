package apperror

import (
	"errors"
	"net/http"
)

// Sentinel errors for cross-layer error matching.
var (
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("resource conflict")
	ErrValidation    = errors.New("validation failed")
	ErrUnprocessable = errors.New("unprocessable entity")
	ErrRateLimited       = errors.New("rate limited")
	ErrVersionConflict   = errors.New("version conflict")
	ErrInternal          = errors.New("internal server error")
)

// AppError is the unified application error type.
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatusCode maps sentinel errors to HTTP status codes.
func HTTPStatusCode(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrValidation):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrUnprocessable):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, ErrVersionConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// Convenience constructors.

func NotFound(msg string, err error) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: msg, Err: errors.Join(ErrNotFound, err)}
}

func Unauthorized(msg string, err error) *AppError {
	return &AppError{Code: http.StatusUnauthorized, Message: msg, Err: errors.Join(ErrUnauthorized, err)}
}

func Forbidden(msg string, err error) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: msg, Err: errors.Join(ErrForbidden, err)}
}

func Conflict(msg string, err error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: errors.Join(ErrConflict, err)}
}

func Validation(msg string, err error) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Message: msg, Err: errors.Join(ErrValidation, err)}
}

func VersionConflict(msg string, err error) *AppError {
	return &AppError{Code: http.StatusConflict, Message: msg, Err: errors.Join(ErrVersionConflict, err)}
}

func Internal(msg string, err error) *AppError {
	return &AppError{Code: http.StatusInternalServerError, Message: msg, Err: errors.Join(ErrInternal, err)}
}
