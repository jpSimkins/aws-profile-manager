package core

import (
	"runtime"
	"strings"
	"testing"
)

// Test GetVersion function
func TestGetVersion(t *testing.T) {
	info := GetVersion()

	// Test required fields are set
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	if info.Framework == "" {
		t.Error("Framework should not be empty")
	}

	if info.Framework != Framework {
		t.Errorf("Expected Framework %s, got %s", Framework, info.Framework)
	}

	if info.FrameworkVersion == "" {
		t.Error("FrameworkVersion should not be empty")
	}

	// Test that GoVersion matches runtime
	if info.GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion %s, got %s", runtime.Version(), info.GoVersion)
	}

	// Test platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Expected Platform %s, got %s", expectedPlatform, info.Platform)
	}

	// Test that version defaults to AppVersion when not overridden
	if Version == AppVersion && info.Version != AppVersion {
		t.Errorf("Expected Version %s, got %s", AppVersion, info.Version)
	}
}

// Test GetVersionString with different build info scenarios
func TestGetVersionString(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
	}()

	tests := []struct {
		name        string
		version     string
		commit      string
		date        string
		expectParts []string
	}{
		{
			name:        "version only",
			version:     "1.0.0",
			commit:      "",
			date:        "",
			expectParts: []string{"1.0.0"},
		},
		{
			name:        "version with long commit",
			version:     "1.0.0",
			commit:      "abcdef1234567890",
			date:        "",
			expectParts: []string{"1.0.0", "(commit: abcdef1)"},
		},
		{
			name:        "version with short commit",
			version:     "1.0.0",
			commit:      "abc123",
			date:        "",
			expectParts: []string{"1.0.0", "(commit: abc123)"},
		},
		{
			name:        "version with date",
			version:     "1.0.0",
			commit:      "",
			date:        "2025-09-28",
			expectParts: []string{"1.0.0", "built on 2025-09-28"},
		},
		{
			name:        "version with commit and date",
			version:     "1.0.0",
			commit:      "abcdef1234567890",
			date:        "2025-09-28",
			expectParts: []string{"1.0.0", "(commit: abcdef1)", "built on 2025-09-28"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test values
			Version = tt.version
			Commit = tt.commit
			Date = tt.date

			versionStr := GetVersionString()

			// Check that all expected parts are present
			for _, part := range tt.expectParts {
				if !strings.Contains(versionStr, part) {
					t.Errorf("Expected version string to contain '%s', got '%s'", part, versionStr)
				}
			}

			// Check that version string starts with the version
			if !strings.HasPrefix(versionStr, tt.version) {
				t.Errorf("Expected version string to start with '%s', got '%s'", tt.version, versionStr)
			}
		})
	}
}

// Test GetSimpleVersion
func TestGetSimpleVersion(t *testing.T) {
	// Save original value
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()

	testVersion := "2.1.0"
	Version = testVersion

	result := GetSimpleVersion()
	if result != testVersion {
		t.Errorf("Expected GetSimpleVersion to return '%s', got '%s'", testVersion, result)
	}

	// Test with default version
	Version = AppVersion
	result = GetSimpleVersion()
	if result != AppVersion {
		t.Errorf("Expected GetSimpleVersion to return '%s', got '%s'", AppVersion, result)
	}
}

// Test GetFullVersionString
func TestGetFullVersionString(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
	}()

	// Set test values
	Version = "1.2.3"
	Commit = "abc1234567890def"
	Date = "2025-09-28T12:00:00Z"

	result := GetFullVersionString()

	// Test that all expected components are present
	expectedParts := []string{
		AppName,
		"Version:    " + Version,
		"Go version: " + runtime.Version(),
		"Platform:   " + runtime.GOOS + "/" + runtime.GOARCH,
		"Framework:  " + Framework,
		"Commit:     " + Commit,
		"Built:      " + FormatBuildDate(Date),
		"Author:     " + AppAuthor,
		"URL:        " + AppURL,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected full version string to contain '%s'\nGot: %s", part, result)
		}
	}
}

// Test GetFullVersionString without optional fields
func TestGetFullVersionStringMinimal(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
	}()

	// Set minimal values (no commit or date)
	Version = "1.0.0"
	Commit = ""
	Date = ""

	result := GetFullVersionString()

	// Test required components are present
	requiredParts := []string{
		AppName,
		"Version:    " + Version,
		"Go version: " + runtime.Version(),
		"Platform:   " + runtime.GOOS + "/" + runtime.GOARCH,
		"Framework:  " + Framework,
		"Author:     " + AppAuthor,
		"URL:        " + AppURL,
	}

	for _, part := range requiredParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected full version string to contain '%s'\nGot: %s", part, result)
		}
	}

	// Test optional components are not present
	if strings.Contains(result, "Commit:") {
		t.Error("Full version string should not contain 'Commit:' when commit is empty")
	}
	if strings.Contains(result, "Built:") {
		t.Error("Full version string should not contain 'Built:' when date is empty")
	}
}

// Test GetAppName
func TestGetAppName(t *testing.T) {
	result := GetAppName()
	if result != AppName {
		t.Errorf("Expected GetAppName to return '%s', got '%s'", AppName, result)
	}

	// Test that it returns the constant value
	if result != "AWS Profile Manager" {
		t.Errorf("Expected GetAppName to return 'AWS Profile Manager', got '%s'", result)
	}
}

// Test GetAppURL
func TestGetAppURL(t *testing.T) {
	result := GetAppURL()
	if result != AppURL {
		t.Errorf("Expected GetAppURL to return '%s', got '%s'", AppURL, result)
	}

	// Test that it returns the expected URL
	expectedURL := "https://github.com/jpsimkins/aws-profile-manager"
	if result != expectedURL {
		t.Errorf("Expected GetAppURL to return '%s', got '%s'", expectedURL, result)
	}
}

// Test GetFrameworkVersion
func TestGetFrameworkVersion(t *testing.T) {
	result := GetFrameworkVersion()
	if result == "" {
		t.Error("Expected GetFrameworkVersion to return a non-empty value")
	}
}

// Test constants have expected values
func TestConstants(t *testing.T) {
	if AppName == "" {
		t.Error("AppName constant should not be empty")
	}

	if AppVersion == "" {
		t.Error("AppVersion constant should not be empty")
	}

	if AppAuthor == "" {
		t.Error("AppAuthor constant should not be empty")
	}

	if AppURL == "" {
		t.Error("AppURL constant should not be empty")
	}

	// Test expected constant values
	expectedValues := map[string]string{
		"AppName":    "AWS Profile Manager",
		"AppVersion": "0.0.1",
		"AppAuthor":  "jpSimkins",
		"AppURL":     "https://github.com/jpsimkins/aws-profile-manager",
		"Framework":  "Fyne",
	}

	actualValues := map[string]string{
		"AppName":    AppName,
		"AppVersion": AppVersion,
		"AppAuthor":  AppAuthor,
		"AppURL":     AppURL,
		"Framework":  Framework,
	}

	for key, expected := range expectedValues {
		if actual := actualValues[key]; actual != expected {
			t.Errorf("Expected %s to be '%s', got '%s'", key, expected, actual)
		}
	}
}

// Test build variables initialization
func TestBuildVariables(t *testing.T) {
	// Test that Version defaults to AppVersion
	if Version != AppVersion {
		// This is okay if Version was overridden at build time
		t.Logf("Version is '%s', expected default '%s' (may be overridden at build time)", Version, AppVersion)
	}

	// Test that GoVersion is set
	if GoVersion == "" {
		t.Error("GoVersion should be set")
	}

	if GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion to be '%s', got '%s'", runtime.Version(), GoVersion)
	}

	// Commit and Date are optional and can be empty (set via ldflags)
	t.Logf("Commit: '%s' (may be empty if not set via ldflags)", Commit)
	t.Logf("Date: '%s' (may be empty if not set via ldflags)", Date)
}

// Test Info struct JSON tags
func TestInfoStructure(t *testing.T) {
	info := GetVersion()

	// Test that Info struct has all expected fields
	if info.Version == "" {
		t.Error("Info.Version should not be empty")
	}

	// Commit and Date can be empty, but fields should exist
	_ = info.Commit
	_ = info.Date

	if info.GoVersion == "" {
		t.Error("Info.GoVersion should not be empty")
	}

	if info.Platform == "" {
		t.Error("Info.Platform should not be empty")
	}

	if info.Framework == "" {
		t.Error("Info.Framework should not be empty")
	}

	if info.FrameworkVersion == "" {
		t.Error("Info.FrameworkVersion should not be empty")
	}
}

// Test version string edge cases
func TestVersionStringEdgeCases(t *testing.T) {
	// Save original values
	originalVersion := Version
	originalCommit := Commit
	originalDate := Date
	defer func() {
		Version = originalVersion
		Commit = originalCommit
		Date = originalDate
	}()

	// Test with empty version
	Version = ""
	Commit = ""
	Date = ""

	versionStr := GetVersionString()
	if versionStr != "" {
		t.Errorf("Expected empty version string when version is empty, got '%s'", versionStr)
	}

	// Test with exactly 7 character commit
	Version = "1.0.0"
	Commit = "abcdef1"
	Date = ""

	versionStr = GetVersionString()
	expected := "1.0.0 (commit: abcdef1)"
	if versionStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, versionStr)
	}

	// Test with 8 character commit (should be truncated)
	Commit = "abcdef12"
	versionStr = GetVersionString()
	expected = "1.0.0 (commit: abcdef1)"
	if versionStr != expected {
		t.Errorf("Expected '%s', got '%s'", expected, versionStr)
	}
}

// Benchmark version functions
func BenchmarkGetVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetVersion()
	}
}

func BenchmarkGetVersionString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetVersionString()
	}
}

func BenchmarkGetFullVersionString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetFullVersionString()
	}
}
