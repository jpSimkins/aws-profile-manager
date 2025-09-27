package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// tokyoNightTheme implements Tokyo Night theme
type tokyoNightTheme struct {
	base fyne.Theme
}

// NewTokyoNightTheme creates a Tokyo Night theme
func NewTokyoNightTheme() fyne.Theme {
	return &tokyoNightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Tokyo Night colors
func (t *tokyoNightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Tokyo Night color palette
	// Background: #1a1b26, Selection: #33467c, Foreground: #c0caf5
	// Red: #f7768e, Orange: #ff9e64, Yellow: #e0af68
	// Green: #9ece6a, Cyan: #7dcfff, Blue: #7aa2f7, Purple: #bb9af7

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 26, G: 27, B: 38, A: 255} // Background #1a1b26
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 36, G: 40, B: 59, A: 255} // Darker #24283b
	case theme.ColorNameForeground:
		return color.NRGBA{R: 192, G: 202, B: 245, A: 255} // Foreground #c0caf5
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 86, G: 95, B: 137, A: 255} // Comment #565f89
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 122, G: 162, B: 247, A: 255} // Blue #7aa2f7
	case theme.ColorNameFocus:
		return color.NRGBA{R: 125, G: 207, B: 255, A: 255} // Cyan #7dcfff
	case theme.ColorNameSelection:
		return color.NRGBA{R: 51, G: 70, B: 124, A: 255} // Selection #33467c
	case theme.ColorNameButton:
		return color.NRGBA{R: 36, G: 40, B: 59, A: 255} // Darker #24283b
	case theme.ColorNameHover:
		return color.NRGBA{R: 51, G: 70, B: 124, A: 255} // Selection #33467c
	case theme.ColorNamePressed:
		return color.NRGBA{R: 122, G: 162, B: 247, A: 255} // Blue #7aa2f7
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 36, G: 40, B: 59, A: 255} // Darker #24283b
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 86, G: 95, B: 137, A: 255} // Comment #565f89
	case theme.ColorNameError:
		return color.NRGBA{R: 247, G: 118, B: 142, A: 255} // Red #f7768e
	case theme.ColorNameWarning:
		return color.NRGBA{R: 224, G: 175, B: 104, A: 255} // Yellow #e0af68
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 158, G: 206, B: 106, A: 255} // Green #9ece6a
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 36, G: 40, B: 59, A: 255} // Darker #24283b
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 86, G: 95, B: 137, A: 255} // Comment #565f89
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 187, G: 154, B: 247, A: 255} // Purple #bb9af7
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *tokyoNightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *tokyoNightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *tokyoNightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
