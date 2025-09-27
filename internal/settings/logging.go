package settings

import (
	"fmt"
)

// LoggingSettings holds logging-specific configuration.
//
// This section controls application logging behavior including log level
// and debug mode settings.
type LoggingSettings struct {
	LogLevel    string `json:"log_level"`    // Log level (debug, info, warn, error)
	EnableDebug bool   `json:"enable_debug"` // Enable debug logging
}

// GetDefaultLogging returns the runtime default for logging settings.
//
// Consumed by GetDefaults() to build the live *Settings struct on first launch.
// The Default values in GetSchema() are separate — those are UI metadata only.
//
// Returns:
//   - LoggingSettings: Settings with warn level and debug disabled
func GetDefaultLogging() LoggingSettings {
	return LoggingSettings{
		LogLevel:    "warn",
		EnableDebug: false,
	}
}

// Validate validates logging settings.
//
// Validation Rules:
//   - LogLevel must be one of: debug, info, warn, error, silent
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *LoggingSettings) Validate() error {
	// Validate log level
	validLogLevels := map[string]bool{
		"debug":  true,
		"info":   true,
		"warn":   true,
		"error":  true,
		"silent": true,
	}
	if !validLogLevels[s.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, error, or silent)", s.LogLevel)
	}

	return nil
}

// GetSchema returns the schema definition for logging settings.
//
// The Default value in each FieldSchema is UI metadata for the settings form
// renderer. Runtime defaults live in GetDefaultLogging() and must stay in sync.
//
// Returns:
//   - Schema: Field schema definitions for all logging settings
func (s *LoggingSettings) GetSchema() Schema {
	return Schema{
		Version:     "1.0",
		Description: "Controls how much output the application writes to the console.",
		Fields: map[string]FieldSchema{
			"log_level": {
				Type:  "string",
				Label: "Log Level",
				Description: "Controls which messages are printed to the console. " +
					"debug shows everything; info adds general progress messages; " +
					"warn (default) shows only warnings and errors; " +
					"error shows only errors; " +
					"silent suppresses all output.",
				Required: true,
				Default:  "warn",
				Enum:     []string{"debug", "info", "warn", "error", "silent"},
				Order:    1,
			},
			"enable_debug": {
				Type:  "bool",
				Label: "Enable Debug Trace",
				Description: "Enables low-level trace logging from the Debug logger — a separate, " +
					"more verbose channel that records internal program flow (function entry, " +
					"state transitions, cache hits, etc). " +
					"This is independent of Log Level: you can have Log Level set to 'warn' " +
					"and still see debug traces, or vice-versa. " +
					"The environment variable AWS_PROFILE_MANAGER_DEBUG=1 also enables this " +
					"and takes precedence over this setting at startup.",
				Required: true,
				Default:  false,
				Order:    2,
				HelpText: "AWS_PROFILE_MANAGER_DEBUG=1 env var overrides this at process start",
			},
		},
	}
}
