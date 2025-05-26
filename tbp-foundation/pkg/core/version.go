// File: version.go
// Title: Version Information Management for TBP Core
// Description: Provides version information, build metadata, and semantic
//              versioning support for TBP components. Includes build-time
//              injection of version data and runtime version comparison
//              functionality for compatibility checks.
// Author: msto63 with Claude Sonnet 4.0
// Version: v0.1.0
// Created: 2025-05-26
// Modified: 2025-05-26
//
// Change History:
// - 2025-05-26 v0.1.0: Initial implementation with semantic versioning support

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Version information variables that are set at build time.
// These can be overridden using ldflags during build:
// go build -ldflags "-X github.com/msto63/tbp/tbp-foundation/pkg/core.Version=v1.2.3"
var (
	// Version is the semantic version of the TBP foundation
	Version = "v0.1.0-dev"
	
	// GitCommit is the git commit hash this binary was built from
	GitCommit = "unknown"
	
	// BuildDate is the date and time when this binary was built
	BuildDate = "unknown"
	
	// BuildUser is the user who built this binary
	BuildUser = "unknown"
	
	// BuildHost is the host where this binary was built
	BuildHost = "unknown"
	
	// GoVersion is the Go version used to build this binary
	GoVersion = runtime.Version()
	
	// Platform is the OS/Arch this binary was built for
	Platform = runtime.GOOS + "/" + runtime.GOARCH
)

// buildFlags stores custom build flags
var buildFlags = make(map[string]string)

// VersionInfo contains comprehensive version and build information.
type VersionInfo struct {
	// Version is the semantic version
	Version string `json:"version"`
	
	// GitCommit is the git commit hash
	GitCommit string `json:"git_commit"`
	
	// GitBranch is the git branch (if available)
	GitBranch string `json:"git_branch,omitempty"`
	
	// BuildDate is when the binary was built
	BuildDate string `json:"build_date"`
	
	// BuildUser is who built the binary
	BuildUser string `json:"build_user"`
	
	// BuildHost is where the binary was built
	BuildHost string `json:"build_host"`
	
	// GoVersion is the Go compiler version
	GoVersion string `json:"go_version"`
	
	// Platform is the target platform (OS/Arch)
	Platform string `json:"platform"`
	
	// IsRelease indicates if this is a release build
	IsRelease bool `json:"is_release"`
	
	// IsDevelopment indicates if this is a development build
	IsDevelopment bool `json:"is_development"`
	
	// ComponentName is the name of the component (set by each service)
	ComponentName string `json:"component_name,omitempty"`
	
	// Dependencies contains version info of key dependencies
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// GetVersionInfo returns comprehensive version information.
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:       Version,
		GitCommit:     GitCommit,
		BuildDate:     BuildDate,
		BuildUser:     BuildUser,
		BuildHost:     BuildHost,
		GoVersion:     GoVersion,
		Platform:      Platform,
		IsRelease:     IsRelease(),
		IsDevelopment: IsDevelopment(),
		Dependencies:  make(map[string]string),
	}
}

// GetVersionInfoForComponent returns version information for a specific component.
func GetVersionInfoForComponent(componentName string) *VersionInfo {
	info := GetVersionInfo()
	info.ComponentName = componentName
	return info
}

// String returns a human-readable version string.
func (vi *VersionInfo) String() string {
	var parts []string
	
	if vi.ComponentName != "" {
		parts = append(parts, vi.ComponentName)
	}
	
	parts = append(parts, vi.Version)
	
	if vi.GitCommit != "unknown" && vi.GitCommit != "" {
		commit := vi.GitCommit
		if len(commit) > 7 {
			commit = commit[:7] // Short commit hash
		}
		parts = append(parts, fmt.Sprintf("commit:%s", commit))
	}
	
	if vi.BuildDate != "unknown" && vi.BuildDate != "" {
		parts = append(parts, fmt.Sprintf("built:%s", vi.BuildDate))
	}
	
	return strings.Join(parts, " ")
}

// GetVersion returns the current version string.
func GetVersion() string {
	return Version
}

// GetShortVersion returns the version without the 'v' prefix.
func GetShortVersion() string {
	return strings.TrimPrefix(Version, "v")
}

// GetGitCommit returns the git commit hash.
func GetGitCommit() string {
	return GitCommit
}

// GetShortGitCommit returns the short git commit hash (7 characters).
func GetShortGitCommit() string {
	if len(GitCommit) > 7 {
		return GitCommit[:7]
	}
	return GitCommit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return BuildDate
}

// GetBuildTime returns the build date as time.Time if parseable.
func GetBuildTime() (time.Time, error) {
	if BuildDate == "unknown" || BuildDate == "" {
		return time.Time{}, fmt.Errorf("build date unknown")
	}
	
	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, BuildDate); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse build date: %s", BuildDate)
}

// IsRelease checks if this is a release version (no dev/alpha/beta/rc suffix).
func IsRelease() bool {
	v := strings.ToLower(GetShortVersion())
	return !strings.Contains(v, "dev") &&
		!strings.Contains(v, "alpha") &&
		!strings.Contains(v, "beta") &&
		!strings.Contains(v, "rc")
}

// IsDevelopment checks if this is a development version.
func IsDevelopment() bool {
	return !IsRelease()
}

// SemVer represents a semantic version with major, minor, and patch components.
type SemVer struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	PreRelease string `json:"pre_release,omitempty"`
	Build      string `json:"build,omitempty"`
}

// String returns the semantic version as a string.
func (sv SemVer) String() string {
	version := fmt.Sprintf("%d.%d.%d", sv.Major, sv.Minor, sv.Patch)
	
	if sv.PreRelease != "" {
		version += "-" + sv.PreRelease
	}
	
	if sv.Build != "" {
		version += "+" + sv.Build
	}
	
	return version
}

// Compare compares this version with another version.
// Returns -1 if this version is less, 0 if equal, 1 if greater.
func (sv SemVer) Compare(other SemVer) int {
	// Compare major version
	if sv.Major < other.Major {
		return -1
	}
	if sv.Major > other.Major {
		return 1
	}
	
	// Compare minor version
	if sv.Minor < other.Minor {
		return -1
	}
	if sv.Minor > other.Minor {
		return 1
	}
	
	// Compare patch version
	if sv.Patch < other.Patch {
		return -1
	}
	if sv.Patch > other.Patch {
		return 1
	}
	
	// Compare pre-release versions
	// Version without pre-release is greater than with pre-release
	if sv.PreRelease == "" && other.PreRelease != "" {
		return 1
	}
	if sv.PreRelease != "" && other.PreRelease == "" {
		return -1
	}
	
	// Both have pre-release, compare lexically
	if sv.PreRelease < other.PreRelease {
		return -1
	}
	if sv.PreRelease > other.PreRelease {
		return 1
	}
	
	return 0
}

// IsCompatible checks if this version is compatible with another version.
// Uses semantic versioning compatibility rules.
func (sv SemVer) IsCompatible(other SemVer) bool {
	// Major version must match for compatibility
	if sv.Major != other.Major {
		return false
	}
	
	// Minor version of this should be >= other
	if sv.Minor < other.Minor {
		return false
	}
	
	// If minor versions match, patch version of this should be >= other
	if sv.Minor == other.Minor && sv.Patch < other.Patch {
		return false
	}
	
	return true
}

// ParseSemVer parses a semantic version string.
func ParseSemVer(version string) (*SemVer, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	
	// Split on '+' to separate build metadata
	var buildMeta string
	if idx := strings.Index(version, "+"); idx >= 0 {
		buildMeta = version[idx+1:]
		version = version[:idx]
	}
	
	// Split on '-' to separate pre-release
	var preRelease string
	if idx := strings.Index(version, "-"); idx >= 0 {
		preRelease = version[idx+1:]
		version = version[:idx]
	}
	
	// Split version into major.minor.patch
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid semantic version format: %s", version)
	}
	
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}
	
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", parts[2])
	}
	
	return &SemVer{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		Build:      buildMeta,
	}, nil
}

// GetCurrentSemVer returns the current version as a SemVer struct.
func GetCurrentSemVer() (*SemVer, error) {
	return ParseSemVer(Version)
}

// IsVersionCompatible checks if the current version is compatible with a required version.
func IsVersionCompatible(requiredVersion string) (bool, error) {
	current, err := GetCurrentSemVer()
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %w", err)
	}
	
	required, err := ParseSemVer(requiredVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse required version: %w", err)
	}
	
	return current.IsCompatible(*required), nil
}

// MustBeCompatible panics if the current version is not compatible with required version.
func MustBeCompatible(requiredVersion string) {
	compatible, err := IsVersionCompatible(requiredVersion)
	if err != nil {
		panic(fmt.Sprintf("version compatibility check failed: %v", err))
	}
	
	if !compatible {
		panic(fmt.Sprintf("version incompatibility: current %s is not compatible with required %s", 
			Version, requiredVersion))
	}
}

// VersionHeader returns version information as HTTP header value.
func VersionHeader() string {
	return fmt.Sprintf("TBP/%s (%s; %s)", GetShortVersion(), Platform, GoVersion)
}

// UserAgent returns a user agent string for HTTP clients.
func UserAgent(componentName string) string {
	return fmt.Sprintf("%s/%s TBP-Foundation/%s (%s; %s)", 
		componentName, GetShortVersion(), GetShortVersion(), Platform, GoVersion)
}

// BuildInfo returns build information for debugging and support.
type BuildInfo struct {
	Version   string            `json:"version"`
	GitCommit string            `json:"git_commit"`
	BuildDate string            `json:"build_date"`
	GoVersion string            `json:"go_version"`
	Platform  string            `json:"platform"`
	Runtime   RuntimeInfo       `json:"runtime"`
	Flags     map[string]string `json:"build_flags,omitempty"`
}

// RuntimeInfo contains runtime information.
type RuntimeInfo struct {
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	Compiler     string `json:"compiler"`
}

// SetBuildFlag sets a build flag for inclusion in build info.
// This can be used to track custom build flags or configuration.
func SetBuildFlag(key, value string) {
	buildFlags[key] = value
}

// GetBuildInfo returns comprehensive build and runtime information.
func GetBuildInfo() *BuildInfo {
	// Copy build flags
	flags := make(map[string]string)
	for k, v := range buildFlags {
		flags[k] = v
	}
	
	return &BuildInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
		Runtime: RuntimeInfo{
			GOOS:         runtime.GOOS,
			GOARCH:       runtime.GOARCH,
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
			Compiler:     runtime.Compiler,
		},
		Flags: flags,
	}
}

// CheckMinimumVersion checks if the current version meets minimum requirements.
func CheckMinimumVersion(minimumVersion string) error {
	current, err := GetCurrentSemVer()
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}
	
	minimum, err := ParseSemVer(minimumVersion)
	if err != nil {
		return fmt.Errorf("failed to parse minimum version: %w", err)
	}
	
	if current.Compare(*minimum) < 0 {
		return fmt.Errorf("version %s does not meet minimum requirement %s", 
			current.String(), minimum.String())
	}
	
	return nil
}

// PrintVersion prints version information to stdout in a formatted way.
func PrintVersion(componentName string) {
	info := GetVersionInfoForComponent(componentName)
	fmt.Printf("Component: %s\n", info.ComponentName)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Git Commit: %s\n", info.GitCommit)
	fmt.Printf("Build Date: %s\n", info.BuildDate)
	fmt.Printf("Build User: %s\n", info.BuildUser)
	fmt.Printf("Go Version: %s\n", info.GoVersion)
	fmt.Printf("Platform: %s\n", info.Platform)
	fmt.Printf("Release Build: %t\n", info.IsRelease)
}

// PrintVersionJSON prints version information as JSON.
func PrintVersionJSON(componentName string) error {
	info := GetVersionInfoForComponent(componentName)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(info)
}