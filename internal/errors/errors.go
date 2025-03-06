package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mcncl/snagbot/internal/logging"
)

// Error types for the application
var (
	// ErrInvalidDollarValue is returned when a dollar value is invalid
	ErrInvalidDollarValue = errors.New("invalid dollar value")

	// ErrInvalidRequest is returned for invalid requests
	ErrInvalidRequest = errors.New("invalid request")

	// ErrInvalidSignature is returned for requests with invalid signatures
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrStorageOperation is returned for storage operation failures
	ErrStorageOperation = errors.New("storage operation failed")

	// ErrSlackAPIError is returned when the Slack API returns an error
	ErrSlackAPIError = errors.New("slack API error")

	// ErrInternalServer is returned for general server errors
	ErrInternalServer = errors.New("internal server error")
)

// AppError represents an application error with context
type AppError struct {
	Err        error  // The underlying error
	Message    string // User-friendly error message
	StatusCode int    // HTTP status code (if applicable)
	Context    string // Additional context about the error
	Cause      error  // The cause of this error, if wrapping another error
}

// New creates a new AppError
func New(err error, message string) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
	}
}

// Newf creates a new AppError with formatted message
func Newf(err error, format string, args ...interface{}) *AppError {
	return &AppError{
		Err:     err,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithStatus adds an HTTP status code to the error
func (e *AppError) WithStatus(code int) *AppError {
	e.StatusCode = code
	return e
}

// WithContext adds additional context to the error
func (e *AppError) WithContext(context string) *AppError {
	e.Context = context
	return e
}

// WithCause adds a causing error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// Error implements the error interface
func (e *AppError) Error() string {
	// Generate a detailed error string including context and cause if available
	var sb strings.Builder

	// Start with the base error message
	sb.WriteString(e.Message)

	// Add context if available
	if e.Context != "" {
		sb.WriteString(" [")
		sb.WriteString(e.Context)
		sb.WriteString("]")
	}

	// Add cause if available
	if e.Cause != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Cause.Error())
	} else if e.Err != nil && e.Message != e.Err.Error() {
		sb.WriteString(": ")
		sb.WriteString(e.Err.Error())
	}

	return sb.String()
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return e.Err
}

// Is checks if the error is of the given type
func (e *AppError) Is(target error) bool {
	return errors.Is(e.Err, target) || (e.Cause != nil && errors.Is(e.Cause, target))
}

// LogAndReturn logs an error and returns it for handling by the caller
func LogAndReturn(err error) error {
	// If it's already an AppError, log it with its details
	if appErr, ok := err.(*AppError); ok {
		logging.Error("Error: %s", appErr.Error())
		return appErr
	}

	// For regular errors, wrap in an AppError and log
	logging.Error("Error: %s", err.Error())
	return New(err, err.Error())
}

// Wrap wraps an error with context without logging
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, just add the new message as context
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Err:        appErr.Err,
			Message:    message,
			StatusCode: appErr.StatusCode,
			Context:    appErr.Context,
			Cause:      appErr,
		}
	}

	// For regular errors, create a new AppError
	return &AppError{
		Err:     err,
		Message: message,
		Cause:   err,
	}
}

// WrapAndLog wraps an error with context and logs it
func WrapAndLog(err error, message string) *AppError {
	wrapped := Wrap(err, message)
	if wrapped != nil {
		logging.Error("Error: %s", wrapped.Error())
	}
	return wrapped
}

// UserFriendlyError returns a sanitized, user-friendly error message
func UserFriendlyError(err error) string {
	if err == nil {
		return ""
	}

	// If it's an AppError, use its user-friendly message
	if appErr, ok := err.(*AppError); ok {
		return appErr.Message
	}

	// For regular errors, return a generic message
	return "Something went wrong. Please try again."
}
