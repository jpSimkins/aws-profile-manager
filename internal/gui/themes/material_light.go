package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// materialLightTheme implements Material Light theme
type materialLightTheme struct {
	base fyne.Theme
}

// NewMaterialLightTheme creates a Material Light theme
func NewMaterialLightTheme() fyne.Theme {
	return &materialLightTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Material Light colors
func (t *materialLightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Material Light palette (Google Material Design)
	// Background: #ffffff, Foreground: #212121
	// Primary: #6200ee (purple), Secondary: #03dac6 (teal)

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // #f5f5f5
	case theme.ColorNameForeground:
		return color.NRGBA{R: 33, G: 33, B: 33, A: 255} // #212121
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 158, G: 158, B: 158, A: 255} // #9e9e9e
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 98, G: 0, B: 238, A: 255} // Purple #6200ee
	case theme.ColorNameFocus:
		return color.NRGBA{R: 3, G: 218, B: 198, A: 255} // Teal #03dac6
	case theme.ColorNameSelection:
		return color.NRGBA{R: 98, G: 0, B: 238, A: 100} // Purple with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 245, G: 245, B: 245, A: 255} // #f5f5f5
	case theme.ColorNameHover:
		return color.NRGBA{R: 238, G: 238, B: 238, A: 255} // #eeeeee
	case theme.ColorNamePressed:
		return color.NRGBA{R: 98, G: 0, B: 238, A: 255} // Purple
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 250, G: 250, B: 250, A: 255} // #fafafa
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 224, G: 224, B: 224, A: 255} // #e0e0e0
	case theme.ColorNameError:
		return color.NRGBA{R: 211, G: 47, B: 47, A: 255} // Red #d32f2f
	case theme.ColorNameWarning:
		return color.NRGBA{R: 251, G: 192, B: 45, A: 255} // Amber #fbc02d
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 56, G: 142, B: 60, A: 255} // Green #388e3c
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 224, G: 224, B: 224, A: 255} // #e0e0e0
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 30}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 189, G: 189, B: 189, A: 255} // #bdbdbd
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 21, G: 101, B: 192, A: 255} // Blue #1565c0
	default:
		return t.base.Color(name, theme.VariantLight)
	}
}

func (t *materialLightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *materialLightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *materialLightTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
