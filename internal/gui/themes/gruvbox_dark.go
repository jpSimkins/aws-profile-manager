package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// gruvboxDarkTheme implements Gruvbox Dark theme
type gruvboxDarkTheme struct {
	base fyne.Theme
}

// NewGruvboxDarkTheme creates a Gruvbox Dark theme
func NewGruvboxDarkTheme() fyne.Theme {
	return &gruvboxDarkTheme{
		base: theme.DefaultTheme(),
	}
}

// Color returns Gruvbox Dark colors
func (t *gruvboxDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Gruvbox Dark palette - warm, retro colors
	// Background: #282828, #3c3836, #504945, #665c54, #7c6f64
	// Foreground: #ebdbb2, #d5c4a1, #bdae93, #a89984
	// Colors: red, green, yellow, blue, purple, aqua, orange

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 40, G: 40, B: 40, A: 255} // bg0 #282828
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 60, G: 56, B: 54, A: 255} // bg1 #3c3836
	case theme.ColorNameForeground:
		return color.NRGBA{R: 235, G: 219, B: 178, A: 255} // fg0 #ebdbb2
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 168, G: 153, B: 132, A: 255} // fg4 #a89984
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 184, G: 187, B: 38, A: 255} // green #b8bb26
	case theme.ColorNameFocus:
		return color.NRGBA{R: 131, G: 165, B: 152, A: 255} // aqua #83a598
	case theme.ColorNameSelection:
		return color.NRGBA{R: 131, G: 165, B: 152, A: 100} // aqua with transparency
	case theme.ColorNameButton:
		return color.NRGBA{R: 80, G: 73, B: 69, A: 255} // bg2 #504945
	case theme.ColorNameHover:
		return color.NRGBA{R: 102, G: 92, B: 84, A: 255} // bg3 #665c54
	case theme.ColorNamePressed:
		return color.NRGBA{R: 184, G: 187, B: 38, A: 255} // green #b8bb26
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 60, G: 56, B: 54, A: 255} // bg1 #3c3836
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 124, G: 111, B: 100, A: 255} // bg4 #7c6f64
	case theme.ColorNameError:
		return color.NRGBA{R: 251, G: 73, B: 52, A: 255} // red #fb4934
	case theme.ColorNameWarning:
		return color.NRGBA{R: 250, G: 189, B: 47, A: 255} // yellow #fabd2f
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 184, G: 187, B: 38, A: 255} // green #b8bb26
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 80, G: 73, B: 69, A: 255} // bg2 #504945
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 124, G: 111, B: 100, A: 255} // bg4 #7c6f64
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 131, G: 165, B: 152, A: 255} // aqua #83a598
	default:
		return t.base.Color(name, theme.VariantDark)
	}
}

func (t *gruvboxDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *gruvboxDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *gruvboxDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}
