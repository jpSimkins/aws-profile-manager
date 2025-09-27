package gui

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"aws-profile-manager/internal/test"
)

func TestNewCustomTheme(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	// Test light theme
	lightTheme := NewCustomTheme(theme.VariantLight)
	if lightTheme == nil {
		t.Fatal("NewCustomTheme(VariantLight) should not return nil")
	}

	// Test dark theme
	darkTheme := NewCustomTheme(theme.VariantDark)
	if darkTheme == nil {
		t.Fatal("NewCustomTheme(VariantDark) should not return nil")
	}

	// Verify they return different colors for background
	lightBg := lightTheme.Color(theme.ColorNameBackground, theme.VariantLight)
	darkBg := darkTheme.Color(theme.ColorNameBackground, theme.VariantDark)

	if lightBg == darkBg {
		t.Error("Light and dark themes should have different background colors")
	}
}

func TestCustomTheme_IgnoresPassedVariant(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	// Create a light theme
	lightTheme := NewCustomTheme(theme.VariantLight)

	// Request a color with dark variant - should still return light color
	bgWithDarkRequest := lightTheme.Color(theme.ColorNameBackground, theme.VariantDark)
	bgWithLightRequest := lightTheme.Color(theme.ColorNameBackground, theme.VariantLight)

	// Should return the same color regardless of requested variant
	if bgWithDarkRequest != bgWithLightRequest {
		t.Error("Custom theme should ignore passed variant and use forced variant")
	}
}

func TestApplyTheme(t *testing.T) {
	tests := []struct {
		name         string
		themeSetting string
	}{
		{"Light theme", "Light"},
		{"Dark theme", "Dark"},
		{"Darcula theme", "Darcula"},
		{"System theme", "System"},
		{"Default fallback", "Invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testApp := fyneTest.NewApp()
			defer testApp.Quit()

			test.SetupTestEnvironment(t)

			// Apply theme should not panic
			ApplyTheme(testApp, tt.themeSetting)

			// Verify theme was set
			currentTheme := testApp.Settings().Theme()
			if currentTheme == nil {
				t.Error("Theme should not be nil after ApplyTheme")
			}
		})
	}
}

func TestCustomTheme_FontAndIcon(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	test.SetupTestEnvironment(t)

	customTheme := NewCustomTheme(theme.VariantLight)

	// Test Font
	textStyle := fyne.TextStyle{Bold: true}
	font := customTheme.Font(textStyle)
	if font == nil {
		t.Error("Font() should not return nil")
	}

	// Test Icon
	icon := customTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}

	// Test Size
	size := customTheme.Size(theme.SizeNameText)
	if size <= 0 {
		t.Errorf("Size() should return positive value, got %f", size)
	}
}
