package profiles

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestNewImporter tests Importer constructor
func TestNewImporter(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	importer := NewImporter(config)

	if importer == nil {
		t.Fatal("NewImporter should not return nil")
	}
}

// TestImporter_Import_ManagedOnly tests importing only managed profiles
func TestImporter_Import_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath
	backupPath := filepath.Join(test.GetTestConfigDir(t), "backup.json")

	// Create backup file with managed profiles
	testSchema := schematest.NewManagedSsoSingle()
	backupData, err := json.Marshal(testSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Import managed profiles
	importer := NewImporter(config)
	result, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.ManagedStats.ProfilesWritten == 0 {
		t.Error("Should have written profiles")
	}

	// Verify config was created
	if !fileExists(configPath) {
		t.Error("Config file should have been created")
	}

	// Verify config contains markers
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# START") {
		t.Error("Config should contain start marker")
	}
	if !strings.Contains(contentStr, "# END") {
		t.Error("Config should contain end marker")
	}
}

// TestImporter_Import_FullRestore tests importing all sections
func TestImporter_Import_FullRestore(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath
	backupPath := filepath.Join(test.GetTestConfigDir(t), "full-backup.json")

	// Create backup with all sections
	testSchema := schematest.NewMixedSimple()
	backupData, err := json.Marshal(testSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Import all sections
	importer := NewImporter(config)
	result, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.ManagedStats.ProfilesWritten == 0 {
		t.Error("Should have imported managed profiles")
	}

	// Verify config was created
	if !fileExists(configPath) {
		t.Error("Config file should have been created")
	}
}

// TestImporter_Import_DryRun tests dry-run mode
func TestImporter_Import_DryRun(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	t.Run("no existing config - should not create file", func(t *testing.T) {
		configPath := config.ConfigPath
		backupPath := filepath.Join(test.GetTestConfigDir(t), "backup-dryrun-noconfig.json")

		// Ensure no config exists
		_ = os.Remove(configPath)

		// Create backup file
		testSchema := schematest.NewManagedSsoSingle()
		backupData, _ := json.Marshal(testSchema)
		_ = os.WriteFile(backupPath, backupData, 0644)

		// Dry run import
		importer := NewImporter(config)
		result, err := importer.Import(context.Background(), ImportOptions{
			BackupPath:     backupPath,
			IncludeManaged: true,
			DryRun:         true,
		}, task.NoOpReporter{})

		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		// Verify stats are populated
		if result.ManagedStats.ProfilesWritten == 0 {
			t.Error("Expected profile counts in dry-run result")
		}

		// Config file should NOT be created in dry-run
		if fileExists(configPath) {
			t.Error("Config file should not be created in dry-run mode")
		}
	})

	t.Run("with existing config - should not modify", func(t *testing.T) {
		configPath := config.ConfigPath
		backupPath := filepath.Join(test.GetTestConfigDir(t), "backup-dryrun-existing.json")

		// Create initial config with some content
		initialContent := "# Existing config\n[profile existing]\nregion = us-east-1\n"
		_ = os.WriteFile(configPath, []byte(initialContent), 0644)

		// Create backup file with different profiles
		testSchema := schematest.NewManagedSsoSingle()
		backupData, _ := json.Marshal(testSchema)
		_ = os.WriteFile(backupPath, backupData, 0644)

		// Dry run import
		importer := NewImporter(config)
		result, err := importer.Import(context.Background(), ImportOptions{
			BackupPath:     backupPath,
			IncludeManaged: true,
			DryRun:         true,
		}, task.NoOpReporter{})

		if err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		if result == nil {
			t.Fatal("Result should not be nil")
		}

		// Config should not be modified
		content, _ := os.ReadFile(configPath)
		if string(content) != initialContent {
			t.Errorf("Config file was modified in dry-run mode.\nExpected: %q\nGot: %q",
				initialContent, string(content))
		}
	})
}

// TestImporter_Import_InvalidBackup tests error handling
func TestImporter_Import_InvalidBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "invalid.json")

	// Create invalid JSON file
	if err := os.WriteFile(backupPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Import should fail
	importer := NewImporter(config)
	_, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Import should fail with invalid JSON")
	}
}

// TestImporter_Import_NonExistentBackup tests missing backup file
func TestImporter_Import_NonExistentBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "nonexistent.json")

	importer := NewImporter(config)
	_, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Import should fail when backup file doesn't exist")
	}
}

// TestImporter_Import_NoSectionsSelected tests importing with no sections selected
// Note: Importer may allow this and just do nothing, which is valid behavior
func TestImporter_Import_NoSectionsSelected(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "backup.json")

	// Create valid backup
	testSchema := schematest.NewManagedSsoSingle()
	backupData, err := json.Marshal(testSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	importer := NewImporter(config)
	result, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: false,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	// Either fail or succeed with 0 profiles
	totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if err == nil && totalWritten != 0 {
		t.Error("Should not have written profiles when no sections selected")
	}
}

// TestImporter_Import_UnmanagedOnly tests importing only personal profiles
func TestImporter_Import_UnmanagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath
	backupPath := filepath.Join(test.GetTestConfigDir(t), "personal-backup.json")

	// Create empty config file first
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Use unmanaged test schema
	testSchema := schematest.NewUnmanagedGenericSingle()
	backupData, err := json.Marshal(testSchema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
		t.Fatalf("Failed to write backup file: %v", err)
	}

	// Import personal profiles only
	importer := NewImporter(config)
	result, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: false,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if result.ManagedStats.ProfilesWritten != 0 {
		t.Error("Should not have imported managed profiles")
	}

	totalPersonal := result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	if totalPersonal == 0 {
		t.Error("Should have imported personal profiles")
	}

	// Verify config was created
	if !fileExists(configPath) {
		t.Error("Config file should have been created")
	}
}

// TestImporter_Import_MultipleSchemas tests different schema types
func TestImporter_Import_MultipleSchemas(t *testing.T) {
	tests := []struct {
		name   string
		schema func() *schema.Schema
	}{
		{"managed SSO single", schematest.NewManagedSsoSingle},
		{"managed SSO multi", schematest.NewManagedSsoMultiAccount},
		{"managed IAM", schematest.NewManagedIamSingle},
		{"managed all types", schematest.NewManagedAll},
		{"mixed simple", schematest.NewMixedSimple},
		{"unmanaged generic", schematest.NewUnmanagedGenericSingle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SetupTestEnvironment(t)

			config := newTestConfig(t)

			configPath := config.ConfigPath
			backupPath := filepath.Join(test.GetTestConfigDir(t), "backup.json")

			// Create empty config file if needed (for unmanaged imports)
			if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Create backup
			testSchema := tt.schema()
			backupData, err := json.Marshal(testSchema)
			if err != nil {
				t.Fatalf("Failed to marshal schema: %v", err)
			}
			if err := os.WriteFile(backupPath, backupData, 0644); err != nil {
				t.Fatalf("Failed to write backup file: %v", err)
			}

			// Import all sections
			importer := NewImporter(config)
			result, err := importer.Import(context.Background(), ImportOptions{
				BackupPath:     backupPath,
				IncludeManaged: true,
				IncludeAbove:   true,
				IncludeBelow:   true,
			}, task.NoOpReporter{})

			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			totalWritten := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
			if totalWritten == 0 {
				t.Error("Should have written profiles")
			}

			// Verify config was created
			if !fileExists(configPath) {
				t.Error("Config file should have been created")
			}
		})
	}
}

// TestImporter_Import_ContextCancellation tests context cancellation
func TestImporter_Import_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	backupPath := filepath.Join(test.GetTestConfigDir(t), "backup.json")

	// Create backup
	installer := NewInstaller(config)
	_, _ = installer.Install(context.Background(), InstallOptions{
		Schema: schematest.NewManagedSsoSingle(),
	}, task.NoOpReporter{})

	exporter := NewExporter(config)
	_, _ = exporter.Export(context.Background(), ExportOptions{
		OutputPath:     backupPath,
		IncludeManaged: true,
	}, task.NoOpReporter{})

	// Cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	importer := NewImporter(config)
	_, err := importer.Import(ctx, ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Should return error when context is cancelled")
	}
}

// TestImporter_Import_DuplicateDetection tests duplicate personal profile detection
func TestImporter_Import_DuplicateDetection(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath
	backupPath := filepath.Join(test.GetTestConfigDir(t), "dup-backup.json")

	// Create config with existing personal profile BELOW managed section
	existingContent := `# START
# END

[profile my-personal]
region = us-east-1
`
	_ = os.WriteFile(configPath, []byte(existingContent), 0644)

	// Create backup with same personal profile
	testSchema := schematest.NewUnmanagedGenericSingle()

	// Ensure structure exists before accessing (Below, not Above)
	if testSchema.Unmanaged == nil || testSchema.Unmanaged.Below == nil ||
		len(testSchema.Unmanaged.Below.GenericProfiles) == 0 {
		t.Fatal("Test schema should have unmanaged below generic profiles")
	}

	testSchema.Unmanaged.Below.GenericProfiles[0].ProfileName = "my-personal" // Duplicate!

	backupData, _ := json.Marshal(testSchema)
	_ = os.WriteFile(backupPath, backupData, 0644)

	// Import - should detect duplicate
	importer := NewImporter(config)
	result, err := importer.Import(context.Background(), ImportOptions{
		BackupPath:   backupPath,
		IncludeBelow: true, // Changed to Below
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Should have detected duplicate
	totalDuplicates := result.ManagedDuplicates.TotalDuplicates + result.UnmanagedAboveDuplicates.TotalDuplicates + result.UnmanagedBelowDuplicates.TotalDuplicates
	if totalDuplicates == 0 {
		t.Error("Should have detected duplicate profile")
	}
}
