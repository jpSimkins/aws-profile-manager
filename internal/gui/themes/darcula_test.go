package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewDarculaTheme(t *testing.T) {
	draculaTheme := NewDarculaTheme()
	if draculaTheme == nil {
		t.Fatal("NewDarculaTheme() should not return nil")
	}
}

func TestDarculaTheme_Color(t *testing.T) {
	draculaTheme := NewDarculaTheme()

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
			color := draculaTheme.Color(colorName, theme.VariantDark)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestDarculaTheme_Font(t *testing.T) {
	draculaTheme := NewDarculaTheme()
	font := draculaTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestDarculaTheme_Icon(t *testing.T) {
	draculaTheme := NewDarculaTheme()
	icon := draculaTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestDarculaTheme_Size(t *testing.T) {
	draculaTheme := NewDarculaTheme()
	size := draculaTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
