package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewSolarizedDarkTheme(t *testing.T) {
	themeInstance := NewSolarizedDarkTheme()
	if themeInstance == nil {
		t.Fatal("NewSolarizedDarkTheme() should not return nil")
	}
}

func TestSolarizedDarkTheme_Color(t *testing.T) {
	themeInstance := NewSolarizedDarkTheme()

	colors := []fyne.ThemeColorName{
		theme.ColorNameBackground,
		theme.ColorNameForeground,
		theme.ColorNamePrimary,
		theme.ColorNameButton,
		theme.ColorNameSuccess,
		theme.ColorNameWarning,
		theme.ColorNameError,
		theme.ColorNameSelection,
		theme.ColorNameInputBackground,
	}

	for _, colorName := range colors {
		t.Run(string(colorName), func(t *testing.T) {
			color := themeInstance.Color(colorName, theme.VariantDark)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestSolarizedDarkTheme_Font(t *testing.T) {
	themeInstance := NewSolarizedDarkTheme()
	font := themeInstance.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestSolarizedDarkTheme_Icon(t *testing.T) {
	themeInstance := NewSolarizedDarkTheme()
	icon := themeInstance.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestSolarizedDarkTheme_Size(t *testing.T) {
	themeInstance := NewSolarizedDarkTheme()
	size := themeInstance.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
