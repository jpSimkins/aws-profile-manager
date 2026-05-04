package awscli

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// Extractor handles parsing of AWS CLI configuration files.
//
// The Extractor reads and parses AWS CLI config files (~/.aws/config), extracting
// profile and SSO session configurations. It supports both modern SSO profiles
// (using sso-session) and legacy SSO profiles (using sso_start_url directly).
//
// Parsing Features:
//   - Profile extraction: [profile name] sections
//   - SSO session extraction: [sso-session name] sections
//   - Metadata extraction: Special comments (# Organization:, # Account:, etc.)
//   - Profile type detection: SSO, IAM, AssumeRole, Unknown
//   - Property preservation: All key=value pairs stored
//
// The extractor delegates to specialized helpers for field classification:
//   - extractSsoFields: SSO-specific properties
//   - extractIamFields: IAM and AssumeRole properties
//   - extractProfileFields: Common profile properties (region, output, etc.)
type Extractor struct {
	configPath string // Path to AWS CLI config file
}

// NewExtractor creates a new AWS CLI config extractor with default config path.
//
// The default path is determined by settings.GetAwsDir(), which respects the
// AWS_PROFILE_MANAGER_AWS_DIR environment variable or defaults to ~/.aws/.
//
// Returns:
//   - *Extractor: Configured extractor using default config path (~/.aws/config)
func NewExtractor() *Extractor {
	// Use settings helper to get AWS directory (respects environment variable)
	awsDir := settings.GetAwsDir()
	configPath := filepath.Join(awsDir, "config")

	logging.Debug.Logf("AWS CLI extractor created : %s", configPath)

	return &Extractor{
		configPath: configPath,
	}
}

// NewExtractorWithPath creates a new extractor with a custom config file path.
//
// Use this when you need to parse a config file at a non-standard location,
// such as during testing or when processing multiple config files.
//
// Parameters:
//   - configPath: Absolute path to AWS CLI config file
//
// Returns:
//   - *Extractor: Configured extractor using the specified config path
func NewExtractorWithPath(configPath string) *Extractor {
	logging.Debug.Logf("AWS CLI extractor created with custom path: %s", configPath)
	return &Extractor{
		configPath: configPath,
	}
}

// ExtractFromFile extracts AWS CLI profiles and SSO sessions from the configured file.
//
// This is the main method for extracting data from the extractor's configured
// config file path.
//
// Returns:
//   - *ExtractedData: Parsed profiles, sessions, and metadata
//   - error: Any error encountered during file reading or parsing
func (e *Extractor) ExtractFromFile() (*ExtractedData, error) {
	return e.ExtractFromPath(e.configPath)
}

// ExtractFromPath extracts AWS CLI profiles and SSO sessions from a specific file path.
//
// This method allows extracting from a custom path regardless of the extractor's
// configured path. Useful for one-off extractions or testing.
//
// Process:
//  1. Open config file
//  2. Parse line by line, detecting sections ([profile ...], [sso-session ...])
//  3. Extract properties for each section
//  4. Parse metadata comments (# Organization:, # Account:)
//  5. Classify profile types (SSO, IAM, AssumeRole)
//  6. Return complete ExtractedData
//
// Parameters:
//   - configPath: Absolute path to AWS CLI config file
//
// Returns:
//   - *ExtractedData: Parsed profiles, sessions, and metadata
//   - error: Any error encountered during file reading or parsing
func (e *Extractor) ExtractFromPath(configPath string) (*ExtractedData, error) {
	logging.Debug.Logf("Extracting AWS CLI configuration from file: %s", configPath)

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, logging.Log.Errorf("AWS CLI config file not found: %s", configPath)
		}
		return nil, logging.Log.ErrorWithDetails("Failed to open AWS CLI config file", err)
	}
	defer file.Close()

	data := &ExtractedData{
		Profiles:    []AwsCliProfile{},
		SsoSessions: []SsoSession{},
		ExtractedAt: time.Now(),
		SourceFile:  configPath,
	}

	if err := e.parseConfigFile(file, data); err != nil {
		return nil, err
	}

	logging.Debug.Log("AWS CLI configuration extracted successfully")
	logging.Debug.Logf("\t🔹 Profiles Found: %d", len(data.Profiles))
	logging.Debug.Logf("\t🔹 SSO Sessions Found: %d", len(data.SsoSessions))
	logging.Debug.Logf("\t🔹 Source File: %s", configPath)

	return data, nil
}

// parseConfigFile reads and parses the AWS CLI config file line by line.
//
// This method implements the core parsing logic using a state machine approach:
//   - Tracks current section type (profile or sso-session)
//   - Accumulates properties for each section
//   - Detects section boundaries and saves completed sections
//   - Parses metadata comments
//
// Parsing Rules:
//   - Lines starting with [profile ...] begin a profile section
//   - Lines starting with [sso-session ...] begin an SSO session section
//   - Lines with = are property assignments
//   - Lines starting with # may contain metadata
//   - Empty/whitespace-only lines are ignored
//
// Parameters:
//   - file: Open file handle to AWS CLI config
//   - data: ExtractedData struct to populate with parsed data
//
// Returns:
//   - error: Any error encountered during parsing
func (e *Extractor) parseConfigFile(file *os.File, data *ExtractedData) error {
	logging.Debug.Log("Parsing AWS CLI config file")

	scanner := bufio.NewScanner(file)
	var currentSection string
	var currentProfile *AwsCliProfile
	var currentSession *SsoSession
	var pendingMetadata map[string]string // Store metadata from comments before next section

	// Regex patterns for parsing sections
	profileRegex := regexp.MustCompile(`^\[profile\s+([^\]]+)\]$`)
	defaultRegex := regexp.MustCompile(`^\[default\]$`)
	sessionRegex := regexp.MustCompile(`^\[sso-session\s+([^\]]+)\]$`)
	propertyRegex := regexp.MustCompile(`^([a-z_]+)\s*=\s*(.*)$`)
	metadataRegex := regexp.MustCompile(`^#\s*([A-Za-z]+):\s*(.+)$`) // Matches "# Key: Value"

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for metadata comments (# Key: Value)
		if strings.HasPrefix(line, "#") {
			if match := metadataRegex.FindStringSubmatch(line); match != nil {
				if pendingMetadata == nil {
					pendingMetadata = make(map[string]string)
				}
				key := strings.TrimSpace(match[1])
				value := strings.TrimSpace(match[2])
				pendingMetadata[key] = value
				logging.Debug.Logf("Found metadata comment: %s = %s", key, value)
			}
			continue
		}

		// Check for [profile name] section
		if match := profileRegex.FindStringSubmatch(line); match != nil {
			e.saveCurrentProfile(currentProfile, data)
			e.saveCurrentSession(currentSession, data)

			currentProfile = &AwsCliProfile{
				Name:       match[1],
				Type:       ProfileTypeUnknown,
				Properties: make(map[string]string),
			}

			// Apply pending metadata to profile
			if pendingMetadata != nil {
				if orgName, ok := pendingMetadata["Organization"]; ok {
					currentProfile.OrganizationName = orgName
				}
				if acctName, ok := pendingMetadata["Account"]; ok {
					currentProfile.AccountName = acctName
				}
				pendingMetadata = nil // Clear after applying
			}

			currentSession = nil
			currentSection = "profile"
			logging.Debug.Logf("Found profile section: %s", match[1])
			continue
		}

		// Check for [default] section
		if defaultRegex.MatchString(line) {
			e.saveCurrentProfile(currentProfile, data)
			e.saveCurrentSession(currentSession, data)

			currentProfile = &AwsCliProfile{
				Name:       "default",
				Type:       ProfileTypeUnknown,
				Properties: make(map[string]string),
			}

			// Apply pending metadata to profile
			if pendingMetadata != nil {
				if orgName, ok := pendingMetadata["Organization"]; ok {
					currentProfile.OrganizationName = orgName
				}
				if acctName, ok := pendingMetadata["Account"]; ok {
					currentProfile.AccountName = acctName
				}
				pendingMetadata = nil // Clear after applying
			}

			currentSession = nil
			currentSection = "profile"
			logging.Debug.Log("Found default profile section")
			continue
		}

		// Check for [sso-session name] section
		if match := sessionRegex.FindStringSubmatch(line); match != nil {
			e.saveCurrentProfile(currentProfile, data)
			e.saveCurrentSession(currentSession, data)

			currentSession = &SsoSession{
				Name:       match[1],
				Properties: make(map[string]string),
			}

			// Apply pending metadata to session
			if pendingMetadata != nil {
				if orgName, ok := pendingMetadata["Organization"]; ok {
					currentSession.OrganizationName = orgName
				}
				if desc, ok := pendingMetadata["Description"]; ok {
					currentSession.Description = desc
				}
				pendingMetadata = nil // Clear after applying
			}

			currentProfile = nil
			currentSection = "sso-session"
			logging.Debug.Logf("\t🔹 Found SSO session section: %s", match[1])
			continue
		}

		// Check for other sections (ignore them but reset current objects)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			e.saveCurrentProfile(currentProfile, data)
			e.saveCurrentSession(currentSession, data)
			currentProfile = nil
			currentSession = nil
			currentSection = "other"
			pendingMetadata = nil // Clear pending metadata for unknown sections
			logging.Debug.Logf("\t🔹 Found other section, ignoring: %s", line)
			continue
		}

		// Parse properties
		if match := propertyRegex.FindStringSubmatch(line); match != nil {
			key := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])

			if currentSection == "profile" && currentProfile != nil {
				currentProfile.Properties[key] = value
				e.extractProfileFields(currentProfile, key, value)
			} else if currentSection == "sso-session" && currentSession != nil {
				currentSession.Properties[key] = value
				e.extractSessionFields(currentSession, key, value)
			}
		}
	}

	// Save the last profile/session
	e.saveCurrentProfile(currentProfile, data)
	e.saveCurrentSession(currentSession, data)

	if err := scanner.Err(); err != nil {
		return logging.Log.ErrorWithDetails("Error reading AWS CLI config file", err)
	}

	return nil
}

// extractProfileFields extracts and classifies profile properties.
//
// This method handles common profile fields (like region) and delegates to
// specialized extractors for type-specific fields (SSO, IAM, AssumeRole).
//
// Delegation Strategy:
//   - Common fields: Handled directly (region, output)
//   - SSO fields: Delegated to extractSsoFields
//   - IAM/AssumeRole fields: Delegated to extractIamFields
//
// Parameters:
//   - profile: Profile to update with extracted fields
//   - key: Config property name
//   - value: Config property value
func (e *Extractor) extractProfileFields(profile *AwsCliProfile, key, value string) {
	// Handle common fields first
	switch key {
	case "region":
		profile.Region = value
		logging.Debug.Logf("\t🔹 Extracted region: %s", value)
		return
	}

	// Delegate to type-specific extractors
	e.extractSsoFields(profile, key, value)
	e.extractIamFields(profile, key, value)
}

// determineProfileType finalizes the profile type classification.
//
// This method is called after all properties have been extracted for a profile.
// If the profile type is still Unknown after property extraction, it remains Unknown.
//
// Parameters:
//   - profile: Profile whose type to finalize
func (e *Extractor) determineProfileType(profile *AwsCliProfile) {
	if profile.Type == ProfileTypeUnknown {
		// If no specific type was detected, it's likely a basic profile
		logging.Debug.Logf("\t🔹 Profile type remains unknown: %s", profile.Name)
	}
}

// extractSessionFields extracts SSO session properties.
//
// This method processes SSO session configuration properties and populates
// the SsoSession struct with the extracted values.
//
// SSO Session Properties:
//   - sso_start_url: SSO portal URL
//   - sso_region: AWS region for SSO service
//   - sso_registration_scopes: OAuth scopes for registration
//
// Parameters:
//   - session: SSO session to update with extracted fields
//   - key: Config property name
//   - value: Config property value
func (e *Extractor) extractSessionFields(session *SsoSession, key, value string) {
	switch key {
	case "sso_start_url":
		session.StartURL = value
	case "sso_region":
		session.Region = value
	case "sso_registration_scopes":
		session.RegistrationScopes = value
	}
}

// saveCurrentProfile saves a completed profile to the ExtractedData.
//
// This method is called when a section boundary is detected or end of file
// is reached. It finalizes the profile type and appends it to the result.
//
// Parameters:
//   - profile: Profile to save (nil-safe)
//   - data: ExtractedData to append profile to
func (e *Extractor) saveCurrentProfile(profile *AwsCliProfile, data *ExtractedData) {
	if profile != nil {
		// Finalize profile type determination
		e.determineProfileType(profile)

		data.Profiles = append(data.Profiles, *profile)
	}
}

// saveCurrentSession saves a completed SSO session to the ExtractedData.
//
// This method is called when a section boundary is detected or end of file
// is reached.
//
// Parameters:
//   - session: SSO session to save (nil-safe)
//   - data: ExtractedData to append session to
func (e *Extractor) saveCurrentSession(session *SsoSession, data *ExtractedData) {
	if session != nil {
		data.SsoSessions = append(data.SsoSessions, *session)
	}
}

// GetConfigPath returns the configured AWS CLI config file path.
//
// Returns:
//   - string: Absolute path to AWS CLI config file
func (e *Extractor) GetConfigPath() string {
	return e.configPath
}

// ValidateConfigFile checks if the AWS CLI config file exists and is readable.
//
// This method performs basic validation without parsing the file contents.
// Useful for pre-flight checks before attempting extraction.
//
// Returns:
//   - error: Error if file doesn't exist, isn't readable, or is a directory
func (e *Extractor) ValidateConfigFile() error {
	fileInfo, err := os.Stat(e.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return logging.Log.Errorf("AWS CLI config file does not exist: %s", e.configPath)
		}
		return logging.Log.ErrorWithDetails("Cannot access AWS CLI config file", err)
	}

	if fileInfo.IsDir() {
		return logging.Log.Errorf("AWS CLI config path is a directory, not a file: %s", e.configPath)
	}

	// Try to open for reading
	file, err := os.Open(e.configPath)
	if err != nil {
		return logging.Log.ErrorWithDetails("Cannot read AWS CLI config file", err)
	}
	file.Close()

	logging.Debug.Logf("AWS CLI config file validation passed: %s", e.configPath)
	return nil
}

// GetFileModTime returns the modification time of the AWS CLI config file.
//
// This method is used by the caching system to detect when the config file
// has been modified and the cache needs to be invalidated.
//
// Returns:
//   - time.Time: File modification time
//   - error: Error if file cannot be accessed
func (e *Extractor) GetFileModTime() (time.Time, error) {
	fileInfo, err := os.Stat(e.configPath)
	if err != nil {
		return time.Time{}, err
	}
	return fileInfo.ModTime(), nil
}
