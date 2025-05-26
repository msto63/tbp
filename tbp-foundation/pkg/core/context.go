// File: context.go
// Title: Extended Context Management for TBP
// Description: Provides extended context functionality for propagating
//              user information, tenant data, request IDs, and correlation IDs
//              throughout the entire call chain in a type-safe manner.
//              Extends Go's standard context.Context with enterprise features.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial implementation with user, tenant, and request tracking

package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

// contextKey is a private type used for context keys to avoid collisions
// with other packages that might use context values.
type contextKey string

// Context keys for storing values in context.Context
const (
	keyUserID        contextKey = "tbp:user_id"
	keyTenantID      contextKey = "tbp:tenant_id"
	keyRequestID     contextKey = "tbp:request_id"
	keyCorrelationID contextKey = "tbp:correlation_id"
	keyStartTime     contextKey = "tbp:start_time"
	keyUserRoles     contextKey = "tbp:user_roles"
	keySessionID     contextKey = "tbp:session_id"
)

// UserInfo represents user information stored in context
type UserInfo struct {
	ID       string    `json:"id"`
	Username string    `json:"username,omitempty"`
	Email    string    `json:"email,omitempty"`
	Roles    []string  `json:"roles,omitempty"`
	TenantID string    `json:"tenant_id,omitempty"`
	LoginAt  time.Time `json:"login_at,omitempty"`
}

// TenantInfo represents tenant information stored in context
type TenantInfo struct {
	ID          string            `json:"id"`
	Name        string            `json:"name,omitempty"`
	Domain      string            `json:"domain,omitempty"`
	Settings    map[string]string `json:"settings,omitempty"`
	IsActive    bool              `json:"is_active"`
	Permissions []string          `json:"permissions,omitempty"`
}

// RequestInfo represents request tracking information
type RequestInfo struct {
	ID            string        `json:"id"`
	CorrelationID string        `json:"correlation_id,omitempty"`
	StartTime     time.Time     `json:"start_time"`
	UserAgent     string        `json:"user_agent,omitempty"`
	RemoteAddr    string        `json:"remote_addr,omitempty"`
	Method        string        `json:"method,omitempty"`
	Path          string        `json:"path,omitempty"`
	Duration      time.Duration `json:"duration,omitempty"`
}

// WithUser adds user information to the context.
// Returns a new context with the user info attached.
func WithUser(ctx context.Context, user *UserInfo) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if user == nil {
		return ctx
	}
	return context.WithValue(ctx, keyUserID, user)
}

// WithUserID adds a user ID to the context.
// This is a convenience function for cases where only the ID is available.
func WithUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	user := &UserInfo{ID: userID}
	return context.WithValue(ctx, keyUserID, user)
}

// WithTenant adds tenant information to the context.
// Returns a new context with the tenant info attached.
func WithTenant(ctx context.Context, tenant *TenantInfo) context.Context {
	if tenant == nil {
		return ctx
	}
	return context.WithValue(ctx, keyTenantID, tenant)
}

// WithTenantID adds a tenant ID to the context.
// This is a convenience function for cases where only the ID is available.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	tenant := &TenantInfo{ID: tenantID, IsActive: true}
	return context.WithValue(ctx, keyTenantID, tenant)
}

// WithRequestID adds a request ID to the context.
// If requestID is empty, a new UUID-like ID is generated automatically.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		requestID = generateRequestID()
	}

	request := &RequestInfo{
		ID:        requestID,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, keyRequestID, request)
}

// WithCorrelationID adds a correlation ID to the context.
// Correlation IDs are used to trace requests across multiple services.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return ctx
	}

	// If we already have a RequestInfo, update it
	if req, exists := GetRequestInfo(ctx); exists {
		req.CorrelationID = correlationID
		return context.WithValue(ctx, keyRequestID, req)
	}

	// Otherwise create new RequestInfo with correlation ID
	request := &RequestInfo{
		ID:            generateRequestID(),
		CorrelationID: correlationID,
		StartTime:     time.Now(),
	}
	return context.WithValue(ctx, keyRequestID, request)
}

// WithStartTime adds a start time to the context.
// This is useful for tracking request duration.
func WithStartTime(ctx context.Context, startTime time.Time) context.Context {
	if req, exists := GetRequestInfo(ctx); exists {
		req.StartTime = startTime
		return context.WithValue(ctx, keyRequestID, req)
	}

	request := &RequestInfo{
		ID:        generateRequestID(),
		StartTime: startTime,
	}
	return context.WithValue(ctx, keyRequestID, request)
}

// WithSessionID adds a session ID to the context.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	if sessionID == "" {
		return ctx
	}
	return context.WithValue(ctx, keySessionID, sessionID)
}

// GetUser retrieves user information from the context.
// Returns the UserInfo and true if found, nil and false otherwise.
func GetUser(ctx context.Context) (*UserInfo, bool) {
	if user, ok := ctx.Value(keyUserID).(*UserInfo); ok && user != nil {
		return user, true
	}
	return nil, false
}

// GetUserID retrieves the user ID from the context.
// Returns the user ID and true if found, empty string and false otherwise.
func GetUserID(ctx context.Context) (string, bool) {
	if user, ok := GetUser(ctx); ok {
		return user.ID, true
	}
	return "", false
}

// GetTenant retrieves tenant information from the context.
// Returns the TenantInfo and true if found, nil and false otherwise.
func GetTenant(ctx context.Context) (*TenantInfo, bool) {
	if tenant, ok := ctx.Value(keyTenantID).(*TenantInfo); ok && tenant != nil {
		return tenant, true
	}
	return nil, false
}

// GetTenantID retrieves the tenant ID from the context.
// Returns the tenant ID and true if found, empty string and false otherwise.
func GetTenantID(ctx context.Context) (string, bool) {
	if tenant, ok := GetTenant(ctx); ok {
		return tenant.ID, true
	}
	return "", false
}

// GetRequestInfo retrieves request information from the context.
// Returns the RequestInfo and true if found, nil and false otherwise.
func GetRequestInfo(ctx context.Context) (*RequestInfo, bool) {
	if req, ok := ctx.Value(keyRequestID).(*RequestInfo); ok && req != nil {
		return req, true
	}
	return nil, false
}

// GetRequestID retrieves the request ID from the context.
// Returns the request ID and true if found, empty string and false otherwise.
func GetRequestID(ctx context.Context) (string, bool) {
	if req, ok := GetRequestInfo(ctx); ok {
		return req.ID, true
	}
	return "", false
}

// GetCorrelationID retrieves the correlation ID from the context.
// Returns the correlation ID and true if found, empty string and false otherwise.
func GetCorrelationID(ctx context.Context) (string, bool) {
	if req, ok := GetRequestInfo(ctx); ok && req.CorrelationID != "" {
		return req.CorrelationID, true
	}
	return "", false
}

// GetSessionID retrieves the session ID from the context.
// Returns the session ID and true if found, empty string and false otherwise.
func GetSessionID(ctx context.Context) (string, bool) {
	if sessionID, ok := ctx.Value(keySessionID).(string); ok && sessionID != "" {
		return sessionID, true
	}
	return "", false
}

// GetStartTime retrieves the start time from the context.
// Returns the start time and true if found, zero time and false otherwise.
func GetStartTime(ctx context.Context) (time.Time, bool) {
	if req, ok := GetRequestInfo(ctx); ok && !req.StartTime.IsZero() {
		return req.StartTime, true
	}
	return time.Time{}, false
}

// GetDuration calculates the duration since the start time in the context.
// Returns the duration and true if start time is found, zero duration and false otherwise.
func GetDuration(ctx context.Context) (time.Duration, bool) {
	if startTime, ok := GetStartTime(ctx); ok {
		return time.Since(startTime), true
	}
	return 0, false
}

// MustGetUserID retrieves the user ID from the context or panics if not found.
// This should only be used in contexts where the user ID is guaranteed to exist.
func MustGetUserID(ctx context.Context) string {
	if userID, ok := GetUserID(ctx); ok {
		return userID
	}
	panic("user ID not found in context")
}

// MustGetTenantID retrieves the tenant ID from the context or panics if not found.
// This should only be used in contexts where the tenant ID is guaranteed to exist.
func MustGetTenantID(ctx context.Context) string {
	if tenantID, ok := GetTenantID(ctx); ok {
		return tenantID
	}
	panic("tenant ID not found in context")
}

// MustGetRequestID retrieves the request ID from the context or panics if not found.
// This should only be used in contexts where the request ID is guaranteed to exist.
func MustGetRequestID(ctx context.Context) string {
	if requestID, ok := GetRequestID(ctx); ok {
		return requestID
	}
	panic("request ID not found in context")
}

// IsAuthenticated checks if the context contains valid user information.
// Returns true if user information is present and has a valid ID.
func IsAuthenticated(ctx context.Context) bool {
	if user, ok := GetUser(ctx); ok {
		return user.ID != ""
	}
	return false
}

// HasRole checks if the authenticated user has a specific role.
// Returns true if the user is authenticated and has the specified role.
func HasRole(ctx context.Context, role string) bool {
	if user, ok := GetUser(ctx); ok {
		for _, userRole := range user.Roles {
			if userRole == role {
				return true
			}
		}
	}
	return false
}

// HasAnyRole checks if the authenticated user has any of the specified roles.
// Returns true if the user is authenticated and has at least one of the roles.
func HasAnyRole(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the authenticated user has all of the specified roles.
// Returns true if the user is authenticated and has all of the roles.
func HasAllRoles(ctx context.Context, roles ...string) bool {
	for _, role := range roles {
		if !HasRole(ctx, role) {
			return false
		}
	}
	return len(roles) > 0 // Return false if no roles specified
}

// NewRequestContext creates a new context with request tracking information.
// This is typically called at the beginning of request handling.
func NewRequestContext(ctx context.Context) context.Context {
	requestID := generateRequestID()
	request := &RequestInfo{
		ID:        requestID,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, keyRequestID, request)
}

// NewUserContext creates a new context with user and request information.
// This is a convenience function for creating a complete context.
func NewUserContext(ctx context.Context, userID, tenantID string) context.Context {
	ctx = NewRequestContext(ctx)
	if userID != "" {
		ctx = WithUserID(ctx, userID)
	}
	if tenantID != "" {
		ctx = WithTenantID(ctx, tenantID)
	}
	return ctx
}

// generateRequestID creates a new unique request ID.
// Uses crypto/rand for cryptographically secure random bytes.
func generateRequestID() string {
	bytes := make([]byte, 16) // 128-bit random ID
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return "req_" + hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000")))
	}
	return "req_" + hex.EncodeToString(bytes)
}

// ContextSummary returns a summary of all context values for debugging.
// This is useful for logging and troubleshooting context propagation.
func ContextSummary(ctx context.Context) map[string]interface{} {
	summary := make(map[string]interface{})

	if user, ok := GetUser(ctx); ok {
		summary["user_id"] = user.ID
		summary["username"] = user.Username
		if len(user.Roles) > 0 {
			summary["roles"] = user.Roles
		}
	}

	if tenant, ok := GetTenant(ctx); ok {
		summary["tenant_id"] = tenant.ID
		summary["tenant_name"] = tenant.Name
	}

	if req, ok := GetRequestInfo(ctx); ok {
		summary["request_id"] = req.ID
		if req.CorrelationID != "" {
			summary["correlation_id"] = req.CorrelationID
		}
		if !req.StartTime.IsZero() {
			summary["duration_ms"] = time.Since(req.StartTime).Milliseconds()
		}
	}

	if sessionID, ok := GetSessionID(ctx); ok {
		summary["session_id"] = sessionID
	}

	return summary
}
