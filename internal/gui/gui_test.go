package gui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestNewApp(t *testing.T) {
	// Initialize core app state
	test.SetupTestEnvironment(t)

	// Create GUI app
	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	if app == nil {
		t.Fatal("NewApp() should not return nil app")
	}

	if app.GetFyneApp() == nil {
		t.Error("Fyne app should not be nil")
	}
}

func TestApp_GetFyneApp(t *testing.T) {
	test.SetupTestEnvironment(t)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	fyneApp := app.GetFyneApp()
	if fyneApp == nil {
		t.Error("GetFyneApp() should not return nil")
	}
}

func TestApp_createTabContent(t *testing.T) {
	// Create Fyne test app
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// createTabContent requires a window
	app.window = testApp.NewWindow("Test")
	defer app.window.Close()

	content := app.createTabContent()
	if content == nil {
		t.Error("createTabContent() should not return nil")
	}
}

func TestApp_createMenu(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create window and footer
	app.window = testApp.NewWindow("Test")
	defer app.window.Close()
	app.footer = components.NewFooter()

	menu := components.CreateMainMenu(components.MenuCallbacks{})
	if menu == nil {
		t.Fatal("CreateMainMenu() should not return nil")
	}

	// Verify File menu exists
	if len(menu.Items) == 0 {
		t.Error("Menu should have items")
	}

	// Verify we have at least 2 menus (File and Help)
	if len(menu.Items) < 2 {
		t.Errorf("Expected at least 2 menus, got %d", len(menu.Items))
	}

	// Check File menu
	fileMenu := menu.Items[0]
	if fileMenu.Label != "File" {
		t.Errorf("First menu should be 'File', got '%s'", fileMenu.Label)
	}

	// Check Help menu
	helpMenu := menu.Items[1]
	if helpMenu.Label != "Help" {
		t.Errorf("Second menu should be 'Help', got '%s'", helpMenu.Label)
	}
}

func TestApp_handleSyncNow(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

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

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create window and footer
	app.window = testApp.NewWindow("Test")
	defer app.window.Close()
	app.footer = components.NewFooter()

	// Call handleSyncNow - should not panic
	app.handleSyncNow()

	// Wait a bit for the async goroutine to complete
	// This prevents cleanup warnings when test ends
	time.Sleep(100 * time.Millisecond)
}

func TestApp_handleSettings(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create footer component
	app.footer = components.NewFooter()

	// Create a test window
	app.window = testApp.NewWindow("Test Settings")

	// overlayStack is needed by handleSettings
	app.overlayStack = container.NewStack()
	tabs := container.NewAppTabs()
	app.overlayStack.Add(tabs)

	// Call handleSettings
	app.handleSettings()

	// Verify it doesn't panic
	// The actual settings dialog is tested in views/settings_view_test.go
}

func TestApp_handleAbout(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create window and footer
	app.window = testApp.NewWindow("Test")
	defer app.window.Close()
	app.footer = components.NewFooter()

	// Call handleAbout - verify it doesn't panic
	app.handleAbout()

	// Note: Footer status update happens async via fyne.Do()
	// We just verify the method doesn't crash
}
