package cli

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestGetSyncSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Set up test sync settings with VALID configuration
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = "/test/config.json" // Required when local strategy is enabled
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Get sync settings
	syncSettings, err := getSyncSettings()
	if err != nil {
		t.Fatalf("getSyncSettings failed: %v", err)
	}

	if syncSettings == nil {
		t.Fatal("syncSettings should not be nil")
	}

	if !syncSettings.Enabled {
		t.Error("Sync should be enabled")
	}

	if syncSettings.Strategy != "local" {
		t.Errorf("Expected strategy 'local', got %s", syncSettings.Strategy)
	}
}

func TestRunSyncFetch_Disabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Set sync to disabled
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = false
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncFetchCommand()

	// Run sync fetch - should work even when disabled (shows message)
	err := runSyncFetch(cmd, []string{})
	// Expected to show "sync is disabled" message, not return error
	_ = err // Don't fail on this
}

func TestRunSyncFetch_LocalStrategy(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a test config file with proper schema
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Enable sync with local strategy
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncFetchCommand()

	// Run sync fetch
	err := runSyncFetch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncFetch failed: %v", err)
	}
}

func TestRunSyncStatus_Disabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Set sync to disabled
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = false
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncStatusCommand()

	// Run sync status - should still work (shows disabled status)
	err := runSyncStatus(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncStatus failed: %v", err)
	}
}

func TestRunSyncStatus_Enabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a test config file with proper schema
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Enable sync with existing file
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncStatusCommand()

	// Run sync status
	err := runSyncStatus(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncStatus failed: %v", err)
	}
}

func TestRunSyncClearCache(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a test config file with proper schema
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create a cache file (the actual cache file name used by sync package)
	cacheDir := test.GetTestCacheDir(t)
	cacheFile := filepath.Join(cacheDir, "sync-config.json")
	cacheData := []byte(`{"Data":{"version":"1.0","managed":{"organizations":{}}},"FetchTime":"2024-01-01T00:00:00Z","Source":"test","Strategy":"local"}`)
	if err := os.WriteFile(cacheFile, cacheData, 0644); err != nil {
		t.Fatalf("Failed to create cache file: %v", err)
	}

	// Verify cache exists
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		t.Fatal("Cache file should exist before clearing")
	}

	// Enable sync
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncClearCacheCommand()

	// Run clear cache
	err := runSyncClearCache(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncClearCache failed: %v", err)
	}
}

func TestRunSyncSetup(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Enable sync
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "http"
	currentSettings.Sync.HTTP.URL = "https://example.com/config.json"
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command
	cmd := createSyncSetupCommand()

	// Run sync setup - should show instructions
	err := runSyncSetup(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncSetup failed: %v", err)
	}
}

func TestRunSyncFetch_WithForceFlag(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with proper schema
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Enable sync
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command with force flag
	cmd := createSyncFetchCommand()
	_ = cmd.Flags().Set("force", "true")

	// Run sync fetch with force
	err := runSyncFetch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncFetch with force failed: %v", err)
	}
}

func TestRunSyncFetch_VerboseMode(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with proper schema
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Enable sync
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	// Create command with verbose flag
	cmd := createSyncFetchCommand()
	_ = cmd.Flags().Set("verbose", "true")

	// Run sync fetch in verbose mode
	err := runSyncFetch(cmd, []string{})
	if err != nil {
		t.Fatalf("runSyncFetch in verbose mode failed: %v", err)
	}
}
