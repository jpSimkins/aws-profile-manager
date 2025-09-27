package views

import (
	"testing"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/test"
)

func TestNewProfilesView_WithoutWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	view := NewProfilesView(nil)
	if view == nil {
		t.Fatal("NewProfilesView should not return nil")
	}
}

func TestNewProfilesView_WithoutWindow_DisablesRefreshButton(t *testing.T) {
	test.SetupTestEnvironment(t)

	view := NewProfilesView(nil)
	if view == nil {
		t.Fatal("NewProfilesView should not return nil")
	}

	refreshButton := findButtonByText(view, "Refresh")
	if refreshButton == nil {
		t.Fatal("expected to find Refresh button")
	}

	if !refreshButton.Disabled() {
		t.Fatal("expected Refresh button to be disabled when window is nil")
	}
}

func TestNewProfilesView_WithWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Profiles")
	defer window.Close()

	view := NewProfilesView(window)
	if view == nil {
		t.Fatal("NewProfilesView should not return nil")
	}
}

func findButtonByText(obj fyne.CanvasObject, text string) *widget.Button {
	if obj == nil {
		return nil
	}

	if button, ok := obj.(*widget.Button); ok {
		if button.Text == text {
			return button
		}
	}

	if containerObj, ok := obj.(*fyne.Container); ok {
		for _, child := range containerObj.Objects {
			if found := findButtonByText(child, text); found != nil {
				return found
			}
		}
	}

	return nil
}
