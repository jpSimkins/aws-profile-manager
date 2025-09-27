package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewMaterialDarkTheme(t *testing.T) {
	themeInstance := NewMaterialDarkTheme()
	if themeInstance == nil {
		t.Fatal("NewMaterialDarkTheme() should not return nil")
	}
}

func TestMaterialDarkTheme_Color(t *testing.T) {
	themeInstance := NewMaterialDarkTheme()

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

func TestMaterialDarkTheme_Font(t *testing.T) {
	themeInstance := NewMaterialDarkTheme()
	font := themeInstance.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestMaterialDarkTheme_Icon(t *testing.T) {
	themeInstance := NewMaterialDarkTheme()
	icon := themeInstance.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestMaterialDarkTheme_Size(t *testing.T) {
	themeInstance := NewMaterialDarkTheme()
	size := themeInstance.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
