// File: config_test.go
// Title: Tests for Configuration Management
// Description: Comprehensive test suite for TBP configuration management
//              including multi-source loading, type conversion, validation,
//              hot-reloading, and struct unmarshaling. Tests cover edge cases,
//              concurrency, and performance characteristics.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.1
// Created: 2025-05-26
// Modified: 2025-05-27
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation with comprehensive coverage
// - 2025-05-27 v0.1.1: Updated for interface segregation and enhanced validation

package config

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates config with defaults", func(t *testing.T) {
		ctx := context.Background()
		
		config, err := New(ctx, LoadOptions{
			Environment: "test",
			EnvPrefix:   "TEST",
			Defaults: map[string]interface{}{
				"server.port": 8080,
				"debug":       false,
			},
		})
		
		require.NoError(t, err)
		assert.NotNil(t, config)
		
		// Check default values are loaded
		port, err := config.GetInt("server.port")
		assert.NoError(t, err)
		assert.Equal(t, 8080, port)
		
		debug, err := config.GetBool("debug")
		assert.NoError(t, err)
		assert.False(t, debug)
		
		// Check environment
		assert.Equal(t, "test", config.GetEnvironment())
	})

	t.Run("handles missing environment gracefully", func(t *testing.T) {
		ctx := context.Background()
		
		config, err := New(ctx, LoadOptions{
			Environment: "test",
			EnvPrefix:   "NONEXISTENT",
		})
		
		require.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("fails on missing required sources", func(t *testing.T) {
		ctx := context.Background()
		
		_, err := New(ctx, LoadOptions{
			Environment:   "test",
			ConfigPaths:   []string{"nonexistent.toml"},
			FailOnMissing: true,
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create file source")
	})

	t.Run("handles validation failure", func(t *testing.T) {
		ctx := context.Background()
		
		metadata := &Metadata{
			Name:        "test-config",
			Environment: "test",
			Fields: map[string]Field{
				"required.field": {
					Name:     "required.field",
					Required: true,
				},
			},
		}

		_, err := New(ctx, LoadOptions{
			Environment:   "test",
			Validation:    true,
			Metadata:      metadata,
		})
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration validation failed")
	})
}

func TestConfig_Get(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets existing value", func(t *testing.T) {
		value, exists := config.Get("test.key")
		assert.True(t, exists)
		assert.Equal(t, "test_value", value)
	})

	t.Run("returns false for non-existing key", func(t *testing.T) {
		_, exists := config.Get("nonexistent.key")
		assert.False(t, exists)
	})

	t.Run("gets all keys", func(t *testing.T) {
		keys := config.GetKeys()
		assert.Contains(t, keys, "test.key")
		assert.Contains(t, keys, "test.number")
	})

	t.Run("checks key existence", func(t *testing.T) {
		assert.True(t, config.HasKey("test.key"))
		assert.False(t, config.HasKey("nonexistent.key"))
	})
}

func TestConfig_GetString(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets string value", func(t *testing.T) {
		value, err := config.GetString("test.key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("converts non-string to string", func(t *testing.T) {
		value, err := config.GetString("test.number")
		assert.NoError(t, err)
		assert.Equal(t, "42", value)
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		_, err := config.GetString("missing.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_GetInt(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets integer value", func(t *testing.T) {
		value, err := config.GetInt("test.number")
		assert.NoError(t, err)
		assert.Equal(t, 42, value)
	})

	t.Run("converts string to integer", func(t *testing.T) {
		value, err := config.GetInt("test.string_number")
		assert.NoError(t, err)
		assert.Equal(t, 123, value)
	})

	t.Run("converts different int types", func(t *testing.T) {
		// Test int32
		value, err := config.GetInt("test.int32")
		assert.NoError(t, err)
		assert.Equal(t, 2147483647, value)

		// Test int64
		value, err = config.GetInt("test.int64")
		assert.NoError(t, err)
		assert.Equal(t, 1000, value) // Will be converted to int

		// Test float to int
		value, err = config.GetInt("test.float_to_int")
		assert.NoError(t, err)
		assert.Equal(t, 99, value)
	})

	t.Run("returns error for invalid integer", func(t *testing.T) {
		_, err := config.GetInt("test.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to int")
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		_, err := config.GetInt("missing.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_GetBool(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets boolean value", func(t *testing.T) {
		value, err := config.GetBool("test.enabled")
		assert.NoError(t, err)
		assert.True(t, value)
	})

	t.Run("converts string to boolean", func(t *testing.T) {
		testCases := []struct {
			key      string
			expected bool
		}{
			{"test.bool_true", true},
			{"test.bool_false", false},
			{"test.bool_yes", true},
			{"test.bool_no", false},
			{"test.bool_1", true},
			{"test.bool_0", false},
		}

		for _, tc := range testCases {
			value, err := config.GetBool(tc.key)
			assert.NoError(t, err, "Failed for key: %s", tc.key)
			assert.Equal(t, tc.expected, value, "Failed for key: %s", tc.key)
		}
	})

	t.Run("converts numeric to boolean", func(t *testing.T) {
		// Non-zero int should be true
		value, err := config.GetBool("test.number")
		assert.NoError(t, err)
		assert.True(t, value)

		// Zero should be false
		value, err = config.GetBool("test.zero")
		assert.NoError(t, err)
		assert.False(t, value)
	})

	t.Run("returns error for invalid boolean", func(t *testing.T) {
		_, err := config.GetBool("test.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to bool")
	})
}

func TestConfig_GetDuration(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets duration value", func(t *testing.T) {
		value, err := config.GetDuration("test.timeout")
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, value)
	})

	t.Run("converts numeric to duration", func(t *testing.T) {
		// Integer seconds
		value, err := config.GetDuration("test.number")
		assert.NoError(t, err)
		assert.Equal(t, 42*time.Second, value)

		// Float seconds
		value, err = config.GetDuration("test.float_seconds")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(2.5*float64(time.Second)), value)
	})

	t.Run("returns error for invalid duration", func(t *testing.T) {
		_, err := config.GetDuration("test.key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to duration")
	})
}

func TestConfig_WithDefault(t *testing.T) {
	config := createTestConfig(t)

	t.Run("returns value when key exists", func(t *testing.T) {
		value := config.GetStringWithDefault("test.key", "default")
		assert.Equal(t, "test_value", value)
	})

	t.Run("returns default when key missing", func(t *testing.T) {
		value := config.GetStringWithDefault("missing.key", "default")
		assert.Equal(t, "default", value)
	})

	t.Run("works with different types", func(t *testing.T) {
		intValue := config.GetIntWithDefault("missing.int", 999)
		assert.Equal(t, 999, intValue)

		boolValue := config.GetBoolWithDefault("missing.bool", true)
		assert.True(t, boolValue)

		durationValue := config.GetDurationWithDefault("missing.duration", 5*time.Minute)
		assert.Equal(t, 5*time.Minute, durationValue)
	})
}

func TestConfig_Unmarshal(t *testing.T) {
	config := createTestConfig(t)

	t.Run("unmarshals into struct", func(t *testing.T) {
		type TestConfig struct {
			Key     string `config:"test.key"`
			Number  int    `config:"test.number"`
			Enabled bool   `config:"test.enabled"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", cfg.Key)
		assert.Equal(t, 42, cfg.Number)
		assert.True(t, cfg.Enabled)
	})

	t.Run("handles default values from struct tags", func(t *testing.T) {
		type TestConfig struct {
			Missing string `config:"missing.key" default:"default_value"`
			Port    int    `config:"missing.port" default:"8080"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, "default_value", cfg.Missing)
		assert.Equal(t, 8080, cfg.Port)
	})

	t.Run("handles required fields", func(t *testing.T) {
		type TestConfig struct {
			Required string `config:"missing.required" required:"true"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required configuration field")
	})

	t.Run("skips ignored fields", func(t *testing.T) {
		type TestConfig struct {
			Included string `config:"test.key"`
			Ignored  string `config:"-"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", cfg.Included)
		assert.Empty(t, cfg.Ignored)
	})

	t.Run("handles nested structs", func(t *testing.T) {
		type ServerConfig struct {
			Host string `config:"host"`
			Port int    `config:"port"`
		}

		type TestConfig struct {
			Server ServerConfig `config:"test.server"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})

	t.Run("handles special types", func(t *testing.T) {
		type TestConfig struct {
			Timeout   time.Duration `config:"test.timeout"`
			CreatedAt time.Time     `config:"test.created_at"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, 30*time.Second, cfg.Timeout)
		assert.False(t, cfg.CreatedAt.IsZero())
	})

	t.Run("handles numeric type conversions", func(t *testing.T) {
		type TestConfig struct {
			Int8Value   int8    `config:"test.small_int"`
			Uint32Value uint32  `config:"test.unsigned"`
			Float32Val  float32 `config:"test.float_val"`
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, int8(127), cfg.Int8Value)
		assert.Equal(t, uint32(12345), cfg.Uint32Value)
		assert.Equal(t, float32(3.14), cfg.Float32Val)
	})

	t.Run("returns error for non-pointer", func(t *testing.T) {
		type TestConfig struct {
			Key string `config:"test.key"`
		}

		var cfg TestConfig
		err := config.Unmarshal(cfg) // Not a pointer
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")
	})

	t.Run("returns error for nil pointer", func(t *testing.T) {
		type TestConfig struct {
			Key string `config:"test.key"`
		}

		var cfg *TestConfig
		err := config.Unmarshal(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a non-nil pointer")
	})

	t.Run("handles overflow errors", func(t *testing.T) {
		type TestConfig struct {
			SmallInt int8 `config:"test.large_number"` // Will overflow int8
		}

		var cfg TestConfig
		err := config.Unmarshal(&cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "overflows")
	})
}

func TestConfig_AddSource(t *testing.T) {
	ctx := context.Background()
	config, err := New(ctx, LoadOptions{Environment: "test"})
	require.NoError(t, err)

	t.Run("adds source and sorts by priority", func(t *testing.T) {
		// Add sources with different priorities
		lowPrioritySource := &mockSource{priority: 10, values: map[string]interface{}{"key": "low"}}
		highPrioritySource := &mockSource{priority: 100, values: map[string]interface{}{"key": "high"}}

		err := config.AddSource(lowPrioritySource)
		assert.NoError(t, err)
		
		err = config.AddSource(highPrioritySource)
		assert.NoError(t, err)

		// Reload to apply new sources
		err = config.Load(ctx)
		assert.NoError(t, err)

		// Higher priority should win
		value, exists := config.Get("key")
		assert.True(t, exists)
		assert.Equal(t, "high", value)
	})

	t.Run("validates source before adding", func(t *testing.T) {
		invalidSource := &mockValidatableSource{
			mockSource: mockSource{priority: 50},
			valid:      false,
		}

		err := config.AddSource(invalidSource)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("rejects nil source", func(t *testing.T) {
		err := config.AddSource(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source cannot be nil")
	})
}

func TestConfig_Validation(t *testing.T) {
	t.Run("validates successfully with all required fields", func(t *testing.T) {
		config := createTestConfigWithMetadata(t)
		err := config.Validate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("fails validation with missing required field", func(t *testing.T) {
		config := createTestConfigWithMissingRequired(t)
		err := config.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required configuration field")
	})

	t.Run("validates field types", func(t *testing.T) {
		metadata := &Metadata{
			Name:        "test-config",
			Environment: "test",
			Fields: map[string]Field{
				"test.number": {
					Name: "test.number",
					Type: "string", // Wrong type - should be int
				},
			},
		}

		config := createTestConfigWithCustomMetadata(t, metadata)
		err := config.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has type")
	})

	t.Run("validates field ranges", func(t *testing.T) {
		metadata := &Metadata{
			Name:        "test-config",
			Environment: "test",
			Fields: map[string]Field{
				"test.number": {
					Name:     "test.number",
					Type:     "integer",
					MinValue: 50,  // test.number is 42, should fail
					MaxValue: 100,
				},
			},
		}

		config := createTestConfigWithCustomMetadata(t, metadata)
		err := config.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "below minimum")
	})

	t.Run("validates enum values", func(t *testing.T) {
		metadata := &Metadata{
			Name:        "test-config",
			Environment: "test",
			Fields: map[string]Field{
				"test.key": {
					Name: "test.key",
					Type: "string",
					Enum: []string{"production", "staging"}, // test.key is "test_value"
				},
			},
		}

		config := createTestConfigWithCustomMetadata(t, metadata)
		err := config.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not in allowed enum values")
	})

	t.Run("runs custom validators", func(t *testing.T) {
		config := createTestConfig(t)
		
		// Add a custom validator that fails
		config.AddValidator(func(key, value string) error {
			if key == "test.key" && value == "test_value" {
				return fmt.Errorf("custom validation failed")
			}
			return nil
		})

		err := config.Validate(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "custom validation failed")
	})
}

func TestConfig_Watcher(t *testing.T) {
	config := createTestConfig(t)
	
	t.Run("notifies watchers of changes", func(t *testing.T) {
		watcher := &mockWatcher{changes: make(chan map[string]ConfigChange, 1)}
		config.AddWatcher(watcher)

		// Add a new source to trigger changes
		newSource := &mockSource{
			priority: 200, // Higher priority
			values: map[string]interface{}{
				"test.key": "updated_value",
				"new.key":  "new_value",
			},
		}

		err := config.AddSource(newSource)
		require.NoError(t, err)

		// Reload to trigger change detection
		ctx := context.Background()
		err = config.Load(ctx)
		require.NoError(t, err)

		// Wait for change notification
		select {
		case changes := <-watcher.changes:
			assert.Contains(t, changes, "test.key")
			assert.Contains(t, changes, "new.key")
			
			// Check change types
			testKeyChange := changes["test.key"]
			assert.Equal(t, ChangeActionUpdate, testKeyChange.Action)
			assert.Equal(t, "test_value", testKeyChange.OldValue)
			assert.Equal(t, "updated_value", testKeyChange.NewValue)
			
			newKeyChange := changes["new.key"]
			assert.Equal(t, ChangeActionAdd, newKeyChange.Action)
			assert.Equal(t, "new_value", newKeyChange.NewValue)
			
		case <-time.After(1 * time.Second):
			t.Fatal("Did not receive change notification")
		}
	})

	t.Run("removes watchers", func(t *testing.T) {
		watcher := &mockWatcher{changes: make(chan map[string]ConfigChange, 1)}
		config.AddWatcher(watcher)
		config.RemoveWatcher(watcher)

		// Add source to trigger potential changes
		newSource := &mockSource{
			priority: 150,
			values:   map[string]interface{}{"test.key": "another_value"},
		}
		config.AddSource(newSource)
		config.Load(context.Background())

		// Should not receive notification
		select {
		case <-watcher.changes:
			t.Fatal("Should not receive notification after watcher removal")
		case <-time.After(100 * time.Millisecond):
			// Expected - no notification
		}
	})
}

func TestConfig_Sources(t *testing.T) {
	config := createTestConfig(t)

	t.Run("gets source information", func(t *testing.T) {
		sources := config.GetSources()
		assert.NotEmpty(t, sources)

		// Check that we have our mock source
		found := false
		for _, source := range sources {
			if source.Name == "mock" {
				found = true
				assert.Equal(t, 50, source.Priority)
				assert.False(t, source.Watchable)   // mockSource doesn't implement WatchableSource
				assert.False(t, source.Writable)    // mockSource doesn't implement WritableSource
				assert.False(t, source.Validatable) // mockSource doesn't implement ValidatableSource
			}
		}
		assert.True(t, found, "Mock source not found in sources")
	})

	t.Run("writes to writable source", func(t *testing.T) {
		writableSource := &mockWritableSource{
			mockSource: mockSource{
				name:     "writable",
				priority: 75,
				values:   make(map[string]interface{}),
			},
		}

		err := config.AddSource(writableSource)
		require.NoError(t, err)

		values := map[string]interface{}{
			"new.key": "new.value",
		}

		err = config.WriteToSource("writable", values)
		assert.NoError(t, err)
		assert.Equal(t, "new.value", writableSource.writtenValues["new.key"])
	})

	t.Run("fails to write to non-writable source", func(t *testing.T) {
		values := map[string]interface{}{"key": "value"}
		err := config.WriteToSource("mock", values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not writable")
	})

	t.Run("fails to write to non-existent source", func(t *testing.T) {
		values := map[string]interface{}{"key": "value"}
		err := config.WriteToSource("nonexistent", values)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConfig_Summary(t *testing.T) {
	config := createTestConfigWithMetadata(t)

	summary := config.Summary()
	assert.Equal(t, "test", summary.Environment)
	assert.Greater(t, summary.TotalKeys, 0)
	assert.Greater(t, summary.TotalSources, 0)
	assert.GreaterOrEqual(t, summary.TotalWatchers, 0)
	assert.NotEmpty(t, summary.Sources)
	assert.Contains(t, summary.RequiredFields, "required.field")
}

func TestConfig_Metadata(t *testing.T) {
	config := createTestConfig(t)

	t.Run("adds and gets field metadata", func(t *testing.T) {
		field := Field{
			Name:        "new.field",
			Type:        "string",
			Required:    true,
			Description: "A new field for testing",
		}

		config.AddFieldMetadata("new.field", field)

		retrievedField, exists := config.GetFieldMetadata("new.field")
		assert.True(t, exists)
		assert.Equal(t, field, retrievedField)
	})

	t.Run("removes field metadata", func(t *testing.T) {
		field := Field{Name: "temp.field"}
		config.AddFieldMetadata("temp.field", field)

		_, exists := config.GetFieldMetadata("temp.field")
		assert.True(t, exists)

		config.RemoveFieldMetadata("temp.field")

		_, exists = config.GetFieldMetadata("temp.field")
		assert.False(t, exists)
	})
}

func TestConfig_Close(t *testing.T) {
	config := createTestConfig(t)

	err := config.Close()
	assert.NoError(t, err)

	// Config should be empty after close
	keys := config.GetKeys()
	assert.Empty(t, keys)
}

// Helper functions and mock implementations

func createTestConfig(t *testing.T) *Config {
	ctx := context.Background()
	
	// Create a mock source with test data
	mockSrc := &mockSource{
		name:     "mock",
		priority: 50,
		values: map[string]interface{}{
			"test.key":           "test_value",
			"test.number":        42,
			"test.string_number": "123",
			"test.enabled":       true,
			"test.timeout":       "30s",
			"test.bool_true":     "true",
			"test.bool_false":    "false",
			"test.bool_yes":      "yes",
			"test.bool_no":       "no",
			"test.bool_1":        "1",
			"test.bool_0":        "0",
			"test.server.host":   "localhost",
			"test.server.port":   8080,
			"test.int32":         int32(2147483647),
			"test.int64":         int64(1000),
			"test.float_to_int":  99.9,
			"test.zero":          0,
			"test.float_seconds": 2.5,
			"test.created_at":    "2024-01-15T10:30:00Z",
			"test.small_int":     127,
			"test.unsigned":      uint32(12345),
			"test.float_val":     3.14,
			"test.large_number":  300, // Will overflow int8
		},
	}

	config, err := New(ctx, LoadOptions{
		Environment: "test",
		Sources:     []Source{mockSrc},
	})
	require.NoError(t, err)
	
	return config
}

func createTestConfigWithMetadata(t *testing.T) *Config {
	ctx := context.Background()
	
	metadata := &Metadata{
		Name:        "test-config",
		Environment: "test",
		Fields: map[string]Field{
			"required.field": {
				Name:     "required.field",
				Required: true,
			},
		},
	}

	mockSrc := &mockSource{
		name:     "mock",
		priority: 50,
		values: map[string]interface{}{
			"required.field": "present",
		},
	}

	config, err := New(ctx, LoadOptions{
		Environment: "test",
		Sources:     []Source{mockSrc},
		Metadata:    metadata,
	})
	require.NoError(t, err)
	
	return config
}

func createTestConfigWithMissingRequired(t *testing.T) *Config {
	ctx := context.Background()
	
	metadata := &Metadata{
		Name:        "test-config",
		Environment: "test",
		Fields: map[string]Field{
			"required.field": {
				Name:     "required.field",
				Required: true,
			},
		},
	}

	mockSrc := &mockSource{
		name:     "mock",
		priority: 50,
		values:   map[string]interface{}{}, // Missing required field
	}

	config, err := New(ctx, LoadOptions{
		Environment: "test",
		Sources:     []Source{mockSrc},
		Metadata:    metadata,
		Validation:  false, // Don't validate during creation
	})
	require.NoError(t, err)
	
	return config
}

func createTestConfigWithCustomMetadata(t *testing.T, metadata *Metadata) *Config {
	ctx := context.Background()
	
	mockSrc := &mockSource{
		name:     "mock",
		priority: 50,
		values: map[string]interface{}{
			"test.key":    "test_value",
			"test.number": 42,
		},
	}

	config, err := New(ctx, LoadOptions{
		Environment: "test",
		Sources:     []Source{mockSrc},
		Metadata:    metadata,
		Validation:  false, // Don't validate during creation
	})
	require.NoError(t, err)
	
	return config
}

// Mock source for testing
type mockSource struct {
	name     string
	priority int
	values   map[string]interface{}
}

func (m *mockSource) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
}

func (m *mockSource) Priority() int {
	return m.priority
}

func (m *mockSource) Load(ctx context.Context) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range m.values {
		result[k] = v
	}
	return result, nil
}

// Mock validatable source for testing
type mockValidatableSource struct {
	mockSource
	valid bool
}

func (m *mockValidatableSource) Validate() error {
	if !m.valid {
		return fmt.Errorf("mock validation error")
	}
	return nil
}

// Mock writable source for testing
type mockWritableSource struct {
	mockSource
	writtenValues map[string]interface{}
}

func (m *mockWritableSource) WriteConfig(values map[string]interface{}) error {
	if m.writtenValues == nil {
		m.writtenValues = make(map[string]interface{})
	}
	for k, v := range values {
		m.writtenValues[k] = v
	}
	return nil
}

// Mock watchable source for testing
type mockWatchableSource struct {
	mockSource
	callback func(map[string]interface{})
}

func (m *mockWatchableSource) Watch(ctx context.Context, callback func(map[string]interface{})) error {
	m.callback = callback
	return nil
}

func (m *mockWatchableSource) TriggerChange(values map[string]interface{}) {
	if m.callback != nil {
		m.callback(values)
	}
}

// Mock watcher for testing
type mockWatcher struct {
	changes chan map[string]ConfigChange
}

func (m *mockWatcher) OnConfigChange(ctx context.Context, changes map[string]ConfigChange) {
	select {
	case m.changes <- changes:
	default:
		// Channel full, drop the change
	}
}

// Test hot-reloading functionality
func TestConfig_HotReload(t *testing.T) {
	t.Run("starts watching when enabled", func(t *testing.T) {
		ctx := context.Background()
		
		watchableSource := &mockWatchableSource{
			mockSource: mockSource{
				name:     "watchable",
				priority: 50,
				values:   map[string]interface{}{"test.key": "initial"},
			},
		}

		config, err := New(ctx, LoadOptions{
			Environment: "test",
			Sources:     []Source{watchableSource},
			HotReload:   true,
		})
		require.NoError(t, err)

		// Verify initial value
		value, exists := config.Get("test.key")
		assert.True(t, exists)
		assert.Equal(t, "initial", value)

		// Add watcher to detect changes
		watcher := &mockWatcher{changes: make(chan map[string]ConfigChange, 1)}
		config.AddWatcher(watcher)

		// Simulate source change
		watchableSource.values["test.key"] = "updated"
		watchableSource.TriggerChange(watchableSource.values)

		// Wait for change notification
		select {
		case changes := <-watcher.changes:
			assert.Contains(t, changes, "test.key")
			change := changes["test.key"]
			assert.Equal(t, ChangeActionUpdate, change.Action)
			assert.Equal(t, "updated", change.NewValue)
		case <-time.After(1 * time.Second):
			t.Fatal("Did not receive change notification")
		}
	})
}

// Test concurrent access
func TestConfig_ConcurrentAccess(t *testing.T) {
	config := createTestConfig(t)
	
	t.Run("concurrent reads", func(t *testing.T) {
		const numGoroutines = 100
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()
				
				// Read various values
				_, _ = config.GetString("test.key")
				_, _ = config.GetInt("test.number")
				_, _ = config.GetBool("test.enabled")
				_ = config.GetAll()
				_ = config.GetKeys()
			}()
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent reads timed out")
			}
		}
	})
	
	t.Run("concurrent writes", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()
				
				// Add sources concurrently
				source := &mockSource{
					name:     fmt.Sprintf("concurrent-%d", id),
					priority: 60 + id,
					values: map[string]interface{}{
						fmt.Sprintf("key.%d", id): fmt.Sprintf("value-%d", id),
					},
				}
				
				_ = config.AddSource(source)
				_ = config.Load(context.Background())
			}(i)
		}
		
		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent writes timed out")
			}
		}
	})
}

// Test error conditions
func TestConfig_ErrorConditions(t *testing.T) {
	config := createTestConfig(t)
	
	t.Run("handles source load errors", func(t *testing.T) {
		errorSource := &mockErrorSource{
			mockSource: mockSource{name: "error", priority: 60},
			loadError:  fmt.Errorf("mock load error"),
		}
		
		err := config.AddSource(errorSource)
		require.NoError(t, err)
		
		err = config.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock load error")
	})
	
	t.Run("handles watcher panics gracefully", func(t *testing.T) {
		panicWatcher := &mockPanicWatcher{}
		config.AddWatcher(panicWatcher)
		
		// Add source to trigger change
		source := &mockSource{
			name:     "trigger",
			priority: 70,
			values:   map[string]interface{}{"new.key": "new.value"},
		}
		
		err := config.AddSource(source)
		require.NoError(t, err)
		
		// This should not panic the main thread
		err = config.Load(context.Background())
		assert.NoError(t, err)
		
		// Give watcher time to panic
		time.Sleep(100 * time.Millisecond)
		
		// Config should still be functional
		value, exists := config.Get("new.key")
		assert.True(t, exists)
		assert.Equal(t, "new.value", value)
	})
}

// Mock error source for testing error conditions
type mockErrorSource struct {
	mockSource
	loadError error
}

func (m *mockErrorSource) Load(ctx context.Context) (map[string]interface{}, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	return m.mockSource.Load(ctx)
}

// Mock panic watcher for testing error handling
type mockPanicWatcher struct{}

func (m *mockPanicWatcher) OnConfigChange(ctx context.Context, changes map[string]ConfigChange) {
	panic("mock panic in watcher")
}

// Benchmark tests for performance validation
func BenchmarkConfig_Get(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.Get("test.key")
	}
}

func BenchmarkConfig_GetString(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetString("test.key")
	}
}

func BenchmarkConfig_GetInt(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetInt("test.number")
	}
}

func BenchmarkConfig_GetBool(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetBool("test.enabled")
	}
}

func BenchmarkConfig_GetDuration(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetDuration("test.timeout")
	}
}

func BenchmarkConfig_Unmarshal(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	type TestConfig struct {
		Key     string `config:"test.key"`
		Number  int    `config:"test.number"`
		Enabled bool   `config:"test.enabled"`
		Timeout time.Duration `config:"test.timeout"`
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg TestConfig
		_ = config.Unmarshal(&cfg)
	}
}

func BenchmarkConfig_Load(b *testing.B) {
	ctx := context.Background()
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.Load(ctx)
	}
}

func BenchmarkConfig_Validate(b *testing.B) {
	ctx := context.Background()
	config := createTestConfigWithMetadata(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.Validate(ctx)
	}
}

func BenchmarkConfig_AddSource(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		source := &mockSource{
			name:     fmt.Sprintf("bench-%d", i),
			priority: 80 + i,
			values:   map[string]interface{}{"key": fmt.Sprintf("value-%d", i)},
		}
		b.StartTimer()
		
		_ = config.AddSource(source)
	}
}

func BenchmarkConfig_GetAll(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.GetAll()
	}
}

func BenchmarkConfig_GetKeys(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.GetKeys()
	}
}

func BenchmarkConfig_Summary(b *testing.B) {
	config := createTestConfigWithMetadata(b.(*testing.T))
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = config.Summary()
	}
}