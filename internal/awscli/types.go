// Package awscli provides AWS CLI integration for profile extraction and management.
//
// This package handles reading and parsing AWS CLI configuration files, extracting
// profile information, managing caching for performance, and checking SSO session status.
//
// Key Features:
//   - Profile Extraction: Read and parse ~/.aws/config files
//   - Profile Types: Detect SSO, IAM, and AssumeRole profiles
//   - Metadata Parsing: Extract comments with organization and account information
//   - Profile Filtering: Advanced filtering by account, role, region, pattern
//   - Caching: Performance optimization with file modification detection
//   - SSO Sessions: Parse SSO session configurations and check status
//   - Session Status: Verify active/expired SSO sessions via cache files
//
// Architecture:
//   - api.go: High-level API functions (ListProfiles, GetSessionStatus)
//   - extractor.go: Core config file parsing and extraction logic
//   - extractor_sso.go: SSO profile classification
//   - extractor_iam.go: IAM profile classification
//   - filters.go: Profile filtering and criteria matching
//   - cache.go: Caching system with file modification detection
//   - sessions.go: SSO session status checking
//   - types.go: Data structures and type definitions
//
// Profile Types:
//   - SSO: Profiles using AWS SSO with sso_session configuration
//   - IAM: Profiles using static credentials or credential_process
//   - AssumeRole: Profiles that assume a role from a source profile
//   - Unknown: Profiles that don't match standard patterns
//
// Metadata Comments:
//
//	The extractor can parse special comments in AWS CLI config:
//	- # Organization: <name> - Associates profiles/sessions with organization
//	- # Account: <name> - Provides human-readable account names
//	- # Description: <text> - Describes SSO sessions
//
// Example Usage:
//
//	// List all profiles with filtering
//	result, err := awscli.ListProfiles(awscli.FilterCriteria{
//	    AccountIDs: []string{"123456789012"},
//	    RoleNames:  []string{"Administrator"},
//	})
//
//	// Check SSO session status
//	status, err := awscli.GetSessionStatus(awscli.SessionOptions{
//	    IncludeExpired: true,
//	})
//	for _, session := range status.ActiveSessions {
//	    fmt.Printf("Session %s expires at %s\n",
//	        session.SessionName, session.ExpiresAt)
//	}
package awscli

import "time"

// ProfileType represents the authentication method used by an AWS CLI profile.
//
// AWS CLI profiles can use different authentication mechanisms. This type
// helps categorize profiles for filtering and processing.
type ProfileType string

const (
	ProfileTypeSSO        ProfileType = "sso"         // Profile uses AWS SSO authentication
	ProfileTypeIAM        ProfileType = "iam"         // Profile uses IAM user credentials
	ProfileTypeAssumeRole ProfileType = "assume_role" // Profile assumes a role from source profile
	ProfileTypeUnknown    ProfileType = "unknown"     // Profile type could not be determined
)

// AwsCliProfile represents a single AWS CLI profile from ~/.aws/config.
//
// This struct contains all information extracted from a profile entry, including
// authentication details, metadata from comments, and all raw properties for
// maximum flexibility.
type AwsCliProfile struct {
	Name      string      `json:"name"`                 // Profile name (from [profile name])
	Type      ProfileType `json:"type"`                 // Detected profile type (SSO, IAM, AssumeRole, Unknown)
	AccountID string      `json:"account_id,omitempty"` // AWS account ID (from sso_account_id)
	RoleName  string      `json:"role_name,omitempty"`  // IAM role name (from sso_role_name or role_arn)
	Region    string      `json:"region,omitempty"`     // AWS region (from region property)

	// SSO-specific fields
	SsoSession  string `json:"sso_session,omitempty"`   // SSO session name reference
	SsoStartURL string `json:"sso_start_url,omitempty"` // SSO start URL (for legacy SSO profiles)

	// IAM-specific fields
	HasAccessKey      bool   `json:"has_access_key,omitempty"`         // Whether profile has aws_access_key_id
	HasSecretKey      bool   `json:"has_secret_key,omitempty"`         // Whether profile has aws_secret_access_key
	HasCredentialProc bool   `json:"has_credential_process,omitempty"` // Whether profile uses credential_process
	CredentialProcess string `json:"credential_process,omitempty"`     // credential_process command

	// Metadata from comments (organization/account annotations)
	AccountName      string `json:"account_name,omitempty"`      // Human-readable account name from # Account: comment
	OrganizationName string `json:"organization_name,omitempty"` // Organization name from # Organization: comment

	// Raw properties for complete access to all config values
	Properties map[string]string `json:"properties"` // All profile properties as key-value pairs
}

// SsoSession represents an SSO session configuration from ~/.aws/config.
//
// SSO sessions define the authentication endpoints and settings for AWS SSO.
// Multiple profiles can reference the same SSO session.
type SsoSession struct {
	Name               string `json:"name"`                          // Session name (from [sso-session name])
	StartURL           string `json:"start_url,omitempty"`           // SSO portal URL
	Region             string `json:"region,omitempty"`              // AWS region for SSO service
	RegistrationScopes string `json:"registration_scopes,omitempty"` // OAuth scopes for registration

	// Metadata from comments
	OrganizationName string `json:"organization_name,omitempty"` // Organization name from # Organization: comment
	Description      string `json:"description,omitempty"`       // Description from # Description: comment

	// Raw properties for complete access to all config values
	Properties map[string]string `json:"properties"` // All session properties as key-value pairs
}

// ExtractedData represents the complete data extracted from AWS CLI config files.
//
// This struct contains all profiles and SSO sessions found in the config,
// along with metadata about when and from where the data was extracted.
type ExtractedData struct {
	Profiles    []AwsCliProfile `json:"profiles"`     // All profiles found in config
	SsoSessions []SsoSession    `json:"sso_sessions"` // All SSO sessions found in config
	ExtractedAt time.Time       `json:"extracted_at"` // Timestamp of extraction
	SourceFile  string          `json:"source_file"`  // Path to source config file
}

// CacheData represents cached AWS CLI data with file modification metadata.
//
// This struct wraps ExtractedData with caching metadata to enable efficient
// cache invalidation based on file modification times.
type CacheData struct {
	Data         ExtractedData `json:"data"`          // Cached extracted data
	LastModified time.Time     `json:"last_modified"` // File modification time when cached
	CachedAt     time.Time     `json:"cached_at"`     // Timestamp when cache was created
	SourcePath   string        `json:"source_path"`   // Path to source config file
}

// FilterCriteria represents filtering options for AWS CLI profiles.
//
// All criteria are combined with AND logic. Multiple values within a field
// are combined with OR logic. Empty fields are ignored.
type FilterCriteria struct {
	AccountIDs   []string      `json:"account_ids,omitempty"`   // Filter by AWS account IDs
	RoleNames    []string      `json:"role_names,omitempty"`    // Filter by IAM role names
	Regions      []string      `json:"regions,omitempty"`       // Filter by AWS regions
	SsoSessions  []string      `json:"sso_sessions,omitempty"`  // Filter by SSO session names
	ProfileTypes []ProfileType `json:"profile_types,omitempty"` // Filter by profile types (SSO, IAM, etc.)
	NamePattern  string        `json:"name_pattern,omitempty"`  // Filter by profile name regex pattern
}

// FilterOptions represents available filter values for building UI dropdowns.
//
// This struct contains lists of unique values found across all profiles,
// useful for populating filter UI components.
type FilterOptions struct {
	AccountIDs  []string `json:"account_ids"`  // All unique account IDs found
	RoleNames   []string `json:"role_names"`   // All unique role names found
	Regions     []string `json:"regions"`      // All unique regions found
	SsoSessions []string `json:"sso_sessions"` // All unique SSO session names found
}

// SsoCacheFile represents the structure of SSO cache files stored by AWS CLI.
//
// AWS CLI stores SSO session tokens in ~/.aws/sso/cache/*.json files.
// This struct matches the JSON structure of those cache files.
type SsoCacheFile struct {
	StartURL     string    `json:"startUrl"`               // SSO portal URL
	Region       string    `json:"region"`                 // AWS region for SSO
	AccessToken  string    `json:"accessToken"`            // OAuth access token
	ExpiresAt    time.Time `json:"expiresAt"`              // Token expiration time
	ClientID     string    `json:"clientId,omitempty"`     // OAuth client ID
	ClientSecret string    `json:"clientSecret,omitempty"` // OAuth client secret
	RefreshToken string    `json:"refreshToken,omitempty"` // OAuth refresh token
}

// ActiveSessionInfo represents information about an active or expired SSO session.
//
// This struct provides details about SSO sessions found in cache files,
// including their validity status.
type ActiveSessionInfo struct {
	SessionName   string    `json:"session_name"`              // SSO session name from config
	StartURL      string    `json:"start_url"`                 // SSO portal URL
	Region        string    `json:"region"`                    // AWS region for SSO
	AccessToken   string    `json:"access_token,omitempty"`    // OAuth access token (may be redacted)
	ExpiresAt     time.Time `json:"expires_at"`                // Token expiration time
	IsExpired     bool      `json:"is_expired"`                // Whether token has expired
	CacheFilePath string    `json:"cache_file_path,omitempty"` // Path to cache file
}

// SessionStatus represents the overall status of AWS SSO sessions.
//
// This struct provides a summary of all SSO sessions, including active and
// expired sessions, plus information about AWS CLI availability.
type SessionStatus struct {
	ActiveSessions  []ActiveSessionInfo `json:"active_sessions"`       // Sessions with valid tokens
	ExpiredSessions []ActiveSessionInfo `json:"expired_sessions"`      // Sessions with expired tokens
	LastChecked     time.Time           `json:"last_checked"`          // Timestamp of status check
	CLIAvailable    bool                `json:"cli_available"`         // Whether AWS CLI is installed
	CLIVersion      string              `json:"cli_version,omitempty"` // AWS CLI version if available
}
