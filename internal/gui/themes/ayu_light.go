package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ayuLightTheme implements Ayu Light theme
type ayuLightTheme struct {
	base fyne.Theme
}

// NewAyuLightTheme creates an Ayu Light theme
func NewAyuLightTheme() fyne.Theme {
	return &ayuLightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Ayu Light colors
func (t *ayuLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Ayu Light palette - clean and minimalist
	// Background: #fafafa, Foreground: #5c6166
	// Accent: #ff9940 (orange)

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 250, G: 250, B: 250, A: 255} // #fafafa
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255} // Slightly darker
	case theme.ColorNameForeground:
		return color.NRGBA{R: 92, G: 97, B: 102, A: 255} // #5c6166
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 172, G: 178, B: 183, A: 255} // Lighter gray
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 255, G: 153, B: 64, A: 255} // Orange #ff9940
	case theme.ColorNameFocus:
		return color.NRGBA{R: 57, G: 186, B: 230, A: 255} // Blue #39bae6
	case theme.ColorNameSelection:
		return color.NRGBA{R: 57, G: 186, B: 230, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	case theme.ColorNameHover:
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 255, G: 153, B: 64, A: 255} // Orange
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNameError:
		return color.NRGBA{R: 242, G: 88, B: 81, A: 255} // Red #f25851
	case theme.ColorNameWarning:
		return color.NRGBA{R: 232, G: 178, B: 65, A: 255} // Yellow #e8b241
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 134, G: 187, B: 82, A: 255} // Green #86b352
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 30}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 57, G: 186, B: 230, A: 255} // Blue #39bae6
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *ayuLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *ayuLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *ayuLightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
