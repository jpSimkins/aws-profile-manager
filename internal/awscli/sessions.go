package awscli

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// SessionManager manages AWS SSO session information and status checking.
//
// The SessionManager provides functionality for checking SSO session validity,
// managing SSO authentication, and cleaning up expired cache files. It works
// by reading SSO cache files stored by AWS CLI and checking token expiration.
//
// Session Status Checking:
//   - Reads ~/.aws/sso/cache/ for token files
//   - Matches cache files to SSO sessions by start URL
//   - Checks token expiration times
//   - Categorizes sessions as active or expired
//
// Caching Strategy:
//   - Status results are cached for 30 seconds
//   - Prevents excessive file system reads
//   - Automatic cache invalidation after 30s
//
// Additional Features:
//   - AWS CLI availability detection
//   - SSO login/logout operations
//   - Expired cache file cleanup
type SessionManager struct {
	cacheDir     string        // Path to AWS SSO cache directory
	extractor    *Extractor    // Extractor for reading AWS config
	lastChecked  time.Time     // Timestamp of last status check
	cachedStatus SessionStatus // Cached session status (valid for 30s)
}

// NewSessionManager creates a new AWS session manager with default SSO cache directory.
//
// The default cache directory is determined by settings.GetAwsDir(), which respects
// the AWS_PROFILE_MANAGER_AWS_DIR environment variable or defaults to ~/.aws/.
//
// Parameters:
//   - extractor: Extractor for reading AWS CLI config
//
// Returns:
//   - *SessionManager: Configured session manager with default cache directory
func NewSessionManager(extractor *Extractor) *SessionManager {
	// Use settings helper to get AWS directory (respects environment variable)
	awsDir := settings.GetAwsDir()
	cacheDir := filepath.Join(awsDir, "sso", "cache")

	logging.Debug.Logf("Session manager created with cache directory: %s", cacheDir)

	return &SessionManager{
		cacheDir:  cacheDir,
		extractor: extractor,
	}
}

// NewSessionManagerWithPath creates a new AWS session manager with custom cache directory.
//
// Use this when you need to specify a non-standard SSO cache directory, such as
// during testing or when processing multiple AWS configurations.
//
// Parameters:
//   - cacheDir: Path to SSO cache directory
//   - extractor: Extractor for reading AWS CLI config
//
// Returns:
//   - *SessionManager: Configured session manager with custom cache directory
func NewSessionManagerWithPath(cacheDir string, extractor *Extractor) *SessionManager {
	logging.Debug.Logf("Session manager created with custom cache directory: %s", cacheDir)
	return &SessionManager{
		cacheDir:  cacheDir,
		extractor: extractor,
	}
}

// GetSessionStatus returns the current status of all AWS SSO sessions.
//
// This method extracts SSO sessions from AWS CLI config and checks their status
// against SSO cache files. Results are cached for 30 seconds to avoid excessive
// file system reads.
//
// Process:
//  1. Check cache: Return cached status if less than 30 seconds old
//  2. Extract sessions: Read SSO sessions from AWS CLI config
//  3. Check cache files: Match sessions to cache files by start URL
//  4. Verify tokens: Check expiration times
//  5. Categorize: Separate active and expired sessions
//  6. Cache results: Store for subsequent calls
//
// Returns:
//   - SessionStatus: Current session status with active/expired sessions
//   - error: Any error encountered during extraction or cache reading
func (sm *SessionManager) GetSessionStatus() (SessionStatus, error) {
	logging.Debug.Log("Getting AWS SSO session status")

	// Check if we need to refresh the cache (refresh every 30 seconds)
	if time.Since(sm.lastChecked) < 30*time.Second && !sm.cachedStatus.LastChecked.IsZero() {
		logging.Debug.Logf("Using cached session status (cache age: %v)", time.Since(sm.lastChecked))
		return sm.cachedStatus, nil
	}

	status := SessionStatus{
		ActiveSessions:  []ActiveSessionInfo{},
		ExpiredSessions: []ActiveSessionInfo{},
		LastChecked:     time.Now(),
		CLIAvailable:    sm.isAWSCLIAvailable(),
	}

	// Get CLI version if available
	if status.CLIAvailable {
		status.CLIVersion = sm.getAWSCLIVersion()
		logging.Debug.Logf("AWS CLI detected (version: %s)", status.CLIVersion)
	} else {
		logging.Log.Warn("AWS CLI not available")
	}

	// Extract SSO sessions from AWS CLI config
	var ssoSessions []SsoSession
	if sm.extractor != nil {
		data, err := sm.extractor.ExtractFromFile()
		if err != nil {
			logging.Log.Warnf("Failed to extract AWS CLI config for session status: %v", err)
			// Continue with empty sessions rather than failing completely
		} else {
			ssoSessions = data.SsoSessions
			logging.Debug.Logf("Extracted AWS CLI config for session analysis (%d sessions)", len(data.SsoSessions))
		}
	} else {
		logging.Debug.Log("No extractor available, will scan cache files directly")
		// When no extractor is available, we'll scan all cache files and try to determine sessions
		cacheFiles, err := sm.findAllCacheFiles()
		if err != nil {
			logging.Log.Warnf("Failed to scan cache files: %v", err)
		} else {
			// Create synthetic sessions from cache files
			ssoSessions = sm.createSessionsFromCacheFiles(cacheFiles)
		}
	}

	// Check cache for each SSO session
	for _, session := range ssoSessions {
		sessionInfo, err := sm.getSessionInfoFromCache(session)
		if err != nil {
			logging.Debug.Logf("Failed to get session info from cache for session %s: %v", session.Name, err)
			continue
		}

		if sessionInfo.IsExpired {
			status.ExpiredSessions = append(status.ExpiredSessions, sessionInfo)
		} else {
			status.ActiveSessions = append(status.ActiveSessions, sessionInfo)
		}
	}

	// Sort sessions by name for consistent ordering
	sort.Slice(status.ActiveSessions, func(i, j int) bool {
		return status.ActiveSessions[i].SessionName < status.ActiveSessions[j].SessionName
	})
	sort.Slice(status.ExpiredSessions, func(i, j int) bool {
		return status.ExpiredSessions[i].SessionName < status.ExpiredSessions[j].SessionName
	})

	logging.Debug.Log("AWS SSO session status")
	logging.Debug.Logf("\t🔹 Active Sessions: %d", len(status.ActiveSessions))
	logging.Debug.Logf("\t🔹 Expired Sessions: %d", len(status.ExpiredSessions))
	logging.Debug.Logf("\t🔹 AWS CLI Available: %t", status.CLIAvailable)

	sm.cachedStatus = status
	sm.lastChecked = time.Now()
	return status, nil
}

// GetSessionStatusWithSessions returns session status using pre-extracted SSO sessions.
//
// This is an optimization for when SSO sessions have already been extracted as part
// of another operation (e.g., ListProfiles). It avoids re-reading and parsing the
// AWS CLI config file.
//
// Use Cases:
//   - When sessions are already available from profile extraction
//   - When you want to check status multiple times with same sessions
//   - Performance optimization to avoid duplicate config file reads
//
// Parameters:
//   - ssoSessions: Pre-extracted SSO sessions to check status for
//
// Returns:
//   - SessionStatus: Current session status with active/expired sessions
//   - error: Any error encountered during cache reading
func (sm *SessionManager) GetSessionStatusWithSessions(ssoSessions []SsoSession) (SessionStatus, error) {
	logging.Debug.Logf("Getting AWS SSO session status with %d pre-extracted sessions", len(ssoSessions))

	// Check if we need to refresh the cache (refresh every 30 seconds)
	if time.Since(sm.lastChecked) < 30*time.Second && !sm.cachedStatus.LastChecked.IsZero() {
		logging.Debug.Logf("Using cached session status (cache age: %v)", time.Since(sm.lastChecked))
		return sm.cachedStatus, nil
	}

	status := SessionStatus{
		ActiveSessions:  []ActiveSessionInfo{},
		ExpiredSessions: []ActiveSessionInfo{},
		LastChecked:     time.Now(),
		CLIAvailable:    sm.isAWSCLIAvailable(),
	}

	// Get CLI version if available
	if status.CLIAvailable {
		status.CLIVersion = sm.getAWSCLIVersion()
		logging.Debug.Logf("AWS CLI detected (version: %s)", status.CLIVersion)
	} else {
		logging.Log.Warn("AWS CLI not available")
	}

	// Check cache for each SSO session (using provided sessions)
	for _, session := range ssoSessions {
		sessionInfo, err := sm.getSessionInfoFromCache(session)
		if err != nil {
			logging.Debug.Logf("Failed to get session info from cache for session %s: %v", session.Name, err)
			continue
		}

		if sessionInfo.IsExpired {
			status.ExpiredSessions = append(status.ExpiredSessions, sessionInfo)
		} else {
			status.ActiveSessions = append(status.ActiveSessions, sessionInfo)
		}
	}

	// Sort sessions by name for consistent ordering
	sort.Slice(status.ActiveSessions, func(i, j int) bool {
		return status.ActiveSessions[i].SessionName < status.ActiveSessions[j].SessionName
	})
	sort.Slice(status.ExpiredSessions, func(i, j int) bool {
		return status.ExpiredSessions[i].SessionName < status.ExpiredSessions[j].SessionName
	})

	logging.Debug.Log("AWS SSO session status (with pre-extracted sessions)",
		"active_sessions", len(status.ActiveSessions),
		"expired_sessions", len(status.ExpiredSessions),
		"cli_available", status.CLIAvailable)

	sm.cachedStatus = status
	sm.lastChecked = time.Now()
	return status, nil
}

// RefreshSessionStatus forces a refresh of session status, bypassing cache.
//
// This method clears the cached status and forces a fresh check of all SSO
// sessions. Use when you need up-to-date status regardless of cache age.
//
// Returns:
//   - SessionStatus: Freshly checked session status
//   - error: Any error encountered during status check
func (sm *SessionManager) RefreshSessionStatus() (SessionStatus, error) {
	logging.Log.Info("Force refreshing AWS SSO session status")
	sm.lastChecked = time.Time{} // Reset cache
	return sm.GetSessionStatus()
}

// CheckSessionByCLI checks if a specific SSO session is active using AWS CLI.
//
// This method invokes the AWS CLI to check if a session has valid cached credentials.
// Requires AWS CLI to be installed and available in PATH.
//
// Parameters:
//   - sessionName: Name of the SSO session to check
//
// Returns:
//   - bool: true if session is active, false if expired or not found
//   - error: Error if AWS CLI is not available or command fails
func (sm *SessionManager) CheckSessionByCLI(sessionName string) (bool, error) {
	logging.Debug.Logf("Checking SSO session via AWS CLI: %s", sessionName)

	if !sm.isAWSCLIAvailable() {
		return false, logging.Log.Error("AWS CLI is not available for session check")
	}

	// Try to list SSO session accounts
	cmd := exec.Command("aws", "sso", "list-accounts", "--sso-session", sessionName)
	cmd.Env = append(os.Environ(), "AWS_PAGER=")

	err := cmd.Run()
	isActive := err == nil

	logging.Debug.Logf("SSO session CLI check completed for %s (active: %v)", sessionName, isActive)
	return isActive, nil
}

// LoginToSession attempts to initiate AWS SSO login for a specific session.
//
// This method invokes the AWS CLI to start the SSO login process, which typically
// opens a browser for authentication. Requires AWS CLI to be installed.
//
// Parameters:
//   - sessionName: Name of the SSO session to login to
//
// Returns:
//   - error: Error if AWS CLI is not available or login fails
func (sm *SessionManager) LoginToSession(sessionName string) error {
	logging.Log.Infof("Attempting to login to SSO session: %s", sessionName)

	if !sm.isAWSCLIAvailable() {
		return logging.Log.Error("AWS CLI is not available for login")
	}

	cmd := exec.Command("aws", "sso", "login", "--sso-session", sessionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		return logging.Log.ErrorWithDetails("SSO login failed", err)
	}

	logging.Log.Successf("SSO session login completed: %s", sessionName)
	// Force refresh of cached status after login
	sm.lastChecked = time.Time{}
	return nil
}

// Logout attempts to logout from all SSO sessions.
//
// This method performs a global SSO logout using AWS CLI. Note that AWS CLI
// does not support per-session logout - this logs out of ALL sessions.
//
// Side Effects:
//   - Clears all SSO session cache files
//   - Invalidates all active SSO tokens
//   - Requires re-authentication for all sessions
//
// Returns:
//   - error: Error if AWS CLI is not available or logout fails
func (sm *SessionManager) Logout() error {
	logging.Log.Info("Attempting to logout from all SSO sessions")

	if !sm.isAWSCLIAvailable() {
		return logging.Log.Error("AWS CLI is not available for logout")
	}

	// AWS SSO CLI only supports global logout with `aws sso logout`
	// There is no per-session logout functionality
	cmd := exec.Command("aws", "sso", "logout")
	cmd.Env = append(os.Environ(), "AWS_PAGER=")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return logging.Log.ErrorWithDetails("SSO logout failed", err)
	}

	logging.Log.Successf("SSO logout succeeded: %s", string(output))

	// Force refresh of cached status after logout
	sm.lastChecked = time.Time{}
	return nil
}

// GetCacheDir returns the configured SSO cache directory path.
//
// Returns:
//   - string: Path to SSO cache directory
func (sm *SessionManager) GetCacheDir() string {
	return sm.cacheDir
}

// isAWSCLIAvailable checks if AWS CLI is installed and available in PATH.
//
// Returns:
//   - bool: true if AWS CLI is available, false otherwise
func (sm *SessionManager) isAWSCLIAvailable() bool {
	_, err := exec.LookPath("aws")
	return err == nil
}

// getAWSCLIVersion retrieves the AWS CLI version string.
//
// Returns:
//   - string: AWS CLI version (e.g., "2.15.0") or "unknown" if unavailable
func (sm *SessionManager) getAWSCLIVersion() string {
	cmd := exec.Command("aws", "--version")
	output, err := cmd.Output()
	if err != nil {
		logging.Debug.Logf("Failed to get AWS CLI version: %v", err)
		return "unknown"
	}

	// Parse version from output like "aws-cli/2.x.x Python/3.x.x..."
	versionStr := string(output)
	if parts := strings.Fields(versionStr); len(parts) > 0 {
		if versionParts := strings.Split(parts[0], "/"); len(versionParts) > 1 {
			return versionParts[1]
		}
	}

	logging.Debug.Logf("Could not parse AWS CLI version: %s", versionStr)
	return "unknown"
}

// getSessionInfoFromCache reads session information from SSO cache files.
//
// This method searches for cache files matching the session's start URL and region,
// then reads the most recent valid cache file to determine session status.
//
// Process:
//  1. Check if SSO cache directory exists
//  2. Find cache files that might match the session
//  3. Read each cache file and validate it matches the session
//  4. Select the most recent valid cache file
//  5. Check token expiration
//  6. Return session information with validity status
//
// Parameters:
//   - session: SSO session configuration to check
//
// Returns:
//   - ActiveSessionInfo: Session information with token expiration status
//   - error: Any error encountered reading cache files
func (sm *SessionManager) getSessionInfoFromCache(session SsoSession) (ActiveSessionInfo, error) {
	sessionInfo := ActiveSessionInfo{
		SessionName: session.Name,
		StartURL:    session.StartURL,
		Region:      session.Region,
		IsExpired:   true, // Default to expired
	}

	// Check if cache directory exists
	if _, err := os.Stat(sm.cacheDir); os.IsNotExist(err) {
		logging.Debug.Log("SSO cache directory does not exist",
			"cache_dir", sm.cacheDir)
		return sessionInfo, fmt.Errorf("SSO cache directory does not exist: %s", sm.cacheDir)
	}

	// Find cache files that might belong to this session
	cacheFiles, err := sm.findCacheFilesForSession(session)
	if err != nil {
		return sessionInfo, err
	}

	// Find the most recent valid cache file
	var latestCache *SsoCacheFile
	var latestCacheFile string
	var latestModTime time.Time

	for _, cacheFile := range cacheFiles {
		cache, err := sm.readCacheFile(cacheFile)
		if err != nil {
			logging.Debug.Logf("Failed to read cache file %s: %v", cacheFile, err)
			continue
		}

		// Check if this cache file matches our session
		if cache.StartURL != session.StartURL || cache.Region != session.Region {
			continue
		}

		// Get file modification time
		fileInfo, err := os.Stat(cacheFile)
		if err != nil {
			continue
		}

		// Use the most recently modified cache file
		if latestCache == nil || fileInfo.ModTime().After(latestModTime) {
			latestCache = cache
			latestCacheFile = cacheFile
			latestModTime = fileInfo.ModTime()
		}
	}

	if latestCache == nil {
		// No cache file found - this is normal if user hasn't logged in yet
		logging.Debug.Log("No valid cache file found for session",
			"session", session.Name,
			"start_url", session.StartURL)
		return sessionInfo, fmt.Errorf("no valid cache file found for session %s", session.Name)
	}

	// Update session info with cache data
	sessionInfo.AccessToken = latestCache.AccessToken
	sessionInfo.ExpiresAt = latestCache.ExpiresAt
	sessionInfo.IsExpired = time.Now().After(latestCache.ExpiresAt)
	sessionInfo.CacheFilePath = latestCacheFile

	logging.Debug.Log("Session info loaded from cache",
		"session", session.Name,
		"expires_at", latestCache.ExpiresAt,
		"is_expired", sessionInfo.IsExpired,
		"cache_file", latestCacheFile)

	return sessionInfo, nil
}

// findCacheFilesForSession finds all JSON cache files in the SSO cache directory.
//
// This method walks through the SSO cache directory and collects all JSON files
// that might contain session tokens. The caller is responsible for validating
// which cache files actually match the target session.
//
// Parameters:
//   - session: SSO session to find cache files for
//
// Returns:
//   - []string: List of cache file paths
//   - error: Any error encountered walking the directory
func (sm *SessionManager) findCacheFilesForSession(session SsoSession) ([]string, error) {
	var cacheFiles []string

	err := filepath.WalkDir(sm.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Only look at JSON files
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			cacheFiles = append(cacheFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("error walking cache directory", err)
	}

	logging.Debug.Logf("Found cache files for session %s: %d files", session.Name, len(cacheFiles))
	return cacheFiles, nil
}

// readCacheFile reads and parses an SSO cache file.
//
// This method reads a JSON cache file from the SSO cache directory and
// deserializes it into an SsoCacheFile struct.
//
// Parameters:
//   - filePath: Absolute path to cache file
//
// Returns:
//   - *SsoCacheFile: Parsed cache file contents
//   - error: Any error encountered reading or parsing the file
func (sm *SessionManager) readCacheFile(filePath string) (*SsoCacheFile, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, logging.Log.ErrorfWithDetails(fmt.Sprintf("failed to read cache file %s", filePath), err)
	}

	var cache SsoCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, logging.Log.ErrorfWithDetails(fmt.Sprintf("failed to parse cache file %s", filePath), err)
	}

	return &cache, nil
}

// ClearExpiredCache removes expired cache files
func (sm *SessionManager) ClearExpiredCache() error {
	logging.Log.Info("Clearing expired SSO cache files")

	if _, err := os.Stat(sm.cacheDir); os.IsNotExist(err) {
		logging.Debug.Log("No SSO cache directory exists, nothing to clear")
		return nil // No cache directory, nothing to clear
	}

	// Get all cache files
	cacheFiles, err := sm.findAllCacheFiles()
	if err != nil {
		return logging.Log.ErrorWithDetails("Failed to find cache files for cleanup", err)
	}

	now := time.Now()
	removedCount := 0

	for _, cacheFile := range cacheFiles {
		cache, err := sm.readCacheFile(cacheFile)
		if err != nil {
			logging.Debug.Logf("Skipping cache file %s due to read error: %v", cacheFile, err)
			continue // Skip files we can't read
		}

		// Remove if expired
		if now.After(cache.ExpiresAt) {
			if err := os.Remove(cacheFile); err == nil {
				removedCount++
				logging.Debug.Logf("Removed expired cache file %s (expired at: %v)", cacheFile, cache.ExpiresAt)
			} else {
				logging.Log.Warnf("Failed to remove expired cache file %s: %v", cacheFile, err)
			}
		}
	}

	logging.Log.Successf("Expired SSO cache cleanup completed: %d files removed", removedCount)
	return nil
}

// findAllCacheFiles finds all JSON cache files in the SSO cache directory.
//
// This method walks through the SSO cache directory and collects all JSON files.
//
// Returns:
//   - []string: List of all cache file paths
//   - error: Any error encountered walking the directory
func (sm *SessionManager) findAllCacheFiles() ([]string, error) {
	var cacheFiles []string

	err := filepath.WalkDir(sm.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Only look at JSON files
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			cacheFiles = append(cacheFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, logging.Log.ErrorfWithDetails("error walking cache directory", err)
	}

	return cacheFiles, nil
}

// createSessionsFromCacheFiles creates synthetic SSO sessions from cache files.
//
// This method is used as a fallback when the extractor is unavailable. It creates
// SSO session objects by reading cache files and extracting session information,
// allowing session status checking even without access to AWS CLI config.
//
// Process:
//  1. Read each cache file
//  2. Extract start URL and region
//  3. Create synthetic session name from start URL
//  4. Deduplicate sessions by start URL + region
//  5. Return list of synthetic sessions
//
// Parameters:
//   - cacheFiles: List of cache file paths to process
//
// Returns:
//   - []SsoSession: List of synthetic SSO sessions
func (sm *SessionManager) createSessionsFromCacheFiles(cacheFiles []string) []SsoSession {
	var sessions []SsoSession
	seenSessions := make(map[string]bool)

	for _, cacheFile := range cacheFiles {
		cache, err := sm.readCacheFile(cacheFile)
		if err != nil {
			logging.Debug.Logf("Skipping cache file %s due to read error: %v", cacheFile, err)
			continue
		}

		// Create a unique key for this session (startURL + region)
		sessionKey := cache.StartURL + "|" + cache.Region
		if seenSessions[sessionKey] {
			continue // Skip duplicates
		}

		// Create synthetic session name from startURL
		sessionName := sm.extractSessionNameFromStartURL(cache.StartURL)

		session := SsoSession{
			Name:       sessionName,
			StartURL:   cache.StartURL,
			Region:     cache.Region,
			Properties: map[string]string{},
		}

		sessions = append(sessions, session)
		seenSessions[sessionKey] = true
		logging.Debug.Logf("Created synthetic SSO session from cache: %s (start URL: %s)", sessionName, cache.StartURL)
	}

	return sessions
}

// extractSessionNameFromStartURL creates a session name from the SSO start URL.
//
// This method extracts a meaningful session name from the start URL by parsing
// the domain name. Used for creating synthetic sessions from cache files.
//
// Extraction Logic:
//   - https://example.awsapps.com/start → "example"
//   - Falls back to "sso-session" if parsing fails
//
// Parameters:
//   - startURL: SSO start URL to extract name from
//
// Returns:
//   - string: Extracted session name
func (sm *SessionManager) extractSessionNameFromStartURL(startURL string) string {
	// Extract domain name from URL and use as session name
	// https://example.awsapps.com/start -> example
	if strings.HasPrefix(startURL, "https://") {
		url := strings.TrimPrefix(startURL, "https://")
		if parts := strings.Split(url, "."); len(parts) > 0 {
			return parts[0]
		}
	}

	// Fallback to a generic name
	return "sso-session"
}
