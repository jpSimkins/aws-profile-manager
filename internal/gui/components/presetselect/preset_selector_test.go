package presetselect

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/schema"
)

func TestNewPresetSelector(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	presets := map[string]*schema.Preset{
		"developer": {
			Label:       "Developer",
			Description: "Standard developer access",
		},
	}

	var callbackCalled bool

	ps := NewPresetSelector(presets, func(preset *schema.Preset) {
		callbackCalled = true
	})

	if ps == nil {
		t.Fatal("PresetSelector should not be nil")
	}

	if ps.dropdown == nil {
		t.Fatal("Dropdown should not be nil")
	}

	if ps.description == nil {
		t.Fatal("Description label should not be nil")
	}

	// Test initial state
	if callbackCalled {
		t.Error("Callback should not be called on initialization")
	}
}

func TestPresetSelector_GetContent(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	ps := NewPresetSelector(nil, nil)
	content := ps.GetContent()

	if content == nil {
		t.Fatal("GetContent should not return nil")
	}
}

func TestPresetSelector_Selection(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	presets := map[string]*schema.Preset{
		"developer": {
			Label:       "Developer",
			Description: "Standard developer access",
			Roles:       []string{"Developer"},
		},
		"admin": {
			Label:       "Admin",
			Description: "Administrator access",
			Roles:       []string{"Admin"},
		},
	}

	var callbackCount int
	var lastPreset *schema.Preset

	ps := NewPresetSelector(presets, func(preset *schema.Preset) {
		callbackCount++
		lastPreset = preset
	})

	// Test selecting a preset
	ps.dropdown.SetSelected("Standard developer access")

	if callbackCount != 1 {
		t.Errorf("Expected callback to be called once, got %d", callbackCount)
	}

	if lastPreset == nil {
		t.Fatal("Preset should not be nil after selection")
	}

	if lastPreset.Label != "Developer" {
		t.Errorf("Expected preset label 'Developer', got '%s'", lastPreset.Label)
	}

	if len(lastPreset.Roles) != 1 || lastPreset.Roles[0] != "Developer" {
		t.Errorf("Expected roles ['Developer'], got %v", lastPreset.Roles)
	}

	if ps.description.Text != "Developer" {
		t.Errorf("Expected description label 'Developer', got '%s'", ps.description.Text)
	}

	// Test selecting none
	ps.dropdown.SetSelected("None (Manual Selection)")

	if callbackCount != 2 {
		t.Errorf("Expected callback to be called twice, got %d", callbackCount)
	}

	if lastPreset != nil {
		t.Error("Preset should be nil after selecting 'None'")
	}

	if ps.description.Text != "" {
		t.Errorf("Description should be empty after selecting 'None', got '%s'", ps.description.Text)
	}
}

func TestPresetSelector_Reset(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	presets := map[string]*schema.Preset{
		"developer": {
			Label:       "Developer",
			Description: "Standard developer access",
		},
	}

	ps := NewPresetSelector(presets, nil)

	// Select a preset
	ps.dropdown.SetSelected("Standard developer access")

	if ps.dropdown.Selected != "Standard developer access" {
		t.Error("Preset should be selected")
	}

	if ps.description.Text == "" {
		t.Error("Description should not be empty")
	}

	// Reset
	ps.Reset()

	if ps.dropdown.Selected != "None (Manual Selection)" {
		t.Errorf("Expected 'None (Manual Selection)', got '%s'", ps.dropdown.Selected)
	}

	if ps.description.Text != "" {
		t.Errorf("Description should be empty after reset, got '%s'", ps.description.Text)
	}
}

func TestPresetSelector_SetPresets(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	initialPresets := map[string]*schema.Preset{
		"developer": {
			Label: "Developer",
		},
	}

	ps := NewPresetSelector(initialPresets, nil)

	// Verify initial options
	initialOptions := len(ps.dropdown.Options)
	if initialOptions != 2 { // "None" + 1 preset
		t.Errorf("Expected 2 initial options, got %d", initialOptions)
	}

	// Update presets
	newPresets := map[string]*schema.Preset{
		"developer": {
			Label: "Developer",
		},
		"admin": {
			Label: "Admin",
		},
		"readonly": {
			Label: "Read Only",
		},
	}

	ps.SetPresets(newPresets)

	// Verify updated options
	newOptions := len(ps.dropdown.Options)
	if newOptions != 4 { // "None" + 3 presets
		t.Errorf("Expected 4 options after update, got %d", newOptions)
	}

	// Verify reset was called (selection should be "None")
	if ps.dropdown.Selected != "None (Manual Selection)" {
		t.Errorf("Expected 'None (Manual Selection)' after SetPresets, got '%s'", ps.dropdown.Selected)
	}
}

func TestPresetSelector_GetSelectedPreset(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	presets := map[string]*schema.Preset{
		"developer": {
			Label:       "Developer",
			Description: "Standard developer access",
			Roles:       []string{"Developer"},
		},
	}

	ps := NewPresetSelector(presets, nil)

	// Initially no selection
	if ps.GetSelectedPreset() != nil {
		t.Error("GetSelectedPreset should return nil when no selection")
	}

	// Select a preset
	ps.dropdown.SetSelected("Standard developer access")

	selectedPreset := ps.GetSelectedPreset()
	if selectedPreset == nil {
		t.Fatal("GetSelectedPreset should return preset after selection")
	}

	if selectedPreset.Label != "Developer" {
		t.Errorf("Expected label 'Developer', got '%s'", selectedPreset.Label)
	}

	// Select none
	ps.dropdown.SetSelected("None (Manual Selection)")

	if ps.GetSelectedPreset() != nil {
		t.Error("GetSelectedPreset should return nil when 'None' selected")
	}
}

func TestPresetSelector_NilPresets(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	ps := NewPresetSelector(nil, nil)

	if ps == nil {
		t.Fatal("PresetSelector should not be nil even with nil presets")
	}

	// Should only have "None" option
	if len(ps.dropdown.Options) != 1 {
		t.Errorf("Expected 1 option with nil presets, got %d", len(ps.dropdown.Options))
	}

	if ps.dropdown.Options[0] != "None (Manual Selection)" {
		t.Errorf("Expected 'None (Manual Selection)', got '%s'", ps.dropdown.Options[0])
	}
}
