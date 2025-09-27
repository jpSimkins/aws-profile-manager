package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewTokyoNightTheme(t *testing.T) {
	themeInstance := NewTokyoNightTheme()
	if themeInstance == nil {
		t.Fatal("NewTokyoNightTheme() should not return nil")
	}
}

func TestTokyoNightTheme_Color(t *testing.T) {
	themeInstance := NewTokyoNightTheme()

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

func TestTokyoNightTheme_Font(t *testing.T) {
	themeInstance := NewTokyoNightTheme()
	font := themeInstance.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestTokyoNightTheme_Icon(t *testing.T) {
	themeInstance := NewTokyoNightTheme()
	icon := themeInstance.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestTokyoNightTheme_Size(t *testing.T) {
	themeInstance := NewTokyoNightTheme()
	size := themeInstance.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
