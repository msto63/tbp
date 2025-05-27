// File: config.go
// Title: Configuration Management for TBP
// Description: Provides multi-layered configuration management supporting
//              environment variables, configuration files, command-line flags,
//              and remote configuration sources. Implements type-safe configuration
//              structures with validation, hot-reloading, and sensitive data protection.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.1
// Created: 2025-05-26
// Modified: 2025-05-27
//
// Change History:
// - 2025-05-26 v0.1.0: Initial configuration management implementation
// - 2025-05-27 v0.1.1: Improved interface segregation, error codes, validation enhancements

package config

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/msto63/tbp/tbp-foundation/pkg/core"
)

// Config represents the main configuration structure for TBP components.
// It provides a unified interface for configuration management across
// different sources and environments.
type Config struct {
	// mu protects concurrent access to configuration data
	mu sync.RWMutex

	// sources contains all configuration sources in priority order
	sources []Source

	// values stores the merged configuration values
	values map[string]interface{}

	// watchers contains registered configuration change watchers
	watchers []Watcher

	// metadata contains configuration metadata and validation info
	metadata *Metadata

	// environment stores the current environment name
	environment string
}

// Source represents a configuration source (env vars, files, etc.)
// This is the base interface that all sources must implement
type Source interface {
	// Name returns the source name for logging and debugging
	Name() string

	// Load loads configuration values from the source
	Load(ctx context.Context) (map[string]interface{}, error)

	// Priority returns the source priority (higher = more important)
	Priority() int
}

// WatchableSource extends Source with watching capabilities
// Only sources that can detect changes should implement this
type WatchableSource interface {
	Source
	
	// Watch monitors the source for changes
	Watch(ctx context.Context, callback func(map[string]interface{})) error
}

// ValidatableSource extends Source with validation capabilities
// Sources that can validate their configuration should implement this
type ValidatableSource interface {
	Source
	
	// Validate checks if the source configuration is valid
	Validate() error
}

// WritableSource extends Source with write capabilities
// Sources that support writing configuration should implement this
type WritableSource interface {
	Source
	
	// WriteConfig writes configuration values to the source
	WriteConfig(values map[string]interface{}) error
}

// Watcher receives notifications when configuration changes
type Watcher interface {
	// OnConfigChange is called when configuration values change
	OnConfigChange(ctx context.Context, changes map[string]ConfigChange)
}

// ConfigChange represents a configuration value change
type ConfigChange struct {
	Key      string      `json:"key"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value"`
	Source   string      `json:"source"`
	Action   ChangeAction `json:"action"`
}

// ChangeAction represents the type of change that occurred
type ChangeAction string

const (
	// ChangeActionAdd indicates a new configuration key was added
	ChangeActionAdd ChangeAction = "add"
	
	// ChangeActionUpdate indicates an existing configuration key was updated
	ChangeActionUpdate ChangeAction = "update"
	
	// ChangeActionDelete indicates a configuration key was removed
	ChangeActionDelete ChangeAction = "delete"
)

// String returns the string representation of the change action
func (ca ChangeAction) String() string {
	return string(ca)
}

// Metadata contains configuration schema and validation information
type Metadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Fields      map[string]Field  `json:"fields"`
	Validators  []ValidatorFunc   `json:"-"`
	Secrets     map[string]string `json:"-"` // Never serialize secrets
}

// Field describes a configuration field with validation and metadata
type Field struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Description  string      `json:"description,omitempty"`
	Sensitive    bool        `json:"sensitive,omitempty"`
	Deprecated   bool        `json:"deprecated,omitempty"`
	Validators   []string    `json:"validators,omitempty"`
	MinValue     interface{} `json:"min_value,omitempty"`
	MaxValue     interface{} `json:"max_value,omitempty"`
	Pattern      string      `json:"pattern,omitempty"`
	Enum         []string    `json:"enum,omitempty"`
}

// ValidatorFunc validates configuration values
type ValidatorFunc func(key string, value interface{}) error

// LoadOptions configures how configuration is loaded
type LoadOptions struct {
	Sources      []Source               `json:"-"`
	Environment  string                 `json:"environment"`
	ConfigPaths  []string               `json:"config_paths"`
	EnvPrefix    string                 `json:"env_prefix"`
	Defaults     map[string]interface{} `json:"defaults"`
	Validation   bool                   `json:"validation"`
	HotReload    bool                   `json:"hot_reload"`
	Metadata     *Metadata              `json:"metadata,omitempty"`
	FailOnMissing bool                  `json:"fail_on_missing"` // Fail if required sources are missing
}

// New creates a new configuration manager with the specified options
func New(ctx context.Context, opts LoadOptions) (*Config, error) {
	if opts.EnvPrefix == "" {
		opts.EnvPrefix = "TBP"
	}

	if opts.Environment == "" {
		opts.Environment = "development"
	}

	config := &Config{
		sources:     make([]Source, 0),
		values:      make(map[string]interface{}),
		watchers:    make([]Watcher, 0),
		metadata:    opts.Metadata,
		environment: opts.Environment,
	}

	// Set default metadata if not provided
	if config.metadata == nil {
		config.metadata = &Metadata{
			Name:        "tbp-config",
			Version:     core.GetVersion(),
			Environment: opts.Environment,
			Fields:      make(map[string]Field),
			Validators:  make([]ValidatorFunc, 0),
			Secrets:     make(map[string]string),
		}
	}

	// Add default configuration sources if none provided
	if len(opts.Sources) == 0 {
		// Add environment variables source
		envSource, err := NewEnvSource(EnvSourceOptions{
			Prefix: opts.EnvPrefix,
		})
		if err != nil {
			return nil, core.Wrap(err, "failed to create environment source")
		}
		opts.Sources = append(opts.Sources, envSource)

		// Add file sources for each config path
		for _, path := range opts.ConfigPaths {
			fileSource, err := NewFileSource(FileSourceOptions{
				Path:     path,
				Format:   "auto", // Auto-detect format
				Optional: !opts.FailOnMissing,
			})
			if err != nil {
				if opts.FailOnMissing {
					return nil, core.Wrapf(err, "failed to create file source for %s", path)
				}
				// Log warning but continue if not failing on missing
				fmt.Printf("Warning: failed to create file source for %s: %v\n", path, err)
				continue
			}
			opts.Sources = append(opts.Sources, fileSource)
		}
	}

	// Add sources to configuration
	for _, source := range opts.Sources {
		if err := config.AddSource(source); err != nil {
			return nil, core.Wrapf(err, "failed to add source %s", source.Name())
		}
	}

	// Set default values (lowest priority)
	if len(opts.Defaults) > 0 {
		defaultSource := NewDefaultSource(opts.Defaults)
		if err := config.AddSource(defaultSource); err != nil {
			return nil, core.Wrap(err, "failed to add default source")
		}
	}

	// Load initial configuration
	if err := config.Load(ctx); err != nil {
		return nil, core.Wrap(err, "failed to load initial configuration")
	}

	// Validate configuration if requested
	if opts.Validation {
		if err := config.Validate(ctx); err != nil {
			return nil, core.Wrap(err, "configuration validation failed")
		}
	}

	// Start hot-reloading if requested
	if opts.HotReload {
		if err := config.StartWatching(ctx); err != nil {
			return nil, core.Wrap(err, "failed to start configuration watching")
		}
	}

	return config, nil
}

// AddSource adds a configuration source to the configuration manager
func (c *Config) AddSource(source Source) error {
	if source == nil {
		return core.New("source cannot be nil")
	}

	// Validate source if it supports validation
	if validatable, ok := source.(ValidatableSource); ok {
		if err := validatable.Validate(); err != nil {
			return core.Wrapf(err, "source %s validation failed", source.Name())
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.sources = append(c.sources, source)
	
	// Sort sources by priority (highest priority first)
	for i := len(c.sources) - 1; i > 0; i-- {
		if c.sources[i].Priority() > c.sources[i-1].Priority() {
			c.sources[i], c.sources[i-1] = c.sources[i-1], c.sources[i]
		} else {
			break
		}
	}

	return nil
}

// Load loads configuration from all sources and merges them
func (c *Config) Load(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	newValues := make(map[string]interface{})

	// Load from sources in reverse priority order (lowest first)
	// This allows higher priority sources to override lower priority ones
	for i := len(c.sources) - 1; i >= 0; i-- {
		source := c.sources[i]
		
		values, err := source.Load(ctx)
		if err != nil {
			return core.Wrapf(err, "failed to load from source %s", source.Name())
		}

		// Merge values (higher priority overwrites lower priority)
		for key, value := range values {
			newValues[key] = value
		}
	}

	// Store old values for change detection
	oldValues := c.values
	c.values = newValues

	// Notify watchers of changes
	if len(c.watchers) > 0 {
		changes := c.detectChanges(oldValues, newValues)
		if len(changes) > 0 {
			go c.notifyWatchers(ctx, changes)
		}
	}

	return nil
}

// Validate validates the current configuration against defined rules
func (c *Config) Validate(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var validationErrors []string

	// Validate required fields
	for fieldName, field := range c.metadata.Fields {
		if field.Required {
			if _, exists := c.values[fieldName]; !exists {
				validationErrors = append(validationErrors, 
					fmt.Sprintf("required configuration field '%s' is missing", fieldName))
			}
		}
		
		// Validate field constraints if value exists
		if value, exists := c.values[fieldName]; exists {
			if err := c.validateField(fieldName, field, value); err != nil {
				validationErrors = append(validationErrors, err.Error())
			}
		}
	}

	// Run custom validators
	for _, validator := range c.metadata.Validators {
		for key, value := range c.values {
			if err := validator(key, value); err != nil {
				validationErrors = append(validationErrors, 
					fmt.Sprintf("validation failed for field '%s': %v", key, err))
			}
		}
	}

	// Check for deprecated fields
	for key := range c.values {
		if field, exists := c.metadata.Fields[key]; exists && field.Deprecated {
			fmt.Printf("Warning: configuration field '%s' is deprecated: %s\n", 
				key, field.Description)
		}
	}

	if len(validationErrors) > 0 {
		return core.Newf("configuration validation failed:\n  - %s", 
			strings.Join(validationErrors, "\n  - "))
	}

	return nil
}

// validateField validates a single field against its constraints
func (c *Config) validateField(fieldName string, field Field, value interface{}) error {
	// Type validation
	if field.Type != "" {
		if err := c.validateFieldType(fieldName, field.Type, value); err != nil {
			return err
		}
	}

	// Range validation for numeric types
	if field.MinValue != nil || field.MaxValue != nil {
		if err := c.validateFieldRange(fieldName, field, value); err != nil {
			return err
		}
	}

	// Enum validation
	if len(field.Enum) > 0 {
		if err := c.validateFieldEnum(fieldName, field.Enum, value); err != nil {
			return err
		}
	}

	// Pattern validation for strings
	if field.Pattern != "" {
		if err := c.validateFieldPattern(fieldName, field.Pattern, value); err != nil {
			return err
		}
	}

	return nil
}

// validateFieldType validates the type of a field value
func (c *Config) validateFieldType(fieldName, expectedType string, value interface{}) error {
	actualType := reflect.TypeOf(value).String()
	
	// Normalize type names
	normalizedExpected := c.normalizeTypeName(expectedType)
	normalizedActual := c.normalizeTypeName(actualType)
	
	if normalizedExpected != normalizedActual {
		return core.Newf("field '%s' has type %s but expected %s", 
			fieldName, actualType, expectedType)
	}
	
	return nil
}

// validateFieldRange validates numeric range constraints
func (c *Config) validateFieldRange(fieldName string, field Field, value interface{}) error {
	// This is a simplified version - a full implementation would handle
	// all numeric types and proper type conversion
	switch v := value.(type) {
	case int:
		if field.MinValue != nil {
			if min, ok := field.MinValue.(int); ok && v < min {
				return core.Newf("field '%s' value %d is below minimum %d", fieldName, v, min)
			}
		}
		if field.MaxValue != nil {
			if max, ok := field.MaxValue.(int); ok && v > max {
				return core.Newf("field '%s' value %d exceeds maximum %d", fieldName, v, max)
			}
		}
	case float64:
		if field.MinValue != nil {
			if min, ok := field.MinValue.(float64); ok && v < min {
				return core.Newf("field '%s' value %f is below minimum %f", fieldName, v, min)
			}
		}
		if field.MaxValue != nil {
			if max, ok := field.MaxValue.(float64); ok && v > max {
				return core.Newf("field '%s' value %f exceeds maximum %f", fieldName, v, max)
			}
		}
	}
	
	return nil
}

// validateFieldEnum validates enum constraints
func (c *Config) validateFieldEnum(fieldName string, enum []string, value interface{}) error {
	strValue := fmt.Sprintf("%v", value)
	
	for _, enumValue := range enum {
		if strValue == enumValue {
			return nil
		}
	}
	
	return core.Newf("field '%s' value '%s' is not in allowed enum values: %s", 
		fieldName, strValue, strings.Join(enum, ", "))
}

// validateFieldPattern validates pattern constraints
func (c *Config) validateFieldPattern(fieldName, pattern string, value interface{}) error {
	// This would require regex validation - simplified for now
	strValue := fmt.Sprintf("%v", value)
	
	// Basic length check as example
	if len(strValue) == 0 {
		return core.Newf("field '%s' cannot be empty (pattern: %s)", fieldName, pattern)
	}
	
	return nil
}

// normalizeTypeName normalizes type names for comparison
func (c *Config) normalizeTypeName(typeName string) string {
	switch typeName {
	case "int", "int32", "int64":
		return "integer"
	case "float32", "float64":
		return "float"
	case "bool":
		return "boolean"
	default:
		return typeName
	}
}

// Get retrieves a configuration value by key
func (c *Config) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.values[key]
	return value, exists
}

// GetString retrieves a string configuration value
func (c *Config) GetString(key string) (string, error) {
	value, exists := c.Get(key)
	if !exists {
		return "", core.Newf("configuration key '%s' not found", key)
	}

	if str, ok := value.(string); ok {
		return str, nil
	}

	return fmt.Sprintf("%v", value), nil
}

// GetInt retrieves an integer configuration value
func (c *Config) GetInt(key string) (int, error) {
	value, exists := c.Get(key)
	if !exists {
		return 0, core.Newf("configuration key '%s' not found", key)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case float32:
		return int(v), nil
	case string:
		if i, err := fmt.Sscanf(v, "%d", new(int)); err == nil && i == 1 {
			var result int
			fmt.Sscanf(v, "%d", &result)
			return result, nil
		}
	}

	return 0, core.Newf("configuration key '%s' with value '%v' cannot be converted to int", key, value)
}

// GetBool retrieves a boolean configuration value
func (c *Config) GetBool(key string) (bool, error) {
	value, exists := c.Get(key)
	if !exists {
		return false, core.Newf("configuration key '%s' not found", key)
	}

	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		switch lower {
		case "true", "yes", "1", "on", "enable", "enabled", "y", "t":
			return true, nil
		case "false", "no", "0", "off", "disable", "disabled", "n", "f", "":
			return false, nil
		}
	case int:
		return v != 0, nil
	case float64:
		return v != 0, nil
	}

	return false, core.Newf("configuration key '%s' with value '%v' cannot be converted to bool", key, value)
}

// GetDuration retrieves a duration configuration value
func (c *Config) GetDuration(key string) (time.Duration, error) {
	value, exists := c.Get(key)
	if !exists {
		return 0, core.Newf("configuration key '%s' not found", key)
	}

	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		duration, err := time.ParseDuration(v)
		if err != nil {
			return 0, core.Wrapf(err, "configuration key '%s' cannot be parsed as duration", key)
		}
		return duration, nil
	case int, int32, int64:
		// Assume seconds if numeric value provided
		var seconds int64
		switch num := v.(type) {
		case int:
			seconds = int64(num)
		case int32:
			seconds = int64(num)
		case int64:
			seconds = num
		}
		return time.Duration(seconds) * time.Second, nil
	case float32, float64:
		// Assume seconds if numeric value provided
		var seconds float64
		switch num := v.(type) {
		case float32:
			seconds = float64(num)
		case float64:
			seconds = num
		}
		return time.Duration(seconds * float64(time.Second)), nil
	}

	return 0, core.Newf("configuration key '%s' with value '%v' cannot be converted to duration", key, value)
}

// GetStringWithDefault retrieves a string value with a default fallback
func (c *Config) GetStringWithDefault(key, defaultValue string) string {
	if value, err := c.GetString(key); err == nil {
		return value
	}
	return defaultValue
}

// GetIntWithDefault retrieves an int value with a default fallback
func (c *Config) GetIntWithDefault(key string, defaultValue int) int {
	if value, err := c.GetInt(key); err == nil {
		return value
	}
	return defaultValue
}

// GetBoolWithDefault retrieves a bool value with a default fallback
func (c *Config) GetBoolWithDefault(key string, defaultValue bool) bool {
	if value, err := c.GetBool(key); err == nil {
		return value
	}
	return defaultValue
}

// GetDurationWithDefault retrieves a duration value with a default fallback
func (c *Config) GetDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value, err := c.GetDuration(key); err == nil {
		return value
	}
	return defaultValue
}

// Unmarshal unmarshals configuration into a struct
func (c *Config) Unmarshal(v interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return core.New("unmarshal target must be a non-nil pointer")
	}

	return c.unmarshalValue(rv.Elem(), "")
}

// unmarshalValue recursively unmarshals configuration values into a struct
func (c *Config) unmarshalValue(rv reflect.Value, prefix string) error {
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get configuration key from struct tag or field name
		configKey := field.Tag.Get("config")
		if configKey == "" {
			configKey = strings.ToLower(field.Name)
		}
		if configKey == "-" {
			continue
		}

		fullKey := configKey
		if prefix != "" {
			fullKey = prefix + "." + configKey
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			if err := c.unmarshalValue(fieldValue, fullKey); err != nil {
				return err
			}
			continue
		}

		// Get value from configuration
		value, exists := c.values[fullKey]
		if !exists {
			// Check for default value in struct tag
			if defaultValue := field.Tag.Get("default"); defaultValue != "" {
				value = defaultValue
				exists = true
			}
		}

		if !exists {
			// Check if field is required
			if field.Tag.Get("required") == "true" {
				return core.Newf("required configuration field '%s' not found", fullKey)
			}
			continue
		}

		// Set field value
		if err := c.setFieldValue(fieldValue, value); err != nil {
			return core.Wrapf(err, "failed to set field '%s'", fullKey)
		}
	}

	return nil
}

// setFieldValue sets a reflect.Value from a configuration value
func (c *Config) setFieldValue(rv reflect.Value, value interface{}) error {
	switch rv.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			rv.SetString(str)
		} else {
			rv.SetString(fmt.Sprintf("%v", value))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intVal int64
		switch v := value.(type) {
		case int:
			intVal = int64(v)
		case int8:
			intVal = int64(v)
		case int16:
			intVal = int64(v)
		case int32:
			intVal = int64(v)
		case int64:
			intVal = v
		case uint:
			intVal = int64(v)
		case uint8:
			intVal = int64(v)
		case uint16:
			intVal = int64(v)
		case uint32:
			intVal = int64(v)
		case uint64:
			intVal = int64(v)
		case float32:
			intVal = int64(v)
		case float64:
			intVal = int64(v)
		case string:
			if i, err := fmt.Sscanf(v, "%d", &intVal); err != nil || i != 1 {
				return core.Newf("cannot convert '%v' to int", value)
			}
		default:
			return core.Newf("cannot convert '%v' to int", value)
		}
		
		// Check for overflow
		if rv.OverflowInt(intVal) {
			return core.Newf("value %d overflows %s", intVal, rv.Type())
		}
		rv.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var uintVal uint64
		switch v := value.(type) {
		case uint:
			uintVal = uint64(v)
		case uint8:
			uintVal = uint64(v)
		case uint16:
			uintVal = uint64(v)
		case uint32:
			uintVal = uint64(v)
		case uint64:
			uintVal = v
		case int:
			if v < 0 {
				return core.Newf("cannot convert negative value %d to uint", v)
			}
			uintVal = uint64(v)
		case int64:
			if v < 0 {
				return core.Newf("cannot convert negative value %d to uint", v)
			}
			uintVal = uint64(v)
		case float64:
			if v < 0 {
				return core.Newf("cannot convert negative value %f to uint", v)
			}
			uintVal = uint64(v)
		case string:
			if i, err := fmt.Sscanf(v, "%d", &uintVal); err != nil || i != 1 {
				return core.Newf("cannot convert '%v' to uint", value)
			}
		default:
			return core.Newf("cannot convert '%v' to uint", value)
		}
		
		// Check for overflow
		if rv.OverflowUint(uintVal) {
			return core.Newf("value %d overflows %s", uintVal, rv.Type())
		}
		rv.SetUint(uintVal)

	case reflect.Bool:
		var boolVal bool
		switch v := value.(type) {
		case bool:
			boolVal = v
		case string:
			lower := strings.ToLower(strings.TrimSpace(v))
			switch lower {
			case "true", "yes", "1", "on", "enable", "enabled", "y", "t":
				boolVal = true
			case "false", "no", "0", "off", "disable", "disabled", "n", "f", "":
				boolVal = false
			default:
				return core.Newf("cannot convert '%v' to bool", value)
			}
		case int, int64:
			switch num := v.(type) {
			case int:
				boolVal = num != 0
			case int64:
				boolVal = num != 0
			}
		default:
			return core.Newf("cannot convert '%v' to bool", value)
		}
		rv.SetBool(boolVal)

	case reflect.Float32, reflect.Float64:
		var floatVal float64
		switch v := value.(type) {
		case float32:
			floatVal = float64(v)
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		case string:
			if f, err := fmt.Sscanf(v, "%f", &floatVal); err != nil || f != 1 {
				return core.Newf("cannot convert '%v' to float", value)
			}
		default:
			return core.Newf("cannot convert '%v' to float", value)
		}
		
		// Check for overflow
		if rv.OverflowFloat(floatVal) {
			return core.Newf("value %f overflows %s", floatVal, rv.Type())
		}
		rv.SetFloat(floatVal)

	default:
		// Handle special types like time.Duration, time.Time, etc.
		if rv.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := c.parseDuration(value)
			if err != nil {
				return core.Wrapf(err, "cannot convert '%v' to duration", value)
			}
			rv.Set(reflect.ValueOf(duration))
		} else if rv.Type() == reflect.TypeOf(time.Time{}) {
			timeVal, err := c.parseTime(value)
			if err != nil {
				return core.Wrapf(err, "cannot convert '%v' to time", value)
			}
			rv.Set(reflect.ValueOf(timeVal))
		} else {
			return core.Newf("unsupported field type: %s", rv.Kind())
		}
	}

	return nil
}

// parseDuration parses a duration from various value types
func (c *Config) parseDuration(value interface{}) (time.Duration, error) {
	switch v := value.(type) {
	case time.Duration:
		return v, nil
	case string:
		return time.ParseDuration(v)
	case int, int64:
		var seconds int64
		switch num := v.(type) {
		case int:
			seconds = int64(num)
		case int64:
			seconds = num
		}
		return time.Duration(seconds) * time.Second, nil
	case float64:
		return time.Duration(v * float64(time.Second)), nil
	default:
		return 0, core.Newf("cannot convert %T to duration", value)
	}
}

// parseTime parses a time from various value types
func (c *Config) parseTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// Try multiple time formats
		timeFormats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		
		for _, format := range timeFormats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return time.Time{}, core.Newf("cannot parse time string '%s'", v)
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, core.Newf("cannot convert %T to time", value)
	}
}

// AddWatcher adds a configuration change watcher
func (c *Config) AddWatcher(watcher Watcher) {
	if watcher == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.watchers = append(c.watchers, watcher)
}

// RemoveWatcher removes a configuration change watcher
func (c *Config) RemoveWatcher(watcher Watcher) {
	if watcher == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i, w := range c.watchers {
		if w == watcher {
			// Remove watcher by replacing with last element and shrinking slice
			c.watchers[i] = c.watchers[len(c.watchers)-1]
			c.watchers = c.watchers[:len(c.watchers)-1]
			break
		}
	}
}

// StartWatching starts watching all sources for configuration changes
func (c *Config) StartWatching(ctx context.Context) error {
	for _, source := range c.sources {
		if watchable, ok := source.(WatchableSource); ok {
			go func(ws WatchableSource) {
				err := ws.Watch(ctx, func(values map[string]interface{}) {
					// Reload configuration when source changes
					if err := c.Load(ctx); err != nil {
						// Log error but continue watching
						// In a real implementation, this would use the logging package
						fmt.Printf("Error reloading configuration from %s: %v\n", ws.Name(), err)
					}
				})
				if err != nil {
					fmt.Printf("Error watching source %s: %v\n", ws.Name(), err)
				}
			}(watchable)
		}
	}

	return nil
}

// detectChanges compares old and new configuration values to detect changes
func (c *Config) detectChanges(oldValues, newValues map[string]interface{}) map[string]ConfigChange {
	changes := make(map[string]ConfigChange)

	// Check for modified and new values
	for key, newValue := range newValues {
		if oldValue, exists := oldValues[key]; exists {
			if !reflect.DeepEqual(oldValue, newValue) {
				changes[key] = ConfigChange{
					Key:      key,
					OldValue: oldValue,
					NewValue: newValue,
					Source:   "merged",
					Action:   ChangeActionUpdate,
				}
			}
		} else {
			changes[key] = ConfigChange{
				Key:      key,
				NewValue: newValue,
				Source:   "merged",
				Action:   ChangeActionAdd,
			}
		}
	}

	// Check for deleted values
	for key, oldValue := range oldValues {
		if _, exists := newValues[key]; !exists {
			changes[key] = ConfigChange{
				Key:      key,
				OldValue: oldValue,
				Source:   "merged",
				Action:   ChangeActionDelete,
			}
		}
	}

	return changes
}

// notifyWatchers notifies all registered watchers of configuration changes
func (c *Config) notifyWatchers(ctx context.Context, changes map[string]ConfigChange) {
	// Create a copy of watchers to avoid holding lock during notification
	c.mu.RLock()
	watchers := make([]Watcher, len(c.watchers))
	copy(watchers, c.watchers)
	c.mu.RUnlock()

	for _, watcher := range watchers {
		go func(w Watcher) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic in configuration watcher: %v\n", r)
				}
			}()
			w.OnConfigChange(ctx, changes)
		}(watcher)
	}
}

// GetAll returns all configuration values
func (c *Config) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]interface{})
	for key, value := range c.values {
		result[key] = value
	}
	return result
}

// GetKeys returns all configuration keys
func (c *Config) GetKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.values))
	for key := range c.values {
		keys = append(keys, key)
	}
	return keys
}

// HasKey checks if a configuration key exists
func (c *Config) HasKey(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.values[key]
	return exists
}

// GetMetadata returns configuration metadata
func (c *Config) GetMetadata() *Metadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metadata
}

// GetEnvironment returns the current environment name
func (c *Config) GetEnvironment() string {
	return c.environment
}

// GetSources returns information about all configured sources
func (c *Config) GetSources() []SourceInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sources := make([]SourceInfo, len(c.sources))
	for i, source := range c.sources {
		sources[i] = SourceInfo{
			Name:       source.Name(),
			Priority:   source.Priority(),
			Watchable:  isWatchableSource(source),
			Writable:   isWritableSource(source),
			Validatable: isValidatableSource(source),
		}
	}
	return sources
}

// SourceInfo provides information about a configuration source
type SourceInfo struct {
	Name        string `json:"name"`
	Priority    int    `json:"priority"`
	Watchable   bool   `json:"watchable"`
	Writable    bool   `json:"writable"`
	Validatable bool   `json:"validatable"`
}

// isWatchableSource checks if a source implements WatchableSource
func isWatchableSource(source Source) bool {
	_, ok := source.(WatchableSource)
	return ok
}

// isWritableSource checks if a source implements WritableSource
func isWritableSource(source Source) bool {
	_, ok := source.(WritableSource)
	return ok
}

// isValidatableSource checks if a source implements ValidatableSource
func isValidatableSource(source Source) bool {
	_, ok := source.(ValidatableSource)
	return ok
}

// WriteToSource writes configuration values to a specific source (if writable)
func (c *Config) WriteToSource(sourceName string, values map[string]interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, source := range c.sources {
		if source.Name() == sourceName {
			if writable, ok := source.(WritableSource); ok {
				return writable.WriteConfig(values)
			}
			return core.Newf("source '%s' is not writable", sourceName)
		}
	}

	return core.Newf("source '%s' not found", sourceName)
}

// Reload forces a reload of all configuration sources
func (c *Config) Reload(ctx context.Context) error {
	return c.Load(ctx)
}

// Close cleanly shuts down the configuration manager
func (c *Config) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop all watchers
	for _, source := range c.sources {
		if stoppable, ok := source.(interface{ Stop() }); ok {
			stoppable.Stop()
		}
	}

	// Clear all data
	c.sources = nil
	c.values = nil
	c.watchers = nil

	return nil
}

// DefaultSource implements the Source interface for default configuration values
type DefaultSource struct {
	values   map[string]interface{}
	priority int
}

// NewDefaultSource creates a new default value configuration source
func NewDefaultSource(defaults map[string]interface{}) *DefaultSource {
	values := make(map[string]interface{})
	for k, v := range defaults {
		values[k] = v
	}
	
	return &DefaultSource{
		values:   values,
		priority: 10, // Lowest priority
	}
}

// Name implements the Source interface
func (ds *DefaultSource) Name() string {
	return "defaults"
}

// Priority implements the Source interface
func (ds *DefaultSource) Priority() int {
	return ds.priority
}

// Load implements the Source interface
func (ds *DefaultSource) Load(ctx context.Context) (map[string]interface{}, error) {
	// Return a copy of default values
	result := make(map[string]interface{})
	for k, v := range ds.values {
		result[k] = v
	}
	return result, nil
}

// AddDefault adds or updates a default value
func (ds *DefaultSource) AddDefault(key string, value interface{}) {
	ds.values[key] = value
}

// RemoveDefault removes a default value
func (ds *DefaultSource) RemoveDefault(key string) {
	delete(ds.values, key)
}

// GetDefaults returns all default values
func (ds *DefaultSource) GetDefaults() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range ds.values {
		result[k] = v
	}
	return result
}

// SetPriority sets the priority of the default source
func (ds *DefaultSource) SetPriority(priority int) {
	ds.priority = priority
}

// AddValidator adds a custom validator function to the configuration
func (c *Config) AddValidator(validator ValidatorFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metadata.Validators = append(c.metadata.Validators, validator)
}

// AddFieldMetadata adds metadata for a configuration field
func (c *Config) AddFieldMetadata(fieldName string, field Field) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.metadata.Fields == nil {
		c.metadata.Fields = make(map[string]Field)
	}
	c.metadata.Fields[fieldName] = field
}

// RemoveFieldMetadata removes metadata for a configuration field
func (c *Config) RemoveFieldMetadata(fieldName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.metadata.Fields, fieldName)
}

// GetFieldMetadata returns metadata for a specific field
func (c *Config) GetFieldMetadata(fieldName string) (Field, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	field, exists := c.metadata.Fields[fieldName]
	return field, exists
}

// Summary returns a summary of the configuration state
func (c *Config) Summary() ConfigSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := ConfigSummary{
		Environment:    c.environment,
		TotalKeys:      len(c.values),
		TotalSources:   len(c.sources),
		TotalWatchers:  len(c.watchers),
		Sources:        c.GetSources(),
		RequiredFields: make([]string, 0),
		SensitiveFields: make([]string, 0),
	}

	// Collect required and sensitive fields
	for fieldName, field := range c.metadata.Fields {
		if field.Required {
			summary.RequiredFields = append(summary.RequiredFields, fieldName)
		}
		if field.Sensitive {
			summary.SensitiveFields = append(summary.SensitiveFields, fieldName)
		}
	}

	return summary
}

// ConfigSummary provides a summary of configuration state
type ConfigSummary struct {
	Environment     string       `json:"environment"`
	TotalKeys       int          `json:"total_keys"`
	TotalSources    int          `json:"total_sources"`
	TotalWatchers   int          `json:"total_watchers"`
	Sources         []SourceInfo `json:"sources"`
	RequiredFields  []string     `json:"required_fields"`
	SensitiveFields []string     `json:"sensitive_fields"`
}
		