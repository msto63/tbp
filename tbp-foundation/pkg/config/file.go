// File: file.go
// Title: File-based Configuration for TBP
// Description: Provides file-based configuration loading with support for TOML,
//              YAML, and JSON formats. Includes file watching for hot-reload,
//              environment variable substitution, and hierarchical configuration
//              merging with validation and error handling.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial file-based configuration implementation

package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"

	"github.com/msto63/tbp/tbp-foundation/pkg/core"
)

// FileSource implements the Source interface for file-based configuration
type FileSource struct {
	// mu protects concurrent access to file source data
	mu sync.RWMutex

	// path is the file path to load configuration from
	path string

	// format specifies the file format (toml, yaml, json, auto)
	format string

	// optional indicates whether the file is optional (no error if missing)
	optional bool

	// watchEnabled indicates whether file watching is enabled
	watchEnabled bool

	// lastModified tracks the last modification time for change detection
	lastModified time.Time

	// values stores the loaded configuration values
	values map[string]interface{}

	// callbacks stores registered change callbacks
	callbacks []func(map[string]interface{})

	// stopWatching is used to stop the file watcher
	stopWatching chan struct{}
}

// FileSourceOptions configures file source creation
type FileSourceOptions struct {
	Path         string `json:"path"`
	Format       string `json:"format"`       // toml, yaml, json, auto (default: auto)
	Optional     bool   `json:"optional"`     // true if file is optional
	WatchEnabled bool   `json:"watch_enabled"` // true to enable file watching
}

// NewFileSource creates a new file-based configuration source
func NewFileSource(opts FileSourceOptions) (*FileSource, error) {
	if opts.Path == "" {
		return nil, core.New("file path is required")
	}

	if opts.Format == "" {
		opts.Format = "auto"
	}

	fs := &FileSource{
		path:         opts.Path,
		format:       opts.Format,
		optional:     opts.Optional,
		watchEnabled: opts.WatchEnabled,
		values:       make(map[string]interface{}),
		callbacks:    make([]func(map[string]interface{}), 0),
		stopWatching: make(chan struct{}),
	}

	return fs, nil
}

// Name implements the Source interface
func (fs *FileSource) Name() string {
	return fmt.Sprintf("file:%s", fs.path)
}

// Priority implements the Source interface
func (fs *FileSource) Priority() int {
	// File sources have medium priority (higher than defaults, lower than env vars)
	return 50
}

// Load implements the Source interface
func (fs *FileSource) Load(ctx context.Context) (map[string]interface{}, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Check if file exists
	info, err := os.Stat(fs.path)
	if err != nil {
		if os.IsNotExist(err) && fs.optional {
			// File doesn't exist but is optional - return empty values
			return make(map[string]interface{}), nil
		}
		return nil, core.Wrapf(err, "failed to access configuration file %s", fs.path)
	}

	// Check if file has been modified since last load
	if !fs.lastModified.IsZero() && !info.ModTime().After(fs.lastModified) {
		// File hasn't changed, return cached values
		return fs.copyValues(), nil
	}

	// Read file content
	content, err := os.ReadFile(fs.path)
	if err != nil {
		return nil, core.Wrapf(err, "failed to read configuration file %s", fs.path)
	}

	// Substitute environment variables
	content, err = fs.substituteEnvVars(content)
	if err != nil {
		return nil, core.Wrapf(err, "failed to substitute environment variables in %s", fs.path)
	}

	// Determine file format
	format := fs.format
	if format == "auto" {
		format = fs.detectFormat()
	}

	// Parse file content based on format
	values, err := fs.parseContent(content, format)
	if err != nil {
		return nil, core.Wrapf(err, "failed to parse configuration file %s as %s", fs.path, format)
	}

	// Flatten nested structures for consistent key access
	flatValues := fs.flattenMap(values, "")

	// Update cached values and modification time
	fs.values = flatValues
	fs.lastModified = info.ModTime()

	return fs.copyValues(), nil
}

// Watch implements the Source interface
func (fs *FileSource) Watch(ctx context.Context, callback func(map[string]interface{})) error {
	if !fs.watchEnabled {
		return nil // Watching is disabled
	}

	fs.mu.Lock()
	fs.callbacks = append(fs.callbacks, callback)
	fs.mu.Unlock()

	// Start file watcher in a separate goroutine
	go fs.watchFile(ctx)

	return nil
}

// detectFormat automatically detects the file format based on extension
func (fs *FileSource) detectFormat() string {
	ext := strings.ToLower(filepath.Ext(fs.path))
	switch ext {
	case ".toml", ".tml":
		return "toml"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		// Default to TOML if extension is unknown
		return "toml"
	}
}

// parseContent parses file content based on the specified format
func (fs *FileSource) parseContent(content []byte, format string) (map[string]interface{}, error) {
	var values map[string]interface{}

	switch format {
	case "toml":
		if err := toml.Unmarshal(content, &values); err != nil {
			return nil, core.Wrap(err, "failed to parse TOML")
		}

	case "yaml":
		if err := yaml.Unmarshal(content, &values); err != nil {
			return nil, core.Wrap(err, "failed to parse YAML")
		}

	case "json":
		if err := json.Unmarshal(content, &values); err != nil {
			return nil, core.Wrap(err, "failed to parse JSON")
		}

	default:
		return nil, core.Newf("unsupported configuration format: %s", format)
	}

	return values, nil
}

// substituteEnvVars substitutes environment variables in the configuration content
// Supports ${VAR} and ${VAR:-default} syntax
func (fs *FileSource) substituteEnvVars(content []byte) ([]byte, error) {
	// Regex to match ${VAR} and ${VAR:-default} patterns
	envVarRegex := regexp.MustCompile(`\$\{([^}]+)\}`)
	
	result := envVarRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		// Extract variable specification (without ${})
		varSpec := string(match[2 : len(match)-1])
		
		// Check for default value syntax (VAR:-default)
		var varName, defaultValue string
		if idx := strings.Index(varSpec, ":-"); idx != -1 {
			varName = varSpec[:idx]
			defaultValue = varSpec[idx+2:]
		} else {
			varName = varSpec
		}
		
		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return []byte(value)
		}
		
		// Return default value if provided, otherwise return original match
		if defaultValue != "" {
			return []byte(defaultValue)
		}
		
		return match
	})
	
	return result, nil
}

// flattenMap flattens nested maps into dot-separated keys
func (fs *FileSource) flattenMap(data map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively flatten nested maps
			for nestedKey, nestedValue := range fs.flattenMap(nestedMap, fullKey) {
				result[nestedKey] = nestedValue
			}
		} else {
			result[fullKey] = value
		}
	}
	
	return result
}

// copyValues returns a copy of the current values to prevent external modification
func (fs *FileSource) copyValues() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range fs.values {
		result[key] = value
	}
	return result
}

// watchFile monitors the configuration file for changes
func (fs *FileSource) watchFile(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // Check for changes every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fs.stopWatching:
			return
		case <-ticker.C:
			fs.checkForChanges(ctx)
		}
	}
}

// checkForChanges checks if the file has been modified and reloads if necessary
func (fs *FileSource) checkForChanges(ctx context.Context) {
	info, err := os.Stat(fs.path)
	if err != nil {
		if !os.IsNotExist(err) {
			// Log error but continue watching
			fmt.Printf("Error checking file %s: %v\n", fs.path, err)
		}
		return
	}

	fs.mu.RLock()
	lastModified := fs.lastModified
	fs.mu.RUnlock()

	// Check if file has been modified
	if info.ModTime().After(lastModified) {
		// File has changed, reload configuration
		values, err := fs.Load(ctx)
		if err != nil {
			fmt.Printf("Error reloading configuration from %s: %v\n", fs.path, err)
			return
		}

		// Notify callbacks
		fs.mu.RLock()
		callbacks := make([]func(map[string]interface{}), len(fs.callbacks))
		copy(callbacks, fs.callbacks)
		fs.mu.RUnlock()

		for _, callback := range callbacks {
			go func(cb func(map[string]interface{})) {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Panic in file watcher callback: %v\n", r)
					}
				}()
				cb(values)
			}(callback)
		}
	}
}

// Stop stops the file watcher
func (fs *FileSource) Stop() {
	close(fs.stopWatching)
}

// WriteConfig writes configuration values to the file
func (fs *FileSource) WriteConfig(values map[string]interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Determine format for writing
	format := fs.format
	if format == "auto" {
		format = fs.detectFormat()
	}

	// Convert flat values back to nested structure
	nestedValues := fs.unflattenMap(values)

	// Serialize based on format
	var content []byte
	var err error

	switch format {
	case "toml":
		var buf strings.Builder
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(nestedValues); err != nil {
			return core.Wrap(err, "failed to encode TOML")
		}
		content = []byte(buf.String())

	case "yaml":
		content, err = yaml.Marshal(nestedValues)
		if err != nil {
			return core.Wrap(err, "failed to encode YAML")
		}

	case "json":
		content, err = json.MarshalIndent(nestedValues, "", "  ")
		if err != nil {
			return core.Wrap(err, "failed to encode JSON")
		}

	default:
		return core.Newf("unsupported format for writing: %s", format)
	}

	// Write to file
	if err := os.WriteFile(fs.path, content, 0644); err != nil {
		return core.Wrapf(err, "failed to write configuration file %s", fs.path)
	}

	return nil
}

// unflattenMap converts flat dot-separated keys back to nested structure
func (fs *FileSource) unflattenMap(flat map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range flat {
		parts := strings.Split(key, ".")
		current := result

		// Navigate/create nested structure
		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part, set the value
				current[part] = value
			} else {
				// Intermediate part, ensure nested map exists
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]interface{})
				}
				if nested, ok := current[part].(map[string]interface{}); ok {
					current = nested
				}
			}
		}
	}

	return result
}

// Validate validates the file source configuration
func (fs *FileSource) Validate() error {
	if fs.path == "" {
		return core.New("file path cannot be empty")
	}

	// Check if file exists (if not optional)
	if !fs.optional {
		if _, err := os.Stat(fs.path); err != nil {
			if os.IsNotExist(err) {
				return core.Newf("required configuration file %s does not exist", fs.path)
			}
			return core.Wrapf(err, "cannot access configuration file %s", fs.path)
		}
	}

	// Validate format
	validFormats := []string{"auto", "toml", "yaml", "json"}
	validFormat := false
	for _, format := range validFormats {
		if fs.format == format {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return core.Newf("unsupported file format: %s", fs.format)
	}

	return nil
}

// GetPath returns the file path
func (fs *FileSource) GetPath() string {
	return fs.path
}

// GetFormat returns the file format
func (fs *FileSource) GetFormat() string {
	return fs.format
}

// IsOptional returns whether the file is optional
func (fs *FileSource) IsOptional() bool {
	return fs.optional
}

// IsWatchEnabled returns whether file watching is enabled
func (fs *FileSource) IsWatchEnabled() bool {
	return fs.watchEnabled
}

// GetLastModified returns the last modification time
func (fs *FileSource) GetLastModified() time.Time {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lastModified
}

// LoadFromReader loads configuration from an io.Reader (useful for testing)
func (fs *FileSource) LoadFromReader(reader io.Reader, format string) (map[string]interface{}, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, core.Wrap(err, "failed to read content")
	}

	// Substitute environment variables
	content, err = fs.substituteEnvVars(content)
	if err != nil {
		return nil, core.Wrap(err, "failed to substitute environment variables")
	}

	// Parse content
	values, err := fs.parseContent(content, format)
	if err != nil {
		return nil, core.Wrapf(err, "failed to parse content as %s", format)
	}

	// Flatten the values
	return fs.flattenMap(values, ""), nil
}