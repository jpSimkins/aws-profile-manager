package settings

import (
	"fmt"
	"strings"
)

// GUISettings holds GUI-specific configuration.
//
// This section controls graphical user interface settings including theme,
// window dimensions, and dialog sizes.
type GUISettings struct {
	Theme                     string `json:"theme"`                       // UI theme
	WindowWidth               int    `json:"window_width"`                // Main window width in pixels
	WindowHeight              int    `json:"window_height"`               // Main window height in pixels
	ShowSidebar               bool   `json:"show_sidebar"`                // Whether to show sidebar
	DialogWidth               int    `json:"dialog_width"`                // Dialog width in pixels
	DialogHeight              int    `json:"dialog_height"`               // Dialog height in pixels
	FilterSearchOrganizations bool   `json:"filter_search_organizations"` // Show search in Organizations filter
	FilterSearchPartitions    bool   `json:"filter_search_partitions"`    // Show search in Partitions filter
	FilterSearchRegions       bool   `json:"filter_search_regions"`       // Show search in Regions filter
	FilterSearchRoles         bool   `json:"filter_search_roles"`         // Show search in Roles filter
	FilterSearchAccounts      bool   `json:"filter_search_accounts"`      // Show search in Accounts filter
}

// GetDefaultGUI returns the runtime default for GUI settings.
//
// Consumed by GetDefaults() to build the live *Settings struct on first launch.
// The Default values in GetSchema() are separate — those are UI metadata only.
//
// Returns:
//   - GUISettings: Settings with system theme and standard window dimensions
func GetDefaultGUI() GUISettings {
	return GUISettings{
		Theme:                     "System",
		WindowWidth:               1024,
		WindowHeight:              768,
		ShowSidebar:               true,
		DialogWidth:               600,
		DialogHeight:              500,
		FilterSearchOrganizations: false,
		FilterSearchPartitions:    false,
		FilterSearchRegions:       false,
		FilterSearchRoles:         false,
		FilterSearchAccounts:      true,
	}
}

// Validate validates GUI settings.
//
// Validation Rules:
//   - Theme must be a valid theme name (case-insensitive)
//   - Window dimensions must be positive
//   - Dialog dimensions must be positive
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *GUISettings) Validate() error {
	// Validate theme (case-insensitive since ApplyTheme uses strings.ToLower)
	validThemes := map[string]bool{
		"system":          true,
		"light":           true,
		"dark":            true,
		"ayu-dark":        true,
		"ayu-light":       true,
		"darcula":         true,
		"dracula":         true,
		"github-dark":     true,
		"github-light":    true,
		"gruvbox-dark":    true,
		"gruvbox-light":   true,
		"material-dark":   true,
		"material-light":  true,
		"monokai":         true,
		"nord":            true,
		"one-dark":        true,
		"solarized-dark":  true,
		"solarized-light": true,
		"tokyo-night":     true,
	}
	themeLower := strings.ToLower(s.Theme)
	if !validThemes[themeLower] {
		return fmt.Errorf("invalid theme: %s", s.Theme)
	}

	// Validate window dimensions
	if s.WindowWidth < 800 {
		return fmt.Errorf("window width must be at least 800 pixels (got %d)", s.WindowWidth)
	}
	if s.WindowHeight < 600 {
		return fmt.Errorf("window height must be at least 600 pixels (got %d)", s.WindowHeight)
	}

	// Validate dialog dimensions
	if s.DialogWidth < 400 {
		return fmt.Errorf("dialog width must be at least 400 pixels (got %d)", s.DialogWidth)
	}
	if s.DialogHeight < 300 {
		return fmt.Errorf("dialog height must be at least 300 pixels (got %d)", s.DialogHeight)
	}

	return nil
}

// GetSchema returns the schema definition for GUI settings.
//
// The Default value in each FieldSchema is UI metadata for the settings form
// renderer. Runtime defaults live in GetDefaultGUI() and must stay in sync.
//
// Returns:
//   - Schema: Field schema definitions for all GUI settings
func (s *GUISettings) GetSchema() Schema {
	minWidth := 800.0
	maxWidth := 3840.0
	minHeight := 600.0
	maxHeight := 2160.0
	minDialogWidth := 400.0
	maxDialogWidth := 2560.0
	minDialogHeight := 300.0
	maxDialogHeight := 1440.0

	return Schema{
		Version:     "1.0",
		Description: "Appearance and window preferences. Changes take effect immediately after saving.",
		Fields: map[string]FieldSchema{
			"theme": {
				Type:        "string",
				Label:       "Theme",
				Description: "Application theme",
				Required:    true,
				Default:     "System",
				Order:       1,
				Enum: []string{
					"System",
					"Light",
					"Dark",
					"Ayu-Dark",
					"Ayu-Light",
					"Darcula",
					"Dracula",
					"GitHub-Dark",
					"GitHub-Light",
					"Gruvbox-Dark",
					"Gruvbox-Light",
					"Material-Dark",
					"Material-Light",
					"Monokai",
					"Nord",
					"One-Dark",
					"Solarized-Dark",
					"Solarized-Light",
					"Tokyo-Night",
				},
			},
			"window_width": {
				Type:        "int",
				Label:       "Window Width",
				Description: "Window width in pixels",
				Required:    true,
				Default:     1024,
				Order:       2,
				Min:         &minWidth,
				Max:         &maxWidth,
			},
			"window_height": {
				Type:        "int",
				Label:       "Window Height",
				Description: "Window height in pixels",
				Required:    true,
				Default:     768,
				Order:       3,
				Min:         &minHeight,
				Max:         &maxHeight,
			},
			"show_sidebar": {
				Type:        "bool",
				Label:       "Show Sidebar",
				Description: "Show sidebar navigation",
				Required:    true,
				Default:     true,
				Order:       4,
			},
			"dialog_width": {
				Type:        "int",
				Label:       "Dialog Width",
				Description: "Dialog width in pixels",
				Required:    true,
				Default:     600,
				Order:       5,
				Min:         &minDialogWidth,
				Max:         &maxDialogWidth,
			},
			"dialog_height": {
				Type:        "int",
				Label:       "Dialog Height",
				Description: "Dialog height in pixels",
				Required:    true,
				Default:     500,
				Order:       6,
				Min:         &minDialogHeight,
				Max:         &maxDialogHeight,
			},
			"filter_search_organizations": {
				Type:            "bool",
				Label:           "Filter Search: Organizations",
				Description:     "Show a search box in the Organizations filter",
				Required:        true,
				Default:         false,
				Order:           7,
				RequiresRestart: true,
			},
			"filter_search_partitions": {
				Type:            "bool",
				Label:           "Filter Search: Partitions",
				Description:     "Show a search box in the Partitions filter",
				Required:        true,
				Default:         false,
				Order:           8,
				RequiresRestart: true,
			},
			"filter_search_regions": {
				Type:            "bool",
				Label:           "Filter Search: Regions",
				Description:     "Show a search box in the Regions filter",
				Required:        true,
				Default:         false,
				Order:           9,
				RequiresRestart: true,
			},
			"filter_search_roles": {
				Type:            "bool",
				Label:           "Filter Search: Roles",
				Description:     "Show a search box in the Roles filter",
				Required:        true,
				Default:         false,
				Order:           10,
				RequiresRestart: true,
			},
			"filter_search_accounts": {
				Type:            "bool",
				Label:           "Filter Search: Accounts",
				Description:     "Show a search box in the Accounts filter",
				Required:        true,
				Default:         true,
				Order:           11,
				RequiresRestart: true,
			},
		},
	}
}
