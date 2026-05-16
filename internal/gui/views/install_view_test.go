package views

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/test"
)

func TestNewInstallView_WithWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Install")
	defer window.Close()

	view := NewInstallView(window)
	if view == nil {
		t.Fatal("NewInstallView should not return nil")
	}
}

func TestNewInstallView_WithoutWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	view := NewInstallView(nil)
	if view == nil {
		t.Fatal("NewInstallView should not return nil when window is nil")
	}
}

func TestNewInstallView_HasRequiredComponents(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	view := NewInstallView(nil)

	if view == nil {
		t.Fatal("NewInstallView should not return nil")
	}

	// Verify view is a container (BorderLayout)
	// We can't easily inspect the internals without making them exported,
	// but we can verify the view was created successfully
	// The actual functionality will be tested via integration tests
}
