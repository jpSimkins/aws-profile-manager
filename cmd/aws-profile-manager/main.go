package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"aws-profile-manager/internal/cli"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

var (
	// Global flags
	configFile string
)

// initializeApp handles the core application initialization
func initializeApp() error {
	logging.Debug.Log("Application started")

	// Initialize application - loads settings from disk
	if err := core.App.Initialize(); err != nil {
		_ = logging.Log.ErrorWithDetails("❗ Failed to initialize application", err)
		return err
	}

	// Handle --config flag override for sync.local.path
	if configFile != "" {
		logging.Debug.Logf("AWS config file specified: %s", configFile)

		// Get current settings
		currentSettings := settings.Get()

		// Override sync.local.path
		currentSettings.Sync.Local.Path = configFile

		// Update settings in memory
		if err := settings.Set(currentSettings); err != nil {
			return logging.Log.ErrorfWithDetails("failed to set sync.local.path override", err)
		}

		logging.Debug.Log("Set sync.local.path override from --config flag")
	}

	return nil
}

// loadDotEnv loads environment variables from .env file for development
// Only loads in development (when not in test environment)
func loadDotEnv() {
	// Skip .env loading in tests to preserve test isolation
	if os.Getenv("GO_ENV") == "test" {
		return
	}

	// Try to load .env file (ignore errors - it's optional)
	if err := godotenv.Load(); err != nil {
		// .env file not found or couldn't be loaded - this is fine
		logging.Debug.Log("No .env file found (this is normal for production)")
	} else {
		logging.Debug.Log(".env file loaded successfully")
	}
}

func main() {
	// Load .env file for development (before any other initialization)
	loadDotEnv()

	// Create the root command
	var rootCmd = &cobra.Command{
		Use:   "aws-profile-manager",
		Short: "AWS Profile Manager - A cross-platform AWS profile management tool",
		Long: `AWS Profile Manager is a comprehensive tool for managing AWS profiles
and credentials across different environments. It provides both CLI and GUI
interfaces for seamless AWS profile management.

Features:
  - CLI and GUI interfaces
  - Cross-platform support
  - Profile import/export functionality
  - Active profile insights`,
		Version: core.GetVersionString(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Initialize the application before any command runs
			_ = initializeApp()
		},
		// When invoked with no subcommand (e.g. from a desktop launcher), launch the GUI.
		RunE: func(cmd *cobra.Command, args []string) error {
			guiCmd, _, err := cmd.Find([]string{"gui"})
			if err != nil || guiCmd == nil {
				return cmd.Help()
			}
			return guiCmd.RunE(guiCmd, args)
		},
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file for AWS profile data (not the same as settings config)")

	// Register all commands (CLI + GUI/version)
	cli.RegisterCommands(rootCmd)

	// Set custom version template
	rootCmd.SetVersionTemplate(`{{printf "%s\n" .Version}}`)

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		_ = logging.Log.ErrorWithDetails("Command execution failed", err)
		os.Exit(1)
	}

	// Log application exit
	logging.Debug.Log("Application exited")
}
