package logging

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// printColoredMessage prints a formatted message with colored values.
//
// This internal function handles format string logging with proper color
// differentiation between the message text and interpolated values.
//
// Format specifiers in the string are detected and their corresponding
// values are colored differently for visual distinction.
//
// Parameters:
//   - msgColor: Color for the main message text
//   - valueColor: Color for interpolated values
//   - format: Format string with placeholders (%s, %d, %v, etc.)
//   - args: Values to interpolate into format string
//
// Example:
//
//	printColoredMessage(infoColor, valueColor, "Processing %d files", 5)
//	// Output: "Processing " (colored) + "5" (white) + " files" (colored)
func printColoredMessage(msgColor *color.Color, valueColor *color.Color, format string, args ...any) {
	if len(args) == 0 {
		// No args, just print the message in color
		msgColor.Println(format)
		return
	}

	// Check if this is a format string (contains % formatting verbs)
	if strings.Contains(format, "%") {
		// Format string pattern: we need to parse and print with colors
		// Split the format string and inject colored values

		// Parse the format string to find %v, %s, %d, etc.
		formatIdx := 0
		argIdx := 0

		for i := 0; i < len(format); i++ {
			if format[i] == '%' && i+1 < len(format) {
				// Found a format verb
				// Print everything before this in the message color
				if i > formatIdx {
					msgColor.Print(format[formatIdx:i])
				}

				// Find the end of the format verb
				verbEnd := i + 1
				for verbEnd < len(format) && !strings.ContainsRune("vsTtbcdoOqxXUeEfFgGsp%", rune(format[verbEnd])) {
					verbEnd++
				}

				if verbEnd < len(format) {
					verb := format[i : verbEnd+1]

					if verb == "%%" {
						// Literal %
						msgColor.Print("%")
					} else if argIdx < len(args) {
						// Format and print this argument in white
						valueColor.Print(fmt.Sprintf(verb, args[argIdx]))
						argIdx++
					}

					i = verbEnd
					formatIdx = i + 1
				}
			}
		}

		// Print any remaining text in message color
		if formatIdx < len(format) {
			msgColor.Print(format[formatIdx:])
		}
		fmt.Println()
	} else {
		// Structured logging pattern: print message in color, then key=value pairs in white
		msgColor.Print(format)

		for i := 0; i < len(args); i += 2 {
			if i+1 < len(args) {
				// We have a key-value pair
				key := fmt.Sprint(args[i])
				value := fmt.Sprint(args[i+1])
				fmt.Print(" ")
				valueColor.Printf("%s=%s", key, value)
			} else {
				// Odd number of args, just append the last one
				fmt.Print(" ")
				valueColor.Print(args[i])
			}
		}
		fmt.Println()
	}
}

// formatKeyValues formats key-value pairs as "key=value key2=value2".
//
// This internal function converts variadic key-value pairs into a
// space-separated string of key=value pairs for inline display.
//
// Parameters:
//   - keyValues: Alternating keys and values (key, value, key, value, ...)
//
// Returns:
//   - string: Formatted key-value string
//
// Example:
//
//	formatKeyValues("count", 5, "status", "active")
//	// Returns: "count=5 status=active"
func formatKeyValues(keyValues ...any) string {
	var parts []string
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			key := fmt.Sprint(keyValues[i])
			value := fmt.Sprint(keyValues[i+1])
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		} else {
			// Odd number of arguments, just append the last one
			parts = append(parts, fmt.Sprint(keyValues[i]))
		}
	}
	return strings.Join(parts, " ")
}
