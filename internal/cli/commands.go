// Package cli provides the command-line interface for AWS Profile Manager.
//
// This package implements the CLI presentation layer using Cobra for command parsing
// and flag handling. It follows the architecture pattern where CLI is a thin layer
// that calls package API functions and displays results.
//
// Architecture Pattern (CRITICAL):
//
//	CLI Layer Responsibility:
//	  1. Parse command flags and arguments
//	  2. Call ONE package API function with parsed data
//	  3. Display results to user
//
//	CLI Layer MUST NOT:
//	  - Create internal objects (extractors, filters, managers)
//	  - Orchestrate multiple package calls
//	  - Contain business logic
//
// Available Commands:
//   - install: Install AWS profiles from configuration
//   - profiles: List and filter existing AWS profiles
//   - sessions: Check SSO session status
//   - export: Export AWS profiles to JSON backup
//   - import: Import AWS profiles from JSON backup
//   - sync: Fetch remote configuration files
//   - gui: Launch graphical user interface
//   - version: Display version information
//
// Command Structure:
//
//	Each command has its own file (install.go, profiles.go, etc.) containing:
//	- createXCommand() - Command definition and flags
//	- runXCommand() - Command execution logic
//	- Helper functions for flag parsing and display
//
// Example Command Pattern:
//
//	func runCommand(cmd *cobra.Command, args []string) error {
//	    // Step 1: Parse flags (CLI responsibility)
//	    criteria := parseFlags(cmd)
//
//	    // Step 2: Call ONE package function
//	    result, err := packageAPI.DoOperation(criteria)
//	    if err != nil {
//	        return err
//	    }
//
//	    // Step 3: Display results (CLI responsibility)
//	    displayResults(result)
//	    return nil
//	}
package cli

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// RegisterCommands registers all application commands with the root command.
//
// This function is called during application initialization to set up the complete
// command tree. It registers all available commands and their subcommands.
//
// Command Registration Order:
//  1. version - Version information
//  2. sync - Configuration synchronization
//  3. install - Profile installation
//  4. profiles - Profile listing
//  5. sessions - Session status
//  6. export - Profile export
//  7. import - Profile import
//  8. gui - GUI launcher
//
// Parameters:
//   - rootCmd: Root Cobra command to attach subcommands to
func RegisterCommands(rootCmd *cobra.Command) {
	logging.Debug.Log("Starting registration of CLI commands")

	// Register version command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "version")
	versionCmd := createVersionCommand()
	rootCmd.AddCommand(versionCmd)

	// Register the sync command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "sync")
	syncCmd := createSyncCommand()
	rootCmd.AddCommand(syncCmd)

	// Register the install command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "install")
	installCmd := createInstallCommand()
	rootCmd.AddCommand(installCmd)

	// Register the profiles command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "profiles")
	profilesCmd := createProfilesCommand()
	rootCmd.AddCommand(profilesCmd)

	// Register the sessions command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "sessions")
	sessionsCmd := createSessionsCommand()
	rootCmd.AddCommand(sessionsCmd)

	// Register the export command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "export")
	exportCmd := createExportCommand()
	rootCmd.AddCommand(exportCmd)

	// Register the import command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "import")
	importCmd := createImportCommand()
	rootCmd.AddCommand(importCmd)

	// Register gui command
	logging.Debug.Logf("\t🔹 Registering cli command: %s", "gui")
	guiCmd := createGUICommand()
	rootCmd.AddCommand(guiCmd)

	logging.Debug.Log("All CLI commands registered successfully")
}

// createInstallCommand creates the install subcommand with all its flags.
//
// The install command generates AWS CLI configuration entries from a centralized
// configuration file, with support for filtering and profile removal.
//
// Returns:
//   - *cobra.Command: Configured install command ready to use
func createInstallCommand() *cobra.Command {
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install AWS CLI profiles from configuration",
		Long: `Install AWS CLI profiles based on the configured organizations, accounts, and roles.
This command generates AWS CLI configuration entries in your ~/.aws/config file
(or specified config file) with proper SSO session configurations.

The command supports filtering to install only specific profiles based on
organizations, partitions, accounts, roles, or regions.

Use --remove flag to remove previously installed profiles and cheat sheets.

Examples:
  # Install all profiles
  aws-profile-manager install

  # Install profiles for specific organizations
  aws-profile-manager install --organizations=org1,org2

  # Install profiles for development accounts only
  aws-profile-manager install --accounts=dev,staging

  # Install profiles for specific roles
  aws-profile-manager install --roles=Developer,ReadOnly

  # Install profiles with all regions (not just default)
  aws-profile-manager install --all-regions

  # Combine multiple filters
  aws-profile-manager install --organizations=org1 --partitions=commercial --roles=Developer

  # Remove installed profiles and cheat sheet
  aws-profile-manager install --remove`,
		RunE: runInstall,
	}

	// Add install-specific flags
	installCmd.Flags().StringSlice("organizations", nil, "Filter by organization names (comma-separated)")
	installCmd.Flags().StringSlice("partitions", nil, "Filter by partition names: commercial,govcloud (comma-separated)")
	installCmd.Flags().StringSlice("accounts", nil, "Filter by account aliases (comma-separated)")
	installCmd.Flags().StringSlice("roles", nil, "Filter by role names (comma-separated)")
	installCmd.Flags().StringSlice("regions", nil, "Filter by specific regions (comma-separated)")
	installCmd.Flags().Bool("all-regions", false, "Include all available regions for each account (overrides --regions)")
	installCmd.Flags().String("output", "", "AWS config file path (defaults to ~/.aws/config)")
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without making changes")
	installCmd.Flags().Bool("verbose", false, "Enable verbose output")
	installCmd.Flags().Bool("remove", false, "Remove installed profiles and cheat sheet")

	// Make --cheat-sheet accept an optional value.
	// If used without a value, we set it to the DEFAULT path via NoOptDefVal.
	// If omitted entirely, we leave it empty and won't generate a cheat sheet.
	installCmd.Flags().String("cheat-sheet", "", "Generate a Markdown cheat sheet (optionally specify path)")
	// (optional) filename completion hint
	_ = installCmd.Flags().SetAnnotation("cheat-sheet", cobra.BashCompFilenameExt, []string{"md"})

	// We need the default cheat sheet path now so we can wire NoOptDefVal.
	// This does NOT force cheat sheet generation; it only applies when the user
	// actually passes --cheat-sheet with no value.
	desktopDir := settings.GetDesktopDir()
	defaultCheatSheetPath := filepath.Join(desktopDir, "AWS_Profile_Cheat_Sheet.md")
	installCmd.Flags().Lookup("cheat-sheet").NoOptDefVal = defaultCheatSheetPath

	installCmd.Flags().Bool("cheat-sheet-only", false, "Skip updating the AWS config file; only generate cheat sheet")

	return installCmd
}

// createSessionsCommand creates the sessions subcommand to check SSO session status.
//
// The sessions command displays the status of AWS SSO sessions, showing which
// sessions have valid tokens and which have expired.
//
// Returns:
//   - *cobra.Command: Configured sessions command ready to use
func createSessionsCommand() *cobra.Command {
	sessionsCmd := &cobra.Command{
		Use:   "sessions",
		Short: "List AWS SSO sessions and their status",
		Long: `List all AWS SSO sessions configured in your AWS CLI configuration along with their current status.

This command shows:
- Active sessions (not expired)
- Expired sessions that need re-authentication
- Session details (start URL, region, expiration)

The status is determined by checking the AWS CLI SSO cache files.

Examples:
  # List all sessions with status
  aws-profile-manager sessions

  # List sessions with detailed output
  aws-profile-manager sessions --verbose`,
		RunE: runSessions,
	}

	sessionsCmd.Flags().Bool("verbose", false, "Show detailed session information")
	sessionsCmd.Flags().Bool("refresh", false, "Force refresh of session status")

	return sessionsCmd
}

// createProfilesCommand creates the profiles subcommand to list and filter AWS CLI profiles.
//
// The profiles command lists existing AWS CLI profiles with support for filtering
// by account, role, region, and other criteria.
//
// Returns:
//   - *cobra.Command: Configured profiles command ready to use
func createProfilesCommand() *cobra.Command {
	profilesCmd := &cobra.Command{
		Use:   "profiles",
		Short: "List AWS CLI profiles from your configuration",
		Long: `List all AWS CLI profiles found in your AWS configuration file (~/.aws/config).

This command shows:
- Profile names and their associated accounts
- SSO configuration details
- Regional configuration
- Role information

Profiles can be filtered by various criteria.

Examples:
  # List all profiles
  aws-profile-manager profiles

  # List profiles with detailed output
  aws-profile-manager profiles --verbose

  # Filter profiles by account
  aws-profile-manager profiles --account-id 123456789012

  # Filter profiles by role
  aws-profile-manager profiles --role Developer`,
		RunE: runProfiles,
	}

	profilesCmd.Flags().Bool("verbose", false, "Show detailed profile information")
	profilesCmd.Flags().String("account-id", "", "Filter profiles by account ID")
	profilesCmd.Flags().String("role", "", "Filter profiles by role name")
	profilesCmd.Flags().String("region", "", "Filter profiles by region")
	profilesCmd.Flags().String("session", "", "Filter profiles by SSO session name")
	profilesCmd.Flags().String("pattern", "", "Filter profiles by name pattern (regex)")

	return profilesCmd
}

// createExportCommand creates the export subcommand for backing up AWS profiles.
//
// The export command saves AWS CLI profiles to a JSON backup file, supporting
// different export modes (managed-only, full backup, personal-only).
//
// Returns:
//   - *cobra.Command: Configured export command ready to use
func createExportCommand() *cobra.Command {
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export AWS profiles to backup JSON format",
		Long: `Export AWS profiles from ~/.aws/config to backup JSON format.

This command is useful for:
1. Company admins creating installer configs (managed profiles only)
2. Personal backup before OS reinstall (full backup with all sections)
3. Disaster recovery and configuration migration

The exported JSON file can be:
- Used as an installer config (aws-profile-manager install --config exported.json)
- Imported to restore profiles (aws-profile-manager import exported.json)
- Stored in Git, S3, or distributed via sync system

By default, exports everything (full backup). Use flags for granular control.

Examples:
  # Full backup (default: all sections)
  aws-profile-manager export --output backup.json

  # Managed profiles only (for installer config)
  aws-profile-manager export --include-managed --output installer.json

  # Personal profiles only
  aws-profile-manager export --include-above --include-below --output personal.json

  # With metadata description
  aws-profile-manager export --output backup.json --description "Pre-OS-wipe backup"

  # Without application settings
  aws-profile-manager export --output backup.json --exclude-settings`,
		RunE: runExport,
	}

	exportCmd.Flags().StringP("output", "o", "", "Output JSON file path (required)")
	exportCmd.Flags().Bool("include-managed", false, "Include managed profiles (between markers)")
	exportCmd.Flags().Bool("include-above", false, "Include personal profiles above managed section")
	exportCmd.Flags().Bool("include-below", false, "Include personal profiles below managed section")
	exportCmd.Flags().String("description", "", "Add description metadata to export")
	exportCmd.Flags().Bool("exclude-settings", false, "Exclude application settings from backup")
	exportCmd.Flags().Bool("verbose", false, "Show detailed export information")

	_ = exportCmd.MarkFlagRequired("output")

	return exportCmd
}

// createImportCommand creates the import subcommand for restoring AWS profiles.
//
// The import command restores AWS CLI profiles from a JSON backup file, with
// support for different import modes and duplicate detection.
//
// Returns:
//   - *cobra.Command: Configured import command ready to use
func createImportCommand() *cobra.Command {
	importCmd := &cobra.Command{
		Use:   "import [backup-file]",
		Short: "Import AWS profiles from backup JSON file",
		Long: `Import AWS profiles from a backup JSON file to ~/.aws/config.

This command restores profiles from a backup created with the export command.
It intelligently merges profiles, detects duplicates, and can restore settings.

The backup file can be specified as:
- Positional argument: aws-profile-manager import backup.json
- Flag: aws-profile-manager import --backup backup.json

By default, imports everything (full restore). Use flags for granular control.

Examples:
  # Import full backup (default: all sections)
  aws-profile-manager import backup.json

  # Import managed profiles only
  aws-profile-manager import --include-managed backup.json

  # Import personal profiles only
  aws-profile-manager import --include-above --include-below personal.json

  # Preview import without making changes
  aws-profile-manager import --dry-run backup.json

  # Import without restoring settings
  aws-profile-manager import backup.json --ignore-settings

  # Backup current settings before restoring
  aws-profile-manager import backup.json --backup-current-settings`,
		RunE: runImport,
		Args: cobra.MaximumNArgs(1),
	}

	importCmd.Flags().String("backup", "", "Backup JSON file to import (can also use positional arg)")
	importCmd.Flags().Bool("include-managed", false, "Include managed profiles (between markers)")
	importCmd.Flags().Bool("include-above", false, "Include personal profiles above managed section")
	importCmd.Flags().Bool("include-below", false, "Include personal profiles below managed section")
	importCmd.Flags().Bool("dry-run", false, "Preview import without making changes")
	importCmd.Flags().Bool("ignore-settings", false, "Don't restore application settings from backup")
	importCmd.Flags().Bool("backup-current-settings", true, "Backup current settings before restoring")
	importCmd.Flags().Bool("cheat-sheet", false, "Generate cheat sheet after import")
	importCmd.Flags().Bool("verbose", false, "Show detailed import information")

	return importCmd
}

// createGUICommand creates the GUI launcher subcommand.
//
// The gui command launches the graphical user interface, providing a window-based
// alternative to the command-line interface.
//
// Returns:
//   - *cobra.Command: Configured GUI command ready to use
func createGUICommand() *cobra.Command {
	guiCmd := &cobra.Command{
		Use:   "gui",
		Short: "Launch the GUI interface",
		Long: `Launch the graphical user interface for AWS Profile Manager.
This provides a user-friendly window-based interface for managing
AWS profiles and credentials.`,
		Run: runGUI,
	}

	return guiCmd
}

// createVersionCommand creates the version information subcommand.
//
// The version command displays application version, build information, and
// copyright details.
//
// Returns:
//   - *cobra.Command: Configured version command ready to use
func createVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long: `Display detailed version information including build details,
Git commit hash, build date, and Go version used to compile the application.`,
		Run: runVersion,
	}

	return versionCmd
}
