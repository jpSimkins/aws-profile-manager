package profiles

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestRoundTrip_FullBackupAndRestore tests complete export → remove → import cycle
func TestRoundTrip_FullBackupAndRestore(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Step 1: Install profiles
	installer := NewInstaller(config)
	originalSchema := schematest.NewMixedSimple()

	installResult, err := installer.Install(context.Background(), InstallOptions{
		Schema: originalSchema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if installResult.TotalProfiles == 0 {
		t.Fatal("Should have installed profiles")
	}

	// Step 2: Export all sections
	exporter := NewExporter(config)
	backupPath := filepath.Join(test.GetTestConfigDir(t), "full-backup.json")

	exportResult, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
		Description:    "Full backup test",
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if exportResult.TotalProfiles == 0 {
		t.Fatal("Should have exported profiles")
	}

	// Step 3: Remove all profiles
	remover := NewRemover(config)
	removeResult, err := remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if !removeResult.RemovedConfig {
		t.Error("Should have removed config")
	}

	// Step 4: Import everything back
	importer := NewImporter(config)
	importResult, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	totalWritten := importResult.ManagedStats.ProfilesWritten + importResult.UnmanagedAboveStats.ProfilesWritten + importResult.UnmanagedBelowStats.ProfilesWritten
	if totalWritten == 0 {
		t.Fatal("Should have imported profiles")
	}

	// Step 5: Export again and compare
	backupPath2 := filepath.Join(test.GetTestConfigDir(t), "full-backup2.json")
	exportResult2, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     backupPath2,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Second export failed: %v", err)
	}

	// Verify profile counts match
	if exportResult2.TotalProfiles != exportResult.TotalProfiles {
		t.Errorf("Profile count mismatch: got %d, want %d",
			exportResult2.TotalProfiles, exportResult.TotalProfiles)
	}
}

// TestRoundTrip_ManagedOnly tests managed section backup and restore
func TestRoundTrip_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Install managed profiles
	installer := NewInstaller(config)
	originalSchema := schematest.NewManagedSsoMultiAccount()

	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: originalSchema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Export managed only
	exporter := NewExporter(config)
	backupPath := filepath.Join(test.GetTestConfigDir(t), "managed-backup.json")

	exportResult, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify only managed section exported
	if exportResult.ManagedProfiles == 0 {
		t.Error("Should have exported managed profiles")
	}
	if exportResult.UnmanagedAbove != 0 {
		t.Error("Should not have exported above section")
	}
	if exportResult.UnmanagedBelow != 0 {
		t.Error("Should not have exported below section")
	}

	// Remove and restore
	remover := NewRemover(config)
	_, err = remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	importer := NewImporter(config)
	importResult, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if importResult.ManagedStats.ProfilesWritten == 0 {
		t.Error("Should have imported managed profiles")
	}
}

// TestImporter_Import_DryRunWithVariousSchemas tests dry run with different schemas
func TestImporter_Import_DryRunWithVariousSchemas(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	importer := NewImporter(config)

	tests := []struct {
		name   string
		schema *schema.Schema
	}{
		{"empty schema", schematest.NewEmpty()},
		{"managed only", schematest.NewManagedSsoSingle()},
		{"unmanaged only", schematest.NewUnmanagedSsoSingle()},
		{"mixed", schematest.NewMixedSimple()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create backup
			backupPath := filepath.Join(test.GetTestConfigDir(t), tt.name+"-backup.json")
			data, _ := json.Marshal(tt.schema)
			_ = os.WriteFile(backupPath, data, 0644)

			// Dry run
			result, err := importer.Import(context.Background(), ImportOptions{
				BackupPath:     backupPath,
				IncludeManaged: true,
				IncludeAbove:   true,
				IncludeBelow:   true,
				DryRun:         true,
			}, task.NoOpReporter{})

			if err != nil {
				t.Fatalf("Dry run should not error: %v", err)
			}

			// Result should be returned
			if result == nil {
				t.Fatal("Result should not be nil")
			}
		})
	}
}

// TestInstaller_Install_WithLargeScale tests installation with large schema
func TestInstaller_Install_WithLargeScale(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	installer := NewInstaller(config)
	schema := schematest.NewLargeScale()

	result, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Large scale install should succeed: %v", err)
	}

	if result.TotalProfiles < 1000 {
		t.Errorf("Should have installed many profiles, got %d", result.TotalProfiles)
	}

	// Verify file was created
	if !fileExists(config.ConfigPath) {
		t.Error("Config file should exist")
	}

	// Verify file has content
	size, _ := getFileSize(config.ConfigPath)
	if size < 10000 {
		t.Errorf("Config file should be large, got %d bytes", size)
	}
}

// TestExporter_Export_WithAllProfileTypes tests export with all profile types
func TestExporter_Export_WithAllProfileTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Install all profile types
	installer := NewInstaller(config)
	schema := schematest.NewManagedAll()

	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Export
	exporter := NewExporter(config)
	backupPath := filepath.Join(test.GetTestConfigDir(t), "all-types-backup.json")

	result, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     backupPath,
		IncludeManaged: true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify exported schema has all types
	if result.Schema == nil {
		t.Fatal("Schema should not be nil")
	}

	if result.Schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	managed := result.Schema.Managed

	// Should have all types
	hasSSO := len(managed.Organizations) > 0
	hasIAM := len(managed.IamUsers) > 0
	hasAssumeRole := len(managed.AssumeRoleChains) > 0
	hasGeneric := len(managed.GenericProfiles) > 0

	if !hasSSO {
		t.Error("Should have SSO profiles")
	}
	if !hasIAM {
		t.Error("Should have IAM profiles")
	}
	if !hasAssumeRole {
		t.Error("Should have AssumeRole profiles")
	}
	if !hasGeneric {
		t.Error("Should have Generic profiles")
	}
}
