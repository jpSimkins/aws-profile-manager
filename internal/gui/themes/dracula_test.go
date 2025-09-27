package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewDraculaTheme(t *testing.T) {
	draculaTheme := NewDraculaTheme()
	if draculaTheme == nil {
		t.Fatal("NewDraculaTheme() should not return nil")
	}
}

func TestDraculaTheme_Color(t *testing.T) {
	draculaTheme := NewDraculaTheme()

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

func TestDraculaTheme_Font(t *testing.T) {
	draculaTheme := NewDraculaTheme()
	font := draculaTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestDraculaTheme_Icon(t *testing.T) {
	draculaTheme := NewDraculaTheme()
	icon := draculaTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestDraculaTheme_Size(t *testing.T) {
	draculaTheme := NewDraculaTheme()
	size := draculaTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
