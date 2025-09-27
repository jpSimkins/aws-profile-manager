package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewGruvboxDarkTheme(t *testing.T) {
	gruvboxTheme := NewGruvboxDarkTheme()
	if gruvboxTheme == nil {
		t.Fatal("NewGruvboxDarkTheme() should not return nil")
	}
}

func TestGruvboxDarkTheme_Color(t *testing.T) {
	gruvboxTheme := NewGruvboxDarkTheme()

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
			color := gruvboxTheme.Color(colorName, theme.VariantDark)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestGruvboxDarkTheme_Font(t *testing.T) {
	gruvboxTheme := NewGruvboxDarkTheme()
	font := gruvboxTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestGruvboxDarkTheme_Icon(t *testing.T) {
	gruvboxTheme := NewGruvboxDarkTheme()
	icon := gruvboxTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestGruvboxDarkTheme_Size(t *testing.T) {
	gruvboxTheme := NewGruvboxDarkTheme()
	size := gruvboxTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
