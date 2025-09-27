package viewheader

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/test"
)

func TestNew(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test Title", "Test description.")

	if h == nil {
		t.Fatal("New returned nil")
	}
	if h.title != "# Test Title" {
		t.Errorf("title = %q, want '# Test Title'", h.title)
	}
	if h.description != "Test description." {
		t.Errorf("description = %q, want 'Test description.'", h.description)
	}
	if h.infoWidget != nil {
		t.Error("infoWidget should be nil by default")
	}
	if len(h.buttons) != 0 {
		t.Error("buttons should be empty by default")
	}
}

func TestViewHeader_WithInfo(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test", "Desc")
	label := widget.NewLabel("info")

	result := h.WithInfo(label)

	if result != h {
		t.Error("WithInfo should return the same ViewHeader for chaining")
	}
	if h.infoWidget == nil {
		t.Error("infoWidget should be set after WithInfo")
	}
}

func TestViewHeader_WithButtons(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test", "Desc")
	btn1 := widget.NewButton("Button 1", nil)
	btn2 := widget.NewButton("Button 2", nil)

	result := h.WithButtons(btn1, btn2)

	if result != h {
		t.Error("WithButtons should return the same ViewHeader for chaining")
	}
	if len(h.buttons) != 2 {
		t.Errorf("expected 2 buttons, got %d", len(h.buttons))
	}
}

func TestViewHeader_WithButtons_Chaining(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test", "Desc")
	btn1 := widget.NewButton("Button 1", nil)
	btn2 := widget.NewButton("Button 2", nil)
	btn3 := widget.NewButton("Button 3", nil)

	h.WithButtons(btn1).WithButtons(btn2, btn3)

	if len(h.buttons) != 3 {
		t.Errorf("expected 3 buttons after chained WithButtons, got %d", len(h.buttons))
	}
}

func TestViewHeader_GetContent(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test Title", "Description text.")
	btn := widget.NewButton("Action", nil)
	h.WithButtons(btn)

	content := h.GetContent()
	if content == nil {
		t.Fatal("GetContent should not return nil")
	}
}

func TestViewHeader_GetContent_WithInfo(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	h := New("# Test Title", "Description.")
	info := widget.NewLabel("ⓘ")
	h.WithInfo(info)

	content := h.GetContent()
	if content == nil {
		t.Fatal("GetContent should not return nil with info widget")
	}
}

func TestViewHeader_GetContent_Empty(t *testing.T) {
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()

	// Should work with no buttons or info widget.
	h := New("# Minimal", "")
	content := h.GetContent()
	if content == nil {
		t.Fatal("GetContent should not return nil for minimal header")
	}
}
