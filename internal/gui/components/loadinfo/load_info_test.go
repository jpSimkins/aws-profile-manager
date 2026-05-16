package loadinfo

import (
	"testing"
	"time"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/test"
)

func TestNewLoadInfo(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)

	if l == nil {
		t.Fatal("NewLoadInfo returned nil")
	}
	if l.button == nil {
		t.Fatal("button should not be nil")
	}
	// Button should be disabled until a source is set.
	if !l.button.Disabled() {
		t.Error("button should be disabled initially")
	}
}

func TestNewLoadInfo_NilWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Should not panic with a nil window.
	l := NewLoadInfo(nil)
	if l == nil {
		t.Fatal("NewLoadInfo returned nil")
	}
	if !l.button.Disabled() {
		t.Error("button should be disabled with nil window")
	}
}

func TestLoadInfo_SetSource(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)

	l.SetSource("/path/to/config.json")

	if l.source != "/path/to/config.json" {
		t.Errorf("source = %q, want /path/to/config.json", l.source)
	}
	if l.button.Disabled() {
		t.Error("button should be enabled after SetSource with non-empty path")
	}
}

func TestLoadInfo_SetSource_Empty(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)

	l.SetSource("/some/path")
	l.SetSource("")

	if !l.button.Disabled() {
		t.Error("button should be disabled when source is cleared")
	}
}

func TestLoadInfo_SetSource_NilWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Button should stay disabled even with a source when window is nil.
	l := NewLoadInfo(nil)
	l.SetSource("/some/path")

	if !l.button.Disabled() {
		t.Error("button should stay disabled with nil window even when source is set")
	}
}

func TestLoadInfo_SetLoadedAt(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)
	now := time.Now()
	l.SetLoadedAt(now)

	if !l.loadedAt.Equal(now) {
		t.Error("loadedAt should be updated")
	}
}

func TestLoadInfo_GetContent(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)

	content := l.GetContent()
	if content == nil {
		t.Fatal("GetContent should not return nil")
	}
}

func TestLoadInfo_ShowDetails_NilWindow(t *testing.T) {
	test.SetupTestEnvironment(t)

	// showDetails should not panic when window is nil.
	l := NewLoadInfo(nil)
	l.source = "/some/path"
	l.showDetails() // should be a no-op
}

func TestLoadInfo_ShowDetails_EmptySource(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	window := testApp.NewWindow("Test")
	l := NewLoadInfo(window)
	l.showDetails() // should be a no-op when source is empty
}
