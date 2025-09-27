package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestNewGithubDarkTheme(t *testing.T) {
	githubTheme := NewGithubDarkTheme()
	if githubTheme == nil {
		t.Fatal("NewGithubDarkTheme() should not return nil")
	}
}

func TestGithubDarkTheme_Color(t *testing.T) {
	githubTheme := NewGithubDarkTheme()

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
			color := githubTheme.Color(colorName, theme.VariantDark)
			if color == nil {
				t.Errorf("Color %v should not be nil", colorName)
			}
		})
	}
}

func TestGithubDarkTheme_Font(t *testing.T) {
	githubTheme := NewGithubDarkTheme()
	font := githubTheme.Font(fyne.TextStyle{})
	if font == nil {
		t.Error("Font() should not return nil")
	}
}

func TestGithubDarkTheme_Icon(t *testing.T) {
	githubTheme := NewGithubDarkTheme()
	icon := githubTheme.Icon(theme.IconNameHome)
	if icon == nil {
		t.Error("Icon() should not return nil")
	}
}

func TestGithubDarkTheme_Size(t *testing.T) {
	githubTheme := NewGithubDarkTheme()
	size := githubTheme.Size(theme.SizeNamePadding)
	if size <= 0 {
		t.Error("Size() should return a positive value")
	}
}
