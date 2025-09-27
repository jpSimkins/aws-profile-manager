package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ayuDarkTheme implements Ayu Dark theme
type ayuDarkTheme struct {
	base fyne.Theme
}

// NewAyuDarkTheme creates an Ayu Dark theme
func NewAyuDarkTheme() fyne.Theme {
	return &ayuDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Ayu Dark colors
func (t *ayuDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Ayu Dark palette
	// Background: #0f1419, Foreground: #bfbdb6
	// Accent: #ff9940 (orange)

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 15, G: 20, B: 25, A: 255} // #0f1419
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 26, G: 31, B: 36, A: 255} // Slightly lighter
	case theme.ColorNameForeground:
		return color.NRGBA{R: 191, G: 189, B: 182, A: 255} // #bfbdb6
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 92, G: 97, B: 102, A: 255} // Darker gray
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 255, G: 153, B: 64, A: 255} // Orange #ff9940
	case theme.ColorNameFocus:
		return color.NRGBA{R: 89, G: 196, B: 228, A: 255} // Blue #59c4e4
	case theme.ColorNameSelection:
		return color.NRGBA{R: 89, G: 196, B: 228, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 26, G: 31, B: 36, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 36, G: 41, B: 46, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 255, G: 153, B: 64, A: 255} // Orange
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 26, G: 31, B: 36, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 60, G: 65, B: 70, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 242, G: 88, B: 81, A: 255} // Red #f25851
	case theme.ColorNameWarning:
		return color.NRGBA{R: 232, G: 178, B: 65, A: 255} // Yellow #e8b241
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 134, G: 187, B: 82, A: 255} // Green #86b352
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 36, G: 41, B: 46, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 60, G: 65, B: 70, A: 255}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 89, G: 196, B: 228, A: 255} // Blue #59c4e4
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *ayuDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *ayuDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *ayuDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
