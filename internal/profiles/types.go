// Package profiles provides unified profile management functionality.
//
// This package consolidates installer and backup packages into a clean
// component-based architecture with complete dependency injection. Each
// component is self-contained with its own constructor and methods.
//
// Components:
//   - Installer: Install profiles to AWS config
//   - Exporter: Export profiles to JSON
//   - Importer: Import profiles from JSON
//   - Remover: Remove managed section
//   - CheatSheet: Generate markdown guide
//
// Architecture Principles:
//   - NO api.go file - each component IS the API
//   - Complete dependency injection - package NEVER imports settings
//   - Config built by CLI/GUI from settings and injected
//   - Granular options for independent section control
//
// Usage Pattern:
//
//	// CLI/GUI builds config from settings
//	config := buildConfigFromSettings()
//
//	// Create and use component
//	installer := profiles.NewInstaller(config)
//	result, err := installer.Install(ctx, opts, reporter)
package profiles

import (
	"time"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/schema"
)

// =============================================================================
// CONFIGURATION (Injected by CLI/GUI)
// =============================================================================

// Config contains all configuration for profile operations.
//
// Built by CLI/GUI from settings and injected into components.
// Package NEVER imports settings - complete dependency injection.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	installer := profiles.NewInstaller(config)
type Config struct {
	// File paths (from settings)
	ConfigPath          string // ~/.aws/config
	CheatSheetOutputDir string // Directory for cheat sheet output (typically ~/Desktop)
	CacheDir            string // Cache directory
	AwsDir              string // ~/.aws

	// Markers (from settings)
	StartMarker string // "# START - Managed by AWS Profile Manager"
	EndMarker   string // "# END - Managed by AWS Profile Manager"

	// Header metadata (from settings)
	IncludeTimestamp bool // Append generation timestamp to the start marker comment block
	IncludeVersion   bool // Append app version to the start marker comment block
}

// =============================================================================
// OPTIONS (Per Operation)
// =============================================================================

// InstallOptions configures profile installation.
//
// Use granular filters to control which profiles are installed.
// Empty filter fields mean "include all" for that dimension.
type InstallOptions struct {
	Schema             *schema.Schema // REQUIRED: Profiles to install
	Organizations      []string       // Filter by organization aliases
	Partitions         []string       // Filter by partition names
	Accounts           []string       // Filter by account aliases
	Roles              []string       // Filter by role names
	Regions            []string       // Filter by regions
	AllRegions         bool           // Include all regions (overrides Regions filter)
	GenerateCheatSheet bool           // Create markdown reference
	CheatSheetPath     string         // Override default cheat sheet path
	CheatSheetOnly     bool           // Skip config writes and only generate cheat sheet
	DryRun             bool           // Preview only, don't write files
	OutputFilePath     string         // Override default AWS config output path (Config.ConfigPath)
}

// ExportOptions configures profile export.
//
// Use granular Include* flags to control which sections are exported.
// Independent control over managed section and Above/Below personal profiles.
//
// Common Scenarios:
//   - Team config: IncludeManaged=true
//   - Full backup: All three =true
//   - Personal only: IncludeAbove=true, IncludeBelow=true
type ExportOptions struct {
	OutputPath     string // REQUIRED: Path for JSON output
	IncludeManaged bool   // Export managed section (between markers)
	IncludeAbove   bool   // Export personal profiles above managed section
	IncludeBelow   bool   // Export personal profiles below managed section
	Description    string // Metadata description for export
}

// SchemaReadOptions configures schema read operations.
//
// Controls which sections are included when reading schema from AWS config.
// At least one section must be enabled.
type SchemaReadOptions struct {
	IncludeManaged        bool // Include managed section (between markers)
	IncludeUnmanagedAbove bool // Include unmanaged profiles above managed section
	IncludeUnmanagedBelow bool // Include unmanaged profiles below managed section
}

// ExportStats tracks statistics during export operations.
//
// Returned by config_reader to avoid double-processing with generators.
type ExportStats struct {
	ManagedProfiles    int // Profiles in managed section
	UnmanagedAbove     int // Personal profiles above markers
	UnmanagedBelow     int // Personal profiles below markers
	TotalProfiles      int // Total across all sections
	SsoSessions        int // SSO sessions in managed section
	SsoProfiles        int // SSO profiles in managed section
	IamProfiles        int // IAM profiles in managed section
	AssumeRoleProfiles int // AssumeRole profiles in managed section
	GenericProfiles    int // Generic profiles in managed section
}

// ImportOptions configures profile import.
//
// Use granular Include* flags to control which sections are imported.
// Independent control over managed section and Above/Below personal profiles.
//
// Common Scenarios:
//   - Full restore: All three =true
//   - Work only: IncludeManaged=true
//   - Personal only: IncludeAbove=true, IncludeBelow=true
type ImportOptions struct {
	BackupPath     string // REQUIRED: Path to JSON backup
	IncludeManaged bool   // Import managed section (replaces existing)
	IncludeAbove   bool   // Import personal profiles above managed section
	IncludeBelow   bool   // Import personal profiles below managed section
	DryRun         bool   // Preview only, don't write files
}

// RemoveOptions configures profile removal.
//
// Controls removal of managed section and optional cheat sheet cleanup.
type RemoveOptions struct {
	RemoveCheatSheet bool   // Also delete cheat sheet file
	CheatSheetPath   string // Override default cheat sheet path for cleanup
	DryRun           bool   // Preview only, don't delete files
}

// CheatSheetOptions configures cheat sheet generation.
//
// Generates markdown reference guide for profiles in collection.
type CheatSheetOptions struct {
	Collection *schema.ProfileCollection // REQUIRED: Profiles to document
	OutputPath string                    // Override default output path
}

// =============================================================================
// RESULTS (Per Operation)
// =============================================================================

// InstallResult contains installation results.
//
// Provides comprehensive statistics about what was installed including
// profile counts, file paths, and metadata about the operation.
type InstallResult struct {
	TotalProfiles   int                     // Total profiles written to config
	ManagedProfiles int                     // Managed section profile count
	UnmanagedAbove  int                     // Personal profiles above managed section
	UnmanagedBelow  int                     // Personal profiles below managed section
	SsoSessions     int                     // SSO session configurations written
	ManagedStats    generators.SectionStats // Detailed breakdown of managed profiles
	ConfigPath      string                  // Path to AWS config file
	CheatSheetPath  string                  // Path to generated cheat sheet (if created)
	Duration        time.Duration           // Time taken for operation
	CreatedMarkers  bool                    // Whether markers were created (vs replaced)
}

// ExportResult contains export results.
//
// Provides statistics about what was exported and references to the
// exported data for inspection.
type ExportResult struct {
	TotalProfiles   int            // Total profiles exported
	ManagedProfiles int            // Managed section profile count
	UnmanagedAbove  int            // Personal profiles above managed section
	UnmanagedBelow  int            // Personal profiles below managed section
	SsoSessions     int            // SSO session count in export
	OutputPath      string         // Path to JSON output file
	Schema          *schema.Schema // Exported schema
	Duration        time.Duration  // Time taken for operation
	Timestamp       time.Time      // When export was performed
}

// SectionDuplicateStats tracks duplicate profiles by type.
//
// Used during import to show users exactly which profile types
// were skipped due to duplication. GUI can subtract these from
// raw stats to calculate actual profiles written.
type SectionDuplicateStats struct {
	TotalDuplicates    int // Total duplicate profiles in this section
	SsoProfiles        int // SSO profile duplicates
	IamProfiles        int // IAM profile duplicates
	AssumeRoleProfiles int // AssumeRole profile duplicates
	GenericProfiles    int // Generic profile duplicates
}

// ImportPlan contains prepared import data from PrepareImport.
//
// This caches the parsed schema and pre-generated content so ExecuteImport
// doesn't need to re-parse or re-generate. Used for two-phase import where
// GUI shows preview (PrepareImport) then executes (ExecuteImport) on confirmation.
type ImportPlan struct {
	Schema     *schema.Schema // Parsed schema (ready to use)
	BackupPath string         // Source backup file path

	// Pre-generated content (cached from generators)
	ManagedContent        string // Generated managed section content
	UnmanagedAboveContent string // Generated unmanaged above content
	UnmanagedBelowContent string // Generated unmanaged below content

	// Raw statistics from generators (what's in the backup/merged collection)
	// These represent the FULL content before duplicate removal
	ManagedStats        generators.SectionStats // Comprehensive stats for managed section
	UnmanagedAboveStats generators.SectionStats // Stats for unmanaged above section (includes duplicates)
	UnmanagedBelowStats generators.SectionStats // Stats for unmanaged below section (includes duplicates)

	// Duplicate statistics (what WON'T be written due to duplication)
	// GUI subtracts these from raw stats to get actual profiles written
	ManagedDuplicates        SectionDuplicateStats // Duplicates in managed section
	UnmanagedAboveDuplicates SectionDuplicateStats // Duplicates in above section
	UnmanagedBelowDuplicates SectionDuplicateStats // Duplicates in below section
}

// ImportResult contains import results.
//
// Provides statistics about what was imported, including duplicate
// detection for personal profiles.
type ImportResult struct {
	BackupPath string         // Path to source JSON backup
	Schema     *schema.Schema // Imported schema
	Duration   time.Duration  // Time taken for operation

	// Raw statistics from generators (what was in the backup/merged collection)
	ManagedStats        generators.SectionStats // Comprehensive stats for managed section
	UnmanagedAboveStats generators.SectionStats // Stats for unmanaged above section (includes duplicates)
	UnmanagedBelowStats generators.SectionStats // Stats for unmanaged below section (includes duplicates)

	// Duplicate statistics (what was skipped)
	ManagedDuplicates        SectionDuplicateStats // Duplicates in managed section
	UnmanagedAboveDuplicates SectionDuplicateStats // Duplicates in above section
	UnmanagedBelowDuplicates SectionDuplicateStats // Duplicates in below section
}

// RemoveResult contains removal results.
//
// Indicates what was removed during the operation.
type RemoveResult struct {
	ConfigPath        string        // Path to AWS config file
	CheatSheetPath    string        // Path to cheat sheet file checked for cleanup
	RemovedConfig     bool          // Whether managed section was removed
	RemovedCheatSheet bool          // Whether cheat sheet was removed
	ProfilesRemoved   int           // Number of profiles removed from config
	Duration          time.Duration // Time taken for operation
}

// CheatSheetResult contains cheat sheet results.
//
// Provides information about the generated cheat sheet.
type CheatSheetResult struct {
	OutputPath string        // Path to generated markdown file
	Duration   time.Duration // Time taken for operation
	Profiles   int           // Number of profiles documented
	Sessions   int           // Number of SSO sessions documented
	FileSize   int64         // Size of generated file in bytes
}

// =============================================================================
// INTERNAL TYPES (Package-Private)
// =============================================================================

// sectionStats tracks profile generation statistics.
//
// Used internally by config_writer to accumulate counts during
// profile generation.
type sectionStats struct {
	TotalProfiles      int // Total profiles generated
	SsoSessions        int // SSO sessions generated
	SsoProfiles        int // SSO profiles generated
	IamProfiles        int // IAM profiles generated
	AssumeRoleProfiles int // Assume role profiles generated
	GenericProfiles    int // Generic profiles generated
	// Organizational details (from SSO generator)
	OrganizationCount int // Number of organizations
	PartitionCount    int // Number of partitions
	AccountCount      int // Number of accounts
	RoleCount         int // Number of roles
	RegionCount       int // Number of regions
}

// markerPosition tracks the line positions of managed section markers.
//
// Used internally for detecting and manipulating managed section boundaries.
type markerPosition struct {
	StartLine int  // Line number of start marker (-1 if not found)
	EndLine   int  // Line number of end marker (-1 if not found)
	Found     bool // Whether both markers were found
}
