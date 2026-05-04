// Package schema defines data models for AWS configuration management.
//
// This package provides the core schema types for representing AWS profiles,
// organizations, accounts, and credentials in a structured format. The schema
// supports multiple authentication methods (SSO, IAM, AssumeRole) and handles
// both managed and unmanaged profile sections.
//
// # Schema Architecture
//
// The package follows a hierarchical structure designed to separate concerns
// between orchestration and business logic:
//
//	┌─────────────────────────────────────────┐
//	│              Schema                     │  <- Top level (orchestration)
//	├─────────────────────────────────────────┤
//	│  Managed:   ProfileCollection           │  <- Company-managed profiles
//	│  Unmanaged: ProfileCollection (2 parts) │  <- User personal profiles
//	└─────────────────────────────────────────┘
//	              ↓
//	┌─────────────────────────────────────────┐
//	│       ProfileCollection                 │  <- Business logic level
//	├─────────────────────────────────────────┤
//	│  Organizations (SSO)                    │  <- Section-agnostic
//	│  IamUsers                               │  <- Reusable components
//	│  AssumeRoleChains                       │  <- work with this
//	│  GenericProfiles                        │
//	└─────────────────────────────────────────┘
//
// # Component Usage
//
//   - Orchestrator (installer, backup): Works with Schema, understands sections
//   - Writers/Generators: Work with ProfileCollection, section-agnostic
//   - Sync: Provides Schema instances from remote sources
//
// # Key Principle
//
// Heavy-lifting components (writers, generators, filters) work with
// ProfileCollection. They don't know or care about managed vs unmanaged -
// that's orchestration logic. This keeps business logic clean and reusable.
//
// # Profile Types
//
//   - SSO: AWS IAM Identity Center authentication
//   - IAM: Static credentials or credential_process
//   - AssumeRole: Role assumption chains
//   - Generic: Catch-all for custom configurations
package schema

// Schema represents the complete AWS profile configuration.
//
// This is the top-level schema used throughout the application for sync,
// import, export, and installation operations. It contains both managed
// (company-controlled) and unmanaged (user-personal) profile sections.
//
// Section Awareness:
//   - Managed: Always replaced during sync/install operations
//   - Unmanaged: Always preserved and merged with managed profiles
//
// Fields:
//   - Version: Schema version for compatibility checking
//   - Managed: Company-managed profiles (replaced on sync)
//   - Unmanaged: User's personal profiles (preserved, split above/below managed)
//   - Presets: Pre-built filter configurations for installation
//   - Settings: Optional application settings for backup/restore
//   - Metadata: Optional export information
type Schema struct {
	Version   string             `json:"version"`             // Schema version (e.g., "1.0")
	Managed   *ProfileCollection `json:"managed,omitempty"`   // Company-managed profiles
	Unmanaged *UnmanagedProfiles `json:"unmanaged,omitempty"` // User personal profiles
	Presets   map[string]*Preset `json:"presets,omitempty"`   // Installation presets
	Settings  interface{}        `json:"settings,omitempty"`  // Application settings
	Metadata  *Metadata          `json:"metadata,omitempty"`  // Export metadata
}

// ProfileCollection represents a collection of AWS profiles.
//
// This is the core working type for business logic components. It is
// section-agnostic - the same structure is used for managed, unmanaged-above,
// and unmanaged-below sections.
//
// Authentication Methods:
//   - Organizations: SSO-based profiles
//   - IamUsers: Static credentials or credential_process
//   - AssumeRoleChains: Role assumption profiles
//   - GenericProfiles: Custom configurations
type ProfileCollection struct {
	Organizations    map[string]*Organization `json:"organizations,omitempty"`      // SSO organizations by alias
	IamUsers         []*IamUser               `json:"iam_users,omitempty"`          // IAM user profiles
	AssumeRoleChains []*AssumeRoleChain       `json:"assume_role_chains,omitempty"` // Role chains
	GenericProfiles  []*GenericProfile        `json:"generic_profiles,omitempty"`   // Custom profiles
}

// UnmanagedProfiles contains user's personal profiles.
//
// Personal profiles are separated by their position relative to the managed
// section in the AWS config file. This preserves the user's organization
// preference when merging profiles.
type UnmanagedProfiles struct {
	Above *ProfileCollection `json:"above_managed,omitempty"` // Profiles appearing above managed section
	Below *ProfileCollection `json:"below_managed,omitempty"` // Profiles appearing below managed section
}

// Metadata contains optional schema information.
//
// Used primarily for export operations to track when and where a configuration
// was exported from.
type Metadata struct {
	ExportedAt  string `json:"exported_at,omitempty"`  // ISO 8601 timestamp
	ExportedBy  string `json:"exported_by,omitempty"`  // User or system identifier
	ToolVersion string `json:"tool_version,omitempty"` // Tool version used for export
	Description string `json:"description,omitempty"`  // Optional user description
}

// Organization represents an AWS organization configuration.
//
// Organizations can span multiple AWS partitions (commercial and GovCloud),
// each with their own SSO portal URL and accounts.
type Organization struct {
	Name        string               `json:"name"`                  // Display name
	Description string               `json:"description,omitempty"` // Optional description
	Partitions  map[string]Partition `json:"partitions"`            // Partitions by name (e.g., "commercial", "govcloud")
}

// Partition represents AWS partition-specific configuration.
//
// A partition is either commercial (standard AWS) or GovCloud, each with
// its own SSO portal, regions, accounts, and available roles.
type Partition struct {
	URL           string    `json:"url"`            // SSO start URL (e.g., https://my-org.awsapps.com/start)
	DefaultRegion string    `json:"default_region"` // Default AWS region (e.g., us-east-1)
	Regions       []string  `json:"regions"`        // Available regions for profiles
	Accounts      []Account `json:"accounts"`       // AWS accounts in this partition
	Roles         []string  `json:"roles"`          // Available IAM roles
}

// Account represents an individual AWS account.
//
// Each account has a human-readable alias, display name, and AWS account ID.
type Account struct {
	Alias string `json:"alias"` // Short identifier (e.g., "prod", "dev")
	Name  string `json:"name"`  // Display name (e.g., "Production Environment")
	ID    string `json:"id"`    // AWS account ID (12-digit number)
}

// IamUser represents an IAM user profile configuration.
//
// Supports both static credentials and credential_process authentication.
type IamUser struct {
	ProfileName       string `json:"profile_name"`                    // AWS CLI profile name
	Region            string `json:"region,omitempty"`                // AWS region
	AwsAccessKeyID    string `json:"aws_access_key_id,omitempty"`     // Static access key ID
	AwsSecretKey      string `json:"aws_secret_access_key,omitempty"` // Static secret access key
	CredentialProcess string `json:"credential_process,omitempty"`    // Credential process command
}

// AssumeRoleChain represents a role assumption profile.
//
// These profiles use an existing source profile to assume an IAM role,
// optionally requiring MFA authentication.
type AssumeRoleChain struct {
	ProfileName   string `json:"profile_name"`           // AWS CLI profile name
	SourceProfile string `json:"source_profile"`         // Source profile to use for credentials
	RoleArn       string `json:"role_arn"`               // IAM role ARN to assume
	MfaSerial     string `json:"mfa_serial,omitempty"`   // MFA device ARN (if MFA required)
	Region        string `json:"region,omitempty"`       // AWS region (optional)
	ExternalID    string `json:"external_id,omitempty"`  // External ID for cross-account access
	SessionName   string `json:"session_name,omitempty"` // Custom role session name
}

// GenericProfile represents a catch-all profile with arbitrary properties.
//
// Used for profiles that don't fit the SSO, IAM, or AssumeRole patterns.
// Simply stores key-value pairs that are written as-is to the AWS config file.
type GenericProfile struct {
	ProfileName string            `json:"profile_name"` // AWS CLI profile name
	Properties  map[string]string `json:"properties"`   // Key-value properties
}

// Preset represents a pre-built filter configuration for profile installation.
//
// Presets allow users to quickly select common filter combinations when
// installing AWS profiles. They automatically populate filter selections
// in the GUI and can be referenced in CLI operations.
//
// Empty filter fields mean "include all" for that dimension. For example:
//   - Empty Organizations = all organizations
//   - Empty Roles = all roles
//   - Empty Regions = use default region only (unless AllRegions is true)
//
// Example use cases:
//   - "Developer" preset: Single org, Developer role only
//   - "DevSecOps" preset: Single org, multiple admin roles
//   - "Break Glass" preset: All orgs, emergency access role only
type Preset struct {
	Label         string   `json:"label"`                   // Display name (e.g., "Developer", "DevSecOps")
	Description   string   `json:"description,omitempty"`   // Optional explanation of preset purpose
	Organizations []string `json:"organizations,omitempty"` // Filter by organization aliases (empty = all)
	Partitions    []string `json:"partitions,omitempty"`    // Filter by partition names (empty = all)
	Accounts      []string `json:"accounts,omitempty"`      // Filter by account aliases (empty = all)
	Roles         []string `json:"roles,omitempty"`         // Filter by role names (empty = all)
	Regions       []string `json:"regions,omitempty"`       // Filter by regions (empty = default only)
	AllRegions    bool     `json:"all_regions,omitempty"`   // Include all regions (overrides Regions)
}
