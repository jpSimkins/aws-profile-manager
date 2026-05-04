// Package core provides core utilities for the application.
package core

import (
	"context"
	"os"
	"sync"
	"time"

	"aws-profile-manager/internal/logging"
)

// throttleDuration is the parsed duration from AWS_PROFILE_MANAGER_THROTTLE.
// Zero duration means no throttling.
var throttleDuration time.Duration

// throttleOnce ensures throttle is only initialized once
var throttleOnce = &sync.Once{}

// initThrottle initializes the throttle duration from environment variable.
// Uses sync.Once to ensure it only runs once, even if called multiple times.
func initThrottle() {
	throttleOnce.Do(func() {
		throttleStr := os.Getenv("AWS_PROFILE_MANAGER_THROTTLE")
		if throttleStr == "" {
			return
		}

		duration, err := time.ParseDuration(throttleStr)
		if err != nil {
			logging.Log.Warn("Invalid AWS_PROFILE_MANAGER_THROTTLE value, ignoring",
				"value", throttleStr,
				"error", err,
			)
			return
		}

		if duration > 0 {
			throttleDuration = duration
			logging.Debug.Log("Throttle enabled",
				"duration", duration.String(),
			)
		}
	})
}

// Throttle pauses execution if AWS_PROFILE_MANAGER_THROTTLE is set.
//
// This is primarily used for manual testing of cancellation and progress reporting.
// The function respects context cancellation and returns immediately if the context
// is cancelled during the sleep.
//
// Throttle initialization is lazy - the environment variable is read on first call,
// which allows it to work with .env files loaded in main().
//
// Parameters:
//   - ctx: Context for cancellation support
//
// Returns:
//   - error: ctx.Err() if context was cancelled, nil otherwise
//
// Example:
//
//	if err := core.Throttle(ctx); err != nil {
//	    return err
//	}
func Throttle(ctx context.Context) error {
	// Lazy initialization - reads env var on first call
	initThrottle()

	if throttleDuration == 0 {
		return nil
	}

	select {
	case <-time.After(throttleDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetThrottleDuration returns the configured throttle duration.
// Returns zero duration if throttling is disabled.
// Initializes throttle on first call if not already initialized.
func GetThrottleDuration() time.Duration {
	initThrottle()
	return throttleDuration
}

// IsThrottleEnabled returns true if throttling is enabled.
// Initializes throttle on first call if not already initialized.
func IsThrottleEnabled() bool {
	initThrottle()
	return throttleDuration > 0
}
