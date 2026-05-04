// Package test provides test fixtures for settings.
//
// This package is for use by OTHER packages that need settings test data
// (e.g., backup tests, CLI tests, GUI tests). It CANNOT be used by the
// settings package itself due to import cycles.
//
// The settings package has its own internal test fixtures in *_test.go files.
//
// Usage:
//
//	import settingstest "aws-profile-manager/internal/settings/test"
//
//	func TestMyFeature(t *testing.T) {
//	    settings := settingstest.NewWithDarkTheme()
//	    // ... test with settings
//	}
package test

import (
	"aws-profile-manager/internal/settings"
)

// NewDefault returns default settings configuration.
//
// This is the baseline settings that the application starts with.
// Use this when you just need valid settings without customization.
//
// Returns:
//   - *settings.Settings: Default settings
func NewDefault() *settings.Settings {
	return settings.GetDefaults()
}

// NewWithDarkTheme returns settings with dark theme configured.
//
// Useful for testing theme-related functionality.
//
// Returns:
//   - *settings.Settings: Settings with dark theme
func NewWithDarkTheme() *settings.Settings {
	s := settings.GetDefaults()
	s.GUI.Theme = "dark"
	return s
}

// NewWithLightTheme returns settings with light theme configured.
//
// Useful for testing theme-related functionality.
//
// Returns:
//   - *settings.Settings: Settings with light theme
func NewWithLightTheme() *settings.Settings {
	s := settings.GetDefaults()
	s.GUI.Theme = "light"
	return s
}

// NewWithHttpSync returns settings with HTTP sync configured.
//
// Useful for testing sync functionality with HTTP strategy.
//
// Returns:
//   - *settings.Settings: Settings with HTTP sync
func NewWithHttpSync() *settings.Settings {
	s := settings.GetDefaults()
	s.Sync.Strategy = "http"
	s.Sync.HTTP.URL = "https://example.com/config.json"
	return s
}

// NewWithS3Sync returns settings with S3 sync configured.
//
// Useful for testing sync functionality with S3 strategy.
//
// Returns:
//   - *settings.Settings: Settings with S3 sync
func NewWithS3Sync() *settings.Settings {
	s := settings.GetDefaults()
	s.Sync.Strategy = "s3"
	s.Sync.S3.Bucket = "my-bucket"
	s.Sync.S3.Key = "config.json"
	s.Sync.S3.Region = "us-east-1"
	return s
}

// NewWithLocalSync returns settings with local sync configured.
//
// Useful for testing sync functionality with local strategy.
//
// Returns:
//   - *settings.Settings: Settings with local sync
func NewWithLocalSync() *settings.Settings {
	s := settings.GetDefaults()
	s.Sync.Strategy = "local"
	s.Sync.Local.Path = "/tmp/config.json"
	return s
}

// NewWithCustomMarkers returns settings with custom managed section markers.
//
// Useful for testing marker-related functionality.
//
// Returns:
//   - *settings.Settings: Settings with custom markers
func NewWithCustomMarkers() *settings.Settings {
	s := settings.GetDefaults()
	s.Application.ManagedSectionStart = "CUSTOM START"
	s.Application.ManagedSectionEnd = "CUSTOM END"
	return s
}

// NewWithDebugLogging returns settings with debug logging enabled.
//
// Useful for testing logging functionality.
//
// Returns:
//   - *settings.Settings: Settings with debug logging
func NewWithDebugLogging() *settings.Settings {
	s := settings.GetDefaults()
	s.Logging.LogLevel = "debug"
	s.Logging.EnableDebug = true
	return s
}

// NewWithAutoRefresh returns settings with auto-refresh enabled.
//
// Useful for testing auto-refresh functionality.
//
// Returns:
//   - *settings.Settings: Settings with auto-refresh
func NewWithAutoRefresh() *settings.Settings {
	s := settings.GetDefaults()
	s.AwsCLI.AutoRefresh = true
	s.AwsCLI.RefreshIntervalMins = 5 // 5 minutes
	return s
}

// NewWithAllCustomized returns settings with many fields customized.
//
// Useful for testing complex scenarios with non-default values.
//
// Returns:
//   - *settings.Settings: Settings with many customizations
func NewWithAllCustomized() *settings.Settings {
	s := settings.GetDefaults()
	s.GUI.Theme = "dark"
	s.GUI.WindowWidth = 1920
	s.GUI.WindowHeight = 1080
	s.Sync.Strategy = "http"
	s.Sync.HTTP.URL = "https://example.com/config.json"
	s.AwsCLI.AutoRefresh = true
	s.AwsCLI.RefreshIntervalMins = 10 // 10 minutes
	s.Logging.LogLevel = "debug"
	s.Logging.EnableDebug = true
	return s
}

// NewInvalid returns settings with invalid values.
//
// Useful for testing validation and error handling.
//
// Returns:
//   - *settings.Settings: Invalid settings (will fail validation)
func NewInvalid() *settings.Settings {
	s := settings.GetDefaults()
	s.Sync.Strategy = "invalid-strategy" // Invalid
	s.Sync.HTTP.URL = "not-a-url"        // Invalid URL
	return s
}

// NewWithMissingRequired returns settings with required fields missing.
//
// Useful for testing validation and error handling.
// Note: This may not always fail validation depending on strategy requirements.
//
// Returns:
//   - *settings.Settings: Settings potentially missing required fields
func NewWithMissingRequired() *settings.Settings {
	s := settings.GetDefaults()
	// Set strategy to none which might not be valid
	s.Sync.Strategy = ""
	return s
}

// NewMinimal returns minimal valid settings.
//
// Useful for testing with the absolute minimum configuration.
//
// Returns:
//   - *settings.Settings: Minimal valid settings
func NewMinimal() *settings.Settings {
	return settings.GetDefaults()
}
