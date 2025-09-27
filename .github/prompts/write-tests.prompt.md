---
agent: agent
description: Write tests for AWS Profile Manager — covers patterns, fixtures, and isolation requirements
---

You are writing tests for the AWS Profile Manager. Follow these rules exactly.

## Rule 1 — Every test file that touches the filesystem MUST call SetupTestEnvironment

```go
func TestMyFunction(t *testing.T) {
    test.SetupTestEnvironment(t)  // Creates temp dirs, sets env vars, auto-cleanup
    // ...
}
```

What it does:
- Creates isolated temp directories: `config/`, `config/cache/`, `.aws/`, `Desktop/`
- Points all environment variables at these temp dirs
- Registers `t.Cleanup()` — no manual teardown needed

**Skip it only** for pure unit tests that have zero file I/O and zero env var dependency.

## Rule 2 — Always use schematest fixtures

**NEVER hardcode a schema inline**. Use the centralized fixtures:

```go
import schematest "aws-profile-manager/internal/schema/test"

// Pick the scenario that matches your test:
schema := schematest.NewManagedSsoSingle()          // 1 org, 1 account, 1 role
schema := schematest.NewManagedAll()                 // all managed profile types
schema := schematest.NewMixedSimple()                // managed + personal profiles
schema := schematest.NewPartialSsoMissingUrl()       // SSO org missing URL (error case)
schema := schematest.NewInvalid()                    // fully invalid (error case)
schema := schematest.NewEmpty()                      // empty schema (edge case)
schema := schematest.NewLargeScale()                 // 2100+ profiles (performance)
```

Full list: see `internal/schema/test/` (managed.go, unmanaged.go, mixed.go, specialized.go).

## Rule 3 — Use NoOpReporter in tests

```go
result, err := mypackage.DoWork(ctx, cfg, task.NoOpReporter{})
//                                         ↑ always — never task.CliReporter{}
```

## Rule 4 — Handle ALL errors explicitly

```go
// ✅ Check it
if err := os.MkdirAll(dir, 0755); err != nil {
    t.Fatalf("setup failed: %v", err)
}

// ✅ Discard it intentionally
_ = cmd.Flags().Set("verbose", "true")

// ❌ Unchecked (lint error)
os.WriteFile(path, data, 0644)
```

## Rule 5 — Use t.Fatal for nil pointer checks

```go
// ✅ Prevents SA5011 lint error
if result == nil {
    t.Fatal("expected non-nil result")
}
result.SomeMethod()  // linter knows it's safe

// ❌ Will trigger SA5011
if result == nil {
    t.Error("expected non-nil result")
}
result.SomeMethod()  // linter warns: possible nil dereference
```

## Standard test file template

```go
package mypackage

import (
    "context"
    "testing"

    "aws-profile-manager/internal/task"
    "aws-profile-manager/internal/test"
    schematest "aws-profile-manager/internal/schema/test"
)

func TestDoWork_Success(t *testing.T) {
    test.SetupTestEnvironment(t)

    cfg := MyConfig{
        SomeField: "test-value",
    }
    schema := schematest.NewManagedSsoSingle()

    result, err := DoWork(context.Background(), cfg, schema, task.NoOpReporter{})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result == nil {
        t.Fatal("expected non-nil result")
    }
    // assert specific values
    if result.Count != 2 {
        t.Errorf("got count %d, want 2", result.Count)
    }
}

func TestDoWork_MissingField(t *testing.T) {
    test.SetupTestEnvironment(t)

    cfg := MyConfig{SomeField: ""}  // invalid — triggers error

    _, err := DoWork(context.Background(), cfg, schematest.NewEmpty(), task.NoOpReporter{})
    if err == nil {
        t.Fatal("expected error for missing field, got nil")
    }
}
```

## Table-driven tests for multiple scenarios

```go
func TestDoWork_Scenarios(t *testing.T) {
    tests := []struct {
        name        string
        schema      *schema.Schema
        wantErr     bool
        wantCount   int
    }{
        {"single SSO org", schematest.NewManagedSsoSingle(), false, 2},
        {"all types",      schematest.NewManagedAll(),       false, 10},
        {"empty schema",   schematest.NewEmpty(),            false, 0},
        {"invalid schema", schematest.NewInvalid(),          true,  0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            test.SetupTestEnvironment(t)

            result, err := DoWork(context.Background(), validCfg, tt.schema, task.NoOpReporter{})
            if (err != nil) != tt.wantErr {
                t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
            }
            if !tt.wantErr {
                if result == nil {
                    t.Fatal("expected non-nil result")
                }
                if result.Count != tt.wantCount {
                    t.Errorf("got count %d, want %d", result.Count, tt.wantCount)
                }
            }
        })
    }
}
```

## Tests needing app state

```go
func TestWithAppState(t *testing.T) {
    test.SetupTestEnvironment(t)

    if err := core.App.Initialize(nil); err != nil {
        t.Fatalf("failed to initialize app: %v", err)
    }

    // Now settings, state, markers are available
    appSettings := settings.GetApplication()
    startMarker := appSettings.GetFormattedStartMarker()
    // ...
}
```

## GUI tests

```go
import fynetest "fyne.io/fyne/v2/test"

func TestMyView(t *testing.T) {
    test.SetupTestEnvironment(t)

    testApp := fynetest.NewApp()
    window := testApp.NewWindow("test")

    view := NewMyView(window)
    if view == nil {
        t.Fatal("expected non-nil view")
    }
}
```

## Coverage targets

- Every `.go` file must have a `_test.go` companion
- Every exported function should be tested
- Every error return path must have at least one test
- Target: 95%+ per package

## Common mistakes

- ❌ Inline schema: `schema := &schema.Schema{...}` — use `schematest.New*()`
- ❌ Unchecked errors: `os.WriteFile(...)` — use `_ = os.WriteFile(...)` or check
- ❌ Missing `test.SetupTestEnvironment(t)` for file I/O tests
- ❌ Using `task.CliReporter{}` in tests — use `task.NoOpReporter{}`
- ❌ Hardcoded markers: `"# START - Managed by..."` — use settings helpers
- ❌ Invalid settings: `settings.Sync.Strategy = "invalid"` will fail on `settings.Set()`
