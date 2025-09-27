package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// githubLightTheme implements GitHub Light theme
type githubLightTheme struct {
	base fyne.Theme
}

// NewGithubLightTheme creates a GitHub Light theme
func NewGithubLightTheme() fyne.Theme {
	return &githubLightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns GitHub Light colors
func (t *githubLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// GitHub Light palette
	// Background: #ffffff, Foreground: #24292f
	// Blue: #0969da, Green: #1a7f37

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 246, G: 248, B: 250, A: 255} // #f6f8fa
	case theme.ColorNameForeground:
		return color.NRGBA{R: 36, G: 41, B: 47, A: 255} // #24292f
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 139, G: 148, B: 158, A: 255} // #8b949e
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 9, G: 105, B: 218, A: 255} // Blue #0969da
	case theme.ColorNameFocus:
		return color.NRGBA{R: 9, G: 105, B: 218, A: 255} // Blue #0969da
	case theme.ColorNameSelection:
		return color.NRGBA{R: 9, G: 105, B: 218, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 246, G: 248, B: 250, A: 255} // #f6f8fa
	case theme.ColorNameHover:
		return color.NRGBA{R: 236, G: 239, B: 241, A: 255} // Slightly darker
	case theme.ColorNamePressed:
		return color.NRGBA{R: 9, G: 105, B: 218, A: 255} // Blue
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 208, G: 215, B: 222, A: 255} // #d0d7de
	case theme.ColorNameError:
		return color.NRGBA{R: 207, G: 34, B: 46, A: 255} // Red #cf222e
	case theme.ColorNameWarning:
		return color.NRGBA{R: 191, G: 135, B: 0, A: 255} // Yellow #bf8700
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 26, G: 127, B: 55, A: 255} // Green #1a7f37
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 216, G: 222, B: 228, A: 255} // #d8dee4
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 20}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 208, G: 215, B: 222, A: 255} // #d0d7de
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 9, G: 105, B: 218, A: 255} // Blue #0969da
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *githubLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *githubLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *githubLightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
