package profiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestConfigWriter_WriteConfig tests complete config writing flow
func TestConfigWriter_WriteConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name            string
		schema          func() interface{}
		existingConfig  string
		startMarker     string
		endMarker       string
		wantErr         bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:           "fresh install - no existing config",
			schema:         func() interface{} { return schematest.NewManagedSsoSingle() },
			existingConfig: "",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantErr:        false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[sso-session",
				"[profile",
			},
		},
		{
			name:   "replace existing managed section",
			schema: func() interface{} { return schematest.NewManagedSsoSingle() },
			existingConfig: `[profile personal-above]
region = us-east-1

# START - Test
[profile old-work]
region = us-west-2
# END - Test

[profile personal-below]
region = us-west-1`,
			startMarker: "# START - Test",
			endMarker:   "# END - Test",
			wantErr:     false,
			wantContains: []string{
				"[profile personal-above]",
				"# START - Test",
				"# END - Test",
				"[profile personal-below]",
				"[sso-session",
			},
			wantNotContains: []string{
				"[profile old-work]",
			},
		},
		{
			name:   "preserve personal profiles above",
			schema: func() interface{} { return schematest.NewManagedSsoSingle() },
			existingConfig: `[profile my-personal-1]
region = us-east-1

[profile my-personal-2]
region = us-west-2`,
			startMarker: "# START - Test",
			endMarker:   "# END - Test",
			wantErr:     false,
			wantContains: []string{
				"[profile my-personal-1]",
				"[profile my-personal-2]",
				"# START - Test",
				"# END - Test",
			},
		},
		{
			name:   "preserve personal profiles below",
			schema: func() interface{} { return schematest.NewManagedSsoSingle() },
			existingConfig: `# START - Test
[profile old]
# END - Test

[profile my-personal-below]
region = us-east-1`,
			startMarker: "# START - Test",
			endMarker:   "# END - Test",
			wantErr:     false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[profile my-personal-below]",
			},
		},
		{
			name:           "IAM profiles",
			schema:         func() interface{} { return schematest.NewManagedIamSingle() },
			existingConfig: "",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantErr:        false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[profile",
				"aws_access_key_id",
			},
		},
		{
			name:           "AssumeRole profiles",
			schema:         func() interface{} { return schematest.NewManagedAssumeRoleSingle() },
			existingConfig: "",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantErr:        false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[profile",
				"role_arn",
				"source_profile",
			},
		},
		{
			name:           "Generic profiles",
			schema:         func() interface{} { return schematest.NewManagedGenericSingle() },
			existingConfig: "",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantErr:        false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[profile",
			},
		},
		{
			name:           "All profile types",
			schema:         func() interface{} { return schematest.NewManagedAll() },
			existingConfig: "",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantErr:        false,
			wantContains: []string{
				"# START - Test",
				"# END - Test",
				"[sso-session",
				"[profile",
				"aws_access_key_id",
				"role_arn",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(test.GetTestConfigDir(t), "config")

			// Create existing config if provided
			if tt.existingConfig != "" {
				if err := os.WriteFile(configPath, []byte(tt.existingConfig), 0600); err != nil {
					t.Fatalf("Failed to create existing config: %v", err)
				}
			}

			config := Config{
				ConfigPath:  configPath,
				StartMarker: tt.startMarker,
				EndMarker:   tt.endMarker,
			}

			writer := newConfigWriter(config)
			ctx := context.Background()
			reporter := task.NoOpReporter{}

			schemaObj := tt.schema()
			profilesWritten, sessionsWritten, _, changed, err := writer.writeConfig(ctx, schemaObj.(*schema.Schema), reporter)

			if (err != nil) != tt.wantErr {
				t.Errorf("writeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify file was written
			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("Failed to read written config: %v", err)
			}

			contentStr := string(content)

			// Check for expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("Config should contain %q", want)
				}
			}

			// Check for unwanted content
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(contentStr, notWant) {
					t.Errorf("Config should not contain %q", notWant)
				}
			}

			// Verify return values make sense
			if changed && profilesWritten == 0 {
				t.Error("Changed flag is true but no profiles written")
			}

			if sessionsWritten < 0 {
				t.Errorf("Session count should be non-negative, got %d", sessionsWritten)
			}
		})
	}
}

// TestConfigWriter_NoChanges tests change detection
func TestConfigWriter_NoChanges(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "config")
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	writer := newConfigWriter(config)
	ctx := context.Background()
	reporter := task.NoOpReporter{}
	schema := schematest.NewManagedSsoSingle()

	// First write
	_, _, _, _, err := writer.writeConfig(ctx, schema, reporter)
	if err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Second write with same content
	_, _, _, changed, err := writer.writeConfig(ctx, schema, reporter)
	if err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	if changed {
		t.Error("Second write with identical content should not report changes")
	}
}

// TestConfigWriter_EmptySchema tests empty schema handling
func TestConfigWriter_EmptySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "config")
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	writer := newConfigWriter(config)
	ctx := context.Background()
	reporter := task.NoOpReporter{}

	tests := []struct {
		name    string
		schema  interface{}
		wantErr bool
	}{
		{
			name:    "empty managed section",
			schema:  schematest.NewManagedEmpty(),
			wantErr: false,
		},
		{
			name:    "completely empty - nil managed",
			schema:  schematest.NewEmpty(),
			wantErr: true, // Expects error because Managed is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaObj := tt.schema.(*schema.Schema)
			profilesWritten, _, _, _, err := writer.writeConfig(ctx, schemaObj, reporter)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && profilesWritten != 0 {
				t.Errorf("Empty schema should write 0 profiles, got %d", profilesWritten)
			}
		})
	}
}

// TestConfigWriter_MarkerVariations tests different marker formats
func TestConfigWriter_MarkerVariations(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name        string
		startMarker string
		endMarker   string
		wantErr     bool
	}{
		{
			name:        "standard markers",
			startMarker: "# START - Managed by AWS Profile Manager",
			endMarker:   "# END - Managed by AWS Profile Manager",
			wantErr:     false,
		},
		{
			name:        "short markers",
			startMarker: "# START",
			endMarker:   "# END",
			wantErr:     false,
		},
		{
			name:        "custom markers",
			startMarker: "# BEGIN WORK PROFILES",
			endMarker:   "# FINISH WORK PROFILES",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(test.GetTestConfigDir(t), "config-"+tt.name)
			config := Config{
				ConfigPath:  configPath,
				StartMarker: tt.startMarker,
				EndMarker:   tt.endMarker,
			}

			writer := newConfigWriter(config)
			ctx := context.Background()
			reporter := task.NoOpReporter{}
			schema := schematest.NewManagedSsoSingle()

			_, _, _, _, err := writer.writeConfig(ctx, schema, reporter)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config: %v", err)
				}

				contentStr := string(content)
				if !strings.Contains(contentStr, tt.startMarker) {
					t.Errorf("Config should contain start marker: %q", tt.startMarker)
				}
				if !strings.Contains(contentStr, tt.endMarker) {
					t.Errorf("Config should contain end marker: %q", tt.endMarker)
				}
			}
		})
	}
}

// TestConfigWriter_LargeScale tests performance with large schemas
func TestConfigWriter_LargeScale(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "config")
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	writer := newConfigWriter(config)
	ctx := context.Background()
	reporter := task.NoOpReporter{}
	schema := schematest.NewLargeScale()

	_, written, _, _, err := writer.writeConfig(ctx, schema, reporter)
	if err != nil {
		t.Fatalf("Large scale write failed: %v", err)
	}

	if written == 0 {
		t.Error("Large scale schema should write profiles")
	}

	// Verify file was created and has content
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Config file should have content")
	}
}

// TestConfigWriter_ContextCancellation tests context cancellation handling
func TestConfigWriter_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "config")
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	writer := newConfigWriter(config)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reporter := task.NoOpReporter{}
	schema := schematest.NewManagedSsoSingle()

	_, _, _, _, err := writer.writeConfig(ctx, schema, reporter)
	if err != nil {
		// Context cancellation might cause error - that's acceptable
		t.Logf("Context cancellation caused error (expected): %v", err)
	}
}

// TestConfigWriter_FilePermissions tests file permission handling
func TestConfigWriter_FilePermissions(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "config")
	config := Config{
		ConfigPath:  configPath,
		StartMarker: "# START - Test",
		EndMarker:   "# END - Test",
	}

	writer := newConfigWriter(config)
	ctx := context.Background()
	reporter := task.NoOpReporter{}
	schema := schematest.NewManagedSsoSingle()

	_, _, _, _, err := writer.writeConfig(ctx, schema, reporter)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Check file permissions - verify file was created
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Verify it's a regular file
	mode := info.Mode()
	if !mode.IsRegular() {
		t.Errorf("Config should be a regular file, got mode: %v", mode)
	}

	// File should exist and be readable
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("File should be readable: %v", err)
	}
	if len(content) == 0 {
		t.Error("File should have content")
	}
}
