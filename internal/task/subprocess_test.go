package task

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestSubprocessTask_Success verifies successful command execution.
func TestSubprocessTask_Success(t *testing.T) {
	task := &SubprocessTask{
		Name: "echo",
		Args: []string{"hello", "world"},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}
	if !strings.Contains(string(result.Output), "hello world") {
		t.Errorf("Expected 'hello world' in output, got: %s", result.Output)
	}
	if result.Canceled {
		t.Error("Expected result.Canceled to be false")
	}
	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// TestSubprocessTask_Failure verifies command failure handling.
func TestSubprocessTask_Failure(t *testing.T) {
	task := &SubprocessTask{
		Name: "false", // Command that always fails
		Args: []string{},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err == nil {
		t.Fatal("Expected error for failed command")
	}
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}
}

// TestSubprocessTask_ContextCancellation verifies context cancellation kills process.
func TestSubprocessTask_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	task := &SubprocessTask{
		Name: "sleep",
		Args: []string{"10"},
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
		t.Fatal("Expected error due to cancellation")
	}
	if !result.Canceled {
		t.Error("Expected result.Canceled to be true")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Process took too long to cancel: %s", elapsed)
	}
}

// TestSubprocessTask_Timeout verifies timeout handling.
func TestSubprocessTask_Timeout(t *testing.T) {
	task := &SubprocessTask{
		Name:    "sleep",
		Args:    []string{"10"},
		Timeout: 100 * time.Millisecond,
	}

	start := time.Now()
	result, err := Run(context.Background(), task, NoOpReporter{})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected timeout error")
	}
	if elapsed > 2*time.Second {
		t.Errorf("Timeout took too long: %s", elapsed)
	}
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code for timeout")
	}
}

// TestSubprocessTask_Environment verifies environment variables are passed.
func TestSubprocessTask_Environment(t *testing.T) {
	task := &SubprocessTask{
		Name: "sh",
		Args: []string{"-c", "echo $TEST_VAR"},
		Env: map[string]string{
			"TEST_VAR": "hello-from-test",
		},
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !strings.Contains(string(result.Output), "hello-from-test") {
		t.Errorf("Expected 'hello-from-test' in output, got: %s", result.Output)
	}
}

// TestSubprocessTask_WorkingDirectory verifies working directory is set.
func TestSubprocessTask_WorkingDirectory(t *testing.T) {
	task := &SubprocessTask{
		Name: "pwd",
		Args: []string{},
		Dir:  "/tmp",
	}

	result, err := Run(context.Background(), task, NoOpReporter{})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if !strings.Contains(string(result.Output), "/tmp") {
		t.Errorf("Expected '/tmp' in output, got: %s", result.Output)
	}
}

// TestSubprocessTask_String verifies String() method.
func TestSubprocessTask_String(t *testing.T) {
	task := &SubprocessTask{
		Name: "aws",
		Args: []string{"s3", "ls"},
	}

	str := task.String()
	if !strings.Contains(str, "subprocess") {
		t.Errorf("Expected 'subprocess' in string, got: %s", str)
	}
	if !strings.Contains(str, "aws") {
		t.Errorf("Expected 'aws' in string, got: %s", str)
	}
	if !strings.Contains(str, "s3 ls") {
		t.Errorf("Expected 's3 ls' in string, got: %s", str)
	}
}

// TestSubprocessTask_ProgressReporting verifies progress is reported.
func TestSubprocessTask_ProgressReporting(t *testing.T) {
	reporter := NewChannelReporter()
	defer reporter.Close()

	task := &SubprocessTask{
		Name: "echo",
		Args: []string{"test"},
	}

	// Run task in background
	done := make(chan struct{})
	go func() {
		_, _ = Run(context.Background(), task, reporter)
		close(done)
	}()

	// Wait for at least one status update
	select {
	case status := <-reporter.Status():
		if status == "" {
			t.Error("Expected non-empty status")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for status update")
	}

	// Wait for completion
	<-done
}

// TestEnvMapToSlice verifies environment map conversion.
func TestEnvMapToSlice(t *testing.T) {
	env := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	}

	slice := envMapToSlice(env)

	if len(slice) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(slice))
	}

	// Check both entries exist (order may vary due to map iteration)
	found1 := false
	found2 := false
	for _, entry := range slice {
		if entry == "KEY1=value1" {
			found1 = true
		}
		if entry == "KEY2=value2" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Expected KEY1=value1 in result")
	}
	if !found2 {
		t.Error("Expected KEY2=value2 in result")
	}
}
