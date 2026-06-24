package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

// TestReadBackupFile tests reading a valid backup file.
func TestReadBackupFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a valid backup file
	backupPath := filepath.Join(test.GetTestConfigDir(t), "test-backup.json")
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
			Description: "Test backup",
		},
		Data:     schematest.NewManagedSsoSingle(),
		Settings: settings.GetDefaults(),
	}

	// Write backup file
	if err := WriteBackupFile(backupPath, backup); err != nil {
		t.Fatalf("Failed to write test backup: %v", err)
	}

	// Read backup file
	restored, err := ReadBackupFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	// Verify
	if restored.Version != backup.Version {
		t.Errorf("Version mismatch: got %s, want %s", restored.Version, backup.Version)
	}
	if restored.Metadata.ToolVersion != backup.Metadata.ToolVersion {
		t.Errorf("ToolVersion mismatch: got %s, want %s",
			restored.Metadata.ToolVersion, backup.Metadata.ToolVersion)
	}
	if restored.Data == nil {
		t.Error("Data should not be nil")
	}
	if restored.Settings == nil {
		t.Error("Settings should not be nil")
	}
}

// TestReadBackupFile_NonExistent tests reading a non-existent file.
func TestReadBackupFile_NonExistent(t *testing.T) {
	test.SetupTestEnvironment(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "nonexistent.json")

	_, err := ReadBackupFile(backupPath)
	if err == nil {
		t.Fatal("Expected error for non-existent file")
	}
}

// TestReadBackupFile_InvalidJSON tests reading invalid JSON.
func TestReadBackupFile_InvalidJSON(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Write invalid JSON
	backupPath := filepath.Join(test.GetTestConfigDir(t), "invalid.json")
	_ = os.WriteFile(backupPath, []byte("not valid json"), 0600)

	_, err := ReadBackupFile(backupPath)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

// TestReadBackupFile_InvalidStructure tests reading backup with missing fields.
func TestReadBackupFile_InvalidStructure(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Write backup missing version
	backupPath := filepath.Join(test.GetTestConfigDir(t), "invalid-structure.json")
	data := []byte(`{
		"metadata": {
			"exported_at": "2025-01-01T12:00:00Z",
			"tool_version": "1.0.0"
		}
	}`)
	_ = os.WriteFile(backupPath, data, 0600)

	_, err := ReadBackupFile(backupPath)
	if err == nil {
		t.Fatal("Expected error for invalid structure")
	}
}

// TestWriteBackupFile tests writing a backup file.
func TestWriteBackupFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "write-test.json")
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
			Description: "Write test",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	// Write
	if err := WriteBackupFile(backupPath, backup); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file was not created")
	}

	// Read back and verify
	restored, err := ReadBackupFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read written backup: %v", err)
	}

	if restored.Version != backup.Version {
		t.Errorf("Version mismatch: got %s, want %s", restored.Version, backup.Version)
	}
}

// TestWriteBackupFile_CreatesDirectory tests that parent directory is created.
func TestWriteBackupFile_CreatesDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Use nested directory that doesn't exist
	backupPath := filepath.Join(test.GetTestConfigDir(t), "nested", "dir", "backup.json")
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	// Write (should create directories)
	if err := WriteBackupFile(backupPath, backup); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file was not created")
	}
}

// TestWriteBackupFile_InvalidBackup tests writing an invalid backup.
func TestWriteBackupFile_InvalidBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "invalid.json")

	// Missing version
	backup := &BackupFile{
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	err := WriteBackupFile(backupPath, backup)
	if err == nil {
		t.Fatal("Expected error for invalid backup")
	}
}

// TestValidateBackupFile_Valid tests validation of valid backup.
func TestValidateBackupFile_Valid(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	if err := ValidateBackupFile(backup); err != nil {
		t.Errorf("Valid backup failed validation: %v", err)
	}
}

// TestValidateBackupFile_NilBackup tests validation of nil backup.
func TestValidateBackupFile_NilBackup(t *testing.T) {
	err := ValidateBackupFile(nil)
	if err == nil {
		t.Fatal("Expected error for nil backup")
	}
}

// TestValidateBackupFile_MissingVersion tests validation without version.
func TestValidateBackupFile_MissingVersion(t *testing.T) {
	backup := &BackupFile{
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	err := ValidateBackupFile(backup)
	if err == nil {
		t.Fatal("Expected error for missing version")
	}
}

// TestValidateBackupFile_EmptyBackup tests validation with no schema or settings.
func TestValidateBackupFile_EmptyBackup(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
	}

	err := ValidateBackupFile(backup)
	if err == nil {
		t.Fatal("Expected error for empty backup (no schema or settings)")
	}
}

// TestValidateBackupFile_MissingTimestamp tests validation without timestamp.
func TestValidateBackupFile_MissingTimestamp(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	err := ValidateBackupFile(backup)
	if err == nil {
		t.Fatal("Expected error for missing timestamp")
	}
}

// TestValidateBackupFile_MissingToolVersion tests validation without tool version.
func TestValidateBackupFile_MissingToolVersion(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt: time.Now(),
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	err := ValidateBackupFile(backup)
	if err == nil {
		t.Fatal("Expected error for missing tool version")
	}
}

// TestValidateBackupFile_SchemaOnly tests validation with schema only.
func TestValidateBackupFile_SchemaOnly(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	if err := ValidateBackupFile(backup); err != nil {
		t.Errorf("Schema-only backup failed validation: %v", err)
	}
}

// TestValidateBackupFile_SettingsOnly tests validation with settings only.
func TestValidateBackupFile_SettingsOnly(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Settings: settings.GetDefaults(),
	}

	if err := ValidateBackupFile(backup); err != nil {
		t.Errorf("Settings-only backup failed validation: %v", err)
	}
}
