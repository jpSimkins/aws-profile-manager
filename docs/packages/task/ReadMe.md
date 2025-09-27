# Task Package

The task package provides a unified execution model for external subprocesses and Go functions. It solves a specific problem: business logic that needs to run commands (like `aws` or `git`) or perform long work must work identically in CLI and GUI contexts — with the same progress reporting, cancellation support, and error handling.

## Table of Contents
- [Why This Package Exists](#why-this-package-exists)
- [Core Concepts](#core-concepts)
- [Task Types](#task-types)
- [Reporter Types](#reporter-types)
- [Usage Patterns](#usage-patterns)
- [Adding a New Task](#adding-a-new-task)

---

## Why This Package Exists

Without this package, business logic that runs subprocesses would have to choose between printing to stdout (which breaks GUI) or taking a channel parameter (which complicates CLI). The task package solves this with a `Reporter` interface — CLI passes `CliReporter{}` which prints to console, GUI passes `ChannelReporter` which sends updates to the UI, and tests pass `NoOpReporter{}` which discards everything.

The result is that the same business logic function works unmodified in all three contexts.

---

## Core Concepts

### Task Interface

Any unit of work that can be executed:

```go
type Task interface {
    Execute(ctx context.Context, reporter Reporter) (*Result, error)
    String() string
}
```

### Result

Every task returns a consistent `Result`:

```go
type Result struct {
    Output   []byte                 // stdout+stderr (subprocess) or return value (function)
    ExitCode int                    // 0 = success
    Duration time.Duration          // total execution time
    Canceled bool                   // true if context was canceled
    Metadata map[string]interface{} // task-specific extras
}
```

### Reporter Interface

Progress reporting decoupled from the execution environment:

```go
type Reporter interface {
    ReportStatus(status string)             // "Fetching from S3..."
    ReportProgress(current, total int64)    // 5/10 items processed
    ReportError(err error)                  // non-fatal warning
}
```

`ReportError` does **not** stop execution. It signals a retry or warning. Fatal errors are returned from `Execute()`.

---

## Task Types

### SubprocessTask

Runs an external command. Uses `exec.CommandContext` for automatic process killing on context cancellation.

```go
t := &task.SubprocessTask{
    Name:    "aws",
    Args:    []string{"s3", "cp", "s3://my-bucket/config.json", "/tmp/config.json"},
    Env:     map[string]string{"AWS_PROFILE": "prod"},
    Timeout: 30 * time.Second,
    Dir:     "/tmp",  // working directory (optional)
}

result, err := task.Run(ctx, t, reporter)
if err != nil {
    return fmt.Errorf("aws s3 cp failed: %w", err)
}
```

**Security**: The `Env` map is the only environment the subprocess receives — it does not inherit `os.Environ()`. Callers must explicitly pass any env vars the command needs. This prevents accidental credential leakage and makes the environment predictable.

**Process killing**: On cancellation, the process receives SIGKILL (Unix) or `TerminateProcess` (Windows). There is no graceful shutdown — commands must be idempotent or use transactions if partial writes are a concern.

### FunctionTask

Runs a Go function with the same interface as a subprocess. Useful for long-running operations that need progress reporting but don't involve external processes.

```go
t := &task.FunctionTask{
    Name: "install-profiles",
    Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
        r.ReportStatus("Writing managed section...")

        for i, profile := range profiles {
            // Check cancellation inside the loop
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            default:
            }

            if err := writeProfile(profile); err != nil {
                return nil, fmt.Errorf("failed to write %s: %w", profile.Name, err)
            }

            r.ReportProgress(int64(i+1), int64(len(profiles)))
        }

        r.ReportStatus("Done")
        return []byte("ok"), nil
    },
}

result, err := task.Run(ctx, t, reporter)
```

**Cancellation requirement**: Functions **must** check `ctx.Done()` inside any loop or blocking operation. A function that ignores context cancellation will block until it finishes, even if the user has already cancelled the operation.

---

## Reporter Types

### NoOpReporter

Silently discards all updates. Use in tests and background operations where progress is irrelevant.

```go
result, err := task.Run(ctx, t, task.NoOpReporter{})
```

### CliReporter

Prints status updates to console using the logging package. Use in CLI command handlers.

```go
result, err := task.Run(ctx, t, task.CliReporter{})
// Status messages appear as info log lines
// Progress appears as "Progress: 50.0% (5/10)"
// Non-fatal errors appear as warnings
```

### ChannelReporter

Sends updates to buffered channels. Use in GUI handlers.

```go
reporter := task.NewChannelReporter()

// Read updates in a goroutine
go func() {
    for status := range reporter.Status() {
        fyne.Do(func() { label.SetText(status) })
    }
}()

// Run task in another goroutine
go func() {
    defer reporter.Close()
    result, err := task.Run(ctx, t, reporter)
    // handle result
}()
```

Channels are buffered (size 10). If the consumer is too slow, updates are dropped — the reporter never blocks task execution.

**Always call `Close()`** when the task is done to avoid goroutine leaks on the reader side.

---

## Usage Patterns

### In Business Logic (Package Functions)

Accept `task.Reporter` as a parameter. Never import CLI or GUI packages.

```go
// internal/sync/sync.go
func Sync(ctx context.Context, cfg SyncConfig, opts Options, reporter task.Reporter) (*Result, error) {
    reporter.ReportStatus("Validating configuration...")

    fetcher, err := createFetcher(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create fetcher: %w", err)
    }

    t := &task.SubprocessTask{
        Name: "aws",
        Args: []string{"s3", "cp", "s3://" + cfg.S3Bucket + "/" + cfg.S3Key, "-"},
        Env:  cfg.S3AWSEnv,
    }

    result, err := task.Run(ctx, t, reporter)
    if err != nil {
        return nil, fmt.Errorf("s3 fetch failed: %w", err)
    }

    reporter.ReportStatus("Configuration fetched successfully")
    return &Result{Data: result.Output}, nil
}
```

### In CLI Commands

Pass `task.CliReporter{}`. It prints automatically.

```go
// internal/cli/sync.go
func runSyncFetch(cmd *cobra.Command, args []string) error {
    cfg := sync.ConfigFromSettings(&settings.Get().Sync)
    result, err := sync.Sync(ctx, cfg, sync.Options{}, task.CliReporter{})
    if err != nil {
        return logging.Log.ErrorfWithDetails("sync failed", err)
    }
    // display result
}
```

### In GUI Handlers

Pass `task.NoOpReporter{}` if you're managing your own progress dialog, or `task.ChannelReporter` if you want to pipe status into a label.

```go
// internal/gui/views/sync_dialog.go
func (d *SyncDialog) runSync() {
    go func() {
        reporter := task.NoOpReporter{}
        result, err := sync.Sync(ctx, cfg, sync.Options{}, reporter)

        fyne.Do(func() {
            d.progressDialog.Hide()
            if err != nil {
                dialog.ShowError(err, d.window)
                return
            }
            d.showSuccess(result)
        })
    }()
}
```

---

## Adding a New Task

Implement the `Task` interface:

```go
type MyTask struct {
    // configuration fields
}

func (t *MyTask) Execute(ctx context.Context, reporter task.Reporter) (*task.Result, error) {
    start := time.Now()
    reporter.ReportStatus("Starting my task...")

    // Do work, checking ctx.Done() in any loop
    select {
    case <-ctx.Done():
        return &task.Result{Canceled: true, Duration: time.Since(start)}, ctx.Err()
    default:
    }

    // ... rest of work ...

    return &task.Result{
        Output:   []byte("result"),
        ExitCode: 0,
        Duration: time.Since(start),
    }, nil
}

func (t *MyTask) String() string {
    return "my-task"
}
```

Then run it with `task.Run(ctx, &MyTask{...}, reporter)`.
