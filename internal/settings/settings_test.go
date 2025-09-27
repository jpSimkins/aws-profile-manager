package settings

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestGet(t *testing.T) {
	// Reset global state
	current = nil

	// Get should return defaults when not set
	s := Get()
	if s == nil {
		t.Fatal("Get() returned nil")
	}
	if s.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", s.Version)
	}
}

func TestSet(t *testing.T) {
	// Reset global state
	current = nil

	// Create custom settings
	custom := &Settings{
		Version: "1.0",
		Application: ApplicationSettings{
			ManagedSectionStart: "CUSTOM START",
			ManagedSectionEnd:   "CUSTOM END",
		},
		Logging: GetDefaultLogging(),
		GUI:     GetDefaultGUI(),
		Sync:    GetDefaultSync(),
		AwsCLI:  GetDefaultAwsCLI(),
	}

	// Set and verify
	if err := Set(custom); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}
	retrieved := Get()
	if retrieved.Application.ManagedSectionStart != "CUSTOM START" {
		t.Errorf("Expected CUSTOM START, got %s", retrieved.Application.ManagedSectionStart)
	}
}

func TestSet_RejectsInvalidSettings(t *testing.T) {
	// Reset global state
	current = nil

	tests := []struct {
		name     string
		settings *Settings
		wantErr  bool
		errMsg   string
	}{
		{
			name: "empty managed section start marker",
			settings: &Settings{
				Version: "1.0",
				Application: ApplicationSettings{
					ManagedSectionStart: "", // Invalid!
					ManagedSectionEnd:   "END",
				},
				Logging: GetDefaultLogging(),
				GUI:     GetDefaultGUI(),
				Sync:    GetDefaultSync(),
				AwsCLI:  GetDefaultAwsCLI(),
			},
			wantErr: true,
			errMsg:  "managed_section_start cannot be empty",
		},
		{
			name: "sync enabled without required fields",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging:     GetDefaultLogging(),
				GUI:         GetDefaultGUI(),
				Sync: SyncSettings{
					Enabled:  true,
					Strategy: "local",
					Local:    LocalSettings{Path: ""}, // Missing required field!
				},
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
			errMsg:  "local.path is required",
		},
		{
			name: "sync s3 enabled without bucket",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging:     GetDefaultLogging(),
				GUI:         GetDefaultGUI(),
				Sync: SyncSettings{
					Enabled:  true,
					Strategy: "s3",
					S3: S3Settings{
						Bucket: "", // Missing required field!
						Key:    "config.json",
						Region: "us-east-1",
					},
				},
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
			errMsg:  "s3.bucket is required",
		},
		{
			name: "invalid log level",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging: LoggingSettings{
					LogLevel:    "invalid-level", // Invalid!
					EnableDebug: false,
				},
				GUI:    GetDefaultGUI(),
				Sync:   GetDefaultSync(),
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
			errMsg:  "log level",
		},
		{
			name:     "valid settings should succeed",
			settings: GetDefaults(),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Set(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Set() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	// Check version
	if defaults.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", defaults.Version)
	}

	// Check application defaults
	if defaults.Application.ManagedSectionStart == "" {
		t.Error("Application ManagedSectionStart should not be empty")
	}

	// Check logging defaults
	if defaults.Logging.LogLevel != "warn" {
		t.Errorf("Expected log level warn, got %s", defaults.Logging.LogLevel)
	}

	// Check GUI defaults
	if defaults.GUI.WindowWidth != 1024 {
		t.Errorf("Expected window width 1024, got %d", defaults.GUI.WindowWidth)
	}

	// Check sync defaults
	if defaults.Sync.Enabled {
		t.Error("Sync should be disabled by default")
	}

	// Check AWS CLI defaults
	if !defaults.AwsCLI.AutoRefresh {
		t.Error("AWS CLI auto-refresh should be enabled by default")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings *Settings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaults(),
			wantErr:  false,
		},
		{
			name: "empty managed section start",
			settings: &Settings{
				Version: "1.0",
				Application: ApplicationSettings{
					ManagedSectionStart: "",
					ManagedSectionEnd:   "END",
				},
				Logging: GetDefaultLogging(),
				GUI:     GetDefaultGUI(),
				Sync:    GetDefaultSync(),
				AwsCLI:  GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging: LoggingSettings{
					LogLevel:    "invalid",
					EnableDebug: false,
				},
				GUI:    GetDefaultGUI(),
				Sync:   GetDefaultSync(),
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "window too small",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging:     GetDefaultLogging(),
				GUI: GUISettings{
					Theme:        "System",
					WindowWidth:  500, // Too small
					WindowHeight: 768,
					ShowSidebar:  true,
					DialogWidth:  600,
					DialogHeight: 500,
				},
				Sync:   GetDefaultSync(),
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadSave(t *testing.T) {
	// Setup test environment
	test.SetupTestEnvironment(t)
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}
	settingsPath := filepath.Join(configDir, "settings.json")

	// Create custom settings
	original := GetDefaults()
	original.Application.ManagedSectionStart = "TEST START"
	original.GUI.Theme = "Dark"

	// Set and save
	if err := Set(original); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}
	if err := Save(settingsPath); err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Fatal("Settings file was not created")
	}

	// Reset global state
	current = nil

	// Load settings
	if err := Load(settingsPath); err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}

	// Verify loaded settings
	loaded := Get()
	if loaded.Application.ManagedSectionStart != "TEST START" {
		t.Errorf("Expected TEST START, got %s", loaded.Application.ManagedSectionStart)
	}
	if loaded.GUI.Theme != "Dark" {
		t.Errorf("Expected Dark theme, got %s", loaded.GUI.Theme)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	// Setup test environment
	test.SetupTestEnvironment(t)
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}
	settingsPath := filepath.Join(configDir, "nonexistent.json")

	// Reset global state
	current = nil

	// Load should create default settings file
	if err := Load(settingsPath); err != nil {
		t.Fatalf("Load should create file with defaults, got error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file should have been created")
	}

	// Verify defaults were loaded
	loaded := Get()
	if loaded.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", loaded.Version)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Setup test environment
	test.SetupTestEnvironment(t)
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("Failed to get config dir: %v", err)
	}
	settingsPath := filepath.Join(configDir, "invalid.json")

	// Write invalid JSON
	_ = os.WriteFile(settingsPath, []byte("not valid json{"), 0644)

	// Load should fail
	if err := Load(settingsPath); err == nil {
		t.Error("Load should fail with invalid JSON")
	}
}

func TestGetAllSchemas(t *testing.T) {
	s := GetDefaults()
	schemas := s.GetAllSchemas()

	expectedSchemas := []string{"application", "logging", "gui", "sync", "awscli"}
	for _, name := range expectedSchemas {
		if _, exists := schemas[name]; !exists {
			t.Errorf("Expected schema %s not found", name)
		}
	}

	// Verify schema has required fields
	guiSchema := schemas["gui"]
	if guiSchema.Version != "1.0" {
		t.Errorf("Expected GUI schema version 1.0, got %s", guiSchema.Version)
	}
	if len(guiSchema.Fields) == 0 {
		t.Error("GUI schema should have fields")
	}
}

func TestThreadSafety(t *testing.T) {
	// Reset global state
	current = nil

	// Test concurrent Get/Set operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			s := GetDefaults()
			s.Application.ManagedSectionStart = "CONCURRENT TEST"
			_ = Set(s) // Ignore error in concurrent test
			retrieved := Get()
			if retrieved == nil {
				t.Error("Get() returned nil in concurrent test")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is consistent
	final := Get()
	if final == nil {
		t.Error("Final state is nil")
	}
}

// TestSet_InvalidSettings tests that Set() rejects invalid settings
func TestSet_InvalidSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	testCases := []struct {
		name     string
		settings *Settings
		wantErr  bool
	}{
		{
			name: "empty managed section start marker",
			settings: &Settings{
				Version: "1.0",
				Application: ApplicationSettings{
					ManagedSectionStart: "", // Empty - invalid
					ManagedSectionEnd:   "END",
				},
				Logging: GetDefaultLogging(),
				GUI:     GetDefaultGUI(),
				Sync:    GetDefaultSync(),
				AwsCLI:  GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "same start and end markers",
			settings: &Settings{
				Version: "1.0",
				Application: ApplicationSettings{
					ManagedSectionStart: "SAME",
					ManagedSectionEnd:   "SAME", // Same as start - invalid
				},
				Logging: GetDefaultLogging(),
				GUI:     GetDefaultGUI(),
				Sync:    GetDefaultSync(),
				AwsCLI:  GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging: LoggingSettings{
					LogLevel: "invalid-level", // Invalid log level
				},
				GUI:    GetDefaultGUI(),
				Sync:   GetDefaultSync(),
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "invalid sync strategy",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging:     GetDefaultLogging(),
				GUI:         GetDefaultGUI(),
				Sync: SyncSettings{
					Strategy: "invalid-strategy", // Invalid strategy
				},
				AwsCLI: GetDefaultAwsCLI(),
			},
			wantErr: true,
		},
		{
			name: "valid settings",
			settings: &Settings{
				Version:     "1.0",
				Application: GetDefaultApplication(),
				Logging:     GetDefaultLogging(),
				GUI:         GetDefaultGUI(),
				Sync:        GetDefaultSync(),
				AwsCLI:      GetDefaultAwsCLI(),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Set(tc.settings)

			if tc.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tc.wantErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestSet_ValidSettingsAccepted tests that valid settings are accepted
func TestSet_ValidSettingsAccepted(t *testing.T) {
	test.SetupTestEnvironment(t)

	validSettings := GetDefaults()

	// Should not return error
	if err := Set(validSettings); err != nil {
		t.Errorf("Valid settings should be accepted, got error: %v", err)
	}

	// Verify settings were actually set
	retrieved := Get()
	if retrieved.Version != validSettings.Version {
		t.Error("Settings were not properly set")
	}
}
