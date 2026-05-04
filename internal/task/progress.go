package task

import (
	"sync"

	"aws-profile-manager/internal/logging"
)

// Reporter receives real-time updates during task execution.
//
// Implementations must be safe for concurrent use if task reports
// from multiple goroutines (though most tasks report from single goroutine).
//
// Use cases:
//   - NoOpReporter: Silent execution (tests, background tasks)
//   - ChannelReporter: GUI applications (real-time updates)
//   - CliReporter: Console output (user feedback)
//   - Custom: Application-specific reporting
type Reporter interface {
	// ReportStatus updates the current operation status.
	// Examples: "Fetching from S3...", "Writing profiles...", "Complete"
	ReportStatus(status string)

	// ReportProgress updates progress with current and total values.
	//
	// Parameters:
	//   - current: Current progress value (items processed, bytes read, etc.)
	//   - total: Total expected value (0 if unknown/indeterminate)
	//
	// Example:
	//   reporter.ReportProgress(5, 10)  // 50% complete
	//   reporter.ReportProgress(100, 0) // Indeterminate (processed 100, unknown total)
	ReportProgress(current, total int64)

	// ReportError reports a non-fatal error (warning, retry, etc.).
	// Does NOT stop task execution - task continues.
	// Final error is returned by task.Execute().
	//
	// Use cases:
	//   - Network retry attempts
	//   - Cache misses (fallback in progress)
	//   - Validation warnings
	//   - Partial failures
	ReportError(err error)
}

// NoOpReporter does nothing with updates.
//
// Use cases:
//   - Unit tests (don't need progress)
//   - Background tasks (no user to show progress)
//   - Benchmarking (avoid overhead)
//
// This is the default if nil reporter passed to Run().
type NoOpReporter struct{}

// ReportStatus does nothing.
func (NoOpReporter) ReportStatus(string) {}

// ReportProgress does nothing.
func (NoOpReporter) ReportProgress(int64, int64) {}

// ReportError does nothing.
func (NoOpReporter) ReportError(error) {}

// ChannelReporter sends updates to channels for GUI consumption.
//
// Channels are buffered (size 10) to prevent blocking if consumer is slow.
// If buffer is full, updates are dropped (non-blocking).
//
// Use cases:
//   - GUI applications (Fyne, etc.)
//   - Web applications (websocket updates)
//   - Event-driven systems
//
// Usage pattern:
//  1. Create reporter with NewChannelReporter()
//  2. Start goroutine to read from Status(), Progress(), Errors()
//  3. Pass reporter to task.Run()
//  4. Call Close() when done to cleanup
type ChannelReporter struct {
	statusCh   chan string
	progressCh chan Progress
	errorCh    chan error
	once       sync.Once // Ensure Close() is safe to call multiple times
}

// NewChannelReporter creates a new channel-based reporter.
func NewChannelReporter() *ChannelReporter {
	return &ChannelReporter{
		statusCh:   make(chan string, 10),
		progressCh: make(chan Progress, 10),
		errorCh:    make(chan error, 10),
	}
}

// ReportStatus sends status update to channel (non-blocking).
func (c *ChannelReporter) ReportStatus(status string) {
	select {
	case c.statusCh <- status:
	default: // Don't block if buffer full
	}
}

// ReportProgress sends progress update to channel (non-blocking).
func (c *ChannelReporter) ReportProgress(current, total int64) {
	select {
	case c.progressCh <- Progress{Current: current, Total: total}:
	default:
	}
}

// ReportError sends error to channel (non-blocking).
func (c *ChannelReporter) ReportError(err error) {
	select {
	case c.errorCh <- err:
	default:
	}
}

// Status returns the read-only status channel.
func (c *ChannelReporter) Status() <-chan string {
	return c.statusCh
}

// Progress returns the read-only progress channel.
func (c *ChannelReporter) Progress() <-chan Progress {
	return c.progressCh
}

// Errors returns the read-only error channel.
func (c *ChannelReporter) Errors() <-chan error {
	return c.errorCh
}

// Close closes all channels. Safe to call multiple times.
func (c *ChannelReporter) Close() {
	c.once.Do(func() {
		close(c.statusCh)
		close(c.progressCh)
		close(c.errorCh)
	})
}

// CliReporter prints updates to console using the logging package.
//
// Use cases:
//   - CLI commands
//   - Console applications
//   - Debug output
type CliReporter struct{}

// ReportStatus logs status message.
func (CliReporter) ReportStatus(status string) {
	logging.Log.Info(status)
}

// ReportProgress logs progress update.
func (CliReporter) ReportProgress(current, total int64) {
	if total > 0 {
		pct := float64(current) / float64(total) * 100
		logging.Log.Infof("Progress: %.1f%% (%d/%d)", pct, current, total)
	} else {
		logging.Log.Infof("Progress: %d items processed", current)
	}
}

// ReportError logs non-fatal error as warning.
func (CliReporter) ReportError(err error) {
	logging.Log.Warn("Non-fatal error",
		"error", err,
		"note", "Continuing execution",
	)
}
