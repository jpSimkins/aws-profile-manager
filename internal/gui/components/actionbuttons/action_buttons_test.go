package actionbuttons

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/test"
)

// setupApp initialises a headless Fyne test app and the isolated test environment.
// It must be called at the start of every test that constructs widgets.
func setupApp(t *testing.T) {
	t.Helper()
	test.SetupTestEnvironment(t)
	_ = fyneTest.NewApp()
}

// --- Terminal ---

func TestTerminal_ReturnsButton(t *testing.T) {
	setupApp(t)
	btn := Terminal("my-profile", "us-east-1")
	if btn == nil {
		t.Fatal("Terminal should not return nil")
	}
}

func TestTerminal_IsLowImportance(t *testing.T) {
	setupApp(t)
	btn := Terminal("my-profile", "us-east-1")
	if btn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance, got %v", btn.Importance)
	}
}

func TestTerminal_HasNoLabel(t *testing.T) {
	setupApp(t)
	btn := Terminal("my-profile", "us-east-1")
	if btn.Text != "" {
		t.Errorf("expected empty label, got %q", btn.Text)
	}
}

// --- SsoLogin ---

func TestSsoLogin_ReturnsButton(t *testing.T) {
	setupApp(t)
	btn := SsoLogin("my-session")
	if btn == nil {
		t.Fatal("SsoLogin should not return nil")
	}
}

func TestSsoLogin_IsLowImportance(t *testing.T) {
	setupApp(t)
	btn := SsoLogin("my-session")
	if btn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance, got %v", btn.Importance)
	}
}

func TestSsoLogin_HasNoLabel(t *testing.T) {
	setupApp(t)
	btn := SsoLogin("my-session")
	if btn.Text != "" {
		t.Errorf("expected empty label, got %q", btn.Text)
	}
}

// --- OpenURL ---

func TestOpenURL_ReturnsButton(t *testing.T) {
	setupApp(t)
	btn := OpenURL("https://example.com")
	if btn == nil {
		t.Fatal("OpenURL should not return nil")
	}
}

func TestOpenURL_IsLowImportance(t *testing.T) {
	setupApp(t)
	btn := OpenURL("https://example.com")
	if btn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance, got %v", btn.Importance)
	}
}

func TestOpenURL_DisabledWhenEmpty(t *testing.T) {
	setupApp(t)
	btn := OpenURL("")
	if !btn.Disabled() {
		t.Error("OpenURL with empty rawURL should be disabled")
	}
}

func TestOpenURL_EnabledWhenNonEmpty(t *testing.T) {
	setupApp(t)
	btn := OpenURL("https://example.com")
	if btn.Disabled() {
		t.Error("OpenURL with valid rawURL should not be disabled")
	}
}

func TestOpenURL_HasNoLabel(t *testing.T) {
	setupApp(t)
	btn := OpenURL("https://example.com")
	if btn.Text != "" {
		t.Errorf("expected empty label, got %q", btn.Text)
	}
}

// --- Copy ---

func TestCopy_ReturnsButton(t *testing.T) {
	setupApp(t)
	btn := Copy("some-value")
	if btn == nil {
		t.Fatal("Copy should not return nil")
	}
}

func TestCopy_IsLowImportance(t *testing.T) {
	setupApp(t)
	btn := Copy("some-value")
	if btn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance, got %v", btn.Importance)
	}
}

func TestCopy_HasNoLabel(t *testing.T) {
	setupApp(t)
	btn := Copy("some-value")
	if btn.Text != "" {
		t.Errorf("expected empty label, got %q", btn.Text)
	}
}

func TestCopy_NotDisabledWithValue(t *testing.T) {
	setupApp(t)
	btn := Copy("something")
	if btn.Disabled() {
		t.Error("Copy should not be disabled when value is non-empty")
	}
}
