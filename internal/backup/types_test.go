package backup

import (
	"encoding/json"
	"testing"
	"time"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/settings"
)

// TestBackupFileJSON tests BackupFile JSON serialization.
func TestBackupFileJSON(t *testing.T) {
	// Create test backup
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			ToolVersion: "1.0.0",
			Description: "Test backup",
		},
		Data:     schematest.NewManagedSsoSingle(),
		Settings: settings.GetDefaults(),
	}

	// Marshal to JSON
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("Failed to marshal backup: %v", err)
	}

	// Unmarshal back
	var restored BackupFile
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	// Verify fields
	if restored.Version != backup.Version {
		t.Errorf("Version mismatch: got %s, want %s", restored.Version, backup.Version)
	}
	if restored.Metadata.ToolVersion != backup.Metadata.ToolVersion {
		t.Errorf("ToolVersion mismatch: got %s, want %s",
			restored.Metadata.ToolVersion, backup.Metadata.ToolVersion)
	}
	if restored.Data == nil {
		t.Error("Schema should not be nil")
	}
	if restored.Settings == nil {
		t.Error("Settings should not be nil")
	}
}

// TestBackupFileJSON_WithoutSettings tests backup without settings.
func TestBackupFileJSON_WithoutSettings(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewManagedSsoSingle(),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("Failed to marshal backup: %v", err)
	}

	var restored BackupFile
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	if restored.Settings != nil {
		t.Error("Settings should be nil")
	}
}

// TestBackupFileJSON_WithoutSchema tests backup without schema.
func TestBackupFileJSON_WithoutSchema(t *testing.T) {
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			ToolVersion: "1.0.0",
		},
		Settings: settings.GetDefaults(),
	}

	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("Failed to marshal backup: %v", err)
	}

	var restored BackupFile
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal backup: %v", err)
	}

	if restored.Data != nil {
		t.Error("Schema should be nil")
	}
}

// TestExportOptions tests ExportOptions struct.
func TestExportOptions(t *testing.T) {
	opts := ExportOptions{
		OutputPath:      "/tmp/backup.json",
		IncludeManaged:  true,
		IncludeAbove:    true,
		IncludeBelow:    true,
		ExcludeSettings: false,
		Description:     "Full backup",
	}

	if opts.OutputPath != "/tmp/backup.json" {
		t.Errorf("OutputPath mismatch: got %s", opts.OutputPath)
	}
	if !opts.IncludeManaged {
		t.Error("IncludeManaged should be true")
	}
	if !opts.IncludeAbove {
		t.Error("IncludeAbove should be true")
	}
	if !opts.IncludeBelow {
		t.Error("IncludeBelow should be true")
	}
	if opts.ExcludeSettings {
		t.Error("ExcludeSettings should be false")
	}
}

// TestImportOptions tests ImportOptions struct.
func TestImportOptions(t *testing.T) {
	opts := ImportOptions{
		BackupPath:            "/tmp/backup.json",
		IncludeManaged:        true,
		IncludeAbove:          true,
		IncludeBelow:          true,
		IgnoreSettings:        false,
		BackupCurrentSettings: true,
		DryRun:                false,
	}

	if opts.BackupPath != "/tmp/backup.json" {
		t.Errorf("BackupPath mismatch: got %s", opts.BackupPath)
	}
	if !opts.IncludeManaged {
		t.Error("IncludeManaged should be true")
	}
	if !opts.BackupCurrentSettings {
		t.Error("BackupCurrentSettings should be true")
	}
}

// TestExportResult tests ExportResult struct.
func TestExportResult(t *testing.T) {
	result := &ExportResult{
		BackupFile: &BackupFile{
			Version: "2.0",
			Metadata: BackupMetadata{
				ExportedAt:  time.Now(),
				ToolVersion: "1.0.0",
			},
			Data: &schema.Schema{},
		},
		OutputPath:       "/tmp/backup.json",
		ManagedProfiles:  10,
		UnmanagedAbove:   2,
		UnmanagedBelow:   3,
		TotalProfiles:    15,
		SettingsExported: true,
		Duration:         time.Second,
	}

	if result.TotalProfiles != 15 {
		t.Errorf("TotalProfiles mismatch: got %d, want 15", result.TotalProfiles)
	}
	if !result.SettingsExported {
		t.Error("SettingsExported should be true")
	}
}

// TestImportResult tests ImportResult struct.
func TestImportResult(t *testing.T) {
	result := ImportResult{
		BackupFile: &BackupFile{
			Version: "2.0",
			Metadata: BackupMetadata{
				ExportedAt:  time.Now(),
				ToolVersion: "1.0.0",
			},
		},
		ConfigPath: "/home/user/.aws/config",
		ManagedStats: generators.SectionStats{
			ProfilesWritten: 10,
		},
		UnmanagedAboveStats: generators.SectionStats{
			ProfilesWritten: 2,
		},
		UnmanagedBelowStats: generators.SectionStats{
			ProfilesWritten: 3,
		},
		ManagedDuplicates: profiles.SectionDuplicateStats{
			TotalDuplicates: 1,
		},
		SettingsRestored:   true,
		SettingsBackupPath: "/tmp/settings-backup.json",
		Duration:           time.Second,
	}

	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if totalWritten != 15 {
		t.Errorf("TotalProfilesWritten mismatch: got %d, want 15", totalWritten)
	}
	if result.ManagedDuplicates.TotalDuplicates != 1 {
		t.Errorf("DuplicatesSkipped mismatch: got %d, want 1", result.ManagedDuplicates.TotalDuplicates)
	}
	if !result.SettingsRestored {
		t.Error("SettingsRestored should be true")
	}
}
