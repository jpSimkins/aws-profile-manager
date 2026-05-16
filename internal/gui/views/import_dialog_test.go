package views

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/test"
)

func TestShowImportDialog(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	footer := components.NewFooter()

	// Should not panic
	ShowImportDialog(window, footer)
}

func TestShowImportOptionsDialog(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	footer := components.NewFooter()

	// Should not panic - error path is handled gracefully when file doesn't exist
	showImportPreview(window, footer, "/tmp/nonexistent-backup.json")
}
