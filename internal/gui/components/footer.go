package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/logging"
)

// Footer represents the application footer/status bar
type Footer struct {
	statusLabel *widget.Label
	container   *fyne.Container
}

// NewFooter creates a new footer component with status display
func NewFooter() *Footer {
	logging.Debug.Log("Creating footer component")

	f := &Footer{}

	// Create status label
	f.statusLabel = widget.NewLabel("Ready")
	f.statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Create footer container
	f.container = container.NewHBox(
		f.statusLabel,
	)

	logging.Debug.Log("Footer component created")
	return f
}

// SetStatus updates the status message in the footer (thread-safe)
func (f *Footer) SetStatus(status string) {
	logging.Debug.Logf("Footer status updated: %s", status)
	// Schedule UI update on main thread (required by Fyne)
	fyne.Do(func() {
		f.statusLabel.SetText(status)
	})
}

// GetStatus returns the current status message
func (f *Footer) GetStatus() string {
	return f.statusLabel.Text
}

// GetContainer returns the footer container
func (f *Footer) GetContainer() *fyne.Container {
	return f.container
}

// GetStatusLabel returns the status label (useful for testing)
func (f *Footer) GetStatusLabel() *widget.Label {
	return f.statusLabel
}
