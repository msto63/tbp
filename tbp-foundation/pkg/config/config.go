// File: config.go
// Title: Configuration Management for TBP
// Description: Provides multi-layered configuration management supporting
//              environment variables, configuration files, command-line flags,
//              and remote configuration sources. Implements type-safe configuration
//              structures with validation, hot-reloading, and sensitive data protection.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial configuration management implementation

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
}

// Source represents a configuration source (env vars, files, etc.)
type Source interface {
	// Name returns the source name for logging and debugging
	Name() string

	// Load loads configuration values from the source
	Load(ctx context.Context) (map[string]interface{}, error)

	// Watch monitors the source for changes (optional)
	Watch(ctx context.Context, callback func(map[string]interface{})) error

	// Priority returns the source priority (higher = more important)
	Priority() int
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
}

// ValidatorFunc validates configuration values
type ValidatorFunc func(key string, value interface{}) error

// LoadOptions configures how configuration is loaded
type LoadOptions struct {
	Sources      []Source          `json:"-"`
	Environment  string            `json:"environment"`
	ConfigPaths  []string          `json:"config_paths"`
	EnvPrefix    string            `json:"env_prefix"`
	Defaults     map[string]interface{} `json:"defaults"`
	Validation   bool              `json:"validation"`
	HotReload    bool              `json:"hot_reload"`
	Metadata     *Metadata         `json:"metadata,omitempty"`
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
		sources:  make([]Source, 0),
		values:   make(map[string]interface{}),
		watchers: make([]Watcher, 0),
		metadata: opts.Metadata,
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

		// Add file sources for each config path (TOML format)
		for _, path := range opts.ConfigPaths {
			fileSource, err := NewFileSource(FileSourceOptions{
				Path:     path,
				Format:   "toml", // Default to TOML format
				Optional: true,
			})
			if err != nil {
				return nil, core.Wrapf(err, "failed to create TOML file source for %s", path)
			}
			opts.Sources = append(opts.Sources, fileSource)
		}
	}

	// Add sources to configuration
	for _, source := range opts.Sources {
		config.AddSource(source)
	}

	// Set default values
	if len(opts.Defaults) > 0 {
		defaultSource := NewDefaultSource(opts.Defaults)
		config.AddSource(defaultSource)
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
func (c *Config) AddSource(source Source) {
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

	// Validate required fields
	for fieldName, field := range c.metadata.Fields {
		if field.Required {
			if _, exists := c.values[fieldName]; !exists {
				return core.Newf("required configuration field '%s' is missing", fieldName)
			}
		}
	}

	// Run custom validators
	for _, validator := range c.metadata.Validators {
		for key, value := range c.values {
			if err := validator(key, value); err != nil {
				return core.Wrapf(err, "validation failed for field '%s'", key)
			}
		}
	}

	return nil
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
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		if i, err := fmt.Sscanf(v, "%d", new(int)); err == nil && i == 1 {
			var result int
			fmt.Sscanf(v, "%d", &result)
			return result, nil
		}
	}

	return 0, core.Newf("configuration key '%s' cannot be converted to int", key)
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
		lower := strings.ToLower(v)
		switch lower {
		case "true", "yes", "1", "on", "enable", "enabled":
			return true, nil
		case "false", "no", "0", "off", "disable", "disabled":
			return false, nil
		}
	}

	return false, core.Newf("configuration key '%s' cannot be converted to bool", key)
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
	case int, int64, float64:
		// Assume seconds if numeric value provided
		seconds := 0.0
		switch num := v.(type) {
		case int:
			seconds = float64(num)
		case int64:
			seconds = float64(num)
		case float64:
			seconds = num
		}
		return time.Duration(seconds * float64(time.Second)), nil
	}

	return 0, core.Newf("configuration key '%s' cannot be converted to duration", key)
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
		if fieldValue.Kind() == reflect.Struct {
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
		case int64:
			intVal = v
		case float64:
			intVal = int64(v)
		case string:
			if i, err := fmt.Sscanf(v, "%d", &intVal); err != nil || i != 1 {
				return core.Newf("cannot convert '%v' to int", value)
			}
		default:
			return core.Newf("cannot convert '%v' to int", value)
		}
		rv.SetInt(intVal)

	case reflect.Bool:
		var boolVal bool
		switch v := value.(type) {
		case bool:
			boolVal = v
		case string:
			lower := strings.ToLower(v)
			switch lower {
			case "true", "yes", "1", "on":
				boolVal = true
			case "false", "no", "0", "off":
				boolVal = false
			default:
				return core.Newf("cannot convert '%v' to bool", value)
			}
		default:
			return core.Newf("cannot convert '%v' to bool", value)
		}
		rv.SetBool(boolVal)

	case reflect.Float32, reflect.Float64:
		var floatVal float64
		switch v := value.(type) {
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		case string:
			if f, err := fmt.Sscanf(v, "%f", &floatVal); err != nil || f != 1 {
				return core.Newf("cannot convert '%v' to float", value)
			}
		default:
			return core.Newf("cannot convert '%v' to float", value)
		}
		rv.SetFloat(floatVal)

	default:
		return core.Newf("unsupported field type: %s", rv.Kind())
	}

	return nil
}

// AddWatcher adds a configuration change watcher
func (c *Config) AddWatcher(watcher Watcher) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.watchers = append(c.watchers, watcher)
}

// StartWatching starts watching all sources for configuration changes
func (c *Config) StartWatching(ctx context.Context) error {
	for _, source := range c.sources {
		go func(s Source) {
			err := s.Watch(ctx, func(values map[string]interface{}) {
				// Reload configuration when source changes
				if err := c.Load(ctx); err != nil {
					// Log error but continue watching
					// In a real implementation, this would use the logging package
					fmt.Printf("Error reloading configuration from %s: %v\n", s.Name(), err)
				}
			})
			if err != nil {
				fmt.Printf("Error watching source %s: %v\n", s.Name(), err)
			}
		}(source)
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
				}
			}
		} else {
			changes[key] = ConfigChange{
				Key:      key,
				NewValue: newValue,
				Source:   "merged",
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
			}
		}
	}

	return changes
}

// notifyWatchers notifies all registered watchers of configuration changes
func (c *Config) notifyWatchers(ctx context.Context, changes map[string]ConfigChange) {
	for _, watcher := range c.watchers {
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

// GetMetadata returns configuration metadata
func (c *Config) GetMetadata() *Metadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.metadata
}

// DefaultSource implements the Source interface for default configuration values
type DefaultSource struct {
	values map[string]interface{}
}

// NewDefaultSource creates a new default value configuration source
func NewDefaultSource(defaults map[string]interface{}) *DefaultSource {
	values := make(map[string]interface{})
	for k, v := range defaults {
		values[k] = v
	}
	
	return &DefaultSource{
		values: values,
	}
}

// Name implements the Source interface
func (ds *DefaultSource) Name() string {
	return "defaults"
}

// Priority implements the Source interface
func (ds *DefaultSource) Priority() int {
	// Default values have the lowest priority
	return 10
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

// Watch implements the Source interface
func (ds *DefaultSource) Watch(ctx context.Context, callback func(map[string]interface{})) error {
	// Default values don't change, so no watching needed
	return nil
}

// AddDefault adds or updates a default value
func (ds *DefaultSource) AddDefault(key string, value interface{}) {
	ds.values[key] = value
}

// RemoveDefault removes a default value
func (ds *DefaultSource) RemoveDefault(key string) {
	delete(ds.values, key)
}

// Close cleanly shuts down the configuration manager
func (c *Config) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear all data
	c.sources = nil
	c.values = nil
	c.watchers = nil

	return nil
}