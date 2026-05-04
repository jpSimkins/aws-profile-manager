package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// solarizedLightTheme implements Solarized Light theme
type solarizedLightTheme struct {
	base fyne.Theme
}

// NewSolarizedLightTheme creates a Solarized Light theme
func NewSolarizedLightTheme() fyne.Theme {
	return &solarizedLightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Solarized Light colors
func (t *solarizedLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 253, G: 246, B: 227, A: 255} // Base3 #fdf6e3
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 238, G: 232, B: 213, A: 255} // Base2 #eee8d5
	case theme.ColorNameForeground:
		return color.NRGBA{R: 101, G: 123, B: 131, A: 255} // Base00 #657b83
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 147, G: 161, B: 161, A: 255} // Base1 #93a1a1
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 42, G: 161, B: 152, A: 255} // Cyan #2aa198
	case theme.ColorNameFocus:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 255} // Blue #268bd2
	case theme.ColorNameSelection:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 100} // Blue with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 238, G: 232, B: 213, A: 255} // Base2 #eee8d5
	case theme.ColorNameHover:
		return color.NRGBA{R: 147, G: 161, B: 161, A: 255} // Base1 #93a1a1
	case theme.ColorNamePressed:
		return color.NRGBA{R: 42, G: 161, B: 152, A: 255} // Cyan #2aa198
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 238, G: 232, B: 213, A: 255} // Base2 #eee8d5
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 147, G: 161, B: 161, A: 255} // Base1 #93a1a1
	case theme.ColorNameError:
		return color.NRGBA{R: 220, G: 50, B: 47, A: 255} // Red #dc322f
	case theme.ColorNameWarning:
		return color.NRGBA{R: 181, G: 137, B: 0, A: 255} // Yellow #b58900
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 133, G: 153, B: 0, A: 255} // Green #859900
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 238, G: 232, B: 213, A: 255} // Base2 #eee8d5
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 50}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 147, G: 161, B: 161, A: 255} // Base1 #93a1a1
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 38, G: 139, B: 210, A: 255} // Blue #268bd2
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *solarizedLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *solarizedLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *solarizedLightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
