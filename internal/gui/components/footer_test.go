package components

import (
	"aws-profile-manager/internal/test"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
)

func TestNewFooter(t *testing.T) {
	// Create Fyne test app (headless)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	// Initialize core app state
	test.SetupTestEnvironment(t)

	// Create footer
	footer := NewFooter()

	if footer == nil {
		t.Fatal("NewFooter() should not return nil")
	}

	if footer.GetContainer() == nil {
		t.Error("Footer container should not be nil")
	}

	if footer.GetStatusLabel() == nil {
		t.Error("Footer status label should not be nil")
	}

	// Check default status
	if footer.GetStatus() != "Ready" {
		t.Errorf("Default status should be 'Ready', got '%s'", footer.GetStatus())
	}
}

func TestFooter_SetStatus(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	footer := NewFooter()

	// Set new status
	testStatus := "Testing status message"
	footer.SetStatus(testStatus)

	// Verify status was updated
	if footer.GetStatus() != testStatus {
		t.Errorf("Expected status '%s', got '%s'", testStatus, footer.GetStatus())
	}
}

func TestFooter_GetStatus(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	footer := NewFooter()

	// Initial status should be "Ready"
	status := footer.GetStatus()
	if status != "Ready" {
		t.Errorf("Expected initial status 'Ready', got '%s'", status)
	}

	// Change status
	footer.SetStatus("New Status")
	status = footer.GetStatus()
	if status != "New Status" {
		t.Errorf("Expected status 'New Status', got '%s'", status)
	}
}

func TestFooter_GetContainer(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	footer := NewFooter()
	container := footer.GetContainer()

	if container == nil {
		t.Fatal("GetContainer() should not return nil")
	}

	// Verify container has content
	if len(container.Objects) == 0 {
		t.Error("Footer container should have objects")
	}
}

func TestFooter_GetStatusLabel(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	footer := NewFooter()
	label := footer.GetStatusLabel()

	if label == nil {
		t.Fatal("GetStatusLabel() should not return nil")
	}

	// Verify label has correct initial text
	if label.Text != "Ready" {
		t.Errorf("Status label should have initial text 'Ready', got '%s'", label.Text)
	}
}

func TestFooter_MultipleStatusUpdates(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	footer := NewFooter()

	// Test multiple status updates
	statuses := []string{
		"Status 1",
		"Status 2",
		"Loading...",
		"Complete",
		"Ready",
	}

	for _, status := range statuses {
		footer.SetStatus(status)
		if footer.GetStatus() != status {
			t.Errorf("Expected status '%s', got '%s'", status, footer.GetStatus())
		}
	}
}
