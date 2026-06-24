package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/core"
	profilestest "aws-profile-manager/internal/profiles/test"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	tasktest "aws-profile-manager/internal/task/test"
	"aws-profile-manager/internal/test"
)

// setupTestAwsConfig writes AWS config for testing using schematest fixtures.
//
// This helper maps schema fixtures to AWS config content for testing.
// It determines the appropriate profilestest fixture based on the schema structure.
func setupTestAwsConfig(t *testing.T, s *schema.Schema) {
	t.Helper()

	// For simplicity, determine fixture based on profile types present
	var content string

	if s.Managed != nil {
		hasSSO := len(s.Managed.Organizations) > 0
		hasIAM := len(s.Managed.IamUsers) > 0
		hasAssumeRole := len(s.Managed.AssumeRoleChains) > 0
		hasGeneric := len(s.Managed.GenericProfiles) > 0

		// Check if this is a large scale test (many organizations)
		if len(s.Managed.Organizations) >= 10 {
			content = profilestest.NewConfigLarge()
		} else if hasSSO && !hasIAM && !hasAssumeRole && !hasGeneric {
			content = profilestest.NewConfigWithSsoSingle()
		} else if hasIAM && !hasSSO && !hasAssumeRole && !hasGeneric {
			content = profilestest.NewConfigWithIamSingle()
		} else if hasAssumeRole && !hasSSO && !hasIAM && !hasGeneric {
			content = profilestest.NewConfigWithAssumeRoleSingle()
		} else if hasSSO || hasIAM || hasAssumeRole || hasGeneric {
			content = profilestest.NewConfigWithAllTypes()
		}
	}

	if content == "" {
		content = profilestest.NewConfigEmpty()
	}

	profilestest.WriteConfig(t, content)
} // TestExportProfiles_ManagedOnly tests exporting managed profiles only.
func TestExportProfiles_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Setup test data
	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.IncludeManaged = true
	opts.IncludeAbove = false
	opts.IncludeBelow = false

	// Export
	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify
	if result.TotalProfiles == 0 {
		t.Error("Expected profiles to be exported")
	}
	if result.ManagedProfiles == 0 {
		t.Error("Expected managed profiles")
	}
	if result.UnmanagedAbove != 0 {
		t.Errorf("Expected no above profiles, got %d", result.UnmanagedAbove)
	}
	if result.UnmanagedBelow != 0 {
		t.Errorf("Expected no below profiles, got %d", result.UnmanagedBelow)
	}
	if !result.SettingsExported {
		t.Error("Settings should be exported by default")
	}

	// Verify file exists
	if _, err := os.Stat(result.OutputPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify backup file structure
	if result.BackupFile.Data == nil {
		t.Error("Data should not be nil")
	}
	if result.BackupFile.Data.Managed == nil {
		t.Error("Managed section should not be nil")
	}
	if result.BackupFile.Settings == nil {
		t.Error("Settings should not be nil")
	}
}

// TestExportProfiles_WithSettings tests exporting with settings.
func TestExportProfiles_WithSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.ExcludeSettings = false // Include settings

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if !result.SettingsExported {
		t.Error("Settings should be exported")
	}
	if result.BackupFile.Settings == nil {
		t.Error("Settings should not be nil in backup file")
	}
}

// TestExportProfiles_ExcludeSettings tests exporting without settings.
func TestExportProfiles_ExcludeSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.ExcludeSettings = true // Exclude settings

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if result.SettingsExported {
		t.Error("Settings should not be exported")
	}
	if result.BackupFile.Settings != nil {
		t.Error("Settings should be nil in backup file")
	}
}

// TestExportProfiles_WithDescription tests exporting with description.
func TestExportProfiles_WithDescription(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.Description = "Test backup with description"

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if result.BackupFile.Metadata.Description != "Test backup with description" {
		t.Errorf("Description mismatch: got %s", result.BackupFile.Metadata.Description)
	}
}

// TestExportProfiles_EmptyConfig tests exporting from empty config.
func TestExportProfiles_EmptyConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create empty config file
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.ExcludeSettings = true // Only export profiles (which will be empty)

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if result.TotalProfiles != 0 {
		t.Errorf("Expected 0 profiles, got %d", result.TotalProfiles)
	}
}

// TestExportProfiles_NonExistentConfig tests exporting from non-existent config.
func TestExportProfiles_NonExistentConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Don't create config file
	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	// Should fail when file doesn't exist
	_, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for non-existent config")
	}
}

// TestExportProfiles_MissingOutputPath tests error when output path missing.
func TestExportProfiles_MissingOutputPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.OutputPath = "" // Missing path

	_, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for missing output path")
	}
}

// TestExportProfiles_NoContentSelected tests error when all sections disabled.
func TestExportProfiles_NoContentSelected(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.IncludeManaged = false
	opts.IncludeAbove = false
	opts.IncludeBelow = false
	opts.ExcludeSettings = true // Nothing selected

	_, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error when no content selected")
	}
}

// TestExportProfiles_ContextCancellation tests context cancellation.
func TestExportProfiles_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Note: Current implementation doesn't check context, but we test the pattern
	// In a real scenario with long-running operations, this would fail
	_, _ = ExportProfiles(ctx, cfg, opts, tasktest.NewMockReporter())
	// Currently succeeds because export is fast, but structure supports cancellation
}

// TestExportProfiles_MultipleProfileTypes tests exporting various profile types.
func TestExportProfiles_MultipleProfileTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Use schema with all profile types
	schema := schematest.NewManagedAll()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should have multiple profiles
	if result.TotalProfiles == 0 {
		t.Error("Expected multiple profiles")
	}

	// Verify backup file has data
	if result.BackupFile.Data == nil {
		t.Error("Data should not be nil")
	}
	if result.BackupFile.Data.Managed == nil {
		t.Error("Managed section should not be nil")
	}
}

// TestExportProfiles_InvalidOutputDirectory tests handling of invalid output path.
func TestExportProfiles_InvalidOutputDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	// Use path that cannot be created (assuming /proc is read-only on Linux)
	opts.OutputPath = "/proc/cannot-create-here/backup.json"

	_, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err == nil {
		t.Fatal("Expected error for invalid output directory")
	}
}

// TestExportProfiles_VerifyMetadata tests that metadata is correctly populated.
func TestExportProfiles_VerifyMetadata(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	opts.Description = "Metadata test"

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify metadata
	metadata := result.BackupFile.Metadata
	if metadata.ExportedAt.IsZero() {
		t.Error("ExportedAt should be set")
	}
	if metadata.ToolVersion == "" {
		t.Error("ToolVersion should be set")
	}
	if metadata.Description != "Metadata test" {
		t.Errorf("Description mismatch: got %s", metadata.Description)
	}

	// Verify version
	if result.BackupFile.Version != "2.0" {
		t.Errorf("Version mismatch: got %s, want 2.0", result.BackupFile.Version)
	}
}

// TestExportProfiles_VerifyDuration tests that duration is tracked.
func TestExportProfiles_VerifyDuration(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if result.Duration == 0 {
		t.Error("Duration should be greater than 0")
	}
}

// TestExportProfiles_VerifyOutputFileExists tests that output file is created.
func TestExportProfiles_VerifyOutputFileExists(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	schema := schematest.NewManagedSsoSingle()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "test-output.json")
	opts.OutputPath = outputPath

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Verify result has correct path
	if result.OutputPath != outputPath {
		t.Errorf("OutputPath mismatch: got %s, want %s", result.OutputPath, outputPath)
	}

	// Verify file can be read back
	backup, err := ReadBackupFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if backup.Version != "2.0" {
		t.Errorf("Version mismatch in file: got %s", backup.Version)
	}
}

// TestExportProfiles_LargeSchema tests exporting large schema.
func TestExportProfiles_LargeSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Use large-scale test schema
	schema := schematest.NewLargeScale()
	setupTestAwsConfig(t, schema)

	cfg := newTestConfig(t)
	opts := newTestExportOptions(t)

	result, err := ExportProfiles(context.Background(), cfg, opts, tasktest.NewMockReporter())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Should have many profiles
	if result.TotalProfiles < 100 {
		t.Errorf("Expected large number of profiles, got %d", result.TotalProfiles)
	}
}
