package task

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestFunctionTask_Success verifies successful function execution.
func TestFunctionTask_Success(t *testing.T) {
	task := &FunctionTask{
		Name: "test-function",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			r.ReportStatus("Working...")
			r.ReportProgress(50, 100)
			return []byte("success"), nil
		},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
	if string(result.Output) != "success" {
		t.Errorf("Expected 'success', got: %s", result.Output)
	}
	if result.Canceled {
		t.Error("Expected result.Canceled to be false")
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// TestFunctionTask_Failure verifies function error handling.
func TestFunctionTask_Failure(t *testing.T) {
	testErr := errors.New("test error")

	task := &FunctionTask{
		Name: "failing-function",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			return nil, testErr
		},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err == nil {
		t.Fatal("Expected error from function")
	}
	if err != testErr {
		t.Errorf("Expected specific error, got: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got: %d", result.ExitCode)
	}
}

// TestFunctionTask_ContextCancellation verifies context cancellation stops function.
func TestFunctionTask_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	task := &FunctionTask{
		Name: "long-running",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			for i := 0; i < 100; i++ {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}
				time.Sleep(10 * time.Millisecond)
			}
			return []byte("done"), nil
		},
	}

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, err := Run(ctx, task, NoOpReporter{})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected cancellation error")
	}
	if !result.Canceled {
		t.Error("Expected result.Canceled to be true")
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("Function took too long to cancel: %s", elapsed)
	}
}

// TestFunctionTask_ProgressReporting verifies progress is reported.
func TestFunctionTask_ProgressReporting(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	task := &FunctionTask{
		Name: "progress-test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			r.ReportStatus("step 1")
			r.ReportProgress(1, 3)
			r.ReportStatus("step 2")
			r.ReportProgress(2, 3)
			r.ReportStatus("step 3")
			r.ReportProgress(3, 3)
			return []byte("done"), nil
		},
	}

	// Collect updates
	statuses := []string{}
	progresses := []Progress{}

	// Start collector goroutine
	done := make(chan struct{})
	go func() {
		timeout := time.After(1 * time.Second)
		for {
			select {
			case status, ok := <-reporter.Status():
				if !ok {
					return
				}
				statuses = append(statuses, status)
			case prog, ok := <-reporter.Progress():
				if !ok {
					return
				}
				progresses = append(progresses, prog)
			case <-timeout:
				return
			case <-done:
				return
			}
		}
	}()

	// Run task
	_, err := Run(context.Background(), task, reporter)
	if err != nil {
		t.Fatalf("Task failed: %v", err)
	}

	// Wait a bit for updates to be collected
	time.Sleep(100 * time.Millisecond)
	close(done)

	// Verify updates were received
	if len(statuses) < 3 {
		t.Errorf("Expected at least 3 status updates, got %d", len(statuses))
	}
	if len(progresses) < 3 {
		t.Errorf("Expected at least 3 progress updates, got %d", len(progresses))
	}
}

// TestFunctionTask_String verifies String() method.
func TestFunctionTask_String(t *testing.T) {
	task := &FunctionTask{
		Name: "my-function",
		Fn:   func(ctx context.Context, r Reporter) ([]byte, error) { return nil, nil },
	}

	str := task.String()
	if !strings.Contains(str, "function") {
		t.Errorf("Expected 'function' in string, got: %s", str)
	}
	if !strings.Contains(str, "my-function") {
		t.Errorf("Expected 'my-function' in string, got: %s", str)
	}
}

// TestFunctionTask_ContextValues verifies context values are accessible.
func TestFunctionTask_ContextValues(t *testing.T) {
	type contextKey string
	const testKey contextKey = "test-key"

	ctx := context.WithValue(context.Background(), testKey, "test-value")

	task := &FunctionTask{
		Name: "context-test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			value := ctx.Value(testKey)
			if value == nil {
				return nil, errors.New("context value not found")
			}
			return []byte(value.(string)), nil
		},
	}

	result, err := Run(ctx, task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if string(result.Output) != "test-value" {
		t.Errorf("Expected 'test-value', got: %s", result.Output)
	}
}

// TestFunctionTask_MultipleIterations verifies function with loop.
func TestFunctionTask_MultipleIterations(t *testing.T) {
	task := &FunctionTask{
		Name: "iterator",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			total := 5
			for i := 0; i < total; i++ {
				// Check cancellation
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}

				r.ReportProgress(int64(i+1), int64(total))
				time.Sleep(10 * time.Millisecond)
			}
			return []byte("completed"), nil
		},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if string(result.Output) != "completed" {
		t.Errorf("Expected 'completed', got: %s", result.Output)
	}
}

// TestFunctionTask_ErrorReporting verifies non-fatal errors can be reported.
func TestFunctionTask_ErrorReporting(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	task := &FunctionTask{
		Name: "error-reporter",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			// Report non-fatal errors
			r.ReportError(errors.New("warning 1"))
			r.ReportError(errors.New("warning 2"))
			// But succeed overall
			return []byte("success despite warnings"), nil
		},
	}

	// Collect errors
	errors := []error{}
	done := make(chan struct{})
	go func() {
		timeout := time.After(1 * time.Second)
		for {
			select {
			case err, ok := <-reporter.Errors():
				if !ok {
					return
				}
				errors = append(errors, err)
			case <-timeout:
				return
			case <-done:
				return
			}
		}
	}()

	// Run task
	result, err := Run(context.Background(), task, reporter)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Wait for error collection
	time.Sleep(100 * time.Millisecond)
	close(done)

	// Verify non-fatal errors were reported
	if len(errors) < 2 {
		t.Errorf("Expected at least 2 error reports, got %d", len(errors))
	}

	// Verify task succeeded despite errors
	if result.ExitCode != 0 {
		t.Error("Expected exit code 0")
	}
}
