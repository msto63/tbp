// File: env_test.go
// Title: Tests for Environment Variable Configuration
// Description: Comprehensive test suite for environment variable configuration
//              source including type conversion, prefix filtering, key mapping,
//              validation, and edge cases. Tests performance characteristics
//              and concurrent access patterns with enhanced type support.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.1
// Created: 2025-05-26
// Modified: 2025-05-27
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation with comprehensive coverage
// - 2025-05-27 v0.1.1: Enhanced tests for expanded type conversions and new features

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
			Priority:      150,
		})
		require.NoError(t, err)
		assert.Equal(t, "MYAPP__", envSrc.prefix)
		assert.Equal(t, "__", envSrc.separator)
		assert.True(t, envSrc.caseSensitive)
		assert.Equal(t, 150, envSrc.Priority())
	})

	t.Run("auto-adds separator to prefix", func(t *testing.T) {
		envSrc, err := NewEnvSource(EnvSourceOptions{
			Prefix: "TEST", // No separator
		})
		require.NoError(t, err)
		assert.Equal(t, "TEST_", envSrc.prefix) // Should add separator
	})
}

func TestEnvSource_Load(t *testing.T) {
	// Setup test environment variables
	testEnvVars := map[string]string{
		"TEST_STRING_VALUE":    "hello_world",
		"TEST_INT_VALUE":       "42",
		"TEST_INT8_VALUE":      "127",
		"TEST_INT16_VALUE":     "32767",
		"TEST_INT32_VALUE":     "2147483647",
		"TEST_INT64_VALUE":     "9223372036854775807",
		"TEST_UINT_VALUE":      "4294967295",
		"TEST_UINT8_VALUE":     "255",
		"TEST_UINT16_VALUE":    "65535",
		"TEST_UINT32_VALUE":    "4294967295",
		"TEST_UINT64_VALUE":    "18446744073709551615",
		"TEST_FLOAT32_VALUE":   "3.14",
		"TEST_FLOAT64_VALUE":   "2.718281828",
		"TEST_BOOL_TRUE":       "true",
		"TEST_BOOL_FALSE":      "false",
		"TEST_BOOL_YES":        "yes",
		"TEST_BOOL_NO":         "no",
		"TEST_BOOL_1":          "1",
		"TEST_BOOL_0":          "0",
		"TEST_BOOL_ON":         "on",
		"TEST_BOOL_OFF":        "off",
		"TEST_BOOL_Y":          "y",
		"TEST_BOOL_N":          "n",
		"TEST_BOOL_T":          "t",
		"TEST_BOOL_F":          "f",
		"TEST_DURATION_VALUE":  "30s",
		"TEST_TIME_VALUE":      "2024-01-15T10:30:00Z",
		"TEST_SLICE_VALUE":     "item1,item2,item3",
		"TEST_INT_SLICE":       "1,2,3,4,5",
		"TEST_FLOAT_SLICE":     "1.1,2.2,3.3",
		"TEST_BOOL_SLICE":      "true,false,yes,no",
		"TEST_NESTED_KEY":      "nested_value",
		"OTHER_PREFIX_KEY":     "should_be_ignored",
		"TEST_EMPTY_VALUE":     "",
		"TEST_SPACED_SLICE":    "  item1  ,  item2  ,  item3  ",
		"TEST_MIXED_CASE":      "MixedCaseValue",
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

		// Test basic types
		assert.Equal(t, "hello_world", values["string.value"])
		assert.Equal(t, 42, values["int.value"])
		assert.Equal(t, int8(127), values["int8.value"])
		assert.Equal(t, int16(32767), values["int16.value"])
		assert.Equal(t, int32(2147483647), values["int32.value"])
		assert.Equal(t, int64(9223372036854775807), values["int64.value"])
		
		// Test unsigned types
		assert.Equal(t, uint(4294967295), values["uint.value"])
		assert.Equal(t, uint8(255), values["uint8.value"])
		assert.Equal(t, uint16(65535), values["uint16.value"])
		assert.Equal(t, uint32(4294967295), values["uint32.value"])
		assert.Equal(t, uint64(18446744073709551615), values["uint64.value"])
		
		// Test float types
		assert.Equal(t, float32(3.14), values["float32.value"])
		assert.Equal(t, 2.718281828, values["float64.value"])

		// Test boolean variations
		assert.Equal(t, true, values["bool.true"])
		assert.Equal(t, false, values["bool.false"])
		assert.Equal(t, true, values["bool.yes"])
		assert.Equal(t, false, values["bool.no"])
		assert.Equal(t, true, values["bool.1"])
		assert.Equal(t, false, values["bool.0"])
		assert.Equal(t, true, values["bool.on"])
		assert.Equal(t, false, values["bool.off"])
		assert.Equal(t, true, values["bool.y"])
		assert.Equal(t, false, values["bool.n"])
		assert.Equal(t, true, values["bool.t"])
		assert.Equal(t, false, values["bool.f"])

		// Test duration and time
		assert.Equal(t, 30*time.Second, values["duration.value"])
		timeVal, ok := values["time.value"].(time.Time)
		assert.True(t, ok)
		assert.Equal(t, "2024-01-15T10:30:00Z", timeVal.Format(time.RFC3339))

		// Test slices
		assert.Equal(t, []string{"item1", "item2", "item3"}, values["slice.value"])
		assert.Equal(t, []int{1, 2, 3, 4, 5}, values["int.slice"])
		assert.Equal(t, []float64{1.1, 2.2, 3.3}, values["float.slice"])
		assert.Equal(t, []bool{true, false, true, false}, values["bool.slice"])

		// Test other values
		assert.Equal(t, "nested_value", values["nested.key"])
		assert.Equal(t, "", values["empty.value"])
		assert.Equal(t, []string{"item1", "item2", "item3"}, values["spaced.slice"])

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
		assert.Equal(t, "MixedCaseValue", values["mixed.case"])
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
			{"y", true},
			{"n", false},
			{"t", true},
			{"f", false},
			{"", false},
		}

		for _, tc := range testCases {
			result, err := envSrc.parseBool(tc.input)
			assert.NoError(t, err, "Failed for input: %s", tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})

	t.Run("handles invalid boolean values", func(t *testing.T) {
		invalidValues := []string{"maybe", "sometimes", "invalid", "2", "-1", "abc"}
		
		for _, value := range invalidValues {
			_, err := envSrc.parseBool(value)
			assert.Error(t, err, "Should fail for input: %s", value)
			assert.Contains(t, err.Error(), "cannot convert")
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
			{"trailing,comma,", []string{"trailing", "comma"}}, // Empty strings filtered
			{",,middle,,", []string{"middle"}},
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
			{"1,,3", []int{1, 3}, false}, // Empty elements filtered
			{"-1,0,1", []int{-1, 0, 1}, false},
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

	t.Run("converts float slices", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []float64
			hasError bool
		}{
			{"", []float64{}, false},
			{"3.14", []float64{3.14}, false},
			{"1.1,2.2,3.3", []float64{1.1, 2.2, 3.3}, false},
			{" 1.5 , 2.5 , 3.5 ", []float64{1.5, 2.5, 3.5}, false},
			{"1.1,invalid,3.3", nil, true},
			{"1.1,,3.3", []float64{1.1, 3.3}, false}, // Empty elements filtered
			{"-1.5,0,1.5", []float64{-1.5, 0, 1.5}, false},
		}

		for _, tc := range testCases {
			result, err := envSrc.parseFloatSlice(tc.input)
			if tc.hasError {
				assert.Error(t, err, "Should fail for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Should succeed for input: %s", tc.input)
				assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
			}
		}
	})

	t.Run("converts bool slices", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected []bool
			hasError bool
		}{
			{"", []bool{}, false},
			{"true", []bool{true}, false},
			{"true,false,yes,no", []bool{true, false, true, false}, false},
			{" true , false ", []bool{true, false}, false},
			{"true,invalid,false", nil, true},
			{"true,,false", []bool{true, false}, false}, // Empty elements filtered
			{"1,0,on,off", []bool{true, false, true, false}, false},
		}

		for _, tc := range testCases {
			result, err := envSrc.parseBoolSlice(tc.input)
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
			{"123.0", 123.0}, // Should stay float due to decimal point
			{"30s", 30 * time.Second},
			{"item1,item2", []string{"item1", "item2"}},
			{"plain string", "plain string"},
			{"", ""},
			{"-42", -42},
			{"2147483648", int64(2147483648)}, // Too large for int32, becomes int64
		}

		for _, tc := range testCases {
			result := envSrc.autoConvertValue(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})
}

func TestEnvSource_AdvancedTypeConversion(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("converts by type hint", func(t *testing.T) {
		testCases := []struct {
			value    string
			typeHint string
			expected interface{}
			hasError bool
		}{
			// Integer types
			{"42", "int", 42, false},
			{"127", "int8", int8(127), false},
			{"32767", "int16", int16(32767), false},
			{"2147483647", "int32", int32(2147483647), false},
			{"9223372036854775807", "int64", int64(9223372036854775807), false},
			
			// Unsigned types
			{"42", "uint", uint(42), false},
			{"255", "uint8", uint8(255), false},
			{"65535", "uint16", uint16(65535), false},
			{"4294967295", "uint32", uint32(4294967295), false},
			{"18446744073709551615", "uint64", uint64(18446744073709551615), false},
			
			// Float types
			{"3.14", "float32", float32(3.14), false},
			{"2.718281828", "float64", 2.718281828, false},
			
			// Boolean
			{"true", "bool", true, false},
			{"yes", "boolean", true, false},
			
			// Duration
			{"30s", "duration", 30 * time.Second, false},
			
			// Time
			{"2024-01-15T10:30:00Z", "time", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), false},
			{"2024-01-15", "timestamp", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), false},
			
			// Slices
			{"a,b,c", "stringslice", []string{"a", "b", "c"}, false},
			{"1,2,3", "intslice", []int{1, 2, 3}, false},
			{"1.1,2.2", "floatslice", []float64{1.1, 2.2}, false},
			{"true,false", "boolslice", []bool{true, false}, false},
			
			// Alternative slice names
			{"a,b", "[]string", []string{"a", "b"}, false},
			{"1,2", "[]int", []int{1, 2}, false},
			{"true,false", "[]bool", []bool{true, false}, false},
			
			// String (explicit)
			{"test", "string", "test", false},
			{"test", "str", "test", false},
			
			// Errors
			{"invalid", "int", nil, true},
			{"3.14", "unsupported", nil, true},
			{"invalid-time", "time", nil, true},
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

	t.Run("lists supported types", func(t *testing.T) {
		supportedTypes := envSrc.GetSupportedTypes()
		
		expectedTypes := []string{
			"string", "str",
			"int", "integer", "int8", "int16", "int32", "int64",
			"uint", "unsigned", "uint8", "uint16", "uint32", "uint64",
			"float32", "float", "float64",
			"bool", "boolean",
			"duration",
			"time", "timestamp",
			"stringslice", "[]string", "strings",
			"intslice", "[]int", "integers",
			"floatslice", "[]float64", "floats",
			"boolslice", "[]bool", "booleans",
		}
		
		for _, expectedType := range expectedTypes {
			assert.Contains(t, supportedTypes, expectedType)
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

func TestEnvSource_EdgeCases(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("handles empty and whitespace values", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected interface{}
		}{
			{"", ""},
			{"   ", "   "},
			{"\t", "\t"},
			{"\n", "\n"},
		}

		for _, tc := range testCases {
			result := envSrc.autoConvertValue(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %q", tc.input)
		}
	})

	t.Run("handles single item slices", func(t *testing.T) {
		// Single item without comma should not become slice
		result := envSrc.autoConvertValue("single_item")
		assert.Equal(t, "single_item", result)

		// Single item with comma should become slice
		result = envSrc.autoConvertValue("single_item,")
		expected := []string{"single_item"}
		assert.Equal(t, expected, result)
	})

	t.Run("handles numeric edge cases", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected interface{}
		}{
			{"0", 0},
			{"-0", 0},
			{"00", 0},
			{"0.0", 0.0},
			{"-1", -1},
			{"1e10", 1e10},
			{"1.23e-4", 1.23e-4},
		}

		for _, tc := range testCases {
			result := envSrc.autoConvertValue(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})

	t.Run("handles duration edge cases", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected interface{}
		}{
			{"0s", 0 * time.Second},
			{"1ns", 1 * time.Nanosecond},
			{"1µs", 1 * time.Microsecond},
			{"1us", 1 * time.Microsecond}, // Alternative microsecond notation
			{"1ms", 1 * time.Millisecond},
			{"1h30m", 90 * time.Minute},
			{"24h", 24 * time.Hour},
		}

		for _, tc := range testCases {
			result := envSrc.autoConvertValue(tc.input)
			assert.Equal(t, tc.expected, result, "Failed for input: %s", tc.input)
		}
	})
}

func TestEnvSource_ErrorHandling(t *testing.T) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(t, err)

	t.Run("handles type conversion errors gracefully", func(t *testing.T) {
		// Setup environment variable with invalid value
		os.Setenv("TEST_INVALID_INT", "not_a_number")
		defer os.Unsetenv("TEST_INVALID_INT")

		// Add type hint that will cause conversion error
		envSrc.AddTypeHint("invalid.int", "int")

		ctx := context.Background()
		_, err := envSrc.Load(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to convert environment variable")
	})

	t.Run("provides helpful error messages", func(t *testing.T) {
		testCases := []struct {
			value    string
			typeHint string
			errorMsg string
		}{
			{"not_a_number", "int", "failed to convert 'not_a_number' to integer"},
			{"maybe", "bool", "cannot convert 'maybe' to boolean"},
			{"invalid_duration", "duration", "failed to convert 'invalid_duration' to duration"},
			{"bad_time", "time", "failed to parse 'bad_time' as time"},
		}

		for _, tc := range testCases {
			_, err := envSrc.convertByType(tc.value, tc.typeHint)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errorMsg)
		}
	})
}

func TestEnvSource_Performance(t *testing.T) {
	// Setup many environment variables
	for i := 0; i < 100; i++ {
		os.Setenv(fmt.Sprintf("PERF_TEST_VAR_%d", i), fmt.Sprintf("value_%d", i))
	}
	defer func() {
		for i := 0; i < 100; i++ {
			os.Unsetenv(fmt.Sprintf("PERF_TEST_VAR_%d", i))
		}
	}()

	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "PERF_TEST"})
	require.NoError(t, err)

	t.Run("loads many variables efficiently", func(t *testing.T) {
		ctx := context.Background()
		values, err := envSrc.Load(ctx)
		require.NoError(t, err)

		// Should have loaded all 100 variables
		assert.Len(t, values, 100)

		// Verify some values
		assert.Equal(t, "value_0", values["var.0"])
		assert.Equal(t, "value_99", values["var.99"])
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
		"2024-01-15T10:30:00Z", "yes", "no", "123.456", "1,2,3,4,5",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			_ = envSrc.autoConvertValue(value)
		}
	}
}

func BenchmarkEnvSource_TypeConversion(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	testCases := []struct {
		value string
		hint  string
	}{
		{"42", "int"},
		{"3.14", "float64"},
		{"true", "bool"},
		{"30s", "duration"},
		{"2024-01-15T10:30:00Z", "time"},
		{"a,b,c", "stringslice"},
		{"1,2,3", "intslice"},
		{"1.1,2.2,3.3", "floatslice"},
		{"true,false,true", "boolslice"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_, _ = envSrc.convertByType(tc.value, tc.hint)
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

func BenchmarkEnvSource_BoolParsing(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	boolValues := []string{
		"true", "false", "yes", "no", "1", "0", "on", "off",
		"enable", "disable", "enabled", "disabled", "y", "n", "t", "f",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, value := range boolValues {
			_, _ = envSrc.parseBool(value)
		}
	}
}

func BenchmarkEnvSource_SliceParsing(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	sliceValue := "item1,item2,item3,item4,item5,item6,item7,item8,item9,item10"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = envSrc.parseStringSlice(sliceValue)
	}
}

func BenchmarkEnvSource_IntSliceParsing(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	intSliceValue := "1,2,3,4,5,6,7,8,9,10"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = envSrc.parseIntSlice(intSliceValue)
	}
}

func BenchmarkEnvSource_FloatSliceParsing(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	floatSliceValue := "1.1,2.2,3.3,4.4,5.5,6.6,7.7,8.8,9.9,10.0"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = envSrc.parseFloatSlice(floatSliceValue)
	}
}

func BenchmarkEnvSource_BoolSliceParsing(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	boolSliceValue := "true,false,yes,no,1,0,on,off,enable,disable"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = envSrc.parseBoolSlice(boolSliceValue)
	}
}

func BenchmarkEnvSource_ListEnvironmentVariables(b *testing.B) {
	// Setup test environment variables
	for i := 0; i < 50; i++ {
		os.Setenv(fmt.Sprintf("BENCH_LIST_%d", i), fmt.Sprintf("value_%d", i))
	}
	defer func() {
		for i := 0; i < 50; i++ {
			os.Unsetenv(fmt.Sprintf("BENCH_LIST_%d", i))
		}
	}()

	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "BENCH_LIST"})
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = envSrc.ListEnvironmentVariables()
	}
}

func BenchmarkEnvSource_GetSupportedTypes(b *testing.B) {
	envSrc, err := NewEnvSource(EnvSourceOptions{Prefix: "TEST"})
	require.NoError(b, err)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = envSrc.GetSupportedTypes()
	}
}