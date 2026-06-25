package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/security"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/sync"
	"aws-profile-manager/internal/task"
)

// runInstall executes the install command to write AWS CLI profiles to config.
//
// This command handler installs AWS CLI profiles from a configuration schema,
// with support for filtering by organization, partition, account, role, and region.
//
// Command Flow:
//  1. Parse flags into InstallFlags struct
//  2. Handle remove mode if --remove flag is set
//  3. Load schema (from --config file or via sync system)
//  4. Build InstallOptions from flags
//  5. Call profiles.Installer API (ONE call)
//  6. Display results
//
// Configuration Loading Priority:
//   - Explicit --config file: Use file directly
//   - No --config: Use sync system (remote → cache → bundled fallback)
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (currently unused)
//
// Returns:
//   - error: Any error encountered during execution
func runInstall(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Install command started")

	// Parse flags into options
	logging.Debug.Log("Parsing command flags")
	flags, err := parseInstallFlags(cmd)
	if err != nil {
		return logging.Log.Errorf("failed to parse install flags: %v", err)
	}

	// Enable verbose logging if requested
	if flags.Verbose {
		logging.Log.Info("Verbose mode enabled")
	}

	// Handle remove mode
	if flags.Remove {
		return runRemove(cmd, flags)
	}

	// Get config file from global flag
	configFile, _ := cmd.Root().PersistentFlags().GetString("config")
	forceRefresh, _ := cmd.Root().PersistentFlags().GetBool("force")

	// Load schema (from explicit file or via sync)
	logging.Debug.Log("Loading configuration schema")
	var s *schema.Schema

	if configFile != "" {
		// Load from explicit file (--config flag)
		logging.Debug.Log("Loading from explicit config file", "path", configFile)
		data, err := security.ReadFile(configFile, security.ReadOptions{
			AllowedExtensions: []string{".json"},
		})
		if err != nil {
			return logging.Log.ErrorfWithDetails("failed to read config file", err)
		}

		s, err = schema.SchemaFromJSON(data)
		if err != nil {
			return logging.Log.ErrorfWithDetails("failed to parse config file", err)
		}
	} else {
		// Load via sync (handles sync → cache)
		logging.Debug.Log("Loading via sync", "force", forceRefresh)

		// Get sync settings and build config
		syncSettings := settings.Get().Sync
		if !syncSettings.Enabled {
			return logging.Log.Errorf("sync is not enabled - please use --config flag or enable sync in settings")
		}

		cfg := sync.ConfigFromSettings(&syncSettings)
		opts := sync.Options{
			ForceRefresh: forceRefresh,
		}

		// Use CLI reporter for progress
		reporter := task.CliReporter{}
		ctx := context.Background()

		result, err := sync.Sync(ctx, cfg, opts, reporter)
		if err != nil {
			return logging.Log.ErrorfWithDetails("failed to sync configuration", err)
		}

		s = result.Data
	}

	// Build install options from flags
	logging.Debug.Log("Building install options")

	// Build config from settings
	awsDir := settings.GetAwsDir()
	desktopDir := settings.GetDesktopDir()
	appSettings := settings.Get().Application

	config := profiles.Config{
		ConfigPath:          filepath.Join(awsDir, "config"),
		CheatSheetOutputDir: desktopDir,
		AwsDir:              awsDir,
		StartMarker:         appSettings.GetFormattedStartMarker(),
		EndMarker:           appSettings.GetFormattedEndMarker(),
		IncludeTimestamp:    appSettings.IncludeTimestamp,
		IncludeVersion:      appSettings.IncludeVersion,
	}

	// Override config path if specified
	if flags.Output != "" {
		config.ConfigPath = flags.Output
	}

	// cheat-sheet-only implies cheat sheet generation.
	generateCheatSheet := flags.CheatSheet != "" || flags.CheatSheetOnly
	cheatSheetPath := flags.CheatSheet
	if flags.CheatSheetOnly && cheatSheetPath == "" {
		cheatSheetPath = filepath.Join(desktopDir, "AWS_Profile_Cheat_Sheet.md")
	}

	options := profiles.InstallOptions{
		Schema:             s, // Pass loaded schema
		Organizations:      flags.Organizations,
		Partitions:         flags.Partitions,
		Accounts:           flags.Accounts,
		Roles:              flags.Roles,
		Regions:            flags.Regions,
		AllRegions:         flags.AllRegions,
		GenerateCheatSheet: generateCheatSheet,
		CheatSheetPath:     cheatSheetPath,
		CheatSheetOnly:     flags.CheatSheetOnly,
		DryRun:             flags.DryRun,
	}

	// Call API to install profiles (business logic happens here)
	logging.Debug.Log("Calling profiles.Install API")
	reporter := task.CliReporter{}
	ctx := context.Background()

	installer := profiles.NewInstaller(config)
	result, err := installer.Install(ctx, options, reporter)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to install profiles", err)
	}

	logging.Debug.Log("Installation completed",
		"total_sessions", result.SsoSessions,
		"total_profiles", result.TotalProfiles)

	// Display results (presentation only)
	return displayInstallResult(result, s, flags)
}

// displayInstallResult formats and displays the installation operation results.
//
// This function presents a user-friendly summary of the installation operation,
// including profile counts, session information, cheat sheet generation status,
// and special messaging for dry-run and cheat-sheet-only modes.
//
// Display Logic:
//   - Configuration info: Shows schema version and organization count
//   - Filter warnings: Alerts if no profiles match filters
//   - Profile counts: Shows sessions, profiles, and cheat sheet status
//   - Dry-run: Shows "no changes made" message
//   - Cheat-sheet-only: Shows "no profiles installed" message
//   - Verbose mode: Shows detailed breakdown
//
// Parameters:
//   - result: Installation operation results from installer package
//   - flags: Command flags for conditional display (dry-run, verbose, etc.)
//
// Returns:
//   - error: Any error encountered during display (currently always nil)
func displayInstallResult(result *profiles.InstallResult, loadedSchema *schema.Schema, flags *InstallFlags) error {
	// Show configuration info
	if loadedSchema != nil {
		orgCount := 0
		if loadedSchema.Managed != nil && loadedSchema.Managed.Organizations != nil {
			orgCount = len(loadedSchema.Managed.Organizations)
		}
		logging.Log.Infof("AWS configuration loaded (version: %s, organizations: %d)",
			loadedSchema.Version, orgCount)
	}

	// Check if anything matches filters (skip this check in cheat-sheet-only mode)
	if !flags.CheatSheetOnly && result.SsoSessions == 0 && result.TotalProfiles == 0 {
		logging.Log.Warn("No profiles match the specified filters")
		logging.Log.Info("Try adjusting your filter criteria or run without filters to see all available profiles")
		return nil
	}

	// Show profile counts (unless cheat-sheet-only)
	if !flags.CheatSheetOnly {
		logging.Log.Infof("Profiles to be generated (SSO sessions: %d, profiles: %d)",
			result.SsoSessions, result.TotalProfiles)
	}

	// Show what would be generated in dry-run mode
	if flags.DryRun {
		logging.Log.Infof("Would write %d SSO sessions and %d Profiles to: %s",
			result.SsoSessions, result.TotalProfiles, result.ConfigPath)

		if flags.CheatSheet != "" || flags.CheatSheetOnly {
			logging.Log.Info("Would generate cheat sheet")
		}
		logging.Log.Warn("DRY RUN: No files were modified")
		return nil
	}

	// Display profile installation results (unless cheat-sheet-only)
	if !flags.CheatSheetOnly {
		logging.Log.Success("AWS CLI profiles installed successfully")
		logging.Log.InfoWithDetails("Installation completed",
			fmt.Sprintf("SSO sessions: %d, Profiles: %d, Config file: %s",
				result.SsoSessions,
				result.TotalProfiles,
				result.ConfigPath))
	}

	// Display cheat sheet results if generated
	if result.CheatSheetPath != "" {
		logging.Log.Success("Cheat sheet generated successfully")
		logging.Log.InfoWithDetails("Cheat sheet details", fmt.Sprintf("SSO sessions: %d, Profiles: %d, Output file: %s",
			result.SsoSessions,
			result.TotalProfiles,
			result.CheatSheetPath))
	}

	// Display verbose information
	if flags.Verbose {
		logging.Log.Info("Installation details:")

		if !flags.CheatSheetOnly {
			logging.Log.Infof("  Config path: %s", result.ConfigPath)
			logging.Log.Infof("  Sessions written: %d", result.SsoSessions)
			logging.Log.Infof("  Profiles written: %d", result.TotalProfiles)
		}

		if result.CheatSheetPath != "" {
			logging.Log.Infof("  Cheat sheet path: %s", result.CheatSheetPath)
			if fileInfo, err := os.Stat(result.CheatSheetPath); err == nil {
				logging.Log.Infof("  Cheat sheet size: %d bytes", fileInfo.Size())
			}
		}
	}

	return nil
}

// runRemove executes the remove command to clean up installed profiles and cheat sheet.
//
// This command handler removes profiles from the managed section and optionally
// deletes the cheat sheet file. It's called when --remove flag is set.
//
// Command Flow:
//  1. Build RemoveOptions from flags
//  2. Call profiles.Remover API (ONE call)
//  3. Display results
//
// Parameters:
//   - cmd: Cobra command context (unused but required by signature)
//   - flags: Command flags for remove operation
//
// Returns:
//   - error: Any error encountered during execution
func runRemove(_ *cobra.Command, flags *InstallFlags) error {
	logging.Debug.Log("Remove mode started")

	// Build config from settings
	awsDir := settings.GetAwsDir()
	desktopDir := settings.GetDesktopDir()
	appSettings := settings.Get().Application

	cheatSheetPath := filepath.Join(desktopDir, "AWS_Profile_Cheat_Sheet.md")
	config := profiles.Config{
		ConfigPath:          filepath.Join(awsDir, "config"),
		CheatSheetOutputDir: desktopDir,
		AwsDir:              awsDir,
		StartMarker:         appSettings.GetFormattedStartMarker(),
		EndMarker:           appSettings.GetFormattedEndMarker(),
		IncludeTimestamp:    appSettings.IncludeTimestamp,
		IncludeVersion:      appSettings.IncludeVersion,
	}

	// Override config path if specified
	if flags.Output != "" {
		config.ConfigPath = flags.Output
	}

	// Build remove options from flags
	logging.Debug.Log("\t🔹 Building remove options")
	options := profiles.RemoveOptions{
		RemoveCheatSheet: true,
		CheatSheetPath:   cheatSheetPath,
		DryRun:           flags.DryRun,
	}

	// Call API to remove profiles (business logic happens here)
	logging.Debug.Log("\t🔹 Calling profiles.Remove API")
	reporter := task.CliReporter{}
	ctx := context.Background()

	remover := profiles.NewRemover(config)
	result, err := remover.Remove(ctx, options, reporter)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to remove profiles", err)
	}

	logging.Debug.Log("\t🔹 Removal completed",
		"profiles_removed", result.ProfilesRemoved,
		"cheat_sheet_removed", result.RemovedCheatSheet)

	// Display results (presentation only)
	return displayRemoveResult(result, flags)
}

// displayRemoveResult formats and displays the removal operation results.
//
// This function presents a user-friendly summary of the removal operation,
// showing what was removed (profiles and cheat sheet) with special messaging
// for dry-run mode.
//
// Display Logic:
//   - Dry-run: Shows what would be removed without making changes
//   - Profile removal: Shows count and config file path
//   - Cheat sheet: Shows removal status if it was present
//   - Verbose mode: Shows detailed paths and counts
//
// Parameters:
//   - result: Removal operation results from profiles package
//   - flags: Command flags for conditional display (dry-run, verbose, etc.)
//
// Returns:
//   - error: Any error encountered during display (currently always nil)
func displayRemoveResult(result *profiles.RemoveResult, flags *InstallFlags) error {
	// Dry-run mode - show what would be removed
	if flags.DryRun {
		logging.Log.Info("🔍 Dry run mode - showing what would be removed:")

		if result.RemovedConfig {
			logging.Log.Info("  ✓ Would remove AWS CLI profiles from:", result.ConfigPath)
		} else {
			logging.Log.Info("  ℹ No managed profiles found in:", result.ConfigPath)
		}

		if result.RemovedCheatSheet {
			logging.Log.Info("  ✓ Would remove cheat sheet:", result.CheatSheetPath)
		} else {
			logging.Log.Info("  ℹ No cheat sheet found at:", result.CheatSheetPath)
		}

		return nil
	}

	// Actual removal results
	if result.ProfilesRemoved > 0 {
		logging.Log.Success("AWS CLI profiles removed successfully")
		logging.Log.Info("Removal completed",
			"config file", result.ConfigPath,
		)
	} else {
		logging.Log.Info("No managed profiles found to remove")
	}

	if result.RemovedCheatSheet {
		logging.Log.Success("Cheat sheet removed successfully")
		logging.Log.Info("Removal completed",
			"file", result.CheatSheetPath,
		)
	} else {
		logging.Log.Info("Cheat sheet file not found")
	}

	// Summary
	if result.ProfilesRemoved == 0 && !result.RemovedCheatSheet {
		logging.Log.Warn("Nothing was removed - no managed profiles or cheat sheet found")
	}

	return nil
}

// InstallFlags holds the parsed install command flags
type InstallFlags struct {
	Organizations  []string
	Partitions     []string
	Accounts       []string
	Roles          []string
	Regions        []string
	AllRegions     bool
	Output         string
	DryRun         bool
	Verbose        bool
	CheatSheet     string // Path for cheat sheet (empty = disabled, populated path = enabled)
	CheatSheetOnly bool   // Only generate cheat sheet, skip profile generation
	Remove         bool   // Remove installed profiles and cheat sheet
}

// parseInstallFlags extracts and validates install command flags.
//
// This function parses all install command flags from the Cobra command,
// providing error handling for each flag and handling special cases like
// cheat sheet path determination.
//
// Flag Validation:
//   - All slice flags: Parsed with error checking
//   - Boolean flags: Direct extraction
//   - CheatSheet: Special handling via parseCheatSheetFlag helper
//
// Parameters:
//   - cmd: Cobra command with parsed flags
//
// Returns:
//   - *InstallFlags: Parsed flags with validated values
//   - error: Any error encountered during flag parsing
func parseInstallFlags(cmd *cobra.Command) (*InstallFlags, error) {
	flags := &InstallFlags{}

	var err error
	flags.Organizations, err = cmd.Flags().GetStringSlice("organizations")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse organizations flag: %v", err)
	}

	flags.Partitions, err = cmd.Flags().GetStringSlice("partitions")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse partitions flag: %v", err)
	}

	flags.Accounts, err = cmd.Flags().GetStringSlice("accounts")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse accounts flag: %v", err)
	}

	flags.Roles, err = cmd.Flags().GetStringSlice("roles")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse roles flag: %v", err)
	}

	flags.Regions, err = cmd.Flags().GetStringSlice("regions")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse regions flag: %v", err)
	}

	flags.AllRegions, err = cmd.Flags().GetBool("all-regions")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse all-regions flag: %v", err)
	}

	flags.Output, err = cmd.Flags().GetString("output")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse output flag: %v", err)
	}

	flags.DryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse dry-run flag: %v", err)
	}

	flags.Verbose, err = cmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse verbose flag: %v", err)
	}

	flags.CheatSheetOnly, err = cmd.Flags().GetBool("cheat-sheet-only")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse cheat-sheet-only flag: %v", err)
	}

	flags.Remove, err = cmd.Flags().GetBool("remove")
	if err != nil {
		return nil, logging.Log.Errorf("failed to parse remove flag: %v", err)
	}

	// Parse cheat sheet flag with sophisticated logic for optional values
	// Debug flag info first
	if cheatSheetFlag := cmd.Flags().Lookup("cheat-sheet"); cheatSheetFlag != nil {
		logging.Debug.Logf("\t🔹 Flag changed: %t", cmd.Flags().Changed("cheat-sheet"))
		logging.Debug.Log("\t🔹 Flag Details",
			"value", cheatSheetFlag.Value.String(),
			"NoOptDefVal", cheatSheetFlag.NoOptDefVal,
		)
	}
	flags.CheatSheet = parseCheatSheetFlag(cmd, cmd.Flags().Args())
	logging.Debug.Logf("Parsed CheatSheet flag value: '%s'", flags.CheatSheet)

	// Validate partition values
	if len(flags.Partitions) > 0 {
		for _, partition := range flags.Partitions {
			if partition != "commercial" && partition != "govcloud" {
				return nil, logging.Log.Errorf("invalid partition '%s': must be 'commercial' or 'govcloud'", partition)
			}
		}
	}

	return flags, nil
}

// parseCheatSheetFlag handles the --cheat-sheet flag parsing with optional values.
//
// This function handles the complex --cheat-sheet flag which supports two usage modes:
//   - --cheat-sheet: Uses default path (Desktop directory)
//   - --cheat-sheet=<path>: Uses custom path
//
// Implementation Note:
//
//	Due to Cobra limitations, the space syntax (--cheat-sheet <path>) is not
//	supported for custom paths. Users must use = syntax for custom paths.
//
// Parameters:
//   - cmd: Cobra command with parsed flags
//   - _: Command arguments (unused but required by signature)
//
// Returns:
//   - string: Cheat sheet path (empty if flag not set, default path if flag set without value, custom path if provided)
func parseCheatSheetFlag(cmd *cobra.Command, _ []string) string {
	// Get the flag value from Cobra
	cheatSheet, _ := cmd.Flags().GetString("cheat-sheet")

	// If flag was set, return its value
	// - If --cheat-sheet alone: NoOptDefVal provides default desktop path
	// - If --cheat-sheet=<path>: cheatSheet contains the custom path
	return cheatSheet
}
