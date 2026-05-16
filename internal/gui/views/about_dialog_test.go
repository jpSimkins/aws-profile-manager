package views

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/test"
)

func TestShowAboutDialog(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test app and window
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	footer := components.NewFooter()

	// Show about dialog - should not panic
	ShowAboutDialog(window, footer)

	// Verify footer status was updated
	if footer.GetStatus() != "About dialog shown" {
		t.Errorf("Expected footer status 'About dialog shown', got '%s'", footer.GetStatus())
	}
}

func TestShowAboutDialog_WithoutFooter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test app and window
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	// Show about dialog without footer - should not panic
	ShowAboutDialog(window, nil)
}

func TestShowAboutDialog_VersionInfo(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Verify version info is accessible
	versionInfo := core.GetVersion()

	if versionInfo.Version == "" {
		t.Error("Version should not be empty")
	}

	if versionInfo.Platform == "" {
		t.Error("Platform should not be empty")
	}

	if versionInfo.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if versionInfo.Framework == "" {
		t.Error("Framework should not be empty")
	}

	if versionInfo.FrameworkVersion == "" {
		t.Error("FrameworkVersion should not be empty")
	}

	// Verify app constants are set
	if core.AppName == "" {
		t.Error("AppName should not be empty")
	}

	if core.AppAuthor == "" {
		t.Error("AppAuthor should not be empty")
	}

	if core.AppURL == "" {
		t.Error("AppURL should not be empty")
	}
}

func TestGetAboutLogoResourceForTheme(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Theme access requires an active app instance.
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	resource := getAboutLogoResourceForTheme()
	if resource == nil {
		t.Fatal("Expected about logo resource to be non-nil")
	}

	if resource.Name() == "" {
		t.Error("Expected about logo resource to have a name")
	}

	if len(resource.Content()) == 0 {
		t.Error("Expected about logo resource to have content")
	}
}
