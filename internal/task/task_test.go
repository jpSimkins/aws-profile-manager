package task

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRun_NilTask verifies nil task is rejected.
func TestRun_NilTask(t *testing.T) {
	result, err := Run(context.Background(), nil, NoOpReporter{})

	if err == nil {
		t.Fatal("Expected error for nil task")
	}
	if result != nil {
		t.Error("Expected nil result for nil task")
	}
}

// TestRun_NilReporter verifies nil reporter defaults to NoOpReporter.
func TestRun_NilReporter(t *testing.T) {
	task := &FunctionTask{
		Name: "test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			// Reporter should be non-nil (NoOpReporter)
			r.ReportStatus("test")
			return []byte("success"), nil
		},
	}

	result, err := Run(context.Background(), task, nil)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

// TestRun_SubprocessTask verifies Run() works with SubprocessTask.
func TestRun_SubprocessTask(t *testing.T) {
	task := &SubprocessTask{
		Name: "echo",
		Args: []string{"test"},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}

// TestRun_FunctionTask verifies Run() works with FunctionTask.
func TestRun_FunctionTask(t *testing.T) {
	task := &FunctionTask{
		Name: "test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			return []byte("success"), nil
		},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
}

// TestRun_WithChannelReporter verifies Run() works with ChannelReporter.
func TestRun_WithChannelReporter(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	task := &FunctionTask{
		Name: "test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			r.ReportStatus("test status")
			return []byte("success"), nil
		},
	}

	done := make(chan struct{})
	go func() {
		_, _ = Run(context.Background(), task, reporter)
		close(done)
	}()

	// Should receive status
	select {
	case status := <-reporter.Status():
		if status == "" {
			t.Error("Expected non-empty status")
		}
	case <-done:
		t.Fatal("Task completed before status received")
	}

	<-done
}

// TestRun_WithCliReporter verifies Run() works with CliReporter.
func TestRun_WithCliReporter(t *testing.T) {
	task := &FunctionTask{
		Name: "test",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			r.ReportStatus("test status")
			r.ReportProgress(1, 2)
			r.ReportError(errors.New("test warning"))
			return []byte("success"), nil
		},
	}

	// Should not panic
	result, err := Run(context.Background(), task, CliReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
}

// TestRun_ContextCancellation verifies cancellation works through Run().
func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	task := &FunctionTask{
		Name: "cancellable",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}

	// Cancel immediately
	cancel()

	result, err := Run(ctx, task, NoOpReporter{})

	if err == nil {
		t.Fatal("Expected cancellation error")
	}
	if result == nil {
		t.Fatal("Expected result even on error")
	}
	if !result.Canceled {
		t.Error("Expected result.Canceled to be true")
	}
}

// TestRunAsync_Success verifies RunAsync invokes callback with successful result.
func TestRunAsync_Success(t *testing.T) {
	done := make(chan struct{})
	taskDef := &FunctionTask{
		Name: "async-success",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			return []byte("ok"), nil
		},
	}

	RunAsync(context.Background(), taskDef, NoOpReporter{}, func(result *Result, err error) {
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		} else if result.ExitCode != 0 {
			t.Errorf("expected exit code 0, got %d", result.ExitCode)
		}
		close(done)
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RunAsync callback")
	}
}

// TestRunAsync_Error verifies RunAsync passes execution errors to callback.
func TestRunAsync_Error(t *testing.T) {
	done := make(chan struct{})
	taskDef := &FunctionTask{
		Name: "async-error",
		Fn: func(ctx context.Context, r Reporter) ([]byte, error) {
			return nil, errors.New("boom")
		},
	}

	RunAsync(context.Background(), taskDef, NoOpReporter{}, func(result *Result, err error) {
		if err == nil {
			t.Error("expected error")
		}
		if result == nil {
			t.Error("expected non-nil result even on error")
		}
		close(done)
	})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for RunAsync callback")
	}
}
