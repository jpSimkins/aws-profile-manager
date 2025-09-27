package task

import (
	"errors"
	"testing"
	"time"
)

// TestNoOpReporter verifies NoOpReporter does nothing (no panics).
func TestNoOpReporter(t *testing.T) {
	reporter := NoOpReporter{}

	// Should not panic
	reporter.ReportStatus("test status")
	reporter.ReportProgress(50, 100)
	reporter.ReportError(errors.New("test error"))
}

// TestChannelReporter_StatusUpdates verifies status updates are sent to channel.
func TestChannelReporter_StatusUpdates(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	// Send status
	reporter.ReportStatus("test status")

	// Receive with timeout
	select {
	case status := <-reporter.Status():
		if status != "test status" {
			t.Errorf("Expected 'test status', got: %s", status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for status update")
	}
}

// TestChannelReporter_ProgressUpdates verifies progress updates are sent to channel.
func TestChannelReporter_ProgressUpdates(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	// Send progress
	reporter.ReportProgress(50, 100)

	// Receive with timeout
	select {
	case prog := <-reporter.Progress():
		if prog.Current != 50 || prog.Total != 100 {
			t.Errorf("Expected progress 50/100, got: %d/%d", prog.Current, prog.Total)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for progress update")
	}
}

// TestChannelReporter_ErrorUpdates verifies error updates are sent to channel.
func TestChannelReporter_ErrorUpdates(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	testErr := errors.New("test error")

	// Send error
	reporter.ReportError(testErr)

	// Receive with timeout
	select {
	case err := <-reporter.Errors():
		if err.Error() != testErr.Error() {
			t.Errorf("Expected error '%v', got: %v", testErr, err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for error update")
	}
}

// TestChannelReporter_NonBlocking verifies reporter doesn't block when buffer is full.
func TestChannelReporter_NonBlocking(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	// Fill the buffer (10 items) and send one more
	for i := 0; i < 15; i++ {
		reporter.ReportStatus("test")
	}

	// Should not have blocked
	// Verify at least some messages were sent
	count := 0
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-reporter.Status():
			count++
		case <-timeout:
			if count < 1 {
				t.Fatal("Expected at least one status message")
			}
			return
		}
	}
}

// TestChannelReporter_Close verifies channels are closed properly.
func TestChannelReporter_Close(t *testing.T) {
	reporter := NewChannelReporter()

	// Close reporter
	reporter.Close()

	// Verify channels are closed
	select {
	case _, ok := <-reporter.Status():
		if ok {
			t.Error("Expected status channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout checking channel closure")
	}

	// Calling Close again should not panic
	reporter.Close()
}

// TestChannelReporter_MultipleUpdates verifies multiple updates work correctly.
func TestChannelReporter_MultipleUpdates(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	// Send multiple updates
	statuses := []string{"starting", "processing", "complete"}
	for _, status := range statuses {
		reporter.ReportStatus(status)
	}

	// Receive all updates
	received := make([]string, 0, len(statuses))
	timeout := time.After(500 * time.Millisecond)
	for len(received) < len(statuses) {
		select {
		case status := <-reporter.Status():
			received = append(received, status)
		case <-timeout:
			t.Fatalf("Timeout: received %d/%d updates", len(received), len(statuses))
		}
	}

	// Verify all received
	for i, expected := range statuses {
		if received[i] != expected {
			t.Errorf("Update %d: expected '%s', got '%s'", i, expected, received[i])
		}
	}
}

// TestCliReporter verifies CliReporter doesn't panic.
func TestCliReporter(t *testing.T) {
	reporter := CliReporter{}

	// Should not panic (will log to console in real usage)
	reporter.ReportStatus("test status")
	reporter.ReportProgress(50, 100)
	reporter.ReportProgress(50, 0) // Indeterminate progress
	reporter.ReportError(errors.New("test error"))
}
