package views

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/test"
)

func TestShowExportDialog(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	footer := components.NewFooter()

	// Should not panic
	ShowExportDialog(window, footer)
}
