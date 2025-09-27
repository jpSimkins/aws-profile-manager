package settings

import (
	"fmt"
)

// SectionHeaderText is the default text used in managed-section markers.
//
// Changing this constant updates the default value everywhere — the
// GetDefaultApplication defaults, schema field defaults, and godoc examples
// all derive from this single source.
const SectionHeaderText = "Managed by AWS Profile Manager"

// ApplicationSettings holds application-wide configuration.
//
// This section contains settings for managed section markers and metadata
// options used when generating AWS config files.
type ApplicationSettings struct {
	ManagedSectionStart string `json:"managed_section_start"` // Start marker for managed section (e.g., "START - Managed by...")
	ManagedSectionEnd   string `json:"managed_section_end"`   // End marker for managed section (e.g., "END - Managed by...")
	IncludeTimestamp    bool   `json:"include_timestamp"`     // Include generation timestamp in generated configs
	IncludeVersion      bool   `json:"include_version"`       // Include schema version in generated configs
}

// GetFormattedStartMarker returns the start marker formatted for AWS config files.
//
// This adds the comment prefix (#) to the marker text, creating the actual
// marker that appears in AWS config files.
//
// Returns:
//   - string: Formatted start marker (e.g., "# START - " + SectionHeaderText)
func (s *ApplicationSettings) GetFormattedStartMarker() string {
	return "# " + s.ManagedSectionStart
}

// GetFormattedEndMarker returns the end marker formatted for AWS config files.
//
// This adds the comment prefix (#) to the marker text, creating the actual
// marker that appears in AWS config files.
//
// Returns:
//   - string: Formatted end marker (e.g., "# END - " + SectionHeaderText)
func (s *ApplicationSettings) GetFormattedEndMarker() string {
	return "# " + s.ManagedSectionEnd
}

// GetDefaultApplication returns the runtime default for application settings.
//
// This is consumed by GetDefaults() to initialize the live *Settings struct
// when no settings.json exists. It is NOT the same as the Default values in
// GetSchema() — those are UI metadata shown in the settings form. Both must
// be kept in sync; use SectionHeaderText to ensure they share a single source.
//
// Returns:
//   - ApplicationSettings: Settings with default marker text and metadata options enabled
func GetDefaultApplication() ApplicationSettings {
	return ApplicationSettings{
		ManagedSectionStart: "START - " + SectionHeaderText,
		ManagedSectionEnd:   "END - " + SectionHeaderText,
		IncludeTimestamp:    true,
		IncludeVersion:      true,
	}
}

// Validate validates application settings.
//
// Validation Rules:
//   - ManagedSectionStart must not be empty
//   - ManagedSectionEnd must not be empty
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *ApplicationSettings) Validate() error {
	// Validate managed section markers
	if s.ManagedSectionStart == "" {
		return fmt.Errorf("managed_section_start cannot be empty")
	}
	if s.ManagedSectionEnd == "" {
		return fmt.Errorf("managed_section_end cannot be empty")
	}
	if s.ManagedSectionStart == s.ManagedSectionEnd {
		return fmt.Errorf("managed_section_start and managed_section_end must be different")
	}

	return nil
}

// GetSchema returns the schema definition for application settings.
//
// The Default value in each FieldSchema is UI metadata used by the settings
// form renderer — it is NOT the runtime default. Runtime defaults live in
// GetDefaultApplication(). Both must stay in sync; use SectionHeaderText.
//
// Returns:
//   - Schema: Field schema definitions for all application settings
func (s *ApplicationSettings) GetSchema() Schema {
	return Schema{
		Version:     "1.0",
		Description: "Controls the markers used to identify the managed section inside ~/.aws/config.",
		Fields: map[string]FieldSchema{
			"managed_section_start": {
				Type:        "string",
				Label:       "Managed Section Start Marker",
				Description: "Start marker for company-managed section in AWS config file (# will be added automatically)",
				Required:    true,
				Default:     "START - " + SectionHeaderText,
				Order:       1,
				Group:       "Managed Section Markers",
			},
			"managed_section_end": {
				Type:        "string",
				Label:       "Managed Section End Marker",
				Description: "End marker for company-managed section in AWS config file (# will be added automatically)",
				Required:    true,
				Default:     "END - " + SectionHeaderText,
				Order:       2,
				Group:       "Managed Section Markers",
			},
			"include_timestamp": {
				Type:        "bool",
				Label:       "Include Timestamp",
				Description: "Include generation timestamp in config file",
				Required:    true,
				Default:     true,
				Order:       3,
				Group:       "Metadata Settings",
			},
			"include_version": {
				Type:        "bool",
				Label:       "Include Version",
				Description: "Include schema version in config file",
				Required:    true,
				Default:     true,
				Order:       4,
				Group:       "Metadata Settings",
			},
		},
	}
}
