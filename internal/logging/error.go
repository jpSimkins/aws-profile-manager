package logging

import (
	"fmt"
	"strings"
)

// Error prints an error message with optional key-value pairs and returns the error.
//
// Error messages are always displayed regardless of log level filtering, as they
// are considered critical. The function returns an error for convenient error
// propagation.
//
// Parameters:
//   - message: Error message
//   - keyValues: Optional key-value pairs for metadata (key, value, key, value, ...)
//
// Returns:
//   - error: Error object with the formatted message
//
// Example:
//
//	return logging.Log.Error("Operation failed",
//		"file", filename,
//		"reason", "not found",
//	)
//	// Output:
//	// Operation failed
//	//    🔹 file: config.json
//	//    🔹 reason: not found
func (l *Logger) Error(message string, keyValues ...any) error {
	errorMsg := message
	if len(keyValues) > 0 {
		errorMsg = message + " " + formatKeyValues(keyValues...)
	}

	// Check if logger is silenced
	if !isSilenced() {
		// Always log errors regardless of log level (errors are critical)
		printKeyValuePairs(l.errorColor, message, "", keyValues...)
	}

	return fmt.Errorf("%s", errorMsg)
}

// Errorf logs a formatted error message and returns an error.
//
// This is the formatted string variant of Error(). Supports error wrapping
// with %w verb like fmt.Errorf. Always displayed regardless of log level.
//
// Parameters:
//   - format: Format string with placeholders (%s, %d, %v, %w, etc.)
//   - args: Values to interpolate into format string
//
// Returns:
//   - error: Error object with the formatted message
//
// Example:
//
//	return logging.Log.Errorf("failed to load %s: %w", filename, err)
//	// Output: "failed to load config.json: file not found"
//	// Returns: wrapped error
func (l *Logger) Errorf(format string, args ...any) error {
	// Create the error first to preserve %w wrapping behavior
	err := fmt.Errorf(format, args...)

	// Check if logger is silenced
	if !isSilenced() {
		// Always log errors regardless of log level (errors are critical)
		// Use printColoredMessage for colored values
		printColoredMessage(l.errorColor, l.valueColor, format, args...)
	}

	return err
}

// ErrorWithDetails logs an error message with detailed explanation and returns an error.
//
// This method provides a two-part error display: a bold title followed by
// indented details. Useful for error messages that need context or explanation.
//
// Parameters:
//   - title: Error title message
//   - details: Detailed explanation (error, string, or any value)
//
// Returns:
//   - error: Error object (returns details if it's an error, otherwise creates new error)
//
// Example:
//
//	return logging.Log.ErrorWithDetails("Failed to load config", err)
//	// Output:
//	// Failed to load config:
//	//    🔹 file not found
func (l *Logger) ErrorWithDetails(title string, details any) error {
	// Check if logger is silenced - still return error but don't print
	if isSilenced() {
		if err, ok := details.(error); ok {
			return err
		}
		return fmt.Errorf("%s", title)
	}

	l.errorColor.Printf("%s:\n", title)

	// Handle error chains - split intelligently
	if err, ok := details.(error); ok {
		errorText := err.Error()
		parts := strings.Split(errorText, ": ")

		// Merge system call errors back together
		var finalParts []string
		i := 0
		for i < len(parts) {
			part := parts[i]
			// If this looks like a system call and we have a next part, combine them
			if i < len(parts)-1 &&
				(strings.HasPrefix(part, "open ") ||
					strings.HasPrefix(part, "mkdir ") ||
					strings.HasPrefix(part, "stat ")) {
				combined := part + ": " + parts[i+1]
				finalParts = append(finalParts, combined)
				i += 2 // Skip both parts
			} else {
				finalParts = append(finalParts, part)
				i++
			}
		}

		// Print each part
		for _, part := range finalParts {
			if strings.TrimSpace(part) != "" {
				fmt.Printf("\t🔹 %s\n", part)
			}
		}
	} else {
		fmt.Printf("\t🔹 %v\n", details)
	}
	fmt.Println()

	// Return the error - use the original error if details is an error, otherwise create one from the title
	if err, ok := details.(error); ok {
		return err
	}
	return fmt.Errorf("%s", title)
}

// ErrorfWithDetails logs a formatted error message with detailed explanation.
//
// This is the formatted string variant of ErrorWithDetails(). It formats
// the title before displaying it with the details.
//
// Parameters:
//   - titleFormat: Format string for error title
//   - details: Detailed explanation (error, string, or any value)
//   - args: Values to interpolate into title format
//
// Returns:
//   - error: Error object
//
// Example:
//
//	return logging.Log.ErrorfWithDetails("Failed to load %s", err, filename)
//	// Output:
//	// Failed to load config.json:
//	//    🔹 file not found
func (l *Logger) ErrorfWithDetails(titleFormat string, details any, args ...any) error {
	title := fmt.Sprintf(titleFormat, args...)
	return l.ErrorWithDetails(title, details)
}
