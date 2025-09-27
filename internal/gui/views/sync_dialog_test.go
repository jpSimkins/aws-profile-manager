package views

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestShowSyncDialog_Disabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	// Explicitly set sync to disabled (don't assume defaults)
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = false
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	window := testApp.NewWindow("Test")
	defer window.Close()

	footer := components.NewFooter()

	// Should not panic and should show "Sync is disabled" message
	ShowSyncDialog(window, footer)

	// Check footer status
	status := footer.GetStatus()
	if status != "Sync is disabled" {
		t.Errorf("Expected footer status 'Sync is disabled', got '%s'", status)
	}
}

func TestShowSyncDialog_Enabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	// Create a test config file for local sync
	configDir := test.GetTestConfigDir(t)
	testConfigFile := filepath.Join(configDir, "test-config.json")
	testConfigData := []byte(`{"version":"2.0","managed":{},"unmanaged":{}}`)
	if err := os.WriteFile(testConfigFile, testConfigData, 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Enable sync with local strategy using new settings system
	currentSettings := settings.Get()
	currentSettings.Sync.Enabled = true
	currentSettings.Sync.Strategy = "local"
	currentSettings.Sync.Local.Path = testConfigFile
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	window := testApp.NewWindow("Test")
	defer window.Close()

	footer := components.NewFooter()

	// Should not panic and should start syncing
	ShowSyncDialog(window, footer)

	// Check footer status was updated to syncing
	status := footer.GetStatus()
	if status != "Syncing configuration..." {
		t.Errorf("Expected footer status 'Syncing configuration...', got '%s'", status)
	}

	// Wait a bit for the async goroutine to complete
	// This prevents cleanup warnings when test ends
	time.Sleep(100 * time.Millisecond)
}
