package logging

// Info logs an informational message with optional key-value pairs.
//
// Info messages are displayed at info level or higher. Key-value pairs are
// formatted on separate indented lines with bullet points for better readability.
//
// Parameters:
//   - message: Info message
//   - keyValues: Optional key-value pairs for metadata (key, value, key, value, ...)
//
// Example:
//
//	logging.Log.Info("Processing started",
//		"count", 5,
//		"mode", "fast",
//	)
//	// Output:
//	// Processing started
//	//    🔹 count: 5
//	//    🔹 mode: fast
func (l *Logger) Info(message string, keyValues ...any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}

	printKeyValuePairs(l.infoColor, message, "", keyValues...)
}

// Infof logs a formatted informational message.
//
// This is the formatted string variant of Info(). Use this when you need
// string interpolation with format specifiers.
//
// Parameters:
//   - format: Format string with placeholders (%s, %d, %v, etc.)
//   - args: Values to interpolate into format string
//
// Example:
//
//	logging.Log.Infof("Processing %d files", count)
//	// Output: "Processing 5 files"
func (l *Logger) Infof(format string, args ...any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}

	printColoredMessage(l.infoColor, l.valueColor, format, args...)
}

// InfoWithDetails logs an info message with detailed explanation.
//
// This method provides a two-part info display: a title followed by
// indented details. Useful for informational messages that need context.
//
// Parameters:
//   - title: Info title message
//   - details: Detailed explanation (string or any value)
//
// Example:
//
//	logging.Log.InfoWithDetails("Configuration loaded", "path: /etc/config")
//	// Output:
//	// Configuration loaded:
//	//    🔹 path: /etc/config
func (l *Logger) InfoWithDetails(title string, details any) {
	if isSilenced() || !shouldLog(l, LogLevelInfo) {
		return
	}
	l.infoColor.Printf("%s:\n\t🔹 %v\n", title, details)
}
