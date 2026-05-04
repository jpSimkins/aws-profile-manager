package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewMaterialLightTheme(t *testing.T) {
	themeInstance := NewMaterialLightTheme()
	if themeInstance == nil {
		t.Fatal("NewMaterialLightTheme() should not return nil")
	}
}

func TestMaterialLightTheme_Color(t *testing.T) {
	themeInstance := NewMaterialLightTheme()

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

func TestMaterialLightTheme_Font(t *testing.T) {
	themeInstance := NewMaterialLightTheme()
	font := themeInstance.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestMaterialLightTheme_Icon(t *testing.T) {
	themeInstance := NewMaterialLightTheme()
	icon := themeInstance.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestMaterialLightTheme_Size(t *testing.T) {
	themeInstance := NewMaterialLightTheme()
	size := themeInstance.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
