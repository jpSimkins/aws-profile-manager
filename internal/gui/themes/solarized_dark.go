package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// solarizedDarkTheme implements Solarized Dark theme
type solarizedDarkTheme struct {
	base fyne.Theme
}

// NewSolarizedDarkTheme creates a Solarized Dark theme
func NewSolarizedDarkTheme() fyne.Theme {
	return &solarizedDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Solarized Dark colors
func (t *solarizedDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Solarized Dark base colors
	// Base03: #002b36, Base02: #073642, Base01: #586e75, Base00: #657b83
	// Base0: #839496, Base1: #93a1a1, Base2: #eee8d5, Base3: #fdf6e3
	// Accent colors: yellow, orange, red, magenta, violet, blue, cyan, green

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0, G: 43, B: 54, A: 255} // Base03 #002b36
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 7, G: 54, B: 66, A: 255} // Base02 #073642
	case theme.ColorNameForeground:
		return color.NRGBA{R: 131, G: 148, B: 150, A: 255} // Base0 #839496
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 88, G: 110, B: 117, A: 255} // Base01 #586e75
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 42, G: 161, B: 152, A: 255} // Cyan #2aa198
	case theme.ColorNameFocus:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 255} // Blue #268bd2
	case theme.ColorNameSelection:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 7, G: 54, B: 66, A: 255} // Base02 #073642
	case theme.ColorNameHover:
		return color.NRGBA{R: 88, G: 110, B: 117, A: 255} // Base01 #586e75
	case theme.ColorNamePressed:
		return color.NRGBA{R: 42, G: 161, B: 152, A: 255} // Cyan #2aa198
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 7, G: 54, B: 66, A: 255} // Base02 #073642
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 88, G: 110, B: 117, A: 255} // Base01 #586e75
	case theme.ColorNameError:
		return color.NRGBA{R: 220, G: 50, B: 47, A: 255} // Red #dc322f
	case theme.ColorNameWarning:
		return color.NRGBA{R: 181, G: 137, B: 0, A: 255} // Yellow #b58900
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 133, G: 153, B: 0, A: 255} // Green #859900
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 7, G: 54, B: 66, A: 255} // Base02 #073642
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 88, G: 110, B: 117, A: 255} // Base01 #586e75
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 255} // Blue #268bd2
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *solarizedDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *solarizedDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *solarizedDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
