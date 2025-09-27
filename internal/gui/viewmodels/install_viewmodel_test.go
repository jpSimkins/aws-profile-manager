package viewmodels

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/settings"
	syncpkg "aws-profile-manager/internal/sync"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

func TestNewInstallViewModel(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()

	if vm == nil {
		t.Fatal("ViewModel should not be nil")
	}

	if vm.IsLoading {
		t.Error("IsLoading should be false initially")
	}

	if vm.GetSelectedPreset() != "" {
		t.Error("Selected preset should be empty initially")
	}
}

func TestInstallViewModel_ConfigPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()
	cachePath, err := vm.GetSyncCachePath()

	if err != nil {
		t.Fatalf("GetSyncCachePath failed: %v", err)
	}

	if cachePath == "" {
		t.Error("Cache path should not be empty")
	}

	// Should end with sync-config.json
	if !strings.HasSuffix(cachePath, "sync-config.json") {
		t.Errorf("Cache path should end with 'sync-config.json', got: %s", cachePath)
	}
}

func TestInstallViewModel_FormatLoadedAt(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()

	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	formatted := vm.FormatLoadedAt(testTime)

	expected := "2024-01-15 14:30:45"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}
}

func TestInstallViewModel_EmptyDisplaySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()
	emptySchema := vm.EmptyDisplaySchema()

	if emptySchema == nil {
		t.Fatal("EmptyDisplaySchema should not be nil")
	}

	if emptySchema.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", emptySchema.Version)
	}

	if emptySchema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	if len(emptySchema.Managed.Organizations) != 0 {
		t.Error("Organizations should be empty")
	}
}

func TestInstallViewModel_PresetSelection(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()

	// Initially no preset selected
	if vm.GetSelectedPreset() != "" {
		t.Error("Initial preset should be empty")
	}

	// Select a preset
	vm.SelectPreset("developer")
	if vm.GetSelectedPreset() != "developer" {
		t.Errorf("Expected 'developer', got '%s'", vm.GetSelectedPreset())
	}

	// Change preset
	vm.SelectPreset("devsecops")
	if vm.GetSelectedPreset() != "devsecops" {
		t.Errorf("Expected 'devsecops', got '%s'", vm.GetSelectedPreset())
	}

	// Clear preset
	vm.SelectPreset("")
	if vm.GetSelectedPreset() != "" {
		t.Error("Preset should be cleared")
	}
}

func TestInstallViewModel_GetPresets(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()

	tests := []struct {
		name           string
		schema         *schema.Schema
		expectedCount  int
		shouldBeNil    bool
		expectedPreset string
	}{
		{
			name:        "Nil schema",
			schema:      nil,
			shouldBeNil: true,
		},
		{
			name:        "Schema without presets",
			schema:      schematest.NewManagedSsoSingle(),
			shouldBeNil: true,
		},
		{
			name:           "Schema with presets",
			schema:         schematest.NewManagedSsoWithPresets(),
			expectedCount:  4,
			expectedPreset: "developer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presets := vm.GetPresets(tt.schema)

			if tt.shouldBeNil {
				if presets != nil {
					t.Error("Expected nil presets")
				}
				return
			}

			if presets == nil {
				t.Fatal("Expected presets, got nil")
			}

			if len(presets) != tt.expectedCount {
				t.Errorf("Expected %d presets, got %d", tt.expectedCount, len(presets))
			}

			if tt.expectedPreset != "" {
				if _, exists := presets[tt.expectedPreset]; !exists {
					t.Errorf("Expected preset '%s' to exist", tt.expectedPreset)
				}
			}
		})
	}
}

func TestInstallViewModel_LoadDisplaySchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create test schema
	testSchema := schematest.NewManagedSsoSingle()

	// Write schema to cache file (not AWS config)
	cacheDir, _ := settings.GetCacheDir()
	cachePath := filepath.Join(cacheDir, "sync-config.json")

	// Create cache entry
	cacheEntry := syncpkg.CacheEntry{
		Data:      testSchema,
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  syncpkg.StrategyLocal,
	}

	// Marshal and write cache file
	data, err := json.MarshalIndent(cacheEntry, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal cache entry: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}

	// Now test loading
	vm := NewInstallViewModel()
	ctx := context.Background()
	displaySchema, sourcePath, loadErr := vm.LoadDisplaySchema(ctx, task.NoOpReporter{})

	if loadErr != nil {
		t.Fatalf("LoadDisplaySchema failed: %v", loadErr)
	}

	if displaySchema == nil {
		t.Fatal("displaySchema should not be nil")
	}

	if sourcePath == "" {
		t.Error("sourcePath should not be empty")
	}

	// Verify source is cache path
	if !strings.HasSuffix(sourcePath, "sync-config.json") {
		t.Errorf("Expected source to be cache file, got: %s", sourcePath)
	}

	// Verify managed section loaded
	if displaySchema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	if len(displaySchema.Managed.Organizations) == 0 {
		t.Error("Should have loaded organizations")
	}
}

func TestInstallViewModel_StartLoad(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	// Create test schema
	testSchema := schematest.NewManagedSsoSingle()

	// Write schema to cache file (not AWS config)
	cacheDir, _ := settings.GetCacheDir()
	cachePath := filepath.Join(cacheDir, "sync-config.json")

	// Create cache entry
	cacheEntry := syncpkg.CacheEntry{
		Data:      testSchema,
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  syncpkg.StrategyLocal,
	}

	// Marshal and write cache file
	data, err := json.MarshalIndent(cacheEntry, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal cache entry: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}

	// Test StartLoad with callback
	vm := NewInstallViewModel()
	ctx := context.Background()

	done := make(chan bool)
	var callbackSchema *schema.Schema
	var callbackError error

	vm.StartLoad(ctx, task.NoOpReporter{}, func(schema *schema.Schema, path string, err error, loadedAt time.Time) {
		callbackSchema = schema
		callbackError = err
		done <- true
	})

	// Wait for callback
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Callback timeout")
	}

	if callbackError != nil {
		t.Fatalf("Callback received error: %v", callbackError)
	}

	if callbackSchema == nil {
		t.Fatal("Callback schema should not be nil")
	}
}

func TestInstallViewModel_Install(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()
	testSchema := schematest.NewManagedSsoSingle()

	opts := profiles.InstallOptions{
		Schema:             testSchema,
		Organizations:      []string{},
		GenerateCheatSheet: false,
		DryRun:             true, // Dry run to avoid actual file writes
	}

	ctx := context.Background()
	result, err := vm.Install(ctx, testSchema, testSchema, opts, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// In dry-run mode, should still report what would be installed
	if result.TotalProfiles == 0 {
		t.Error("Should report profiles that would be installed")
	}
}

func TestInstallViewModel_StartInstall(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()
	testSchema := schematest.NewManagedSsoSingle()

	opts := profiles.InstallOptions{
		Schema: testSchema,
		DryRun: true,
	}

	ctx := context.Background()
	done := make(chan bool)
	var callbackResult *profiles.InstallResult
	var callbackError error

	vm.StartInstall(ctx, testSchema, testSchema, opts, task.NoOpReporter{}, func(result *profiles.InstallResult, err error) {
		callbackResult = result
		callbackError = err
		done <- true
	})

	// Wait for callback
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Callback timeout")
	}

	if callbackError != nil {
		t.Errorf("Callback received error: %v", callbackError)
	}

	if callbackResult == nil {
		t.Error("Callback result should not be nil")
	}
}

func TestInstallViewModel_ThreadSafety(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	vm := NewInstallViewModel()

	// Concurrent preset selection
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			presetKey := "preset-" + string(rune(id))
			vm.SelectPreset(presetKey)
			_ = vm.GetSelectedPreset()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic (test passes if we get here)
}
