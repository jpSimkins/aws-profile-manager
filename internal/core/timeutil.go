package core

import (
	"fmt"
	"time"
)

// Time formatting constants for consistent display across the application.
//
// These constants define standard time formats used throughout the application
// to ensure consistent time display in logs, UI, and output.
const (
	DateFormat                     = "2006-01-02"                  // Standard date format (YYYY-MM-DD)
	TimeFormat                     = "15:04:05"                    // Standard time format (HH:MM:SS)
	DateTimeFormat                 = "2006-01-02 15:04:05"         // Standard datetime format (YYYY-MM-DD HH:MM:SS)
	DateTimeWithZoneFormat         = "2006-01-02 15:04:05 MST"     // Datetime with timezone (YYYY-MM-DD HH:MM:SS MST)
	FriendlyDateTimeFormat         = "Mon Jan 2, 2006 3:04 PM"     // User-friendly datetime (Mon Jan 2, 2006 3:04 PM)
	FriendlyDateTimeWithZoneFormat = "Mon Jan 2, 2006 3:04 PM MST" // Friendly datetime with timezone
	RFC3339Local                   = "2006-01-02T15:04:05-07:00"   // RFC3339 format in local timezone
)

// FormatDate formats a time as date only.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Formatted date string (YYYY-MM-DD)
//
// Example:
//
//	date := core.FormatDate(time.Now())  // "2025-10-27"
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatTime formats a time as time only.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Formatted time string (HH:MM:SS)
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// FormatDateTime formats a time as datetime.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Formatted datetime string (YYYY-MM-DD HH:MM:SS)
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// FormatDateTimeWithZone formats a time with timezone.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Formatted datetime string with timezone (YYYY-MM-DD HH:MM:SS MST)
func FormatDateTimeWithZone(t time.Time) string {
	return t.Format(DateTimeWithZoneFormat)
}

// FormatFriendlyDateTime formats a time in user-friendly format.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Friendly datetime string (Mon Jan 2, 2006 3:04 PM)
func FormatFriendlyDateTime(t time.Time) string {
	return t.Format(FriendlyDateTimeFormat)
}

// FormatFriendlyDateTimeWithZone formats a time in friendly format with timezone.
//
// Parameters:
//   - t: Time to format
//
// Returns:
//   - string: Friendly datetime string with timezone (Mon Jan 2, 2006 3:04 PM MST)
func FormatFriendlyDateTimeWithZone(t time.Time) string {
	return t.Format(FriendlyDateTimeWithZoneFormat)
}

// ParseRFC3339 parses an RFC3339 formatted string and returns local time.
//
// This function is useful for parsing build dates and other timestamps that
// are stored in RFC3339 format.
//
// Parameters:
//   - s: RFC3339 formatted string (e.g., "2025-10-27T14:29:21Z")
//
// Returns:
//   - time.Time: Parsed time in local timezone
//   - error: Any error encountered during parsing
func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.Local(), nil
}

// ParseRFC3339OrNow parses an RFC3339 string, returning current time on error.
//
// This function provides a safe parsing alternative that never fails, returning
// the current time if parsing fails.
//
// Parameters:
//   - s: RFC3339 formatted string
//
// Returns:
//   - time.Time: Parsed time in local timezone, or current time if parsing fails
func ParseRFC3339OrNow(s string) time.Time {
	t, err := ParseRFC3339(s)
	if err != nil {
		return time.Now()
	}
	return t
}

// FormatBuildDate formats a build date string to friendly local time.
//
// This function converts build dates (typically from ldflags in RFC3339 format)
// to a user-friendly display format in local timezone.
//
// Parameters:
//   - buildDate: Build date string in RFC3339 format (e.g., "2025-10-27T14:29:21Z")
//
// Returns:
//   - string: Friendly formatted date with timezone (e.g., "Mon Oct 27, 2025 10:29 AM EDT")
//     or original string if parsing fails, or empty string if input is empty
func FormatBuildDate(buildDate string) string {
	if buildDate == "" {
		return ""
	}

	t, err := ParseRFC3339(buildDate)
	if err != nil {
		// If parsing fails, return the original string
		return buildDate
	}

	return FormatFriendlyDateTimeWithZone(t)
}

// FormatDuration formats a duration in human-readable format.
//
// This function converts Go durations into friendly strings suitable for
// displaying to users, automatically choosing appropriate units.
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Human-readable duration string
//
// Examples:
//   - 45 seconds: "45s"
//   - 5 minutes 30 seconds: "5m 30s"
//   - 2 hours 30 minutes: "2h 30m"
//   - 3 days 2 hours: "3d 2h"
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Minute).String()
	}

	// For durations >= 24 hours, show days
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours == 0 {
		return formatPlural(days, "day")
	}
	return formatPlural(days, "day") + " " + formatPlural(hours, "hour")
}

// FormatRelativeTime formats a time relative to the current time.
//
// This function creates user-friendly relative time strings, automatically
// handling both past and future times.
//
// Parameters:
//   - t: Time to format relative to now
//
// Returns:
//   - string: Relative time string
//
// Examples:
//   - Recent past: "just now", "2 minutes ago", "3 hours ago", "yesterday"
//   - Recent future: "in a few seconds", "in 5 minutes", "in 2 hours"
//   - Distant: "3 days ago", "2 weeks ago", "5 months ago", "1 year ago"
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		// Future time
		diff = -diff
		if diff < time.Minute {
			return "in a few seconds"
		}
		if diff < time.Hour {
			minutes := int(diff.Round(time.Minute).Minutes())
			return "in " + formatPlural(minutes, "minute")
		}
		if diff < 24*time.Hour {
			hours := int(diff.Round(time.Hour).Hours())
			return "in " + formatPlural(hours, "hour")
		}
		days := int(diff.Round(24*time.Hour).Hours() / 24)
		return "in " + formatPlural(days, "day")
	}

	// Past time
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		minutes := int(diff.Round(time.Minute).Minutes())
		return formatPlural(minutes, "minute") + " ago"
	}
	if diff < 24*time.Hour {
		hours := int(diff.Round(time.Hour).Hours())
		return formatPlural(hours, "hour") + " ago"
	}
	if diff < 48*time.Hour {
		return "yesterday"
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return formatPlural(days, "day") + " ago"
	}
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		return formatPlural(weeks, "week") + " ago"
	}
	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		return formatPlural(months, "month") + " ago"
	}
	years := int(diff.Hours() / 24 / 365)
	return formatPlural(years, "year") + " ago"
}

// formatPlural formats a number with appropriate singular/plural suffix.
//
// This helper function adds proper pluralization to time units, making
// relative time strings grammatically correct.
//
// Parameters:
//   - n: Number of units
//   - unit: Unit name (singular form)
//
// Returns:
//   - string: Formatted string with correct singular/plural form
//
// Examples:
//   - formatPlural(1, "minute") → "1 minute"
//   - formatPlural(5, "minute") → "5 minutes"
func formatPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d", n) + " " + unit + "s"
}
