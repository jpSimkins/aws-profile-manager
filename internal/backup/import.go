package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/security"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ImportProfiles imports AWS CLI profiles and settings from a backup file.
//
// This function orchestrates:
//  1. Reading and validating backup file
//  2. Backing up current settings if requested
//  3. Calling profiles.Import() for profile restoration
//  4. Restoring settings if present
//
// Parameters:
//   - ctx: Context for cancellation
//   - cfg: Configuration (paths)
//   - opts: Import options
//   - reporter: Progress reporter (task.Reporter)
//
// Returns:
//   - *ImportResult: Statistics and paths
//   - error: Any error encountered
//
// Example:
//
//	cfg := backup.Config{
//	    ConfigPath:  settings.GetAwsConfigPath(),
//	    AwsDir:      settings.GetAwsDir(),
//	    StartMarker: app.GetFormattedStartMarker(),
//	    EndMarker:   app.GetFormattedEndMarker(),
//	}
//	opts := backup.ImportOptions{
//	    BackupPath:     "/path/to/backup.json",
//	    IncludeManaged: true,
//	    IncludeAbove:   true,
//	    IncludeBelow:   true,
//	}
//	result, err := backup.ImportProfiles(ctx, cfg, opts, reporter)

// PrepareImport prepares an import by parsing backup and generating content.
//
// This is Phase 1 of a two-phase import. Use this for GUI preview functionality
// where you want to show what WILL be imported before actually writing files.
//
// The returned ImportPlan contains:
//   - Parsed schema
//   - Pre-generated content (cached)
//   - Accurate profile counts
//
// Pass the plan to ExecuteImport to complete the import without re-processing.
//
// Parameters:
//   - ctx: Context for cancellation
//   - cfg: Configuration (paths and markers)
//   - opts: Import options (what to include)
//   - reporter: Progress reporter
//
// Returns:
//   - *profiles.ImportPlan: Prepared import with cached content
//   - error: Any error during preparation
//
// Example:
//
//	plan, err := backup.PrepareImport(ctx, cfg, opts, reporter)
//	// Show preview...
//	result, err := backup.ExecuteImport(ctx, cfg, plan, opts, reporter)
func PrepareImport(
	ctx context.Context,
	cfg Config,
	opts ImportOptions,
	reporter task.Reporter,
) (*profiles.ImportPlan, error) {
	reporter.ReportStatus("Preparing import...")

	// Build profiles.Config from backup.Config
	profilesConfig := profiles.Config{
		ConfigPath:          cfg.ConfigPath,
		AwsDir:              cfg.AwsDir,
		StartMarker:         cfg.StartMarker,
		EndMarker:           cfg.EndMarker,
		CheatSheetOutputDir: "", // Not needed for import
		CacheDir:            "", // Not needed for import
	}

	// Create importer
	importer := profiles.NewImporter(profilesConfig)

	// Build import options for profiles package
	profilesOpts := profiles.ImportOptions{
		BackupPath:     opts.BackupPath,
		IncludeManaged: opts.IncludeManaged,
		IncludeAbove:   opts.IncludeAbove,
		IncludeBelow:   opts.IncludeBelow,
		DryRun:         false, // Not applicable for PrepareImport
	}

	// Prepare import (parse + generate)
	plan, err := importer.PrepareImport(ctx, profilesOpts, reporter)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare import: %w", err)
	}

	return plan, nil
}

// ExecuteImport executes a prepared import by writing cached content.
//
// This is Phase 2 of a two-phase import. It takes the ImportPlan from
// PrepareImport and writes the pre-generated content to files. Much faster
// than ImportProfiles since no re-parsing or re-generation happens.
//
// This also handles:
//   - Settings backup (if requested)
//   - Settings restoration on cancellation
//   - Cheat sheet generation (if requested)
//
// Parameters:
//   - ctx: Context for cancellation
//   - cfg: Configuration (paths and markers)
//   - plan: Prepared import plan from PrepareImport
//   - opts: Import options (same as used in PrepareImport)
//   - reporter: Progress reporter
//
// Returns:
//   - *ImportResult: Import results with actual counts
//   - error: Any error during execution
//
// Example:
//
//	plan, _ := backup.PrepareImport(ctx, cfg, opts, reporter)
//	result, err := backup.ExecuteImport(ctx, cfg, plan, opts, reporter)
func ExecuteImport(
	ctx context.Context,
	cfg Config,
	plan *profiles.ImportPlan,
	opts ImportOptions,
	reporter task.Reporter,
) (*ImportResult, error) {
	startTime := time.Now()
	reporter.ReportStatus("Executing import...")

	// Step 1: Backup current settings if requested
	var settingsBackupPath string
	if opts.BackupCurrentSettings && !opts.IgnoreSettings && !opts.DryRun {
		reporter.ReportStatus("Backing up current settings...")

		var err error
		settingsBackupPath, err = BackupSettings(cfg)
		if err != nil {
			logging.Log.Warn("Failed to backup current settings",
				"error", err,
			)
			// Continue anyway - this is not fatal
		} else {
			logging.Log.Info("Current settings backed up",
				"path", settingsBackupPath,
			)
		}
	}

	// Setup deferred cleanup: restore settings if cancelled
	defer func() {
		// Check if operation was cancelled
		if ctx.Err() == context.Canceled && settingsBackupPath != "" {
			logging.Log.Info("Import cancelled - restoring settings from backup",
				"backup", settingsBackupPath,
			)

			// Read settings from backup file (JSON file contains settings directly)
			data, err := security.ReadFile(settingsBackupPath, security.ReadOptions{
				AllowedExtensions: []string{".json"},
			})
			if err != nil {
				_ = logging.Log.Error("Failed to read settings backup for restore",
					"error", err,
					"path", settingsBackupPath,
				)
				return
			}

			var backupSettings settings.Settings
			if err := json.Unmarshal(data, &backupSettings); err != nil {
				_ = logging.Log.Error("Failed to parse settings backup",
					"error", err,
				)
				return
			}

			// Restore settings
			if err := RestoreSettings(&backupSettings); err != nil {
				_ = logging.Log.Error("Failed to restore settings after cancellation",
					"error", err,
				)
			} else {
				logging.Log.Info("Settings successfully restored after cancellation")
			}
		}
	}()

	// Step 2: Execute the prepared import
	profilesConfig := profiles.Config{
		ConfigPath:          cfg.ConfigPath,
		AwsDir:              cfg.AwsDir,
		StartMarker:         cfg.StartMarker,
		EndMarker:           cfg.EndMarker,
		CheatSheetOutputDir: settings.GetDesktopDir(),
		CacheDir:            "",
	}

	importer := profiles.NewImporter(profilesConfig)

	profilesOpts := profiles.ImportOptions{
		BackupPath:     opts.BackupPath,
		IncludeManaged: opts.IncludeManaged,
		IncludeAbove:   opts.IncludeAbove,
		IncludeBelow:   opts.IncludeBelow,
		DryRun:         opts.DryRun,
	}

	profilesResult, err := importer.ExecuteImport(ctx, plan, profilesOpts, reporter)
	if err != nil {
		// Don't wrap cancellation errors
		if err == context.Canceled {
			return nil, err
		}
		return nil, fmt.Errorf("failed to execute import: %w", err)
	}

	// Step 3: Restore settings if present and requested
	settingsRestored := false
	if !opts.IgnoreSettings && !opts.DryRun && opts.BackupPath != "" {
		reporter.ReportStatus("Checking backup for settings...")

		backupFile, readErr := ReadBackupFile(opts.BackupPath)
		if readErr != nil {
			logging.Log.Warn("Could not read backup file for settings restoration",
				"error", readErr,
			)
		} else if backupFile.Settings != nil {
			reporter.ReportStatus("Restoring application settings...")

			if restoreErr := RestoreSettings(backupFile.Settings); restoreErr != nil {
				return nil, fmt.Errorf("failed to restore settings: %w", restoreErr)
			}

			settingsRestored = true
			logging.Log.Info("Settings restored via ExecuteImport")
		}
	}

	// Step 4: Generate cheat sheet if requested
	var cheatSheetGenerated bool
	var cheatSheetPath string

	if opts.GenerateCheatSheet && opts.IncludeManaged && !opts.DryRun && plan.Schema.Managed != nil {
		reporter.ReportStatus("Generating cheat sheet...")

		// Use profiles.Install to generate cheat sheet
		installer := profiles.NewInstaller(profilesConfig)

		installOpts := profiles.InstallOptions{
			Schema:             plan.Schema,
			GenerateCheatSheet: true,
		}

		installResult, err := installer.Install(ctx, installOpts, reporter)
		if err != nil {
			logging.Log.Warn("Failed to generate cheat sheet",
				"error", err,
			)
		} else if installResult.CheatSheetPath != "" {
			cheatSheetGenerated = true
			cheatSheetPath = installResult.CheatSheetPath
		}
	} else if opts.GenerateCheatSheet && !opts.IncludeManaged {
		logging.Log.Info("Skipping cheat sheet generation because managed import is disabled")
	}

	// Build result
	duration := time.Since(startTime)
	result := &ImportResult{
		ConfigPath:               cfg.ConfigPath,
		SettingsRestored:         settingsRestored,
		SettingsBackupPath:       settingsBackupPath,
		CheatSheetGenerated:      cheatSheetGenerated,
		CheatSheetPath:           cheatSheetPath,
		Duration:                 duration,
		ManagedStats:             profilesResult.ManagedStats,
		UnmanagedAboveStats:      profilesResult.UnmanagedAboveStats,
		UnmanagedBelowStats:      profilesResult.UnmanagedBelowStats,
		ManagedDuplicates:        profilesResult.ManagedDuplicates,
		UnmanagedAboveDuplicates: profilesResult.UnmanagedAboveDuplicates,
		UnmanagedBelowDuplicates: profilesResult.UnmanagedBelowDuplicates,
		// Note: BackupFile not set since we used ImportPlan directly
	}

	return result, nil
}

func ImportProfiles(
	ctx context.Context,
	cfg Config,
	opts ImportOptions,
	reporter task.Reporter,
) (*ImportResult, error) {
	startTime := time.Now()

	reporter.ReportStatus("Starting profile import...")

	// Validate options
	if opts.BackupPath == "" {
		return nil, fmt.Errorf("backup path is required")
	}

	// Check if anything to import
	if !opts.IncludeManaged && !opts.IncludeAbove && !opts.IncludeBelow && opts.IgnoreSettings {
		return nil, fmt.Errorf("no content selected for import (all sections disabled)")
	}

	// Step 1: Prepare import (parse + generate content + compute stats ONCE)
	plan, err := PrepareImport(ctx, cfg, opts, reporter)
	if err != nil {
		return nil, err
	}

	// Step 2: Read backup file for settings
	reporter.ReportStatus(fmt.Sprintf("Reading backup file from %s...", opts.BackupPath))

	backupFile, err := ReadBackupFile(opts.BackupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	logging.Log.Info("Backup file loaded",
		"version", backupFile.Version,
		"has_data", backupFile.Data != nil,
		"has_settings", backupFile.Settings != nil,
	)

	// Step 2: Backup current settings if requested
	var settingsBackupPath string
	if opts.BackupCurrentSettings && !opts.IgnoreSettings && backupFile.Settings != nil && !opts.DryRun {
		reporter.ReportStatus("Backing up current settings...")

		settingsBackupPath, err = BackupSettings(cfg)
		if err != nil {
			logging.Log.Warn("Failed to backup current settings",
				"error", err,
			)
			// Continue anyway - this is not fatal
		} else {
			logging.Log.Info("Current settings backed up",
				"path", settingsBackupPath,
			)
		}
	}

	// Setup deferred cleanup: restore settings if cancelled
	defer func() {
		// Check if operation was cancelled
		if ctx.Err() == context.Canceled && settingsBackupPath != "" {
			logging.Log.Info("Import cancelled - restoring settings from backup",
				"backup", settingsBackupPath,
			)

			// Read settings from backup file (JSON file contains settings directly)
			data, err := security.ReadFile(settingsBackupPath, security.ReadOptions{
				AllowedExtensions: []string{".json"},
			})
			if err != nil {
				_ = logging.Log.Error("Failed to read settings backup for restore",
					"error", err,
					"path", settingsBackupPath,
				)
				return
			}

			var backupSettings settings.Settings
			if err := json.Unmarshal(data, &backupSettings); err != nil {
				_ = logging.Log.Error("Failed to parse settings backup",
					"error", err,
				)
				return
			}

			// Restore settings
			if err := RestoreSettings(&backupSettings); err != nil {
				_ = logging.Log.Error("Failed to restore settings after cancellation",
					"error", err,
				)
			} else {
				logging.Log.Info("Settings successfully restored after cancellation")
			}
		}
	}()

	// Step 3: For DryRun, return early with plan stats (NO actual import)
	if opts.DryRun {
		duration := time.Since(startTime)
		reporter.ReportStatus("Dry run complete - no changes made")
		logging.Log.Info("Import dry run complete",
			"duration", duration,
		)

		return &ImportResult{
			BackupFile:          backupFile,
			ConfigPath:          cfg.ConfigPath,
			SettingsRestored:    false,
			SettingsBackupPath:  "",
			CheatSheetGenerated: false,
			CheatSheetPath:      "",
			Duration:            duration,
			ManagedStats:        plan.ManagedStats,
			UnmanagedAboveStats: plan.UnmanagedAboveStats,
			UnmanagedBelowStats: plan.UnmanagedBelowStats,
		}, nil
	}

	// Step 4: Execute the prepared import (uses cached content from plan - NO re-processing)
	executeResult, err := ExecuteImport(ctx, cfg, plan, opts, reporter)
	if err != nil {
		return nil, err
	}

	logging.Log.Info("Profiles imported",
		"managed_profiles", executeResult.ManagedStats.ProfilesWritten,
		"above_profiles", executeResult.UnmanagedAboveStats.ProfilesWritten,
		"below_profiles", executeResult.UnmanagedBelowStats.ProfilesWritten,
		"above_duplicates", executeResult.UnmanagedAboveDuplicates.TotalDuplicates,
		"below_duplicates", executeResult.UnmanagedBelowDuplicates.TotalDuplicates,
	)

	// Step 5: Restore settings if present and requested
	settingsRestored := false
	if !opts.IgnoreSettings && backupFile.Settings != nil && !opts.DryRun {
		reporter.ReportStatus("Restoring application settings...")

		if err := RestoreSettings(backupFile.Settings); err != nil {
			return nil, fmt.Errorf("failed to restore settings: %w", err)
		}

		settingsRestored = true
		logging.Log.Info("Settings restored")
	}

	// Step 6: Generate cheat sheet if requested (AFTER settings restoration)
	// This is done last so that settings (which may control cheat sheet location in future)
	// are already restored before generation
	cheatSheetGenerated := false
	cheatSheetPath := ""
	if opts.GenerateCheatSheet && opts.IncludeManaged && backupFile.Data != nil && !opts.DryRun {
		reporter.ReportStatus("Generating cheat sheet...")

		// Use profiles.Install to generate cheat sheet
		// Pass desktop directory from settings (respects ENV variables)
		profilesConfig := profiles.Config{
			ConfigPath:          cfg.ConfigPath,
			AwsDir:              cfg.AwsDir,
			StartMarker:         cfg.StartMarker,
			EndMarker:           cfg.EndMarker,
			CheatSheetOutputDir: settings.GetDesktopDir(),
			CacheDir:            "",
		}

		installer := profiles.NewInstaller(profilesConfig)

		// Generate cheat sheet from the imported schema
		// This will also re-write profiles (idempotent operation)
		installOpts := profiles.InstallOptions{
			Schema:             backupFile.Data,
			GenerateCheatSheet: true,
			// Note: Not using DryRun so cheat sheet is actually written
		}

		installResult, err := installer.Install(ctx, installOpts, reporter)
		if err != nil {
			logging.Log.Warn("Failed to generate cheat sheet",
				"error", err,
			)
		} else if installResult.CheatSheetPath != "" {
			cheatSheetGenerated = true
			cheatSheetPath = installResult.CheatSheetPath
			logging.Log.Info("Cheat sheet generated",
				"path", cheatSheetPath,
			)
		}
	} else if opts.GenerateCheatSheet && !opts.IncludeManaged {
		logging.Log.Info("Skipping cheat sheet generation because managed import is disabled")
	}

	duration := time.Since(startTime)

	reporter.ReportStatus("Import complete")
	totalWritten := executeResult.ManagedStats.ProfilesWritten +
		executeResult.UnmanagedAboveStats.ProfilesWritten +
		executeResult.UnmanagedBelowStats.ProfilesWritten -
		executeResult.UnmanagedAboveDuplicates.TotalDuplicates -
		executeResult.UnmanagedBelowDuplicates.TotalDuplicates

	logging.Log.Success("Backup restored",
		"profiles", totalWritten,
		"settings", settingsRestored,
		"duration", duration,
	)

	// Build result from executeResult (reuse stats from ExecuteImport - no duplication)
	result := &ImportResult{
		BackupFile:               backupFile,
		ConfigPath:               executeResult.ConfigPath,
		SettingsRestored:         settingsRestored,
		SettingsBackupPath:       settingsBackupPath,
		CheatSheetGenerated:      executeResult.CheatSheetGenerated || cheatSheetGenerated,
		CheatSheetPath:           executeResult.CheatSheetPath,
		Duration:                 duration,
		ManagedStats:             executeResult.ManagedStats,
		UnmanagedAboveStats:      executeResult.UnmanagedAboveStats,
		UnmanagedBelowStats:      executeResult.UnmanagedBelowStats,
		ManagedDuplicates:        executeResult.ManagedDuplicates,
		UnmanagedAboveDuplicates: executeResult.UnmanagedAboveDuplicates,
		UnmanagedBelowDuplicates: executeResult.UnmanagedBelowDuplicates,
	}

	if cheatSheetPath != "" {
		result.CheatSheetPath = cheatSheetPath
	}

	return result, nil
}
