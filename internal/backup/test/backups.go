// Package test provides test fixtures for backup package.
//
// This package is for use by OTHER packages that need backup test data
// (e.g., CLI tests, integration tests). It CANNOT be used by the
// backup package itself due to import cycles.
//
// The backup package has its own internal test fixtures in *_test.go files.
//
// This package provides pre-built BackupFile objects and helpers for
// writing them to test environments.
//
// Usage:
//
//	import backuptest "aws-profile-manager/internal/backup/test"
//
//	func TestMyFeature(t *testing.T) {
//	    test.SetupTestEnvironment(t)
//	    backupPath := backuptest.WriteBackup(t, backuptest.NewManagedOnlyBackup())
//	    // ... test with backup file
//	}
package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/backup"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	settingstest "aws-profile-manager/internal/settings/test"
	"aws-profile-manager/internal/test"
)

// NewManagedOnlyBackup returns a backup file with managed profiles only.
//
// Uses schema/test's NewManagedSsoSingle for consistency.
//
// Returns:
//   - *backup.BackupFile: Backup with managed SSO profile
func NewManagedOnlyBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test backup with managed profiles only",
		},
		Data:     schematest.NewManagedSsoSingle(),
		Settings: nil, // No settings
	}
}

// NewFullBackup returns a backup file with profiles and settings.
//
// Uses schema/test's NewManagedAll and settings/test's NewDefault.
//
// Returns:
//   - *backup.BackupFile: Complete backup with profiles and settings
func NewFullBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test full backup with profiles and settings",
		},
		Data:     schematest.NewManagedAll(),
		Settings: settingstest.NewDefault(),
	}
}

// NewMixedBackup returns a backup file with managed and unmanaged profiles.
//
// Uses schema/test's NewMixedSimple for consistency.
//
// Returns:
//   - *backup.BackupFile: Backup with mixed profiles
func NewMixedBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test backup with managed and unmanaged profiles",
		},
		Data:     schematest.NewMixedSimple(),
		Settings: nil,
	}
}

// NewSettingsOnlyBackup returns a backup file with settings only (no profiles).
//
// Uses settings/test's NewDefault for consistency.
//
// Returns:
//   - *backup.BackupFile: Backup with settings only
func NewSettingsOnlyBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test backup with settings only",
		},
		Data:     nil, // No profiles
		Settings: settingstest.NewDefault(),
	}
}

// NewEmptyBackup returns a backup file with no profiles or settings.
//
// Returns:
//   - *backup.BackupFile: Empty backup (only metadata)
func NewEmptyBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test empty backup",
		},
		Data:     &schema.Schema{Version: "2.0"}, // Empty schema
		Settings: nil,
	}
}

// NewLargeBackup returns a backup file with many profiles.
//
// Uses schema/test's NewLargeScale for performance testing.
//
// Returns:
//   - *backup.BackupFile: Backup with 2100+ profiles
func NewLargeBackup() *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: "Test large backup with 2100+ profiles",
		},
		Data:     schematest.NewLargeScale(),
		Settings: nil,
	}
}

// NewCustomBackup returns a backup file with custom description.
//
// Parameters:
//   - description: Custom description for the backup
//
// Returns:
//   - *backup.BackupFile: Backup with custom description
func NewCustomBackup(description string) *backup.BackupFile {
	return &backup.BackupFile{
		Version: "2.0",
		Metadata: backup.BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: description,
		},
		Data:     schematest.NewManagedSsoSingle(),
		Settings: settingstest.NewDefault(),
	}
}

// WriteBackup writes a backup file to the test environment and returns the path.
//
// The file is written to the test config directory with a predictable name.
//
// Parameters:
//   - t: Testing context
//   - backupFile: BackupFile to write
//
// Returns:
//   - string: Path to written backup file
//
// Example:
//
//	test.SetupTestEnvironment(t)
//	backupPath := backuptest.WriteBackup(t, backuptest.NewFullBackup())
//	// ... test with backup at backupPath
func WriteBackup(t *testing.T, backupFile *backup.BackupFile) string {
	t.Helper()

	backupPath := filepath.Join(test.GetTestConfigDir(t), "test-backup.json")
	if err := backup.WriteBackupFile(backupPath, backupFile); err != nil {
		t.Fatalf("Failed to write test backup: %v", err)
	}

	return backupPath
}

// WriteBackupWithName writes a backup file with a specific name.
//
// Parameters:
//   - t: Testing context
//   - backupFile: BackupFile to write
//   - filename: Name for the backup file (e.g., "my-backup.json")
//
// Returns:
//   - string: Path to written backup file
func WriteBackupWithName(t *testing.T, backupFile *backup.BackupFile, filename string) string {
	t.Helper()

	backupPath := filepath.Join(test.GetTestConfigDir(t), filename)
	if err := backup.WriteBackupFile(backupPath, backupFile); err != nil {
		t.Fatalf("Failed to write test backup: %v", err)
	}

	return backupPath
}

// WriteInvalidBackup writes an invalid JSON file for error testing.
//
// Parameters:
//   - t: Testing context
//
// Returns:
//   - string: Path to invalid backup file
func WriteInvalidBackup(t *testing.T) string {
	t.Helper()

	backupPath := filepath.Join(test.GetTestConfigDir(t), "invalid-backup.json")
	if err := os.WriteFile(backupPath, []byte("invalid json {["), 0600); err != nil {
		t.Fatalf("Failed to write invalid backup: %v", err)
	}

	return backupPath
}
