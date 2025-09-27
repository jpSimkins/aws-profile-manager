package cli

import (
	"github.com/spf13/cobra"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
)

// runVersion displays application version and build information.
//
// This command handler retrieves version details from the core package and
// displays them in a user-friendly format.
//
// Displayed Information:
//   - Application version number
//   - Git commit hash (if available)
//   - Build date (if available)
//   - Go version used for compilation
//   - Platform (OS/architecture)
//
// Parameters:
//   - cmd: Cobra command context (unused)
//   - args: Command arguments (unused)
func runVersion(cmd *cobra.Command, args []string) {
	versionInfo := core.GetVersion()

	logging.Log.Info("AWS Profile Manager Version Information")
	logging.Log.Infof("Version: %s", versionInfo.Version)

	if versionInfo.Commit != "" {
		logging.Log.Infof("Commit: %s", versionInfo.Commit)
	}

	if versionInfo.Date != "" {
		logging.Log.Infof("Build Date: %s", versionInfo.Date)
	}

	logging.Log.Infof("Go Version: %s", versionInfo.GoVersion)
	logging.Log.Infof("Platform: %s", versionInfo.Platform)
}
