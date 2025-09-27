---
agent: agent
description: Add a new CLI command to aws-profile-manager
---

You are adding a new CLI command to the AWS Profile Manager. CLI commands live in `internal/cli/`. They are **thin wrappers** — they parse flags, build config from settings, call one package function, and display results.

## Structure

Every CLI command file contains:
1. A `create<Name>Command()` function that returns `*cobra.Command`
2. A `run<Name>Command(cmd, args)` function that implements the handler
3. A `display<Name>Result(result)` function for output formatting

## Template

```go
package cli

import (
    "github.com/spf13/cobra"

    "aws-profile-manager/internal/logging"
    "aws-profile-manager/internal/mypackage"
    "aws-profile-manager/internal/settings"
    "aws-profile-manager/internal/task"
)

// create<Name>Command creates the <name> command.
//
// <Describe what the command does.>
func create<Name>Command() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "<name>",
        Short: "Short one-line description",
        Long:  `Longer description shown in --help output.`,
        RunE:  run<Name>Command,
    }

    cmd.Flags().String("my-flag", "", "Description of the flag")

    return cmd
}

// run<Name>Command implements the <name> command handler.
func run<Name>Command(cmd *cobra.Command, args []string) error {
    // 1. Parse flags
    myFlag, _ := cmd.Flags().GetString("my-flag")

    // 2. Build config from settings (ONLY CLI/GUI touches settings)
    currentSettings := settings.Get()
    cfg := mypackage.<Name>Config{
        SomeValue: currentSettings.Section.Field,
        OtherValue: myFlag,
    }

    // 3. Call ONE package function with CliReporter
    result, err := mypackage.Do<Name>(cmd.Context(), cfg, task.CliReporter{})
    if err != nil {
        return logging.Log.ErrorfWithDetails("<name> failed", err)
    }

    // 4. Display results
    display<Name>Result(result)
    return nil
}

// display<Name>Result prints the result to the console.
func display<Name>Result(result *mypackage.<Name>Result) {
    logging.Log.Success("Operation complete",
        "count", result.Count,
    )
    // ... display logic
}
```

## Register the command

Add the new command to `createCommands()` in `internal/cli/commands.go`:

```go
rootCmd.AddCommand(create<Name>Command())
```

## Rules

- **One package call**: `run<Name>Command` calls exactly ONE business logic function
- **Always `task.CliReporter{}`**: Pass to the package function — it prints status automatically
- **Always `logging.Log.ErrorfWithDetails()`**: For error presentation at the CLI layer
- **Never create objects in CLI**: No `NewExtractor()`, no `NewFilter()` — the package does that
- **Never `fmt.Println()`**: Use `logging.Log.Info/Success/Warn/Error()`
- **Build config from settings here**: Business logic accepts config structs, not settings

## Test file

Create `internal/cli/<name>_test.go`:

```go
package cli

import (
    "testing"

    "aws-profile-manager/internal/core"
    "aws-profile-manager/internal/test"
)

func Test<Name>Command(t *testing.T) {
    test.SetupTestEnvironment(t)

    if err := core.App.Initialize(nil); err != nil {
        t.Fatalf("failed to initialize: %v", err)
    }

    cmd := create<Name>Command()
    cmd.SetArgs([]string{"--my-flag", "test-value"})

    // For commands that need file setup, write fixtures first
    // then execute and verify output or side effects
    err := cmd.Execute()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```
