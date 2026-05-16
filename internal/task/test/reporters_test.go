package test

import (
	"errors"
	"testing"
)

// TestMockReporter_ReportStatus verifies status message capturing.
func TestMockReporter_ReportStatus(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportStatus("Starting")
	reporter.ReportStatus("Processing")
	reporter.ReportStatus("Complete")

	if reporter.MessageCount() != 3 {
		t.Errorf("Expected 3 messages, got %d", reporter.MessageCount())
	}

	if !reporter.HasMessage("Starting") {
		t.Error("Expected 'Starting' message")
	}

	if !reporter.HasMessage("Complete") {
		t.Error("Expected 'Complete' message")
	}
}

// TestMockReporter_ReportError verifies error capturing.
func TestMockReporter_ReportError(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportError(errors.New("test error"))

	if reporter.MessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", reporter.MessageCount())
	}

	if !reporter.HasMessageContaining("test error") {
		t.Error("Expected error message")
	}
}

// TestMockReporter_HasMessage verifies exact message matching.
func TestMockReporter_HasMessage(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportStatus("Exact message")

	if !reporter.HasMessage("Exact message") {
		t.Error("Should find exact message")
	}

	if reporter.HasMessage("Different message") {
		t.Error("Should not find different message")
	}
}

// TestMockReporter_HasMessageContaining verifies substring matching.
func TestMockReporter_HasMessageContaining(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportStatus("Processing files...")

	if !reporter.HasMessageContaining("Processing") {
		t.Error("Should find substring")
	}

	if !reporter.HasMessageContaining("files") {
		t.Error("Should find substring")
	}

	if reporter.HasMessageContaining("complete") {
		t.Error("Should not find non-existent substring")
	}
}

// TestMockReporter_LastMessage verifies last message retrieval.
func TestMockReporter_LastMessage(t *testing.T) {
	reporter := NewMockReporter()

	if reporter.LastMessage() != "" {
		t.Error("Empty reporter should return empty string")
	}

	reporter.ReportStatus("First")
	reporter.ReportStatus("Second")
	reporter.ReportStatus("Third")

	if reporter.LastMessage() != "Third" {
		t.Errorf("Expected 'Third', got %s", reporter.LastMessage())
	}
}

// TestMockReporter_Clear verifies clearing messages.
func TestMockReporter_Clear(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportStatus("Message 1")
	reporter.ReportStatus("Message 2")

	if reporter.MessageCount() != 2 {
		t.Error("Should have 2 messages")
	}

	reporter.Clear()

	if reporter.MessageCount() != 0 {
		t.Error("Should have 0 messages after clear")
	}
}

// TestMockReporter_GetMessages verifies message copy retrieval.
func TestMockReporter_GetMessages(t *testing.T) {
	reporter := NewMockReporter()

	reporter.ReportStatus("Message 1")
	reporter.ReportStatus("Message 2")

	messages := reporter.GetMessages()

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Modifying copy shouldn't affect reporter
	messages[0] = "Modified"

	if reporter.HasMessage("Modified") {
		t.Error("Modifying copy should not affect original")
	}
}

// TestMockReporter_Concurrent verifies thread safety.
func TestMockReporter_Concurrent(t *testing.T) {
	reporter := NewMockReporter()

	// Report from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			reporter.ReportStatus("Test")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if reporter.MessageCount() != 10 {
		t.Errorf("Expected 10 messages, got %d", reporter.MessageCount())
	}
}

// TestCountingReporter_Count verifies message counting.
func TestCountingReporter_Count(t *testing.T) {
	reporter := NewCountingReporter()

	if reporter.Count() != 0 {
		t.Error("Initial count should be 0")
	}

	reporter.ReportStatus("Message 1")
	reporter.ReportStatus("Message 2")
	reporter.ReportStatus("Message 3")

	if reporter.Count() != 3 {
		t.Errorf("Expected count 3, got %d", reporter.Count())
	}
}

// TestCountingReporter_Reset verifies count reset.
func TestCountingReporter_Reset(t *testing.T) {
	reporter := NewCountingReporter()

	reporter.ReportStatus("Message")
	reporter.ReportStatus("Message")

	if reporter.Count() != 2 {
		t.Error("Should have count 2")
	}

	reporter.Reset()

	if reporter.Count() != 0 {
		t.Error("Count should be 0 after reset")
	}
}

// TestCountingReporter_IgnoresContent verifies content is ignored.
func TestCountingReporter_IgnoresContent(t *testing.T) {
	reporter := NewCountingReporter()

	reporter.ReportStatus("Any message")
	reporter.ReportProgress(50, 100)
	reporter.ReportError(errors.New("test error"))

	// Only ReportStatus increments count
	if reporter.Count() != 1 {
		t.Errorf("Expected count 1, got %d", reporter.Count())
	}
}

// TestFailingReporter_Panic verifies panic behavior.
func TestFailingReporter_Panic(t *testing.T) {
	reporter := NewFailingReporter(true)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic")
		}
	}()

	reporter.ReportStatus("This should panic")
}

// TestFailingReporter_NoPanic verifies non-panic behavior.
func TestFailingReporter_NoPanic(t *testing.T) {
	reporter := NewFailingReporter(false)

	reporter.ReportStatus("Message")
	reporter.ReportError(errors.New("test error"))

	if len(reporter.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(reporter.Messages))
	}
}
