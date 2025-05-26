// File: context_test.go
// Title: Tests for Extended Context Management
// Description: Comprehensive test suite for the extended context functionality
//              including user tracking, tenant management, request tracing,
//              and all context manipulation functions. Tests edge cases,
//              concurrent access, and performance characteristics.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation with comprehensive coverage

package core

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithUser(t *testing.T) {
	t.Run("adds user to context", func(t *testing.T) {
		ctx := context.Background()
		user := &UserInfo{
			ID:       "user123",
			Username: "testuser",
			Email:    "test@example.com",
			Roles:    []string{"admin", "user"},
			TenantID: "tenant123",
			LoginAt:  time.Now(),
		}

		newCtx := WithUser(ctx, user)

		retrievedUser, exists := GetUser(newCtx)
		assert.True(t, exists)
		assert.Equal(t, user, retrievedUser)
	})

	t.Run("handles nil user", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithUser(ctx, nil)

		_, exists := GetUser(newCtx)
		assert.False(t, exists)
	})

	t.Run("overwrites existing user", func(t *testing.T) {
		ctx := context.Background()
		user1 := &UserInfo{ID: "user1"}
		user2 := &UserInfo{ID: "user2"}

		ctx = WithUser(ctx, user1)
		ctx = WithUser(ctx, user2)

		retrievedUser, exists := GetUser(ctx)
		assert.True(t, exists)
		assert.Equal(t, "user2", retrievedUser.ID)
	})
}

func TestWithUserID(t *testing.T) {
	t.Run("creates user with ID", func(t *testing.T) {
		ctx := context.Background()
		userID := "user123"

		newCtx := WithUserID(ctx, userID)

		retrievedUser, exists := GetUser(newCtx)
		assert.True(t, exists)
		assert.Equal(t, userID, retrievedUser.ID)
		assert.Empty(t, retrievedUser.Username)
		assert.Empty(t, retrievedUser.Email)
	})

	t.Run("handles empty user ID", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithUserID(ctx, "")

		_, exists := GetUser(newCtx)
		assert.False(t, exists)
	})
}

func TestWithTenant(t *testing.T) {
	t.Run("adds tenant to context", func(t *testing.T) {
		ctx := context.Background()
		tenant := &TenantInfo{
			ID:       "tenant123",
			Name:     "Test Tenant",
			Domain:   "test.example.com",
			IsActive: true,
			Settings: map[string]string{"theme": "dark"},
		}

		newCtx := WithTenant(ctx, tenant)

		retrievedTenant, exists := GetTenant(newCtx)
		assert.True(t, exists)
		assert.Equal(t, tenant, retrievedTenant)
	})

	t.Run("handles nil tenant", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithTenant(ctx, nil)

		_, exists := GetTenant(newCtx)
		assert.False(t, exists)
	})
}

func TestWithTenantID(t *testing.T) {
	t.Run("creates tenant with ID", func(t *testing.T) {
		ctx := context.Background()
		tenantID := "tenant123"

		newCtx := WithTenantID(ctx, tenantID)

		retrievedTenant, exists := GetTenant(newCtx)
		assert.True(t, exists)
		assert.Equal(t, tenantID, retrievedTenant.ID)
		assert.True(t, retrievedTenant.IsActive)
	})

	t.Run("handles empty tenant ID", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithTenantID(ctx, "")

		_, exists := GetTenant(newCtx)
		assert.False(t, exists)
	})
}

func TestWithRequestID(t *testing.T) {
	t.Run("adds request ID to context", func(t *testing.T) {
		ctx := context.Background()
		requestID := "req_123456"

		newCtx := WithRequestID(ctx, requestID)

		retrievedID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, requestID, retrievedID)

		// Check that start time is set
		startTime, exists := GetStartTime(newCtx)
		assert.True(t, exists)
		assert.WithinDuration(t, time.Now(), startTime, time.Second)
	})

	t.Run("generates ID when empty", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithRequestID(ctx, "")

		retrievedID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.NotEmpty(t, retrievedID)
		assert.True(t, strings.HasPrefix(retrievedID, "req_"))
	})

	t.Run("generated IDs are unique", func(t *testing.T) {
		ctx := context.Background()

		var ids []string
		for i := 0; i < 100; i++ {
			newCtx := WithRequestID(ctx, "")
			id, exists := GetRequestID(newCtx)
			require.True(t, exists)
			ids = append(ids, id)
		}

		// Check all IDs are unique
		idSet := make(map[string]bool)
		for _, id := range ids {
			assert.False(t, idSet[id], "Duplicate ID found: %s", id)
			idSet[id] = true
		}
	})
}

func TestWithCorrelationID(t *testing.T) {
	t.Run("adds correlation ID to existing request", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithRequestID(ctx, "req_123")
		correlationID := "corr_456"

		newCtx := WithCorrelationID(ctx, correlationID)

		retrievedCorrelationID, exists := GetCorrelationID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, correlationID, retrievedCorrelationID)

		// Original request ID should still exist
		requestID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, "req_123", requestID)
	})

	t.Run("creates new request when no existing request", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "corr_456"

		newCtx := WithCorrelationID(ctx, correlationID)

		retrievedCorrelationID, exists := GetCorrelationID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, correlationID, retrievedCorrelationID)

		// Request ID should be auto-generated
		requestID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
	})

	t.Run("handles empty correlation ID", func(t *testing.T) {
		ctx := context.Background()
		newCtx := WithCorrelationID(ctx, "")

		_, exists := GetCorrelationID(newCtx)
		assert.False(t, exists)
	})
}

func TestGetMethods(t *testing.T) {
	t.Run("returns false for empty context", func(t *testing.T) {
		ctx := context.Background()

		_, exists := GetUser(ctx)
		assert.False(t, exists)

		_, exists = GetUserID(ctx)
		assert.False(t, exists)

		_, exists = GetTenant(ctx)
		assert.False(t, exists)

		_, exists = GetTenantID(ctx)
		assert.False(t, exists)

		_, exists = GetRequestID(ctx)
		assert.False(t, exists)

		_, exists = GetCorrelationID(ctx)
		assert.False(t, exists)

		_, exists = GetSessionID(ctx)
		assert.False(t, exists)

		_, exists = GetStartTime(ctx)
		assert.False(t, exists)

		_, exists = GetDuration(ctx)
		assert.False(t, exists)
	})

	t.Run("extracts user ID correctly", func(t *testing.T) {
		ctx := context.Background()
		user := &UserInfo{ID: "user123", Username: "testuser"}
		ctx = WithUser(ctx, user)

		userID, exists := GetUserID(ctx)
		assert.True(t, exists)
		assert.Equal(t, "user123", userID)
	})

	t.Run("extracts tenant ID correctly", func(t *testing.T) {
		ctx := context.Background()
		tenant := &TenantInfo{ID: "tenant123", Name: "Test Tenant"}
		ctx = WithTenant(ctx, tenant)

		tenantID, exists := GetTenantID(ctx)
		assert.True(t, exists)
		assert.Equal(t, "tenant123", tenantID)
	})
}

func TestMustGetMethods(t *testing.T) {
	t.Run("returns value when present", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithUserID(ctx, "user123")
		ctx = WithTenantID(ctx, "tenant123")
		ctx = WithRequestID(ctx, "req123")

		assert.Equal(t, "user123", MustGetUserID(ctx))
		assert.Equal(t, "tenant123", MustGetTenantID(ctx))
		assert.Equal(t, "req123", MustGetRequestID(ctx))
	})

	t.Run("panics when not present", func(t *testing.T) {
		ctx := context.Background()

		assert.Panics(t, func() {
			MustGetUserID(ctx)
		})

		assert.Panics(t, func() {
			MustGetTenantID(ctx)
		})

		assert.Panics(t, func() {
			MustGetRequestID(ctx)
		})
	})
}

func TestAuthenticationMethods(t *testing.T) {
	t.Run("IsAuthenticated", func(t *testing.T) {
		ctx := context.Background()

		// Not authenticated when no user
		assert.False(t, IsAuthenticated(ctx))

		// Not authenticated when user has empty ID
		ctx = WithUser(ctx, &UserInfo{ID: ""})
		assert.False(t, IsAuthenticated(ctx))

		// Authenticated when user has valid ID
		ctx = WithUser(ctx, &UserInfo{ID: "user123"})
		assert.True(t, IsAuthenticated(ctx))
	})

	t.Run("HasRole", func(t *testing.T) {
		ctx := context.Background()
		user := &UserInfo{
			ID:    "user123",
			Roles: []string{"admin", "user", "moderator"},
		}
		ctx = WithUser(ctx, user)

		assert.True(t, HasRole(ctx, "admin"))
		assert.True(t, HasRole(ctx, "user"))
		assert.True(t, HasRole(ctx, "moderator"))
		assert.False(t, HasRole(ctx, "superadmin"))
		assert.False(t, HasRole(ctx, "guest"))

		// No user in context
		emptyCtx := context.Background()
		assert.False(t, HasRole(emptyCtx, "admin"))
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		ctx := context.Background()
		user := &UserInfo{
			ID:    "user123",
			Roles: []string{"user", "moderator"},
		}
		ctx = WithUser(ctx, user)

		assert.True(t, HasAnyRole(ctx, "admin", "user"))
		assert.True(t, HasAnyRole(ctx, "moderator", "superadmin"))
		assert.False(t, HasAnyRole(ctx, "admin", "superadmin"))
		assert.False(t, HasAnyRole(ctx))

		// No user in context
		emptyCtx := context.Background()
		assert.False(t, HasAnyRole(emptyCtx, "admin", "user"))
	})

	t.Run("HasAllRoles", func(t *testing.T) {
		ctx := context.Background()
		user := &UserInfo{
			ID:    "user123",
			Roles: []string{"admin", "user", "moderator"},
		}
		ctx = WithUser(ctx, user)

		assert.True(t, HasAllRoles(ctx, "admin", "user"))
		assert.True(t, HasAllRoles(ctx, "user"))
		assert.False(t, HasAllRoles(ctx, "admin", "superadmin"))
		assert.False(t, HasAllRoles(ctx))

		// No user in context
		emptyCtx := context.Background()
		assert.False(t, HasAllRoles(emptyCtx, "admin"))
	})
}

func TestConvenienceMethods(t *testing.T) {
	t.Run("NewRequestContext", func(t *testing.T) {
		ctx := context.Background()
		newCtx := NewRequestContext(ctx)

		requestID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		assert.True(t, strings.HasPrefix(requestID, "req_"))

		startTime, exists := GetStartTime(newCtx)
		assert.True(t, exists)
		assert.WithinDuration(t, time.Now(), startTime, time.Second)
	})

	t.Run("NewUserContext", func(t *testing.T) {
		ctx := context.Background()
		newCtx := NewUserContext(ctx, "user123", "tenant456")

		userID, exists := GetUserID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, "user123", userID)

		tenantID, exists := GetTenantID(newCtx)
		assert.True(t, exists)
		assert.Equal(t, "tenant456", tenantID)

		requestID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
	})

	t.Run("NewUserContext with empty values", func(t *testing.T) {
		ctx := context.Background()
		newCtx := NewUserContext(ctx, "", "")

		_, exists := GetUserID(newCtx)
		assert.False(t, exists)

		_, exists = GetTenantID(newCtx)
		assert.False(t, exists)

		// Request ID should still be generated
		requestID, exists := GetRequestID(newCtx)
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
	})
}

func TestGetDuration(t *testing.T) {
	t.Run("calculates duration correctly", func(t *testing.T) {
		ctx := context.Background()
		startTime := time.Now().Add(-100 * time.Millisecond)
		ctx = WithStartTime(ctx, startTime)

		// Wait a bit to ensure some duration
		time.Sleep(10 * time.Millisecond)

		duration, exists := GetDuration(ctx)
		assert.True(t, exists)
		assert.True(t, duration >= 10*time.Millisecond)
		assert.True(t, duration <= 200*time.Millisecond) // Generous upper bound
	})

	t.Run("returns false when no start time", func(t *testing.T) {
		ctx := context.Background()
		_, exists := GetDuration(ctx)
		assert.False(t, exists)
	})
}

func TestContextSummary(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		summary := ContextSummary(ctx)
		assert.Empty(t, summary)
	})

	t.Run("full context", func(t *testing.T) {
		ctx := context.Background()

		user := &UserInfo{
			ID:       "user123",
			Username: "testuser",
			Roles:    []string{"admin", "user"},
		}
		ctx = WithUser(ctx, user)

		tenant := &TenantInfo{
			ID:   "tenant456",
			Name: "Test Tenant",
		}
		ctx = WithTenant(ctx, tenant)

		ctx = WithRequestID(ctx, "req789")
		ctx = WithCorrelationID(ctx, "corr000")
		ctx = WithSessionID(ctx, "sess111")

		summary := ContextSummary(ctx)

		assert.Equal(t, "user123", summary["user_id"])
		assert.Equal(t, "testuser", summary["username"])
		assert.Equal(t, []string{"admin", "user"}, summary["roles"])
		assert.Equal(t, "tenant456", summary["tenant_id"])
		assert.Equal(t, "Test Tenant", summary["tenant_name"])
		assert.Equal(t, "req789", summary["request_id"])
		assert.Equal(t, "corr000", summary["correlation_id"])
		assert.Equal(t, "sess111", summary["session_id"])
		assert.Contains(t, summary, "duration_ms")
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent context creation", func(t *testing.T) {
		const numGoroutines = 100
		var wg sync.WaitGroup
		results := make(chan string, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx := context.Background()
				ctx = WithRequestID(ctx, "")
				requestID, _ := GetRequestID(ctx)
				results <- requestID
			}()
		}

		wg.Wait()
		close(results)

		// Collect all results
		var ids []string
		for id := range results {
			ids = append(ids, id)
		}

		// Verify all IDs are unique
		require.Len(t, ids, numGoroutines)
		idSet := make(map[string]bool)
		for _, id := range ids {
			assert.False(t, idSet[id], "Duplicate ID found: %s", id)
			idSet[id] = true
		}
	})
}

func TestSessionID(t *testing.T) {
	t.Run("adds and retrieves session ID", func(t *testing.T) {
		ctx := context.Background()
		sessionID := "sess_12345"

		ctx = WithSessionID(ctx, sessionID)

		retrievedID, exists := GetSessionID(ctx)
		assert.True(t, exists)
		assert.Equal(t, sessionID, retrievedID)
	})

	t.Run("handles empty session ID", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithSessionID(ctx, "")

		_, exists := GetSessionID(ctx)
		assert.False(t, exists)
	})
}

// Benchmark tests for performance validation
func BenchmarkWithUser(b *testing.B) {
	ctx := context.Background()
	user := &UserInfo{
		ID:       "user123",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"admin", "user"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WithUser(ctx, user)
	}
}

func BenchmarkGetUser(b *testing.B) {
	ctx := context.Background()
	user := &UserInfo{ID: "user123"}
	ctx = WithUser(ctx, user)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GetUser(ctx)
	}
}

func BenchmarkWithRequestID(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WithRequestID(ctx, "")
	}
}

func BenchmarkWithRequestID_Existing(b *testing.B) {
	ctx := context.Background()
	requestID := "req_existing_12345"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WithRequestID(ctx, requestID)
	}
}

func BenchmarkGenerateRequestID(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = generateRequestID()
	}
}

func BenchmarkWithUserContext_Complete(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewUserContext(ctx, "user123", "tenant456")
	}
}

func BenchmarkContextSummary(b *testing.B) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user123")
	ctx = WithTenantID(ctx, "tenant456")
	ctx = WithRequestID(ctx, "req789")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ContextSummary(ctx)
	}
}

func BenchmarkContextSummary_Empty(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ContextSummary(ctx)
	}
}

func BenchmarkWithCorrelationID(b *testing.B) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req123")
	correlationID := "corr456"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = WithCorrelationID(ctx, correlationID)
	}
}

func BenchmarkHasRole(b *testing.B) {
	ctx := context.Background()
	user := &UserInfo{
		ID:    "user123",
		Roles: []string{"admin", "user", "moderator", "editor", "viewer"},
	}
	ctx = WithUser(ctx, user)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = HasRole(ctx, "moderator")
	}
}

func BenchmarkHasAnyRole(b *testing.B) {
	ctx := context.Background()
	user := &UserInfo{
		ID:    "user123",
		Roles: []string{"admin", "user", "moderator", "editor", "viewer"},
	}
	ctx = WithUser(ctx, user)
	searchRoles := []string{"superadmin", "moderator", "guest"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = HasAnyRole(ctx, searchRoles...)
	}
}
