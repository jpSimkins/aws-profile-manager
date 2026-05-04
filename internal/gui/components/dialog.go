package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/settings"
)

// DialogButton represents a button in a dialog
type DialogButton struct {
	Label      string
	OnTapped   func()
	Importance widget.Importance // Default, Primary, Danger, Warning, Success
}

// DialogOptions contains configuration for creating a dialog
type DialogOptions struct {
	Title       string
	Content     fyne.CanvasObject
	Buttons     []DialogButton
	Window      fyne.Window
	Scrollable  bool // If true, content is wrapped in a scroll container
	UseSettings bool // If true, uses dialog width/height from settings
}

// ShowCustomDialog creates and shows a standardized dialog with custom buttons
// This ensures consistent look and feel across all dialogs in the application
func ShowCustomDialog(opts DialogOptions) *dialog.CustomDialog {
	// Wrap content in scroll container if requested
	content := opts.Content
	if opts.Scrollable {
		content = container.NewScroll(content)
	}

	// Create dialog with Cancel as default dismiss button
	// The actual Cancel button will be added via SetButtons
	customDialog := dialog.NewCustom(opts.Title, "Cancel", content, opts.Window)

	// Apply size from settings if requested
	if opts.UseSettings {
		guiSettings := settings.Get().GUI
		width := float32(guiSettings.DialogWidth)
		height := float32(guiSettings.DialogHeight)
		customDialog.Resize(fyne.NewSize(width, height))
	}

	// Create buttons if provided
	if len(opts.Buttons) > 0 {
		buttons := make([]fyne.CanvasObject, len(opts.Buttons))
		for i, btn := range opts.Buttons {
			// Capture variables for closure
			buttonDef := btn

			button := widget.NewButton(buttonDef.Label, func() {
				if buttonDef.OnTapped != nil {
					buttonDef.OnTapped()
				}
				// Don't auto-hide - let the callback decide
			})

			// Set button importance if specified
			if buttonDef.Importance != widget.MediumImportance {
				button.Importance = buttonDef.Importance
			}

			buttons[i] = button
		}

		customDialog.SetButtons(buttons)
	}

	return customDialog
}

// CreateStandardButtons creates a standard Cancel + Action button pair
// The action button is automatically set to HighImportance (primary)
func CreateStandardButtons(cancelLabel, actionLabel string, onCancel, onAction func()) []DialogButton {
	return []DialogButton{
		{
			Label:      cancelLabel,
			OnTapped:   onCancel,
			Importance: widget.MediumImportance,
		},
		{
			Label:      actionLabel,
			OnTapped:   onAction,
			Importance: widget.HighImportance,
		},
	}
}
