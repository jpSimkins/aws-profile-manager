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

// ExportFlags holds all flags for the export command.
//
// This struct provides a typed container for export command flags, making
// flag parsing and validation cleaner.
type ExportFlags struct {
	OutputPath      string
	IncludeManaged  bool
	IncludeAbove    bool
	IncludeBelow    bool
	Description     string
	ExcludeSettings bool
	Verbose         bool
}

// runExport executes the export command to backup AWS CLI profiles.
//
// This command handler exports AWS CLI profiles to a JSON backup file,
// supporting different export modes (managed-only, full, personal-only).
//
// Command Flow:
//  1. Parse flags into ExportFlags struct
//  2. Build Config from settings (DI)
//  3. Build ExportOptions from flags
//  4. Call backup.ExportProfiles() with context and reporter (ONE call)
//  5. Display results
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during execution
func runExport(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Export command started")

	// Parse flags
	logging.Debug.Log("Parsing command flags")
	flags, err := parseExportFlags(cmd)
	if err != nil {
		return logging.Log.Errorf("failed to parse export flags: %v", err)
	}

	// Enable verbose logging if requested
	if flags.Verbose {
		logging.Log.Info("Verbose mode enabled")
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

	// Build export options from flags
	logging.Debug.Log("Building export options")
	opts := backup.ExportOptions{
		OutputPath:      flags.OutputPath,
		IncludeManaged:  flags.IncludeManaged,
		IncludeAbove:    flags.IncludeAbove,
		IncludeBelow:    flags.IncludeBelow,
		Description:     flags.Description,
		ExcludeSettings: flags.ExcludeSettings,
	}

	// Create reporter for progress updates
	reporter := task.CliReporter{}

	// Call API to export profiles (business logic happens here)
	logging.Debug.Log("Calling backup.ExportProfiles API")
	ctx := context.Background()
	result, err := backup.ExportProfiles(ctx, cfg, opts, reporter)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to export profiles", err)
	}

	logging.Debug.Log("Export completed",
		"managed", result.ManagedProfiles,
		"above", result.UnmanagedAbove,
		"below", result.UnmanagedBelow,
		"total", result.TotalProfiles)

	// Display results (presentation only)
	return displayExportResult(result, flags)
}

// parseExportFlags parses command flags into ExportFlags struct.
//
// This function extracts all export-related flags from the cobra command and
// validates them. Default behavior is to export everything (full backup).
//
// Parameters:
//   - cmd: Cobra command with parsed flags
//
// Returns:
//   - *ExportFlags: Parsed flag values
//   - error: Any error encountered during parsing
func parseExportFlags(cmd *cobra.Command) (*ExportFlags, error) {
	outputPath, _ := cmd.Flags().GetString("output")
	includeManaged, _ := cmd.Flags().GetBool("include-managed")
	includeAbove, _ := cmd.Flags().GetBool("include-above")
	includeBelow, _ := cmd.Flags().GetBool("include-below")
	description, _ := cmd.Flags().GetString("description")
	excludeSettings, _ := cmd.Flags().GetBool("exclude-settings")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Default: include everything (full backup)
	if !includeManaged && !includeAbove && !includeBelow {
		includeManaged = true
		includeAbove = true
		includeBelow = true
		logging.Debug.Log("No sections specified, defaulting to full backup")
	}

	return &ExportFlags{
		OutputPath:      outputPath,
		IncludeManaged:  includeManaged,
		IncludeAbove:    includeAbove,
		IncludeBelow:    includeBelow,
		Description:     description,
		ExcludeSettings: excludeSettings,
		Verbose:         verbose,
	}, nil
}

// displayExportResult formats and displays the export result to the user.
//
// This function handles all display logic for the export command, showing:
//   - Success message
//   - Export mode (managed, personal above/below, full)
//   - Profile counts by type
//   - Settings inclusion status
//   - Output file location
//
// Parameters:
//   - result: Export result with profile counts and metadata
//   - flags: Export flags containing display preferences
//
// Returns:
//   - error: Any error encountered during display
func displayExportResult(result *backup.ExportResult, flags *ExportFlags) error {
	logging.Log.Success("✓ Export completed successfully")

	// Show what was exported
	if flags.IncludeManaged && !flags.IncludeAbove && !flags.IncludeBelow {
		logging.Log.Info("Export mode: Managed profiles only (installer config)")
	} else if !flags.IncludeManaged && (flags.IncludeAbove || flags.IncludeBelow) {
		logging.Log.Info("Export mode: Personal profiles only")
	} else {
		logging.Log.Info("Export mode: Full backup (managed + personal profiles)")
	}

	// Show counts by section
	if flags.Verbose {
		if result.ManagedProfiles > 0 {
			logging.Log.Info("Managed profiles exported",
				"count", result.ManagedProfiles,
			)
		}
		if result.UnmanagedAbove > 0 {
			logging.Log.Info("Personal profiles above exported",
				"count", result.UnmanagedAbove,
			)
		}
		if result.UnmanagedBelow > 0 {
			logging.Log.Info("Personal profiles below exported",
				"count", result.UnmanagedBelow,
			)
		}
	}

	logging.Log.Info("Total profiles exported",
		"count", result.TotalProfiles,
		"output", result.OutputPath,
	)

	// Show settings status
	if result.SettingsExported {
		logging.Log.Info("✓ Application settings included in backup")
	} else {
		logging.Log.Info("⊘ Application settings excluded from backup")
	}

	// Show metadata if present and verbose
	if flags.Verbose && result.BackupFile != nil {
		metadata := result.BackupFile.Metadata
		logging.Log.Info("Backup metadata",
			"version", result.BackupFile.Version,
			"timestamp", metadata.ExportedAt.Format("2006-01-02 15:04:05"),
			"tool_version", metadata.ToolVersion,
		)
		if metadata.Description != "" {
			logging.Log.Info("Description", "text", metadata.Description)
		}
	}

	return nil
}
