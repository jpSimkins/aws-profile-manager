package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// oneDarkTheme implements One Dark theme (Atom editor)
type oneDarkTheme struct {
	base fyne.Theme
}

// NewOneDarkTheme creates a One Dark theme
func NewOneDarkTheme() fyne.Theme {
	return &oneDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns One Dark colors
func (t *oneDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// One Dark color palette (Atom)
	// Background: #282c34, Gutter: #21252b, Foreground: #abb2bf
	// Red: #e06c75, Green: #98c379, Yellow: #e5c07b, Blue: #61afef
	// Magenta: #c678dd, Cyan: #56b6c2

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 40, G: 44, B: 52, A: 255} // Background #282c34
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 33, G: 37, B: 43, A: 255} // Gutter #21252b
	case theme.ColorNameForeground:
		return color.NRGBA{R: 171, G: 178, B: 191, A: 255} // Foreground #abb2bf
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 92, G: 99, B: 112, A: 255} // Comment #5c6370
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 152, G: 195, B: 121, A: 255} // Green #98c379
	case theme.ColorNameFocus:
		return color.NRGBA{R: 97, G: 175, B: 239, A: 255} // Blue #61afef
	case theme.ColorNameSelection:
		return color.NRGBA{R: 57, G: 63, B: 77, A: 255} // Selection #393f4d
	case theme.ColorNameButton:
		return color.NRGBA{R: 51, G: 55, B: 64, A: 255} // Lighter #333740
	case theme.ColorNameHover:
		return color.NRGBA{R: 57, G: 63, B: 77, A: 255} // Selection #393f4d
	case theme.ColorNamePressed:
		return color.NRGBA{R: 152, G: 195, B: 121, A: 255} // Green #98c379
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 33, G: 37, B: 43, A: 255} // Gutter #21252b
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 92, G: 99, B: 112, A: 255} // Comment #5c6370
	case theme.ColorNameError:
		return color.NRGBA{R: 224, G: 108, B: 117, A: 255} // Red #e06c75
	case theme.ColorNameWarning:
		return color.NRGBA{R: 229, G: 192, B: 123, A: 255} // Yellow #e5c07b
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 152, G: 195, B: 121, A: 255} // Green #98c379
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 51, G: 55, B: 64, A: 255} // Lighter #333740
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 92, G: 99, B: 112, A: 255} // Comment #5c6370
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 86, G: 182, B: 194, A: 255} // Cyan #56b6c2
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *oneDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *oneDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *oneDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
