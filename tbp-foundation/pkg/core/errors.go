// File: errors.go
// Title: Basic Error Types and Handling for TBP Core
// Description: Provides fundamental error types and basic error handling
//              functionality that serves as the foundation for the more
//              comprehensive error system in the errors package.
//              Implements Go 1.13+ error wrapping with TBP-specific extensions.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial implementation with basic error types and wrapping

package core

import (
	"errors"
	"fmt"
)

// Error represents a basic TBP error with additional context.
// This is the foundation error type that other packages can extend.
type Error struct {
	// Message is the human-readable error message
	Message string `json:"message"`
	
	// Code is a machine-readable error identifier
	Code string `json:"code,omitempty"`
	
	// Cause is the underlying error that caused this error
	Cause error `json:"-"`
	
	// Context provides additional key-value pairs for debugging
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface.
// Returns the error message, optionally with the underlying cause.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap implements the Go 1.13+ error unwrapping interface.
// This allows errors.Is() and errors.As() to work with wrapped errors.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is implements the Go 1.13+ error comparison interface.
// This allows errors.Is() to work with TBP errors.
func (e *Error) Is(target error) bool {
	if target == nil {
		return false
	}
	
	// Check if target is also a TBP Error with the same code
	if tbpErr, ok := target.(*Error); ok {
		return e.Code != "" && e.Code == tbpErr.Code
	}
	
	// Use standard error comparison
	return e.Message == target.Error()
}

// WithContext adds context information to the error.
// Returns a new error with the additional context.
func (e *Error) WithContext(key string, value interface{}) *Error {
	newErr := &Error{
		Message: e.Message,
		Code:    e.Code,
		Cause:   e.Cause,
		Context: make(map[string]interface{}),
	}
	
	// Copy existing context
	for k, v := range e.Context {
		newErr.Context[k] = v
	}
	
	// Add new context
	newErr.Context[key] = value
	
	return newErr
}

// WithCode sets the error code.
// Returns a new error with the specified code.
func (e *Error) WithCode(code string) *Error {
	return &Error{
		Message: e.Message,
		Code:    code,
		Cause:   e.Cause,
		Context: e.Context,
	}
}

// GetContext retrieves a context value by key.
// Returns the value and true if found, nil and false otherwise.
func (e *Error) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// Common error codes used throughout TBP.
// These provide a standardized set of error classifications.
const (
	// ErrCodeInternal represents an internal system error
	ErrCodeInternal = "INTERNAL_ERROR"
	
	// ErrCodeInvalidInput represents invalid user input
	ErrCodeInvalidInput = "INVALID_INPUT"
	
	// ErrCodeNotFound represents a resource that could not be found
	ErrCodeNotFound = "NOT_FOUND"
	
	// ErrCodeUnauthorized represents an authentication failure
	ErrCodeUnauthorized = "UNAUTHORIZED"
	
	// ErrCodeForbidden represents an authorization failure
	ErrCodeForbidden = "FORBIDDEN"
	
	// ErrCodeConflict represents a resource conflict
	ErrCodeConflict = "CONFLICT"
	
	// ErrCodeTimeout represents a timeout error
	ErrCodeTimeout = "TIMEOUT"
	
	// ErrCodeUnavailable represents a service unavailability
	ErrCodeUnavailable = "UNAVAILABLE"
)

// Predefined error instances for common scenarios.
// These can be used directly or as base errors for wrapping.
var (
	// ErrInternal represents a generic internal error
	ErrInternal = &Error{
		Message: "internal server error",
		Code:    ErrCodeInternal,
	}
	
	// ErrInvalidInput represents invalid user input
	ErrInvalidInput = &Error{
		Message: "invalid input provided",
		Code:    ErrCodeInvalidInput,
	}
	
	// ErrNotFound represents a resource not found
	ErrNotFound = &Error{
		Message: "resource not found",
		Code:    ErrCodeNotFound,
	}
	
	// ErrUnauthorized represents an authentication failure
	ErrUnauthorized = &Error{
		Message: "authentication required",
		Code:    ErrCodeUnauthorized,
	}
	
	// ErrForbidden represents an authorization failure
	ErrForbidden = &Error{
		Message: "access forbidden",
		Code:    ErrCodeForbidden,
	}
	
	// ErrConflict represents a resource conflict
	ErrConflict = &Error{
		Message: "resource conflict",
		Code:    ErrCodeConflict,
	}
	
	// ErrTimeout represents a timeout error
	ErrTimeout = &Error{
		Message: "operation timed out",
		Code:    ErrCodeTimeout,
	}
	
	// ErrUnavailable represents service unavailability
	ErrUnavailable = &Error{
		Message: "service unavailable",
		Code:    ErrCodeUnavailable,
	}
)

// New creates a new TBP error with the given message.
// This is similar to errors.New() but creates a TBP Error instance.
func New(message string) *Error {
	return &Error{
		Message: message,
	}
}

// Newf creates a new TBP error with a formatted message.
// This is similar to fmt.Errorf() but creates a TBP Error instance.
func Newf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context.
// If the provided error is nil, returns nil.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	
	return &Error{
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with a formatted message.
// If the provided error is nil, returns nil.
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// WrapWithCode wraps an existing error with a message and error code.
// If the provided error is nil, returns nil.
func WrapWithCode(err error, code, message string) *Error {
	if err == nil {
		return nil
	}
	
	return &Error{
		Message: message,
		Code:    code,
		Cause:   err,
	}
}

// WrapWithContext wraps an existing error with additional context.
// If the provided error is nil, returns nil.
func WrapWithContext(err error, message string, context map[string]interface{}) *Error {
	if err == nil {
		return nil
	}
	
	return &Error{
		Message: message,
		Cause:   err,
		Context: context,
	}
}

// IsCode checks if an error has a specific error code.
// Works with both TBP errors and standard errors.
func IsCode(err error, code string) bool {
	if err == nil {
		return false
	}
	
	// Walk through the error chain to find any error with the specified code
	current := err
	for current != nil {
		// Check if current error is a TBP error with the specified code
		if tbpErr, ok := current.(*Error); ok && tbpErr.Code == code {
			return true
		}
		
		// Try to get the next error in the chain
		var next error
		
		// First try TBP Error's Cause field
		if tbpErr, ok := current.(*Error); ok && tbpErr.Cause != nil {
			next = tbpErr.Cause
		} else if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			// Then try standard Unwrap interface
			next = unwrapper.Unwrap()
		}
		
		if next == nil || next == current {
			break // Avoid infinite loops
		}
		current = next
	}
	
	return false
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return IsCode(err, ErrCodeInternal)
}

// IsInvalidInput checks if an error is an invalid input error.
func IsInvalidInput(err error) bool {
	return IsCode(err, ErrCodeInvalidInput)
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return IsCode(err, ErrCodeNotFound)
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return IsCode(err, ErrCodeUnauthorized)
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return IsCode(err, ErrCodeForbidden)
}

// IsConflict checks if an error is a conflict error.
func IsConflict(err error) bool {
	return IsCode(err, ErrCodeConflict)
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	return IsCode(err, ErrCodeTimeout)
}

// IsUnavailable checks if an error is an unavailable error.
func IsUnavailable(err error) bool {
	return IsCode(err, ErrCodeUnavailable)
}

// GetCode extracts the error code from an error.
// Returns the code and true if found, empty string and false otherwise.
func GetCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	
	// Walk through the error chain to find the first error with a code
	current := err
	for current != nil {
		// Check if current error is a TBP error with a code
		if tbpErr, ok := current.(*Error); ok && tbpErr.Code != "" {
			return tbpErr.Code, true
		}
		
		// Try to get the next error in the chain
		var next error
		
		// First try TBP Error's Cause field
		if tbpErr, ok := current.(*Error); ok && tbpErr.Cause != nil {
			next = tbpErr.Cause
		} else if unwrapper, ok := current.(interface{ Unwrap() error }); ok {
			// Then try standard Unwrap interface
			next = unwrapper.Unwrap()
		}
		
		if next == nil || next == current {
			break // Avoid infinite loops
		}
		current = next
	}
	
	return "", false
}

// GetRootCause returns the root cause of an error by unwrapping all layers.
// If the error doesn't wrap other errors, returns the error itself.
func GetRootCause(err error) error {
	if err == nil {
		return nil
	}
	
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// ErrorChain returns all errors in the error chain as a slice.
// The first element is the outermost error, the last is the root cause.
func ErrorChain(err error) []error {
	if err == nil {
		return nil
	}
	
	var chain []error
	current := err
	
	for current != nil {
		chain = append(chain, current)
		current = errors.Unwrap(current)
	}
	
	return chain
}

// ErrorMessages returns all error messages in the error chain.
// Useful for detailed error logging and debugging.
func ErrorMessages(err error) []string {
	chain := ErrorChain(err)
	messages := make([]string, len(chain))
	
	for i, e := range chain {
		messages[i] = e.Error()
	}
	
	return messages
}

// JoinErrors combines multiple errors into a single error.
// Uses Go 1.20+ errors.Join if available, otherwise creates a TBP error.
func JoinErrors(errs ...error) error {
	// Filter out nil errors
	var validErrors []error
	for _, err := range errs {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}
	
	if len(validErrors) == 0 {
		return nil
	}
	
	if len(validErrors) == 1 {
		return validErrors[0]
	}
	
	// Use Go 1.20+ errors.Join if available
	// Note: This would require Go 1.20+, for earlier versions we'd implement our own
	return errors.Join(validErrors...)
}

// RetryableError indicates whether an error might succeed if retried.
// This is a basic implementation that can be extended by the errors package.
type RetryableError interface {
	error
	IsRetryable() bool
}

// IsRetryable checks if an error might succeed if retried.
// Returns true for timeout and unavailable errors by default.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if the error implements RetryableError interface
	if retryable, ok := err.(RetryableError); ok {
		return retryable.IsRetryable()
	}
	
	// Default retry logic for known error types
	return IsTimeout(err) || IsUnavailable(err)
}

// TemporaryError indicates whether an error is temporary.
// This interface is compatible with net.Error.
type TemporaryError interface {
	error
	Temporary() bool
}

// IsTemporary checks if an error is temporary.
// Returns true for errors that implement the Temporary interface.
func IsTemporary(err error) bool {
	if err == nil {
		return false
	}
	
	if temp, ok := err.(TemporaryError); ok {
		return temp.Temporary()
	}
	
	return false
}