package components

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/test"
)

func TestShowCustomDialog(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	// Test basic dialog creation
	content := widget.NewLabel("Test content")
	opts := DialogOptions{
		Title:   "Test Dialog",
		Content: content,
		Window:  window,
	}

	dialog := ShowCustomDialog(opts)
	if dialog == nil {
		t.Fatal("Dialog should not be nil")
	}
}

func TestShowCustomDialog_WithButtons(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	buttons := CreateStandardButtons("Cancel", "Save",
		func() { /* cancel action */ },
		func() { /* save action */ },
	)

	content := widget.NewLabel("Test content")
	opts := DialogOptions{
		Title:   "Test Dialog",
		Content: content,
		Window:  window,
		Buttons: buttons,
	}

	dialog := ShowCustomDialog(opts)
	if dialog == nil {
		t.Fatal("Dialog should not be nil")
	}

	// Verify buttons were created
	if len(buttons) != 2 {
		t.Errorf("Expected 2 buttons, got %d", len(buttons))
	}

	// Verify button importance
	if buttons[0].Importance != widget.MediumImportance {
		t.Error("Cancel button should have MediumImportance")
	}
	if buttons[1].Importance != widget.HighImportance {
		t.Error("Action button should have HighImportance")
	}
}

func TestShowCustomDialog_Scrollable(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	content := widget.NewLabel("Test content")
	opts := DialogOptions{
		Title:      "Test Dialog",
		Content:    content,
		Window:     window,
		Scrollable: true,
	}

	dialog := ShowCustomDialog(opts)
	if dialog == nil {
		t.Fatal("Dialog should not be nil")
	}
}

func TestShowCustomDialog_WithSettings(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	content := widget.NewLabel("Test content")
	opts := DialogOptions{
		Title:       "Test Dialog",
		Content:     content,
		Window:      window,
		UseSettings: true,
	}

	dialog := ShowCustomDialog(opts)
	if dialog == nil {
		t.Fatal("Dialog should not be nil")
	}
}

func TestCreateStandardButtons(t *testing.T) {
	cancelCalled := false
	actionCalled := false

	buttons := CreateStandardButtons("Cancel", "Save",
		func() { cancelCalled = true },
		func() { actionCalled = true },
	)

	if len(buttons) != 2 {
		t.Fatalf("Expected 2 buttons, got %d", len(buttons))
	}

	// Test cancel button
	if buttons[0].Label != "Cancel" {
		t.Errorf("Expected Cancel label, got %s", buttons[0].Label)
	}
	buttons[0].OnTapped()
	if !cancelCalled {
		t.Error("Cancel callback should have been called")
	}

	// Test action button
	if buttons[1].Label != "Save" {
		t.Errorf("Expected Save label, got %s", buttons[1].Label)
	}
	buttons[1].OnTapped()
	if !actionCalled {
		t.Error("Action callback should have been called")
	}
}
