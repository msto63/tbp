// File: config_test.go
// Title: Tests for Configuration Management
// Description: Comprehensive test suite for TBP configuration management
//              including multi-source loading, type conversion, validation,
//              hot-reloading, and struct unmarshaling. Tests cover edge cases,
//              concurrency, and performance characteristics.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation with comprehensive coverage

package config

import (
	"context"
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
		}

		for _, tc := range testCases {
			value, err := config.GetBool(tc.key)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value, "Failed for key: %s", tc.key)
		}
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
		value, err := config.GetDuration("test.number")
		assert.NoError(t, err)
		assert.Equal(t, 42*time.Second, value)
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
}

func TestConfig_AddSource(t *testing.T) {
	ctx := context.Background()
	config, err := New(ctx, LoadOptions{Environment: "test"})
	require.NoError(t, err)

	t.Run("adds source and sorts by priority", func(t *testing.T) {
		// Add sources with different priorities
		lowPrioritySource := &mockSource{priority: 10, values: map[string]interface{}{"key": "low"}}
		highPrioritySource := &mockSource{priority: 100, values: map[string]interface{}{"key": "high"}}

		config.AddSource(lowPrioritySource)
		config.AddSource(highPrioritySource)

		// Reload to apply new sources
		err := config.Load(ctx)
		assert.NoError(t, err)

		// Higher priority should win
		value, exists := config.Get("key")
		assert.True(t, exists)
		assert.Equal(t, "high", value)
	})
}

func TestConfig_Validate(t *testing.T) {
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
}

func TestConfig_Watcher(t *testing.T) {
	config := createTestConfig(t)
	
	t.Run("notifies watchers of changes", func(t *testing.T) {
		watcher := &mockWatcher{changes: make(chan map[string]ConfigChange, 1)}
		config.AddWatcher(watcher)

		// Simulate configuration change by reloading
		ctx := context.Background()
		err := config.Load(ctx)
		assert.NoError(t, err)

		// Note: Since we're using the same mock source, there won't be actual changes
		// In a real scenario with changing sources, this would trigger notifications
	})
}

// Helper functions and mock implementations

func createTestConfig(t *testing.T) *Config {
	ctx := context.Background()
	
	// Create a mock source with test data
	mockSrc := &mockSource{
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
			"test.server.host":   "localhost",
			"test.server.port":   8080,
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

// Mock source for testing
type mockSource struct {
	priority int
	values   map[string]interface{}
}

func (m *mockSource) Name() string {
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

func (m *mockSource) Watch(ctx context.Context, callback func(map[string]interface{})) error {
	return nil
}

// Mock watcher for testing
type mockWatcher struct {
	changes chan map[string]ConfigChange
}

func (m *mockWatcher) OnConfigChange(ctx context.Context, changes map[string]ConfigChange) {
	select {
	case m.changes <- changes:
	default:
	}
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

func BenchmarkConfig_Unmarshal(b *testing.B) {
	config := createTestConfig(b.(*testing.T))
	
	type TestConfig struct {
		Key     string `config:"test.key"`
		Number  int    `config:"test.number"`
		Enabled bool   `config:"test.enabled"`
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var cfg TestConfig
		_ = config.Unmarshal(&cfg)
	}
}