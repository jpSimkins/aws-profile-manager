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
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ImportViewModel manages the state and business logic for profile import operations.
//
// This ViewModel follows MVVM pattern by:
//   - Separating business logic (backup API calls) from presentation (View)
//   - Managing import state (in progress, results, preview)
//   - Handling dependency injection (building Config from settings)
//   - Providing preview functionality via dry-run mode
type ImportViewModel struct {
	IsImporting bool // Is an import operation currently in progress?

	// Cached plan from preview (reused for import to avoid re-processing)
	cachedPlan       *profiles.ImportPlan
	cachedBackupPath string // Track which backup the plan is for

	mu sync.RWMutex
}

// ImportOptions contains all user selections for import operation.
//
// These options map directly to backup.ImportOptions but are GUI-friendly
// with clear naming and defaults suitable for user interaction.
type ImportOptions struct {
	BackupPath            string
	IncludeManaged        bool
	IncludeAbove          bool
	IncludeBelow          bool
	IgnoreSettings        bool
	BackupCurrentSettings bool
	GenerateCheatSheet    bool // Generate cheat sheet after import
}

// ImportResult contains the results of an import operation for GUI display.
//
// Wraps backup.ImportResult with additional GUI-specific information.
type ImportResult struct {
	Success             bool
	ConfigPath          string
	SettingsRestored    bool
	SettingsBackupPath  string
	CheatSheetGenerated bool
	CheatSheetPath      string
	Error               error

	// Raw statistics from generators (what's in the backup/merged collection)
	ManagedStats        generators.SectionStats // Comprehensive stats for managed section
	UnmanagedAboveStats generators.SectionStats // Stats for unmanaged above section (includes duplicates)
	UnmanagedBelowStats generators.SectionStats // Stats for unmanaged below section (includes duplicates)

	// Duplicate statistics (what was skipped)
	ManagedDuplicates        profiles.SectionDuplicateStats // Duplicates in managed section
	UnmanagedAboveDuplicates profiles.SectionDuplicateStats // Duplicates in above section
	UnmanagedBelowDuplicates profiles.SectionDuplicateStats // Duplicates in below section
}

// ImportPreview contains preview information for GUI display before import.
//
// Generated using PrepareImport to show what will be imported with full detail.
type ImportPreview struct {
	Success     bool
	HasSettings bool
	Description string // Optional description from backup metadata
	Error       error

	// Raw statistics from generators (what's in the backup/merged collection)
	ManagedStats        generators.SectionStats // Comprehensive stats for managed section
	UnmanagedAboveStats generators.SectionStats // Stats for unmanaged above section (includes duplicates)
	UnmanagedBelowStats generators.SectionStats // Stats for unmanaged below section (includes duplicates)

	// Duplicate statistics (what WON'T be written)
	// GUI uses these to calculate actual profiles written = Raw - Duplicates
	ManagedDuplicates        profiles.SectionDuplicateStats // Duplicates in managed section
	UnmanagedAboveDuplicates profiles.SectionDuplicateStats // Duplicates in above section
	UnmanagedBelowDuplicates profiles.SectionDuplicateStats // Duplicates in below section
}

// NewImportViewModel creates a new import view model and registers it.
func NewImportViewModel() *ImportViewModel {
	logging.Debug.Log("\t🔹 Creating import view model")

	vm := &ImportViewModel{
		IsImporting: false,
	}

	// Register in core state
	core.App.RegisterState("import-dialog", vm)

	logging.Debug.Log("\t🔹 Import view model created")
	return vm
}

// GetImportPreview generates a preview of what will be imported using dry-run mode.
//
// This allows the GUI to show the user what will be imported before they commit.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Import options (will run in dry-run mode)
//   - reporter: Progress reporter
//
// Returns:
//   - *ImportPreview: Preview information for display
func (vm *ImportViewModel) GetImportPreview(ctx context.Context, opts ImportOptions, reporter task.Reporter) *ImportPreview {
	logging.Debug.Log("Import view model: Getting preview")

	// Build Config from settings (dependency injection)
	currentSettings := settings.Get()
	cfg := backup.Config{
		ConfigPath:  filepath.Join(settings.GetAwsDir(), "config"),
		AwsDir:      settings.GetAwsDir(),
		StartMarker: currentSettings.Application.GetFormattedStartMarker(),
		EndMarker:   currentSettings.Application.GetFormattedEndMarker(),
	}

	// Build backup options for preview
	backupOpts := backup.ImportOptions{
		BackupPath:            opts.BackupPath,
		IncludeManaged:        opts.IncludeManaged,
		IncludeAbove:          opts.IncludeAbove,
		IncludeBelow:          opts.IncludeBelow,
		IgnoreSettings:        opts.IgnoreSettings,
		BackupCurrentSettings: false,
		DryRun:                false, // Not used by PrepareImport
	}

	// Call PrepareImport ONLY (no writes - just parse and generate)
	plan, err := backup.PrepareImport(ctx, cfg, backupOpts, reporter)
	if err != nil {
		_ = logging.Log.Error("Preview failed", "error", err)
		return &ImportPreview{
			Success: false,
			Error:   err,
		}
	}

	// Read backup file to check for settings
	backupFile, err := backup.ReadBackupFile(opts.BackupPath)
	if err != nil {
		_ = logging.Log.Error("Failed to read backup file", "error", err)
		return &ImportPreview{
			Success: false,
			Error:   err,
		}
	}

	hasSettings := backupFile.Settings != nil

	// Cache the plan for later use (avoid re-processing on import)
	vm.mu.Lock()
	vm.cachedPlan = plan
	vm.cachedBackupPath = opts.BackupPath
	vm.mu.Unlock()

	logging.Debug.Log("Plan cached for import", "backup", opts.BackupPath)

	preview := &ImportPreview{
		Success:                  true,
		HasSettings:              hasSettings,
		Description:              backupFile.Metadata.Description,
		ManagedStats:             plan.ManagedStats,
		UnmanagedAboveStats:      plan.UnmanagedAboveStats,
		UnmanagedBelowStats:      plan.UnmanagedBelowStats,
		ManagedDuplicates:        plan.ManagedDuplicates,
		UnmanagedAboveDuplicates: plan.UnmanagedAboveDuplicates,
		UnmanagedBelowDuplicates: plan.UnmanagedBelowDuplicates,
		Error:                    nil,
	}

	logging.Debug.Log("Preview created",
		"managed_profiles", preview.ManagedStats.ProfilesWritten,
		"above_profiles", preview.UnmanagedAboveStats.ProfilesWritten,
		"below_profiles", preview.UnmanagedBelowStats.ProfilesWritten,
		"above_duplicates", preview.UnmanagedAboveDuplicates.TotalDuplicates,
		"below_duplicates", preview.UnmanagedBelowDuplicates.TotalDuplicates,
	)

	return preview
}

// ImportProfiles performs the import operation with proper dependency injection.
//
// This is the main business logic method that:
//  1. Builds Config from settings (DI)
//  2. Calls backup.ImportProfiles with context and reporter
//  3. Transforms result for GUI consumption
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Import options from UI
//   - reporter: Progress reporter (typically task.NoOpReporter for GUI)
//
// Returns:
//   - *ImportResult: Results for GUI display
func (vm *ImportViewModel) ImportProfiles(ctx context.Context, opts ImportOptions, reporter task.Reporter) *ImportResult {
	vm.mu.Lock()
	vm.IsImporting = true
	vm.mu.Unlock()

	defer func() {
		vm.mu.Lock()
		vm.IsImporting = false
		vm.mu.Unlock()
	}()

	logging.Debug.Log("Import view model: Starting import")

	// Build Config from settings (dependency injection)
	currentSettings := settings.Get()
	cfg := backup.Config{
		ConfigPath:  filepath.Join(settings.GetAwsDir(), "config"),
		AwsDir:      settings.GetAwsDir(),
		StartMarker: currentSettings.Application.GetFormattedStartMarker(),
		EndMarker:   currentSettings.Application.GetFormattedEndMarker(),
	}

	// Build backup options
	backupOpts := backup.ImportOptions{
		BackupPath:            opts.BackupPath,
		IncludeManaged:        opts.IncludeManaged,
		IncludeAbove:          opts.IncludeAbove,
		IncludeBelow:          opts.IncludeBelow,
		IgnoreSettings:        opts.IgnoreSettings,
		BackupCurrentSettings: opts.BackupCurrentSettings,
		GenerateCheatSheet:    opts.GenerateCheatSheet,
	}

	// Check if we have a cached plan from preview (avoid re-processing)
	vm.mu.RLock()
	plan := vm.cachedPlan
	cachedPath := vm.cachedBackupPath
	vm.mu.RUnlock()

	var result *backup.ImportResult
	var err error

	if plan != nil && cachedPath == opts.BackupPath {
		// Use cached plan from preview (NO re-processing!)
		logging.Debug.Log("Using cached plan from preview - skipping re-processing")
		result, err = backup.ExecuteImport(ctx, cfg, plan, backupOpts, reporter)

		// Clear cache after use
		vm.mu.Lock()
		vm.cachedPlan = nil
		vm.cachedBackupPath = ""
		vm.mu.Unlock()
	} else {
		// No cached plan (user didn't preview or changed options)
		logging.Debug.Log("No cached plan - doing full import")
		result, err = backup.ImportProfiles(ctx, cfg, backupOpts, reporter)
	}
	if err != nil {
		_ = logging.Log.Error("Import failed", "error", err)
		return &ImportResult{
			Success: false,
			Error:   err,
		}
	}

	logging.Log.Success("Import completed successfully")

	// Transform to GUI result (copy stats from backup result)
	return &ImportResult{
		Success:                  true,
		ConfigPath:               result.ConfigPath,
		SettingsRestored:         result.SettingsRestored,
		SettingsBackupPath:       result.SettingsBackupPath,
		CheatSheetGenerated:      result.CheatSheetGenerated,
		CheatSheetPath:           result.CheatSheetPath,
		ManagedStats:             result.ManagedStats,
		UnmanagedAboveStats:      result.UnmanagedAboveStats,
		UnmanagedBelowStats:      result.UnmanagedBelowStats,
		ManagedDuplicates:        result.ManagedDuplicates,
		UnmanagedAboveDuplicates: result.UnmanagedAboveDuplicates,
		UnmanagedBelowDuplicates: result.UnmanagedBelowDuplicates,
		Error:                    nil,
	}
}

// GetImportMode returns a human-readable description of what will be imported.
//
// This helper method provides consistent messaging across the GUI.
func (vm *ImportViewModel) GetImportMode(opts ImportOptions) string {
	if opts.IncludeManaged && !opts.IncludeAbove && !opts.IncludeBelow {
		return "Managed profiles only"
	} else if !opts.IncludeManaged && (opts.IncludeAbove || opts.IncludeBelow) {
		return "Personal profiles only"
	} else if opts.IncludeManaged && opts.IncludeAbove && opts.IncludeBelow {
		return "Full restore (managed + personal profiles)"
	} else {
		return "Custom selection"
	}
}

// FormatPreview returns a styled widget showing a preview of what will be imported.
//
// This generates the preview content shown to the user before import.
// On failure, returns a simple error label.
func (vm *ImportViewModel) FormatPreview(preview *ImportPreview, opts ImportOptions) fyne.CanvasObject {
	if !preview.Success {
		return widget.NewLabel(fmt.Sprintf("Preview failed: %v", preview.Error))
	}

	// Calculate actual profiles that will be written.
	// Managed: net new (ProfilesWritten - duplicates) - the managed section is replaced from schema.
	// Unmanaged: full merged count (ProfilesWritten) - the merged above/below sections are written entirely.
	managedActual := preview.ManagedStats.ProfilesWritten - preview.ManagedDuplicates.TotalDuplicates
	aboveTotal := preview.UnmanagedAboveStats.ProfilesWritten
	belowTotal := preview.UnmanagedBelowStats.ProfilesWritten
	totalActual := managedActual + aboveTotal + belowTotal

	settingsText := "⊘ No settings in backup"
	if preview.HasSettings {
		settingsText = "✓ Application settings will be restored"
	}

	items := []fyne.CanvasObject{}

	items = appendStatsSection(items, preview.ManagedStats, managedActual, "📦 Managed Data", preview.ManagedDuplicates)
	items = appendStatsSection(items, preview.UnmanagedAboveStats, aboveTotal, "👤 Personal Profiles (Above)", preview.UnmanagedAboveDuplicates)
	items = appendStatsSection(items, preview.UnmanagedBelowStats, belowTotal, "👤 Personal Profiles (Below)", preview.UnmanagedBelowDuplicates)

	items = append(items,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown(fmt.Sprintf("## Grand Total\n\n✓ %d profiles will be written", totalActual)),
		widget.NewRichTextFromMarkdown(settingsText),
	)

	return container.NewVBox(items...)
}

// FormatResult returns a styled widget summarising the import result.
//
// This generates the success content shown to the user after import.
// On failure, returns a simple error label.
func (vm *ImportViewModel) FormatResult(result *ImportResult, opts ImportOptions) fyne.CanvasObject {
	if !result.Success {
		return widget.NewLabel(fmt.Sprintf("Import failed: %v", result.Error))
	}

	mode := vm.GetImportMode(opts)

	// Calculate actual profiles written.
	// Managed: net new (ProfilesWritten - duplicates) - the managed section is replaced from schema.
	// Unmanaged: full merged count (ProfilesWritten) - the merged above/below sections are written entirely,
	// including any pre-existing profiles that were preserved through the merge.
	managedWritten := result.ManagedStats.ProfilesWritten - result.ManagedDuplicates.TotalDuplicates
	aboveTotal := result.UnmanagedAboveStats.ProfilesWritten
	belowTotal := result.UnmanagedBelowStats.ProfilesWritten
	totalWritten := managedWritten + aboveTotal + belowTotal
	totalInBackup := result.ManagedStats.ProfilesWritten + aboveTotal + belowTotal

	settingsText := "⊘ No settings restored"
	if result.SettingsRestored {
		settingsText = "✓ Application settings restored from backup"
		if result.SettingsBackupPath != "" {
			settingsText += fmt.Sprintf("\n\n- Previous settings backed up to: [%s](file://%s)", filepath.Base(result.SettingsBackupPath), result.SettingsBackupPath)
		}
	}

	items := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown(fmt.Sprintf("**Mode:** %s", mode)),
	}

	items = appendStatsSection(items, result.ManagedStats, managedWritten, "📦 Managed Data", result.ManagedDuplicates)
	items = appendStatsSection(items, result.UnmanagedAboveStats, aboveTotal, "👤 Personal Profiles (Above)", result.UnmanagedAboveDuplicates)
	items = appendStatsSection(items, result.UnmanagedBelowStats, belowTotal, "👤 Personal Profiles (Below)", result.UnmanagedBelowDuplicates)

	grandTotalText := fmt.Sprintf("## Grand Total\n\n✓ %d profiles written\n\n- Backup contained: %d profiles", totalWritten, totalInBackup)

	items = append(items,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown(grandTotalText),
		widget.NewRichTextFromMarkdown(settingsText),
	)

	if result.CheatSheetGenerated {
		items = append(items, widget.NewRichTextFromMarkdown("✓ Cheat sheet generated"))
	}

	return container.NewVBox(items...)
}

// StartPreview runs GetImportPreview asynchronously using the task package.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Import options
//   - reporter: Progress reporter (typically GuiProgressReporter)
//   - onComplete: Callback invoked with the preview result
func (vm *ImportViewModel) StartPreview(
	ctx context.Context,
	opts ImportOptions,
	reporter task.Reporter,
	onComplete func(*ImportPreview),
) {
	logging.Debug.Log("Import view model: StartPreview triggered")

	var preview *ImportPreview

	asyncTask := &task.FunctionTask{
		Name: "import-preview-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			preview = vm.GetImportPreview(runCtx, opts, runReporter)
			if !preview.Success {
				return nil, preview.Error
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		logging.Debug.Log("Import view model: StartPreview complete", "error", err)
		if preview == nil {
			preview = &ImportPreview{Success: false, Error: err}
		}
		if onComplete != nil {
			onComplete(preview)
		}
	})
}

// StartImport runs ImportProfiles asynchronously using the task package.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Import options built from the UI
//   - reporter: Progress reporter (typically GuiProgressReporter)
//   - onComplete: Callback invoked with the import result
func (vm *ImportViewModel) StartImport(
	ctx context.Context,
	opts ImportOptions,
	reporter task.Reporter,
	onComplete func(*ImportResult),
) {
	logging.Debug.Log("Import view model: StartImport triggered")

	var importResult *ImportResult

	asyncTask := &task.FunctionTask{
		Name: "import-profiles-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			importResult = vm.ImportProfiles(runCtx, opts, runReporter)
			if !importResult.Success {
				return nil, importResult.Error
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		logging.Debug.Log("Import view model: StartImport complete", "error", err)
		if importResult == nil {
			importResult = &ImportResult{Success: false, Error: err}
		}
		if onComplete != nil {
			onComplete(importResult)
		}
	})
}
