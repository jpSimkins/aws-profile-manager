---
agent: agent
description: Add a new GUI view or dialog to the AWS Profile Manager Fyne application
---

You are adding a new GUI view or dialog. The GUI uses Fyne 2.x with an MVVM pattern. Read `docs/Architecture.md` before starting.

## Structure

```
internal/gui/
├── viewmodels/           ← Business state, data loading, operation execution
│   └── <name>_viewmodel.go
└── views/                ← Fyne widgets, zero business logic
    └── <name>_view.go    (or <name>_dialog.go for dialogs)
```

## Step 1 — ViewModel

The ViewModel owns all state and calls business logic packages. It never creates Fyne widgets.

```go
package viewmodels

import (
    "context"
    "sync"

    "aws-profile-manager/internal/core"
    "aws-profile-manager/internal/mypackage"
    "aws-profile-manager/internal/settings"
    "aws-profile-manager/internal/task"
)

// <Name>ViewModel manages state for the <Name> view.
type <Name>ViewModel struct {
    IsLoading bool
    mu        sync.RWMutex
}

// New<Name>ViewModel creates and registers a <Name> view model.
func New<Name>ViewModel() *<Name>ViewModel {
    vm := &<Name>ViewModel{}
    core.App.RegisterState("<name>-view", vm)
    return vm
}

// Do<Action> performs <description> and returns the result.
//
// This function is called from the view in a goroutine.
// It MUST NOT touch any Fyne widgets — the view handles UI updates via fyne.Do().
func (vm *<Name>ViewModel) Do<Action>(ctx context.Context) (*mypackage.<Action>Result, error) {
    vm.mu.Lock()
    vm.IsLoading = true
    vm.mu.Unlock()
    defer func() {
        vm.mu.Lock()
        vm.IsLoading = false
        vm.mu.Unlock()
    }()

    // Build config from settings (ViewModel CAN access settings)
    currentSettings := settings.Get()
    cfg := mypackage.<Action>Config{
        SomeValue: currentSettings.Section.Field,
    }

    // Call business logic — pass NoOpReporter when managing your own progress UI,
    // or task.ChannelReporter to pipe status into a label
    return mypackage.Do<Action>(ctx, cfg, task.NoOpReporter{})
}
```

## Step 2 — View

The View is pure Fyne widgets. It wires user actions to ViewModel calls and updates the UI.

**Thread safety is non-negotiable**: Any update to a Fyne widget from a goroutine MUST be inside `fyne.Do()`.

```go
package views

import (
    "context"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"

    "aws-profile-manager/internal/gui/components"
    "aws-profile-manager/internal/gui/viewmodels"
)

// New<Name>View creates the <Name> tab/view content.
func New<Name>View(window fyne.Window) fyne.CanvasObject {
    vm := viewmodels.New<Name>ViewModel()

    statusLabel := widget.NewLabel("Ready")

    actionButton := widget.NewButton("Do Action", func() {
        // Disable button immediately to prevent double-clicks
        actionButton.Disable()
        statusLabel.SetText("Working...")

        // ALL long work runs in a goroutine — never block the main thread
        go func() {
            result, err := vm.Do<Action>(context.Background())

            // ALL Fyne widget updates MUST be inside fyne.Do()
            fyne.Do(func() {
                actionButton.Enable()

                if err != nil {
                    statusLabel.SetText("Failed")
                    dialog.ShowError(err, window)
                    return
                }

                statusLabel.SetText("Complete")
                // update other widgets with result...
                _ = result
            })
        }()
    })

    return container.NewVBox(statusLabel, actionButton)
}
```

## Step 3 — For dialogs, use components.ShowCustomDialog

**NEVER** use `dialog.NewCustom()` directly — always use the project's wrapper:

```go
func ShowMyDialog(window fyne.Window) {
    content := buildDialogContent()

    buttons := []*widget.Button{
        widget.NewButton("Cancel", func() { /* close */ }),
        widget.NewButton("OK", func() { /* action */ }),
    }

    components.ShowCustomDialog(components.DialogOptions{
        Title:       "My Dialog",
        Content:     content,
        Buttons:     buttons,
        Window:      window,
        Scrollable:  true,
        UseSettings: true,
    })
}
```

## Step 4 — Text styling

Use Markdown for ALL non-plain text. Never use `canvas.NewText()` with manual sizing.

```go
// ✅ CORRECT
title := widget.NewRichTextFromMarkdown("## My Section")
plain := widget.NewLabel("Status: Ready")

// ❌ WRONG
text := canvas.NewText("Title", theme.ForegroundColor())
text.TextSize = 20
```

## Step 5 — Register the tab/view

Add it to the tab container in `internal/gui/gui.go` or wherever tabs are assembled.

## Step 6 — Write tests

```go
package views

import (
    "testing"

    fynetest "fyne.io/fyne/v2/test"

    "aws-profile-manager/internal/test"
)

func TestNew<Name>View(t *testing.T) {
    test.SetupTestEnvironment(t)

    testApp := fynetest.NewApp()
    window := testApp.NewWindow("test")

    view := New<Name>View(window)
    if view == nil {
        t.Fatal("expected non-nil view")
    }
}
```

## Rules checklist

- [ ] ViewModel does NOT create Fyne widgets
- [ ] View does NOT call business logic packages directly — goes through ViewModel
- [ ] All goroutines in the View update UI only via `fyne.Do()`
- [ ] All dialogs use `components.ShowCustomDialog()`
- [ ] All text uses `widget.NewRichTextFromMarkdown()` or `widget.NewLabel()` — never `canvas.NewText()`
- [ ] No `exec.Command()` or raw `go func() { doWork() }()` in ViewModel — use task package
- [ ] Error display uses `dialog.ShowError(err, window)` — never `fmt.Println()`
