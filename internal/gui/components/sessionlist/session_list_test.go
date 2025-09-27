package sessionlist

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/test"
)

// TestNew_WithWindow verifies that the component constructs without panicking
// when a real window is provided.
func TestNew_WithWindow(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	defer window.Close()

	slc := New(window)
	if slc == nil {
		t.Fatal("New should not return nil")
	}
}

// TestNew_NilWindow verifies that the component constructs without panicking
// when window is nil (headless / test mode).
func TestNew_NilWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if slc == nil {
		t.Fatal("New should not return nil with nil window")
	}
}

// TestNew_RefreshButtonDisabledWhenNilWindow verifies the refresh button is
// disabled in headless mode so taps are ignored.
func TestNew_RefreshButtonDisabledWhenNilWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if !slc.refreshButton.Disabled() {
		t.Error("refreshButton should be disabled when window is nil")
	}
}

// TestNew_LogoutButtonHiddenInitially verifies the logout button starts hidden
// and is only shown once active sessions are detected.
func TestNew_LogoutButtonHiddenInitially(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if slc.logoutButton.Visible() {
		t.Error("logoutButton should be hidden initially")
	}
}

// TestNew_LogoutDescLabelHiddenInitially verifies the logout description label
// starts hidden together with the logout button.
func TestNew_LogoutDescLabelHiddenInitially(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if slc.logoutDescLabel.Visible() {
		t.Error("logoutDescLabel should be hidden initially")
	}
}

// TestContent_ReturnsNonNil verifies that Content() returns a usable canvas object.
func TestContent_ReturnsNonNil(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if slc.Content() == nil {
		t.Fatal("Content() should not return nil")
	}
}

// TestGetLoadInfo_ReturnsNonNil verifies that GetLoadInfo returns the embedded
// load info component so callers can surface it in a header.
func TestGetLoadInfo_ReturnsNonNil(t *testing.T) {
	test.SetupTestEnvironment(t)

	slc := New(nil)
	if slc.GetLoadInfo() == nil {
		t.Fatal("GetLoadInfo() should not return nil")
	}
}

// TestUpdateRows_EmptyShowsPlaceholder verifies that when sessions is nil the
// placeholder "no sessions" message is displayed.
func TestUpdateRows_EmptyShowsPlaceholder(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	slc := New(nil)
	slc.updateRows(nil)

	if len(slc.sessionRows.Objects) != 1 {
		t.Fatalf("expected 1 placeholder row, got %d", len(slc.sessionRows.Objects))
	}
}

// TestUpdateRows_PopulatesRows verifies that one canvas object is added per session.
func TestUpdateRows_PopulatesRows(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	slc := New(nil)

	sessions := makeTestSessions(2)
	slc.updateRows(sessions)

	if len(slc.sessionRows.Objects) != 2 {
		t.Fatalf("expected 2 session rows, got %d", len(slc.sessionRows.Objects))
	}
}

// TestUpdateRows_ClearsPreviousRows verifies that calling updateRows a second time
// replaces the previous content rather than appending to it.
func TestUpdateRows_ClearsPreviousRows(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	slc := New(nil)

	slc.updateRows(makeTestSessions(3))
	slc.updateRows(makeTestSessions(1))

	if len(slc.sessionRows.Objects) != 1 {
		t.Fatalf("expected 1 row after second update, got %d", len(slc.sessionRows.Objects))
	}
}
