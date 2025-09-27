# Task Package

Unified execution for subprocesses and Go functions with context cancellation and progress reporting.

## Documentation

Full documentation is available at [`docs/packages/task/ReadMe.md`](../../docs/packages/task/ReadMe.md).

## Quick Start

### Subprocess Task

```go
import (
    "context"
    "aws-profile-manager/internal/task"
)

task := &task.SubprocessTask{
    Name: "aws",
    Args: []string{"s3", "ls", "s3://my-bucket"},
    Env:  map[string]string{"AWS_PROFILE": "prod"},
}

result, err := task.Run(context.Background(), task, task.CliReporter{})
```

### Function Task

```go
task := &task.FunctionTask{
    Name: "process-data",
    Fn: func(ctx context.Context, r task.Reporter) ([]byte, error) {
        r.ReportStatus("Processing...")
        // Do work
        return []byte("done"), nil
    },
}

result, err := task.Run(context.Background(), task, task.NoOpReporter{})
```

## Why Use This Package?

The task package enables **CLI/GUI agnostic business logic**:

- ✅ Same code works in CLI and GUI
- ✅ Progress reporting built-in
- ✅ Context cancellation support
- ✅ No subprocess boilerplate

See [full documentation](../../docs/packages/task/README.md) for complete guide.
