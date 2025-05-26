// File: types_test.go
// Title: Tests for Common Types and Interfaces
// Description: Comprehensive test suite for TBP core types including generics,
//              JSON serialization, business logic validation, pagination,
//              and interface compliance. Tests cover edge cases, performance,
//              and type safety for the foundation layer.
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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntity is a mock entity for testing generic repository functionality
type TestEntity struct {
	BaseEntity
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

func (e *TestEntity) GetTestName() string {
	return e.Name
}

func TestID(t *testing.T) {
	t.Run("string conversion", func(t *testing.T) {
		id := ID("test123")
		assert.Equal(t, "test123", id.String())
		assert.Equal(t, "test123", string(id))
	})

	t.Run("empty check", func(t *testing.T) {
		var emptyID ID
		assert.True(t, emptyID.IsEmpty())
		assert.Equal(t, "", emptyID.String())

		nonEmptyID := ID("test")
		assert.False(t, nonEmptyID.IsEmpty())
	})

	t.Run("JSON marshaling", func(t *testing.T) {
		id := ID("test123")

		// Marshal
		data, err := json.Marshal(id)
		require.NoError(t, err)
		assert.Equal(t, `"test123"`, string(data))

		// Unmarshal
		var unmarshaled ID
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		assert.Equal(t, id, unmarshaled)
	})

	t.Run("JSON unmarshal empty", func(t *testing.T) {
		var id ID
		err := json.Unmarshal([]byte(`""`), &id)
		require.NoError(t, err)
		assert.True(t, id.IsEmpty())
	})

	t.Run("JSON unmarshal invalid", func(t *testing.T) {
		var id ID
		err := json.Unmarshal([]byte(`123`), &id)
		assert.Error(t, err)
	})
}

func TestBaseEntity(t *testing.T) {
	t.Run("implements Entity interface", func(t *testing.T) {
		entity := &BaseEntity{
			ID:        ID("test123"),
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Test interface compliance
		var e Entity = entity
		assert.Equal(t, ID("test123"), e.GetID())
		assert.Equal(t, int64(1), e.GetVersion())
		assert.False(t, e.GetCreatedAt().IsZero())
		assert.False(t, e.GetUpdatedAt().IsZero())
	})

	t.Run("set ID", func(t *testing.T) {
		entity := &BaseEntity{}
		entity.SetID(ID("new123"))
		assert.Equal(t, ID("new123"), entity.GetID())
	})

	t.Run("increment version", func(t *testing.T) {
		entity := &BaseEntity{Version: 1}
		originalTime := entity.UpdatedAt

		time.Sleep(1 * time.Millisecond) // Ensure time difference
		entity.IncrementVersion()

		assert.Equal(t, int64(2), entity.Version)
		assert.True(t, entity.UpdatedAt.After(originalTime))
	})

	t.Run("touch updates timestamp", func(t *testing.T) {
		entity := &BaseEntity{Version: 1}
		originalTime := entity.UpdatedAt
		originalVersion := entity.Version

		time.Sleep(1 * time.Millisecond) // Ensure time difference
		entity.Touch()

		assert.Equal(t, originalVersion, entity.Version) // Version unchanged
		assert.True(t, entity.UpdatedAt.After(originalTime))
	})
}

func TestListOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		opts := NewListOptions()
		assert.Equal(t, int64(0), opts.Offset)
		assert.Equal(t, int64(50), opts.Limit)
		assert.Equal(t, SortAsc, opts.SortOrder)
		assert.NotNil(t, opts.Filters)
		assert.Empty(t, opts.Filters)
	})

	t.Run("with offset", func(t *testing.T) {
		opts := NewListOptions().WithOffset(100)
		assert.Equal(t, int64(100), opts.Offset)
	})

	t.Run("with limit", func(t *testing.T) {
		opts := NewListOptions().WithLimit(25)
		assert.Equal(t, int64(25), opts.Limit)
	})

	t.Run("with sort", func(t *testing.T) {
		opts := NewListOptions().WithSort("name", SortDesc)
		assert.Equal(t, "name", opts.SortBy)
		assert.Equal(t, SortDesc, opts.SortOrder)
	})

	t.Run("with filter", func(t *testing.T) {
		opts := NewListOptions().WithFilter("status", "active")
		assert.Equal(t, "active", opts.Filters["status"])
	})

	t.Run("with search", func(t *testing.T) {
		opts := NewListOptions().WithSearch("test query")
		assert.Equal(t, "test query", opts.Search)
	})

	t.Run("method chaining", func(t *testing.T) {
		opts := NewListOptions().
			WithOffset(10).
			WithLimit(20).
			WithSort("created_at", SortDesc).
			WithFilter("status", "active").
			WithSearch("test")

		assert.Equal(t, int64(10), opts.Offset)
		assert.Equal(t, int64(20), opts.Limit)
		assert.Equal(t, "created_at", opts.SortBy)
		assert.Equal(t, SortDesc, opts.SortOrder)
		assert.Equal(t, "active", opts.Filters["status"])
		assert.Equal(t, "test", opts.Search)
	})

	t.Run("get page calculation", func(t *testing.T) {
		// Page 1
		opts := ListOptions{Offset: 0, Limit: 10}
		assert.Equal(t, int64(1), opts.GetPage())

		// Page 2
		opts = ListOptions{Offset: 10, Limit: 10}
		assert.Equal(t, int64(2), opts.GetPage())

		// Page 3
		opts = ListOptions{Offset: 20, Limit: 10}
		assert.Equal(t, int64(3), opts.GetPage())

		// Zero limit
		opts = ListOptions{Offset: 100, Limit: 0}
		assert.Equal(t, int64(1), opts.GetPage())
	})
}

func TestSortOrder(t *testing.T) {
	t.Run("valid sort orders", func(t *testing.T) {
		assert.True(t, SortAsc.IsValid())
		assert.True(t, SortDesc.IsValid())

		assert.Equal(t, "asc", SortAsc.String())
		assert.Equal(t, "desc", SortDesc.String())
	})

	t.Run("invalid sort order", func(t *testing.T) {
		invalid := SortOrder("invalid")
		assert.False(t, invalid.IsValid())
		assert.Equal(t, "invalid", invalid.String())
	})
}

func TestListResult(t *testing.T) {
	t.Run("create with items", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		opts := ListOptions{Offset: 0, Limit: 10}
		result := NewListResult(items, 100, opts)

		assert.Equal(t, items, result.Items)
		assert.Equal(t, int64(100), result.Total)
		assert.Equal(t, int64(0), result.Offset)
		assert.Equal(t, int64(10), result.Limit)
		assert.True(t, result.HasMore) // 3 items, total 100, so more available
	})

	t.Run("no more items", func(t *testing.T) {
		items := []string{"item1", "item2"}
		opts := ListOptions{Offset: 0, Limit: 10}
		result := NewListResult(items, 2, opts)

		assert.False(t, result.HasMore) // 2 items, total 2, no more available
	})

	t.Run("is empty", func(t *testing.T) {
		emptyResult := NewListResult([]string{}, 0, ListOptions{})
		assert.True(t, emptyResult.IsEmpty())

		nonEmptyResult := NewListResult([]string{"item"}, 1, ListOptions{})
		assert.False(t, nonEmptyResult.IsEmpty())
	})

	t.Run("get page info", func(t *testing.T) {
		items := []string{"item1", "item2", "item3"}
		opts := ListOptions{Offset: 20, Limit: 10}
		result := NewListResult(items, 100, opts)

		pageInfo := result.GetPageInfo()
		assert.Equal(t, int64(3), pageInfo.CurrentPage) // Offset 20, Limit 10 = Page 3
		assert.Equal(t, int64(10), pageInfo.TotalPages) // 100 total / 10 per page = 10 pages
		assert.Equal(t, int64(100), pageInfo.TotalItems)
		assert.Equal(t, int64(10), pageInfo.ItemsPerPage)
		assert.True(t, pageInfo.HasNext) // Page 3 of 10
		assert.True(t, pageInfo.HasPrev) // Page 3, so has previous
	})

	t.Run("edge cases for pagination", func(t *testing.T) {
		// First page
		opts := ListOptions{Offset: 0, Limit: 10}
		result := NewListResult([]string{"item"}, 50, opts)
		pageInfo := result.GetPageInfo()
		assert.Equal(t, int64(1), pageInfo.CurrentPage)
		assert.False(t, pageInfo.HasPrev)
		assert.True(t, pageInfo.HasNext)

		// Last page
		opts = ListOptions{Offset: 40, Limit: 10}
		result = NewListResult([]string{"item"}, 50, opts)
		pageInfo = result.GetPageInfo()
		assert.Equal(t, int64(5), pageInfo.CurrentPage)
		assert.True(t, pageInfo.HasPrev)
		assert.True(t, pageInfo.HasNext) // Still has next because 50 items / 10 per page = 5 pages
	})
}

func TestStatus(t *testing.T) {
	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []Status{
			StatusActive, StatusInactive, StatusPending,
			StatusCompleted, StatusCancelled, StatusDeleted,
		}

		for _, status := range validStatuses {
			assert.True(t, status.IsValid())
			assert.NotEmpty(t, status.String())
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		invalid := Status("invalid")
		assert.False(t, invalid.IsValid())
		assert.Equal(t, "invalid", invalid.String())
	})

	t.Run("string conversion", func(t *testing.T) {
		assert.Equal(t, "active", StatusActive.String())
		assert.Equal(t, "inactive", StatusInactive.String())
		assert.Equal(t, "pending", StatusPending.String())
		assert.Equal(t, "completed", StatusCompleted.String())
		assert.Equal(t, "cancelled", StatusCancelled.String())
		assert.Equal(t, "deleted", StatusDeleted.String())
	})
}

func TestPriority(t *testing.T) {
	t.Run("valid priorities", func(t *testing.T) {
		validPriorities := []Priority{
			PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical,
		}

		for _, priority := range validPriorities {
			assert.True(t, priority.IsValid())
			assert.NotEmpty(t, priority.String())
		}
	})

	t.Run("invalid priority", func(t *testing.T) {
		invalid := Priority(0)
		assert.False(t, invalid.IsValid())
		assert.Equal(t, "unknown", invalid.String())

		invalid = Priority(10)
		assert.False(t, invalid.IsValid())
		assert.Equal(t, "unknown", invalid.String())
	})

	t.Run("string conversion", func(t *testing.T) {
		assert.Equal(t, "low", PriorityLow.String())
		assert.Equal(t, "medium", PriorityMedium.String())
		assert.Equal(t, "high", PriorityHigh.String())
		assert.Equal(t, "critical", PriorityCritical.String())
	})

	t.Run("JSON marshaling", func(t *testing.T) {
		// Marshal as string
		data, err := json.Marshal(PriorityHigh)
		require.NoError(t, err)
		assert.Equal(t, `"high"`, string(data))

		// Unmarshal from string
		var priority Priority
		err = json.Unmarshal([]byte(`"medium"`), &priority)
		require.NoError(t, err)
		assert.Equal(t, PriorityMedium, priority)
	})

	t.Run("JSON unmarshaling from integer", func(t *testing.T) {
		var priority Priority
		err := json.Unmarshal([]byte(`3`), &priority)
		require.NoError(t, err)
		assert.Equal(t, PriorityHigh, priority)
	})

	t.Run("JSON unmarshaling invalid string", func(t *testing.T) {
		var priority Priority
		err := json.Unmarshal([]byte(`"invalid"`), &priority)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid priority")
	})

	t.Run("JSON unmarshaling invalid format", func(t *testing.T) {
		var priority Priority
		err := json.Unmarshal([]byte(`true`), &priority)
		assert.Error(t, err)
	})
}

func TestBaseEvent(t *testing.T) {
	t.Run("implements Event interface", func(t *testing.T) {
		now := time.Now()
		event := &BaseEvent{
			ID:          "event123",
			Type:        "TestEvent",
			AggregateId: "aggregate456",
			Ver:         1,
			OccurredAt:  now,
			Data:        []byte(`{"test": "data"}`),
		}

		// Test interface compliance
		var e Event = event
		assert.Equal(t, "TestEvent", e.EventType())
		assert.Equal(t, "event123", e.EventID())
		assert.Equal(t, "aggregate456", e.AggregateID())
		assert.Equal(t, int64(1), e.Version())
		assert.Equal(t, now, e.Timestamp())
	})
}

func TestHealthStatus(t *testing.T) {
	t.Run("healthy status", func(t *testing.T) {
		status := HealthStatus{Status: HealthStatusHealthy}
		assert.True(t, status.IsHealthy())
	})

	t.Run("unhealthy status", func(t *testing.T) {
		status := HealthStatus{Status: HealthStatusUnhealthy}
		assert.False(t, status.IsHealthy())
	})

	t.Run("degraded status", func(t *testing.T) {
		status := HealthStatus{Status: HealthStatusDegraded}
		assert.False(t, status.IsHealthy())
	})

	t.Run("with details", func(t *testing.T) {
		status := HealthStatus{
			Status:  HealthStatusHealthy,
			Message: "All systems operational",
			Details: map[string]string{
				"database": "connected",
				"cache":    "available",
			},
		}
		assert.True(t, status.IsHealthy())
		assert.Equal(t, "All systems operational", status.Message)
		assert.Equal(t, "connected", status.Details["database"])
	})
}

func TestMetadata(t *testing.T) {
	t.Run("get and set", func(t *testing.T) {
		metadata := make(Metadata)
		metadata.Set("key1", "value1")

		value, exists := metadata.Get("key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		metadata := make(Metadata)
		value, exists := metadata.Get("nonexistent")
		assert.False(t, exists)
		assert.Empty(t, value)
	})

	t.Run("has key", func(t *testing.T) {
		metadata := Metadata{"key1": "value1"}
		assert.True(t, metadata.Has("key1"))
		assert.False(t, metadata.Has("key2"))
	})

	t.Run("nil metadata", func(t *testing.T) {
		var metadata Metadata

		value, exists := metadata.Get("key")
		assert.False(t, exists)
		assert.Empty(t, value)

		assert.False(t, metadata.Has("key"))

		// Set on nil metadata should not panic but also not work
		metadata.Set("key", "value")
		assert.False(t, metadata.Has("key"))
	})

	t.Run("clone", func(t *testing.T) {
		original := Metadata{
			"key1": "value1",
			"key2": "value2",
		}

		cloned := original.Clone()
		assert.Equal(t, original, cloned)

		// Modify clone
		cloned.Set("key3", "value3")
		assert.True(t, cloned.Has("key3"))
		assert.False(t, original.Has("key3"))
	})

	t.Run("clone nil metadata", func(t *testing.T) {
		var metadata Metadata
		cloned := metadata.Clone()
		assert.Nil(t, cloned)
	})
}

func TestParseID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		id, err := ParseID("test123")
		require.NoError(t, err)
		assert.Equal(t, ID("test123"), id)
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := ParseID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestMustParseID(t *testing.T) {
	t.Run("valid ID", func(t *testing.T) {
		id := MustParseID("test123")
		assert.Equal(t, ID("test123"), id)
	})

	t.Run("invalid ID panics", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParseID("")
		})
	})
}

func TestToIntID(t *testing.T) {
	t.Run("valid integer ID", func(t *testing.T) {
		id := ID("123")
		intID, err := ToIntID(id)
		require.NoError(t, err)
		assert.Equal(t, int64(123), intID)
	})

	t.Run("invalid integer ID", func(t *testing.T) {
		id := ID("not-a-number")
		_, err := ToIntID(id)
		assert.Error(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		id := ID("")
		_, err := ToIntID(id)
		assert.Error(t, err)
	})
}

func TestFromIntID(t *testing.T) {
	t.Run("converts integer to ID", func(t *testing.T) {
		id := FromIntID(123)
		assert.Equal(t, ID("123"), id)
	})

	t.Run("handles zero", func(t *testing.T) {
		id := FromIntID(0)
		assert.Equal(t, ID("0"), id)
	})

	t.Run("handles negative", func(t *testing.T) {
		id := FromIntID(-123)
		assert.Equal(t, ID("-123"), id)
	})
}

// Generic Repository Tests
func TestRepository_Generics(t *testing.T) {
	t.Run("repository with test entity", func(t *testing.T) {
		// This test verifies that our generic repository interface
		// can be used with concrete types

		// Create a mock repository for testing
		repo := &mockRepository[*TestEntity]{}

		// Test entity
		entity := &TestEntity{
			BaseEntity: BaseEntity{
				ID:        ID("test123"),
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:   "Test Entity",
			Status: StatusActive,
		}

		// Test Create
		ctx := context.Background()
		err := repo.Create(ctx, entity)
		assert.NoError(t, err)
		assert.Equal(t, 1, repo.createCalled)

		// Test GetByID
		retrieved, err := repo.GetByID(ctx, entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity, retrieved)
		assert.Equal(t, 1, repo.getByIDCalled)

		// Test Update
		entity.Name = "Updated Name"
		err = repo.Update(ctx, entity)
		assert.NoError(t, err)
		assert.Equal(t, 1, repo.updateCalled)

		// Test Delete
		err = repo.Delete(ctx, entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, repo.deleteCalled)

		// Test List
		opts := NewListOptions().WithLimit(10)
		entities, err := repo.List(ctx, opts)
		assert.NoError(t, err)
		assert.NotNil(t, entities)
		assert.Equal(t, 1, repo.listCalled)

		// Test Count
		count, err := repo.Count(ctx, opts)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Equal(t, 1, repo.countCalled)
	})
}

// Mock repository for testing generics
type mockRepository[T Entity] struct {
	createCalled  int
	getByIDCalled int
	updateCalled  int
	deleteCalled  int
	listCalled    int
	countCalled   int
	entity        T
}

func (r *mockRepository[T]) Create(ctx context.Context, entity T) error {
	r.createCalled++
	r.entity = entity
	return nil
}

func (r *mockRepository[T]) GetByID(ctx context.Context, id ID) (T, error) {
	r.getByIDCalled++
	return r.entity, nil
}

func (r *mockRepository[T]) Update(ctx context.Context, entity T) error {
	r.updateCalled++
	r.entity = entity
	return nil
}

func (r *mockRepository[T]) Delete(ctx context.Context, id ID) error {
	r.deleteCalled++
	return nil
}

func (r *mockRepository[T]) List(ctx context.Context, opts ListOptions) ([]T, error) {
	r.listCalled++
	return []T{r.entity}, nil
}

func (r *mockRepository[T]) Count(ctx context.Context, opts ListOptions) (int64, error) {
	r.countCalled++
	return 1, nil
}

// Benchmark tests for performance validation
func BenchmarkID_String(b *testing.B) {
	id := ID("test123")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = id.String()
	}
}

func BenchmarkID_IsEmpty(b *testing.B) {
	id := ID("test123")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = id.IsEmpty()
	}
}

func BenchmarkID_JSON_Marshal(b *testing.B) {
	id := ID("test123")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(id)
	}
}

func BenchmarkID_JSON_Unmarshal(b *testing.B) {
	data := []byte(`"test123"`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var id ID
		_ = json.Unmarshal(data, &id)
	}
}

func BenchmarkBaseEntity_IncrementVersion(b *testing.B) {
	entity := &BaseEntity{Version: 1}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		entity.IncrementVersion()
	}
}

func BenchmarkBaseEntity_Touch(b *testing.B) {
	entity := &BaseEntity{Version: 1}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		entity.Touch()
	}
}

func BenchmarkNewListOptions(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewListOptions()
	}
}

func BenchmarkListOptions_WithFilter(b *testing.B) {
	opts := NewListOptions()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = opts.WithFilter("status", "active")
	}
}

func BenchmarkListOptions_WithFilter_Multiple(b *testing.B) {
	opts := NewListOptions()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = opts.WithFilter("status", "active").
			WithFilter("type", "user").
			WithFilter("created_date", "2024-01-15")
	}
}

func BenchmarkListOptions_GetPage(b *testing.B) {
	opts := ListOptions{Offset: 100, Limit: 25}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = opts.GetPage()
	}
}

func BenchmarkNewListResult(b *testing.B) {
	items := make([]string, 100)
	for i := range items {
		items[i] = "item"
	}
	opts := ListOptions{Offset: 100, Limit: 100}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = NewListResult(items, 1000, opts)
	}
}

func BenchmarkListResult_GetPageInfo(b *testing.B) {
	items := make([]string, 100)
	for i := range items {
		items[i] = "item"
	}
	result := NewListResult(items, 1000, ListOptions{Offset: 100, Limit: 100})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = result.GetPageInfo()
	}
}

func BenchmarkStatus_IsValid(b *testing.B) {
	status := StatusActive
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.IsValid()
	}
}

func BenchmarkStatus_String(b *testing.B) {
	status := StatusActive
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

func BenchmarkPriority_IsValid(b *testing.B) {
	priority := PriorityHigh
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = priority.IsValid()
	}
}

func BenchmarkPriority_String(b *testing.B) {
	priority := PriorityHigh
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = priority.String()
	}
}

func BenchmarkPriority_JSON_Marshal(b *testing.B) {
	priority := PriorityHigh
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(priority)
	}
}

func BenchmarkPriority_JSON_Unmarshal_String(b *testing.B) {
	data := []byte(`"high"`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var priority Priority
		_ = json.Unmarshal(data, &priority)
	}
}

func BenchmarkPriority_JSON_Unmarshal_Int(b *testing.B) {
	data := []byte(`3`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var priority Priority
		_ = json.Unmarshal(data, &priority)
	}
}

func BenchmarkMetadata_Get(b *testing.B) {
	metadata := Metadata{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = metadata.Get("key2")
	}
}

func BenchmarkMetadata_Set(b *testing.B) {
	metadata := make(Metadata)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		metadata.Set("key", "value")
	}
}

func BenchmarkMetadata_Has(b *testing.B) {
	metadata := Metadata{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = metadata.Has("key2")
	}
}

func BenchmarkMetadata_Clone(b *testing.B) {
	metadata := Metadata{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "value4",
		"key5": "value5",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = metadata.Clone()
	}
}

func BenchmarkMetadata_Clone_Large(b *testing.B) {
	metadata := make(Metadata, 100)
	for i := 0; i < 100; i++ {
		metadata[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = metadata.Clone()
	}
}

func BenchmarkParseID(b *testing.B) {
	idStr := "test123456"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseID(idStr)
	}
}

func BenchmarkToIntID(b *testing.B) {
	id := ID("123456")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ToIntID(id)
	}
}

func BenchmarkFromIntID(b *testing.B) {
	intID := int64(123456)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FromIntID(intID)
	}
}

func BenchmarkRepository_Create(b *testing.B) {
	repo := &mockRepository[*TestEntity]{}
	entity := &TestEntity{
		BaseEntity: BaseEntity{
			ID:        ID("test123"),
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:   "Test Entity",
		Status: StatusActive,
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = repo.Create(ctx, entity)
	}
}

func BenchmarkRepository_GetByID(b *testing.B) {
	repo := &mockRepository[*TestEntity]{}
	repo.entity = &TestEntity{
		BaseEntity: BaseEntity{ID: ID("test123")},
		Name:       "Test Entity",
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, ID("test123"))
	}
}

func BenchmarkRepository_List(b *testing.B) {
	repo := &mockRepository[*TestEntity]{}
	repo.entity = &TestEntity{
		BaseEntity: BaseEntity{ID: ID("test123")},
		Name:       "Test Entity",
	}
	ctx := context.Background()
	opts := NewListOptions().WithLimit(10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, opts)
	}
}
