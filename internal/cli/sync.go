package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/sync"
	"aws-profile-manager/internal/task"
)

// getSyncSettings extracts sync settings from the global application settings.
//
// This is a helper function used by sync command handlers to access the
// current sync configuration.
//
// Returns:
//   - *settings.SyncSettings: Current sync settings
//   - error: Always nil (kept for interface consistency)
func getSyncSettings() (*settings.SyncSettings, error) {
	currentSettings := settings.Get()
	return &currentSettings.Sync, nil
}

// createSyncCommand creates the sync command with all subcommands.
//
// The sync command manages remote AWS configuration synchronization, allowing
// users to fetch configurations from centralized sources like HTTP endpoints
// or S3 buckets.
//
// Subcommands:
//   - fetch: Download configuration from remote source
//   - status: Check sync status and cache information
//   - clear-cache: Remove cached configuration
//   - setup: Display setup instructions for new users
//
// Returns:
//   - *cobra.Command: Configured sync command
func createSyncCommand() *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Manage AWS configuration synchronization",
		Long: `Manage AWS configuration synchronization from remote sources.

Sync allows you to fetch AWS configuration from centralized sources:
  - HTTP/HTTPS endpoints (for public configs)
  - S3 buckets (authenticated with SSO or IAM)
  - Git repositories (SSH or HTTPS)
  - Local files (for testing)

Examples:
  # Fetch configuration from remote source
  aws-profile-manager sync fetch

  # Check sync status and cache info
  aws-profile-manager sync status

  # Clear cached configuration
  aws-profile-manager sync clear-cache

  # Setup sync for new hires (bootstrap instructions)
  aws-profile-manager sync setup
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help if no subcommand provided
			return cmd.Help()
		},
	}

	// Add subcommands
	syncCmd.AddCommand(createSyncFetchCommand())
	syncCmd.AddCommand(createSyncStatusCommand())
	syncCmd.AddCommand(createSyncClearCacheCommand())
	syncCmd.AddCommand(createSyncSetupCommand())

	return syncCmd
}

// createSyncFetchCommand creates the sync fetch subcommand.
//
// The fetch command downloads AWS configuration from the configured remote
// source (HTTP endpoint or S3 bucket) and caches it locally for offline use.
//
// Flags:
//   - --force, -f: Force fetch even if cache is recent
//   - --verbose, -v: Show detailed fetch information
//
// Returns:
//   - *cobra.Command: Configured sync fetch command
func createSyncFetchCommand() *cobra.Command {
	fetchCmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch configuration from remote source",
		Long: `Fetch AWS configuration from the configured remote source.

This command will:
  1. Connect to your configured sync source (HTTP or S3)
  2. Download the latest configuration
  3. Cache it locally for offline use
  4. Fall back to cached config if remote is unavailable

The fetched configuration can then be used with 'aws-profile-manager install'.
`,
		RunE: runSyncFetch,
	}

	fetchCmd.Flags().BoolP("force", "f", false, "Force fetch even if cache is recent")
	fetchCmd.Flags().BoolP("verbose", "v", false, "Show detailed fetch information")

	return fetchCmd
}

// createSyncStatusCommand creates the sync status subcommand.
//
// The status command displays current sync configuration and cache information,
// helping users verify their sync setup and check cache age.
//
// Flags:
//   - --verbose, -v: Show detailed status information
//
// Returns:
//   - *cobra.Command: Configured sync status command
func createSyncStatusCommand() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync configuration and cache status",
		Long: `Display the current sync configuration and cache information.

Shows:
  - Sync enabled/disabled status
  - Configured sync strategy (HTTP, S3, Git)
  - Source location
  - Cache status and age
  - Last successful sync
`,
		RunE: runSyncStatus,
	}

	statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed status information")

	return statusCmd
}

// createSyncClearCacheCommand creates the sync clear-cache subcommand.
//
// The clear-cache command removes the locally cached configuration, forcing
// the next fetch to download from the remote source. Useful for testing or
// troubleshooting sync issues.
//
// Flags:
//   - --yes, -y: Skip confirmation prompt
//
// Returns:
//   - *cobra.Command: Configured sync clear-cache command
func createSyncClearCacheCommand() *cobra.Command {
	clearCmd := &cobra.Command{
		Use:   "clear-cache",
		Short: "Clear cached configuration",
		Long: `Remove the locally cached configuration.

This forces the next fetch to download from the remote source.
Useful when:
  - Testing sync configuration changes
  - Troubleshooting sync issues
  - Forcing a fresh download
`,
		RunE: runSyncClearCache,
	}

	clearCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	return clearCmd
}

// createSyncSetupCommand creates the sync setup subcommand.
//
// The setup command displays bootstrap instructions for new team members,
// providing step-by-step guidance based on the organization's sync strategy.
//
// Returns:
//   - *cobra.Command: Configured sync setup command
func createSyncSetupCommand() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Show setup instructions for new hires",
		Long: `Display bootstrap instructions for new team members.

Provides step-by-step guidance based on your organization's sync strategy:
  - HTTP: Simple wget/curl instructions
  - S3 Public: AWS CLI commands without authentication
  - S3 SSO: SSO login and profile setup workflow

Perfect for onboarding new engineers to your AWS environment.
`,
		RunE: runSyncSetup,
	}

	return setupCmd
}

// runSyncFetch executes the sync fetch command.
//
// This function handles the fetch command workflow:
//  1. Parse command flags (force, verbose)
//  2. Get sync settings and build configuration
//  3. Call sync.Sync() with CLI reporter to show progress
//  4. Display fetch results to user
//
// Parameters:
//   - cmd: Cobra command instance
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during fetch
func runSyncFetch(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Sync fetch command started")

	// Parse flags
	force, _ := cmd.Flags().GetBool("force")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get sync settings
	syncSettings, err := getSyncSettings()
	if err != nil {
		return err
	}

	// Check if sync is enabled
	if !syncSettings.Enabled {
		logging.Log.Warn("⚠️  Sync is not enabled in settings")
		logging.Log.Info("")
		logging.Log.Info("To enable sync, update your settings")
		logging.Log.Info("")
		logging.Log.Info("Run 'aws-profile-manager sync setup' for detailed instructions")
		return nil
	}

	// Build sync configuration from settings
	logging.Debug.Log("\t🔹 Building sync configuration")
	cfg := sync.ConfigFromSettings(syncSettings)

	// Build sync options
	opts := sync.Options{
		ForceRefresh: force,
	}

	// Create CLI reporter for progress updates
	reporter := task.CliReporter{}

	// Call sync API with context and reporter
	logging.Log.Info("🔄 Starting configuration sync...")
	logging.Log.Info("")

	ctx := context.Background()
	result, err := sync.Sync(ctx, cfg, opts, reporter)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to fetch configuration", err)
	}

	// Display results
	return displayFetchResult(result, verbose)
}

// displayFetchResult formats and displays the fetch result.
//
// This function presents fetch results to the user in a clear, informative way,
// showing sync strategy, data size, cache status, and relevant details.
//
// Parameters:
//   - result: Sync result from sync package
//   - verbose: Whether to show detailed information
//
// Returns:
//   - error: Always nil (kept for consistency)
func displayFetchResult(result *sync.Result, verbose bool) error {
	logging.Log.Info("")

	// Success message
	logging.Log.Success("✅ Configuration synced successfully!")
	logging.Log.Info("")

	// Show source
	if result.CacheHit {
		cacheAge := time.Since(result.FetchTime)
		logging.Log.Infof("💾 Source: Cache (%.0f seconds old)", cacheAge.Seconds())
	} else {
		logging.Log.Infof("🌐 Source: %s", result.Source)
	}

	// Show data size if fetched
	if result.BytesTransferred > 0 {
		kb := float64(result.BytesTransferred) / 1024.0
		logging.Log.Infof("📦 Size: %.1f KB", kb)
	}

	// Show duration
	logging.Log.Infof("⏱️  Duration: %v", result.Duration.Round(time.Millisecond))

	logging.Log.Info("")
	logging.Log.Info("💡 Next: Run 'aws-profile-manager install' to apply profiles")

	return nil
}

// runSyncStatus executes the sync status command.
//
// This function handles the status command workflow:
//  1. Parse command flags (verbose)
//  2. Get sync settings
//  3. Check cache status
//  4. Display status information to user
//
// Parameters:
//   - cmd: Cobra command instance
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during status check
func runSyncStatus(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Sync status command started")

	// Parse flags
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get sync settings
	syncSettings, err := getSyncSettings()
	if err != nil {
		return err
	}

	// Build sync configuration
	cfg := sync.ConfigFromSettings(syncSettings)

	// Display status
	return displaySyncStatus(syncSettings, cfg, verbose)
}

// displaySyncStatus formats and displays the sync status.
//
// This function presents sync configuration and cache status to the user,
// showing sync strategy, cache age, source information, and helpful tips.
//
// Parameters:
//   - syncSettings: Sync settings from application
//   - cfg: Sync configuration
//   - verbose: Whether to show detailed information
//
// Returns:
//   - error: Any error encountered
func displaySyncStatus(syncSettings *settings.SyncSettings, cfg sync.SyncConfig, verbose bool) error {
	logging.Log.Info("📊 Sync Configuration Status")
	logging.Log.Info("")

	// Enabled status
	if syncSettings.Enabled {
		logging.Log.Success("✅ Sync: Enabled")
	} else {
		logging.Log.Warn("⚠️  Sync: Disabled")
		logging.Log.Info("")
		logging.Log.Info("Run 'aws-profile-manager sync setup' for configuration instructions")
		return nil
	}

	// Strategy and source
	logging.Log.Infof("📡 Strategy: %s", cfg.Strategy)

	// Display source based on strategy
	switch cfg.Strategy {
	case sync.StrategyHTTP:
		logging.Log.Infof("🔗 Source: %s", cfg.HTTPUrl)
	case sync.StrategyS3:
		logging.Log.Infof("🔗 Source: s3://%s/%s", cfg.S3Bucket, cfg.S3Key)
	case sync.StrategyLocal:
		logging.Log.Infof("🔗 Source: %s", cfg.LocalPath)
	}

	logging.Log.Info("")

	// Check cache status
	cache := sync.NewCache(cfg.CacheTTL, cfg.CacheDir)
	entry, err := cache.Get()

	if err != nil {
		logging.Log.Warn("💾 Cache: Error reading cache")
		if verbose {
			logging.Log.Infof("   Error: %v", err)
		}
	} else if entry != nil {
		logging.Log.Success("💾 Cache: Available")

		age := time.Since(entry.FetchTime)
		ageMinutes := int(age.Minutes())
		if ageMinutes < 60 {
			logging.Log.Infof("   Age: %d minutes", ageMinutes)
		} else if ageMinutes < 1440 {
			logging.Log.Infof("   Age: %.1f hours", float64(ageMinutes)/60.0)
		} else {
			logging.Log.Infof("   Age: %.1f days", float64(ageMinutes)/1440.0)
		}

		if verbose {
			logging.Log.Infof("   Fetched from: %s", entry.Source)
		}
	} else {
		logging.Log.Warn("💾 Cache: None")
		logging.Log.Info("   Run 'aws-profile-manager sync fetch' to cache configuration")
	}

	logging.Log.Info("")

	return nil
}

// runSyncClearCache executes the sync clear-cache command.
//
// This function handles the clear-cache command workflow:
//  1. Get sync settings
//  2. Call sync.ClearCache() to remove cached configuration
//  3. Display results to user
//
// Parameters:
//   - cmd: Cobra command instance
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during cache clearing
func runSyncClearCache(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Sync clear-cache command started")

	// Get sync settings
	syncSettings, err := getSyncSettings()
	if err != nil {
		return err
	}

	// Build sync configuration
	cfg := sync.ConfigFromSettings(syncSettings)

	// Clear cache
	logging.Log.Info("�️  Clearing cache...")
	err = sync.ClearCache(cfg)
	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to clear cache", err)
	}

	logging.Log.Success("✅ Cache cleared successfully")
	return nil
}

// runSyncSetup executes the sync setup command.
//
// This function displays setup instructions for configuring sync.
//
// Parameters:
//   - cmd: Cobra command instance
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered
func runSyncSetup(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Sync setup command started")

	// Get sync settings
	syncSettings, err := getSyncSettings()
	if err != nil {
		return err
	}

	// Display setup instructions
	logging.Log.Info("🚀 Sync Setup Instructions")
	logging.Log.Info("")

	if syncSettings.Enabled {
		logging.Log.Success("✅ Sync is already enabled!")
		logging.Log.Info("")
		logging.Log.Infof("Strategy: %s", syncSettings.Strategy)

		switch sync.Strategy(syncSettings.Strategy) {
		case sync.StrategyHTTP:
			logging.Log.Infof("URL: %s", syncSettings.HTTP.URL)
		case sync.StrategyS3:
			logging.Log.Infof("Bucket: s3://%s/%s", syncSettings.S3.Bucket, syncSettings.S3.Key)
		case sync.StrategyLocal:
			logging.Log.Infof("Path: %s", syncSettings.Local.Path)
		}

		logging.Log.Info("")
		logging.Log.Info("💡 Run 'aws-profile-manager sync fetch' to download configuration")
	} else {
		logging.Log.Warn("⚠️  Sync is not enabled")
		logging.Log.Info("")
		logging.Log.Info("To enable sync, configure your settings:")
		logging.Log.Info("")
		logging.Log.Info("HTTP Strategy:")
		logging.Log.Info(`  {
    "sync": {
      "enabled": true,
      "strategy": "http",
      "http": {
        "url": "https://your-server.com/aws-config.json"
      }
    }
  }`)
		logging.Log.Info("")
		logging.Log.Info("S3 Strategy:")
		logging.Log.Info(`  {
    "sync": {
      "enabled": true,
      "strategy": "s3",
      "s3": {
        "bucket": "your-bucket",
        "key": "aws-config.json",
        "region": "us-east-1"
      }
    }
  }`)
		logging.Log.Info("")
		logging.Log.Info("Local Strategy (for testing):")
		logging.Log.Info(`  {
    "sync": {
      "enabled": true,
      "strategy": "local",
      "local": {
        "path": "/path/to/aws-config.json"
      }
    }
  }`)
		logging.Log.Info("")
		logging.Log.Info("Git Strategy:")
		logging.Log.Info(`  {
    "sync": {
      "enabled": true,
      "strategy": "git",
      "git": {
        "repo_url": "https://github.com/org/repo.git",
        "branch": "main",
        "file_path": "config.json"
      }
    }
  }`)
	}

	return nil
}
