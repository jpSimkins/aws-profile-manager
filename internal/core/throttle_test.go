package core

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestThrottle_Disabled(t *testing.T) {
	// Save original state
	originalDuration := throttleDuration
	originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
	t.Cleanup(func() {
		throttleDuration = originalDuration
		throttleOnce = &sync.Once{}
		_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
	})

	// Reset and disable throttle
	throttleDuration = 0
	throttleOnce = &sync.Once{}
	_ = os.Unsetenv("AWS_PROFILE_MANAGER_THROTTLE")

	ctx := context.Background()
	start := time.Now()

	err := Throttle(ctx)

	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Throttle() returned error: %v", err)
	}

	// Should complete almost instantly (< 10ms)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Throttle took too long when disabled: %v", elapsed)
	}
}

func TestThrottle_Enabled(t *testing.T) {
	// Save original state
	originalDuration := throttleDuration
	originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
	t.Cleanup(func() {
		throttleDuration = originalDuration
		throttleOnce = &sync.Once{}
		_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
	})

	// Reset and enable throttle with short duration for testing
	throttleDuration = 0
	throttleOnce = &sync.Once{}
	_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", "50ms")

	ctx := context.Background()
	start := time.Now()

	err := Throttle(ctx)

	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Throttle() returned error: %v", err)
	}

	// Should take at least the throttle duration
	if elapsed < 50*time.Millisecond {
		t.Errorf("Throttle completed too quickly: %v", elapsed)
	}

	// Should not take significantly longer (< 2x duration)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Throttle took too long: %v", elapsed)
	}
}

func TestThrottle_Cancelled(t *testing.T) {
	// Save original state
	originalDuration := throttleDuration
	originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
	t.Cleanup(func() {
		throttleDuration = originalDuration
		throttleOnce = &sync.Once{}
		_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
	})

	// Reset and enable throttle with longer duration
	throttleDuration = 0
	throttleOnce = &sync.Once{}
	_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", "1s")

	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()

	// Cancel after 50ms
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Throttle(ctx)

	elapsed := time.Since(start)

	// Should return context.Canceled error
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	// Should complete quickly after cancellation (< 200ms)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Throttle took too long after cancellation: %v", elapsed)
	}

	// Should not complete before cancellation (>= 50ms)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Throttle completed before cancellation: %v", elapsed)
	}
}

func TestGetThrottleDuration(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     time.Duration
	}{
		{"zero", "", 0},
		{"100ms", "100ms", 100 * time.Millisecond},
		{"1s", "1s", 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalDuration := throttleDuration
			originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
			t.Cleanup(func() {
				throttleDuration = originalDuration
				throttleOnce = &sync.Once{}
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
			})

			// Reset state
			throttleDuration = 0
			throttleOnce = &sync.Once{}
			if tt.envValue != "" {
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", tt.envValue)
			} else {
				_ = os.Unsetenv("AWS_PROFILE_MANAGER_THROTTLE")
			}

			got := GetThrottleDuration()
			if got != tt.want {
				t.Errorf("GetThrottleDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsThrottleEnabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{"disabled", "", false},
		{"enabled_100ms", "100ms", true},
		{"enabled_1s", "1s", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalDuration := throttleDuration
			originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
			t.Cleanup(func() {
				throttleDuration = originalDuration
				throttleOnce = &sync.Once{}
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
			})

			// Reset state
			throttleDuration = 0
			throttleOnce = &sync.Once{}
			if tt.envValue != "" {
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", tt.envValue)
			} else {
				_ = os.Unsetenv("AWS_PROFILE_MANAGER_THROTTLE")
			}

			got := IsThrottleEnabled()
			if got != tt.want {
				t.Errorf("IsThrottleEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottle_InitFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envValue  string
		wantZero  bool
		wantError bool
	}{
		{
			name:     "not set",
			envValue: "",
			wantZero: true,
		},
		{
			name:     "valid 100ms",
			envValue: "100ms",
			wantZero: false,
		},
		{
			name:     "valid 1s",
			envValue: "1s",
			wantZero: false,
		},
		{
			name:     "valid 2m",
			envValue: "2m",
			wantZero: false,
		},
		{
			name:      "invalid",
			envValue:  "invalid",
			wantZero:  true,
			wantError: true,
		},
		{
			name:     "zero",
			envValue: "0",
			wantZero: true,
		},
		{
			name:     "negative",
			envValue: "-1s",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment and state
			originalEnv := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
			originalDuration := throttleDuration
			t.Cleanup(func() {
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", originalEnv)
				throttleDuration = originalDuration
				throttleOnce = &sync.Once{}
			})

			// Reset state for this test
			throttleDuration = 0
			throttleOnce = &sync.Once{}

			// Set test environment
			if tt.envValue != "" {
				_ = os.Setenv("AWS_PROFILE_MANAGER_THROTTLE", tt.envValue)
			} else {
				_ = os.Unsetenv("AWS_PROFILE_MANAGER_THROTTLE")
			}

			// Trigger lazy initialization
			_ = GetThrottleDuration()

			// Check result
			if tt.wantZero {
				if throttleDuration != 0 {
					t.Errorf("Expected zero duration, got: %v", throttleDuration)
				}
			} else {
				if throttleDuration == 0 {
					t.Errorf("Expected non-zero duration, got: %v", throttleDuration)
				}
			}
		})
	}
}
