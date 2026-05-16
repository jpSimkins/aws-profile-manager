package settings

import (
	"fmt"
)

// AwsCliSettings holds AWS CLI integration configuration.
//
// This section controls how the application interacts with AWS CLI profiles
// including caching and refresh behavior.
type AwsCliSettings struct {
	AutoRefresh         bool `json:"auto_refresh"`          // Enable automatic profile cache refresh
	RefreshIntervalMins int  `json:"refresh_interval_mins"` // Minutes between cache refreshes
	ShowSsoSessions     bool `json:"show_sso_sessions"`     // Show SSO sessions panel in the Profiles view
}

// GetDefaultAwsCLI returns the runtime default for AWS CLI settings.
//
// Consumed by GetDefaults() to build the live *Settings struct on first launch.
// The Default values in GetSchema() are separate — those are UI metadata only.
//
// Returns:
//   - AwsCliSettings: Settings with auto-refresh enabled at 15-minute intervals
func GetDefaultAwsCLI() AwsCliSettings {
	return AwsCliSettings{
		AutoRefresh:         true,
		RefreshIntervalMins: 5,
		ShowSsoSessions:     true,
	}
}

// Validate validates AWS CLI settings.
//
// Validation Rules:
//   - RefreshIntervalMins must be between 1 and 1440 (24 hours)
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *AwsCliSettings) Validate() error {
	if s.RefreshIntervalMins < 1 {
		return fmt.Errorf("refresh_interval_mins must be at least 1 (got %d)", s.RefreshIntervalMins)
	}
	if s.RefreshIntervalMins > 1440 {
		return fmt.Errorf("refresh_interval_mins must be at most 1440 (24 hours) (got %d)", s.RefreshIntervalMins)
	}

	return nil
}

// GetSchema returns the schema definition for AWS CLI settings.
//
// The Default value in each FieldSchema is UI metadata for the settings form
// renderer. Runtime defaults live in GetDefaultAwsCLI() and must stay in sync.
//
// Returns:
//   - Schema: Field schema definitions for all AWS CLI settings
func (s *AwsCliSettings) GetSchema() Schema {
	minInterval := 1.0
	maxInterval := 1440.0

	return Schema{
		Version:     "1.0",
		Description: "Controls how the app interacts with the local AWS CLI configuration and credential cache.",
		Fields: map[string]FieldSchema{
			"auto_refresh": {
				Type:            "bool",
				Label:           "Auto Refresh",
				Description:     "Automatically refresh AWS CLI profiles cache",
				Required:        true,
				Default:         true,
				Order:           1,
				RequiresRestart: true,
			},
			"refresh_interval_mins": {
				Type:            "int",
				Label:           "Refresh Interval (Minutes)",
				Description:     "Minutes between automatic cache refreshes (1-1440)",
				Required:        true,
				Default:         5,
				Order:           2,
				Min:             &minInterval,
				Max:             &maxInterval,
				RequiresRestart: true,
			},
			"show_sso_sessions": {
				Type:            "bool",
				Label:           "Show SSO Sessions Panel",
				Description:     "Display the SSO sessions accordion in the Profiles view (disable if you do not use SSO)",
				Required:        true,
				Default:         true,
				Order:           3,
				RequiresRestart: true,
			},
		},
	}
}
