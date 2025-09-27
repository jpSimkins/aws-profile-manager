package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// draculaTheme implements Dracula theme (purple/pink dark theme)
type draculaTheme struct {
	base fyne.Theme
}

// NewDraculaTheme creates a Dracula theme
func NewDraculaTheme() fyne.Theme {
	return &draculaTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Dracula colors
func (t *draculaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Dracula color palette
	// Background: #282a36, Selection: #44475a, Foreground: #f8f8f2
	// Comment: #6272a4, Cyan: #8be9fd, Green: #50fa7b, Orange: #ffb86c
	// Pink: #ff79c6, Purple: #bd93f9, Red: #ff5555, Yellow: #f1fa8c

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 40, G: 42, B: 54, A: 255} // Background #282a36
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 68, G: 71, B: 90, A: 255} // Selection #44475a
	case theme.ColorNameForeground:
		return color.NRGBA{R: 248, G: 248, B: 242, A: 255} // Foreground #f8f8f2
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 98, G: 114, B: 164, A: 255} // Comment #6272a4
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 189, G: 147, B: 249, A: 255} // Purple #bd93f9
	case theme.ColorNameFocus:
		return color.NRGBA{R: 139, G: 233, B: 253, A: 255} // Cyan #8be9fd
	case theme.ColorNameSelection:
		return color.NRGBA{R: 68, G: 71, B: 90, A: 255} // Selection #44475a
	case theme.ColorNameButton:
		return color.NRGBA{R: 68, G: 71, B: 90, A: 255} // Selection #44475a
	case theme.ColorNameHover:
		return color.NRGBA{R: 98, G: 114, B: 164, A: 255} // Comment #6272a4
	case theme.ColorNamePressed:
		return color.NRGBA{R: 189, G: 147, B: 249, A: 255} // Purple #bd93f9
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 68, G: 71, B: 90, A: 255} // Selection #44475a
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 98, G: 114, B: 164, A: 255} // Comment #6272a4
	case theme.ColorNameError:
		return color.NRGBA{R: 255, G: 85, B: 85, A: 255} // Red #ff5555
	case theme.ColorNameWarning:
		return color.NRGBA{R: 255, G: 184, B: 108, A: 255} // Orange #ffb86c
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 80, G: 250, B: 123, A: 255} // Green #50fa7b
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 68, G: 71, B: 90, A: 255} // Selection #44475a
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 98, G: 114, B: 164, A: 255} // Comment #6272a4
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 255, G: 121, B: 198, A: 255} // Pink #ff79c6
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *draculaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *draculaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *draculaTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
