// Package backup provides AWS profile and settings backup/restore functionality.
//
// This package orchestrates the profiles package and settings operations to provide
// unified backup/restore capabilities for disaster recovery and configuration migration.
//
// # Architecture
//
// The backup package is a thin orchestration layer that:
//   - Calls profiles.Export/Import for AWS CLI profile operations
//   - Handles settings backup/restore separately
//   - Combines both into a unified backup file format
//   - Provides comprehensive statistics
//
// # What This Package Does
//
//   - Export: profiles.Export() + settings.Get() → BackupFile → JSON
//   - Import: JSON → BackupFile → profiles.Import() + settings.Set()
//   - Settings: Backup/restore application settings independently
//   - File I/O: Read/write/validate backup files
//
// # What This Package Does NOT Do
//
//   - Parse AWS CLI config files (profiles package does this)
//   - Write AWS CLI config files (profiles package does this)
//   - Detect markers (profiles package does this)
//   - Extract profiles (profiles package does this)
//
// # Usage Pattern
//
//	// CLI/GUI builds config from settings
//	cfg := backup.Config{
//	    ConfigPath: settings.GetAwsConfigPath(),
//	    AwsDir:     settings.GetAwsDir(),
//	}
//
//	// Export profiles and settings
//	result, err := backup.ExportProfiles(ctx, cfg, opts, reporter)
//
//	// Import profiles and settings
//	result, err := backup.ImportProfiles(ctx, cfg, opts, reporter)
package backup

import (
	"time"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
)

// =============================================================================
// CONFIGURATION (Injected by CLI/GUI)
// =============================================================================

// Config contains paths and configuration for backup operations.
//
// Built by CLI/GUI from settings and injected into functions.
// Package NEVER imports settings directly - complete dependency injection.
//
// Usage:
//
//	app := settings.GetApplication()
//	cfg := backup.Config{
//	    ConfigPath:  settings.GetAwsConfigPath(),
//	    AwsDir:      settings.GetAwsDir(),
//	    StartMarker: app.GetFormattedStartMarker(),
//	    EndMarker:   app.GetFormattedEndMarker(),
//	}
type Config struct {
	// ConfigPath is the AWS CLI config file path (~/.aws/config)
	ConfigPath string

	// AwsDir is the AWS directory (~/.aws/)
	AwsDir string

	// StartMarker is the formatted start marker for managed section
	StartMarker string

	// EndMarker is the formatted end marker for managed section
	EndMarker string
}

// =============================================================================
// BACKUP FILE FORMAT
// =============================================================================

// BackupFile represents the complete backup file structure.
//
// This is the unified JSON format written to disk containing both profiles
// and settings for complete disaster recovery. The format supports partial
// backups (profiles only, settings only) via optional fields.
//
// Version History:
//   - 2.0: Current version using profiles package architecture
type BackupFile struct {
	// Version is the backup file format version
	Version string `json:"version"`

	// Metadata tracks backup provenance
	Metadata BackupMetadata `json:"metadata"`

	// Data contains AWS profiles (optional - may be nil for settings-only backup)
	Data *schema.Schema `json:"data,omitempty"`

	// Settings contains application settings (optional - may be nil for profiles-only backup)
	Settings *settings.Settings `json:"settings,omitempty"`
}

// BackupMetadata tracks backup provenance.
//
// Provides information about when, where, and by what the backup was created.
type BackupMetadata struct {
	// ExportedAt is the backup creation timestamp (ISO 8601)
	ExportedAt time.Time `json:"exported_at"`

	// ToolVersion is the application version that created the backup
	ToolVersion string `json:"tool_version"`

	// Description is an optional human-readable description
	Description string `json:"description,omitempty"`
}

// =============================================================================
// EXPORT OPTIONS & RESULTS
// =============================================================================

// ExportOptions configures profile and settings export.
//
// Provides granular control over what is exported. Common scenarios:
//   - Full backup: IncludeManaged=true, IncludeAbove=true, IncludeBelow=true
//   - Work only: IncludeManaged=true
//   - Personal only: IncludeAbove=true, IncludeBelow=true
//   - Without settings: ExcludeSettings=true
type ExportOptions struct {
	// OutputPath is where to write backup file (required)
	OutputPath string

	// IncludeManaged exports managed profiles (between markers)
	IncludeManaged bool

	// IncludeAbove exports personal profiles above managed section
	IncludeAbove bool

	// IncludeBelow exports personal profiles below managed section
	IncludeBelow bool

	// ExcludeSettings prevents settings from being included in backup
	ExcludeSettings bool

	// Description is an optional human-readable description for the backup
	Description string
}

// ExportResult contains export statistics and data.
//
// Provides comprehensive information about what was exported including
// profile counts, file paths, and references to exported data.
type ExportResult struct {
	// BackupFile is the complete backup file that was created
	BackupFile *BackupFile

	// OutputPath is where the backup file was written
	OutputPath string

	// ManagedProfiles is the count of managed profiles exported
	ManagedProfiles int

	// UnmanagedAbove is the count of personal profiles above managed section
	UnmanagedAbove int

	// UnmanagedBelow is the count of personal profiles below managed section
	UnmanagedBelow int

	// TotalProfiles is the total count of all profiles exported
	TotalProfiles int

	// SettingsExported indicates whether settings were included
	SettingsExported bool

	// Duration is the time taken for the operation
	Duration time.Duration

	// Detailed statistics from generators
	ManagedStats   generators.SectionStats // Comprehensive stats for managed section
	UnmanagedStats generators.SectionStats // Combined stats for unmanaged sections
}

// =============================================================================
// IMPORT OPTIONS & RESULTS
// =============================================================================

// ImportOptions configures profile and settings import.
//
// Provides granular control over what is imported. Common scenarios:
//   - Full restore: IncludeManaged=true, IncludeAbove=true, IncludeBelow=true
//   - Work only: IncludeManaged=true
//   - Personal only: IncludeAbove=true, IncludeBelow=true
//   - Without settings: IgnoreSettings=true
type ImportOptions struct {
	// BackupPath is the path to backup file (required)
	BackupPath string

	// IncludeManaged imports managed profiles (replaces existing)
	IncludeManaged bool

	// IncludeAbove imports personal profiles above managed section
	IncludeAbove bool

	// IncludeBelow imports personal profiles below managed section
	IncludeBelow bool

	// IgnoreSettings prevents settings restoration even if present in backup
	IgnoreSettings bool

	// BackupCurrentSettings creates a backup of existing settings before restore
	// (default: true, disable for testing or when no backup needed)
	BackupCurrentSettings bool

	// GenerateCheatSheet creates a markdown cheat sheet after import
	// (happens AFTER settings restoration so settings can influence output)
	GenerateCheatSheet bool

	// DryRun previews import without making changes
	DryRun bool
}

// ImportResult contains import statistics.
//
// Provides information about what was imported including profile counts,
// file paths, and whether settings were restored.
type ImportResult struct {
	// BackupFile is the backup file that was imported
	BackupFile *BackupFile

	// ConfigPath is the AWS config file that was modified
	ConfigPath string

	// SettingsRestored indicates whether settings were restored
	SettingsRestored bool

	// SettingsBackupPath is the path to settings backup (if created)
	SettingsBackupPath string

	// CheatSheetGenerated indicates whether cheat sheet was created
	CheatSheetGenerated bool

	// CheatSheetPath is the path to generated cheat sheet (if created)
	CheatSheetPath string

	// Duration is the time taken for the operation
	Duration time.Duration

	// Raw statistics from generators (what's in the backup/merged collection)
	ManagedStats        generators.SectionStats // Comprehensive stats for managed section
	UnmanagedAboveStats generators.SectionStats // Stats for unmanaged above section (includes duplicates)
	UnmanagedBelowStats generators.SectionStats // Stats for unmanaged below section (includes duplicates)

	// Duplicate statistics (what was skipped)
	ManagedDuplicates        profiles.SectionDuplicateStats // Duplicates in managed section
	UnmanagedAboveDuplicates profiles.SectionDuplicateStats // Duplicates in above section
	UnmanagedBelowDuplicates profiles.SectionDuplicateStats // Duplicates in below section
}
