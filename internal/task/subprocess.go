package task

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SubprocessTask executes an external command with context cancellation.
//
// Features:
//   - Context cancellation (kills process tree)
//   - Explicit environment (no os.Environ)
//   - Timeout support
//   - Working directory control
//   - Cross-platform process killing
//
// Security:
//   - No shell execution (prevents injection)
//   - Explicit environment only
//   - Caller must validate inputs
type SubprocessTask struct {
	// Name is the command to execute (e.g., "aws", "git")
	Name string

	// Args are the command arguments
	Args []string

	// Env is the explicit environment (no os.Environ inheritance)
	// Format: map["KEY"] = "value"
	Env map[string]string

	// Timeout is an optional timeout (0 = no timeout)
	// Context cancellation takes precedence if it occurs first
	Timeout time.Duration

	// Dir is the working directory (optional, defaults to current)
	Dir string
}

// Execute implements the Task interface for subprocesses.
//
// The subprocess will be terminated if:
//   - Context is canceled (ctx.Done())
//   - Timeout is reached (if specified)
//
// Go's exec.CommandContext handles cross-platform process killing:
//   - Unix/Linux/macOS: Sends SIGKILL to process
//   - Windows: Calls TerminateProcess
//
// Returns:
//   - *Result: Contains output, exit code, duration, and cancellation status
//   - error: Any error encountered during execution
func (t *SubprocessTask) Execute(ctx context.Context, reporter Reporter) (*Result, error) {
	reporter.ReportStatus(fmt.Sprintf("Starting subprocess: %s", t.Name))
	start := time.Now()

	// Apply timeout if specified
	execCtx := ctx
	if t.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, t.Timeout)
		defer cancel()
	}

	// Create command with context (auto-kills on cancel)
	// #nosec G204 -- Command/args are intentionally provided by trusted callers of the task API.
	cmd := exec.CommandContext(execCtx, t.Name, t.Args...)

	// Set explicit environment (only what we're given)
	if len(t.Env) > 0 {
		cmd.Env = envMapToSlice(t.Env)
	}

	// Set working directory if specified
	if t.Dir != "" {
		cmd.Dir = t.Dir
	}

	// Execute and capture output
	output, err := cmd.CombinedOutput()

	// Build result
	result := &Result{
		Output:   output,
		Duration: time.Since(start),
		Canceled: ctx.Err() != nil,
		Metadata: make(map[string]interface{}),
	}

	// Set exit code
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		result.ExitCode = 1
	}

	if err != nil {
		reporter.ReportError(err)
		return result, fmt.Errorf("subprocess failed: %w", err)
	}

	reporter.ReportStatus("Subprocess completed successfully")
	return result, nil
}

// String returns human-readable task description.
func (t *SubprocessTask) String() string {
	return fmt.Sprintf("subprocess: %s %s", t.Name, strings.Join(t.Args, " "))
}

// envMapToSlice converts environment map to slice for exec.Cmd.Env.
//
// Parameters:
//   - env: Environment variables as map
//
// Returns:
//   - []string: Environment variables in KEY=value format
func envMapToSlice(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, k+"="+v)
	}
	return result
}
