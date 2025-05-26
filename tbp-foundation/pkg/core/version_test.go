// File: version_test.go
// Title: Tests for Version Information Management
// Description: Comprehensive test suite for TBP version management including
//              semantic versioning, compatibility checks, build information,
//              and version comparison logic. Tests edge cases, parsing,
//              and enterprise version control scenarios.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial test implementation with comprehensive coverage

package core

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersionInfo(t *testing.T) {
	t.Run("returns version info", func(t *testing.T) {
		info := GetVersionInfo()
		
		assert.NotNil(t, info)
		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.GoVersion)
		assert.NotEmpty(t, info.Platform)
		assert.NotNil(t, info.Dependencies)
		
		// Check boolean flags
		assert.Equal(t, IsRelease(), info.IsRelease)
		assert.Equal(t, IsDevelopment(), info.IsDevelopment)
		assert.Equal(t, !IsRelease(), info.IsDevelopment)
	})

	t.Run("returns component-specific info", func(t *testing.T) {
		componentName := "test-service"
		info := GetVersionInfoForComponent(componentName)
		
		assert.Equal(t, componentName, info.ComponentName)
		assert.NotEmpty(t, info.Version)
	})
}

func TestVersionInfo_String(t *testing.T) {
	t.Run("formats version string correctly", func(t *testing.T) {
		info := &VersionInfo{
			ComponentName: "test-service",
			Version:       "v1.2.3",
			GitCommit:     "abc123def456",
			BuildDate:     "2024-01-15T10:30:00Z",
		}
		
		str := info.String()
		assert.Contains(t, str, "test-service")
		assert.Contains(t, str, "v1.2.3")
		assert.Contains(t, str, "commit:abc123d") // Short commit
		assert.Contains(t, str, "built:2024-01-15T10:30:00Z")
	})

	t.Run("handles missing component name", func(t *testing.T) {
		info := &VersionInfo{
			Version:   "v1.2.3",
			GitCommit: "abc123",
		}
		
		str := info.String()
		assert.Contains(t, str, "v1.2.3")
		assert.Contains(t, str, "commit:abc123")
		assert.NotContains(t, str, "test-service")
	})

	t.Run("handles unknown values", func(t *testing.T) {
		info := &VersionInfo{
			Version:   "v1.2.3",
			GitCommit: "unknown",
			BuildDate: "unknown",
		}
		
		str := info.String()
		assert.Contains(t, str, "v1.2.3")
		assert.NotContains(t, str, "commit:")
		assert.NotContains(t, str, "built:")
	})
}

func TestVersionGetters(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalGitCommit := GitCommit
	originalBuildDate := BuildDate
	
	// Set test values
	Version = "v1.2.3"
	GitCommit = "abc123def456ghi789"
	BuildDate = "2024-01-15T10:30:00Z"
	
	defer func() {
		// Restore original values
		Version = originalVersion
		GitCommit = originalGitCommit
		BuildDate = originalBuildDate
	}()

	t.Run("GetVersion", func(t *testing.T) {
		assert.Equal(t, "v1.2.3", GetVersion())
	})

	t.Run("GetShortVersion", func(t *testing.T) {
		assert.Equal(t, "1.2.3", GetShortVersion())
	})

	t.Run("GetGitCommit", func(t *testing.T) {
		assert.Equal(t, "abc123def456ghi789", GetGitCommit())
	})

	t.Run("GetShortGitCommit", func(t *testing.T) {
		assert.Equal(t, "abc123d", GetShortGitCommit())
	})

	t.Run("GetBuildDate", func(t *testing.T) {
		assert.Equal(t, "2024-01-15T10:30:00Z", GetBuildDate())
	})
}

func TestGetBuildTime(t *testing.T) {
	// Save original value
	originalBuildDate := BuildDate
	defer func() {
		BuildDate = originalBuildDate
	}()

	t.Run("parses RFC3339 format", func(t *testing.T) {
		BuildDate = "2024-01-15T10:30:00Z"
		
		buildTime, err := GetBuildTime()
		require.NoError(t, err)
		
		expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		assert.Equal(t, expected, buildTime)
	})

	t.Run("parses simple date format", func(t *testing.T) {
		BuildDate = "2024-01-15"
		
		buildTime, err := GetBuildTime()
		require.NoError(t, err)
		
		expected := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, buildTime)
	})

	t.Run("handles unknown build date", func(t *testing.T) {
		BuildDate = "unknown"
		
		_, err := GetBuildTime()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "build date unknown")
	})

	t.Run("handles invalid format", func(t *testing.T) {
		BuildDate = "invalid-date-format"
		
		_, err := GetBuildTime()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse build date")
	})
}

func TestReleaseDetection(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	testCases := []struct {
		version   string
		isRelease bool
	}{
		{"v1.0.0", true},
		{"v1.2.3", true},
		{"v2.0.0", true},
		{"v1.0.0-dev", false},
		{"v1.0.0-alpha", false},
		{"v1.0.0-alpha.1", false},
		{"v1.0.0-beta", false},
		{"v1.0.0-beta.2", false},
		{"v1.0.0-rc", false},
		{"v1.0.0-rc.1", false},
		{"v0.1.0-dev", false},
		{"v1.0.0-ALPHA", false}, // Case insensitive
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			Version = tc.version
			
			assert.Equal(t, tc.isRelease, IsRelease())
			assert.Equal(t, !tc.isRelease, IsDevelopment())
		})
	}
}

func TestSemVer_String(t *testing.T) {
	t.Run("basic version", func(t *testing.T) {
		v := SemVer{Major: 1, Minor: 2, Patch: 3}
		assert.Equal(t, "1.2.3", v.String())
	})

	t.Run("with pre-release", func(t *testing.T) {
		v := SemVer{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"}
		assert.Equal(t, "1.2.3-alpha.1", v.String())
	})

	t.Run("with build metadata", func(t *testing.T) {
		v := SemVer{Major: 1, Minor: 2, Patch: 3, Build: "20240115.123"}
		assert.Equal(t, "1.2.3+20240115.123", v.String())
	})

	t.Run("with pre-release and build", func(t *testing.T) {
		v := SemVer{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta", Build: "exp.sha.5114f85"}
		assert.Equal(t, "1.2.3-beta+exp.sha.5114f85", v.String())
	})
}

func TestSemVer_Compare(t *testing.T) {
	testCases := []struct {
		name     string
		v1       SemVer
		v2       SemVer
		expected int
	}{
		{
			name:     "equal versions",
			v1:       SemVer{1, 2, 3, "", ""},
			v2:       SemVer{1, 2, 3, "", ""},
			expected: 0,
		},
		{
			name:     "v1 major > v2 major",
			v1:       SemVer{2, 0, 0, "", ""},
			v2:       SemVer{1, 9, 9, "", ""},
			expected: 1,
		},
		{
			name:     "v1 major < v2 major",
			v1:       SemVer{1, 9, 9, "", ""},
			v2:       SemVer{2, 0, 0, "", ""},
			expected: -1,
		},
		{
			name:     "v1 minor > v2 minor",
			v1:       SemVer{1, 3, 0, "", ""},
			v2:       SemVer{1, 2, 9, "", ""},
			expected: 1,
		},
		{
			name:     "v1 minor < v2 minor",
			v1:       SemVer{1, 2, 9, "", ""},
			v2:       SemVer{1, 3, 0, "", ""},
			expected: -1,
		},
		{
			name:     "v1 patch > v2 patch",
			v1:       SemVer{1, 2, 4, "", ""},
			v2:       SemVer{1, 2, 3, "", ""},
			expected: 1,
		},
		{
			name:     "v1 patch < v2 patch",
			v1:       SemVer{1, 2, 3, "", ""},
			v2:       SemVer{1, 2, 4, "", ""},
			expected: -1,
		},
		{
			name:     "release > pre-release",
			v1:       SemVer{1, 2, 3, "", ""},
			v2:       SemVer{1, 2, 3, "alpha", ""},
			expected: 1,
		},
		{
			name:     "pre-release < release",
			v1:       SemVer{1, 2, 3, "alpha", ""},
			v2:       SemVer{1, 2, 3, "", ""},
			expected: -1,
		},
		{
			name:     "alpha < beta",
			v1:       SemVer{1, 2, 3, "alpha", ""},
			v2:       SemVer{1, 2, 3, "beta", ""},
			expected: -1,
		},
		{
			name:     "beta > alpha",
			v1:       SemVer{1, 2, 3, "beta", ""},
			v2:       SemVer{1, 2, 3, "alpha", ""},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.v1.Compare(tc.v2)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSemVer_IsCompatible(t *testing.T) {
	testCases := []struct {
		name       string
		current    SemVer
		required   SemVer
		compatible bool
	}{
		{
			name:       "same version",
			current:    SemVer{1, 2, 3, "", ""},
			required:   SemVer{1, 2, 3, "", ""},
			compatible: true,
		},
		{
			name:       "higher patch version",
			current:    SemVer{1, 2, 4, "", ""},
			required:   SemVer{1, 2, 3, "", ""},
			compatible: true,
		},
		{
			name:       "higher minor version",
			current:    SemVer{1, 3, 0, "", ""},
			required:   SemVer{1, 2, 3, "", ""},
			compatible: true,
		},
		{
			name:       "lower patch version",
			current:    SemVer{1, 2, 2, "", ""},
			required:   SemVer{1, 2, 3, "", ""},
			compatible: false,
		},
		{
			name:       "lower minor version",
			current:    SemVer{1, 1, 9, "", ""},
			required:   SemVer{1, 2, 0, "", ""},
			compatible: false,
		},
		{
			name:       "different major version",
			current:    SemVer{2, 0, 0, "", ""},
			required:   SemVer{1, 9, 9, "", ""},
			compatible: false,
		},
		{
			name:       "lower major version",
			current:    SemVer{1, 9, 9, "", ""},
			required:   SemVer{2, 0, 0, "", ""},
			compatible: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.current.IsCompatible(tc.required)
			assert.Equal(t, tc.compatible, result)
		})
	}
}

func TestParseSemVer(t *testing.T) {
	t.Run("basic version", func(t *testing.T) {
		v, err := ParseSemVer("1.2.3")
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
		assert.Empty(t, v.PreRelease)
		assert.Empty(t, v.Build)
	})

	t.Run("version with v prefix", func(t *testing.T) {
		v, err := ParseSemVer("v1.2.3")
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
	})

	t.Run("version with pre-release", func(t *testing.T) {
		v, err := ParseSemVer("1.2.3-alpha.1")
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
		assert.Equal(t, "alpha.1", v.PreRelease)
		assert.Empty(t, v.Build)
	})

	t.Run("version with build metadata", func(t *testing.T) {
		v, err := ParseSemVer("1.2.3+20240115.123")
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
		assert.Empty(t, v.PreRelease)
		assert.Equal(t, "20240115.123", v.Build)
	})

	t.Run("version with pre-release and build", func(t *testing.T) {
		v, err := ParseSemVer("v1.2.3-beta+exp.sha.5114f85")
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
		assert.Equal(t, "beta", v.PreRelease)
		assert.Equal(t, "exp.sha.5114f85", v.Build)
	})

	t.Run("invalid format - too few parts", func(t *testing.T) {
		_, err := ParseSemVer("1.2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid semantic version format")
	})

	t.Run("invalid format - too many parts", func(t *testing.T) {
		_, err := ParseSemVer("1.2.3.4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid semantic version format")
	})

	t.Run("invalid major version", func(t *testing.T) {
		_, err := ParseSemVer("a.2.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid major version")
	})

	t.Run("invalid minor version", func(t *testing.T) {
		_, err := ParseSemVer("1.b.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid minor version")
	})

	t.Run("invalid patch version", func(t *testing.T) {
		_, err := ParseSemVer("1.2.c")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid patch version")
	})
}

func TestGetCurrentSemVer(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("parses current version", func(t *testing.T) {
		Version = "v1.2.3-alpha"
		
		v, err := GetCurrentSemVer()
		require.NoError(t, err)
		assert.Equal(t, 1, v.Major)
		assert.Equal(t, 2, v.Minor)
		assert.Equal(t, 3, v.Patch)
		assert.Equal(t, "alpha", v.PreRelease)
	})

	t.Run("handles invalid current version", func(t *testing.T) {
		Version = "invalid-version"
		
		_, err := GetCurrentSemVer()
		assert.Error(t, err)
	})
}

func TestIsVersionCompatible(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("compatible versions", func(t *testing.T) {
		Version = "v1.2.3"
		
		compatible, err := IsVersionCompatible("v1.2.0")
		require.NoError(t, err)
		assert.True(t, compatible)
	})

	t.Run("incompatible versions", func(t *testing.T) {
		Version = "v1.1.0"
		
		compatible, err := IsVersionCompatible("v1.2.0")
		require.NoError(t, err)
		assert.False(t, compatible)
	})

	t.Run("invalid current version", func(t *testing.T) {
		Version = "invalid"
		
		_, err := IsVersionCompatible("v1.0.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse current version")
	})

	t.Run("invalid required version", func(t *testing.T) {
		Version = "v1.0.0"
		
		_, err := IsVersionCompatible("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse required version")
	})
}

func TestMustBeCompatible(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("compatible versions - no panic", func(t *testing.T) {
		Version = "v1.2.3"
		
		assert.NotPanics(t, func() {
			MustBeCompatible("v1.2.0")
		})
	})

	t.Run("incompatible versions - panics", func(t *testing.T) {
		Version = "v1.1.0"
		
		assert.Panics(t, func() {
			MustBeCompatible("v1.2.0")
		})
	})

	t.Run("invalid version - panics", func(t *testing.T) {
		Version = "invalid"
		
		assert.Panics(t, func() {
			MustBeCompatible("v1.0.0")
		})
	})
}

func TestVersionHeader(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("formats version header", func(t *testing.T) {
		Version = "v1.2.3"
		
		header := VersionHeader()
		assert.Contains(t, header, "TBP/1.2.3")
		assert.Contains(t, header, Platform)
		assert.Contains(t, header, GoVersion)
	})
}

func TestUserAgent(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("formats user agent", func(t *testing.T) {
		Version = "v1.2.3"
		
		ua := UserAgent("test-service")
		assert.Contains(t, ua, "test-service/1.2.3")
		assert.Contains(t, ua, "TBP-Foundation/1.2.3")
		assert.Contains(t, ua, Platform)
		assert.Contains(t, ua, GoVersion)
	})
}

func TestGetBuildInfo(t *testing.T) {
	t.Run("returns build info", func(t *testing.T) {
		info := GetBuildInfo()
		
		assert.NotNil(t, info)
		assert.NotEmpty(t, info.Version)
		assert.NotEmpty(t, info.GoVersion)
		assert.NotEmpty(t, info.Platform)
		assert.NotNil(t, info.Runtime)
		assert.NotNil(t, info.Flags)
		
		// Check runtime info
		assert.NotEmpty(t, info.Runtime.GOOS)
		assert.NotEmpty(t, info.Runtime.GOARCH)
		assert.Greater(t, info.Runtime.NumCPU, 0)
		assert.GreaterOrEqual(t, info.Runtime.NumGoroutine, 0)
		assert.NotEmpty(t, info.Runtime.Compiler)
	})
}

func TestSetBuildFlag(t *testing.T) {
	t.Run("sets build flag", func(t *testing.T) {
		// Set test flag
		SetBuildFlag("test_flag", "test_value")
		
		// Get fresh build info immediately
		info := GetBuildInfo()
		
		// Debug: Print all flags
		t.Logf("All flags: %+v", info.Flags)
		
		// Check if flag was set
		value, exists := info.Flags["test_flag"]
		assert.True(t, exists, "Flag should exist")
		assert.Equal(t, "test_value", value, "Flag should have correct value")
	})
}

func TestCheckMinimumVersion(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	t.Run("meets minimum version", func(t *testing.T) {
		Version = "v1.2.3"
		
		err := CheckMinimumVersion("v1.2.0")
		assert.NoError(t, err)
	})

	t.Run("equal to minimum version", func(t *testing.T) {
		Version = "v1.2.3"
		
		err := CheckMinimumVersion("v1.2.3")
		assert.NoError(t, err)
	})

	t.Run("below minimum version", func(t *testing.T) {
		Version = "v1.1.0"
		
		err := CheckMinimumVersion("v1.2.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not meet minimum requirement")
	})

	t.Run("invalid current version", func(t *testing.T) {
		Version = "invalid"
		
		err := CheckMinimumVersion("v1.0.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse current version")
	})

	t.Run("invalid minimum version", func(t *testing.T) {
		Version = "v1.0.0"
		
		err := CheckMinimumVersion("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse minimum version")
	})
}

func TestVersionInfoJSON(t *testing.T) {
	t.Run("marshals to JSON", func(t *testing.T) {
		info := &VersionInfo{
			Version:       "v1.2.3",
			GitCommit:     "abc123",
			BuildDate:     "2024-01-15",
			GoVersion:     "go1.21.0",
			Platform:      "linux/amd64",
			IsRelease:     true,
			IsDevelopment: false,
			ComponentName: "test-service",
		}
		
		data, err := json.Marshal(info)
		require.NoError(t, err)
		
		var unmarshaled VersionInfo
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		
		assert.Equal(t, info.Version, unmarshaled.Version)
		assert.Equal(t, info.GitCommit, unmarshaled.GitCommit)
		assert.Equal(t, info.ComponentName, unmarshaled.ComponentName)
		assert.Equal(t, info.IsRelease, unmarshaled.IsRelease)
	})
}

func TestVersionEdgeCases(t *testing.T) {
	t.Run("empty version values", func(t *testing.T) {
		// Save original values
		originalVersion := Version
		originalGitCommit := GitCommit
		originalBuildDate := BuildDate
		
		defer func() {
			Version = originalVersion
			GitCommit = originalGitCommit
			BuildDate = originalBuildDate
		}()
		
		Version = ""
		GitCommit = ""
		BuildDate = ""
		
		info := GetVersionInfo()
		assert.Empty(t, info.Version)
		assert.Empty(t, info.GitCommit)
		assert.Empty(t, info.BuildDate)
		
		str := info.String()
		// Should still have something (at least empty string is valid)
		assert.NotNil(t, str)
	})

	t.Run("short git commit", func(t *testing.T) {
		originalGitCommit := GitCommit
		defer func() {
			GitCommit = originalGitCommit
		}()
		
		GitCommit = "abc"
		assert.Equal(t, "abc", GetShortGitCommit())
	})
}

// Benchmark tests for performance validation
func BenchmarkGetVersionInfo(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetVersionInfo()
	}
}

func BenchmarkGetVersionInfoForComponent(b *testing.B) {
	componentName := "test-service"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetVersionInfoForComponent(componentName)
	}
}

func BenchmarkVersionInfo_String(b *testing.B) {
	info := &VersionInfo{
		ComponentName: "test-service",
		Version:       "v1.2.3",
		GitCommit:     "abc123def456",
		BuildDate:     "2024-01-15T10:30:00Z",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = info.String()
	}
}

func BenchmarkGetVersion(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetVersion()
	}
}

func BenchmarkGetShortVersion(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetShortVersion()
	}
}

func BenchmarkGetShortGitCommit(b *testing.B) {
	originalGitCommit := GitCommit
	GitCommit = "abc123def456ghi789"
	defer func() {
		GitCommit = originalGitCommit
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetShortGitCommit()
	}
}

func BenchmarkGetBuildTime(b *testing.B) {
	originalBuildDate := BuildDate
	BuildDate = "2024-01-15T10:30:00Z"
	defer func() {
		BuildDate = originalBuildDate
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GetBuildTime()
	}
}

func BenchmarkIsRelease(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsRelease()
	}
}

func BenchmarkIsDevelopment(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3-dev"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsDevelopment()
	}
}

func BenchmarkSemVer_String(b *testing.B) {
	v := SemVer{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "build.123"}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = v.String()
	}
}

func BenchmarkSemVer_String_Simple(b *testing.B) {
	v := SemVer{Major: 1, Minor: 2, Patch: 3}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = v.String()
	}
}

func BenchmarkSemVer_Compare(b *testing.B) {
	v1 := SemVer{1, 2, 3, "alpha", ""}
	v2 := SemVer{1, 2, 4, "beta", ""}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = v1.Compare(v2)
	}
}

func BenchmarkSemVer_Compare_Same(b *testing.B) {
	v1 := SemVer{1, 2, 3, "", ""}
	v2 := SemVer{1, 2, 3, "", ""}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = v1.Compare(v2)
	}
}

func BenchmarkSemVer_IsCompatible(b *testing.B) {
	current := SemVer{1, 3, 2, "", ""}
	required := SemVer{1, 2, 1, "", ""}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = current.IsCompatible(required)
	}
}

func BenchmarkParseSemVer(b *testing.B) {
	version := "v1.2.3-alpha.1+build.123"
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseSemVer(version)
	}
}

func BenchmarkParseSemVer_Simple(b *testing.B) {
	version := "1.2.3"
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseSemVer(version)
	}
}

func BenchmarkGetCurrentSemVer(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3-alpha"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GetCurrentSemVer()
	}
}

func BenchmarkIsVersionCompatible(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = IsVersionCompatible("v1.2.0")
	}
}

func BenchmarkVersionHeader(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = VersionHeader()
	}
}

func BenchmarkUserAgent(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = UserAgent("test-service")
	}
}

func BenchmarkGetBuildInfo(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetBuildInfo()
	}
}

func BenchmarkCheckMinimumVersion(b *testing.B) {
	originalVersion := Version
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
	}()
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CheckMinimumVersion("v1.2.0")
	}
}

func BenchmarkVersionInfo_JSON(b *testing.B) {
	info := &VersionInfo{
		Version:       "v1.2.3",
		GitCommit:     "abc123",
		BuildDate:     "2024-01-15",
		GoVersion:     "go1.21.0",
		Platform:      "linux/amd64",
		IsRelease:     true,
		IsDevelopment: false,
		ComponentName: "test-service",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(info)
	}
}

func BenchmarkVersionInfo_JSON_Unmarshal(b *testing.B) {
	data := []byte(`{
		"version": "v1.2.3",
		"git_commit": "abc123",
		"build_date": "2024-01-15",
		"go_version": "go1.21.0",
		"platform": "linux/amd64",
		"is_release": true,
		"is_development": false,
		"component_name": "test-service"
	}`)
	
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var info VersionInfo
		_ = json.Unmarshal(data, &info)
	}
}