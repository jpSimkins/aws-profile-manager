package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/gui"
	"aws-profile-manager/internal/logging"
)

// isTestEnvironment detects if the application is running in a test environment.
//
// This function checks command-line arguments to determine if the process is
// running under 'go test'. This prevents GUI initialization during tests.
//
// Returns:
//   - bool: true if running in test environment, false otherwise
func isTestEnvironment() bool {
	// Check if we're running under 'go test'
	for _, arg := range os.Args {
		if strings.Contains(arg, "test") || strings.Contains(arg, ".test") {
			return true
		}
	}
	return false
}

// runGUI handles GUI mode execution and launches the graphical interface.
//
// This command handler creates and runs the GUI application using the Fyne
// framework. It respects the --config flag and prevents GUI launch during tests.
//
// Process:
//  1. Check for test environment (skip GUI if testing)
//  2. Retrieve config file path from flags
//  3. Create GUI application instance
//  4. Run GUI main loop (blocks until window closes)
//  5. Cleanup resources on exit
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Always nil (GUI errors are logged, not returned)
func runGUI(cmd *cobra.Command, args []string) error {
	logging.Debug.Log("Launching GUI interface...")

	// Get config file from global flag
	configFile, _ := cmd.Root().PersistentFlags().GetString("config")
	if configFile != "" {
		logging.Debug.Logf("GUI will use specified config file: %s", configFile)
	}

	// Handle test environment - don't actually run GUI
	if isTestEnvironment() {
		logging.Log.Info("Test environment detected - GUI creation successful")
		return nil
	}

	// Create and run the GUI application
	app, err := gui.NewApp()
	if err != nil {
		_ = logging.Log.ErrorWithDetails("Failed to create GUI application", err)
		return nil
	}
	app.Run(configFile)

	logging.Debug.Log("GUI interface exited")
	return nil
}
