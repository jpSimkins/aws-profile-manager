package profiles

import (
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestNewTestConfig tests that newTestConfig creates a valid Config for testing.
func TestNewTestConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Verify all fields are set
	if config.ConfigPath == "" {
		t.Error("ConfigPath should be set")
	}

	if config.StartMarker == "" {
		t.Error("StartMarker should be set")
	}

	if config.EndMarker == "" {
		t.Error("EndMarker should be set")
	}

	if config.CheatSheetOutputDir == "" {
		t.Error("CheatSheetOutputDir should be set")
	}

	// Verify paths point to test directories
	expectedConfigPath := test.GetTestAwsConfigPath(t)
	if config.ConfigPath != expectedConfigPath {
		t.Errorf("ConfigPath = %s, want %s", config.ConfigPath, expectedConfigPath)
	}

	expectedDesktopDir := test.GetTestDesktopDir(t)
	if config.CheatSheetOutputDir != expectedDesktopDir {
		t.Errorf("CheatSheetOutputDir = %s, want %s", config.CheatSheetOutputDir, expectedDesktopDir)
	}

	// Verify ConfigPath ends with 'config' file
	if filepath.Base(config.ConfigPath) != "config" {
		t.Errorf("ConfigPath should end with 'config', got %s", filepath.Base(config.ConfigPath))
	}

	// Verify markers are correct
	if config.StartMarker != "# START" {
		t.Errorf("StartMarker = %s, want '# START'", config.StartMarker)
	}

	if config.EndMarker != "# END" {
		t.Errorf("EndMarker = %s, want '# END'", config.EndMarker)
	}
}

// TestNewTestConfig_MultipleCalls tests that multiple calls create consistent configs.
func TestNewTestConfig_MultipleCalls(t *testing.T) {
	test.SetupTestEnvironment(t)

	config1 := newTestConfig(t)
	config2 := newTestConfig(t)

	// Both configs should have the same paths (same test environment)
	if config1.ConfigPath != config2.ConfigPath {
		t.Error("Multiple calls should produce same ConfigPath")
	}

	if config1.CheatSheetOutputDir != config2.CheatSheetOutputDir {
		t.Error("Multiple calls should produce same CheatSheetOutputDir")
	}

	// Markers should be identical
	if config1.StartMarker != config2.StartMarker {
		t.Error("Multiple calls should produce same StartMarker")
	}

	if config1.EndMarker != config2.EndMarker {
		t.Error("Multiple calls should produce same EndMarker")
	}
}
