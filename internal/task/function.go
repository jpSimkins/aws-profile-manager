package task

import (
	"context"
	"fmt"
	"time"
)

// FunctionTask executes a Go function with context cancellation.
//
// The function signature must be:
//
//	func(context.Context, Reporter) ([]byte, error)
//
// Function MUST:
//   - Check ctx.Done() periodically
//   - Return promptly on cancellation
//   - Report progress for user feedback
//
// Features:
//   - Same interface as SubprocessTask
//   - Context cancellation
//   - Progress reporting
//   - Consistent Result structure
type FunctionTask struct {
	// Name is the task name for logging/display
	Name string

	// Fn is the function to execute
	// Must respect ctx.Done() for cancellation
	// Should report progress via reporter
	Fn func(ctx context.Context, reporter Reporter) ([]byte, error)
}

// Execute implements the Task interface for Go functions.
//
// The function is called with the provided context and reporter.
// It is the function's responsibility to:
//   - Check ctx.Done() periodically for cancellation
//   - Report meaningful progress updates
//   - Return promptly when canceled
//
// Returns:
//   - *Result: Contains output, exit code (0 or 1), duration, and cancellation status
//   - error: Any error returned by the function
func (t *FunctionTask) Execute(ctx context.Context, reporter Reporter) (*Result, error) {
	reporter.ReportStatus(fmt.Sprintf("Starting task: %s", t.Name))
	start := time.Now()

	// Execute function (function must respect ctx.Done())
	output, err := t.Fn(ctx, reporter)

	// Build result
	result := &Result{
		Output:   output,
		Duration: time.Since(start),
		Canceled: ctx.Err() != nil,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
	}

	if err != nil {
		result.ExitCode = 1
		reporter.ReportError(err)
		return result, err
	}

	reporter.ReportStatus(fmt.Sprintf("Task completed: %s", t.Name))
	return result, nil
}

// String returns human-readable task description.
func (t *FunctionTask) String() string {
	return fmt.Sprintf("function: %s", t.Name)
}
