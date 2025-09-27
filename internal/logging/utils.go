package logging

import (
	"os"
	"strings"

	"github.com/fatih/color"
)

// logLevelPriority defines log level filtering priorities.
//
// Lower numbers have higher priority. When a log level is set, only messages
// with equal or higher priority will be displayed. "silent" sits above all
// real levels so no message ever satisfies the threshold.
var logLevelPriority = map[string]int{
	LogLevelDebug:  0, // Lowest priority - shows everything
	LogLevelInfo:   1,
	LogLevelWarn:   2,
	LogLevelError:  3,
	LogLevelSilent: 4, // Highest priority - suppresses all output
}

// isSilenced checks if logging is silenced via environment variable.
//
// Used primarily for testing to suppress log output. Set
// AWS_PROFILE_MANAGER_SILENCE_LOGGER=1 or AWS_PROFILE_MANAGER_SILENCE_LOGGER=true
// to silence all logging.
//
// Returns:
//   - bool: true if logging is silenced, false otherwise
func isSilenced() bool {
	envValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	if envValue == "" {
		return false
	}
	// Support true/false and 1/0 (case-insensitive)
	envValue = strings.ToLower(strings.TrimSpace(envValue))
	return envValue == "true" || envValue == "1"
}

// shouldLog checks if a message should be logged based on current log level.
//
// This internal method compares the message's log level priority with the
// current logger level to determine if the message should be displayed.
//
// Parameters:
//   - logger: Logger instance with current log level
//   - messageLevel: Log level of the message to check
//
// Returns:
//   - bool: true if message should be logged, false if filtered out
func shouldLog(logger *Logger, messageLevel string) bool {
	logger.mu.RLock()
	defer logger.mu.RUnlock()

	currentPriority, exists := logLevelPriority[logger.currentLevel]
	if !exists {
		return true // If unknown level, allow logging
	}

	messagePriority, exists := logLevelPriority[messageLevel]
	if !exists {
		return true // If unknown message level, allow logging
	}

	return messagePriority >= currentPriority
}

// printKeyValuePairs prints key-value pairs with hierarchical indentation.
//
// This function handles the beautiful formatting used across all log levels,
// with automatic indentation detection and bullet-point display.
//
// Parameters:
//   - msgColor: Color for the bullet point prefix
//   - message: The main message (may contain leading tabs for indentation)
//   - prefix: Emoji prefix (e.g., "🐞 ", "" for regular logs)
//   - keyValues: Optional key-value pairs (key, value, key, value, ...)
func printKeyValuePairs(msgColor *color.Color, message string, prefix string, keyValues ...any) {
	// Count leading tabs in the message
	tabCount := 0
	for _, char := range message {
		if char == '\t' {
			tabCount++
		} else {
			break
		}
	}

	// Print the main message
	msgColor.Printf("%s%s\n", prefix, message)

	// Print values as indented key-value pairs
	if len(keyValues) > 0 {
		// Add one more level of indentation for the key-value pairs
		indent := strings.Repeat("\t", tabCount+1)
		valueColor := color.New(color.FgHiWhite)

		for i := 0; i < len(keyValues); i += 2 {
			// Print key with 🔹 prefix and proper indentation
			msgColor.Print(prefix + indent + "🔹 ")

			// Print the key-value pair
			if i+1 < len(keyValues) {
				valueColor.Printf("%v: %v\n", keyValues[i], keyValues[i+1])
			} else {
				// Odd number of values, just print the last one
				valueColor.Printf("%v\n", keyValues[i])
			}
		}
	}
}
