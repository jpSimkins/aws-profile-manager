package components

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/test"
)

func TestShowProgressDialog(t *testing.T) {
	// Setup test environment
	test.SetupTestEnvironment(t)

	// Initialize app for settings
	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	app := fyneTest.NewApp()
	defer app.Quit()

	window := app.NewWindow("Test")

	progress := ShowProgressDialog(window, "Test Progress", "Testing...", nil)

	if progress == nil {
		t.Fatal("Expected progress dialog, got nil")
	}

	if progress.dialog == nil {
		t.Error("Progress dialog should have dialog")
	}

	if progress.progressBar == nil {
		t.Error("Progress dialog should have progress bar")
	}

	if progress.statusLabel == nil {
		t.Error("Progress dialog should have status label")
	}

	if progress.detailsLabel == nil {
		t.Error("Progress dialog should have details label")
	}
}

func TestProgressDialog_UpdateStatus(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	app := fyneTest.NewApp()
	defer app.Quit()

	window := app.NewWindow("Test")

	progress := ShowProgressDialog(window, "Test Progress", "Initial message", nil)

	// Update status
	progress.UpdateStatus("Updated message")

	if progress.statusLabel.Text != "Updated message" {
		t.Errorf("Expected status 'Updated message', got '%s'", progress.statusLabel.Text)
	}
}

func TestProgressDialog_UpdateDetails(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	app := fyneTest.NewApp()
	defer app.Quit()

	window := app.NewWindow("Test")

	progress := ShowProgressDialog(window, "Test Progress", "Processing...", nil)

	// Update details
	progress.UpdateDetails("Step 1 of 3")

	if progress.detailsLabel.Text != "Step 1 of 3" {
		t.Errorf("Expected details 'Step 1 of 3', got '%s'", progress.detailsLabel.Text)
	}
}

func TestProgressDialog_ShowHide(t *testing.T) {
	test.SetupTestEnvironment(t)

	if err := core.App.Initialize(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	app := fyneTest.NewApp()
	defer app.Quit()

	window := app.NewWindow("Test")

	progress := ShowProgressDialog(window, "Test Progress", "Processing...", nil)

	// Show should start the progress bar
	progress.Show()

	// Hide should stop the progress bar
	progress.Hide()

	// Test passes if no panic occurs
}
