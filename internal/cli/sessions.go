package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/logging"
)

// runSessions executes the sessions command to list AWS SSO sessions.
//
// This is the main command handler that checks SSO session status and displays
// which sessions are active vs expired.
//
// Command Flow:
//  1. Parse flags (verbose, refresh)
//  2. Call awscli.GetSessionStatus() API (ONE call)
//  3. Display results via helper function
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during execution
func runSessions(cmd *cobra.Command, args []string) error {
	return runSessionsWithExtractor(cmd, args, nil)
}

// runSessionsWithExtractor executes the sessions command with an optional custom extractor.
//
// This function supports both production and test modes. In production, it uses the
// high-level API. In test mode, it accepts a custom extractor for controlled testing.
//
// Parameters:
//   - cmd: Cobra command context
//   - _: Command arguments (unused)
//   - customExtractor: Optional extractor for testing (nil in production)
//
// Returns:
//   - error: Any error encountered during execution
func runSessionsWithExtractor(cmd *cobra.Command, _ []string, customExtractor *awscli.Extractor) error {
	logging.Debug.Log("Sessions command started")

	// Parse flags
	logging.Debug.Log("\t🔹 Parsing command flags")
	verbose, _ := cmd.Flags().GetBool("verbose")
	refresh, _ := cmd.Flags().GetBool("refresh")

	if verbose {
		logging.Log.Info("Verbose mode enabled for sessions command")
	}

	// Get session status via API
	logging.Debug.Log("\t🔹 Calling awscli.GetSessionStatus API", "force_refresh", refresh)
	var status awscli.SessionStatus
	var err error

	if customExtractor != nil {
		// Test mode with custom extractor
		sessionManager := awscli.NewSessionManagerWithExtractor(customExtractor)
		if refresh {
			logging.Log.Info("🔄 Refreshing session status...")
			status, err = sessionManager.RefreshSessionStatus()
		} else {
			status, err = sessionManager.GetSessionStatus()
		}
	} else {
		// Production mode - use high-level API
		status, err = awscli.GetSessionStatus(refresh)
	}

	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to get session status", err)
	}

	logging.Debug.Log("\t🔹 Displaying results",
		"active_sessions", len(status.ActiveSessions),
		"expired_sessions", len(status.ExpiredSessions))

	// Display results (presentation only)
	return displaySessionStatus(status, verbose)
}

// displaySessionStatus formats and displays the session status information.
//
// This function handles the presentation layer for session status, formatting
// the data into a user-friendly console output.
//
// Display Format:
//   - AWS CLI availability and version
//   - Active sessions with expiration times
//   - Expired sessions
//   - Detailed session information (if verbose)
//
// Parameters:
//   - status: Session status data to display
//   - verbose: Whether to show detailed information
//
// Returns:
//   - error: Any error encountered during display
func displaySessionStatus(status awscli.SessionStatus, verbose bool) error {
	logging.Log.Info("📊 AWS SSO Session Status")
	logging.Log.Info("")

	// Display CLI availability
	if status.CLIAvailable {
		logging.Log.Successf("✅ AWS CLI available (version: %s)", status.CLIVersion)
	} else {
		logging.Log.Warn("⚠️  AWS CLI not found or not available")
		logging.Log.Info("   Install AWS CLI v2 to use SSO sessions")
		return nil
	}

	// Display last checked time
	if verbose {
		logging.Log.Infof("📅 Last checked: %s", status.LastChecked.Format(time.RFC3339))
	}

	logging.Log.Info("")

	// Display active sessions
	if len(status.ActiveSessions) > 0 {
		logging.Log.Successf("🟢 Active Sessions (%d)", len(status.ActiveSessions))
		for _, session := range status.ActiveSessions {
			displaySessionInfo(session, verbose, true)
		}
	} else {
		logging.Log.Info("🟢 Active Sessions: None")
	}

	logging.Log.Info("")

	// Display expired sessions
	if len(status.ExpiredSessions) > 0 {
		logging.Log.Warnf("🔴 Expired Sessions (%d)", len(status.ExpiredSessions))
		for _, session := range status.ExpiredSessions {
			displaySessionInfo(session, verbose, false)
		}
		logging.Log.Info("")
		logging.Log.Info("💡 Run 'aws sso login --profile <profile-name>' to refresh expired sessions")
	} else {
		logging.Log.Info("🔴 Expired Sessions: None")
	}

	// Summary
	totalSessions := len(status.ActiveSessions) + len(status.ExpiredSessions)
	if totalSessions == 0 {
		logging.Log.Info("")
		logging.Log.Info("ℹ️  No SSO sessions found")
		logging.Log.Info("   Configure AWS CLI with SSO to see sessions here")
	}

	return nil
}

// displaySessionInfo displays detailed information for a single SSO session.
//
// This function formats and displays session information including status indicator,
// session name, URLs, expiration time, and cache file location. The display adapts
// based on verbose mode and session state (active/expired).
//
// Display Format:
//   - Status indicator: 🟢 (active) or 🔴 (expired)
//   - Session name with indentation
//   - Verbose mode: Start URL, region, expiration details, cache path
//   - Simple mode: Only expiration time in friendly format
//
// Parameters:
//   - session: Session information to display
//   - verbose: Whether to show detailed information
//   - isActive: Whether the session is currently active (affects display)
func displaySessionInfo(session awscli.ActiveSessionInfo, verbose bool, isActive bool) {
	status := "🟢"
	if !isActive {
		status = "🔴"
	}

	// Basic session info
	logging.Log.Infof("   %s %s", status, session.SessionName)

	if verbose {
		logging.Log.Infof("      🌐 Start URL: %s", session.StartURL)
		logging.Log.Infof("      🌍 Region: %s", session.Region)

		if isActive {
			timeUntilExpiry := time.Until(session.ExpiresAt)
			if timeUntilExpiry > 0 {
				logging.Log.Infof("      ⏰ Expires: %s (in %s)",
					session.ExpiresAt.Local().Format("2006-01-02 15:04:05"),
					formatDuration(timeUntilExpiry))
			}
		} else {
			timeSinceExpiry := time.Since(session.ExpiresAt)
			logging.Log.Infof("      ⏰ Expired: %s (%s ago)",
				session.ExpiresAt.Local().Format("2006-01-02 15:04:05"),
				formatDuration(timeSinceExpiry))
		}

		if session.CacheFilePath != "" {
			logging.Log.Infof("      📁 Cache: %s", session.CacheFilePath)
		}
	} else {
		// Simplified view
		if isActive {
			timeUntilExpiry := time.Until(session.ExpiresAt)
			if timeUntilExpiry > 0 {
				logging.Log.Infof("      ⏰ Expires in %s", formatDuration(timeUntilExpiry))
			}
		} else {
			timeSinceExpiry := time.Since(session.ExpiresAt)
			logging.Log.Infof("      ⏰ Expired %s ago", formatDuration(timeSinceExpiry))
		}
	}
}

// formatDuration formats a duration in a human-readable, friendly format.
//
// This function converts Go's time.Duration into a more readable format suitable
// for displaying to users. It handles negative durations by converting them to
// positive and showing the appropriate sign in the calling context.
//
// Format Examples:
//   - 2h 30m for durations over an hour
//   - 45m for durations under an hour but over a minute
//   - 30s for durations under a minute
//
// Parameters:
//   - d: Duration to format
//
// Returns:
//   - string: Human-readable duration (e.g., "2h 30m", "45m", "30s")
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	seconds := int(d.Seconds())
	return fmt.Sprintf("%ds", seconds)
}
