package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents a standardized error code.
type ErrorCode string

const (
	// General error codes
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// Database error codes
	ErrCodeDatabaseError    ErrorCode = "DATABASE_ERROR"
	ErrCodeUniqueViolation  ErrorCode = "UNIQUE_VIOLATION"
	ErrCodeForeignKeyViolation ErrorCode = "FOREIGN_KEY_VIOLATION"

	// Validation error codes
	ErrCodeValidationError ErrorCode = "VALIDATION_ERROR"
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
)

// AppError represents an application error with additional context.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
	Details    any       `json:"details,omitempty"`
	Original   error     `json:"-"`
}

// Error returns the error message.
func (e *AppError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Original.Error())
	}
	return e.Message
}

// Unwrap returns the original error for error wrapping.
func (e *AppError) Unwrap() error {
	return e.Original
}

// New creates a new AppError.
func New(code ErrorCode, message string, httpStatus int, details any, original error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    details,
		Original:   original,
	}
}

// NewInternalServerError creates a new internal server error.
func NewInternalServerError(message string, original error) *AppError {
	return New(ErrCodeInternalError, message, http.StatusInternalServerError, nil, original)
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string, details any) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest, details, nil)
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound, nil, nil)
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized, nil, nil)
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden, nil, nil)
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict, nil, nil)
}

// NewDatabaseError creates a new database error.
func NewDatabaseError(message string, original error) *AppError {
	return New(ErrCodeDatabaseError, message, http.StatusInternalServerError, nil, original)
}

// NewValidationError creates a new validation error.
func NewValidationError(message string, details any) *AppError {
	return New(ErrCodeValidationError, message, http.StatusBadRequest, details, nil)
}

// NewInvalidInputError creates a new invalid input error.
func NewInvalidInputError(message string, details any) *AppError {
	return New(ErrCodeInvalidInput, message, http.StatusBadRequest, details, nil)
}

// Is checks if the error matches a given error code.
func (e *AppError) Is(code ErrorCode) bool {
	return e.Code == code
}

// Wrap wraps an existing error with an AppError.
func Wrap(code ErrorCode, message string, httpStatus int, original error) *AppError {
	return New(code, message, httpStatus, nil, original)
}

// WrapIfNotAppError wraps an error only if it's not already an AppError.
func WrapIfNotAppError(err error, code ErrorCode, message string, httpStatus int) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return Wrap(code, message, httpStatus, err)
}
