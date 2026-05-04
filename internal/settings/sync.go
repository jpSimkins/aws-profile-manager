package settings

import (
	"fmt"
)

// SyncSettings holds configuration sync settings.
//
// This section controls how the application fetches remote configuration
// files from various sources (S3, HTTP, Git, or local).
type SyncSettings struct {
	Enabled      bool   `json:"enabled"`        // Enable configuration syncing
	AutoUpdate   bool   `json:"auto_update"`    // Auto-fetch on startup
	UpdateOnRead bool   `json:"update_on_read"` // Check for updates when reading config
	Strategy     string `json:"strategy"`       // Sync strategy ("s3", "http", "git", "local")

	// Strategy-specific configuration
	Local LocalSettings `json:"local,omitempty"` // Local file sync settings
	S3    S3Settings    `json:"s3,omitempty"`    // S3 bucket sync settings
	HTTP  HTTPSettings  `json:"http,omitempty"`  // HTTP/HTTPS sync settings
	Git   GitSettings   `json:"git,omitempty"`   // Git repository sync settings (future)
}

// LocalSettings holds local file sync configuration.
type LocalSettings struct {
	Path string `json:"path,omitempty"` // Local file path to configuration
}

// S3Settings holds S3-specific sync configuration.
type S3Settings struct {
	Bucket     string `json:"bucket,omitempty"`      // S3 bucket name
	Key        string `json:"key,omitempty"`         // Object key in bucket
	Region     string `json:"region,omitempty"`      // AWS region
	Profile    string `json:"profile,omitempty"`     // AWS CLI profile to use for authentication
	UseSSO     bool   `json:"use_sso,omitempty"`     // Use SSO for authentication (recommended)
	UseIAM     bool   `json:"use_iam,omitempty"`     // Use IAM user credentials
	PublicRead bool   `json:"public_read,omitempty"` // Public bucket (NOT RECOMMENDED for company use)
}

// HTTPSettings holds HTTP/HTTPS-specific sync configuration.
type HTTPSettings struct {
	URL       string            `json:"url,omitempty"`        // HTTP/HTTPS URL to configuration file
	Headers   map[string]string `json:"headers,omitempty"`    // Custom headers for request
	BasicAuth bool              `json:"basic_auth,omitempty"` // Enable HTTP basic authentication
	Username  string            `json:"username,omitempty"`   // Username for basic auth
}

// GitSettings holds Git-specific sync configuration.
type GitSettings struct {
	RepoURL  string `json:"repo_url,omitempty"`  // Git repository URL (SSH or HTTPS)
	Branch   string `json:"branch,omitempty"`    // Branch or ref to checkout (e.g., "main", "v1.0.0")
	FilePath string `json:"file_path,omitempty"` // Path to config file within repo (e.g., "config.json")
	WorkDir  string `json:"work_dir,omitempty"`  // Working directory for git operations (empty for temp)
}

// GetDefaultSync returns the runtime default for sync settings.
//
// Consumed by GetDefaults() to build the live *Settings struct on first launch.
// The Default values in GetSchema() are separate — those are UI metadata only.
//
// Returns:
//   - SyncSettings: Settings with sync disabled and local strategy
func GetDefaultSync() SyncSettings {
	return SyncSettings{
		Enabled:      false,
		AutoUpdate:   false,
		UpdateOnRead: false,
		Strategy:     "local",
		Local:        LocalSettings{},
		S3: S3Settings{
			Region: "us-west-2",
			UseSSO: true,
		},
		HTTP: HTTPSettings{},
		Git: GitSettings{
			Branch:   "main",
			FilePath: "aws-config.json",
		},
	}
}

// Validate validates sync settings.
//
// Validation Rules:
//   - Strategy must be one of: local, s3, http, https, git
//   - When enabled, strategy-specific required fields must be present:
//   - local: Path required
//   - s3: Bucket and Key required; Region required; one auth method required
//   - http/https: URL required
//   - git: Repo required (future)
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *SyncSettings) Validate() error {
	// Validate strategy is a valid option
	validStrategies := map[string]bool{
		"local": true,
		"s3":    true,
		"http":  true,
		"https": true,
		"git":   true,
	}
	if !validStrategies[s.Strategy] {
		return fmt.Errorf("invalid sync strategy: %s", s.Strategy)
	}

	// Only validate strategy-specific fields if sync is ENABLED
	// This allows users to configure settings even when disabled
	if s.Enabled {
		switch s.Strategy {
		case "local":
			if s.Local.Path == "" {
				return fmt.Errorf("local.path is required when local strategy is enabled")
			}
		case "s3":
			if s.S3.Bucket == "" {
				return fmt.Errorf("s3.bucket is required when S3 strategy is enabled")
			}
			if s.S3.Key == "" {
				return fmt.Errorf("s3.key is required when S3 strategy is enabled")
			}
			if s.S3.Region == "" {
				return fmt.Errorf("s3.region is required when S3 strategy is enabled")
			}
		case "http", "https":
			if s.HTTP.URL == "" {
				return fmt.Errorf("http.url is required when HTTP/HTTPS strategy is enabled")
			}
		case "git":
			if s.Git.RepoURL == "" {
				return fmt.Errorf("git.repo_url is required when git strategy is enabled")
			}
			if s.Git.FilePath == "" {
				return fmt.Errorf("git.file_path is required when git strategy is enabled")
			}
		}
	}

	return nil
}

// GetSchema returns the schema definition for sync settings with conditional dependencies.
//
// This schema includes complex field dependencies where certain fields are only
// required when specific strategies are selected.
//
// Returns:
//   - Schema: Field schema definitions for all sync settings
func (s *SyncSettings) GetSchema() Schema {
	return Schema{
		Version:     "1.0",
		Description: "Configure how the app synchronizes with a configuration file.",
		Fields: map[string]FieldSchema{
			"enabled": {
				Type:        "bool",
				Label:       "Enable Sync",
				Description: "Enable configuration sync",
				Required:    true,
				Default:     false,
				Order:       1,
			},
			"auto_update": {
				Type:        "bool",
				Label:       "Auto Update",
				Description: "Automatically fetch configuration on startup",
				Required:    true,
				Default:     false,
				Order:       2,
			},
			"update_on_read": {
				Type:        "bool",
				Label:       "Update On Read",
				Description: "Check for updates when reading configuration",
				Required:    true,
				Default:     false,
				Order:       3,
			},
			"strategy": {
				Type:        "string",
				Label:       "Sync Strategy",
				Description: "Strategy to use for syncing configuration",
				Required:    true,
				Default:     "local",
				Enum:        []string{"local", "s3", "http", "https", "git"},
				Order:       4,
			},
			"local": {
				Type:        "object",
				Label:       "Local File Configuration",
				Description: "Local file system configuration",
				Required:    false,
				Group:       "Local Settings",
				Order:       10,
				DependsOn: &FieldDependency{
					Field:    "strategy",
					Operator: "equals",
					Value:    "local",
				},
				Nested: &Schema{
					Fields: map[string]FieldSchema{
						"path": {
							Type:        "file",
							Label:       "Configuration File Path",
							Description: "Path to local configuration file",
							Required:    false,
							Order:       1,
							Placeholder: "/path/to/aws-config.json",
						},
					},
				},
			},
			"s3": {
				Type:        "object",
				Label:       "S3 Configuration",
				Description: "Amazon S3 bucket configuration",
				Required:    false,
				Group:       "S3 Settings",
				Order:       11,
				DependsOn: &FieldDependency{
					Field:    "strategy",
					Operator: "equals",
					Value:    "s3",
				},
				Nested: &Schema{
					Fields: map[string]FieldSchema{
						"bucket": {
							Type:        "string",
							Label:       "S3 Bucket Name",
							Description: "S3 bucket name where configuration is stored",
							Required:    false,
							Order:       1,
							Placeholder: "my-config-bucket",
						},
						"key": {
							Type:        "string",
							Label:       "S3 Object Key",
							Description: "S3 object key (path to configuration file)",
							Required:    false,
							Order:       2,
							Placeholder: "config/aws-config.json",
						},
						"region": {
							Type:        "string",
							Label:       "AWS Region",
							Description: "AWS region where the bucket is located",
							Required:    false,
							Default:     "us-west-2",
							Order:       3,
						},
						"profile": {
							Type:        "string",
							Label:       "AWS Profile",
							Description: "AWS CLI profile to use for authentication",
							Required:    false,
							Order:       4,
							Placeholder: "my-sso-profile",
						},
						"use_sso": {
							Type:        "bool",
							Label:       "Use SSO Authentication",
							Description: "Use AWS SSO for authentication (recommended)",
							Required:    false,
							Default:     true,
							Order:       5,
						},
						"use_iam": {
							Type:        "bool",
							Label:       "Use IAM Credentials",
							Description: "Use IAM user credentials for authentication",
							Required:    false,
							Default:     false,
							Order:       6,
						},
						"public_read": {
							Type:        "bool",
							Label:       "Public Bucket Access",
							Description: "Bucket allows public read access (not recommended for company use)",
							Required:    false,
							Default:     false,
							Order:       7,
						},
					},
				},
			},
			"http": {
				Type:        "object",
				Label:       "HTTP/HTTPS Configuration",
				Description: "HTTP/HTTPS URL configuration",
				Required:    false,
				Group:       "HTTP/HTTPS Settings",
				Order:       12,
				DependsOn: &FieldDependency{
					Field:    "strategy",
					Operator: "in",
					Value:    []interface{}{"http", "https"},
				},
				Nested: &Schema{
					Fields: map[string]FieldSchema{
						"url": {
							Type:        "string",
							Label:       "Configuration URL",
							Description: "HTTP or HTTPS URL to the configuration file",
							Required:    false,
							Order:       1,
							Placeholder: "https://example.com/config.json",
						},
						"basic_auth": {
							Type:        "bool",
							Label:       "Use Basic Authentication",
							Description: "Use HTTP basic authentication",
							Required:    false,
							Default:     false,
							Order:       2,
						},
						"username": {
							Type:        "string",
							Label:       "Username",
							Description: "Username for basic authentication",
							Required:    false,
							Order:       3,
							Placeholder: "admin",
						},
					},
				},
			},
			"git": {
				Type:        "object",
				Label:       "Git Configuration",
				Description: "Git repository configuration",
				Required:    false,
				Group:       "Git Settings",
				Order:       13,
				DependsOn: &FieldDependency{
					Field:    "strategy",
					Operator: "equals",
					Value:    "git",
				},
				Nested: &Schema{
					Fields: map[string]FieldSchema{
						"repo": {
							Type:        "string",
							Label:       "Repository URL",
							Description: "Git repository URL (HTTPS or SSH)",
							Required:    false,
							Order:       1,
							Placeholder: "https://github.com/org/repo.git",
						},
						"branch": {
							Type:        "string",
							Label:       "Branch",
							Description: "Git branch to clone and sync from",
							Required:    false,
							Default:     "main",
							Order:       2,
						},
						"config_path": {
							Type:        "string",
							Label:       "Config File Path",
							Description: "Path to configuration file within the repository",
							Required:    false,
							Default:     "aws-config.json",
							Order:       3,
						},
						"use_ssh": {
							Type:        "bool",
							Label:       "Use SSH",
							Description: "Use SSH protocol for Git operations",
							Required:    false,
							Default:     false,
							Order:       4,
						},
						"private": {
							Type:        "bool",
							Label:       "Private Repository",
							Description: "Repository is private (requires authentication)",
							Required:    false,
							Default:     false,
							Order:       5,
						},
					},
				},
			},
		},
	}
}
