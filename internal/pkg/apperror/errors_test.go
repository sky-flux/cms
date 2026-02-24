package apperror

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error_WithInnerErr(t *testing.T) {
	inner := errors.New("db timeout")
	appErr := &AppError{Code: 500, Message: "query failed", Err: inner}
	assert.Equal(t, "query failed: db timeout", appErr.Error())
}

func TestAppError_Error_NilInnerErr(t *testing.T) {
	appErr := &AppError{Code: 404, Message: "not found"}
	assert.Equal(t, "not found", appErr.Error())
}

func TestAppError_Unwrap(t *testing.T) {
	inner := errors.New("original")
	appErr := &AppError{Code: 500, Message: "wrapped", Err: inner}

	assert.True(t, errors.Is(appErr, inner))

	var target *AppError
	assert.True(t, errors.As(appErr, &target))
	assert.Equal(t, 500, target.Code)
}

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"ErrNotFound", ErrNotFound, http.StatusNotFound},
		{"ErrUnauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"ErrForbidden", ErrForbidden, http.StatusForbidden},
		{"ErrConflict", ErrConflict, http.StatusConflict},
		{"ErrValidation", ErrValidation, http.StatusUnprocessableEntity},
		{"ErrUnprocessable", ErrUnprocessable, http.StatusUnprocessableEntity},
		{"ErrRateLimited", ErrRateLimited, http.StatusTooManyRequests},
		{"ErrInternal", ErrInternal, http.StatusInternalServerError},
		{"unknown error", errors.New("unknown"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HTTPStatusCode(tt.err))
		})
	}
}

func TestHTTPStatusCode_WrappedSentinel(t *testing.T) {
	appErr := NotFound("user not found", errors.New("sql: no rows"))
	assert.Equal(t, http.StatusNotFound, HTTPStatusCode(appErr))
}

func TestConstructors(t *testing.T) {
	originalErr := errors.New("original cause")

	tests := []struct {
		name         string
		constructor  func(string, error) *AppError
		expectedCode int
		sentinel     error
	}{
		{"NotFound", NotFound, http.StatusNotFound, ErrNotFound},
		{"Unauthorized", Unauthorized, http.StatusUnauthorized, ErrUnauthorized},
		{"Forbidden", Forbidden, http.StatusForbidden, ErrForbidden},
		{"Conflict", Conflict, http.StatusConflict, ErrConflict},
		{"Validation", Validation, http.StatusUnprocessableEntity, ErrValidation},
		{"Internal", Internal, http.StatusInternalServerError, ErrInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor("test message", originalErr)

			assert.Equal(t, tt.expectedCode, appErr.Code)
			assert.Equal(t, "test message", appErr.Message)

			require.True(t, errors.Is(appErr, tt.sentinel),
				"AppError should match sentinel %v", tt.sentinel)

			require.True(t, errors.Is(appErr, originalErr),
				"AppError should match original error")
		})
	}
}

func TestConstructors_NilInnerErr(t *testing.T) {
	appErr := NotFound("not found", nil)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.True(t, errors.Is(appErr, ErrNotFound))
}
