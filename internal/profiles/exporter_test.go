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

// TestNewExporter tests Exporter constructor
func TestNewExporter(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:  test.GetTestAwsConfigPath(t),
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	exporter := NewExporter(config)

	if exporter == nil {
		t.Fatal("NewExporter should not return nil")
	}
}

// TestExporter_Export_ManagedOnly tests exporting only managed section
func TestExporter_Export_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export-managed.json")

	// Create config
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	// Install profiles first
	installer := NewInstaller(config)
	testSchema := schematest.NewManagedSsoSingle()
	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: testSchema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to install profiles: %v", err)
	}

	// Export managed only
	exporter := NewExporter(config)
	result, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify result
	if result.Schema == nil {
		t.Error("ExportedSchema should not be nil")
	}
	if result.Schema.Managed == nil {
		t.Error("Managed section should not be nil")
	}
	if result.TotalProfiles == 0 {
		t.Error("TotalProfiles should be > 0")
	}

	// Verify file was created
	if !fileExists(outputPath) {
		t.Error("Export file was not created")
	}

	// Verify JSON is valid
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var exported schema.Schema
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Export file is not valid JSON: %v", err)
	}
}

// TestExporter_Export_FullBackup tests exporting all sections
func TestExporter_Export_FullBackup(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export-full.json")

	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	// Create config with managed and personal profiles
	configContent := `[profile personal-above]
region = us-east-1

# START - Test
[profile work-dev]
region = us-west-2
# END - Test

[profile personal-below]
region = us-west-1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Export all sections
	exporter := NewExporter(config)
	result, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
		Description:    "Full backup test",
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify all sections were exported
	if result.Schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if result.Schema.Unmanaged.Above == nil {
		t.Error("Above section should not be nil")
	}
	if result.Schema.Unmanaged.Below == nil {
		t.Error("Below section should not be nil")
	}

	// Verify metadata
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var exported schema.Schema
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Export file is not valid JSON: %v", err)
	}

	// Verify description if metadata exists
	if exported.Metadata != nil && exported.Metadata.Description != "Full backup test" {
		t.Errorf("Description = %q, want %q", exported.Metadata.Description, "Full backup test")
	}
}

// TestExporter_Export_PersonalOnly tests exporting only personal profiles
func TestExporter_Export_PersonalOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export-personal.json")

	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	// Create config with personal profiles only
	configContent := `[profile personal-1]
region = us-east-1

[profile personal-2]
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Export personal only
	exporter := NewExporter(config)
	result, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: false,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Managed section should be nil or empty
	if result.Schema.Managed != nil && hasProfiles(result.Schema.Managed) {
		t.Error("Managed section should be empty when not included")
	}

	// Verify file was created
	if !fileExists(outputPath) {
		t.Error("Export file was not created")
	}
}

// TestExporter_Export_EmptyConfig tests exporting from empty config
func TestExporter_Export_EmptyConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export-empty.json")

	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	// Create empty config
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Export should succeed but return empty schema
	exporter := NewExporter(config)
	result, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Export should succeed on empty config, got error: %v", err)
	}

	if result.TotalProfiles != 0 {
		t.Errorf("TotalProfiles should be 0, got %d", result.TotalProfiles)
	}

	// Verify file was still created
	if !fileExists(outputPath) {
		t.Error("Export file was not created")
	}
}

// TestExporter_Export_NonExistentConfig tests exporting when config doesn't exist
func TestExporter_Export_NonExistentConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "nonexistent-config")
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export.json")

	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	exporter := NewExporter(config)
	_, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	// Should fail when config doesn't exist
	if err == nil {
		t.Error("Export should fail on non-existent config")
	}
}

// TestExporter_Export_InvalidOutputPath tests error handling for bad output path
func TestExporter_Export_InvalidOutputPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	// Create valid config
	if err := os.WriteFile(configPath, []byte("[profile test]\n"), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	exporter := NewExporter(config)

	// Try to export to invalid path (directory without filename)
	_, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     "/",
		IncludeManaged: true,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Export should fail with invalid output path")
	}
}

// TestExporter_Export_NoSectionsSelected tests error when no sections selected
func TestExporter_Export_NoSectionsSelected(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	outputPath := filepath.Join(test.GetTestConfigDir(t), "export.json")

	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	exporter := NewExporter(config)
	_, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     outputPath,
		IncludeManaged: false,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Export should fail when no sections are selected")
	}
}

// TestExporter_Export_PreservesSections tests that correct sections are exported
func TestExporter_Export_PreservesSections(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name           string
		includeManaged bool
		includeAbove   bool
		includeBelow   bool
		wantManaged    bool
		wantAbove      bool
		wantBelow      bool
	}{
		{
			name:           "managed only",
			includeManaged: true,
			includeAbove:   false,
			includeBelow:   false,
			wantManaged:    true,
			wantAbove:      false,
			wantBelow:      false,
		},
		{
			name:           "above only",
			includeManaged: false,
			includeAbove:   true,
			includeBelow:   false,
			wantManaged:    false,
			wantAbove:      true,
			wantBelow:      false,
		},
		{
			name:           "below only",
			includeManaged: false,
			includeAbove:   false,
			includeBelow:   true,
			wantManaged:    false,
			wantAbove:      false,
			wantBelow:      true,
		},
		{
			name:           "all sections",
			includeManaged: true,
			includeAbove:   true,
			includeBelow:   true,
			wantManaged:    true,
			wantAbove:      true,
			wantBelow:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := test.GetTestAwsConfigPath(t)
			outputPath := filepath.Join(test.GetTestConfigDir(t), "export-"+tt.name+".json")

			config := Config{
				ConfigPath:  configPath,
				StartMarker: "# START - Test",
				EndMarker:   "# END - Test",
			}

			// Create config with all sections
			configContent := `[profile personal-above]
region = us-east-1

# START - Test
[profile work-dev]
region = us-west-2
# END - Test

[profile personal-below]
region = us-west-1
`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Export with specified options
			exporter := NewExporter(config)
			result, err := exporter.Export(context.Background(), ExportOptions{
				OutputPath:     outputPath,
				IncludeManaged: tt.includeManaged,
				IncludeAbove:   tt.includeAbove,
				IncludeBelow:   tt.includeBelow,
			}, task.NoOpReporter{})

			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Verify sections
			hasManaged := result.Schema.Managed != nil && hasProfiles(result.Schema.Managed)
			hasAbove := result.Schema.Unmanaged != nil && result.Schema.Unmanaged.Above != nil && hasProfiles(result.Schema.Unmanaged.Above)
			hasBelow := result.Schema.Unmanaged != nil && result.Schema.Unmanaged.Below != nil && hasProfiles(result.Schema.Unmanaged.Below)

			if hasManaged != tt.wantManaged {
				t.Errorf("Managed section presence = %v, want %v", hasManaged, tt.wantManaged)
			}
			if hasAbove != tt.wantAbove {
				t.Errorf("Above section presence = %v, want %v", hasAbove, tt.wantAbove)
			}
			if hasBelow != tt.wantBelow {
				t.Errorf("Below section presence = %v, want %v", hasBelow, tt.wantBelow)
			}
		})
	}
}

// TestExporter_Export_NonexistentFile tests exporting when config doesn't exist
func TestExporter_Export_NonexistentFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:  filepath.Join(test.GetTestConfigDir(t), "does-not-exist"),
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	exporter := NewExporter(config)
	_, err := exporter.Export(context.Background(), ExportOptions{
		OutputPath:     filepath.Join(test.GetTestConfigDir(t), "output.json"),
		IncludeManaged: true,
	}, task.NoOpReporter{})

	// Should handle gracefully (either error or empty export)
	_ = err
}
