// Package logging provides structured logging functionality with colorized output.
//
// The logging package offers two distinct logging patterns, each designed for
// specific use cases:
//
// # Key-Value Logging (non-f functions)
//
// For logging with structured metadata. Key-value pairs are displayed with
// hierarchical indentation and bullet points:
//
//	logging.Log.Info("Processing started",
//		"count", 5,
//		"mode", "fast",
//	)
//	// Output:
//	// Processing started
//	//    🔹 count: 5
//	//    🔹 mode: fast
//
// # Formatted String Logging (f functions)
//
// For logging with string interpolation using format specifiers:
//
// logging.Log.Infof("Processing %d files in %s mode", count, mode)
// // Output: "Processing 5 files in fast mode"
//
// # Log Levels
//
// The package supports five log levels with filtering:
//   - debug: Detailed debugging information (lowest priority)
//   - info: General informational messages
//   - warn: Warning messages that need attention (default)
//   - error: Error messages
//   - silent: Suppresses all log output
//
// # Global Logger Instance
//
// The package provides a global Logger instance (Log) that should be used
// throughout the application:
//
// logging.Log.Info("Application started")
// logging.Log.Success("Operation completed")
// logging.Log.Warn("Cache miss", "key", cacheKey)
// logging.Log.Error("Failed to save", "error", err)
//
// # Testing Support
//
// Set AWS_PROFILE_MANAGER_SILENCE_LOGGER=1 to silence all logging during tests.
//
// # Debug Logging
//
// See the Debug variable for specialized debug logging that can be toggled
// independently of the main log level.
package logging

import (
	"sync"

	"github.com/fatih/color"
)

// Log level constants.
//
// These constants define the available log levels. Set the log level using
// Logger.SetLogLevel() to filter messages by priority.
const (
	LogLevelDebug  = "debug"  // Detailed debugging information
	LogLevelInfo   = "info"   // General informational messages
	LogLevelWarn   = "warn"   // Warning messages
	LogLevelError  = "error"  // Error messages
	LogLevelSilent = "silent" // Suppresses all log output
)

// Logger provides structured, colorized logging with level filtering.
//
// Logger supports both key-value logging and formatted string logging, with
// automatic colorization for different log levels and thread-safe operation.
//
// Thread Safety:
//
// All methods are thread-safe and can be called concurrently.
type Logger struct {
	errorColor   *color.Color // Color for error messages (red, bold)
	successColor *color.Color // Color for success messages (green)
	infoColor    *color.Color // Color for info messages (blue)
	valueColor   *color.Color // Color for key-value pairs (white)
	warnColor    *color.Color // Color for warning messages (yellow)
	currentLevel string       // Current log level filter
	mu           sync.RWMutex // Protects currentLevel
}

// NewLogger creates a new logger instance with default colors.
//
// The returned logger has the following color scheme:
//   - Errors: Red, bold
//   - Success: Green
//   - Info: Blue
//   - Values: White
//   - Warnings: Yellow
//
// Returns:
//   - *Logger: Configured logger instance with info level default
func NewLogger() *Logger {
	// Debug statement to confirm logger creation
	Debug.Log("Logger initialized")

	return &Logger{
		errorColor:   color.New(color.FgHiRed, color.Bold),
		successColor: color.New(color.FgHiGreen),
		infoColor:    color.New(color.FgHiBlue),
		valueColor:   color.New(color.FgHiWhite),
		warnColor:    color.New(color.FgHiYellow),
		currentLevel: LogLevelInfo, // Default to info level
	}
}

// SetLogLevel sets the current log level for filtering.
//
// Messages with lower priority than the set level will be suppressed.
// Thread-safe operation.
//
// Parameters:
//   - level: Log level (use LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError)
func (l *Logger) SetLogLevel(level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentLevel = level
}

// GetLogLevel returns the current log level.
//
// Thread-safe read operation.
//
// Returns:
//   - string: Current log level
func (l *Logger) GetLogLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.currentLevel
}

// Global logger instance and initialization state.
var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// GetLogger returns the global logger instance (singleton pattern).
//
// This function ensures only one Logger instance exists throughout the application's
// lifetime using sync.Once for thread-safe initialization.
//
// Returns:
//   - *Logger: Global logger instance
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		globalLogger = NewLogger()
	})
	return globalLogger
}

// UpdateLoggerFromSettings updates the global logger's log level from settings.
//
// This is a convenience function for updating the log level from application
// settings during initialization.
//
// Parameters:
//   - logLevel: New log level to set
func UpdateLoggerFromSettings(logLevel string) {
	Log.SetLogLevel(logLevel)
}

// Log is the global logger instance used throughout the application.
//
// This is the primary logging interface that should be used for all application
// logging. It provides both key-value logging and formatted string logging.
//
// Example Usage:
//
//	logging.Log.Info("Application started",
//		"version", "1.0.0",
//	)
//	logging.Log.Success("Operation completed",
//		"duration", "2s",
//	)
//	logging.Log.Warn("Cache miss",
//		"key", cacheKey,
//	)
//	logging.Log.Error("Failed to process",
//		"error", err,
//	)
//	logging.Log.Infof("Processing %d files", count)
var Log = GetLogger()
