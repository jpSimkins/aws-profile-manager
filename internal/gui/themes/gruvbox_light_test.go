package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewGruvboxLightTheme(t *testing.T) {
	gruvboxTheme := NewGruvboxLightTheme()
	if gruvboxTheme == nil {
		t.Fatal("NewGruvboxLightTheme() should not return nil")
	}
}

func TestGruvboxLightTheme_Color(t *testing.T) {
	gruvboxTheme := NewGruvboxLightTheme()

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
			color := gruvboxTheme.Color(colorName, theme.VariantLight)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestGruvboxLightTheme_Font(t *testing.T) {
	gruvboxTheme := NewGruvboxLightTheme()
	font := gruvboxTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestGruvboxLightTheme_Icon(t *testing.T) {
	gruvboxTheme := NewGruvboxLightTheme()
	icon := gruvboxTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestGruvboxLightTheme_Size(t *testing.T) {
	gruvboxTheme := NewGruvboxLightTheme()
	size := gruvboxTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
