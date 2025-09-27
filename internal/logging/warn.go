package logging

// Warn logs a warning message with optional key-value pairs.
//
// Warning messages are displayed at warn level or higher. Key-value pairs are
// formatted on separate indented lines with bullet points.
//
// Parameters:
//   - message: Warning message
//   - keyValues: Optional key-value pairs for metadata (key, value, key, value, ...)
//
// Example:
//
//	logging.Log.Warn("Deprecated feature used",
//		"feature", "oldMethod",
//		"use", "newMethod",
//	)
//	// Output:
//	// Deprecated feature used
//	//    🔹 feature: oldMethod
//	//    🔹 use: newMethod
func (l *Logger) Warn(message string, keyValues ...any) {
	if isSilenced() || !shouldLog(l, LogLevelWarn) {
		return
	}

	printKeyValuePairs(l.warnColor, message, "", keyValues...)
}

// Warnf logs a formatted warning message.
//
// This is the formatted string variant of Warn(). Use this when you need
// string interpolation with format specifiers.
//
// Parameters:
//   - format: Format string with placeholders (%s, %d, %v, etc.)
//   - args: Values to interpolate into format string
//
// Example:
//
//	logging.Log.Warnf("Failed to process %s", filename)
//	// Output: "Failed to process config.json"
func (l *Logger) Warnf(format string, args ...any) {
	if isSilenced() || !shouldLog(l, LogLevelWarn) {
		return
	}

	printColoredMessage(l.warnColor, l.valueColor, format, args...)
}
