package task

import (
	"context"
	"time"
)

// Task represents a unit of work that can be executed with cancellation and progress.
//
// Implementations:
//   - SubprocessTask: Execute external commands (aws, git, etc.)
//   - FunctionTask: Execute Go functions
//
// All tasks MUST:
//   - Respect ctx.Done() and cancel gracefully
//   - Report progress via reporter for user feedback
//   - Return consistent Result structure
type Task interface {
	// Execute runs the task with context cancellation and progress reporting.
	//
	// MUST check ctx.Done() periodically and cancel if triggered.
	// SHOULD report meaningful status updates via reporter.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadlines
	//   - reporter: Progress reporter for status updates
	//
	// Returns:
	//   - *Result: Execution result with output and metadata
	//   - error: Any error encountered during execution
	Execute(ctx context.Context, reporter Reporter) (*Result, error)

	// String returns human-readable task description for logging.
	String() string
}

// Result contains the output and metadata from task execution.
//
// Returned by all tasks regardless of type (subprocess or function).
// Even on error, Result may contain partial output for debugging.
type Result struct {
	// Output contains the task output:
	//   - Subprocess: Combined stdout/stderr
	//   - Function: Returned byte slice
	Output []byte

	// ExitCode indicates success/failure:
	//   - 0: Success
	//   - Non-zero: Failure
	//   - Subprocess: Process exit code
	//   - Function: 0 on success, 1 on error
	ExitCode int

	// Duration is the total execution time
	Duration time.Duration

	// Canceled is true if task was canceled via context
	Canceled bool

	// Metadata contains task-specific additional information (optional)
	// Examples: bytes transferred, items processed, etc.
	Metadata map[string]interface{}
}

// Progress represents a progress update with current and total values.
type Progress struct {
	Current int64 // Current progress value
	Total   int64 // Total expected value (0 if unknown)
}
