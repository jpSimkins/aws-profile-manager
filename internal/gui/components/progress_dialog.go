package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/settings"
)

// ProgressDialog shows a progress indicator with status text
type ProgressDialog struct {
	dialog       *dialog.CustomDialog
	progressBar  *widget.ProgressBarInfinite
	statusLabel  *widget.Label
	detailsLabel *widget.Label
	cancelFunc   func() // Cancellation function
}

// ShowProgressDialog creates and shows a progress dialog with optional cancel button
//
// Parameters:
//   - window: Parent window
//   - title: Dialog title
//   - initialMessage: Initial status message
//   - cancelFunc: Optional cancellation function (nil = no cancel button)
func ShowProgressDialog(window fyne.Window, title string, initialMessage string, cancelFunc func()) *ProgressDialog {
	progressBar := widget.NewProgressBarInfinite()
	statusLabel := widget.NewLabel(initialMessage)
	detailsLabel := widget.NewLabel("") // Optional details line

	content := container.NewVBox(
		statusLabel,
		detailsLabel,
		progressBar,
	)

	// Get dialog size from settings
	guiSettings := settings.Get().GUI
	width, height := float32(guiSettings.DialogWidth), float32(200) // Fixed height for progress

	pd := &ProgressDialog{
		progressBar:  progressBar,
		statusLabel:  statusLabel,
		detailsLabel: detailsLabel,
		cancelFunc:   cancelFunc,
	}

	// If cancel function provided, show dialog with cancel button
	if cancelFunc != nil {
		cancelBtn := widget.NewButton("Cancel", func() {
			if pd.cancelFunc != nil {
				pd.cancelFunc()
			}
		})

		pd.dialog = dialog.NewCustom(title, "", content, window)
		pd.dialog.SetButtons([]fyne.CanvasObject{cancelBtn})
		pd.dialog.Resize(fyne.NewSize(width, height))
	} else {
		pd.dialog = dialog.NewCustomWithoutButtons(title, content, window)
		pd.dialog.Resize(fyne.NewSize(width, height))
	}

	return pd
}

// Show displays the progress dialog
func (pd *ProgressDialog) Show() {
	pd.progressBar.Start()
	pd.dialog.Show()
}

// UpdateStatus updates the main status message
func (pd *ProgressDialog) UpdateStatus(message string) {
	pd.statusLabel.SetText(message)
	pd.statusLabel.Refresh()
}

// UpdateDetails updates the details line (step information)
func (pd *ProgressDialog) UpdateDetails(details string) {
	pd.detailsLabel.SetText(details)
	pd.detailsLabel.Refresh()
}

// Hide closes the progress dialog
func (pd *ProgressDialog) Hide() {
	pd.progressBar.Stop()
	pd.dialog.Hide()
}
