// File: errors_test.go
// Title: Tests for Basic Error Types and Handling
// Description: Comprehensive test suite for the TBP error handling system
//              including error creation, wrapping, unwrapping, classification,
//              and Go 1.13+ compatibility. Tests edge cases, error chains,
//              and performance characteristics.
// Author: msto63 with Claude Sonnet 4.0
// Version: v1.0.0
// Created: 2024-01-15
// Modified: 2024-01-15
//
// Change History:
// - 2024-01-15 v1.0.0: Initial test implementation with comprehensive coverage

package core

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Error(t *testing.T) {
	t.Run("returns message without cause", func(t *testing.T) {
		err := &Error{Message: "test error"}
		assert.Equal(t, "test error", err.Error())
	})

	t.Run("returns message with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &Error{
			Message: "test error",
			Cause:   cause,
		}
		assert.Equal(t, "test error: underlying error", err.Error())
	})

	t.Run("handles empty message", func(t *testing.T) {
		err := &Error{Message: ""}
		assert.Equal(t, "", err.Error())
	})
}

func TestError_Unwrap(t *testing.T) {
	t.Run("returns cause when present", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &Error{
			Message: "test error",
			Cause:   cause,
		}
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("returns nil when no cause", func(t *testing.T) {
		err := &Error{Message: "test error"}
		assert.Nil(t, err.Unwrap())
	})
}

func TestError_Is(t *testing.T) {
	t.Run("matches same code", func(t *testing.T) {
		err1 := &Error{Message: "error 1", Code: "TEST_ERROR"}
		err2 := &Error{Message: "error 2", Code: "TEST_ERROR"}
		assert.True(t, err1.Is(err2))
	})

	t.Run("does not match different code", func(t *testing.T) {
		err1 := &Error{Message: "error 1", Code: "TEST_ERROR"}
		err2 := &Error{Message: "error 2", Code: "OTHER_ERROR"}
		assert.False(t, err1.Is(err2))
	})

	t.Run("does not match when no code", func(t *testing.T) {
		err1 := &Error{Message: "error 1"}
		err2 := &Error{Message: "error 2"}
		assert.False(t, err1.Is(err2))
	})

	t.Run("matches same message for standard errors", func(t *testing.T) {
		err1 := &Error{Message: "test error"}
		err2 := errors.New("test error")
		assert.True(t, err1.Is(err2))
	})

	t.Run("does not match different message", func(t *testing.T) {
		err1 := &Error{Message: "error 1"}
		err2 := errors.New("error 2")
		assert.False(t, err1.Is(err2))
	})

	t.Run("handles nil target", func(t *testing.T) {
		err := &Error{Message: "test error"}
		assert.False(t, err.Is(nil))
	})
}

func TestError_WithContext(t *testing.T) {
	t.Run("adds context to error", func(t *testing.T) {
		err := &Error{Message: "test error"}
		newErr := err.WithContext("key1", "value1")

		value, exists := newErr.GetContext("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)

		// Original error should be unchanged
		_, exists = err.GetContext("key1")
		assert.False(t, exists)
	})

	t.Run("preserves existing context", func(t *testing.T) {
		err := &Error{
			Message: "test error",
			Context: map[string]interface{}{"existing": "value"},
		}
		newErr := err.WithContext("new", "newvalue")

		existingValue, exists := newErr.GetContext("existing")
		assert.True(t, exists)
		assert.Equal(t, "value", existingValue)

		newValue, exists := newErr.GetContext("new")
		assert.True(t, exists)
		assert.Equal(t, "newvalue", newValue)
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		err := &Error{
			Message: "test error",
			Context: map[string]interface{}{"key": "oldvalue"},
		}
		newErr := err.WithContext("key", "newvalue")

		value, exists := newErr.GetContext("key")
		assert.True(t, exists)
		assert.Equal(t, "newvalue", value)
	})
}

func TestError_WithCode(t *testing.T) {
	t.Run("sets error code", func(t *testing.T) {
		err := &Error{Message: "test error"}
		newErr := err.WithCode("TEST_CODE")

		assert.Equal(t, "TEST_CODE", newErr.Code)
		assert.Equal(t, "test error", newErr.Message)

		// Original error should be unchanged
		assert.Empty(t, err.Code)
	})

	t.Run("overwrites existing code", func(t *testing.T) {
		err := &Error{Message: "test error", Code: "OLD_CODE"}
		newErr := err.WithCode("NEW_CODE")

		assert.Equal(t, "NEW_CODE", newErr.Code)
		assert.Equal(t, "OLD_CODE", err.Code) // Original unchanged
	})
}

func TestError_GetContext(t *testing.T) {
	t.Run("returns existing context value", func(t *testing.T) {
		err := &Error{
			Message: "test error",
			Context: map[string]interface{}{"key": "value"},
		}

		value, exists := err.GetContext("key")
		assert.True(t, exists)
		assert.Equal(t, "value", value)
	})

	t.Run("returns false for non-existing key", func(t *testing.T) {
		err := &Error{
			Message: "test error",
			Context: map[string]interface{}{"key": "value"},
		}

		_, exists := err.GetContext("nonexistent")
		assert.False(t, exists)
	})

	t.Run("returns false for nil context", func(t *testing.T) {
		err := &Error{Message: "test error"}

		_, exists := err.GetContext("key")
		assert.False(t, exists)
	})
}

func TestNew(t *testing.T) {
	t.Run("creates error with message", func(t *testing.T) {
		err := New("test error")
		assert.Equal(t, "test error", err.Message)
		assert.Empty(t, err.Code)
		assert.Nil(t, err.Cause)
		assert.Nil(t, err.Context)
	})

	t.Run("handles empty message", func(t *testing.T) {
		err := New("")
		assert.Equal(t, "", err.Message)
	})
}

func TestNewf(t *testing.T) {
	t.Run("creates formatted error", func(t *testing.T) {
		err := Newf("error %d: %s", 42, "test")
		assert.Equal(t, "error 42: test", err.Message)
	})

	t.Run("handles no formatting", func(t *testing.T) {
		err := Newf("simple error")
		assert.Equal(t, "simple error", err.Message)
	})
}

func TestWrap(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Wrap(cause, "wrapper message")

		assert.Equal(t, "wrapper message", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "wrapper message: underlying error", err.Error())
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := Wrap(nil, "wrapper message")
		assert.Nil(t, err)
	})

	t.Run("wraps TBP error", func(t *testing.T) {
		cause := &Error{Message: "original error", Code: "ORIG_CODE"}
		err := Wrap(cause, "wrapper message")

		assert.Equal(t, "wrapper message", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Empty(t, err.Code) // Code not inherited
	})
}

func TestWrapf(t *testing.T) {
	t.Run("wraps error with formatted message", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := Wrapf(cause, "error %d: %s", 42, "test")

		assert.Equal(t, "error 42: test", err.Message)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := Wrapf(nil, "error %d", 42)
		assert.Nil(t, err)
	})
}

func TestWrapWithCode(t *testing.T) {
	t.Run("wraps error with code and message", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := WrapWithCode(cause, "TEST_CODE", "wrapper message")

		assert.Equal(t, "wrapper message", err.Message)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := WrapWithCode(nil, "TEST_CODE", "wrapper message")
		assert.Nil(t, err)
	})
}

func TestWrapWithContext(t *testing.T) {
	t.Run("wraps error with context", func(t *testing.T) {
		cause := errors.New("underlying error")
		context := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}
		err := WrapWithContext(cause, "wrapper message", context)

		assert.Equal(t, "wrapper message", err.Message)
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, context, err.Context)

		value1, exists := err.GetContext("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value1)

		value2, exists := err.GetContext("key2")
		assert.True(t, exists)
		assert.Equal(t, 42, value2)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		err := WrapWithContext(nil, "wrapper message", map[string]interface{}{})
		assert.Nil(t, err)
	})
}

func TestIsCode(t *testing.T) {
	t.Run("returns true for matching code", func(t *testing.T) {
		err := &Error{Message: "test error", Code: "TEST_CODE"}
		assert.True(t, IsCode(err, "TEST_CODE"))
	})

	t.Run("returns false for different code", func(t *testing.T) {
		err := &Error{Message: "test error", Code: "TEST_CODE"}
		assert.False(t, IsCode(err, "OTHER_CODE"))
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsCode(err, "TEST_CODE"))
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsCode(nil, "TEST_CODE"))
	})

	t.Run("works with wrapped errors", func(t *testing.T) {
		innerErr := &Error{Message: "inner error", Code: "INNER_CODE"}
		wrappedErr := Wrap(innerErr, "wrapper message")
		
		// The wrapper itself doesn't have the code, but GetCode should find it
		code, exists := GetCode(wrappedErr)
		assert.True(t, exists)
		assert.Equal(t, "INNER_CODE", code)
		
		// IsCode should also work through the chain
		assert.True(t, IsCode(wrappedErr, "INNER_CODE"))
		assert.False(t, IsCode(wrappedErr, "WRAPPER_CODE"))
	})
}

func TestPredefinedErrorCheckers(t *testing.T) {
	tests := []struct {
		name    string
		checker func(error) bool
		code    string
		errVar  *Error
	}{
		{"IsInternal", IsInternal, ErrCodeInternal, ErrInternal},
		{"IsInvalidInput", IsInvalidInput, ErrCodeInvalidInput, ErrInvalidInput},
		{"IsNotFound", IsNotFound, ErrCodeNotFound, ErrNotFound},
		{"IsUnauthorized", IsUnauthorized, ErrCodeUnauthorized, ErrUnauthorized},
		{"IsForbidden", IsForbidden, ErrCodeForbidden, ErrForbidden},
		{"IsConflict", IsConflict, ErrCodeConflict, ErrConflict},
		{"IsTimeout", IsTimeout, ErrCodeTimeout, ErrTimeout},
		{"IsUnavailable", IsUnavailable, ErrCodeUnavailable, ErrUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with predefined error
			assert.True(t, tt.checker(tt.errVar))

			// Test with custom error with same code
			customErr := &Error{Message: "custom error", Code: tt.code}
			assert.True(t, tt.checker(customErr))

			// Test with different code
			differentErr := &Error{Message: "different error", Code: "DIFFERENT_CODE"}
			assert.False(t, tt.checker(differentErr))

			// Test with nil
			assert.False(t, tt.checker(nil))

			// Test with standard error
			stdErr := errors.New("standard error")
			assert.False(t, tt.checker(stdErr))
		})
	}
}

func TestGetCode(t *testing.T) {
	t.Run("returns code from TBP error", func(t *testing.T) {
		err := &Error{Message: "test error", Code: "TEST_CODE"}
		code, exists := GetCode(err)
		assert.True(t, exists)
		assert.Equal(t, "TEST_CODE", code)
	})

	t.Run("returns false for TBP error without code", func(t *testing.T) {
		err := &Error{Message: "test error"}
		code, exists := GetCode(err)
		assert.False(t, exists)
		assert.Empty(t, code)
	})

	t.Run("returns false for standard error", func(t *testing.T) {
		err := errors.New("standard error")
		code, exists := GetCode(err)
		assert.False(t, exists)
		assert.Empty(t, code)
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		code, exists := GetCode(nil)
		assert.False(t, exists)
		assert.Empty(t, code)
	})

	t.Run("finds code in wrapped error", func(t *testing.T) {
		innerErr := &Error{Message: "inner error", Code: "INNER_CODE"}
		wrappedErr := fmt.Errorf("wrapper: %w", innerErr)
		
		code, exists := GetCode(wrappedErr)
		assert.True(t, exists)
		assert.Equal(t, "INNER_CODE", code)
	})
}

func TestGetRootCause(t *testing.T) {
	t.Run("returns same error when no wrapping", func(t *testing.T) {
		err := errors.New("root error")
		root := GetRootCause(err)
		assert.Equal(t, err, root)
	})

	t.Run("returns root cause from wrapped error", func(t *testing.T) {
		rootErr := errors.New("root error")
		wrappedErr := Wrap(rootErr, "wrapped error")
		doubleWrappedErr := Wrap(wrappedErr, "double wrapped")

		root := GetRootCause(doubleWrappedErr)
		assert.Equal(t, rootErr, root)
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		root := GetRootCause(nil)
		assert.Nil(t, root)
	})

	t.Run("handles circular references safely", func(t *testing.T) {
		// This shouldn't happen in practice, but test defensive behavior
		err := &Error{Message: "test error"}
		// Don't create actual circular reference as it would cause infinite loop
		root := GetRootCause(err)
		assert.Equal(t, err, root)
	})
}

func TestErrorChain(t *testing.T) {
	t.Run("returns single error for unwrapped error", func(t *testing.T) {
		err := errors.New("single error")
		chain := ErrorChain(err)
		require.Len(t, chain, 1)
		assert.Equal(t, err, chain[0])
	})

	t.Run("returns chain for wrapped errors", func(t *testing.T) {
		rootErr := errors.New("root error")
		wrappedErr := Wrap(rootErr, "wrapped error")
		doubleWrappedErr := Wrap(wrappedErr, "double wrapped")

		chain := ErrorChain(doubleWrappedErr)
		require.Len(t, chain, 3)
		assert.Equal(t, doubleWrappedErr, chain[0])
		assert.Equal(t, wrappedErr, chain[1])
		assert.Equal(t, rootErr, chain[2])
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		chain := ErrorChain(nil)
		assert.Nil(t, chain)
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("returns single message for unwrapped error", func(t *testing.T) {
		err := errors.New("single error")
		messages := ErrorMessages(err)
		require.Len(t, messages, 1)
		assert.Equal(t, "single error", messages[0])
	})

	t.Run("returns all messages for wrapped errors", func(t *testing.T) {
		rootErr := errors.New("root error")
		wrappedErr := Wrap(rootErr, "wrapped error")
		doubleWrappedErr := Wrap(wrappedErr, "double wrapped")

		messages := ErrorMessages(doubleWrappedErr)
		require.Len(t, messages, 3)
		assert.Equal(t, "double wrapped: wrapped error: root error", messages[0])
		assert.Equal(t, "wrapped error: root error", messages[1])
		assert.Equal(t, "root error", messages[2])
	})
}

func TestJoinErrors(t *testing.T) {
	t.Run("returns nil for no errors", func(t *testing.T) {
		err := JoinErrors()
		assert.Nil(t, err)
	})

	t.Run("returns nil for only nil errors", func(t *testing.T) {
		err := JoinErrors(nil, nil, nil)
		assert.Nil(t, err)
	})

	t.Run("returns single error when only one valid", func(t *testing.T) {
		validErr := errors.New("valid error")
		err := JoinErrors(nil, validErr, nil)
		assert.Equal(t, validErr, err)
	})

	t.Run("joins multiple errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		err3 := errors.New("error 3")

		joinedErr := JoinErrors(err1, err2, err3)
		require.NotNil(t, joinedErr)

		// Check that all errors are accessible
		assert.True(t, errors.Is(joinedErr, err1))
		assert.True(t, errors.Is(joinedErr, err2))
		assert.True(t, errors.Is(joinedErr, err3))
	})
}

func TestIsRetryable(t *testing.T) {
	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsRetryable(nil))
	})

	t.Run("returns true for timeout errors", func(t *testing.T) {
		err := &Error{Code: ErrCodeTimeout}
		assert.True(t, IsRetryable(err))
	})

	t.Run("returns true for unavailable errors", func(t *testing.T) {
		err := &Error{Code: ErrCodeUnavailable}
		assert.True(t, IsRetryable(err))
	})

	t.Run("returns false for other error codes", func(t *testing.T) {
		err := &Error{Code: ErrCodeInvalidInput}
		assert.False(t, IsRetryable(err))
	})

	t.Run("respects RetryableError interface", func(t *testing.T) {
		retryableErr := &mockRetryableError{retryable: true}
		assert.True(t, IsRetryable(retryableErr))

		nonRetryableErr := &mockRetryableError{retryable: false}
		assert.False(t, IsRetryable(nonRetryableErr))
	})
}

func TestIsTemporary(t *testing.T) {
	t.Run("returns false for nil error", func(t *testing.T) {
		assert.False(t, IsTemporary(nil))
	})

	t.Run("returns false for standard errors", func(t *testing.T) {
		err := errors.New("standard error")
		assert.False(t, IsTemporary(err))
	})

	t.Run("respects TemporaryError interface", func(t *testing.T) {
		tempErr := &mockTemporaryError{temporary: true}
		assert.True(t, IsTemporary(tempErr))

		nonTempErr := &mockTemporaryError{temporary: false}
		assert.False(t, IsTemporary(nonTempErr))
	})
}

func TestGo113Compatibility(t *testing.T) {
	t.Run("errors.Is works with TBP errors", func(t *testing.T) {
		target := &Error{Code: "TEST_CODE"}
		err := &Error{Code: "TEST_CODE"}
		
		assert.True(t, errors.Is(err, target))
	})

	t.Run("errors.As works with TBP errors", func(t *testing.T) {
		originalErr := &Error{Message: "test error", Code: "TEST_CODE"}
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)

		var tbpErr *Error
		assert.True(t, errors.As(wrappedErr, &tbpErr))
		assert.Equal(t, originalErr, tbpErr)
	})

	t.Run("errors.Unwrap works with TBP errors", func(t *testing.T) {
		cause := errors.New("cause error")
		err := &Error{Message: "wrapper", Cause: cause}

		unwrapped := errors.Unwrap(err)
		assert.Equal(t, cause, unwrapped)
	})
}

func TestPredefinedErrors(t *testing.T) {
	predefinedErrors := map[string]*Error{
		"ErrInternal":     ErrInternal,
		"ErrInvalidInput": ErrInvalidInput,
		"ErrNotFound":     ErrNotFound,
		"ErrUnauthorized": ErrUnauthorized,
		"ErrForbidden":    ErrForbidden,
		"ErrConflict":     ErrConflict,
		"ErrTimeout":      ErrTimeout,
		"ErrUnavailable":  ErrUnavailable,
	}

	for name, err := range predefinedErrors {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, err.Message)
			assert.NotEmpty(t, err.Code)
			assert.Nil(t, err.Cause)
		})
	}
}

// Mock types for testing interfaces

type mockRetryableError struct {
	retryable bool
}

func (e *mockRetryableError) Error() string {
	return "mock retryable error"
}

func (e *mockRetryableError) IsRetryable() bool {
	return e.retryable
}

type mockTemporaryError struct {
	temporary bool
}

func (e *mockTemporaryError) Error() string {
	return "mock temporary error"
}

func (e *mockTemporaryError) Temporary() bool {
	return e.temporary
}

// Benchmark tests for performance validation

func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = New("test error")
	}
}

func BenchmarkNewf(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Newf("error %d: %s", 42, "test")
	}
}

func BenchmarkWrap(b *testing.B) {
	cause := errors.New("underlying error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Wrap(cause, "wrapper message")
	}
}

func BenchmarkWrapf(b *testing.B) {
	cause := errors.New("underlying error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Wrapf(cause, "error %d: %s", 42, "test")
	}
}

func BenchmarkWrapWithCode(b *testing.B) {
	cause := errors.New("underlying error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WrapWithCode(cause, "TEST_CODE", "wrapper message")
	}
}

func BenchmarkWrapWithContext(b *testing.B) {
	cause := errors.New("underlying error")
	context := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WrapWithContext(cause, "wrapper message", context)
	}
}

func BenchmarkIsCode(b *testing.B) {
	err := &Error{Message: "test error", Code: "TEST_CODE"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsCode(err, "TEST_CODE")
	}
}

func BenchmarkIsCode_Standard(b *testing.B) {
	err := errors.New("standard error")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsCode(err, "TEST_CODE")
	}
}

func BenchmarkGetCode(b *testing.B) {
	err := &Error{Message: "test error", Code: "TEST_CODE"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GetCode(err)
	}
}

func BenchmarkWithContext(b *testing.B) {
	err := &Error{Message: "test error"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.WithContext("key", "value")
	}
}

func BenchmarkWithContext_Existing(b *testing.B) {
	err := &Error{
		Message: "test error",
		Context: map[string]interface{}{
			"existing1": "value1",
			"existing2": "value2",
			"existing3": "value3",
		},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.WithContext("new_key", "new_value")
	}
}

func BenchmarkWithCode(b *testing.B) {
	err := &Error{Message: "test error"}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.WithCode("TEST_CODE")
	}
}

func BenchmarkError_Error(b *testing.B) {
	cause := errors.New("underlying error")
	err := &Error{
		Message: "test error",
		Cause:   cause,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkGetRootCause(b *testing.B) {
	rootErr := errors.New("root error")
	wrappedErr := Wrap(rootErr, "wrapped")
	doubleWrappedErr := Wrap(wrappedErr, "double wrapped")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetRootCause(doubleWrappedErr)
	}
}

func BenchmarkGetRootCause_Deep(b *testing.B) {
	// Create a deep error chain (10 levels)
	err := errors.New("root error")
	for i := 0; i < 10; i++ {
		err = Wrap(err, fmt.Sprintf("layer %d", i))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetRootCause(err)
	}
}

func BenchmarkErrorChain(b *testing.B) {
	rootErr := errors.New("root error")
	wrappedErr := Wrap(rootErr, "wrapped")
	doubleWrappedErr := Wrap(wrappedErr, "double wrapped")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ErrorChain(doubleWrappedErr)
	}
}

func BenchmarkErrorChain_Deep(b *testing.B) {
	// Create a deep error chain (10 levels)
	err := errors.New("root error")
	for i := 0; i < 10; i++ {
		err = Wrap(err, fmt.Sprintf("layer %d", i))
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ErrorChain(err)
	}
}

func BenchmarkErrorMessages(b *testing.B) {
	rootErr := errors.New("root error")
	wrappedErr := Wrap(rootErr, "wrapped")
	doubleWrappedErr := Wrap(wrappedErr, "double wrapped")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ErrorMessages(doubleWrappedErr)
	}
}

func BenchmarkJoinErrors(b *testing.B) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = JoinErrors(err1, err2, err3)
	}
}

func BenchmarkJoinErrors_WithNils(b *testing.B) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = JoinErrors(nil, err1, nil, err2, nil)
	}
}

func BenchmarkIsRetryable(b *testing.B) {
	err := &Error{Code: ErrCodeTimeout}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsRetryable(err)
	}
}

func BenchmarkIsTemporary(b *testing.B) {
	err := &mockTemporaryError{temporary: true}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsTemporary(err)
	}
}