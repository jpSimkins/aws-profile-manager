package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/bundled"
	"aws-profile-manager/internal/logging"
)

// Header represents the application header with logo
type Header struct {
	widget.BaseWidget
	logo  *canvas.Image
	title *widget.RichText
}

// NewHeader creates a new header component with theme-aware logo
func NewHeader() *Header {
	logging.Debug.Log("\t🔹 Creating header component")

	h := &Header{}
	h.ExtendBaseWidget(h)

	logging.Debug.Log("\t🔹 Header component created")
	return h
}

// CreateRenderer creates the renderer for the header widget
func (h *Header) CreateRenderer() fyne.WidgetRenderer {
	// Create logo image
	logoResource := h.getLogoResourceForTheme()
	logging.Debug.Logf("\t🔹 Logo resource loaded: %s (size: %d bytes)", logoResource.Name(), len(logoResource.Content()))

	h.logo = canvas.NewImageFromResource(logoResource)
	h.logo.FillMode = canvas.ImageFillContain
	// Logo aspect ratio is ~1.42:1, constrain height to 50px
	// Width = 50 * 1.42 ≈ 71px
	h.logo.SetMinSize(fyne.NewSize(71, 50))

	logging.Debug.Logf("\t🔹 Logo configured: MinSize=%v, FillMode=%v",
		h.logo.MinSize(), h.logo.FillMode)

	// Create title using markdown for clean, theme-aware styling
	h.title = widget.NewRichTextFromMarkdown("# AWS Profile Manager")

	// Create horizontal box with logo and title
	headerContent := container.NewHBox(
		h.logo,
		h.title,
	)

	// Center the entire header
	centeredContainer := container.NewCenter(headerContent)

	return &headerRenderer{
		header:    h,
		container: centeredContainer,
	}
}

// headerRenderer is the renderer for the Header widget
type headerRenderer struct {
	header    *Header
	container *fyne.Container
}

// Layout arranges the header content
func (r *headerRenderer) Layout(size fyne.Size) {
	r.container.Resize(size)
}

// MinSize returns the minimum size of the header
func (r *headerRenderer) MinSize() fyne.Size {
	return r.container.MinSize()
}

// Refresh updates the header when the theme changes
func (r *headerRenderer) Refresh() {
	logging.Debug.Log("\t🔹 Header refresh triggered (theme may have changed, normal to see multiple refreshes due to Fyne internals)")

	// Update logo resource based on current theme
	logoResource := r.header.getLogoResourceForTheme()
	r.header.logo.Resource = logoResource

	// Refresh container (automatically refreshes all children: logo and title)
	r.container.Refresh()
}

// Objects returns the objects to render
func (r *headerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.container}
}

// Destroy cleans up the renderer
func (r *headerRenderer) Destroy() {
	// Nothing to clean up
}

// getLogoResourceForTheme returns the appropriate logo based on the current theme
func (h *Header) getLogoResourceForTheme() fyne.Resource {
	// In Fyne 2.7.0+, use theme colors to determine if we're in dark mode
	// Check the background color brightness to determine theme variant
	bgColor := theme.Color(theme.ColorNameBackground)

	// Calculate brightness using relative luminance formula
	r, g, b, _ := bgColor.RGBA()
	// Convert from 16-bit color to 8-bit and calculate luminance
	brightness := (0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8))

	logging.Debug.Logf("\t🔹 Background brightness: %.2f", brightness)

	// If brightness < 128 (mid-point), we're in dark mode
	if brightness < 128 {
		logging.Debug.Log("\t🔹 Using dark mode logo (brightness-based detection)")
		return bundled.ResourceLogoDarkMode
	}

	logging.Debug.Log("\t🔹 Using light mode logo (brightness-based detection)")
	return bundled.ResourceLogo
}

// GetLogo returns the logo image (useful for testing)
func (h *Header) GetLogo() *canvas.Image {
	return h.logo
}
