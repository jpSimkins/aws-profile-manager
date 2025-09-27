package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// materialDarkTheme implements Material Dark theme
type materialDarkTheme struct {
	base fyne.Theme
}

// NewMaterialDarkTheme creates a Material Dark theme
func NewMaterialDarkTheme() fyne.Theme {
	return &materialDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Material Dark colors
func (t *materialDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Material Dark palette (Google Material Design)
	// Background: #121212, Foreground: #ffffff
	// Primary: #bb86fc (purple), Secondary: #03dac6 (teal)

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // #121212
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 28, G: 28, B: 28, A: 255} // Slightly lighter
	case theme.ColorNameForeground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // #ffffff
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 117, G: 117, B: 117, A: 255} // #757575
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 187, G: 134, B: 252, A: 255} // Purple #bb86fc
	case theme.ColorNameFocus:
		return color.NRGBA{R: 3, G: 218, B: 198, A: 255} // Teal #03dac6
	case theme.ColorNameSelection:
		return color.NRGBA{R: 187, G: 134, B: 252, A: 100} // Purple with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 48, G: 48, B: 48, A: 255} // #303030
	case theme.ColorNameHover:
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255} // #424242
	case theme.ColorNamePressed:
		return color.NRGBA{R: 187, G: 134, B: 252, A: 255} // Purple
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 28, G: 28, B: 28, A: 255}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255} // #424242
	case theme.ColorNameError:
		return color.NRGBA{R: 207, G: 102, B: 121, A: 255} // Red #cf6679
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 224, B: 130, A: 255} // Amber #ffe082
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 129, G: 199, B: 132, A: 255} // Green #81c784
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 48, G: 48, B: 48, A: 255} // #303030
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 66, G: 66, B: 66, A: 255} // #424242
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 100, G: 181, B: 246, A: 255} // Blue #64b5f6
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *materialDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *materialDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *materialDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
