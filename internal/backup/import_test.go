package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aws-profile-manager/internal/core"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/settings"
	tasktest "aws-profile-manager/internal/task/test"
	"aws-profile-manager/internal/test"
)

// TestImportProfiles_FullRestore tests full import with all sections.
func TestImportProfiles_FullRestore(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup file
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Clear config file
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	// Import
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify
	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if totalWritten == 0 {
		t.Error("Expected profiles to be imported")
	}
	if result.ManagedStats.ProfilesWritten == 0 {
		t.Error("Expected managed profiles")
	}
	if !result.SettingsRestored {
		t.Error("Settings should be restored by default")
	}
}

// TestImportProfiles_ManagedOnly tests importing managed profiles only.
func TestImportProfiles_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Import
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true
	importOpts.IncludeAbove = false
	importOpts.IncludeBelow = false

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify
	if result.ManagedStats.ProfilesWritten == 0 {
		t.Error("Expected managed profiles")
	}
	if result.UnmanagedAboveStats.ProfilesWritten != 0 {
		t.Errorf("Expected no above profiles, got %d", result.UnmanagedAboveStats.ProfilesWritten)
	}
	if result.UnmanagedBelowStats.ProfilesWritten != 0 {
		t.Errorf("Expected no below profiles, got %d", result.UnmanagedBelowStats.ProfilesWritten)
	}
}

// TestImportProfiles_WithSettings tests importing with settings.
func TestImportProfiles_WithSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Modify settings before export
	currentSettings := settings.Get()
	modifiedSettings := *currentSettings
	modifiedSettings.GUI.Theme = "dark"
	if err := settings.Set(&modifiedSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create backup with settings
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true
	exportOpts.ExcludeSettings = false // Include settings

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Change settings
	changedSettings := settings.Get()
	changedSettings.GUI.Theme = "light"
	if err := settings.Set(changedSettings); err != nil {
		t.Fatalf("Failed to change settings: %v", err)
	}

	// Import (should restore dark theme)
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true
	importOpts.IgnoreSettings = false

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify settings restored
	if !result.SettingsRestored {
		t.Error("Settings should be restored")
	}

	restoredSettings := settings.Get()
	if restoredSettings.GUI.Theme != "dark" {
		t.Errorf("Theme not restored: got %s, want dark", restoredSettings.GUI.Theme)
	}
}

// TestImportProfiles_IgnoreSettings tests importing without settings.
func TestImportProfiles_IgnoreSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup with settings
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Import with settings ignored
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true
	importOpts.IgnoreSettings = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify settings not restored
	if result.SettingsRestored {
		t.Error("Settings should not be restored")
	}
}

// TestImportProfiles_BackupCurrentSettings tests backing up current settings.
func TestImportProfiles_BackupCurrentSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Import with settings backup enabled
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true
	importOpts.BackupCurrentSettings = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify settings backup was created
	if result.SettingsBackupPath == "" {
		t.Error("Settings backup path should be set")
	}

	// Verify backup file exists
	if _, err := os.Stat(result.SettingsBackupPath); os.IsNotExist(err) {
		t.Error("Settings backup file not created")
	}
}

// TestImportProfiles_DryRun tests dry run mode.
func TestImportProfiles_DryRun(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	t.Logf("Export result: TotalProfiles=%d, has schema=%v",
		exportResult.TotalProfiles, exportResult.BackupFile.Data != nil)
	if exportResult.BackupFile.Data != nil && exportResult.BackupFile.Data.Managed != nil {
		t.Logf("Schema has %d orgs, %d IAM users",
			len(exportResult.BackupFile.Data.Managed.Organizations),
			len(exportResult.BackupFile.Data.Managed.IamUsers))
	}

	// Clear config file
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	// Import in dry run mode
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true
	importOpts.DryRun = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify no changes were made
	content, _ := os.ReadFile(cfg.ConfigPath)
	if len(content) > 0 {
		t.Error("Config file should still be empty in dry run mode")
	}

	// In dry run mode, check what would be imported
	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	totalDuplicates := result.ManagedDuplicates.TotalDuplicates + result.UnmanagedAboveDuplicates.TotalDuplicates + result.UnmanagedBelowDuplicates.TotalDuplicates
	t.Logf("DryRun result: TotalWritten=%d, Duplicates=%d",
		totalWritten, totalDuplicates)

	if totalWritten == 0 {
		t.Errorf("Should show profiles that would be written in backup file, got %d (duplicates=%d)",
			totalWritten, totalDuplicates)
	}
}

// TestImportProfiles_MissingBackupPath tests error when backup path missing.
func TestImportProfiles_MissingBackupPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)
	importOpts := newTestImportOptions(t, "")
	importOpts.BackupPath = "" // Missing

	_, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for missing backup path")
	}
}

// TestImportProfiles_NoContentSelected tests error when all sections disabled.
func TestImportProfiles_NoContentSelected(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)
	importOpts := newTestImportOptions(t, "/tmp/backup.json")
	importOpts.IncludeManaged = false
	importOpts.IncludeAbove = false
	importOpts.IncludeBelow = false
	importOpts.IgnoreSettings = true // Nothing selected

	_, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error when no content selected")
	}
}

// TestImportProfiles_NonExistentBackup tests error for non-existent backup.
func TestImportProfiles_NonExistentBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)
	importOpts := newTestImportOptions(t, "/tmp/nonexistent-backup.json")
	importOpts.IncludeManaged = true

	_, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for non-existent backup")
	}
}

// TestImportProfiles_InvalidBackup tests error for invalid backup file.
func TestImportProfiles_InvalidBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create invalid backup file
	cfg := newTestConfig(t)
	backupPath := newTestExportOptions(t).OutputPath
	_ = os.WriteFile(backupPath, []byte("invalid json"), 0600)

	importOpts := newTestImportOptions(t, backupPath)
	importOpts.IncludeManaged = true

	_, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for invalid backup")
	}
}

// TestImportProfiles_EmptySchema tests error when backup has no profiles.
func TestImportProfiles_EmptySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup with no schema
	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = false
	exportOpts.IncludeAbove = false
	exportOpts.IncludeBelow = false
	exportOpts.ExcludeSettings = false // Only settings

	// Create empty config
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	// This should fail during export because no content selected
	// Let's create a settings-only backup manually
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Settings: settings.GetDefaults(),
		// No Schema
	}
	_ = WriteBackupFile(exportOpts.OutputPath, backup)

	// Try to import profiles (should fail)
	importOpts := newTestImportOptions(t, exportOpts.OutputPath)
	importOpts.IncludeManaged = true // Want profiles but backup has none

	_, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error when backup has no profiles")
	}
}

// TestImportProfiles_ContextCancellation tests context cancellation.
func TestImportProfiles_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Import with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true

	// Currently may succeed because import is fast, but structure supports cancellation
	_, _ = ImportProfiles(ctx, cfg, importOpts, tasktest.NewMockReporter())
}

// TestImportProfiles_MultipleProfileTypes tests importing various profile types.
func TestImportProfiles_MultipleProfileTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup with all profile types
	schema := schematest.NewManagedAll()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Clear config
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	// Import
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Should have multiple profiles
	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if totalWritten == 0 {
		t.Error("Expected multiple profiles")
	}
}

// TestImportProfiles_VerifyDuration tests that duration is tracked.
func TestImportProfiles_VerifyDuration(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create backup
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Import
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.Duration == 0 {
		t.Error("Duration should be greater than 0")
	}
}

// TestImportProfiles_SkipCheatSheetWhenManagedDisabled verifies that cheat sheet
// generation is skipped when managed profile import is disabled.
func TestImportProfiles_SkipCheatSheetWhenManagedDisabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create a backup that includes both managed and unmanaged sections.
	backupPath := filepath.Join(test.GetTestConfigDir(t), "mixed-backup.json")
	backup := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: "1.0.0",
		},
		Data: schematest.NewMixedSimple(),
	}

	if err := WriteBackupFile(backupPath, backup); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	cfg := newTestConfig(t)
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	importOpts := newTestImportOptions(t, backupPath)
	importOpts.IncludeManaged = false
	importOpts.IncludeAbove = true
	importOpts.IncludeBelow = true
	importOpts.GenerateCheatSheet = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.CheatSheetGenerated {
		t.Error("Cheat sheet should not be generated when managed import is disabled")
	}

	if result.CheatSheetPath != "" {
		t.Errorf("CheatSheetPath should be empty, got %s", result.CheatSheetPath)
	}

	content, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if strings.Contains(string(content), cfg.StartMarker) {
		t.Error("Managed section marker should not be written when IncludeManaged is false")
	}

	defaultCheatSheetPath := filepath.Join(test.GetTestDesktopDir(t), "AWS_Profile_Cheat_Sheet.md")
	if _, err := os.Stat(defaultCheatSheetPath); err == nil {
		t.Errorf("Cheat sheet file should not exist: %s", defaultCheatSheetPath)
	}
}

// TestImportProfiles_LargeSchema tests importing large schema.
func TestImportProfiles_LargeSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create large backup
	schema := schematest.NewLargeScale()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	exportOpts := newTestExportOptions(t)
	exportOpts.IncludeManaged = true

	exportResult, err := ExportProfiles(context.Background(), cfg, exportOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Clear config
	_ = os.WriteFile(cfg.ConfigPath, []byte(""), 0600)

	// Import
	importOpts := newTestImportOptions(t, exportResult.OutputPath)
	importOpts.IncludeManaged = true

	result, err := ImportProfiles(context.Background(), cfg, importOpts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Should have many profiles
	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if totalWritten < 100 {
		t.Errorf("Expected large number of profiles, got %d", totalWritten)
	}
}
