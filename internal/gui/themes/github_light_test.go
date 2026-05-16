package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewGithubLightTheme(t *testing.T) {
	githubTheme := NewGithubLightTheme()
	if githubTheme == nil {
		t.Fatal("NewGithubLightTheme() should not return nil")
	}
}

func TestGithubLightTheme_Color(t *testing.T) {
	githubTheme := NewGithubLightTheme()

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
			color := githubTheme.Color(colorName, theme.VariantLight)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestGithubLightTheme_Font(t *testing.T) {
	githubTheme := NewGithubLightTheme()
	font := githubTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestGithubLightTheme_Icon(t *testing.T) {
	githubTheme := NewGithubLightTheme()
	icon := githubTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestGithubLightTheme_Size(t *testing.T) {
	githubTheme := NewGithubLightTheme()
	size := githubTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
