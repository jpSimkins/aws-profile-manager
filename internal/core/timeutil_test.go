package core

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	testTime := time.Date(2025, 10, 27, 14, 30, 45, 0, time.UTC)
	result := FormatDate(testTime)
	expected := "2025-10-27"

	if result != expected {
		t.Errorf("FormatDate() = %q, want %q", result, expected)
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2025, 10, 27, 14, 30, 45, 0, time.UTC)
	result := FormatTime(testTime)
	expected := "14:30:45"

	if result != expected {
		t.Errorf("FormatTime() = %q, want %q", result, expected)
	}
}

func TestFormatDateTime(t *testing.T) {
	testTime := time.Date(2025, 10, 27, 14, 30, 45, 0, time.UTC)
	result := FormatDateTime(testTime)
	expected := "2025-10-27 14:30:45"

	if result != expected {
		t.Errorf("FormatDateTime() = %q, want %q", result, expected)
	}
}

func TestParseRFC3339(t *testing.T) {
	input := "2025-10-27T14:29:21Z"
	result, err := ParseRFC3339(input)

	if err != nil {
		t.Fatalf("ParseRFC3339() error = %v", err)
	}

	// Should be in local time zone
	if result.Location() != time.Local {
		t.Errorf("ParseRFC3339() location = %v, want Local", result.Location())
	}

	// Check the time values (in UTC for comparison)
	utcTime := result.UTC()
	if utcTime.Year() != 2025 || utcTime.Month() != 10 || utcTime.Day() != 27 {
		t.Errorf("ParseRFC3339() date = %v-%v-%v, want 2025-10-27", utcTime.Year(), utcTime.Month(), utcTime.Day())
	}
}

func TestParseRFC3339OrNow(t *testing.T) {
	// Valid input
	input := "2025-10-27T14:29:21Z"
	result := ParseRFC3339OrNow(input)

	if result.Year() != 2025 {
		t.Errorf("ParseRFC3339OrNow() with valid input, year = %d, want 2025", result.Year())
	}

	// Invalid input - should return current time
	before := time.Now()
	result = ParseRFC3339OrNow("invalid")
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Errorf("ParseRFC3339OrNow() with invalid input should return current time")
	}
}

func TestFormatBuildDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDate string // Check for date portion
	}{
		{
			name:     "Valid RFC3339",
			input:    "2025-10-27T14:29:21Z",
			wantDate: "Oct 27, 2025",
		},
		{
			name:     "Empty string",
			input:    "",
			wantDate: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBuildDate(tt.input)

			if tt.wantDate == "" {
				if result != "" {
					t.Errorf("FormatBuildDate(%q) = %q, want empty string", tt.input, result)
				}
				return
			}

			// For non-empty results, verify it's formatted (not the raw input)
			if result == tt.input && tt.input != "" {
				t.Errorf("FormatBuildDate(%q) = %q, should be formatted", tt.input, result)
			}

			// Check it's not empty
			if len(result) == 0 {
				t.Errorf("FormatBuildDate(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "Seconds",
			duration: 45 * time.Second,
			want:     "45s",
		},
		{
			name:     "Minutes",
			duration: 5 * time.Minute,
			want:     "5m0s",
		},
		{
			name:     "Hours",
			duration: 2*time.Hour + 30*time.Minute,
			want:     "2h30m0s",
		},
		{
			name:     "Days with hours",
			duration: 3*24*time.Hour + 2*time.Hour,
			want:     "3 days 2 hours",
		},
		{
			name:     "Days only",
			duration: 5 * 24 * time.Hour,
			want:     "5 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.want)
			}
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "Just now",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "Minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "5 minutes ago",
		},
		{
			name: "Hours ago",
			time: now.Add(-3 * time.Hour),
			want: "3 hours ago",
		},
		{
			name: "Yesterday",
			time: now.Add(-36 * time.Hour),
			want: "yesterday",
		},
		{
			name: "Days ago",
			time: now.Add(-3 * 24 * time.Hour),
			want: "3 days ago",
		},
		{
			name: "Weeks ago",
			time: now.Add(-10 * 24 * time.Hour),
			want: "1 week ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelativeTime(tt.time)
			if result != tt.want {
				t.Errorf("FormatRelativeTime() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestFormatRelativeTime_Future(t *testing.T) {
	// Use a fixed duration to avoid timing issues
	futureTime := time.Now().Add(2 * time.Hour)

	result := FormatRelativeTime(futureTime)
	expected := "in 2 hours"

	if result != expected {
		t.Errorf("FormatRelativeTime(future) = %q, want %q", result, expected)
	}
}

func TestFormatPlural(t *testing.T) {
	tests := []struct {
		n    int
		unit string
		want string
	}{
		{1, "day", "1 day"},
		{2, "day", "2 days"},
		{1, "hour", "1 hour"},
		{5, "minute", "5 minutes"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := formatPlural(tt.n, tt.unit)
			if result != tt.want {
				t.Errorf("formatPlural(%d, %q) = %q, want %q", tt.n, tt.unit, result, tt.want)
			}
		})
	}
}

func TestTimeFormatConstants(t *testing.T) {
	testTime := time.Date(2025, 10, 27, 14, 30, 45, 0, time.UTC)

	// Test that constants produce expected formats
	dateResult := testTime.Format(DateFormat)
	if dateResult != "2025-10-27" {
		t.Errorf("DateFormat constant incorrect: got %q", dateResult)
	}

	timeResult := testTime.Format(TimeFormat)
	if timeResult != "14:30:45" {
		t.Errorf("TimeFormat constant incorrect: got %q", timeResult)
	}

	dateTimeResult := testTime.Format(DateTimeFormat)
	if dateTimeResult != "2025-10-27 14:30:45" {
		t.Errorf("DateTimeFormat constant incorrect: got %q", dateTimeResult)
	}
}
