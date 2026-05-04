package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// DarculaTheme implements a Darcula-inspired dark theme similar to JetBrains IDEs
type DarculaTheme struct{}

// NewDarculaTheme creates a new Darcula theme
func NewDarculaTheme() fyne.Theme {
	return &DarculaTheme{}
}

// Color returns the color for the specified theme color name and variant
func (t *DarculaTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Darcula color palette
	switch name {
	// Background colors
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x2b, G: 0x2b, B: 0x2b, A: 0xff} // #2b2b2b - main background
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x36, G: 0x5a, B: 0x81, A: 0xff} // #365a81 - button background
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x3c, G: 0x3f, B: 0x41, A: 0xff} // #3c3f41 - disabled button
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x45, G: 0x49, B: 0x4a, A: 0xff} // #45494a - input background
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 0x3c, G: 0x3f, B: 0x41, A: 0xff} // #3c3f41 - menu background
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0x3c, G: 0x3f, B: 0x41, A: 0xdd} // Semi-transparent overlay

	// Foreground colors
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xa9, G: 0xb7, B: 0xc6, A: 0xff} // #a9b7c6 - main text
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x60, G: 0x63, B: 0x66, A: 0xff} // #606366 - disabled text
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x78, G: 0x7c, B: 0x7f, A: 0xff} // #787c7f - placeholder text

	// Accent colors
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x46, G: 0x98, B: 0xf9, A: 0xff} // #4698f9 - primary accent (blue)
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x4b, G: 0x6e, B: 0xaf, A: 0xff} // #4b6eaf - hover state
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x46, G: 0x98, B: 0xf9, A: 0xff} // #4698f9 - focus indicator

	// Status colors
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x62, G: 0x9a, B: 0x55, A: 0xff} // #629a55 - success/green
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xbb, G: 0xb5, B: 0x29, A: 0xff} // #bbb529 - warning/yellow
	case theme.ColorNameError:
		return color.NRGBA{R: 0xbc, G: 0x3f, B: 0x3c, A: 0xff} // #bc3f3c - error/red

	// Separator colors
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x32, G: 0x35, B: 0x37, A: 0xff} // #323537 - separator lines

	// Selection colors
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x21, G: 0x43, B: 0x65, A: 0xff} // #214365 - selected item
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0x36, G: 0x5a, B: 0x81, A: 0xff} // #365a81 - pressed state

	// Scrollbar colors
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x3c, G: 0x3f, B: 0x41, A: 0xff} // #3c3f41 - scrollbar background

	// Hyperlink color
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x58, G: 0x9d, B: 0xf6, A: 0xff} // #589df6 - links

	default:
		// For any colors not explicitly defined, use Fyne's default dark theme
		return theme.DefaultTheme().Color(name, theme.VariantDark)
	}
}

// Font returns the font for the specified text style
func (t *DarculaTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon for the specified theme icon name
func (t *DarculaTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the specified theme size name
func (t *DarculaTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
