package test

import (
	"testing"
)

// TestNewDefault verifies default settings creation.
func TestNewDefault(t *testing.T) {
	s := NewDefault()
	if s == nil {
		t.Fatal("NewDefault() returned nil")
	}
	if s.Version == "" {
		t.Error("Version should be set")
	}
}

// TestNewWithDarkTheme verifies dark theme settings.
func TestNewWithDarkTheme(t *testing.T) {
	s := NewWithDarkTheme()
	if s.GUI.Theme != "dark" {
		t.Errorf("Expected dark theme, got %s", s.GUI.Theme)
	}
}

// TestNewWithLightTheme verifies light theme settings.
func TestNewWithLightTheme(t *testing.T) {
	s := NewWithLightTheme()
	if s.GUI.Theme != "light" {
		t.Errorf("Expected light theme, got %s", s.GUI.Theme)
	}
}

// TestNewWithHttpSync verifies HTTP sync settings.
func TestNewWithHttpSync(t *testing.T) {
	s := NewWithHttpSync()
	if s.Sync.Strategy != "http" {
		t.Errorf("Expected http strategy, got %s", s.Sync.Strategy)
	}
	if s.Sync.HTTP.URL == "" {
		t.Error("HTTP URL should be set")
	}
}

// TestNewWithS3Sync verifies S3 sync settings.
func TestNewWithS3Sync(t *testing.T) {
	s := NewWithS3Sync()
	if s.Sync.Strategy != "s3" {
		t.Errorf("Expected s3 strategy, got %s", s.Sync.Strategy)
	}
	if s.Sync.S3.Bucket == "" {
		t.Error("S3 bucket should be set")
	}
}

// TestNewWithLocalSync verifies local sync settings.
func TestNewWithLocalSync(t *testing.T) {
	s := NewWithLocalSync()
	if s.Sync.Strategy != "local" {
		t.Errorf("Expected local strategy, got %s", s.Sync.Strategy)
	}
	if s.Sync.Local.Path == "" {
		t.Error("Local path should be set")
	}
}

// TestNewWithCustomMarkers verifies custom marker settings.
func TestNewWithCustomMarkers(t *testing.T) {
	s := NewWithCustomMarkers()
	if s.Application.ManagedSectionStart != "CUSTOM START" {
		t.Errorf("Expected custom start marker, got %s", s.Application.ManagedSectionStart)
	}
	if s.Application.ManagedSectionEnd != "CUSTOM END" {
		t.Errorf("Expected custom end marker, got %s", s.Application.ManagedSectionEnd)
	}
}

// TestNewWithDebugLogging verifies debug logging settings.
func TestNewWithDebugLogging(t *testing.T) {
	s := NewWithDebugLogging()
	if s.Logging.LogLevel != "debug" {
		t.Errorf("Expected debug log level, got %s", s.Logging.LogLevel)
	}
	if !s.Logging.EnableDebug {
		t.Error("Debug should be enabled")
	}
}

// TestNewWithAutoRefresh verifies auto-refresh settings.
func TestNewWithAutoRefresh(t *testing.T) {
	s := NewWithAutoRefresh()
	if !s.AwsCLI.AutoRefresh {
		t.Error("Auto-refresh should be enabled")
	}
	if s.AwsCLI.RefreshIntervalMins == 0 {
		t.Error("Refresh interval should be set")
	}
}

// TestNewWithAllCustomized verifies fully customized settings.
func TestNewWithAllCustomized(t *testing.T) {
	s := NewWithAllCustomized()
	if s.GUI.Theme != "dark" {
		t.Error("Theme should be dark")
	}
	if s.Sync.Strategy != "http" {
		t.Error("Strategy should be http")
	}
	if s.AwsCLI.AutoRefresh != true {
		t.Error("Auto-refresh should be enabled")
	}
	if s.Logging.LogLevel != "debug" {
		t.Error("Log level should be debug")
	}
}

// TestNewInvalid verifies invalid settings creation.
func TestNewInvalid(t *testing.T) {
	s := NewInvalid()
	if s == nil {
		t.Fatal("NewInvalid() returned nil")
	}
	// Invalid settings should be created but fail validation
	if err := s.Validate(); err == nil {
		t.Error("Invalid settings should fail validation")
	}
}

// TestNewWithMissingRequired verifies settings with missing required fields.
func TestNewWithMissingRequired(t *testing.T) {
	s := NewWithMissingRequired()
	if s == nil {
		t.Fatal("NewWithMissingRequired() returned nil")
	}
	// Note: May or may not fail validation depending on default strategy
	// This is just a fixture for testing missing field scenarios
}

// TestNewMinimal verifies minimal settings creation.
func TestNewMinimal(t *testing.T) {
	s := NewMinimal()
	if s == nil {
		t.Fatal("NewMinimal() returned nil")
	}
	// Minimal settings should be valid
	if err := s.Validate(); err != nil {
		t.Errorf("Minimal settings should be valid: %v", err)
	}
}
