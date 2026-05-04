package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// githubDarkTheme implements GitHub Dark theme
type githubDarkTheme struct {
	base fyne.Theme
}

// NewGithubDarkTheme creates a GitHub Dark theme
func NewGithubDarkTheme() fyne.Theme {
	return &githubDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns GitHub Dark colors
func (t *githubDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// GitHub Dark palette
	// Background: #0d1117, Foreground: #c9d1d9
	// Blue: #58a6ff, Green: #3fb950

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 13, G: 17, B: 23, A: 255} // #0d1117
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 22, G: 27, B: 34, A: 255} // #161b22
	case theme.ColorNameForeground:
		return color.NRGBA{R: 201, G: 209, B: 217, A: 255} // #c9d1d9
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 110, G: 118, B: 129, A: 255} // #6e7681
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // Blue #58a6ff
	case theme.ColorNameFocus:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // Blue #58a6ff
	case theme.ColorNameSelection:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 33, G: 38, B: 45, A: 255} // #21262d
	case theme.ColorNameHover:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255} // Slightly lighter
	case theme.ColorNamePressed:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // Blue
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 13, G: 17, B: 23, A: 255} // #0d1117
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255} // #30363d
	case theme.ColorNameError:
		return color.NRGBA{R: 248, G: 81, B: 73, A: 255} // Red #f85149
	case theme.ColorNameWarning:
		return color.NRGBA{R: 219, G: 179, B: 0, A: 255} // Yellow #dbb300
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 63, G: 185, B: 80, A: 255} // Green #3fb950
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 33, G: 38, B: 45, A: 255} // #21262d
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255} // #30363d
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // Blue #58a6ff
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *githubDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *githubDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *githubDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
