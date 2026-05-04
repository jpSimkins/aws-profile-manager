package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewAyuLightTheme(t *testing.T) {
	ayuTheme := NewAyuLightTheme()
	if ayuTheme == nil {
		t.Fatal("NewAyuLightTheme() should not return nil")
	}
}

func TestAyuLightTheme_Color(t *testing.T) {
	ayuTheme := NewAyuLightTheme()

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
			color := ayuTheme.Color(colorName, theme.VariantLight)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestAyuLightTheme_Font(t *testing.T) {
	ayuTheme := NewAyuLightTheme()
	font := ayuTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestAyuLightTheme_Icon(t *testing.T) {
	ayuTheme := NewAyuLightTheme()
	icon := ayuTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestAyuLightTheme_Size(t *testing.T) {
	ayuTheme := NewAyuLightTheme()
	size := ayuTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
