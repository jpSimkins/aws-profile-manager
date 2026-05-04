// Package test provides test fixtures for task package.package test

// This package is for use by OTHER packages that need task test utilities
// (e.g., backup tests, installer tests, sync tests). It CANNOT be used by the
// task package itself due to import cycles.
//
// The task package has its own internal test fixtures in *_test.go files.
//
// This package provides mock reporters for testing progress reporting and
// status message capture in tests.
//
// Usage:
//
//	import tasktest "aws-profile-manager/internal/task/test"
//
//	func TestMyFeature(t *testing.T) {
//	    reporter := tasktest.NewMockReporter()
//	    result, err := mypackage.DoWork(ctx, opts, reporter)
//	    // Verify progress was reported
//	    if len(reporter.Messages) == 0 {
//	        t.Error("Expected progress messages")
//	    }
//	}
package test

import (
	"sync"

	"aws-profile-manager/internal/task"
)

// MockReporter is a test reporter that captures all status messages.
//
// Thread-safe for use in concurrent tests. Use this to verify that
// business logic properly reports progress during operations.
//
// Example:
//
//	reporter := tasktest.NewMockReporter()
//	result, err := installer.InstallProfiles(ctx, opts, reporter)
//	// Check that status was reported
//	if !reporter.HasMessage("Installing profiles") {
//	    t.Error("Expected installation status message")
//	}
type MockReporter struct {
	Messages []string     // Captured status messages in order
	mu       sync.RWMutex // Protects concurrent access
}

// NewMockReporter creates a new mock reporter for testing.
//
// Returns:
//   - *MockReporter: Ready to use mock reporter
func NewMockReporter() *MockReporter {
	return &MockReporter{
		Messages: make([]string, 0),
	}
}

// ReportStatus captures the status message.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - message: Status message to capture
func (m *MockReporter) ReportStatus(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, message)
}

// ReportProgress captures progress updates (currently ignored).
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - current: Current progress value
//   - total: Total expected value
func (m *MockReporter) ReportProgress(current, total int64) {
	// Mock implementation - could track progress if needed in future
}

// ReportError captures error messages.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - err: Error to capture
func (m *MockReporter) ReportError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, "ERROR: "+err.Error())
}

// HasMessage checks if a specific message was reported.
//
// Performs exact string matching.
//
// Parameters:
//   - message: Message to search for
//
// Returns:
//   - bool: True if message exists
func (m *MockReporter) HasMessage(message string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, msg := range m.Messages {
		if msg == message {
			return true
		}
	}
	return false
}

// HasMessageContaining checks if any message contains the substring.
//
// Useful for partial matching when exact message isn't known.
//
// Parameters:
//   - substring: Substring to search for
//
// Returns:
//   - bool: True if any message contains substring
func (m *MockReporter) HasMessageContaining(substring string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, msg := range m.Messages {
		if len(msg) >= len(substring) {
			for i := 0; i <= len(msg)-len(substring); i++ {
				if msg[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}

// MessageCount returns the total number of messages reported.
//
// Returns:
//   - int: Number of messages
func (m *MockReporter) MessageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Messages)
}

// LastMessage returns the most recent message, or empty string if none.
//
// Returns:
//   - string: Last message or empty string
func (m *MockReporter) LastMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.Messages) == 0 {
		return ""
	}
	return m.Messages[len(m.Messages)-1]
}

// Clear removes all captured messages.
//
// Useful for testing multiple operations with the same reporter.
func (m *MockReporter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = make([]string, 0)
}

// GetMessages returns a copy of all messages.
//
// Returns a copy to prevent external modification.
//
// Returns:
//   - []string: Copy of all messages
func (m *MockReporter) GetMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]string, len(m.Messages))
	copy(result, m.Messages)
	return result
}

// Ensure MockReporter implements task.Reporter interface at compile time
var _ task.Reporter = (*MockReporter)(nil)

// CountingReporter is a test reporter that only counts messages.
//
// Use this when you only care about the number of status updates,
// not their content. More lightweight than MockReporter.
//
// Example:
//
//	reporter := tasktest.NewCountingReporter()
//	result, err := sync.FetchConfig(ctx, cfg, opts, reporter)
//	if reporter.Count() < 3 {
//	    t.Error("Expected at least 3 status updates")
//	}
type CountingReporter struct {
	count int
	mu    sync.RWMutex
}

// NewCountingReporter creates a new counting reporter.
//
// Returns:
//   - *CountingReporter: Ready to use counting reporter
func NewCountingReporter() *CountingReporter {
	return &CountingReporter{}
}

// ReportStatus increments the message count.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - message: Status message (ignored, only counted)
func (c *CountingReporter) ReportStatus(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// ReportProgress does nothing (counting reporter only counts status messages).
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - current: Current progress value (ignored)
//   - total: Total expected value (ignored)
func (c *CountingReporter) ReportProgress(current, total int64) {
	// Counting reporter doesn't track progress
}

// ReportError does nothing (counting reporter only counts status messages).
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - err: Error (ignored)
func (c *CountingReporter) ReportError(err error) {
	// Counting reporter doesn't track errors
}

// Count returns the number of status messages reported.
//
// Returns:
//   - int: Message count
func (c *CountingReporter) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

// Reset resets the count to zero.
func (c *CountingReporter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count = 0
}

// Ensure CountingReporter implements task.Reporter interface at compile time
var _ task.Reporter = (*CountingReporter)(nil)

// FailingReporter is a reporter for testing error handling.
//
// This reporter can be configured to panic or capture but not process messages.
// Useful for testing error paths and recovery.
type FailingReporter struct {
	ShouldPanic bool
	Messages    []string
	mu          sync.RWMutex
}

// NewFailingReporter creates a new failing reporter.
//
// Parameters:
//   - shouldPanic: If true, ReportStatus will panic
//
// Returns:
//   - *FailingReporter: Ready to use failing reporter
func NewFailingReporter(shouldPanic bool) *FailingReporter {
	return &FailingReporter{
		ShouldPanic: shouldPanic,
		Messages:    make([]string, 0),
	}
}

// ReportStatus either panics or captures the message.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - message: Status message
func (f *FailingReporter) ReportStatus(message string) {
	if f.ShouldPanic {
		panic("reporter panic: " + message)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Messages = append(f.Messages, message)
}

// ReportProgress either panics or does nothing.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - current: Current progress value
//   - total: Total expected value
func (f *FailingReporter) ReportProgress(current, total int64) {
	if f.ShouldPanic {
		panic("reporter panic on progress")
	}
}

// ReportError either panics or captures the error.
//
// This implements the task.Reporter interface.
//
// Parameters:
//   - err: Error
func (f *FailingReporter) ReportError(err error) {
	if f.ShouldPanic {
		panic("reporter panic: " + err.Error())
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Messages = append(f.Messages, "ERROR: "+err.Error())
}

// Ensure FailingReporter implements task.Reporter interface at compile time
var _ task.Reporter = (*FailingReporter)(nil)
