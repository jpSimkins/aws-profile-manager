package gui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"aws-profile-manager/internal/gui/themes"
)

// customTheme wraps the default theme but allows forcing a specific variant
type customTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

// NewCustomTheme creates a theme that forces a specific variant (light or dark)
// variant should be theme.VariantLight or theme.VariantDark
func NewCustomTheme(variant fyne.ThemeVariant) fyne.Theme {
	return &customTheme{
		base:    theme.DefaultTheme(),
		variant: variant,
	}
}

// Color returns the color for the given name, always using the forced variant
func (t *customTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	// Ignore the passed variant and use our forced variant
	return t.base.Color(name, t.variant)
}

// Font returns the font for the given text style
func (t *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

// Icon returns the icon for the given name
func (t *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

// Size returns the size for the given name
func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name)
}

// ApplyTheme applies the specified theme to the Fyne app
func ApplyTheme(app fyne.App, themeName string) {
	switch strings.ToLower(themeName) {
	// Default themes (top of list)
	case "system":
		app.Settings().SetTheme(theme.DefaultTheme())
	case "light":
		app.Settings().SetTheme(NewCustomTheme(theme.VariantLight))
	case "dark":
		app.Settings().SetTheme(NewCustomTheme(theme.VariantDark))
	// Custom themes (alphabetical)
	case "ayu-dark":
		app.Settings().SetTheme(themes.NewAyuDarkTheme())
	case "ayu-light":
		app.Settings().SetTheme(themes.NewAyuLightTheme())
	case "darcula":
		app.Settings().SetTheme(themes.NewDarculaTheme())
	case "dracula":
		app.Settings().SetTheme(themes.NewDraculaTheme())
	case "github-dark":
		app.Settings().SetTheme(themes.NewGithubDarkTheme())
	case "github-light":
		app.Settings().SetTheme(themes.NewGithubLightTheme())
	case "gruvbox-dark":
		app.Settings().SetTheme(themes.NewGruvboxDarkTheme())
	case "gruvbox-light":
		app.Settings().SetTheme(themes.NewGruvboxLightTheme())
	case "material-dark":
		app.Settings().SetTheme(themes.NewMaterialDarkTheme())
	case "material-light":
		app.Settings().SetTheme(themes.NewMaterialLightTheme())
	case "monokai":
		app.Settings().SetTheme(themes.NewMonokaiTheme())
	case "nord":
		app.Settings().SetTheme(themes.NewNordTheme())
	case "one-dark":
		app.Settings().SetTheme(themes.NewOneDarkTheme())
	case "solarized-dark":
		app.Settings().SetTheme(themes.NewSolarizedDarkTheme())
	case "solarized-light":
		app.Settings().SetTheme(themes.NewSolarizedLightTheme())
	case "tokyo-night":
		app.Settings().SetTheme(themes.NewTokyoNightTheme())
	default:
		// Fallback to system theme
		app.Settings().SetTheme(theme.DefaultTheme())
	}
}
