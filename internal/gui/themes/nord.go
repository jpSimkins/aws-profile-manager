package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// nordTheme implements Nord theme (Arctic-inspired)
type nordTheme struct {
	base fyne.Theme
}

// NewNordTheme creates a Nord theme
func NewNordTheme() fyne.Theme {
	return &nordTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Nord colors
func (t *nordTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Nord color palette
	// Polar Night: #2e3440, #3b4252, #434c5e, #4c566a
	// Snow Storm: #d8dee9, #e5e9f0, #eceff4
	// Frost: #8fbcbb, #88c0d0, #81a1c1, #5e81ac
	// Aurora: #bf616a (red), #d08770 (orange), #ebcb8b (yellow), #a3be8c (green), #b48ead (purple)

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 46, G: 52, B: 64, A: 255} // Polar Night 0 #2e3440
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 59, G: 66, B: 82, A: 255} // Polar Night 1 #3b4252
	case theme.ColorNameForeground:
		return color.NRGBA{R: 216, G: 222, B: 233, A: 255} // Snow Storm 0 #d8dee9
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 76, G: 86, B: 106, A: 255} // Polar Night 3 #4c566a
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 136, G: 192, B: 208, A: 255} // Frost 1 #88c0d0
	case theme.ColorNameFocus:
		return color.NRGBA{R: 94, G: 129, B: 172, A: 255} // Frost 3 #5e81ac
	case theme.ColorNameSelection:
		return color.NRGBA{R: 94, G: 129, B: 172, A: 100} // Frost 3 with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 67, G: 76, B: 94, A: 255} // Polar Night 2 #434c5e
	case theme.ColorNameHover:
		return color.NRGBA{R: 76, G: 86, B: 106, A: 255} // Polar Night 3 #4c566a
	case theme.ColorNamePressed:
		return color.NRGBA{R: 136, G: 192, B: 208, A: 255} // Frost 1 #88c0d0
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 59, G: 66, B: 82, A: 255} // Polar Night 1 #3b4252
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 76, G: 86, B: 106, A: 255} // Polar Night 3 #4c566a
	case theme.ColorNameError:
		return color.NRGBA{R: 191, G: 97, B: 106, A: 255} // Aurora Red #bf616a
	case theme.ColorNameWarning:
		return color.NRGBA{R: 235, G: 203, B: 139, A: 255} // Aurora Yellow #ebcb8b
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 163, G: 190, B: 140, A: 255} // Aurora Green #a3be8c
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 67, G: 76, B: 94, A: 255} // Polar Night 2 #434c5e
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 76, G: 86, B: 106, A: 255} // Polar Night 3 #4c566a
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 136, G: 192, B: 208, A: 255} // Frost 1 #88c0d0
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *nordTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *nordTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *nordTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
