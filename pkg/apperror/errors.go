package apperror

import (
	"fmt"
	"net/http"
)

// ErrorCode is a machine-readable string identifying the error type.
type ErrorCode string

const (
	// 4xx
	ErrValidation       ErrorCode = "VALIDATION_ERROR"
	ErrBadRequest       ErrorCode = "BAD_REQUEST"
	ErrUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrTokenExpired     ErrorCode = "TOKEN_EXPIRED"
	ErrForbidden        ErrorCode = "FORBIDDEN"
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrConflict         ErrorCode = "CONFLICT"
	ErrFileTooLarge     ErrorCode = "FILE_TOO_LARGE"
	ErrUnsupportedMedia ErrorCode = "UNSUPPORTED_MEDIA_TYPE"

	// 5xx
	ErrInternal ErrorCode = "INTERNAL_ERROR"
)

// AppError is a typed application error that carries HTTP status, error code,
// a user-facing message, and optional field-level validation details.
type AppError struct {
	HTTPStatus int
	Code       ErrorCode
	Message    string
	Details    []FieldError
	Cause      error // internal cause (not exposed to client)
}

// FieldError describes a single validation failure on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap supports errors.As / errors.Is chaining.
func (e *AppError) Unwrap() error { return e.Cause }

// ─── Constructors ────────────────────────────────────────────────────────────

func New(status int, code ErrorCode, message string) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message}
}

func NewWithCause(status int, code ErrorCode, message string, cause error) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message, Cause: cause}
}

func ValidationError(details []FieldError) *AppError {
	return &AppError{
		HTTPStatus: http.StatusBadRequest,
		Code:       ErrValidation,
		Message:    "Validation failed",
		Details:    details,
	}
}

func NotFound(resource string) *AppError {
	return New(http.StatusNotFound, ErrNotFound, fmt.Sprintf("%s not found", resource))
}

func Unauthorized(msg string) *AppError {
	return New(http.StatusUnauthorized, ErrUnauthorized, msg)
}

func TokenExpired() *AppError {
	return New(http.StatusUnauthorized, ErrTokenExpired, "Access token has expired")
}

func Forbidden() *AppError {
	return New(http.StatusForbidden, ErrForbidden, "You do not have permission to perform this action")
}

func Conflict(msg string) *AppError {
	return New(http.StatusConflict, ErrConflict, msg)
}

func FileTooLarge(maxMB int64) *AppError {
	return New(http.StatusRequestEntityTooLarge, ErrFileTooLarge,
		fmt.Sprintf("File exceeds maximum allowed size of %dMB", maxMB))
}

func UnsupportedMedia(msg string) *AppError {
	return New(http.StatusUnsupportedMediaType, ErrUnsupportedMedia, msg)
}

func Internal(cause error) *AppError {
	return NewWithCause(http.StatusInternalServerError, ErrInternal,
		"An unexpected error occurred. Please try again later.", cause)
}
