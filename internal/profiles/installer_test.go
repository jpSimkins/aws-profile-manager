package profiles

import (
	"context"
	"os"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestNewInstaller tests constructor
func TestNewInstaller(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)

	if installer == nil {
		t.Fatal("NewInstaller should not return nil")
	}
}

// TestInstaller_Install tests basic installation
func TestInstaller_Install(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	result, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.TotalProfiles == 0 {
		t.Error("Expected profiles to be installed")
	}

	// Verify config file was created
	if _, err := os.Stat(config.ConfigPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

// TestInstaller_Install_DryRun tests dry run mode
func TestInstaller_Install_DryRun(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	result, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
		DryRun: true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// In dry run, profiles should be counted but not written
	if result.TotalProfiles == 0 {
		t.Error("Expected profile count in dry run")
	}

	// Verify config file was NOT created
	if _, err := os.Stat(config.ConfigPath); !os.IsNotExist(err) {
		t.Error("Config file should not be created in dry run")
	}
}

// TestInstaller_Install_WithCheatSheet tests cheat sheet generation
func TestInstaller_Install_WithCheatSheet(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	result, err := installer.Install(context.Background(), InstallOptions{
		Schema:             schema,
		GenerateCheatSheet: true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.CheatSheetPath == "" {
		t.Error("Expected cheat sheet path to be set")
	}

	// Verify cheat sheet file exists
	if _, err := os.Stat(result.CheatSheetPath); os.IsNotExist(err) {
		t.Errorf("Cheat sheet file was not created at %s", result.CheatSheetPath)
	}

	// Verify it has content
	content, err := os.ReadFile(result.CheatSheetPath)
	if err != nil {
		t.Fatalf("Failed to read cheat sheet: %v", err)
	}
	if len(content) == 0 {
		t.Error("Cheat sheet file is empty")
	}
}

// TestInstaller_Install_CheatSheetOnly tests cheat-sheet-only mode.
func TestInstaller_Install_CheatSheetOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	result, err := installer.Install(context.Background(), InstallOptions{
		Schema:             schema,
		GenerateCheatSheet: true,
		CheatSheetOnly:     true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if result.CheatSheetPath == "" {
		t.Fatal("Expected cheat sheet path to be set")
	}

	if result.TotalProfiles == 0 {
		t.Error("Expected profile counts to be reported in cheat-sheet-only mode")
	}

	if _, err := os.Stat(result.CheatSheetPath); os.IsNotExist(err) {
		t.Fatalf("Cheat sheet file was not created at %s", result.CheatSheetPath)
	}

	if _, err := os.Stat(config.ConfigPath); !os.IsNotExist(err) {
		t.Error("Config file should not be created in cheat-sheet-only mode")
	}
}

// TestInstaller_Install_NilSchema tests nil schema error
func TestInstaller_Install_NilSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)

	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: nil,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Expected error for nil schema")
	}
}

// TestInstaller_Install_EmptySchema tests empty schema error
func TestInstaller_Install_EmptySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewEmpty()

	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Expected error for empty schema")
	}
}

// TestInstaller_Install_ContextCancellation tests context cancellation
func TestInstaller_Install_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := installer.Install(ctx, InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

// TestInstaller_Install_AllProfileTypes tests all profile types
func TestInstaller_Install_AllProfileTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name   string
		schema *schema.Schema
	}{
		{"SSO", schematest.NewManagedSsoSingle()},
		{"IAM", schematest.NewManagedIamSingle()},
		{"AssumeRole", schematest.NewManagedAssumeRoleSingle()},
		{"Generic", schematest.NewManagedGenericSingle()},
		{"All", schematest.NewManagedAll()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := newTestConfig(t)

			installer := NewInstaller(config)

			result, err := installer.Install(context.Background(), InstallOptions{
				Schema: tt.schema,
			}, task.NoOpReporter{})

			if err != nil {
				t.Fatalf("Install() error = %v", err)
			}

			if result.TotalProfiles == 0 {
				t.Error("Expected profiles to be installed")
			}

			// Verify file exists and has content
			content, err := os.ReadFile(config.ConfigPath)
			if err != nil {
				t.Fatalf("Failed to read config: %v", err)
			}

			if len(content) == 0 {
				t.Error("Config file is empty")
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, "# START") {
				t.Error("Missing start marker")
			}
			if !strings.Contains(contentStr, "# END") {
				t.Error("Missing end marker")
			}
		})
	}
}

// TestInstaller_Install_PreservesPersonal tests personal profile preservation
func TestInstaller_Install_PreservesPersonal(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)

	// Create existing config with personal profiles
	existingConfig := `[profile my-personal-dev]
region = us-east-1
output = json

[profile my-personal-prod]
region = us-west-2
`

	if err := os.WriteFile(configPath, []byte(existingConfig), 0600); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	config := newTestConfig(t)

	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()

	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	// Read config and verify personal profiles are preserved
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[profile my-personal-dev]") {
		t.Error("Personal profile my-personal-dev was not preserved")
	}
	if !strings.Contains(contentStr, "[profile my-personal-prod]") {
		t.Error("Personal profile my-personal-prod was not preserved")
	}
}

// TestInstaller_Install_WithPreFilteredSchema tests install with filtered schema
func TestInstaller_Install_WithPreFilteredSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Use multi-account schema and filter to specific account
	schema := schematest.NewManagedSsoMultiAccount()

	installer := NewInstaller(config)
	result, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
		// Note: Filtering happens in schema.FilterSchema, not InstallOptions
		// This test ensures the installer works with pre-filtered schemas
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install with filtered schema failed: %v", err)
	}

	if result.TotalProfiles == 0 {
		t.Error("Should have installed profiles from filtered schema")
	}
}
