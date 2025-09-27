package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewAyuDarkTheme(t *testing.T) {
	ayuTheme := NewAyuDarkTheme()
	if ayuTheme == nil {
		t.Fatal("NewAyuDarkTheme() should not return nil")
	}
}

func TestAyuDarkTheme_Color(t *testing.T) {
	ayuTheme := NewAyuDarkTheme()

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
			color := ayuTheme.Color(colorName, theme.VariantDark)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestAyuDarkTheme_Font(t *testing.T) {
	ayuTheme := NewAyuDarkTheme()
	font := ayuTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestAyuDarkTheme_Icon(t *testing.T) {
	ayuTheme := NewAyuDarkTheme()
	icon := ayuTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestAyuDarkTheme_Size(t *testing.T) {
	ayuTheme := NewAyuDarkTheme()
	size := ayuTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
