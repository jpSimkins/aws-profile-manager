// Package task provides unified execution for subprocesses and Go functions
// with context cancellation, progress reporting, and consistent error handling.
//
// The task package abstracts the execution model, allowing both external
// processes (via SubprocessTask) and Go functions (via FunctionTask) to be
// executed with the same interface, progress reporting, and cancellation support.
//
// Key features:
//   - Context-based cancellation for both subprocess and function tasks
//   - Real-time progress reporting via channels (GUI) or console (CLI)
//   - Consistent Result structure across all task types
//   - Cross-platform subprocess execution (Linux, macOS, Windows)
//   - No global state - all configuration passed explicitly
//
// Example usage (subprocess):
//
//	task := &task.SubprocessTask{
//	    Name: "aws",
//	    Args: []string{"s3", "ls"},
//	    Env: map[string]string{"AWS_PROFILE": "prod"},
//	}
//	result, err := task.Run(ctx, task, task.CliReporter{})
//
// Example usage (Go function):
//
//	task := &task.FunctionTask{
//	    Name: "process-data",
//	    Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
//	        r.ReportStatus("Processing...")
//	        // ... work ...
//	        return []byte("done"), nil
//	    },
//	}
//	result, err := task.Run(ctx, task, task.CliReporter{})
package task

import (
	"context"
	"fmt"
)

// Run executes a task with context cancellation and progress reporting.
//
// This is the primary entry point for all task execution. Works for both
// subprocess commands and Go functions transparently.
//
// Process:
//  1. Validates task and reporter (non-nil)
//  2. Calls task.Execute() with context and reporter
//  3. Returns result and any error
//
// Cancellation:
//
//	Pass a cancellable context to stop execution mid-flight.
//	Task MUST respect ctx.Done() and clean up appropriately.
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - task: Task to execute (SubprocessTask or FunctionTask)
//   - reporter: Progress reporter (use NoOpReporter{} for silent execution)
//
// Returns:
//   - *Result: Execution result (even on error, for debugging)
//   - error: Any error encountered during execution
//
// Example (subprocess):
//
//	task := &task.SubprocessTask{
//	    Name: "aws",
//	    Args: []string{"s3", "ls"},
//	    Env: map[string]string{"AWS_PROFILE": "prod"},
//	}
//	result, err := task.Run(ctx, task, reporter)
//
// Example (Go function):
//
//	task := &task.FunctionTask{
//	    Name: "install-profiles",
//	    Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
//	        r.ReportStatus("Installing profiles...")
//	        // ... installation logic
//	        return []byte("success"), nil
//	    },
//	}
//	result, err := task.Run(ctx, task, reporter)
func Run(ctx context.Context, t Task, reporter Reporter) (*Result, error) {
	if t == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if reporter == nil {
		reporter = NoOpReporter{} // Default to silent
	}

	return t.Execute(ctx, reporter)
}

// RunAsync executes a task in the background and invokes onComplete with result.
//
// This helper centralizes async task execution in the task package so callers
// do not need to create raw goroutines for task orchestration.
func RunAsync(
	ctx context.Context,
	t Task,
	reporter Reporter,
	onComplete func(result *Result, err error),
) {
	go func() {
		result, err := Run(ctx, t, reporter)
		if onComplete != nil {
			onComplete(result, err)
		}
	}()
}
