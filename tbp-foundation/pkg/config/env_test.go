// File: env_test.go
// Title: Tests for Environment Variable Configuration
// Description: Comprehensive test suite for environment variable configuration
//              source including type conversion, prefix filtering, key mapping,
//              validation, and edge cases. Tests performance characteristics
//              and concurrent access patterns.
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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnvSource(t *testing.T) {
	t.Run("creates env source with defaults", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{})
		require.NoError(t, err)
		assert.Equal(t, "TBP_", envSrc.prefix)
		assert.Equal(t, "_", envSrc.separator)
		assert.Equal(t, "env:TBP", envSrc.Name())
		assert.Equal(t, 100, envSrc.Priority())
	})

	t.Run("creates env source with custom options", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix:        "MYAPP",
			Separator:     "__",
			CaseSensitive: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "MYAPP__", envSrc.prefix)
		assert.Equal(t, "__", envSrc.separator)
		assert.True(t, envSrc.caseSensitive)
	})
}

func TestEnvSource_Load(t *testing.T) {
	// Setup test environment variables
	testEnvVars := map[string]string{
		"TEST_STRING_VALUE":    "hello_world",
		"TEST_INT_VALUE":       "42",
		"TEST_FLOAT_VALUE":     "3.14",
		"TEST_BOOL_TRUE":       "true",
		"TEST_BOOL_FALSE":      "false",
		"TEST_BOOL_YES":        "yes",
		"TEST_BOOL_NO":         "no",
		"TEST_DURATION_VALUE":  "30s",
		"TEST_SLICE_VALUE":     "item1,item2,item3",
		"TEST_NESTED_KEY":      "nested_value",
		"OTHER_PREFIX_KEY":     "should_be_ignored",
		"TEST_EMPTY_VALUE":     "",
	}

	// Set environment variables
	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}
	defer func() {
		// Clean up
		for key := range testEnvVars {
			os.Unsetenv(key)
		}
	}()

	t.Run("loads environment variables with prefix", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix: "TEST",
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		// Check that prefixed variables are loaded
		assert.Equal(t, "hello_world", values["string.value"])
		assert.Equal(t, 42, values["int.value"])
		assert.Equal(t, 3.14, values["float.value"])
		assert.Equal(t, true, values["bool.true"])
		assert.Equal(t, false, values["bool.false"])
		assert.Equal(t, true, values["bool.yes"])
		assert.Equal(t, false, values["bool.no"])
		assert.Equal(t, 30*time.Second, values["duration.value"])
		assert.Equal(t, []string{"item1", "item2", "item3"}, values["slice.value"])
		assert.Equal(t, "nested_value", values["nested.key"])
		assert.Equal(t, "", values["empty.value"])

		// Check that non-prefixed variables are ignored
		_, exists := values["prefix.key"]
		assert.False(t, exists)
	})

	t.Run("handles case insensitive matching", func(t *testing.T) {
		os.Setenv("test_lower_case", "lowercase")
		defer os.Unsetenv("test_lower_case")

		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix:        "TEST",
			CaseSensitive: false,
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		assert.Equal(t, "lowercase", values["lower.case"])
	})

	t.Run("uses custom key mappings", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix: "TEST",
			KeyMapping: map[string]string{
				"TEST_STRING_VALUE": "custom.mapped.key",
			},
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		assert.Equal(t, "hello_world", values["custom.mapped.key"])
		_, exists := values["string.value"]
		assert.False(t, exists) // Original key should not exist
	})

	t.Run("uses type hints", func(t *testing.T) {
		os.Setenv("TEST_TYPE_HINT", "123.45")
		defer os.Unsetenv("TEST_TYPE_HINT")

		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix: "TEST",
			TypeHints: map[string]string{
				"type.hint": "int",
			},
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		// Should be converted to int despite decimal point
		assert.Equal(t, 123, values["type.hint"])
	})
}

func TestEnvSource_TypeConversion(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("converts boolean values", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected bool
		}{
			{"true", true},
			{"false", false},
			{"TRUE", true},
			{"FALSE", false},
			{"yes", true},
			{"no", false},
			{"YES", true},
			{"NO", false},
			{"1", true},
			{"0", false},
			{"on", true},
			{"off", false},
			{"enable", true},
			{"disable", false},
			{"enabled", true},
			{"disabled", false},
			{"", false},
		}

		for _, tc := range testCases {
			result, err := envSrc.parseBool(tc.input)
			assert.NoError(t, err, "Failed for input: %s", tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})

	t.Run("handles invalid boolean values", func(t *testing.T) {
		invalidValues := []string{"maybe", "sometimes", "invalid", "2", "-1"}
		
		for _, value := range invalidValues {
			_, err := envSrc.parseBool(value)
			assert.Error(t, err, "Should fail for input: %s", value)
		}
	})

	t.Run("converts string slices", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []string
		}{
			{"", []string{}},
			{"single", []string{"single"}},
			{"one,two,three", []string{"one", "two", "three"}},
			{"  spaced  ,  values  ", []string{"spaced", "values"}},
			{"trailing,comma,", []string{"trailing", "comma", ""}},
		}

		for _, tc := range testCases {
			result := envSrc.parseStringSlice(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})

	t.Run("converts int slices", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []int
			hasError bool
		}{
			{"", []int{}, false},
			{"42", []int{42}, false},
			{"1,2,3", []int{1, 2, 3}, false},
			{" 10 , 20 , 30 ", []int{10, 20, 30}, false},
			{"1,invalid,3", nil, true},
		}

		for _, tc := range testCases {
			result, err := envSrc.parseIntSlice(tc.input)
			if tc.hasError {
				assert.Error(t, err, "Should fail for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Should succeed for input: %s", tc.input)
				assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
			}
		}
	})

	t.Run("auto-converts values", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected interface{}
		}{
			{"true", true},
			{"false", false},
			{"42", 42},
			{"3.14", 3.14},
			{"30s", 30 * time.Second},
			{"item1,item2", []string{"item1", "item2"}},
			{"plain string", "plain string"},
			{"", ""},
		}

		for _, tc := range testCases {
			result := envSrc.autoConvertValue(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})
}

func TestEnvSource_KeyMapping(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("converts env key to config key", func(t *testing.T) {
		testCases := []struct {
			envKey    string
			configKey string
		}{
			{"TEST_SIMPLE", "simple"},
			{"TEST_NESTED_KEY", "nested.key"},
			{"TEST_DEEP_NESTED_KEY", "deep.nested.key"},
			{"TEST_WITH_NUMBERS_123", "with.numbers.123"},
		}

		for _, tc := range testCases {
			result := envSrc.envKeyToConfigKey(tc.envKey)
			assert.Equal(t, tc.configKey, result, "Failed for env key: %s", tc.envKey)
		}
	})

	t.Run("converts config key to env key", func(t *testing.T) {
		testCases := []struct {
			configKey string
			envKey    string
		}{
			{"simple", "TEST_SIMPLE"},
			{"nested.key", "TEST_NESTED_KEY"},
			{"deep.nested.key", "TEST_DEEP_NESTED_KEY"},
		}

		for _, tc := range testCases {
			result := envSrc.configKeyToEnvKey(tc.configKey)
			assert.Equal(t, tc.envKey, result, "Failed for config key: %s", tc.configKey)
		}
	})

	t.Run("adds and uses custom key mappings", func(t *testing.T) {
		envSrc.AddKeyMapping("CUSTOM_ENV_VAR", "custom.config.key")
		
		mappings := envSrc.GetKeyMappings()
		assert.Equal(t, "custom.config.key", mappings["CUSTOM_ENV_VAR"])
		
		// Test reverse lookup
		result := envSrc.configKeyToEnvKey("custom.config.key")
		assert.Equal(t, "CUSTOM_ENV_VAR", result)
	})

	t.Run("adds and uses type hints", func(t *testing.T) {
		envSrc.AddTypeHint("test.key", "duration")
		
		hints := envSrc.GetTypeHints()
		assert.Equal(t, "duration", hints["test.key"])
	})
}

func TestEnvSource_Validation(t *testing.T) {
	// Setup test environment
	os.Setenv("TEST_REQUIRED_KEY", "present")
	defer os.Unsetenv("TEST_REQUIRED_KEY")

	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("validates required environment variables", func(t *testing.T) {
		err := envSrc.ValidateEnvironment([]string{"required.key"})
		assert.NoError(t, err)
	})

	t.Run("fails validation for missing required variables", func(t *testing.T) {
		err := envSrc.ValidateEnvironment([]string{"missing.key"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required environment variables")
		assert.Contains(t, err.Error(), "TEST_MISSING_KEY")
	})

	t.Run("validates multiple required variables", func(t *testing.T) {
		err := envSrc.ValidateEnvironment([]string{"required.key", "missing.key"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TEST_MISSING_KEY")
		assert.NotContains(t, err.Error(), "TEST_REQUIRED_KEY")
	})
}

func TestEnvSource_Utilities(t *testing.T) {
	// Setup test environment
	testVars := map[string]string{
		"TEST_UTIL_KEY1": "value1",
		"TEST_UTIL_KEY2": "value2",
		"OTHER_KEY":      "other_value",
	}
	
	for key, value := range testVars {
		os.Setenv(key, value)
	}
	defer func() {
		for key := range testVars {
			os.Unsetenv(key)
		}
	}()

	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("lists environment variables with prefix", func(t *testing.T) {
		vars := envSrc.ListEnvironmentVariables()
		
		assert.Equal(t, "value1", vars["TEST_UTIL_KEY1"])
		assert.Equal(t, "value2", vars["TEST_UTIL_KEY2"])
		_, exists := vars["OTHER_KEY"]
		assert.False(t, exists)
	})

	t.Run("gets environment variable name for config key", func(t *testing.T) {
		envVarName := envSrc.GetEnvironmentVariableName("util.key1")
		assert.Equal(t, "TEST_UTIL_KEY1", envVarName)
	})

	t.Run("checks if config key is set", func(t *testing.T) {
		assert.True(t, envSrc.IsSet("util.key1"))
		assert.False(t, envSrc.IsSet("nonexistent.key"))
	})

	t.Run("gets raw string value", func(t *testing.T) {
		value, exists := envSrc.GetRaw("util.key1")
		assert.True(t, exists)
		assert.Equal(t, "value1", value)

		_, exists = envSrc.GetRaw("nonexistent.key")
		assert.False(t, exists)
	})

	t.Run("sets and unsets environment variables", func(t *testing.T) {
		// Set
		err := envSrc.SetEnvironmentVariable("test.set.key", "test_value")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", os.Getenv("TEST_TEST_SET_KEY"))

		// Unset
		err = envSrc.UnsetEnvironmentVariable("test.set.key")
		assert.NoError(t, err)
		assert.Empty(t, os.Getenv("TEST_TEST_SET_KEY"))
	})
}

func TestEnvSource_TypeHints(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{
		Prefix: "TEST",
		TypeHints: map[string]string{
			"duration.key": "duration",
			"slice.key":    "stringslice",
			"int.key":      "int",
		},
	})
	require.NoError(t, err)

	t.Run("uses type hints for conversion", func(t *testing.T) {
		testCases := []struct {
			value    string
			typeHint string
			expected interface{}
			hasError bool
		}{
			{"30s", "duration", 30 * time.Second, false},
			{"item1,item2,item3", "stringslice", []string{"item1", "item2", "item3"}, false},
			{"42", "int", 42, false},
			{"3.14", "float", 3.14, false},
			{"true", "bool", true, false},
			{"invalid", "int", nil, true},
			{"value", "unsupported", nil, true},
		}

		for _, tc := range testCases {
			result, err := envSrc.convertByType(tc.value, tc.typeHint)
			if tc.hasError {
				assert.Error(t, err, "Should fail for value %s with type %s", tc.value, tc.typeHint)
			} else {
				assert.NoError(t, err, "Should succeed for value %s with type %s", tc.value, tc.typeHint)
				assert.Equal(t, tc.expected, result, "Failed for value %s with type %s", tc.value, tc.typeHint)
			}
		}
	})
}

func TestEnvSource_CaseSensitivity(t *testing.T) {
	// Setup mixed case environment variables
	os.Setenv("TEST_UPPER_CASE", "upper")
	os.Setenv("test_lower_case", "lower")
	defer func() {
		os.Unsetenv("TEST_UPPER_CASE")
		os.Unsetenv("test_lower_case")
	}()

	t.Run("case sensitive matching", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix:        "TEST",
			CaseSensitive: true,
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		// Only exact prefix match should be included
		_, upperExists := values["upper.case"]
		assert.True(t, upperExists)

		_, lowerExists := values["lower.case"]
		assert.False(t, lowerExists) // Should not match "test_" prefix
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix:        "TEST",
			CaseSensitive: false,
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		// Both should match
		_, upperExists := values["upper.case"]
		assert.True(t, upperExists)

		_, lowerExists := values["lower.case"]
		assert.True(t, lowerExists)
	})
}

func TestEnvSource_CustomSeparator(t *testing.T) {
	// Setup test environment variable
	os.Setenv("TEST__CUSTOM__SEPARATOR", "custom_sep_value")
	defer os.Unsetenv("TEST__CUSTOM__SEPARATOR")

	t.Run("uses custom separator", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix:    "TEST",
			Separator: "__",
		})
		require.NoError(t, err)

		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		assert.Equal(t, "custom_sep_value", values["custom.separator"])
	})
}

// Benchmark tests for performance validation
func BenchmarkEnvSource_Load(b *testing.B) {
	// Setup test environment variables
	for i := 0; i < 100; i++ {
		os.Setenv(fmt.Sprintf("BENCH_KEY_%d", i), fmt.Sprintf("value_%d", i))
	}
	defer func() {
		for i := 0; i < 100; i++ {
			os.Unsetenv(fmt.Sprintf("BENCH_KEY_%d", i))
		}
	}()

	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "BENCH"})
	require.NoError(b, err)

	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = envSrc.Load(ctx)
	}
}

func BenchmarkEnvSource_AutoConvert(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	values := []string{
		"true", "false", "42", "3.14", "30s", "item1,item2,item3", "plain string",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			_ = envSrc.autoConvertValue(value)
		}
	}
}

func BenchmarkEnvSource_KeyConversion(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	envKeys := []string{
		"TEST_SIMPLE", "TEST_NESTED_KEY", "TEST_DEEP_NESTED_KEY_VALUE",
		"TEST_WITH_NUMBERS_123", "TEST_VERY_LONG_KEY_NAME_WITH_MULTIPLE_PARTS",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, key := range envKeys {
			_ = envSrc.envKeyToConfigKey(key)
		}
	}
}