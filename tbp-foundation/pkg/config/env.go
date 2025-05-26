// File: env.go
// Title: Environment Variable Configuration for TBP
// Description: Provides environment variable-based configuration source with
//              type conversion, prefix filtering, nested key mapping, and
//              validation. Supports standard environment variable patterns
//              with automatic type detection and secure handling.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial environment variable configuration implementation

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
}

// EnvSourceOptions configures environment variable source creation
type EnvSourceOptions struct {
	Prefix        string            `json:"prefix"`         // Environment variable prefix (e.g., "TBP")
	Separator     string            `json:"separator"`      // Key separator (default: "_")
	KeyMapping    map[string]string `json:"key_mapping"`    // Custom key mappings
	TypeHints     map[string]string `json:"type_hints"`     // Type hints for conversion
	CaseSensitive bool              `json:"case_sensitive"` // Case-sensitive key matching
}

// NewEnvSource creates a new environment variable-based configuration source
func NewEnvSource(opts EnvSourceOptions) (*EnvSource, error) {
	if opts.Prefix == "" {
		opts.Prefix = "TBP"
	}

	if opts.Separator == "" {
		opts.Separator = "_"
	}

	// Ensure prefix ends with separator for consistent matching
	if !strings.HasSuffix(opts.Prefix, opts.Separator) {
		opts.Prefix += opts.Separator
	}

	es := &EnvSource{
		prefix:        opts.Prefix,
		separator:     opts.Separator,
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
	// Environment variables have high priority (override file config)
	return 100
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
	// but we can implement polling if needed
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
	// Remove prefix
	var key string
	if es.caseSensitive {
		key = strings.TrimPrefix(envKey, es.prefix)
	} else {
		upperEnvKey := strings.ToUpper(envKey)
		upperPrefix := strings.ToUpper(es.prefix)
		key = strings.TrimPrefix(upperEnvKey, upperPrefix)
	}

	// Check for custom key mapping
	if mappedKey, exists := es.keyMapping[envKey]; exists {
		return mappedKey
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
	case "string":
		return value, nil

	case "int", "integer":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to integer", value)
		}
		return intVal, nil

	case "int64":
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to int64", value)
		}
		return intVal, nil

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

	case "stringslice", "[]string":
		return es.parseStringSlice(value), nil

	case "intslice", "[]int":
		return es.parseIntSlice(value)

	default:
		return nil, core.Newf("unsupported type hint: %s", typeHint)
	}
}

// autoConvertValue automatically detects and converts the value type
func (es *EnvSource) autoConvertValue(value string) interface{} {
	// Try boolean first (common for environment variables)
	if boolVal, err := es.parseBool(value); err == nil {
		return boolVal
	}

	// Try integer
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		// Only return float if it has decimal places
		if strings.Contains(value, ".") {
			return floatVal
		}
	}

	// Try duration (if it looks like a duration)
	if strings.ContainsAny(value, "hms") || strings.HasSuffix(value, "us") || strings.HasSuffix(value, "ns") {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}

	// Check if it's a comma-separated list
	if strings.Contains(value, ",") {
		return es.parseStringSlice(value)
	}

	// Default to string
	return value
}

// parseBool parses a boolean value from various string representations
func (es *EnvSource) parseBool(value string) (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(value))
	switch lower {
	case "true", "yes", "1", "on", "enable", "enabled":
		return true, nil
	case "false", "no", "0", "off", "disable", "disabled", "":
		return false, nil
	default:
		return false, core.Newf("cannot convert '%s' to boolean", value)
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
	return result
}

// parseIntSlice parses a comma-separated string into a slice of integers
func (es *EnvSource) parseIntSlice(value string) ([]int, error) {
	if value == "" {
		return []int{}, nil
	}

	parts := strings.Split(value, ",")
	result := make([]int, len(parts))
	for i, part := range parts {
		trimmed := strings.TrimSpace(part)
		intVal, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, core.Wrapf(err, "failed to convert '%s' to integer in slice", trimmed)
		}
		result[i] = intVal
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

// GetPrefix returns the environment variable prefix
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
		if os.Getenv(envKey) == "" {
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