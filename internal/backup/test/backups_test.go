package test

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestNewManagedOnlyBackup verifies managed-only backup creation.
func TestNewManagedOnlyBackup(t *testing.T) {
	backup := NewManagedOnlyBackup()
	if backup == nil {
		t.Fatal("NewManagedOnlyBackup() returned nil")
	}
	if backup.Data == nil {
		t.Error("Backup should have data")
	}
	if backup.Settings != nil {
		t.Error("Backup should not have settings")
	}
}

// TestNewFullBackup verifies full backup creation.
func TestNewFullBackup(t *testing.T) {
	backup := NewFullBackup()
	if backup == nil {
		t.Fatal("NewFullBackup() returned nil")
	}
	if backup.Data == nil {
		t.Error("Backup should have data")
	}
	if backup.Settings == nil {
		t.Error("Backup should have settings")
	}
}

// TestNewMixedBackup verifies mixed backup creation.
func TestNewMixedBackup(t *testing.T) {
	backup := NewMixedBackup()
	if backup == nil {
		t.Fatal("NewMixedBackup() returned nil")
	}
	if backup.Data == nil {
		t.Error("Backup should have data")
	}
}

// TestNewSettingsOnlyBackup verifies settings-only backup creation.
func TestNewSettingsOnlyBackup(t *testing.T) {
	backup := NewSettingsOnlyBackup()
	if backup == nil {
		t.Fatal("NewSettingsOnlyBackup() returned nil")
	}
	if backup.Data != nil {
		t.Error("Backup should not have data")
	}
	if backup.Settings == nil {
		t.Error("Backup should have settings")
	}
}

// TestNewEmptyBackup verifies empty backup creation.
func TestNewEmptyBackup(t *testing.T) {
	backup := NewEmptyBackup()
	if backup == nil {
		t.Fatal("NewEmptyBackup() returned nil")
	}
	if backup.Settings != nil {
		t.Error("Empty backup should not have settings")
	}
}

// TestNewLargeBackup verifies large backup creation.
func TestNewLargeBackup(t *testing.T) {
	backup := NewLargeBackup()
	if backup == nil {
		t.Fatal("NewLargeBackup() returned nil")
	}
	if backup.Data == nil {
		t.Error("Backup should have data")
	}
	// Should have many profiles
	if backup.Data.Managed == nil {
		t.Error("Large backup should have managed profiles")
	}
}

// TestNewCustomBackup verifies custom backup creation.
func TestNewCustomBackup(t *testing.T) {
	description := "My custom backup"
	backup := NewCustomBackup(description)
	if backup == nil {
		t.Fatal("NewCustomBackup() returned nil")
	}
	if backup.Metadata.Description != description {
		t.Errorf("Expected description %q, got %q", description, backup.Metadata.Description)
	}
}

// TestWriteBackup verifies writing backup to test environment.
func TestWriteBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	backup := NewManagedOnlyBackup()
	backupPath := WriteBackup(t, backup)

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
}

// TestWriteBackupWithName verifies writing backup with custom name.
func TestWriteBackupWithName(t *testing.T) {
	test.SetupTestEnvironment(t)

	backup := NewManagedOnlyBackup()
	filename := "my-custom-backup.json"
	backupPath := WriteBackupWithName(t, backup, filename)

	// Verify file exists and has correct name
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}
	if filepath.Base(backupPath) != filename {
		t.Errorf("Expected filename %q, got %q", filename, filepath.Base(backupPath))
	}
}

// TestWriteInvalidBackup verifies writing invalid backup for error testing.
func TestWriteInvalidBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	backupPath := WriteInvalidBackup(t)

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Invalid backup file was not created")
	}

	// Verify content is invalid JSON
	data, _ := os.ReadFile(backupPath)
	if len(data) == 0 {
		t.Error("Invalid backup should have content")
	}
}
