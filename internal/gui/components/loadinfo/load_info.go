// Package loadinfo provides a compact icon button that reveals data source details.
//
// Designed for use in view headers to keep the UI clean while still making
// source path and load timestamp easily accessible on demand.
package loadinfo

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/settings"
)

// LoadInfo is a compact ⓘ icon button that shows data source and timestamp on tap.
//
// The button is disabled until SetSource is called with a non-empty path.
// If the window is nil (headless/test mode), the button stays disabled.
//
// Usage:
//
//	info := loadinfo.NewLoadInfo(window)
//	// ... embed info.GetContent() in the header layout ...
//
//	// When data loads:
//	info.SetSource("/home/user/.config/aws-profile-manager/cache/sync-config.json")
//	info.SetLoadedAt(time.Now())
type LoadInfo struct {
	window   fyne.Window
	source   string
	loadedAt time.Time
	button   *widget.Button
}

// NewLoadInfo creates a new LoadInfo for the given window.
//
// Pass nil for window in headless or test contexts — the button will be visible
// but disabled, and tapping it will have no effect.
func NewLoadInfo(window fyne.Window) *LoadInfo {
	l := &LoadInfo{window: window}
	l.button = widget.NewButtonWithIcon("", theme.InfoIcon(), l.showDetails)
	l.button.Importance = widget.LowImportance
	l.button.Disable()
	return l
}

// SetSource updates the data source path and enables the button.
//
// Pass an empty string to clear the source and disable the button.
// The button remains disabled if the window is nil.
func (l *LoadInfo) SetSource(source string) {
	l.source = source
	if source != "" && l.window != nil {
		l.button.Enable()
	} else {
		l.button.Disable()
	}
}

// SetLoadedAt records when the data was last loaded.
func (l *LoadInfo) SetLoadedAt(t time.Time) {
	l.loadedAt = t
}

// GetContent returns the info icon button to embed in a layout.
func (l *LoadInfo) GetContent() fyne.CanvasObject {
	return l.button
}

// showDetails opens a small dialog displaying the source path and load timestamp.
func (l *LoadInfo) showDetails() {
	if l.source == "" || l.window == nil {
		return
	}

	loadedText := "Unknown"
	if !l.loadedAt.IsZero() {
		loadedText = l.loadedAt.Format("Jan 2, 2006 at 15:04:05")
	}

	// Use a wrapping label for long paths.
	sourceLabel := widget.NewLabel(l.source)
	// sourceLabel.Wrapping = fyne.TextWrapBreak

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("**Source:**"),
		sourceLabel,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("**Last Loaded:**"),
		widget.NewLabel(loadedText),
	)

	// Use dialog.NewCustom directly — loadinfo is a low-level component and
	// importing the parent components package would create a circular dependency.
	// Enforce the configured dialog width via a transparent spacer so the dialog
	// is wide enough to display long paths, while height auto-fits the content.
	guiSettings := settings.Get().GUI
	minWidthSpacer := canvas.NewRectangle(color.Transparent)
	minWidthSpacer.SetMinSize(fyne.NewSize(float32(guiSettings.DialogWidth), 0))
	constrainedContent := container.NewStack(content, minWidthSpacer)

	d := dialog.NewCustom("Information", "OK", constrainedContent, l.window)
	d.Show()
}
