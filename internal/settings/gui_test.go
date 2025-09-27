package settings

import (
	"strings"
	"testing"
)

func TestGetDefaultGUI(t *testing.T) {
	gui := GetDefaultGUI()

	if gui.Theme == "" {
		t.Error("Theme should not be empty")
	}
	if gui.WindowWidth != 1024 {
		t.Errorf("Expected window width 1024, got %d", gui.WindowWidth)
	}
	if gui.WindowHeight != 768 {
		t.Errorf("Expected window height 768, got %d", gui.WindowHeight)
	}
	if !gui.ShowSidebar {
		t.Error("ShowSidebar should be true by default")
	}
	if gui.DialogWidth != 600 {
		t.Errorf("Expected dialog width 600, got %d", gui.DialogWidth)
	}
	if gui.DialogHeight != 500 {
		t.Errorf("Expected dialog height 500, got %d", gui.DialogHeight)
	}
}

func TestGUIValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings GUISettings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaultGUI(),
			wantErr:  false,
		},
		{
			name: "valid light theme",
			settings: GUISettings{
				Theme:        "Light",
				WindowWidth:  1920,
				WindowHeight: 1080,
				ShowSidebar:  true,
				DialogWidth:  800,
				DialogHeight: 600,
			},
			wantErr: false,
		},
		{
			name: "valid dark theme",
			settings: GUISettings{
				Theme:        "Dark",
				WindowWidth:  1024,
				WindowHeight: 768,
				ShowSidebar:  false,
				DialogWidth:  600,
				DialogHeight: 500,
			},
			wantErr: false,
		},
		{
			name: "invalid theme",
			settings: GUISettings{
				Theme:        "InvalidTheme",
				WindowWidth:  1024,
				WindowHeight: 768,
				ShowSidebar:  true,
				DialogWidth:  600,
				DialogHeight: 500,
			},
			wantErr: true,
		},
		{
			name: "window too narrow",
			settings: GUISettings{
				Theme:        "System",
				WindowWidth:  500, // Too small
				WindowHeight: 768,
				ShowSidebar:  true,
				DialogWidth:  600,
				DialogHeight: 500,
			},
			wantErr: true,
		},
		{
			name: "window too short",
			settings: GUISettings{
				Theme:        "System",
				WindowWidth:  1024,
				WindowHeight: 400, // Too small
				ShowSidebar:  true,
				DialogWidth:  600,
				DialogHeight: 500,
			},
			wantErr: true,
		},
		{
			name: "dialog too narrow",
			settings: GUISettings{
				Theme:        "System",
				WindowWidth:  1024,
				WindowHeight: 768,
				ShowSidebar:  true,
				DialogWidth:  200, // Too small
				DialogHeight: 500,
			},
			wantErr: true,
		},
		{
			name: "dialog too short",
			settings: GUISettings{
				Theme:        "System",
				WindowWidth:  1024,
				WindowHeight: 768,
				ShowSidebar:  true,
				DialogWidth:  600,
				DialogHeight: 200, // Too small
			},
			wantErr: true,
		},
		{
			name: "case insensitive theme",
			settings: GUISettings{
				Theme:        "DARK", // Uppercase
				WindowWidth:  1024,
				WindowHeight: 768,
				ShowSidebar:  true,
				DialogWidth:  600,
				DialogHeight: 500,
			},
			wantErr: false, // Should accept case variations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGUIGetSchema(t *testing.T) {
	gui := GetDefaultGUI()
	schema := gui.GetSchema()

	if schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", schema.Version)
	}

	expectedFields := []string{
		"theme",
		"window_width",
		"window_height",
		"show_sidebar",
		"dialog_width",
		"dialog_height",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Verify theme field has enum
	themeField := schema.Fields["theme"]
	if len(themeField.Enum) == 0 {
		t.Error("Theme field should have enum values")
	}

	// Verify enum contains expected themes
	expectedThemes := []string{"System", "Light", "Dark"}
	enumLower := make(map[string]bool)
	for _, theme := range themeField.Enum {
		enumLower[strings.ToLower(theme)] = true
	}
	for _, expected := range expectedThemes {
		if !enumLower[strings.ToLower(expected)] {
			t.Errorf("Expected theme %s not found in enum", expected)
		}
	}

	// Verify numeric fields have min/max
	widthField := schema.Fields["window_width"]
	if widthField.Min == nil {
		t.Error("window_width should have min constraint")
	}
	if widthField.Max == nil {
		t.Error("window_width should have max constraint")
	}
}
