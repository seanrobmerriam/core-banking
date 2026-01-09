package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := New(ErrCodeBadRequest, "bad request occurred", http.StatusBadRequest, map[string]string{"field": "email"}, originalErr)

	t.Run("Error method", func(t *testing.T) {
		errMsg := appErr.Error()
		assert.Contains(t, errMsg, "bad request occurred")
		assert.Contains(t, errMsg, "original error")
	})

	t.Run("Unwrap method", func(t *testing.T) {
		unwrapped := errors.Unwrap(appErr)
		assert.Equal(t, originalErr, unwrapped)
	})

	t.Run("Is method", func(t *testing.T) {
		assert.True(t, appErr.Is(ErrCodeBadRequest))
		assert.False(t, appErr.Is(ErrCodeNotFound))
	})

	t.Run("JSON marshaling", func(t *testing.T) {
		assert.Equal(t, ErrCodeBadRequest, appErr.Code)
		assert.Equal(t, "bad request occurred", appErr.Message)
		assert.Equal(t, http.StatusBadRequest, appErr.HTTPStatus)
		assert.NotNil(t, appErr.Details)
	})
}

func TestNewErrorTypes(t *testing.T) {
	tests := []struct {
		name         string
		errFunc      func(string) *AppError
		expectedCode ErrorCode
		expectedHTTP int
		input        string
	}{
		{
			name:         "NewInternalServerError",
			errFunc:      func(msg string) *AppError { return NewInternalServerError(msg, nil) },
			expectedCode: ErrCodeInternalError,
			expectedHTTP: http.StatusInternalServerError,
			input:        "internal error",
		},
		{
			name:         "NewBadRequestError",
			errFunc:      func(msg string) *AppError { return NewBadRequestError(msg, nil) },
			expectedCode: ErrCodeBadRequest,
			expectedHTTP: http.StatusBadRequest,
			input:        "bad request",
		},
		{
			name:         "NewNotFoundError",
			errFunc:      func(msg string) *AppError { return NewNotFoundError(msg) },
			expectedCode: ErrCodeNotFound,
			expectedHTTP: http.StatusNotFound,
			input:        "resource not found",
		},
		{
			name:         "NewUnauthorizedError",
			errFunc:      func(msg string) *AppError { return NewUnauthorizedError(msg) },
			expectedCode: ErrCodeUnauthorized,
			expectedHTTP: http.StatusUnauthorized,
			input:        "unauthorized",
		},
		{
			name:         "NewForbiddenError",
			errFunc:      func(msg string) *AppError { return NewForbiddenError(msg) },
			expectedCode: ErrCodeForbidden,
			expectedHTTP: http.StatusForbidden,
			input:        "forbidden",
		},
		{
			name:         "NewConflictError",
			errFunc:      func(msg string) *AppError { return NewConflictError(msg) },
			expectedCode: ErrCodeConflict,
			expectedHTTP: http.StatusConflict,
			input:        "conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.input)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedHTTP, err.HTTPStatus)
			assert.Contains(t, err.Message, tt.input)
		})
	}
}

func TestNewDatabaseError(t *testing.T) {
	original := errors.New("connection failed")
	err := NewDatabaseError("database operation failed", original)

	assert.Equal(t, ErrCodeDatabaseError, err.Code)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	assert.Equal(t, "database operation failed: connection failed", err.Error())
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{"email": "invalid format"}
	err := NewValidationError("validation failed", details)

	assert.Equal(t, ErrCodeValidationError, err.Code)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, details, err.Details)
}

func TestWrap(t *testing.T) {
	original := errors.New("original")
	wrapped := Wrap(ErrCodeBadRequest, "wrapped message", http.StatusBadRequest, original)

	assert.Equal(t, ErrCodeBadRequest, wrapped.Code)
	assert.Equal(t, "wrapped message: original", wrapped.Error())
}

func TestWrapIfNotAppError(t *testing.T) {
	t.Run("wraps regular error", func(t *testing.T) {
		original := errors.New("some error")
		wrapped := WrapIfNotAppError(original, ErrCodeBadRequest, "wrapped", http.StatusBadRequest)

		var appErr *AppError
		require.ErrorAs(t, wrapped, &appErr)
		assert.Equal(t, ErrCodeBadRequest, appErr.Code)
	})

	t.Run("does not wrap AppError", func(t *testing.T) {
		original := NewBadRequestError("original", nil)
		wrapped := WrapIfNotAppError(original, ErrCodeBadRequest, "wrapped", http.StatusBadRequest)

		var appErr *AppError
		require.ErrorAs(t, wrapped, &appErr)
		assert.Equal(t, "original", appErr.Message)
	})
}
