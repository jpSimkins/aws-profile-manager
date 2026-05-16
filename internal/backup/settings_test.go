package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

// TestBackupSettings tests backing up current settings.
func TestBackupSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Initialize app to get valid settings
	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)

	// Backup settings
	backupPath, err := BackupSettings(cfg)
	if err != nil {
		t.Fatalf("Failed to backup settings: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Settings backup file was not created")
	}

	// Verify filename is singular backup
	filename := filepath.Base(backupPath)
	expectedFilename := "settings-backup.json"
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	// Read and validate backup content
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	var restoredSettings settings.Settings
	if err := json.Unmarshal(data, &restoredSettings); err != nil {
		t.Fatalf("Failed to parse backup: %v", err)
	}

	// Verify settings match
	currentSettings := settings.Get()
	if restoredSettings.Version != currentSettings.Version {
		t.Errorf("Version mismatch: got %s, want %s",
			restoredSettings.Version, currentSettings.Version)
	}
}

// TestBackupSettings_MultipleBackups tests that subsequent backups overwrite previous.
func TestBackupSettings_MultipleBackups(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)

	// Create first backup
	backup1, err := BackupSettings(cfg)
	if err != nil {
		t.Fatalf("Failed to create first backup: %v", err)
	}

	// Get first backup content
	content1, err := os.ReadFile(backup1)
	if err != nil {
		t.Fatalf("Failed to read first backup: %v", err)
	}

	// Modify settings
	currentSettings := settings.Get()
	currentSettings.Logging.LogLevel = "debug"
	_ = settings.Set(currentSettings)

	// Create second backup (should overwrite first)
	backup2, err := BackupSettings(cfg)
	if err != nil {
		t.Fatalf("Failed to create second backup: %v", err)
	}

	// Paths should be identical (singular backup file)
	if backup1 != backup2 {
		t.Errorf("Expected same backup path, got %s and %s", backup1, backup2)
	}

	// Content should be different (settings changed)
	content2, err := os.ReadFile(backup2)
	if err != nil {
		t.Fatalf("Failed to read second backup: %v", err)
	}

	if string(content1) == string(content2) {
		t.Error("Second backup should have different content (settings changed)")
	}
}

// TestRestoreSettings tests restoring settings.
func TestRestoreSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Get current settings and modify them
	originalSettings := settings.Get()
	modifiedSettings := *originalSettings
	modifiedSettings.GUI.Theme = "dark"

	// Restore modified settings
	if err := RestoreSettings(&modifiedSettings); err != nil {
		t.Fatalf("Failed to restore settings: %v", err)
	}

	// Verify settings were applied
	currentSettings := settings.Get()
	if currentSettings.GUI.Theme != "dark" {
		t.Errorf("Settings were not restored: got theme %s, want dark",
			currentSettings.GUI.Theme)
	}
}

// TestRestoreSettings_NilSettings tests restoring nil settings.
func TestRestoreSettings_NilSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	err := RestoreSettings(nil)
	if err == nil {
		t.Fatal("Expected error for nil settings")
	}
}

// TestRestoreSettings_InvalidSettings tests restoring invalid settings.
func TestRestoreSettings_InvalidSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create settings with invalid values
	invalidSettings := settings.GetDefaults()
	invalidSettings.Sync.Strategy = "invalid_strategy" // Invalid enum value

	// Attempt to restore (should fail validation)
	err := RestoreSettings(invalidSettings)
	if err == nil {
		t.Fatal("Expected error for invalid settings")
	}
}

// TestBackupAndRestore_RoundTrip tests full backup/restore cycle.
func TestBackupAndRestore_RoundTrip(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)

	// Modify settings
	originalSettings := settings.Get()
	modifiedSettings := *originalSettings
	modifiedSettings.GUI.Theme = "dark"
	modifiedSettings.Logging.LogLevel = "debug"
	if err := settings.Set(&modifiedSettings); err != nil {
		t.Fatalf("Failed to set modified settings: %v", err)
	}

	// Backup current settings
	backupPath, err := BackupSettings(cfg)
	if err != nil {
		t.Fatalf("Failed to backup settings: %v", err)
	}

	// Change settings again
	changedSettings := settings.Get()
	changedSettings.GUI.Theme = "light"
	changedSettings.Logging.LogLevel = "info"
	if err := settings.Set(changedSettings); err != nil {
		t.Fatalf("Failed to change settings: %v", err)
	}

	// Read backup and restore
	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	var restoredSettings settings.Settings
	if err := json.Unmarshal(data, &restoredSettings); err != nil {
		t.Fatalf("Failed to parse backup: %v", err)
	}

	if err := RestoreSettings(&restoredSettings); err != nil {
		t.Fatalf("Failed to restore settings: %v", err)
	}

	// Verify restored settings match original modified settings
	finalSettings := settings.Get()
	if finalSettings.GUI.Theme != "dark" {
		t.Errorf("Theme not restored: got %s, want dark", finalSettings.GUI.Theme)
	}
	if finalSettings.Logging.LogLevel != "debug" {
		t.Errorf("Log level not restored: got %s, want debug", finalSettings.Logging.LogLevel)
	}
}
