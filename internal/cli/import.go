package cli

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/backup"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ImportFlags holds all flags for the import command.
//
// This struct provides a typed container for import command flags, making
// flag parsing and validation cleaner.
type ImportFlags struct {
	BackupPath            string
	IncludeManaged        bool
	IncludeAbove          bool
	IncludeBelow          bool
	DryRun                bool
	IgnoreSettings        bool
	BackupCurrentSettings bool
	GenerateCheatSheet    bool
	Verbose               bool
}

// runImport executes the import command to restore AWS CLI profiles.
//
// This command handler imports AWS CLI profiles from a JSON backup file,
// with support for different import modes and duplicate detection.
//
// Command Flow:
//  1. Parse flags into ImportFlags struct
//  2. Build Config from settings (DI)
//  3. Build ImportOptions from flags
//  4. Call backup.ImportProfiles() with context and reporter (ONE call)
//  5. Display results
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (backup file path can be positional)
//
// Returns:
//   - error: Any error encountered during execution
func runImport(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Import command started")

	// Parse flags
	logging.Debug.Log("Parsing command flags")
	flags, err := parseImportFlags(cmd, args)
	if err != nil {
		return logging.Log.Errorf("failed to parse import flags: %v", err)
	}

	// Enable verbose logging if requested
	if flags.Verbose {
		logging.Log.Info("Verbose mode enabled")
	}

	// Show dry-run notice
	if flags.DryRun {
		logging.Log.Warn("🔍 Dry-run mode: No files will be modified")
	}

	// Build Config from settings (dependency injection)
	logging.Debug.Log("Building config from settings")
	currentSettings := settings.Get()
	cfg := backup.Config{
		ConfigPath:  filepath.Join(settings.GetAwsDir(), "config"),
		AwsDir:      settings.GetAwsDir(),
		StartMarker: currentSettings.Application.GetFormattedStartMarker(),
		EndMarker:   currentSettings.Application.GetFormattedEndMarker(),
	}

	// Build import options from flags
	logging.Debug.Log("Building import options")
	opts := backup.ImportOptions{
		BackupPath:            flags.BackupPath,
		IncludeManaged:        flags.IncludeManaged,
		IncludeAbove:          flags.IncludeAbove,
		IncludeBelow:          flags.IncludeBelow,
		DryRun:                flags.DryRun,
		IgnoreSettings:        flags.IgnoreSettings,
		BackupCurrentSettings: flags.BackupCurrentSettings,
		GenerateCheatSheet:    flags.GenerateCheatSheet,
	}

	// Create reporter for progress updates
	reporter := task.CliReporter{}

	// Call API to import profiles (business logic happens here)
	logging.Debug.Log("Calling backup.ImportProfiles API")
	ctx := context.Background()
	result, err := backup.ImportProfiles(ctx, cfg, opts, reporter)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to import profiles", err)
	}

	totalWritten := result.ManagedStats.ProfilesWritten +
		result.UnmanagedAboveStats.ProfilesWritten +
		result.UnmanagedBelowStats.ProfilesWritten -
		result.ManagedDuplicates.TotalDuplicates -
		result.UnmanagedAboveDuplicates.TotalDuplicates -
		result.UnmanagedBelowDuplicates.TotalDuplicates

	totalDuplicates := result.ManagedDuplicates.TotalDuplicates +
		result.UnmanagedAboveDuplicates.TotalDuplicates +
		result.UnmanagedBelowDuplicates.TotalDuplicates

	logging.Debug.Log("Import completed",
		"managed_profiles", result.ManagedStats.ProfilesWritten,
		"above_profiles", result.UnmanagedAboveStats.ProfilesWritten,
		"below_profiles", result.UnmanagedBelowStats.ProfilesWritten,
		"duplicates_skipped", totalDuplicates,
		"total_written", totalWritten)

	// Display results (presentation only)
	return displayImportResult(result, flags)
}

// parseImportFlags parses command flags and arguments into an ImportFlags struct.
//
// This function extracts and validates all import command flags, with special
// handling for the backup path which can be provided as either a positional
// argument or a --backup flag.
//
// Default behavior is to import everything (full restore).
//
// Parameters:
//   - cmd: Cobra command with parsed flags
//   - args: Positional arguments (backup file path may be first arg)
//
// Returns:
//   - *ImportFlags: Parsed flags with validated values
//   - error: Any error encountered during parsing
func parseImportFlags(cmd *cobra.Command, args []string) (*ImportFlags, error) {
	// Backup path can come from positional arg or --backup flag
	backupPath, _ := cmd.Flags().GetString("backup")
	if backupPath == "" && len(args) > 0 {
		backupPath = args[0]
	}

	includeManaged, _ := cmd.Flags().GetBool("include-managed")
	includeAbove, _ := cmd.Flags().GetBool("include-above")
	includeBelow, _ := cmd.Flags().GetBool("include-below")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	ignoreSettings, _ := cmd.Flags().GetBool("ignore-settings")
	backupCurrentSettings, _ := cmd.Flags().GetBool("backup-current-settings")
	generateCheatSheet, _ := cmd.Flags().GetBool("cheat-sheet")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Default: import everything (full restore)
	if !includeManaged && !includeAbove && !includeBelow {
		includeManaged = true
		includeAbove = true
		includeBelow = true
		logging.Debug.Log("No sections specified, defaulting to full restore")
	}

	return &ImportFlags{
		BackupPath:            backupPath,
		IncludeManaged:        includeManaged,
		IncludeAbove:          includeAbove,
		IncludeBelow:          includeBelow,
		DryRun:                dryRun,
		IgnoreSettings:        ignoreSettings,
		BackupCurrentSettings: backupCurrentSettings,
		GenerateCheatSheet:    generateCheatSheet,
		Verbose:               verbose,
	}, nil
}

// displayImportResult formats and displays the import operation results.
//
// This function presents a user-friendly summary of the import operation,
// including profile counts by section, duplicate detection, settings restoration,
// and special messaging for dry-run mode.
//
// Display Logic:
//   - Dry-run: Shows "no changes made" message
//   - Import mode: Indicates managed, personal, or full backup
//   - Profile counts: Shows managed, above, below, and duplicate counts
//   - Settings: Reports settings restoration or backup status
//   - Verbose mode: Shows detailed breakdown by section
//
// Parameters:
//   - result: Import operation results from backup package
//   - flags: Command flags for conditional display (dry-run, verbose, etc.)
//
// Returns:
//   - error: Any error encountered during display (currently always nil)
func displayImportResult(result *backup.ImportResult, flags *ImportFlags) error {
	if flags.DryRun {
		logging.Log.Success("✓ Dry-run completed (no changes made)")
	} else {
		logging.Log.Success("✓ Import completed successfully")
	}

	// Show what was imported
	if flags.IncludeManaged && !flags.IncludeAbove && !flags.IncludeBelow {
		logging.Log.Info("Import mode: Managed profiles only")
	} else if !flags.IncludeManaged && (flags.IncludeAbove || flags.IncludeBelow) {
		logging.Log.Info("Import mode: Personal profiles only")
	} else {
		logging.Log.Info("Import mode: Full backup (managed + personal profiles)")
	}

	// Calculate actual profiles written (Raw - Duplicates)
	managedWritten := result.ManagedStats.ProfilesWritten - result.ManagedDuplicates.TotalDuplicates
	aboveWritten := result.UnmanagedAboveStats.ProfilesWritten - result.UnmanagedAboveDuplicates.TotalDuplicates
	belowWritten := result.UnmanagedBelowStats.ProfilesWritten - result.UnmanagedBelowDuplicates.TotalDuplicates
	totalWritten := managedWritten + aboveWritten + belowWritten
	totalInBackup := result.ManagedStats.ProfilesWritten + result.UnmanagedAboveStats.ProfilesWritten + result.UnmanagedBelowStats.ProfilesWritten
	totalDuplicates := result.ManagedDuplicates.TotalDuplicates + result.UnmanagedAboveDuplicates.TotalDuplicates + result.UnmanagedBelowDuplicates.TotalDuplicates

	// Show counts by section
	if flags.Verbose || flags.DryRun {
		if managedWritten > 0 {
			logging.Log.Infof("Managed profiles imported: %d", managedWritten)
		}
		if aboveWritten > 0 {
			logging.Log.Infof("Personal profiles above imported: %d", aboveWritten)
		}
		if belowWritten > 0 {
			logging.Log.Infof("Personal profiles below imported: %d", belowWritten)
		}
		if totalDuplicates > 0 {
			logging.Log.Infof("Duplicate profiles skipped: %d (%d SSO, %d IAM, %d Generic)",
				totalDuplicates,
				result.UnmanagedAboveDuplicates.SsoProfiles+result.UnmanagedBelowDuplicates.SsoProfiles,
				result.UnmanagedAboveDuplicates.IamProfiles+result.UnmanagedBelowDuplicates.IamProfiles,
				result.UnmanagedAboveDuplicates.GenericProfiles+result.UnmanagedBelowDuplicates.GenericProfiles)
		}
	}

	logging.Log.Info("Total profiles written",
		"count", totalWritten,
		"in_backup", totalInBackup,
		"config", result.ConfigPath,
	)

	// Show settings restoration status
	if result.SettingsRestored {
		logging.Log.Info("✓ Application settings restored from backup")
		if result.SettingsBackupPath != "" {
			logging.Log.Info("Previous settings backed up",
				"path", result.SettingsBackupPath,
			)
		}
	} else if result.BackupFile != nil && result.BackupFile.Settings != nil && flags.IgnoreSettings {
		logging.Log.Info("⊘ Settings found in backup but ignored per user request")
	} else {
		logging.Log.Info("⊘ No settings found in backup")
	}

	// Show cheat sheet generation status
	if result.CheatSheetGenerated {
		logging.Log.Info("✓ Cheat sheet generated",
			"path", result.CheatSheetPath,
		)
	}

	if !flags.DryRun {
		logging.Log.Info("📝 Profiles have been written to your AWS config")
	}

	return nil
}
