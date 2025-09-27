package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// gruvboxLightTheme implements Gruvbox Light theme
type gruvboxLightTheme struct {
	base fyne.Theme
}

// NewGruvboxLightTheme creates a Gruvbox Light theme
func NewGruvboxLightTheme() fyne.Theme {
	return &gruvboxLightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Gruvbox Light colors
func (t *gruvboxLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 251, G: 241, B: 199, A: 255} // bg0 #fbf1c7
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 235, G: 219, B: 178, A: 255} // bg1 #ebdbb2
	case theme.ColorNameForeground:
		return color.NRGBA{R: 60, G: 56, B: 54, A: 255} // fg0 #3c3836
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 189, G: 174, B: 147, A: 255} // fg4 #bdae93
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 121, G: 116, B: 14, A: 255} // green #79740e
	case theme.ColorNameFocus:
		return color.NRGBA{R: 66, G: 123, B: 88, A: 255} // aqua #427b58
	case theme.ColorNameSelection:
		return color.NRGBA{R: 66, G: 123, B: 88, A: 100} // aqua with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 213, G: 196, B: 161, A: 255} // bg2 #d5c4a1
	case theme.ColorNameHover:
		return color.NRGBA{R: 189, G: 174, B: 147, A: 255} // bg3 #bdae93
	case theme.ColorNamePressed:
		return color.NRGBA{R: 121, G: 116, B: 14, A: 255} // green #79740e
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 235, G: 219, B: 178, A: 255} // bg1 #ebdbb2
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 168, G: 153, B: 132, A: 255} // bg4 #a89984
	case theme.ColorNameError:
		return color.NRGBA{R: 204, G: 36, B: 29, A: 255} // red #cc241d
	case theme.ColorNameWarning:
		return color.NRGBA{R: 215, G: 153, B: 33, A: 255} // yellow #d79921
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 121, G: 116, B: 14, A: 255} // green #79740e
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 213, G: 196, B: 161, A: 255} // bg2 #d5c4a1
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 50}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 168, G: 153, B: 132, A: 255} // bg4 #a89984
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 66, G: 123, B: 88, A: 255} // aqua #427b58
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *gruvboxLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *gruvboxLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *gruvboxLightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
