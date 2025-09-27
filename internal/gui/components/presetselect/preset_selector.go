// Package presetselect provides a component for selecting installation presets.
//
// The preset selector displays a dropdown of available presets from the schema
// and triggers a callback when the selection changes. Presets automatically
// populate filter selections in the Install view.
package presetselect

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/schema"
)

// PresetSelector provides a UI component for selecting installation presets.
//
// The component displays:
//   - Help text explaining what presets do
//   - Dropdown with preset options
//   - Description label showing selected preset details
//
// When a preset is selected, it calls the OnPresetChange callback with the
// selected preset configuration.
type PresetSelector struct {
	presets        map[string]*schema.Preset
	dropdown       *widget.Select
	description    *widget.Label
	onPresetChange func(*schema.Preset)
}

// NewPresetSelector creates a new preset selector component.
//
// Parameters:
//   - presets: Map of available presets from schema (key → preset)
//   - onPresetChange: Callback when preset is selected (receives preset config or nil)
//
// The dropdown will include a "None (Manual Selection)" option to clear preset selection.
func NewPresetSelector(presets map[string]*schema.Preset, onPresetChange func(*schema.Preset)) *PresetSelector {
	ps := &PresetSelector{
		presets:        presets,
		onPresetChange: onPresetChange,
	}

	// Build dropdown options
	options := []string{"None (Manual Selection)"}

	// Add preset descriptions (sorted by key for consistency)
	for _, preset := range presets {
		options = append(options, preset.Description)
	}

	ps.dropdown = widget.NewSelect(options, func(selected string) {
		ps.handleSelection(selected)
	})
	ps.dropdown.PlaceHolder = "Select a preset..."

	ps.description = widget.NewLabel("")
	ps.description.Wrapping = fyne.TextWrapWord

	return ps
}

// handleSelection processes preset selection changes.
func (ps *PresetSelector) handleSelection(selectedDescription string) {
	if selectedDescription == "None (Manual Selection)" {
		ps.description.SetText("")
		if ps.onPresetChange != nil {
			ps.onPresetChange(nil)
		}
		return
	}

	// Find preset by description
	for _, preset := range ps.presets {
		if preset.Description == selectedDescription {
			ps.description.SetText(preset.Label)
			if ps.onPresetChange != nil {
				ps.onPresetChange(preset)
			}
			return
		}
	}
}

// GetContent returns the UI content for the preset selector.
//
// Returns a container with:
//   - Help text
//   - Dropdown selector
//   - Description label
func (ps *PresetSelector) GetContent() fyne.CanvasObject {
	helpText := widget.NewLabel("Select from pre-built configurations that automatically set common filter combinations for AWS profile installation")
	helpText.Wrapping = fyne.TextWrapWord

	return container.NewPadded(container.NewVBox(
		helpText,
		ps.dropdown,
		// ps.description,
	))
}

// Reset clears the preset selection.
//
// Sets the dropdown back to "None (Manual Selection)" and clears the description.
func (ps *PresetSelector) Reset() {
	ps.dropdown.SetSelected("None (Manual Selection)")
	ps.description.SetText("")
}

// SetPresets updates the available presets.
//
// This rebuilds the dropdown options with the new presets.
// Useful when the schema is reloaded.
func (ps *PresetSelector) SetPresets(presets map[string]*schema.Preset) {
	ps.presets = presets

	// Rebuild dropdown options
	options := []string{"None (Manual Selection)"}
	for _, preset := range presets {
		options = append(options, preset.Description)
	}

	ps.dropdown.Options = options
	ps.dropdown.Refresh()
	ps.Reset()
}

// GetSelectedPreset returns the currently selected preset.
//
// Returns nil if "None (Manual Selection)" is selected or no selection made.
func (ps *PresetSelector) GetSelectedPreset() *schema.Preset {
	selectedDescription := ps.dropdown.Selected

	if selectedDescription == "" || selectedDescription == "None (Manual Selection)" {
		return nil
	}

	// Find preset by description
	for _, preset := range ps.presets {
		if preset.Description == selectedDescription {
			return preset
		}
	}

	return nil
}
