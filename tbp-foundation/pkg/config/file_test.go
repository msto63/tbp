// File: file_test.go
// Title: Tests for File-based Configuration Source
// Description: Comprehensive test suite for file-based configuration loading
//              including TOML, YAML, JSON parsing, file watching, environment variable
//              expansion, and error handling. Tests cover various file formats,
//              hot-reloading scenarios, and edge cases.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.1
// Created: 2025-05-26
// Modified: 2025-05-27
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation for file-based configuration
// - 2025-05-27 v0.1.1: Fixed tests for array indexing and YAML support

package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileSource(t *testing.T) {
	t.Run("creates file source with valid options", func(t *testing.T) {
		opts := FileSourceOptions{
			Path:     "config.toml",
			Format:   "toml",
			Optional: true,
			Priority: 100,
		}

		source, err := NewFileSource(opts)
		require.NoError(t, err)
		assert.Equal(t, "config.toml", source.GetPath())
		assert.Equal(t, "toml", source.GetFormat())
		assert.True(t, source.IsOptional())
		assert.Equal(t, 100, source.Priority())
		assert.Equal(t, "file:config.toml", source.Name())
	})

	t.Run("detects format from file extension", func(t *testing.T) {
		testCases := []struct {
			path           string
			expectedFormat string
		}{
			{"config.toml", "toml"},
			{"config.yaml", "yaml"},
			{"config.yml", "yaml"},
			{"config.json", "json"},
			{"config.unknown", "toml"}, // Default to TOML
		}

		for _, tc := range testCases {
			t.Run(tc.path, func(t *testing.T) {
				opts := FileSourceOptions{Path: tc.path, Optional: true}
				source, err := NewFileSource(opts)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedFormat, source.GetFormat())
			})
		}
	})

	t.Run("sets default priority", func(t *testing.T) {
		opts := FileSourceOptions{Path: "config.toml", Optional: true}
		source, err := NewFileSource(opts)
		require.NoError(t, err)
		assert.Equal(t, 50, source.Priority()) // Default priority
	})

	t.Run("returns error for empty path", func(t *testing.T) {
		opts := FileSourceOptions{Path: "", Optional: true}
		_, err := NewFileSource(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file path is required")
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		opts := FileSourceOptions{
			Path:     "config.txt",
			Format:   "xml",
			Optional: true,
		}
		_, err := NewFileSource(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported configuration format")
	})
}

func TestFileSource_Load(t *testing.T) {
	t.Run("loads TOML file successfully", func(t *testing.T) {
		// Create temporary TOML file
		tmpFile := createTempFile(t, "config.toml", `
# Test TOML configuration
environment = "test"
debug = true

[server]
host = "localhost"
port = 8080
timeout = "30s"

[database]
host = "db.example.com"
port = 5432
name = "testdb"
ssl_mode = "require"

[array_example]
tags = ["web", "api", "service"]
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "toml",
		})
		require.NoError(t, err)

		values, err := source.Load(context.Background())
		require.NoError(t, err)

		// Test flattened keys
		assert.Equal(t, "test", values["environment"])
		assert.Equal(t, true, values["debug"])
		assert.Equal(t, "localhost", values["server.host"])
		assert.Equal(t, int64(8080), values["server.port"])
		assert.Equal(t, "30s", values["server.timeout"])
		assert.Equal(t, "db.example.com", values["database.host"])
		assert.Equal(t, int64(5432), values["database.port"])
		assert.Equal(t, "testdb", values["database.name"])
		assert.Equal(t, "require", values["database.ssl_mode"])

		// Test array handling - both original array and indexed access
		tags, ok := values["array_example.tags"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, tags, 3)
		assert.Equal(t, "web", tags[0])
		assert.Equal(t, "api", tags[1])
		assert.Equal(t, "service", tags[2])

		// Test indexed access
		assert.Equal(t, "web", values["array_example.tags.0"])
		assert.Equal(t, "api", values["array_example.tags.1"])
		assert.Equal(t, "service", values["array_example.tags.2"])
	})

	t.Run("loads JSON file successfully", func(t *testing.T) {
		tmpFile := createTempFile(t, "config.json", `{
  "environment": "test",
  "debug": true,
  "server": {
    "host": "localhost",
    "port": 8080
  },
  "nested": {
    "deep": {
      "value": "test"
    }
  },
  "array_data": [
    {"name": "item1", "value": 1},
    {"name": "item2", "value": 2}
  ]
}`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "json",
		})
		require.NoError(t, err)

		values, err := source.Load(context.Background())
		require.NoError(t, err)

		assert.Equal(t, "test", values["environment"])
		assert.Equal(t, true, values["debug"])
		assert.Equal(t, "localhost", values["server.host"])
		assert.Equal(t, float64(8080), values["server.port"]) // JSON numbers are float64
		assert.Equal(t, "test", values["nested.deep.value"])

		// Test array with objects
		assert.Equal(t, "item1", values["array_data.0.name"])
		assert.Equal(t, float64(1), values["array_data.0.value"])
		assert.Equal(t, "item2", values["array_data.1.name"])
		assert.Equal(t, float64(2), values["array_data.1.value"])
	})

	t.Run("loads YAML file successfully", func(t *testing.T) {
		tmpFile := createTempFile(t, "config.yaml", `
environment: test
debug: true
server:
  host: localhost
  port: 8080
  timeout: 30s
database:
  host: db.example.com
  port: 5432
  name: testdb
  ssl_mode: require
array_example:
  tags:
    - web
    - api  
    - service
  servers:
    - name: server1
      port: 8080
    - name: server2
      port: 8081
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "yaml",
		})
		require.NoError(t, err)

		values, err := source.Load(context.Background())
		require.NoError(t, err)

		// Test basic values
		assert.Equal(t, "test", values["environment"])
		assert.Equal(t, true, values["debug"])
		assert.Equal(t, "localhost", values["server.host"])
		assert.Equal(t, 8080, values["server.port"])
		assert.Equal(t, "30s", values["server.timeout"])

		// Test arrays
		assert.Equal(t, "web", values["array_example.tags.0"])
		assert.Equal(t, "api", values["array_example.tags.1"])
		assert.Equal(t, "service", values["array_example.tags.2"])

		// Test nested arrays with objects
		assert.Equal(t, "server1", values["array_example.servers.0.name"])
		assert.Equal(t, 8080, values["array_example.servers.0.port"])
		assert.Equal(t, "server2", values["array_example.servers.1.name"])
		assert.Equal(t, 8081, values["array_example.servers.1.port"])
	})

	t.Run("expands environment variables", func(t *testing.T) {
		// Set test environment variable
		os.Setenv("TEST_HOST", "example.com")
		os.Setenv("TEST_PORT", "9000")
		defer func() {
			os.Unsetenv("TEST_HOST")
			os.Unsetenv("TEST_PORT")
		}()

		tmpFile := createTempFile(t, "config.toml", `
environment = "${NODE_ENV:-development}"
host = "${TEST_HOST}"
port = "${TEST_PORT}"
url = "https://${TEST_HOST}:${TEST_PORT}/api"
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "toml",
		})
		require.NoError(t, err)

		values, err := source.Load(context.Background())
		require.NoError(t, err)

		assert.Equal(t, "development", values["environment"]) // Default value
		assert.Equal(t, "example.com", values["host"])
		assert.Equal(t, "9000", values["port"])
		assert.Equal(t, "https://example.com:9000/api", values["url"])
	})

	t.Run("handles optional missing file", func(t *testing.T) {
		source, err := NewFileSource(FileSourceOptions{
			Path:     "nonexistent.toml",
			Optional: true,
		})
		require.NoError(t, err)

		values, err := source.Load(context.Background())
		require.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("returns error for required missing file", func(t *testing.T) {
		source, err := NewFileSource(FileSourceOptions{
			Path:     "nonexistent.toml",
			Optional: false,
		})
		require.NoError(t, err)

		_, err = source.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to access configuration file")
	})

	t.Run("returns error for invalid TOML", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.toml", `
[server
port = 8080
invalid toml content
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "toml",
		})
		require.NoError(t, err)

		_, err = source.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse TOML")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.json", `{
  "server": {
    "port": 8080,
  }
  invalid json
}`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "json",
		})
		require.NoError(t, err)

		_, err = source.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.yaml", `
server:
  host: localhost
  port: 8080
    invalid_indentation: true
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "yaml",
		})
		require.NoError(t, err)

		_, err = source.Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse YAML")
	})
}

func TestFileSource_FlattenMap(t *testing.T) {
	source := &FileSource{}

	t.Run("flattens nested maps", func(t *testing.T) {
		input := map[string]interface{}{
			"simple": "value",
			"nested": map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"deep": "deepvalue",
				},
			},
		}

		result := source.flattenMap(input, "")

		assert.Equal(t, "value", result["simple"])
		assert.Equal(t, "value1", result["nested.key1"])
		assert.Equal(t, "deepvalue", result["nested.key2.deep"])
	})

	t.Run("handles arrays with indexing", func(t *testing.T) {
		input := map[string]interface{}{
			"tags": []interface{}{"web", "api", "service"},
			"servers": []interface{}{
				map[string]interface{}{
					"name": "server1",
					"port": 8080,
				},
				map[string]interface{}{
					"name": "server2",
					"port": 8081,
				},
			},
		}

		result := source.flattenMap(input, "")

		// Check original array is preserved
		tags, ok := result["tags"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, tags, 3)

		// Check indexed access
		assert.Equal(t, "web", result["tags.0"])
		assert.Equal(t, "api", result["tags.1"])
		assert.Equal(t, "service", result["tags.2"])

		// Check nested object arrays
		assert.Equal(t, "server1", result["servers.0.name"])
		assert.Equal(t, 8080, result["servers.0.port"])
		assert.Equal(t, "server2", result["servers.1.name"])
		assert.Equal(t, 8081, result["servers.1.port"])
	})

	t.Run("handles empty arrays", func(t *testing.T) {
		input := map[string]interface{}{
			"empty_array": []interface{}{},
			"null_array":  nil,
		}

		result := source.flattenMap(input, "")

		// Empty array should be preserved
		emptyArray, ok := result["empty_array"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, emptyArray, 0)

		// Nil should be preserved
		assert.Nil(t, result["null_array"])
	})

	t.Run("handles mixed arrays", func(t *testing.T) {
		input := map[string]interface{}{
			"mixed": []interface{}{
				"string",
				42,
				true,
				map[string]interface{}{
					"nested": "object",
				},
			},
		}

		result := source.flattenMap(input, "")

		assert.Equal(t, "string", result["mixed.0"])
		assert.Equal(t, 42, result["mixed.1"])
		assert.Equal(t, true, result["mixed.2"])
		assert.Equal(t, "object", result["mixed.3.nested"])
	})
}

func TestFileSource_Watch(t *testing.T) {
	t.Run("detects file changes", func(t *testing.T) {
		// Create temporary file
		tmpFile := createTempFile(t, "config.toml", `
environment = "test"
port = 8080
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:         tmpFile,
			Format:       "toml",
			WatchEnabled: true,
		})
		require.NoError(t, err)

		// Load initial configuration
		_, err = source.Load(context.Background())
		require.NoError(t, err)

		// Set up watcher
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		changeChan := make(chan map[string]interface{}, 1)
		err = source.Watch(ctx, func(values map[string]interface{}) {
			changeChan <- values
		})
		require.NoError(t, err)

		// Wait a bit to ensure watcher is active
		time.Sleep(100 * time.Millisecond)

		// Modify the file
		newContent := `
environment = "production"
port = 9000
new_key = "new_value"
`
		err = os.WriteFile(tmpFile, []byte(newContent), 0644)
		require.NoError(t, err)

		// Wait for change notification
		select {
		case values := <-changeChan:
			assert.Equal(t, "production", values["environment"])
			assert.Equal(t, int64(9000), values["port"])
			assert.Equal(t, "new_value", values["new_key"])
		case <-ctx.Done():
			t.Fatal("Did not receive change notification within timeout")
		}
	})

	t.Run("does not watch when disabled", func(t *testing.T) {
		tmpFile := createTempFile(t, "config.toml", `environment = "test"`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:         tmpFile,
			Format:       "toml",
			WatchEnabled: false,
		})
		require.NoError(t, err)

		ctx := context.Background()
		err = source.Watch(ctx, func(values map[string]interface{}) {
			t.Fatal("Should not receive callback when watching is disabled")
		})

		// Should not return error, just do nothing
		assert.NoError(t, err)
	})
}

func TestFileSource_Validate(t *testing.T) {
	t.Run("validates existing valid file", func(t *testing.T) {
		tmpFile := createTempFile(t, "config.toml", `
environment = "test"
port = 8080
`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:     tmpFile,
			Format:   "toml",
			Optional: false,
		})
		require.NoError(t, err)

		err = source.Validate()
		assert.NoError(t, err)
	})

	t.Run("validates optional missing file", func(t *testing.T) {
		source, err := NewFileSource(FileSourceOptions{
			Path:     "nonexistent.toml",
			Optional: true,
		})
		require.NoError(t, err)

		err = source.Validate()
		assert.NoError(t, err)
	})

	t.Run("returns error for required missing file", func(t *testing.T) {
		source, err := NewFileSource(FileSourceOptions{
			Path:     "nonexistent.toml",
			Optional: false,
		})
		require.NoError(t, err)

		err = source.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("returns error for invalid file content", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.toml", `invalid toml content`)
		defer os.Remove(tmpFile)

		source, err := NewFileSource(FileSourceOptions{
			Path:     tmpFile,
			Format:   "toml",
			Optional: false,
		})
		require.NoError(t, err)

		err = source.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is invalid")
	})
}

func TestFileSource_WriteConfig(t *testing.T) {
	t.Run("writes TOML config", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "write_test.toml")

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "toml",
		})
		require.NoError(t, err)

		values := map[string]interface{}{
			"environment":   "production",
			"server.host":   "0.0.0.0",
			"server.port":   9000,
			"database.name": "prod_db",
			"tags":          []interface{}{"prod", "live"},
		}

		err = source.WriteConfig(values)
		require.NoError(t, err)

		// Verify file was written and can be read back
		loadedValues, err := source.Load(context.Background())
		require.NoError(t, err)

		assert.Equal(t, "production", loadedValues["environment"])
		assert.Equal(t, "0.0.0.0", loadedValues["server.host"])
		assert.Equal(t, int64(9000), loadedValues["server.port"])
		assert.Equal(t, "prod_db", loadedValues["database.name"])
	})

	t.Run("writes JSON config", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "write_test.json")

		source, err := NewFileSource(FileSourceOptions{
			Path:   tmpFile,
			Format: "json",
		})
		require.NoError(t, err)

		values := map[string]interface{}{
			"environment": "test",
			"server.host": "localhost",
			"server.port": 8080,
		}

		err = source.WriteConfig(values)
		require.NoError(t, err)

		// Verify file was written
		assert.True(t, source.Exists())

		// Verify content can be loaded back
		loadedValues, err := source.Load(context.Background())
		require.NoError(t, err)

		assert.Equal(t, "test", loadedValues["environment"])
		assert.Equal(t, "localhost", loadedValues["server.host"])
		assert.Equal(t, float64(8080), loadedValues["server.port"]) // JSON numbers are float64
	})
}

func TestFileSource_Utilities(t *testing.T) {
	tmpFile := createTempFile(t, "config.toml", "environment = \"test\"")
	defer os.Remove(tmpFile)

	source, err := NewFileSource(FileSourceOptions{
		Path:   tmpFile,
		Format: "toml",
	})
	require.NoError(t, err)

	t.Run("exists check", func(t *testing.T) {
		assert.True(t, source.Exists())

		nonExistentSource, err := NewFileSource(FileSourceOptions{
			Path:   "nonexistent.toml",
			Format: "toml",
		})
		require.NoError(t, err)
		assert.False(t, nonExistentSource.Exists())
	})

	t.Run("get last modified", func(t *testing.T) {
		modTime, err := source.GetLastModified()
		require.NoError(t, err)
		assert.False(t, modTime.IsZero())
		assert.True(t, modTime.Before(time.Now().Add(time.Second)))
	})

	t.Run("load from reader", func(t *testing.T) {
		content := `
environment = "from_reader"
debug = true
`
		reader := strings.NewReader(content)
		values, err := source.LoadFromReader(reader, "toml")
		require.NoError(t, err)

		assert.Equal(t, "from_reader", values["environment"])
		assert.Equal(t, true, values["debug"])
	})
}

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, name, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, name)

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	return tmpFile
}

// Benchmark tests for performance validation
func BenchmarkFileSource_Load_TOML(b *testing.B) {
	tmpFile := createTempFileForBench(b, "config.toml", `
environment = "production"
debug = false

[server]
host = "0.0.0.0"
port = 8080
timeout = "30s"

[database]
host = "localhost"
port = 5432
name = "tbp_production"
ssl_mode = "require"

[logging]
level = "info"
format = "json"
output = "stdout"

[array_example]
tags = ["prod", "api", "service", "backend"]
servers = [
  {name = "web1", port = 8080},
  {name = "web2", port = 8081},
  {name = "api1", port = 9080}
]
`)
	defer os.Remove(tmpFile)

	source, err := NewFileSource(FileSourceOptions{
		Path:   tmpFile,
		Format: "toml",
	})
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := source.Load(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileSource_Load_JSON(b *testing.B) {
	tmpFile := createTempFileForBench(b, "config.json", `{
  "environment": "production",
  "debug": false,
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "timeout": "30s"
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "tbp_production",
    "ssl_mode": "require"
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "array_example": {
    "tags": ["prod", "api", "service", "backend"],
    "servers": [
      {"name": "web1", "port": 8080},
      {"name": "web2", "port": 8081},
      {"name": "api1", "port": 9080}
    ]
  }
}`)
	defer os.Remove(tmpFile)

	source, err := NewFileSource(FileSourceOptions{
		Path:   tmpFile,
		Format: "json",
	})
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := source.Load(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileSource_Load_YAML(b *testing.B) {
	tmpFile := createTempFileForBench(b, "config.yaml", `
environment: production
debug: false
server:
  host: 0.0.0.0
  port: 8080
  timeout: 30s
database:
  host: localhost
  port: 5432
  name: tbp_production
  ssl_mode: require
logging:
  level: info
  format: json
  output: stdout
array_example:
  tags:
    - prod
    - api
    - service
    - backend
  servers:
    - name: web1
      port: 8080
    - name: web2
      port: 8081
    - name: api1
      port: 9080
`)
	defer os.Remove(tmpFile)

	source, err := NewFileSource(FileSourceOptions{
		Path:   tmpFile,
		Format: "yaml",
	})
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := source.Load(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileSource_FlattenMap(b *testing.B) {
	source := &FileSource{}
	input := map[string]interface{}{
		"simple": "value",
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{
				"deep": "deepvalue",
				"array": []interface{}{
					"item1", "item2", "item3",
					map[string]interface{}{
						"nested_array_item": "value",
					},
				},
			},
		},
		"servers": []interface{}{
			map[string]interface{}{
				"name": "server1",
				"port": 8080,
				"config": map[string]interface{}{
					"timeout": "30s",
					"ssl":     true,
				},
			},
			map[string]interface{}{
				"name": "server2",
				"port": 8081,
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = source.flattenMap(input, "")
	}
}

func BenchmarkFileSource_ExpandEnvVars(b *testing.B) {
	source := &FileSource{}

	// Set up environment variables
	os.Setenv("TEST_HOST", "example.com")
	os.Setenv("TEST_PORT", "8080")
	os.Setenv("TEST_DB", "mydb")
	defer func() {
		os.Unsetenv("TEST_HOST")
		os.Unsetenv("TEST_PORT")
		os.Unsetenv("TEST_DB")
	}()

	content := []byte(`
environment = "${NODE_ENV:-production}"
host = "${TEST_HOST}"
port = "${TEST_PORT}"
database = "${TEST_DB}"
url = "https://${TEST_HOST}:${TEST_PORT}/api/v1"
connection_string = "postgres://user:pass@${TEST_HOST}:5432/${TEST_DB}?sslmode=require"
`)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = source.substituteEnvVars(content)
	}
}

// Helper function for benchmarks
func createTempFileForBench(b *testing.B, name, content string) string {
	b.Helper()

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, name)

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(b, err)

	return tmpFile
}
