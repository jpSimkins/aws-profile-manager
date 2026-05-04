package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// monokaiTheme implements Monokai theme (Sublime Text classic)
type monokaiTheme struct {
	base fyne.Theme
}

// NewMonokaiTheme creates a Monokai theme
func NewMonokaiTheme() fyne.Theme {
	return &monokaiTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Monokai colors
func (t *monokaiTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Monokai color palette
	// Background: #272822, Foreground: #f8f8f2
	// Red: #f92672, Orange: #fd971f, Yellow: #e6db74
	// Green: #a6e22e, Cyan: #66d9ef, Blue: #ae81ff

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 39, G: 40, B: 34, A: 255} // Background #272822
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 58, G: 61, B: 52, A: 255} // Slightly lighter
	case theme.ColorNameForeground:
		return color.NRGBA{R: 248, G: 248, B: 242, A: 255} // Foreground #f8f8f2
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 117, G: 113, B: 94, A: 255} // Comment #75715e
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 166, G: 226, B: 46, A: 255} // Green #a6e22e
	case theme.ColorNameFocus:
		return color.NRGBA{R: 102, G: 217, B: 239, A: 255} // Cyan #66d9ef
	case theme.ColorNameSelection:
		return color.NRGBA{R: 73, G: 72, B: 62, A: 255} // Selection #49483e
	case theme.ColorNameButton:
		return color.NRGBA{R: 58, G: 61, B: 52, A: 255} // Slightly lighter
	case theme.ColorNameHover:
		return color.NRGBA{R: 73, G: 72, B: 62, A: 255} // Selection #49483e
	case theme.ColorNamePressed:
		return color.NRGBA{R: 166, G: 226, B: 46, A: 255} // Green #a6e22e
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 58, G: 61, B: 52, A: 255} // Slightly lighter
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 117, G: 113, B: 94, A: 255} // Comment #75715e
	case theme.ColorNameError:
		return color.NRGBA{R: 249, G: 38, B: 114, A: 255} // Red #f92672
	case theme.ColorNameWarning:
		return color.NRGBA{R: 253, G: 151, B: 31, A: 255} // Orange #fd971f
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 166, G: 226, B: 46, A: 255} // Green #a6e22e
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 58, G: 61, B: 52, A: 255} // Slightly lighter
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 117, G: 113, B: 94, A: 255} // Comment #75715e
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 174, G: 129, B: 255, A: 255} // Blue #ae81ff
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *monokaiTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *monokaiTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *monokaiTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
