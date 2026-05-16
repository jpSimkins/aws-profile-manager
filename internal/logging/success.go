package logging

// Success logs a success message with optional key-value pairs.
//
// Success messages are displayed at info level or higher. Used to indicate
// successful completion of operations.
//
// Parameters:
//   - message: Success message
//   - keyValues: Optional key-value pairs for metadata (key, value, key, value, ...)
//
// Example:
//
//	logging.Log.Success("Operation completed",
//		"duration", "2s",
//		"files", 10,
//	)
//	// Output:
//	// Operation completed
//	//    🔹 duration: 2s
//	//    🔹 files: 10
func (l *Logger) Success(message string, keyValues ...any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}

	printKeyValuePairs(l.successColor, message, "", keyValues...)
}

// Successf logs a formatted success message.
//
// This is the formatted string variant of Success(). Use this when you need
// string interpolation with format specifiers.
//
// Parameters:
//   - format: Format string with placeholders (%s, %d, %v, etc.)
//   - args: Values to interpolate into format string
//
// Example:
//
//	logging.Log.Successf("Completed in %d seconds", duration)
//	// Output: "Completed in 5 seconds"
func (l *Logger) Successf(format string, args ...any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}

	printColoredMessage(l.successColor, l.valueColor, format, args...)
}

// SuccessWithDetails logs a success message with detailed explanation.
//
// This method provides a two-part success display: a title followed by
// indented details. Useful for success messages that need context.
//
// Parameters:
//   - title: Success title message
//   - details: Detailed explanation (string or any value)
//
// Example:
//
//	logging.Log.SuccessWithDetails("Deployment completed", "version: 1.2.3")
//	// Output:
//	// Deployment completed:
//	//    🔹 version: 1.2.3
func (l *Logger) SuccessWithDetails(title string, details any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}
	l.successColor.Printf("%s:\n\t🔹 %v\n", title, details)
}
