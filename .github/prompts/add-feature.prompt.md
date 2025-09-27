---
agent: agent
description: Add a new feature to AWS Profile Manager — covers business logic, CLI command, and tests
---

You are adding a new feature to the AWS Profile Manager codebase. Follow this process exactly. Do not skip steps.

## Step 1 — Understand the task

Before writing any code:
1. Read `docs/Architecture.md` to understand the layer model.
2. Identify which existing packages are involved. Search the codebase to confirm what already exists.
3. Do NOT create new packages unless the feature genuinely has no home. Prefer extending existing ones.

## Step 2 — Design the config struct (if the feature uses external data or settings)

Business logic NEVER imports `internal/settings`. Create a config struct in your package:

```go
// In your package: types.go or the main file
type MyFeatureConfig struct {
    SomeValue  string
    OtherValue int
    Timeout    time.Duration
}
```

The CLI/GUI layer will build this from settings and inject it.

## Step 3 — Implement the business logic

The function signature for any operation that is long-running or calls external commands:

```go
func DoMyFeature(ctx context.Context, cfg MyFeatureConfig, reporter task.Reporter) (*MyResult, error) {
    reporter.ReportStatus("Starting...")

    // External command → task.SubprocessTask (NEVER exec.Command)
    t := &task.SubprocessTask{
        Name:    "aws",
        Args:    []string{"..."},
        Env:     map[string]string{"AWS_PROFILE": cfg.Profile},
        Timeout: cfg.Timeout,
    }
    result, err := task.Run(ctx, t, reporter)
    if err != nil {
        return nil, fmt.Errorf("command failed: %w", err)
    }

    // Long-running Go work → task.FunctionTask (NEVER go func(){})
    ft := &task.FunctionTask{
        Name: "process",
        Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
            for i, item := range items {
                select {
                case <-ctx.Done():
                    return nil, ctx.Err()  // MUST check cancellation
                default:
                }
                r.ReportProgress(int64(i+1), int64(len(items)))
                // process item...
            }
            return nil, nil
        },
    }
    _, err = task.Run(ctx, ft, reporter)
    if err != nil {
        return nil, fmt.Errorf("processing failed: %w", err)
    }

    return &MyResult{...}, nil
}
```

Rules:
- `fmt.Errorf("context: %w", err)` for all errors — never `logging.Log.Error*` inside business logic
- Never `panic()`
- Never `fmt.Println()` or `log.Print*()`
- Never import `internal/settings`, `internal/cli`, or `internal/gui`

## Step 4 — Add the CLI command (if needed)

CLI commands live in `internal/cli/`. They are thin wrappers:

```go
func runMyFeatureCommand(cmd *cobra.Command, args []string) error {
    // 1. Parse flags
    someFlag, _ := cmd.Flags().GetString("some-flag")

    // 2. Build config from settings (only CLI/GUI touches settings)
    currentSettings := settings.Get()
    cfg := mypackage.MyFeatureConfig{
        SomeValue: currentSettings.MySection.SomeValue,
        OtherValue: someFlag,
    }

    // 3. Call ONE package function
    result, err := mypackage.DoMyFeature(cmd.Context(), cfg, task.CliReporter{})
    if err != nil {
        return logging.Log.ErrorfWithDetails("feature failed", err)
    }

    // 4. Display results
    displayMyResult(result)
    return nil
}
```

Register the command in `createCommands()` in `internal/cli/commands.go`.

## Step 5 — Write tests

For every `.go` file you create, create a matching `_test.go`.

```go
func TestDoMyFeature(t *testing.T) {
    test.SetupTestEnvironment(t)  // ALWAYS — isolates file I/O

    cfg := mypackage.MyFeatureConfig{
        SomeValue: "test-value",
        Timeout:   5 * time.Second,
    }

    result, err := mypackage.DoMyFeature(context.Background(), cfg, task.NoOpReporter{})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // assertions...
}
```

Test rules:
- Use `schematest.New*()` for any schema fixtures — never hardcode schemas inline
- Use `test.NoOpReporter{}` in tests — never `task.CliReporter{}`
- Handle ALL errors explicitly: either check with `if err != nil` or discard with `_ = err`
- Test error paths (at least one test per error return)

## Step 6 — Verify

Run in order before declaring done:
```bash
make fmt
make vet
make lint        # must be zero errors
make test-coverage
make build
```

## Common mistakes to avoid

- ❌ `go func() { doWork() }()` — use `task.FunctionTask` instead
- ❌ `exec.Command(...)` or `exec.CommandContext(...)` — use `task.SubprocessTask` instead
- ❌ `fmt.Println(...)` — use `logging.Log.Info(...)` instead
- ❌ Importing `internal/settings` in business logic — accept a config struct instead
- ❌ Hardcoding marker strings — use `settings.GetApplication().GetFormattedStartMarker()`
- ❌ Missing godoc on exported symbols — every exported type, func, and field needs a comment
