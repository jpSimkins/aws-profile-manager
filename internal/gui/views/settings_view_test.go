package views

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestNewSettingsView(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	footer := components.NewFooter()

	currentSettings := settings.Get()
	currentSettings.GUI.Theme = "dark"
	currentSettings.GUI.WindowWidth = 1024
	currentSettings.GUI.WindowHeight = 768
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	closed := false
	onClose := func() { closed = true }

	view := NewSettingsView(window, footer, onClose, nil)
	if view == nil {
		t.Fatal("NewSettingsView returned nil")
	}
	_ = closed
}

func TestNewSettingsView_WithNilFooter(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	currentSettings := settings.Get()
	currentSettings.GUI.Theme = "light"
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	view := NewSettingsView(window, nil, nil, nil)
	if view == nil {
		t.Fatal("NewSettingsView returned nil")
	}
}

func TestNewSettingsView_SchemaGeneration(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	currentSettings := settings.Get()
	currentSettings.Application.ManagedSectionStart = "START TEST"
	currentSettings.Application.ManagedSectionEnd = "END TEST"
	currentSettings.Logging.EnableDebug = true
	currentSettings.GUI.Theme = "dark"
	currentSettings.Sync.Enabled = false
	currentSettings.AwsCLI.AutoRefresh = true
	if err := settings.Set(currentSettings); err != nil {
		t.Fatalf("Failed to set settings: %v", err)
	}

	footer := components.NewFooter()

	view := NewSettingsView(window, footer, nil, nil)
	if view == nil {
		t.Fatal("NewSettingsView returned nil")
	}
}
