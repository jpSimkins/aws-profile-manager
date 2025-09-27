package awscli

import (
	"sort"

	"aws-profile-manager/internal/logging"
)

// ProfilesResult contains all data needed for displaying AWS CLI profiles.
//
// This struct is returned by ListProfiles() and provides everything needed
// for CLI or GUI display, including filtered profiles, SSO sessions for context,
// and current session status.
type ProfilesResult struct {
	Profiles      []AwsCliProfile // Filtered profiles matching the criteria
	SsoSessions   []SsoSession    // All SSO sessions for reference
	SessionStatus SessionStatus   // Current session states (active/expired)
	ConfigPath    string          // Path to AWS CLI config file used
}

// FilterOptionsResult contains all available filter values from current profiles.
//
// This struct is returned by GetFilterOptions() to help users build filter UIs
// by showing what values are actually present in their AWS CLI config.
type FilterOptionsResult struct {
	AccountIDs  []string // All unique account IDs found in profiles
	RoleNames   []string // All unique role names found in profiles
	Regions     []string // All unique regions found in profiles
	SsoSessions []string // All unique SSO session names found
}

// =============================================================================
// HIGH-LEVEL API FUNCTIONS
// =============================================================================
// These are the main entry points that CLI and GUI should call.
// They orchestrate internal components (Extractor, Filter, SessionManager)
// and return complete, ready-to-use results.

// ListProfiles retrieves and filters AWS CLI profiles with session status.
//
// This is the main API function that CLI and GUI should call for profile listing.
// It orchestrates all necessary operations in a single call, providing complete
// results ready for display.
//
// Process:
//  1. Extract profiles and SSO sessions from ~/.aws/config (once)
//  2. Apply filter criteria to profiles
//  3. Sort profiles alphabetically by name
//  4. Get session status for all SSO sessions
//  5. Return complete result with all data
//
// Filtering:
//   - All criteria are combined with AND logic
//   - Multiple values within a field use OR logic
//   - Empty criteria matches all profiles
//
// Performance:
//   - Uses caching to avoid repeated file reads
//   - Extracts profiles once, reuses for session status
//   - Efficient for repeated calls
//
// Parameters:
//   - criteria: Filter criteria (account IDs, regions, roles, patterns, etc.)
//
// Returns:
//   - *ProfilesResult: Complete data including filtered profiles, sessions, and status
//   - error: Any error encountered during extraction or processing
//
// Example:
//
//	result, err := awscli.ListProfiles(awscli.FilterCriteria{
//	    AccountIDs: []string{"123456789012"},
//	    RoleNames:  []string{"Administrator", "Developer"},
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, profile := range result.Profiles {
//	    fmt.Printf("Profile: %s (Account: %s, Role: %s)\n",
//	        profile.Name, profile.AccountID, profile.RoleName)
//	}
func ListProfiles(criteria FilterCriteria) (*ProfilesResult, error) {
	logging.Debug.Log("ListProfiles API called")
	logging.Debug.Log("\t🔹 Filter criteria",
		"account_ids", len(criteria.AccountIDs),
		"regions", len(criteria.Regions),
		"roles", len(criteria.RoleNames),
		"sessions", len(criteria.SsoSessions),
		"pattern", criteria.NamePattern,
		"types", len(criteria.ProfileTypes))

	// Extract profiles from AWS CLI config (once)
	logging.Debug.Log("\t🔹 Extracting AWS CLI profiles")
	extractor := NewExtractor()
	data, err := extractor.ExtractFromFile()
	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("failed to extract AWS CLI profiles", err)
	}

	logging.Debug.Log("\t🔹 Extraction complete",
		"profiles", len(data.Profiles),
		"sso_sessions", len(data.SsoSessions))

	// Apply filters
	logging.Debug.Log("\t🔹 Applying filters")
	filter := NewFilter()
	filteredProfiles := filter.FilterProfiles(data.Profiles, criteria)

	logging.Debug.Log("\t🔹 Filtering complete",
		"filtered_profiles", len(filteredProfiles))

	// Sort profiles by name for consistent output
	logging.Debug.Log("\t🔹 Sorting profiles")
	sort.Slice(filteredProfiles, func(i, j int) bool {
		return filteredProfiles[i].Name < filteredProfiles[j].Name
	})

	// Get session status using already-extracted sessions
	logging.Debug.Log("\t🔹 Getting session status")
	sessionManager := NewSessionManager(extractor)
	sessionStatus, err := sessionManager.GetSessionStatusWithSessions(data.SsoSessions)
	if err != nil {
		logging.Log.Warnf("Failed to get session status: %v", err)
		// Continue with empty session status rather than failing
		sessionStatus = SessionStatus{}
	}

	logging.Debug.Log("\t🔹 Session status retrieved",
		"active_sessions", len(sessionStatus.ActiveSessions),
		"expired_sessions", len(sessionStatus.ExpiredSessions))

	// Return complete result
	result := &ProfilesResult{
		Profiles:      filteredProfiles,
		SsoSessions:   data.SsoSessions,
		SessionStatus: sessionStatus,
		ConfigPath:    extractor.GetConfigPath(),
	}

	logging.Debug.Log("✅ ListProfiles completed successfully",
		"total_profiles", len(result.Profiles),
		"total_sessions", len(result.SsoSessions))

	return result, nil
}

// ListProfilesWithExtractor retrieves and filters profiles using a custom extractor.
//
// This is a variant of ListProfiles() that accepts a custom extractor, primarily
// used for testing with specific config file paths or injected test data.
//
// Use Cases:
//   - Testing with custom AWS config file locations
//   - Injecting mock data for unit tests
//   - Processing multiple config files sequentially
//
// Parameters:
//   - criteria: Filter criteria (account IDs, regions, roles, patterns, etc.)
//   - extractor: Custom extractor configured with specific config path or test data
//
// Returns:
//   - *ProfilesResult: Complete data including filtered profiles, sessions, and status
//   - error: Any error encountered during extraction or processing
func ListProfilesWithExtractor(criteria FilterCriteria, extractor *Extractor) (*ProfilesResult, error) {
	logging.Debug.Log("ListProfilesWithExtractor API called (test variant)")

	// Extract profiles using provided extractor
	data, err := extractor.ExtractFromFile()
	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("failed to extract AWS CLI profiles", err)
	}

	logging.Debug.Log("\t🔹 Extracted from custom extractor",
		"profiles", len(data.Profiles),
		"sso_sessions", len(data.SsoSessions))

	// Apply filters
	filter := NewFilter()
	filteredProfiles := filter.FilterProfiles(data.Profiles, criteria)

	logging.Debug.Log("\t🔹 Filtered profiles",
		"count", len(filteredProfiles))

	// Sort profiles by name for consistent output
	sort.Slice(filteredProfiles, func(i, j int) bool {
		return filteredProfiles[i].Name < filteredProfiles[j].Name
	})

	// Get session status using already-extracted sessions
	sessionManager := NewSessionManager(extractor)
	sessionStatus, err := sessionManager.GetSessionStatusWithSessions(data.SsoSessions)
	if err != nil {
		logging.Log.Warnf("Failed to get session status: %v", err)
		sessionStatus = SessionStatus{}
	}

	// Return complete result
	result := &ProfilesResult{
		Profiles:      filteredProfiles,
		SsoSessions:   data.SsoSessions,
		SessionStatus: sessionStatus,
		ConfigPath:    extractor.GetConfigPath(),
	}

	logging.Debug.Log("✅ ListProfilesWithExtractor completed",
		"profiles", len(result.Profiles))

	return result, nil
}

// GetSessionStatus retrieves the current status of all AWS SSO sessions.
//
// This is the main API function that CLI and GUI should call for session management.
// It checks which SSO sessions have valid tokens and which have expired.
//
// Process:
//  1. Extract SSO session configurations from ~/.aws/config
//  2. Check ~/.aws/sso/cache/ for token files matching each session
//  3. Validate token expiration times
//  4. Categorize sessions as active or expired
//  5. Check AWS CLI availability and version
//
// Session Status:
//   - Active: Session has valid, non-expired token in cache
//   - Expired: Session has expired token or no token in cache
//
// Parameters:
//   - forceRefresh: If true, bypasses cache and forces fresh extraction
//
// Returns:
//   - SessionStatus: Complete session status with active/expired sessions and CLI info
//   - error: Any error encountered during extraction or cache checking
//
// Example:
//
//	status, err := awscli.GetSessionStatus(false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Active sessions: %d\n", len(status.ActiveSessions))
//	for _, session := range status.ActiveSessions {
//	    fmt.Printf("  %s expires at %s\n", session.SessionName, session.ExpiresAt)
//	}
func GetSessionStatus(forceRefresh bool) (SessionStatus, error) {
	logging.Debug.Log("GetSessionStatus API called",
		"force_refresh", forceRefresh)

	// Create extractor and session manager
	logging.Debug.Log("\t🔹 Creating extractor and session manager")
	extractor := NewExtractor()
	sessionManager := NewSessionManager(extractor)

	// Get session status (with optional force refresh)
	logging.Debug.Log("\t🔹 Getting session status")
	var status SessionStatus
	var err error

	if forceRefresh {
		logging.Debug.Log("\t\t🔸 Force refresh requested")
		status, err = sessionManager.RefreshSessionStatus()
	} else {
		status, err = sessionManager.GetSessionStatus()
	}

	if err != nil {
		return SessionStatus{}, logging.Log.ErrorfWithDetails("failed to get session status", err)
	}

	logging.Debug.Log("✅ GetSessionStatus completed",
		"active_sessions", len(status.ActiveSessions),
		"expired_sessions", len(status.ExpiredSessions),
		"cli_available", status.CLIAvailable)

	return status, nil
}

// LoginSession initiates AWS SSO login for the named session.
//
// This calls `aws sso login --sso-session <name>` which opens a browser for
// authentication. Requires AWS CLI to be installed. Blocks until login completes.
//
// Parameters:
//   - sessionName: Name of the SSO session to log in to
//
// Returns:
//   - error: Any error from AWS CLI or if CLI is unavailable
func LoginSession(sessionName string) error {
	logging.Debug.Log("LoginSession API called", "session", sessionName)
	extractor := NewExtractor()
	sm := NewSessionManager(extractor)
	return sm.LoginToSession(sessionName)
}

// LogoutAllSessions logs out from all AWS SSO sessions globally.
//
// This calls `aws sso logout` which invalidates all active SSO tokens.
// Note: AWS CLI does not support per-session logout.
//
// Returns:
//   - error: Any error from AWS CLI or if CLI is unavailable
func LogoutAllSessions() error {
	logging.Debug.Log("LogoutAllSessions API called")
	extractor := NewExtractor()
	sm := NewSessionManager(extractor)
	return sm.Logout()
}

// GetConfiguredSessions returns all SSO sessions declared in ~/.aws/config.
//
// Unlike GetSessionStatus, this returns every [sso-session …] entry regardless
// of whether a valid token exists in the cache. Use this to build a full session
// list and then overlay status from GetSessionStatus.
//
// Returns:
//   - []SsoSession: All configured SSO sessions
//   - error: Any error encountered during extraction
func GetConfiguredSessions() ([]SsoSession, error) {
	logging.Debug.Log("GetConfiguredSessions API called")
	extractor := NewExtractor()
	data, err := extractor.ExtractFromFile()
	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("failed to extract configured sessions", err)
	}
	logging.Debug.Log("✅ GetConfiguredSessions completed", "count", len(data.SsoSessions))
	return data.SsoSessions, nil
}

// GetFilterOptions returns all available filter values from current AWS CLI profiles.
//
// This function helps users build filter UIs by providing lists of all unique values
// found across their profiles. Useful for populating dropdown menus and showing
// what filter options are actually available.
//
// Process:
//  1. Extract all profiles and SSO sessions from ~/.aws/config
//  2. Collect unique values for each filterable field
//  3. Sort values alphabetically for consistent display
//
// Returns:
//   - *FilterOptionsResult: All unique values for account IDs, roles, regions, and sessions
//   - error: Any error encountered during extraction
//
// Example:
//
//	options, err := awscli.GetFilterOptions()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Available accounts:", options.AccountIDs)
//	fmt.Println("Available roles:", options.RoleNames)
//	fmt.Println("Available regions:", options.Regions)
func GetFilterOptions() (*FilterOptionsResult, error) {
	logging.Debug.Log("GetFilterOptions API called")

	// Extract profiles and sessions
	logging.Debug.Log("\t🔹 Extracting AWS CLI data")
	extractor := NewExtractor()
	data, err := extractor.ExtractFromFile()
	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("failed to extract AWS CLI data", err)
	}

	logging.Debug.Log("\t🔹 Extraction complete",
		"profiles", len(data.Profiles),
		"sso_sessions", len(data.SsoSessions))

	// Get available filter options
	logging.Debug.Log("\t🔹 Extracting filter options")
	filter := NewFilter()
	options := filter.GetAvailableFilterOptions(data.Profiles, data.SsoSessions)

	result := &FilterOptionsResult{
		AccountIDs:  options.AccountIDs,
		RoleNames:   options.RoleNames,
		Regions:     options.Regions,
		SsoSessions: options.SsoSessions,
	}

	logging.Debug.Log("✅ GetFilterOptions completed",
		"account_ids", len(result.AccountIDs),
		"roles", len(result.RoleNames),
		"regions", len(result.Regions),
		"sso_sessions", len(result.SsoSessions))

	return result, nil
}

// =============================================================================
// ADVANCED API - Direct Component Access
// =============================================================================
// These functions provide direct access to internal components for advanced use cases.
// Most users should use the high-level API functions above. Use these only when you
// need fine-grained control over individual operations.

// NewSessionManagerDefault creates a SessionManager with default AWS config path.
//
// Use this when you need direct access to SessionManager operations beyond what
// GetSessionStatus() provides, such as:
//   - LoginToSession() - Initiate AWS SSO login for a specific session
//   - Logout() - Perform global SSO logout
//   - CheckSessionByCLI() - Check specific session via AWS CLI command
//   - ClearExpiredCache() - Clean up expired SSO cache files
//
// Returns:
//   - *SessionManager: Configured session manager using default config path
func NewSessionManagerDefault() *SessionManager {
	logging.Debug.Log("NewSessionManagerDefault called")
	extractor := NewExtractor()
	return NewSessionManager(extractor)
}

// NewSessionManagerWithExtractor creates a SessionManager with a custom extractor.
//
// Use this when you need to specify a custom AWS config file location or provide
// a pre-configured extractor for testing.
//
// Parameters:
//   - extractor: Custom extractor configured with specific config path or test data
//
// Returns:
//   - *SessionManager: Configured session manager using the provided extractor
func NewSessionManagerWithExtractor(extractor *Extractor) *SessionManager {
	logging.Debug.Log("NewSessionManagerWithExtractor called")
	return NewSessionManager(extractor)
}

// NewExtractorDefault creates an Extractor with default AWS config path.
//
// Use this when you need direct access to extraction logic, such as:
//   - ExtractFromFile() - Read and parse AWS CLI config
//   - Custom extraction workflows
//   - Testing extraction logic
//
// Returns:
//   - *Extractor: Configured extractor using default config path (~/.aws/config)
func NewExtractorDefault() *Extractor {
	logging.Debug.Log("NewExtractorDefault called")
	return NewExtractor()
}

// NewFilterDefault creates a Filter instance for profile filtering operations.
//
// Use this when you need direct access to filtering logic, such as:
//   - FilterProfiles() - Apply filter criteria to profile list
//   - GetAvailableFilterOptions() - Extract unique filter values
//   - Custom filtering workflows
//
// Returns:
//   - *Filter: Filter instance ready to use
func NewFilterDefault() *Filter {
	logging.Debug.Log("NewFilterDefault called")
	return NewFilter()
}
