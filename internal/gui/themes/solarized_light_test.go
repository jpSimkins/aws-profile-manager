package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewSolarizedLightTheme(t *testing.T) {
	themeInstance := NewSolarizedLightTheme()
	if themeInstance == nil {
		t.Fatal("NewSolarizedLightTheme() should not return nil")
	}
}

func TestSolarizedLightTheme_Color(t *testing.T) {
	themeInstance := NewSolarizedLightTheme()

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
			color := themeInstance.Color(colorName, theme.VariantLight)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestSolarizedLightTheme_Font(t *testing.T) {
	themeInstance := NewSolarizedLightTheme()
	font := themeInstance.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestSolarizedLightTheme_Icon(t *testing.T) {
	themeInstance := NewSolarizedLightTheme()
	icon := themeInstance.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestSolarizedLightTheme_Size(t *testing.T) {
	themeInstance := NewSolarizedLightTheme()
	size := themeInstance.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
