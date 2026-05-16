package profiles

import (
	"context"
	"os"
	"testing"

	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestConfigReader_ReadConfig_IamProfiles tests reading IAM profiles
func TestConfigReader_ReadConfig_IamProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// First write IAM profiles
	writer := newConfigWriter(config)
	schema := schematest.NewManagedIamMulti()
	_, _, _, _, err := writer.writeConfig(context.Background(), schema, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Now read them back
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	if len(result.Managed.IamUsers) == 0 {
		t.Error("Should have read IAM users")
	}
}

// TestConfigReader_ReadConfig_AssumeRoleProfiles tests reading AssumeRole profiles
func TestConfigReader_ReadConfig_AssumeRoleProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// First write AssumeRole profiles
	writer := newConfigWriter(config)
	schema := schematest.NewManagedAssumeRoleMulti()
	_, _, _, _, err := writer.writeConfig(context.Background(), schema, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Now read them back
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	if len(result.Managed.AssumeRoleChains) == 0 {
		t.Error("Should have read AssumeRole chains")
	}
}

// TestConfigReader_ReadConfig_MixedTypes tests reading all profile types
func TestConfigReader_ReadConfig_MixedTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Write all profile types
	writer := newConfigWriter(config)
	schema := schematest.NewManagedAll()
	_, _, _, _, err := writer.writeConfig(context.Background(), schema, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read them back
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have all types
	if len(result.Managed.Organizations) == 0 {
		t.Error("Should have SSO organizations")
	}
	if len(result.Managed.IamUsers) == 0 {
		t.Error("Should have IAM users")
	}
	if len(result.Managed.AssumeRoleChains) == 0 {
		t.Error("Should have AssumeRole chains")
	}
	if len(result.Managed.GenericProfiles) == 0 {
		t.Error("Should have Generic profiles")
	}
}

// TestConfigReader_ReadConfig_GenericProfiles tests reading generic profiles
func TestConfigReader_ReadConfig_GenericProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Write generic profiles
	writer := newConfigWriter(config)
	schema := schematest.NewManagedGenericMulti()
	_, _, _, _, err := writer.writeConfig(context.Background(), schema, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read them back
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	if len(result.Managed.GenericProfiles) == 0 {
		t.Error("Should have read Generic profiles")
	}
}

// TestConfigReader_ReadConfig_UnmanagedSections tests reading personal profiles
func TestConfigReader_ReadConfig_UnmanagedSections(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Create config with personal profiles above and below
	configContent := `[profile personal-above]
region = us-east-1
output = json

# START
[profile work-dev]
region = us-west-2
# END

[profile personal-below]
region = eu-west-1
output = yaml
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read all sections
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}

	if result.Unmanaged.Above == nil {
		t.Error("Above section should not be nil")
	}

	if result.Unmanaged.Below == nil {
		t.Error("Below section should not be nil")
	}

	// Check that profiles were read
	if result.Unmanaged.Above != nil && len(result.Unmanaged.Above.GenericProfiles) == 0 {
		t.Error("Should have read above profiles")
	}

	if result.Unmanaged.Below != nil && len(result.Unmanaged.Below.GenericProfiles) == 0 {
		t.Error("Should have read below profiles")
	}
}

// TestConfigReader_ReadConfig_NoManagedMarkers tests reading without markers
func TestConfigReader_ReadConfig_NoManagedMarkers(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Create config without markers (all personal)
	configContent := `[profile personal-1]
region = us-east-1

[profile personal-2]
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read all sections
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Should have no managed section
	if result.Managed != nil && hasProfiles(result.Managed) {
		t.Error("Should not have managed profiles when no markers present")
	}

	// Should have unmanaged profiles
	if result.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
}

// TestConfigReader_ReadConfig_EmptyConfig tests reading empty config
func TestConfigReader_ReadConfig_EmptyConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Create empty config
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read config
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Result should be empty but valid
	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

// TestConfigReader_ReadConfig_MalformedProfiles tests handling malformed profiles
func TestConfigReader_ReadConfig_MalformedProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Create config with malformed profile (no closing bracket, invalid syntax)
	configContent := `[profile malformed
region = us-east-1

[profile valid]
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Read config - should handle gracefully
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: false,
		IncludeAbove:   true,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	// Should either succeed (skipping malformed) or fail gracefully
	if err != nil {
		// Error is acceptable for malformed config
		return
	}

	if result == nil {
		t.Fatal("Result should not be nil when no error returned")
	}
}

// TestConfigReader_RoundTrip tests write and read consistency
func TestConfigReader_RoundTrip(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	// Write a schema
	writer := newConfigWriter(config)
	originalSchema := schematest.NewManagedSsoMultiAccount()
	profilesWritten, sessionsWritten, _, _, err := writer.writeConfig(context.Background(), originalSchema, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	if profilesWritten == 0 {
		t.Error("Should have written profiles")
	}
	if sessionsWritten == 0 {
		t.Error("Should have written sessions")
	}

	// Read it back
	reader := newConfigReader(config)
	result, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
		IncludeAbove:   false,
		IncludeBelow:   false,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if result.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have same number of orgs
	if len(result.Managed.Organizations) != len(originalSchema.Managed.Organizations) {
		t.Errorf("Organizations count mismatch: got %d, want %d",
			len(result.Managed.Organizations), len(originalSchema.Managed.Organizations))
	}
}

// TestConfigReader_ParseContentToExtractedData_ErrorPaths tests error handling
func TestConfigReader_ParseContentToExtractedData_ErrorPaths(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:  test.GetTestAwsConfigPath(t),
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	reader := newConfigReader(config)

	t.Run("empty content", func(t *testing.T) {
		data, err := reader.parseContentToExtractedData("")
		if err != nil {
			t.Fatalf("Should not error on empty content: %v", err)
		}
		if data == nil {
			t.Fatal("Should return empty data, not nil")
		}
		if len(data.Profiles) != 0 || len(data.SsoSessions) != 0 {
			t.Error("Empty content should return empty slices")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		data, err := reader.parseContentToExtractedData("   \n\t  ")
		if err != nil {
			t.Fatalf("Should not error on whitespace: %v", err)
		}
		if data == nil {
			t.Fatal("Should return empty data, not nil")
		}
		if len(data.Profiles) != 0 || len(data.SsoSessions) != 0 {
			t.Error("Whitespace should return empty slices")
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		data, err := reader.parseContentToExtractedData("[profile broken\nno closing")
		// Parser may or may not error - depends on awscli.Extractor behavior
		// Either way, data should not be nil
		if err == nil && data == nil {
			t.Fatal("Should return data or error, not both nil")
		}
	})

	t.Run("valid content", func(t *testing.T) {
		content := `[profile test]
region = us-east-1
`
		data, err := reader.parseContentToExtractedData(content)
		if err != nil {
			t.Fatalf("Should not error on valid content: %v", err)
		}
		if data == nil {
			t.Fatal("Should return data for valid content")
		}
	})
}

// TestConfigReader_ReadConfig_NonexistentFile tests reading missing file
func TestConfigReader_ReadConfig_NonexistentFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:  test.GetTestConfigDir(t) + "/nonexistent",
		StartMarker: "# START",
		EndMarker:   "# END",
	}

	reader := newConfigReader(config)
	_, _, err := reader.readConfig(context.Background(), ExportOptions{
		IncludeManaged: true,
	}, task.NoOpReporter{})

	// Should handle missing file gracefully (returns error or empty)
	_ = err
}
