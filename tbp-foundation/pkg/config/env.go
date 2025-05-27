// File: env.go
// Title: Environment Variable Configuration for TBP
// Description: Provides environment variable-based configuration source with
//              comprehensive type conversion, prefix filtering, nested key mapping,
//              and validation. Supports standard environment variable patterns
//              with automatic type detection and secure handling.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.1
// Created: 2025-05-26
// Modified: 2025-05-27
//
// Change History:
// - 2025-05-26 v0.1.0: Initial environment variable configuration implementation
// - 2025-05-27 v0.1.1: Enhanced type conversions, better error handling, expanded type support

package config

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/msto63/tbp/tbp-foundation/pkg/core"
)

// EnvSource implements the Source interface for environment variable configuration
type EnvSource struct {
	// mu protects concurrent access to environment source data
	mu sync.RWMutex

	// prefix is the environment variable prefix (e.g., "TBP_")
	prefix string

	// separator is used to separate nested keys (default: "_")
	separator string

	// values stores the loaded environment variable values
	values map[string]interface{}

	// keyMapping maps environment variable names to configuration keys
	keyMapping map[string]string

	// typeHints provides type hints for specific keys
	typeHints map[string]string

	// caseSensitive determines whether key matching is case-sensitive
	caseSensitive bool

	// priority sets the source priority for merging
	priority int
}

// EnvSourceOptions configures environment variable source creation
type EnvSourceOptions struct {
	Prefix        string            `json:"prefix"`         // Environment variable prefix (e.g., "TBP")
	Separator     string            `json:"separator"`      // Key separator (default: "_")
	KeyMapping    map[string]string `json:"key_mapping"`    // Custom key mappings
	TypeHints     map[string]string `json:"type_hints"`     // Type hints for conversion
	CaseSensitive bool              `json:"case_sensitive"` // Case-sensitive key matching
	Priority      int               `json:"priority"`       // Source priority (default: 100)
}

// NewEnvSource creates a new environment variable-based configuration source
func NewEnvSource(opts EnvSourceOptions) (*EnvSource, error) {
	if opts.Prefix == "" {
		opts.Prefix = "TBP"
	}

	if opts.Separator == "" {
		opts.Separator = "_"
	}

	// Set default priority if not specified
	if opts.Priority == 0 {
		opts.Priority = 100 // High priority by default
	}

	// Ensure prefix ends with separator for consistent matching
	if !strings.HasSuffix(opts.Prefix, opts.Separator) {
		opts.Prefix += opts.Separator
	}

	es := &EnvSource{
		prefix:        opts.Prefix,
		separator:     opts.Separator,
		priority:      opts.Priority,
		values:        make(map[string]interface{}),
		keyMapping:    opts.KeyMapping,
		typeHints:     opts.TypeHints,
		caseSensitive: opts.CaseSensitive,
	}

	if es.keyMapping == nil {
		es.keyMapping = make(map[string]string)
	}

	if es.typeHints == nil {
		es.typeHints = make(map[string]string)
	}

	return es, nil
}

// Name implements the Source interface
func (es *EnvSource) Name() string {
	return "env:" + strings.TrimSuffix(es.prefix, es.separator)
}

// Priority implements the Source interface
func (es *EnvSource) Priority() int {
	return es.priority
}

// Load implements the Source interface
func (es *EnvSource) Load(ctx context.Context) (map[string]interface{}, error) {
	es.mu.Lock()
	defer es.mu.Unlock()

	values := make(map[string]interface{})

	// Get all environment variables
	environ := os.Environ()

	for _, env := range environ {
		// Split into key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		envKey := parts[0]
		envValue := parts[1]

		// Check if environment variable matches our prefix
		if !es.matchesPrefix(envKey) {
			continue
		}

		// Convert environment variable name to configuration key
		configKey := es.envKeyToConfigKey(envKey)
		if configKey == "" {
			continue
		}

		// Convert value to appropriate type
		convertedValue, err := es.convertValue(configKey, envValue)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert environment variable %s", envKey)
		}

		values[configKey] = convertedValue
	}

	// Cache the values
	es.values = values

	return es.copyValues(), nil
}

// Watch implements the Source interface
func (es *EnvSource) Watch(ctx context.Context, callback func(map[string]interface{})) error {
	// Environment variables don't typically change during runtime,
	// so we don't implement active watching by default
	// This could be extended with polling if needed
	return nil
}

// matchesPrefix checks if an environment variable name matches our prefix
func (es *EnvSource) matchesPrefix(envKey string) bool {
	if es.caseSensitive {
		return strings.HasPrefix(envKey, es.prefix)
	}
	return strings.HasPrefix(strings.ToUpper(envKey), strings.ToUpper(es.prefix))
}

// envKeyToConfigKey converts an environment variable name to a configuration key
func (es *EnvSource) envKeyToConfigKey(envKey string) string {
	// Check for custom key mapping first
	if mappedKey, exists := es.keyMapping[envKey]; exists {
		return mappedKey
	}

	// Remove prefix
	var key string
	if es.caseSensitive {
		key = strings.TrimPrefix(envKey, es.prefix)
	} else {
		upperEnvKey := strings.ToUpper(envKey)
		upperPrefix := strings.ToUpper(es.prefix)
		key = strings.TrimPrefix(upperEnvKey, upperPrefix)
	}

	// Convert UPPER_CASE_WITH_UNDERSCORES to dot.separated.lowercase
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, es.separator, ".")

	return key
}

// convertValue converts a string environment variable value to the appropriate type
func (es *EnvSource) convertValue(key, value string) (interface{}, error) {
	// Check for explicit type hint
	if typeHint, exists := es.typeHints[key]; exists {
		return es.convertByType(value, typeHint)
	}

	// Auto-detect type based on value content
	return es.autoConvertValue(value), nil
}

// convertByType converts a value based on an explicit type hint
func (es *EnvSource) convertByType(value, typeHint string) (interface{}, error) {
	switch strings.ToLower(typeHint) {
	case "string", "str":
		return value, nil

	case "int", "integer":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to integer", value)
		}
		return intVal, nil

	case "int8":
		intVal, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to int8", value)
		}
		return int8(intVal), nil

	case "int16":
		intVal, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to int16", value)
		}
		return int16(intVal), nil

	case "int32":
		intVal, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to int32", value)
		}
		return int32(intVal), nil

	case "int64":
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to int64", value)
		}
		return intVal, nil

	case "uint", "unsigned":
		uintVal, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to uint", value)
		}
		return uint(uintVal), nil

	case "uint8":
		uintVal, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to uint8", value)
		}
		return uint8(uintVal), nil

	case "uint16":
		uintVal, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to uint16", value)
		}
		return uint16(uintVal), nil

	case "uint32":
		uintVal, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to uint32", value)
		}
		return uint32(uintVal), nil

	case "uint64":
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to uint64", value)
		}
		return uintVal, nil

	case "float32":
		floatVal, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to float32", value)
		}
		return float32(floatVal), nil

	case "float", "float64":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to float64", value)
		}
		return floatVal, nil

	case "bool", "boolean":
		return es.parseBool(value)

	case "duration":
		duration, err := time.ParseDuration(value)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to duration", value)
		}
		return duration, nil

	case "time", "timestamp":
		// Try multiple time formats
		timeFormats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		
		for _, format := range timeFormats {
			if t, err := time.Parse(format, value); err == nil {
				return t, nil
			}
		}
		return nil, core.Newf("failed to parse '%s' as time - supported formats: RFC3339, ISO date", value)

	case "stringslice", "[]string", "strings":
		return es.parseStringSlice(value), nil

	case "intslice", "[]int", "integers":
		return es.parseIntSlice(value)

	case "floatslice", "[]float64", "floats":
		return es.parseFloatSlice(value)

	case "boolslice", "[]bool", "booleans":
		return es.parseBoolSlice(value)

	default:
		return nil, core.Newf("unsupported type hint '%s' - supported types: string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, duration, time, stringslice, intslice, floatslice, boolslice", typeHint)
	}
}

// autoConvertValue automatically detects and converts the value type
func (es *EnvSource) autoConvertValue(value string) interface{} {
	// Empty string stays as string
	if value == "" {
		return value
	}

	// Try boolean first (common for environment variables)
	if boolVal, err := es.parseBool(value); err == nil {
		return boolVal
	}

	// Try integer (but avoid converting things like "123.0" to int)
	if !strings.Contains(value, ".") {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			// Return appropriate int type based on size
			if intVal >= -2147483648 && intVal <= 2147483647 {
				return int(intVal)
			}
			return intVal
		}
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Try duration (if it looks like a duration)
	if strings.ContainsAny(value, "hms") || 
	   strings.HasSuffix(value, "us") || 
	   strings.HasSuffix(value, "ns") || 
	   strings.HasSuffix(value, "µs") {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	// Check if it's a comma-separated list
	if strings.Contains(value, ",") && len(strings.Split(value, ",")) > 1 {
		return es.parseStringSlice(value)
	}

	// Default to string
	return value
}

// parseBool parses a boolean value from various string representations
func (es *EnvSource) parseBool(value string) (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "true", "yes", "1", "on", "enable", "enabled", "y", "t":
		return true, nil
	case "false", "no", "0", "off", "disable", "disabled", "n", "f", "":
		return false, nil
	default:
		return false, core.Newf("cannot convert '%s' to boolean - supported values: true/false, yes/no, 1/0, on/off, enable/disable", value)
	}
}

// parseStringSlice parses a comma-separated string into a slice of strings
func (es *EnvSource) parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}

	parts := strings.Split(value, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}
	
	// Remove empty strings
	filtered := make([]string, 0, len(result))
	for _, s := range result {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	
	return filtered
}

// parseIntSlice parses a comma-separated string into a slice of integers
func (es *EnvSource) parseIntSlice(value string) ([]int, error) {
	if value == "" {
		return []int{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, 0, len(parts))
	
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue // Skip empty parts
		}
		
		intVal, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' (element %d) to integer in slice", trimmed, i)
		}
		result = append(result, intVal)
	}
	
	return result, nil
}

// parseFloatSlice parses a comma-separated string into a slice of float64
func (es *EnvSource) parseFloatSlice(value string) ([]float64, error) {
	if value == "" {
		return []float64{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]float64, 0, len(parts))
	
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue // Skip empty parts
		}
		
		floatVal, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' (element %d) to float in slice", trimmed, i)
		}
		result = append(result, floatVal)
	}
	
	return result, nil
}

// parseBoolSlice parses a comma-separated string into a slice of booleans
func (es *EnvSource) parseBoolSlice(value string) ([]bool, error) {
	if value == "" {
		return []bool{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]bool, 0, len(parts))
	
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue // Skip empty parts
		}
		
		boolVal, err := es.parseBool(trimmed)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' (element %d) to boolean in slice", trimmed, i)
		}
		result = append(result, boolVal)
	}
	
	return result, nil
}

// copyValues returns a copy of the current values to prevent external modification
func (es *EnvSource) copyValues() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range es.values {
		result[key] = value
	}
	return result
}

// GetPrefix returns the environment variable prefix (without separator)
func (es *EnvSource) GetPrefix() string {
	return strings.TrimSuffix(es.prefix, es.separator)
}

// GetSeparator returns the key separator
func (es *EnvSource) GetSeparator() string {
	return es.separator
}

// AddKeyMapping adds a custom key mapping
func (es *EnvSource) AddKeyMapping(envKey, configKey string) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.keyMapping[envKey] = configKey
}

// AddTypeHint adds a type hint for a specific configuration key
func (es *EnvSource) AddTypeHint(configKey, typeHint string) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.typeHints[configKey] = typeHint
}

// GetKeyMappings returns all key mappings
func (es *EnvSource) GetKeyMappings() map[string]string {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range es.keyMapping {
		result[k] = v
	}
	return result
}

// GetTypeHints returns all type hints
func (es *EnvSource) GetTypeHints() map[string]string {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range es.typeHints {
		result[k] = v
	}
	return result
}

// ValidateEnvironment validates that required environment variables are set
func (es *EnvSource) ValidateEnvironment(requiredKeys []string) error {
	missing := make([]string, 0)

	for _, key := range requiredKeys {
		// Convert config key back to environment variable name
		envKey := es.configKeyToEnvKey(key)
		if _, exists := os.LookupEnv(envKey); !exists {
			missing = append(missing, envKey)
		}
	}

	if len(missing) > 0 {
		return core.Newf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// configKeyToEnvKey converts a configuration key back to environment variable name
func (es *EnvSource) configKeyToEnvKey(configKey string) string {
	// Check reverse key mapping first
	for envKey, mappedConfigKey := range es.keyMapping {
		if mappedConfigKey == configKey {
			return envKey
		}
	}

	// Convert dot.separated.key to ENV_VAR_FORMAT
	envKey := strings.ToUpper(configKey)
	envKey = strings.ReplaceAll(envKey, ".", es.separator)
	return es.prefix + envKey
}

// ListEnvironmentVariables returns all environment variables that match the prefix
func (es *EnvSource) ListEnvironmentVariables() map[string]string {
	result := make(map[string]string)
	environ := os.Environ()

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		envKey := parts[0]
		envValue := parts[1]

		if es.matchesPrefix(envKey) {
			result[envKey] = envValue
		}
	}

	return result
}

// GetEnvironmentVariableName returns the environment variable name for a config key
func (es *EnvSource) GetEnvironmentVariableName(configKey string) string {
	return es.configKeyToEnvKey(configKey)
}

// IsSet checks if a configuration key has a corresponding environment variable set
func (es *EnvSource) IsSet(configKey string) bool {
	envKey := es.configKeyToEnvKey(configKey)
	_, exists := os.LookupEnv(envKey)
	return exists
}

// GetRaw returns the raw string value of an environment variable
func (es *EnvSource) GetRaw(configKey string) (string, bool) {
	envKey := es.configKeyToEnvKey(configKey)
	return os.LookupEnv(envKey)
}

// SetEnvironmentVariable sets an environment variable (useful for testing)
func (es *EnvSource) SetEnvironmentVariable(configKey string, value string) error {
	envKey := es.configKeyToEnvKey(configKey)
	return os.Setenv(envKey, value)
}

// UnsetEnvironmentVariable unsets an environment variable (useful for testing)
func (es *EnvSource) UnsetEnvironmentVariable(configKey string) error {
	envKey := es.configKeyToEnvKey(configKey)
	return os.Unsetenv(envKey)
}

// GetSupportedTypes returns a list of all supported type hints
func (es *EnvSource) GetSupportedTypes() []string {
	return []string{
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
}