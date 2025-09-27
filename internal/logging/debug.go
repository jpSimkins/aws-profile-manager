package logging

import (
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// DebugLogger handles debug-specific logging with hierarchical output.
//
// Debug logging can be enabled/disabled independently of the main log level,
// making it ideal for detailed troubleshooting without flooding production logs.
//
// Enable Methods:
//  1. AWS_PROFILE_MANAGER_DEBUG=1 environment variable (highest priority)
//  2. Settings.Logging.EnableDebug configuration
//
// Output Format:
//
//	Debug messages use hierarchical indentation and bullet points to show
//	program flow and nested operations clearly.
//
// Thread Safety:
//
//	All methods are thread-safe.
type DebugLogger struct {
	debugColor *color.Color // Color for debug messages (blue)
	enabled    bool         // Whether debug output is enabled
	mu         sync.RWMutex // Protects enabled flag
}

// Global debug logger initialization state.
var (
	globalDebug *DebugLogger
	debugOnce   sync.Once
)

// parseDebugEnv parses the AWS_PROFILE_MANAGER_DEBUG environment variable.
//
// Supports "1", "true" (case-insensitive) as enabled values.
//
// Returns:
//   - bool: true if AWS_PROFILE_MANAGER_DEBUG is set to an enabled value
func parseDebugEnv() bool {
	envValue := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	if envValue == "" {
		return false // Default to false if not set
	}

	// Convert to lowercase for case-insensitive comparison
	envValue = strings.ToLower(strings.TrimSpace(envValue))

	// Support true/false and 1/0
	return envValue == "true" || envValue == "1"
}

// isSilencedDebug checks if debug logging is silenced via environment variable.
//
// Used primarily for testing to suppress debug output.
//
// Returns:
//   - bool: true if logging is silenced
func isSilencedDebug() bool {
	envValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	if envValue == "" {
		return false
	}
	// Support true/false and 1/0 (case-insensitive)
	envValue = strings.ToLower(strings.TrimSpace(envValue))
	return envValue == "true" || envValue == "1"
}

// NewDebugLogger creates a new debug logger instance.
//
// The logger's initial enabled state is determined by the AWS_PROFILE_MANAGER_DEBUG environment
// variable.
//
// Returns:
//   - *DebugLogger: Configured debug logger instance
func NewDebugLogger() *DebugLogger {
	return &DebugLogger{
		debugColor: color.New(color.FgHiBlue),
		enabled:    parseDebugEnv(), // Check AWS_PROFILE_MANAGER_DEBUG environment variable
	}
}

// GetDebugLogger returns the global debug logger instance (singleton pattern).
//
// This function ensures only one DebugLogger instance exists throughout the
// application's lifetime.
//
// Returns:
//   - *DebugLogger: Global debug logger instance
func GetDebugLogger() *DebugLogger {
	debugOnce.Do(func() {
		globalDebug = NewDebugLogger()
	})
	return globalDebug
}

// IsEnvOverrideActive checks if AWS_PROFILE_MANAGER_DEBUG environment variable is set.
//
// When AWS_PROFILE_MANAGER_DEBUG is set, it takes precedence over settings-based debug control.
//
// Returns:
//   - bool: true if AWS_PROFILE_MANAGER_DEBUG environment variable is set
func IsEnvOverrideActive() bool {
	return os.Getenv("AWS_PROFILE_MANAGER_DEBUG") != ""
}

// SetEnabled controls debug output based on settings.
//
// Note: If AWS_PROFILE_MANAGER_DEBUG environment variable is set, it takes precedence over
// this setting. Thread-safe operation.
//
// Parameters:
//   - enabled: Whether to enable debug output
func (d *DebugLogger) SetEnabled(enabled bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If environment variable is set, it overrides settings
	if IsEnvOverrideActive() {
		d.enabled = parseDebugEnv()
	} else {
		d.enabled = enabled
	}
}

// IsEnabled returns current debug state.
//
// Thread-safe read operation.
//
// Returns:
//   - bool: true if debug output is enabled
func (d *DebugLogger) IsEnabled() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.enabled
}

// Log logs a debug message with optional key-value pairs.
//
// Debug messages use hierarchical indentation and bullet points to show
// program flow. Indentation level is automatically detected from leading
// tabs in the message.
//
// Parameters:
//   - message: Debug message (can include leading tabs for indentation)
//   - values: Optional key-value pairs (key, value, key, value, ...)
//
// Example:
//
//	logging.Debug.Log("Processing file",
//		"path", "/etc/config",
//	)
//	// Output:
//	// 🐞 Processing file
//	//    🔹 path: /etc/config
//
//	logging.Debug.Log("\t🔹 Details",
//		"count", 5,
//		"status", "active",
//	)
//	// Output:
//	// 	🔹 Details
//	// 	   🔹 count: 5
//	// 	   🔹 status: active
func (d *DebugLogger) Log(message string, values ...any) {
	if isSilencedDebug() || !d.IsEnabled() {
		return
	}

	// Use the shared printKeyValuePairs function with the debug prefix
	printKeyValuePairs(d.debugColor, message, "🐞 ", values...)
}

// Logf logs a formatted debug message.
//
// This is the formatted string variant of Log(). Use this when you need
// string interpolation with format specifiers.
//
// Parameters:
//   - format: Format string with placeholders (%s, %d, %v, etc.)
//   - args: Values to interpolate into format string
//
// Example:
//
//	logging.Debug.Logf("Processing %d files", count)
//	// Output: "🐞 Processing 5 files"
//
//	logging.Debug.Logf("\t🔹 Config loaded from: %s", configPath)
//	// Output: "	🔹 Config loaded from: /etc/config"
func (d *DebugLogger) Logf(format string, args ...any) {
	if isSilencedDebug() || !d.IsEnabled() {
		return
	}

	// Use the same colored message printing as the main logger
	// Print the 🐞 prefix in debug color
	d.debugColor.Print("🐞 ")

	// Use printColoredMessage for consistent colored value handling
	valueColor := color.New(color.FgHiWhite)
	printColoredMessage(d.debugColor, valueColor, format, args...)
}

// SetDebugEnabled controls debug output globally.
//
// This is a convenience function for updating the debug state without
// needing to call GetDebugLogger() directly.
//
// Parameters:
//   - enabled: Whether to enable debug output
func SetDebugEnabled(enabled bool) {
	GetDebugLogger().SetEnabled(enabled)
}

// IsDebugEnabled returns current global debug state.
//
// This is a convenience function for checking debug state without
// needing to call GetDebugLogger() directly.
//
// Returns:
//   - bool: true if debug output is enabled
func IsDebugEnabled() bool {
	return GetDebugLogger().IsEnabled()
}

// UpdateDebugFromSettings updates debug state from settings.
//
// This is a convenience function for updating the debug state during
// application initialization from settings.
//
// Parameters:
//   - enableDebug: Whether to enable debug output
func UpdateDebugFromSettings(enableDebug bool) {
	SetDebugEnabled(enableDebug)
}

// Debug is the global debug logger instance used throughout the application.
//
// This is the primary debug logging interface that should be used for all
// detailed troubleshooting output. Debug output can be enabled/disabled
// independently of the main log level.
//
// Enable Methods:
//  1. Set AWS_PROFILE_MANAGER_DEBUG=1 environment variable (highest priority)
//  2. Set Settings.Logging.EnableDebug to true in application settings
//
// Example Usage:
//
//	logging.Debug.Log("Starting operation",
//		"id", operationID,
//	)
//	logging.Debug.Logf("Processing %d items", count)
//	logging.Debug.Log("\t🔹 Details",
//		"duration", "2s",
//		"status", "success",
//	)
var Debug = GetDebugLogger()
