// File: types.go
// Title: Common Types and Interfaces for TBP Core
// Description: Defines fundamental types, interfaces, and data structures
//              used throughout the TBP platform. Provides a consistent
//              foundation for domain modeling, service contracts, and
//              data exchange between components.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial implementation with basic types and interfaces

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ID represents a unique identifier in the TBP system.
// Uses string to support various ID formats (UUID, numeric, custom).
type ID string

// String returns the string representation of the ID.
func (id ID) String() string {
	return string(id)
}

// IsEmpty checks if the ID is empty or unset.
func (id ID) IsEmpty() bool {
	return string(id) == ""
}

// MarshalJSON implements json.Marshaler interface.
func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(id))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (id *ID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*id = ID(s)
	return nil
}

// Entity represents the base interface for all domain entities.
// All business objects should implement this interface.
type Entity interface {
	// GetID returns the unique identifier of the entity
	GetID() ID

	// GetVersion returns the version for optimistic locking
	GetVersion() int64

	// GetCreatedAt returns when the entity was created
	GetCreatedAt() time.Time

	// GetUpdatedAt returns when the entity was last updated
	GetUpdatedAt() time.Time
}

// BaseEntity provides a default implementation of the Entity interface.
// Domain entities can embed this struct to inherit standard fields.
type BaseEntity struct {
	ID        ID        `json:"id" db:"id"`
	Version   int64     `json:"version" db:"version"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetID implements Entity interface.
func (e *BaseEntity) GetID() ID {
	return e.ID
}

// GetVersion implements Entity interface.
func (e *BaseEntity) GetVersion() int64 {
	return e.Version
}

// GetCreatedAt implements Entity interface.
func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt implements Entity interface.
func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// SetID sets the entity ID (typically used during creation).
func (e *BaseEntity) SetID(id ID) {
	e.ID = id
}

// IncrementVersion increments the version for optimistic locking.
func (e *BaseEntity) IncrementVersion() {
	e.Version++
	e.UpdatedAt = time.Now()
}

// Touch updates the UpdatedAt timestamp without changing version.
func (e *BaseEntity) Touch() {
	e.UpdatedAt = time.Now()
}

// Service represents the base interface for all business services.
// Services encapsulate business logic and coordinate between repositories.
type Service interface {
	// Name returns the service name for logging and metrics
	Name() string

	// Health checks if the service is healthy and ready to serve requests
	Health(ctx context.Context) error
}

// Repository represents the base interface for data access objects.
// Repositories abstract the data persistence layer.
type Repository[T Entity] interface {
	// Create persists a new entity
	Create(ctx context.Context, entity T) error

	// GetByID retrieves an entity by its ID
	GetByID(ctx context.Context, id ID) (T, error)

	// Update modifies an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity by its ID
	Delete(ctx context.Context, id ID) error

	// List retrieves entities with optional filtering and pagination
	List(ctx context.Context, opts ListOptions) ([]T, error)

	// Count returns the total number of entities matching the criteria
	Count(ctx context.Context, opts ListOptions) (int64, error)
}

// ListOptions defines parameters for list operations.
// Provides standardized pagination, sorting, and filtering.
type ListOptions struct {
	// Offset for pagination (number of records to skip)
	Offset int64 `json:"offset" form:"offset"`

	// Limit for pagination (maximum number of records to return)
	Limit int64 `json:"limit" form:"limit"`

	// SortBy specifies the field to sort by
	SortBy string `json:"sort_by" form:"sort_by"`

	// SortOrder specifies the sort direction (asc/desc)
	SortOrder SortOrder `json:"sort_order" form:"sort_order"`

	// Filters contains field-specific filter criteria
	Filters map[string]interface{} `json:"filters" form:"-"`

	// Search provides full-text search functionality
	Search string `json:"search" form:"search"`

	// IncludeDeleted includes soft-deleted records in results
	IncludeDeleted bool `json:"include_deleted" form:"include_deleted"`
}

// SortOrder represents the direction of sorting.
type SortOrder string

const (
	// SortAsc represents ascending sort order
	SortAsc SortOrder = "asc"

	// SortDesc represents descending sort order
	SortDesc SortOrder = "desc"
)

// IsValid checks if the sort order is valid.
func (so SortOrder) IsValid() bool {
	return so == SortAsc || so == SortDesc
}

// String returns the string representation of sort order.
func (so SortOrder) String() string {
	return string(so)
}

// NewListOptions creates ListOptions with sensible defaults.
func NewListOptions() ListOptions {
	return ListOptions{
		Offset:    0,
		Limit:     50, // Default page size
		SortOrder: SortAsc,
		Filters:   make(map[string]interface{}),
	}
}

// WithOffset sets the offset for pagination.
func (opts ListOptions) WithOffset(offset int64) ListOptions {
	opts.Offset = offset
	return opts
}

// WithLimit sets the limit for pagination.
func (opts ListOptions) WithLimit(limit int64) ListOptions {
	opts.Limit = limit
	return opts
}

// WithSort sets the sort field and order.
func (opts ListOptions) WithSort(field string, order SortOrder) ListOptions {
	opts.SortBy = field
	opts.SortOrder = order
	return opts
}

// WithFilter adds a filter criterion.
func (opts ListOptions) WithFilter(field string, value interface{}) ListOptions {
	if opts.Filters == nil {
		opts.Filters = make(map[string]interface{})
	}
	opts.Filters[field] = value
	return opts
}

// WithSearch sets the search term.
func (opts ListOptions) WithSearch(search string) ListOptions {
	opts.Search = search
	return opts
}

// GetPage calculates the page number based on offset and limit.
func (opts ListOptions) GetPage() int64 {
	if opts.Limit <= 0 {
		return 1
	}
	return (opts.Offset / opts.Limit) + 1
}

// Validate checks if the ListOptions are valid and returns an error if not.
// This ensures that pagination parameters are within acceptable ranges
// and that sort order is valid if specified.
func (opts ListOptions) Validate() error {
	if opts.Limit < 0 {
		return New("limit cannot be negative")
	}
	if opts.Offset < 0 {
		return New("offset cannot be negative")
	}
	if opts.SortOrder != "" && !opts.SortOrder.IsValid() {
		return Newf("invalid sort order: %s", opts.SortOrder)
	}

	// Optional: Add reasonable upper limits to prevent abuse
	if opts.Limit > 1000 {
		return New("limit cannot exceed 1000")
	}

	return nil
}

// ListResult represents the result of a list operation with pagination metadata.
type ListResult[T any] struct {
	// Items contains the actual data
	Items []T `json:"items"`

	// Total is the total number of items matching the criteria
	Total int64 `json:"total"`

	// Offset is the current offset
	Offset int64 `json:"offset"`

	// Limit is the current limit
	Limit int64 `json:"limit"`

	// HasMore indicates if there are more items available
	HasMore bool `json:"has_more"`
}

// NewListResult creates a new ListResult with calculated metadata.
func NewListResult[T any](items []T, total int64, opts ListOptions) *ListResult[T] {
	hasMore := opts.Offset+int64(len(items)) < total

	return &ListResult[T]{
		Items:   items,
		Total:   total,
		Offset:  opts.Offset,
		Limit:   opts.Limit,
		HasMore: hasMore,
	}
}

// IsEmpty checks if the result contains no items.
func (r *ListResult[T]) IsEmpty() bool {
	return len(r.Items) == 0
}

// GetPageInfo returns pagination information.
func (r *ListResult[T]) GetPageInfo() PageInfo {
	var totalPages int64 = 1
	if r.Limit > 0 {
		totalPages = (r.Total + r.Limit - 1) / r.Limit // Ceiling division
	}

	var currentPage int64 = 1
	if r.Limit > 0 {
		currentPage = (r.Offset / r.Limit) + 1
	}

	return PageInfo{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   r.Total,
		ItemsPerPage: r.Limit,
		HasNext:      r.HasMore,
		HasPrev:      r.Offset > 0,
	}
}

// PageInfo provides detailed pagination information.
type PageInfo struct {
	CurrentPage  int64 `json:"current_page"`
	TotalPages   int64 `json:"total_pages"`
	TotalItems   int64 `json:"total_items"`
	ItemsPerPage int64 `json:"items_per_page"`
	HasNext      bool  `json:"has_next"`
	HasPrev      bool  `json:"has_prev"`
}

// Handler represents a generic handler interface for commands, queries, or events.
type Handler[TRequest, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

// Command represents a command in the CQRS pattern.
// Commands change state and typically don't return data.
type Command interface {
	// CommandType returns the type of command for routing
	CommandType() string

	// Validate checks if the command is valid
	Validate() error
}

// Query represents a query in the CQRS pattern.
// Queries read data and don't change state.
type Query interface {
	// QueryType returns the type of query for routing
	QueryType() string

	// Validate checks if the query is valid
	Validate() error
}

// Event represents a domain event that occurred in the system.
type Event interface {
	// EventType returns the type of event
	EventType() string

	// EventID returns a unique identifier for this event occurrence
	EventID() string

	// Timestamp returns when the event occurred
	Timestamp() time.Time

	// AggregateID returns the ID of the aggregate that produced this event
	AggregateID() string

	// Version returns the version of the aggregate after this event
	Version() int64
}

// BaseEvent provides a default implementation of the Event interface.
type BaseEvent struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	AggregateId string    `json:"aggregate_id"`
	Ver         int64     `json:"version"`
	OccurredAt  time.Time `json:"occurred_at"`
	Data        []byte    `json:"data,omitempty"`
}

// EventType implements Event interface.
func (e *BaseEvent) EventType() string {
	return e.Type
}

// EventID implements Event interface.
func (e *BaseEvent) EventID() string {
	return e.ID
}

// Timestamp implements Event interface.
func (e *BaseEvent) Timestamp() time.Time {
	return e.OccurredAt
}

// AggregateID implements Event interface.
func (e *BaseEvent) AggregateID() string {
	return e.AggregateId
}

// Version implements Event interface.
func (e *BaseEvent) Version() int64 {
	return e.Ver
}

// Value represents a value object in domain-driven design.
// Value objects are immutable and defined by their attributes.
type Value interface {
	// Equals checks if two value objects are equal
	Equals(other Value) bool

	// String returns a string representation
	String() string
}

// Status represents a generic status enumeration.
type Status string

const (
	// StatusActive represents an active state
	StatusActive Status = "active"

	// StatusInactive represents an inactive state
	StatusInactive Status = "inactive"

	// StatusPending represents a pending state
	StatusPending Status = "pending"

	// StatusCompleted represents a completed state
	StatusCompleted Status = "completed"

	// StatusCancelled represents a cancelled state
	StatusCancelled Status = "cancelled"

	// StatusDeleted represents a deleted state
	StatusDeleted Status = "deleted"
)

// IsValid checks if the status is one of the predefined values.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusInactive, StatusPending, StatusCompleted, StatusCancelled, StatusDeleted:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status.
func (s Status) String() string {
	return string(s)
}

// Priority represents a priority level enumeration.
type Priority int

const (
	// PriorityLow represents low priority
	PriorityLow Priority = 1

	// PriorityMedium represents medium priority
	PriorityMedium Priority = 2

	// PriorityHigh represents high priority
	PriorityHigh Priority = 3

	// PriorityCritical represents critical priority
	PriorityCritical Priority = 4
)

// IsValid checks if the priority is within valid range.
func (p Priority) IsValid() bool {
	return p >= PriorityLow && p <= PriorityCritical
}

// String returns the string representation of the priority.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler interface for Priority.
func (p Priority) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON implements json.Unmarshaler interface for Priority.
func (p *Priority) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try parsing as integer
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*p = Priority(i)
		return nil
	}

	switch s {
	case "low":
		*p = PriorityLow
	case "medium":
		*p = PriorityMedium
	case "high":
		*p = PriorityHigh
	case "critical":
		*p = PriorityCritical
	default:
		return fmt.Errorf("invalid priority: %s", s)
	}

	return nil
}

// Validator provides validation functionality for any type.
type Validator interface {
	Validate() error
}

// Lifecycle provides lifecycle management for services and components.
type Lifecycle interface {
	// Start initializes and starts the component
	Start(ctx context.Context) error

	// Stop gracefully shuts down the component
	Stop(ctx context.Context) error

	// IsRunning returns true if the component is running
	IsRunning() bool
}

// HealthChecker provides health check functionality.
type HealthChecker interface {
	// Health returns the current health status
	Health(ctx context.Context) HealthStatus
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status  string            `json:"status"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// Health status constants.
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDegraded  = "degraded"
)

// IsHealthy checks if the status indicates healthy state.
func (hs HealthStatus) IsHealthy() bool {
	return hs.Status == HealthStatusHealthy
}

// Metadata represents key-value metadata that can be attached to entities.
type Metadata map[string]string

// Get retrieves a metadata value by key.
func (m Metadata) Get(key string) (string, bool) {
	if m == nil {
		return "", false
	}
	value, exists := m[key]
	return value, exists
}

// Set sets a metadata value.
func (m Metadata) Set(key, value string) {
	if m != nil {
		m[key] = value
	}
}

// Has checks if a metadata key exists.
func (m Metadata) Has(key string) bool {
	if m == nil {
		return false
	}
	_, exists := m[key]
	return exists
}

// Clone creates a deep copy of the metadata.
func (m Metadata) Clone() Metadata {
	if m == nil {
		return nil
	}

	clone := make(Metadata, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// ParseID converts a string to an ID, handling common parsing scenarios.
func ParseID(s string) (ID, error) {
	if s == "" {
		return "", Newf("ID cannot be empty")
	}
	return ID(s), nil
}

// MustParseID converts a string to an ID, panicking on error.
// Should only be used in contexts where the ID is guaranteed to be valid.
func MustParseID(s string) ID {
	id, err := ParseID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ToIntID converts an ID to integer if possible.
func ToIntID(id ID) (int64, error) {
	return strconv.ParseInt(string(id), 10, 64)
}

// FromIntID creates an ID from an integer.
func FromIntID(i int64) ID {
	return ID(strconv.FormatInt(i, 10))
}
