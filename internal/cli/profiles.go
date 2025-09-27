package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/logging"
)

// runProfiles executes the profiles command to list and filter AWS CLI profiles.
//
// This is the main command handler that retrieves existing AWS CLI profiles
// and displays them according to filter criteria.
//
// Command Flow:
//  1. Parse flags (verbose, filters)
//  2. Build FilterCriteria from flags
//  3. Call awscli.ListProfiles() API (ONE call)
//  4. Display results via helper function
//
// Parameters:
//   - cmd: Cobra command context
//   - args: Command arguments (unused)
//
// Returns:
//   - error: Any error encountered during execution
func runProfiles(cmd *cobra.Command, args []string) error {
	return runProfilesWithExtractor(cmd, args, nil)
}

// runProfilesWithExtractor executes the profiles command with an optional custom extractor.
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
func runProfilesWithExtractor(cmd *cobra.Command, _ []string, customExtractor *awscli.Extractor) error {
	logging.Debug.Log("Profiles command started")

	// Parse flags into filter criteria
	logging.Debug.Log("\t🔹 Parsing command flags")
	verbose, _ := cmd.Flags().GetBool("verbose")
	accountID, _ := cmd.Flags().GetString("account-id")
	role, _ := cmd.Flags().GetString("role")
	region, _ := cmd.Flags().GetString("region")
	session, _ := cmd.Flags().GetString("session")
	pattern, _ := cmd.Flags().GetString("pattern")

	if verbose {
		logging.Log.Info("Verbose mode enabled for profiles command")
	}

	// Build filter criteria from command line flags
	logging.Debug.Log("\t🔹 Building filter criteria")
	var criteria awscli.FilterCriteria

	if accountID != "" {
		criteria.AccountIDs = []string{accountID}
	}
	if role != "" {
		criteria.RoleNames = []string{role}
	}
	if region != "" {
		criteria.Regions = []string{region}
	}
	if session != "" {
		criteria.SsoSessions = []string{session}
	}
	if pattern != "" {
		criteria.NamePattern = pattern
	}

	// Call API to get profiles (business logic happens here)
	logging.Debug.Log("\t🔹 Calling awscli.ListProfiles API")
	var result *awscli.ProfilesResult
	var err error

	if customExtractor != nil {
		// Test mode with custom extractor
		result, err = awscli.ListProfilesWithExtractor(criteria, customExtractor)
	} else {
		// Production mode
		result, err = awscli.ListProfiles(criteria)
	}

	if err != nil {
		return logging.Log.ErrorfWithDetails("failed to list profiles", err)
	}

	logging.Debug.Log("\t🔹 Displaying results",
		"profile_count", len(result.Profiles),
		"session_count", len(result.SsoSessions))

	// Display results (presentation only)
	return displayProfiles(result.Profiles, result.SsoSessions, result.SessionStatus, verbose, result.ConfigPath)
}

// displayProfiles formats and displays profile information to the console.
//
// This function handles the presentation layer for profile listing, including:
//   - Header with config file path (verbose mode)
//   - Profile count summary
//   - Individual profile details with color-coding
//   - SSO session status indicators
//   - Organization and account metadata
//
// Display Format:
//   - Profile name with colored status indicator
//   - Account ID, Role, Region
//   - SSO session name (for SSO profiles)
//   - Organization and account names (if available)
//   - Additional properties in verbose mode
//
// Parameters:
//   - profiles: List of profiles to display
//   - sessions: SSO sessions for reference
//   - sessionStatus: SSO session status information
//   - verbose: Whether to show detailed information
//   - configPath: Path to AWS CLI config file
//
// Returns:
//   - error: Any error encountered during display
func displayProfiles(profiles []awscli.AwsCliProfile, sessions []awscli.SsoSession, sessionStatus awscli.SessionStatus, verbose bool, configPath string) error {
	logging.Log.Info("📝 AWS CLI Profiles")

	if verbose {
		logging.Log.Infof("📁 Config file: %s", configPath)
	}

	logging.Log.Info("")

	if len(profiles) == 0 {
		logging.Log.Info("ℹ️  No profiles found")
		if configPath != "" {
			logging.Log.Info("   Check your AWS CLI configuration file")
		}
		return nil
	}

	logging.Log.Infof("📊 Found %d profile(s)", len(profiles))
	logging.Log.Info("")

	// Create a map of active session names for quick lookup
	activeSessions := make(map[string]bool)
	for _, activeSession := range sessionStatus.ActiveSessions {
		activeSessions[activeSession.SessionName] = true
	}

	// Group profiles by SSO session for better organization
	profilesBySession := make(map[string][]awscli.AwsCliProfile)
	profilesWithoutSession := []awscli.AwsCliProfile{}

	for _, profile := range profiles {
		if profile.SsoSession != "" {
			profilesBySession[profile.SsoSession] = append(profilesBySession[profile.SsoSession], profile)
		} else {
			profilesWithoutSession = append(profilesWithoutSession, profile)
		}
	}

	// Display profiles grouped by SSO session
	sessionNames := make([]string, 0, len(profilesBySession))
	for sessionName := range profilesBySession {
		sessionNames = append(sessionNames, sessionName)
	}
	sort.Strings(sessionNames)

	for _, sessionName := range sessionNames {
		sessionProfiles := profilesBySession[sessionName]
		isActive := activeSessions[sessionName]

		sessionIcon := "🔴"
		sessionStatusText := "inactive"
		if isActive {
			sessionIcon = "🟢"
			sessionStatusText = "active"
		}

		logging.Log.Infof("%s SSO Session: %s (%s)", sessionIcon, sessionName, sessionStatusText)

		for _, profile := range sessionProfiles {
			displayProfileInfo(profile, verbose, isActive, "   ")
		}
		logging.Log.Info("")
	}

	// Display profiles without SSO sessions
	if len(profilesWithoutSession) > 0 {
		logging.Log.Info("🔧 Non-SSO Profiles")
		for _, profile := range profilesWithoutSession {
			displayProfileInfo(profile, verbose, false, "   ")
		}
		logging.Log.Info("")
	}

	// Display summary
	if verbose && len(sessions) > 0 {
		logging.Log.Infof("📋 Available SSO Sessions: %d", len(sessions))
		for _, session := range sessions {
			isActive := activeSessions[session.Name]
			icon := "🔴"
			if isActive {
				icon = "🟢"
			}
			logging.Log.Infof("   %s %s (%s)", icon, session.Name, session.StartURL)
		}
	}

	return nil
}

// displayProfileInfo displays detailed information for a single profile.
//
// This function formats and outputs a single profile's information with appropriate
// visual indicators and hierarchical indentation.
//
// Display Elements:
//   - Status icon: 🟢 (active SSO), 🔴 (expired SSO), ⚪ (non-SSO)
//   - Profile name with account/role
//   - Region information
//   - SSO session reference
//   - Organization and account metadata
//   - Additional properties (verbose mode only)
//
// Parameters:
//   - profile: Profile to display
//   - verbose: Whether to show additional properties
//   - sessionActive: Whether the profile's SSO session is active
//   - indent: Indentation string for hierarchical display
func displayProfileInfo(profile awscli.AwsCliProfile, verbose bool, sessionActive bool, indent string) {
	// Profile status indicator
	icon := "⚪"
	if profile.SsoSession != "" {
		if sessionActive {
			icon = "🟢"
		} else {
			icon = "🔴"
		}
	}

	// Main profile line
	profileLine := fmt.Sprintf("%s%s %s", indent, icon, profile.Name)

	// Add account info if available
	if profile.AccountID != "" && profile.RoleName != "" {
		profileLine += fmt.Sprintf(" → %s/%s", profile.AccountID, profile.RoleName)
	}

	logging.Log.Info(profileLine)

	if verbose {
		// Show detailed information
		if profile.AccountID != "" {
			logging.Log.Infof("%s   🏦 Account: %s", indent, profile.AccountID)
		}
		if profile.RoleName != "" {
			logging.Log.Infof("%s   👤 Role: %s", indent, profile.RoleName)
		}
		if profile.Region != "" {
			logging.Log.Infof("%s   🌍 Region: %s", indent, profile.Region)
		}
		if profile.SsoSession != "" {
			status := "inactive"
			if sessionActive {
				status = "active"
			}
			logging.Log.Infof("%s   🔐 SSO Session: %s (%s)", indent, profile.SsoSession, status)
		}
		if profile.SsoStartURL != "" {
			logging.Log.Infof("%s   🌐 SSO URL: %s", indent, profile.SsoStartURL)
		}

		// Show any additional properties
		if len(profile.Properties) > 0 {
			var propKeys []string
			for key := range profile.Properties {
				// Skip properties we've already shown
				if !isStandardProperty(key) {
					propKeys = append(propKeys, key)
				}
			}

			if len(propKeys) > 0 {
				sort.Strings(propKeys)
				logging.Log.Infof("%s   ⚙️  Additional properties:", indent)
				for _, key := range propKeys {
					logging.Log.Infof("%s      %s: %s", indent, key, profile.Properties[key])
				}
			}
		}
	} else {
		// Simplified view - show key info on same line
		details := []string{}

		if profile.Region != "" {
			details = append(details, profile.Region)
		}

		if profile.SsoSession != "" && sessionActive {
			details = append(details, "SSO active")
		} else if profile.SsoSession != "" {
			details = append(details, "SSO inactive")
		}

		if len(details) > 0 {
			logging.Log.Infof("%s     (%s)", indent, strings.Join(details, ", "))
		}
	}
}

// isStandardProperty checks if a property key is a standard AWS CLI property.
//
// Standard properties are displayed separately in the main profile view and
// excluded from the "Additional properties" section in verbose mode. This
// prevents duplicate display of commonly-used properties.
//
// Standard Properties:
//   - account_id, role_name, region
//   - sso_session, sso_start_url, sso_region
//   - sso_account_id, sso_role_name
//
// Parameters:
//   - key: Property key to check
//
// Returns:
//   - bool: true if the property is standard, false if it should be shown in additional properties
func isStandardProperty(key string) bool {
	standardProps := map[string]bool{
		"account_id":     true,
		"role_name":      true,
		"region":         true,
		"sso_session":    true,
		"sso_start_url":  true,
		"sso_region":     true,
		"sso_account_id": true,
		"sso_role_name":  true,
	}
	return standardProps[key]
}
