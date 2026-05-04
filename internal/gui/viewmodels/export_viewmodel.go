package viewmodels

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/backup"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ExportViewModel manages the state and business logic for profile export operations.
//
// This ViewModel follows MVVM pattern by:
//   - Separating business logic (backup API calls) from presentation (View)
//   - Managing export state (in progress, results)
//   - Handling dependency injection (building Config from settings)
//   - Coordinating with task reporters for progress updates
type ExportViewModel struct {
	IsExporting bool // Is an export operation currently in progress?
	mu          sync.RWMutex
}

// ExportOptions contains all user selections for export operation.
//
// These options map directly to backup.ExportOptions but are GUI-friendly
// with clear naming and defaults suitable for user interaction.
type ExportOptions struct {
	OutputPath      string
	IncludeManaged  bool
	IncludeAbove    bool
	IncludeBelow    bool
	Description     string
	ExcludeSettings bool
}

// ExportResult contains the results of an export operation for GUI display.
//
// Wraps backup.ExportResult with additional GUI-specific information.
type ExportResult struct {
	Success          bool
	OutputPath       string
	TotalProfiles    int
	ManagedProfiles  int
	UnmanagedAbove   int
	UnmanagedBelow   int
	SettingsExported bool
	Error            error

	// Detailed statistics from generators
	ManagedStats   generators.SectionStats // Comprehensive stats for managed section
	UnmanagedStats generators.SectionStats // Combined stats for unmanaged sections
}

// NewExportViewModel creates a new export view model and registers it.
func NewExportViewModel() *ExportViewModel {
	logging.Debug.Log("\t🔹 Creating export view model")

	vm := &ExportViewModel{
		IsExporting: false,
	}

	// Register in core state
	core.App.RegisterState("export-dialog", vm)

	logging.Debug.Log("\t🔹 Export view model created")
	return vm
}

// ExportProfiles performs the export operation with proper dependency injection.
//
// This is the main business logic method that:
//  1. Builds Config from settings (DI)
//  2. Calls backup.ExportProfiles with context and reporter
//  3. Transforms result for GUI consumption
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Export options from UI
//   - reporter: Progress reporter (typically task.NoOpReporter for GUI)
//
// Returns:
//   - *ExportResult: Results for GUI display
func (vm *ExportViewModel) ExportProfiles(ctx context.Context, opts ExportOptions, reporter task.Reporter) *ExportResult {
	vm.mu.Lock()
	vm.IsExporting = true
	vm.mu.Unlock()

	defer func() {
		vm.mu.Lock()
		vm.IsExporting = false
		vm.mu.Unlock()
	}()

	logging.Debug.Log("Export view model: Starting export")

	// Build Config from settings (dependency injection)
	currentSettings := settings.Get()
	cfg := backup.Config{
		ConfigPath:  filepath.Join(settings.GetAwsDir(), "config"),
		AwsDir:      settings.GetAwsDir(),
		StartMarker: currentSettings.Application.GetFormattedStartMarker(),
		EndMarker:   currentSettings.Application.GetFormattedEndMarker(),
	}

	// Build backup options
	backupOpts := backup.ExportOptions{
		OutputPath:      opts.OutputPath,
		IncludeManaged:  opts.IncludeManaged,
		IncludeAbove:    opts.IncludeAbove,
		IncludeBelow:    opts.IncludeBelow,
		Description:     opts.Description,
		ExcludeSettings: opts.ExcludeSettings,
	}

	// Call backup package with proper DI
	result, err := backup.ExportProfiles(ctx, cfg, backupOpts, reporter)
	if err != nil {
		_ = logging.Log.Error("Export failed", "error", err)
		return &ExportResult{
			Success: false,
			Error:   err,
		}
	}

	logging.Log.Success("Export completed successfully")

	// Transform to GUI result
	return &ExportResult{
		Success:          true,
		OutputPath:       result.OutputPath,
		TotalProfiles:    result.TotalProfiles,
		ManagedProfiles:  result.ManagedProfiles,
		UnmanagedAbove:   result.UnmanagedAbove,
		UnmanagedBelow:   result.UnmanagedBelow,
		SettingsExported: result.SettingsExported,
		ManagedStats:     result.ManagedStats,
		UnmanagedStats:   result.UnmanagedStats,
		Error:            nil,
	}
}

// GetExportMode returns a human-readable description of what will be exported.
//
// This helper method provides consistent messaging across the GUI.
func (vm *ExportViewModel) GetExportMode(opts ExportOptions) string {
	if opts.IncludeManaged && !opts.IncludeAbove && !opts.IncludeBelow {
		return "Managed profiles only (installer config)"
	} else if !opts.IncludeManaged && (opts.IncludeAbove || opts.IncludeBelow) {
		return "Personal profiles only"
	} else if opts.IncludeManaged && opts.IncludeAbove && opts.IncludeBelow {
		return "Full backup (managed + personal profiles)"
	} else {
		return "Custom selection"
	}
}

// FormatResult returns a styled widget summarising the export result.
//
// This generates the success content shown to the user after export.
// On failure, returns a simple error label.
func (vm *ExportViewModel) FormatResult(result *ExportResult, opts ExportOptions) fyne.CanvasObject {
	if !result.Success {
		return widget.NewLabel(fmt.Sprintf("Export failed: %v", result.Error))
	}

	mode := vm.GetExportMode(opts)
	personalTotal := result.UnmanagedAbove + result.UnmanagedBelow

	settingsText := "⊘ Application settings excluded from backup"
	if result.SettingsExported {
		settingsText = "✓ Application settings included in backup"
	}

	items := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**Mode:** %s", mode)),
	}

	items = appendStatsSection(items, result.ManagedStats, result.ManagedProfiles, "📦 Managed Data")
	items = appendStatsSection(items, result.UnmanagedStats, personalTotal, "👤 Personal Profiles")

	items = append(items,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown(fmt.Sprintf("📊 **Grand Total:** %d profiles", result.TotalProfiles)),
		widget.NewRichTextFromMarkdown(settingsText),
	)

	return container.NewVBox(items...)
}

// StartExport executes the export asynchronously using the task package.
//
// This follows the same pattern as StartInstall — the ViewModel owns the goroutine
// via task.RunAsync, keeping the view free of concurrency concerns.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Export options built from the UI
//   - reporter: Progress reporter (typically GuiProgressReporter)
//   - onComplete: Callback invoked on the goroutine thread with the result
func (vm *ExportViewModel) StartExport(
	ctx context.Context,
	opts ExportOptions,
	reporter task.Reporter,
	onComplete func(*ExportResult),
) {
	logging.Debug.Log("Export view model: StartExport triggered")

	var exportResult *ExportResult

	asyncTask := &task.FunctionTask{
		Name: "export-profiles-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			exportResult = vm.ExportProfiles(runCtx, opts, runReporter)
			if !exportResult.Success {
				return nil, exportResult.Error
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		logging.Debug.Log("Export view model: StartExport complete", "error", err)
		if exportResult == nil {
			exportResult = &ExportResult{Success: false, Error: err}
		}
		if onComplete != nil {
			onComplete(exportResult)
		}
	})
}
