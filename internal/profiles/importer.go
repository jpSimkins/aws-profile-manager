package profiles

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// Importer imports AWS CLI profiles from JSON.
//
// This component reads JSON backup files and imports selected sections
// to AWS config. Uses duplicate detection for personal profiles.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	importer := profiles.NewImporter(config)
//	result, err := importer.Import(ctx, opts, reporter)
type Importer struct {
	config Config
	writer *configWriter
	reader *configReader
	merger *merger
}

// NewImporter creates a new Importer with injected configuration.
//
// Parameters:
//   - config: Configuration injected by CLI/GUI (from settings)
//
// Returns:
//   - *Importer: Ready to use importer instance
//
// Example:
//
//	config := buildConfigFromSettings()
//	importer := profiles.NewImporter(config)
func NewImporter(config Config) *Importer {
	return &Importer{
		config: config,
		writer: newConfigWriter(config),
		reader: newConfigReader(config),
		merger: newMerger(),
	}
}

// PrepareImport prepares an import by parsing the backup and generating content.
//
// This is Phase 1 of a two-phase import. It reads the backup file, parses the schema,
// and runs generators to create content and get accurate counts. The generated content
// and stats are cached in the returned ImportPlan for later use by ExecuteImport.
//
// Use this for GUI preview functionality where you want to show what WILL be imported
// before actually writing files. The ImportPlan can be passed to ExecuteImport to
// complete the import without re-parsing or re-generating.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Import options (what to include)
//   - reporter: Progress reporter
//
// Returns:
//   - *ImportPlan: Prepared import with cached content and stats
//   - error: Any error during preparation
//
// Example:
//
//	// Show preview
//	plan, err := importer.PrepareImport(ctx, opts, reporter)
//	if err != nil { return err }
//	fmt.Printf("Will import %d profiles\n", plan.TotalProfiles)
//
//	// User confirms, then execute
//	result, err := importer.ExecuteImport(ctx, plan, opts, reporter)
func (i *Importer) PrepareImport(
	ctx context.Context,
	opts ImportOptions,
	reporter task.Reporter,
) (*ImportPlan, error) {
	reporter.ReportStatus("Reading backup file...")
	logging.Debug.Log("PrepareImport started", "backup", opts.BackupPath)

	// Validate backup path
	if opts.BackupPath == "" {
		return nil, fmt.Errorf("backup path is required")
	}

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Read and parse backup file
	content, err := readFileContent(opts.BackupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	// Try parsing as new format (with "data" wrapper) first
	var backupFile struct {
		Version string         `json:"version"`
		Data    *schema.Schema `json:"data,omitempty"`
	}

	if err := json.Unmarshal([]byte(content), &backupFile); err != nil {
		return nil, fmt.Errorf("failed to parse backup file: %w", err)
	}

	var backupSchema schema.Schema

	// Check if we have the new format with "data" wrapper
	if backupFile.Data != nil {
		backupSchema = *backupFile.Data
		logging.Debug.Log("Detected new backup format (with data wrapper)")
	} else {
		// Try old flat format
		if err := json.Unmarshal([]byte(content), &backupSchema); err != nil {
			return nil, fmt.Errorf("failed to parse backup file as old format: %w", err)
		}
		logging.Debug.Log("Detected old backup format (flat structure)")
	}

	// For non-dry-run, validate we have some data to import
	// For dry-run, allow empty schemas (will just return zero counts)
	isEmpty := backupSchema.Managed == nil && backupSchema.Unmanaged == nil
	if isEmpty && !opts.DryRun {
		return nil, fmt.Errorf("backup file contains no profiles to import")
	}

	plan := &ImportPlan{
		Schema:     &backupSchema,
		BackupPath: opts.BackupPath,
	}

	// Generate content for managed section (and get counts)
	if opts.IncludeManaged && backupSchema.Managed != nil {
		reporter.ReportStatus("Preparing managed profiles...")

		// Run all generators to get content AND stats
		ssoContent, ssoStats, err := generators.GenerateSsoProfiles(ctx, backupSchema.Managed, reporter)
		if err != nil {
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to generate SSO profiles: %w", err)
		}

		iamContent, iamStats, err := generators.GenerateIamProfiles(ctx, backupSchema.Managed, reporter)
		if err != nil {
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to generate IAM profiles: %w", err)
		}

		assumeRoleContent, assumeRoleStats, err := generators.GenerateAssumeRoleProfiles(ctx, backupSchema.Managed, reporter)
		if err != nil {
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to generate AssumeRole profiles: %w", err)
		}

		genericContent, genericStats, err := generators.GenerateGenericProfiles(ctx, backupSchema.Managed, reporter)
		if err != nil {
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to generate Generic profiles: %w", err)
		}

		// Combine all content
		plan.ManagedContent = ssoContent + iamContent + assumeRoleContent + genericContent

		// Aggregate detailed stats for managed section
		plan.ManagedStats = generators.SectionStats{
			ProfilesWritten:    ssoStats.ProfilesWritten + iamStats.ProfilesWritten + assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten,
			SessionsWritten:    ssoStats.SessionsWritten,
			SsoProfiles:        ssoStats.SsoProfiles,
			IamProfiles:        iamStats.IamProfiles,
			AssumeRoleProfiles: assumeRoleStats.AssumeRoleProfiles,
			GenericProfiles:    genericStats.GenericProfiles,
			OrganizationCount:  ssoStats.OrganizationCount,
			PartitionCount:     ssoStats.PartitionCount,
			AccountCount:       ssoStats.AccountCount,
			RoleCount:          ssoStats.RoleCount,
			RegionCount:        ssoStats.RegionCount,
		}
	}

	// Handle unmanaged sections with duplicate detection
	var aboveDupStats SectionDuplicateStats
	var belowDupStats SectionDuplicateStats

	if (opts.IncludeAbove || opts.IncludeBelow) && backupSchema.Unmanaged != nil {
		reporter.ReportStatus("Reading existing personal profiles...")

		// Read existing unmanaged sections (if config exists)
		existingSchema, _, err := i.reader.readConfig(ctx, ExportOptions{
			IncludeAbove: true,
			IncludeBelow: true,
		}, reporter)

		// If file doesn't exist, treat as no existing profiles (no duplicates to detect)
		var existingUnmanaged *schema.UnmanagedProfiles
		if err != nil {
			if err == context.Canceled {
				return nil, err
			}
			// Config file doesn't exist or can't be read - treat as empty
			// This is expected for first-time imports or restore operations
			logging.Debug.Log("No existing config found - treating as empty", "error", err)
			existingUnmanaged = &schema.UnmanagedProfiles{}
		} else {
			existingUnmanaged = existingSchema.Unmanaged
			if existingUnmanaged == nil {
				existingUnmanaged = &schema.UnmanagedProfiles{}
			}
		}

		// Merge with duplicate detection
		var mergedAbove *schema.ProfileCollection
		var mergedBelow *schema.ProfileCollection

		if opts.IncludeAbove && backupSchema.Unmanaged.Above != nil {
			reporter.ReportStatus("Merging personal profiles (above) - detecting duplicates...")
			mergedAbove, aboveDupStats = i.merger.merge(existingUnmanaged.Above, backupSchema.Unmanaged.Above)
			logging.Debug.Log("Merged above profiles",
				"duplicates_total", aboveDupStats.TotalDuplicates,
				"duplicates_sso", aboveDupStats.SsoProfiles,
				"duplicates_iam", aboveDupStats.IamProfiles,
			)
		}

		if opts.IncludeBelow && backupSchema.Unmanaged.Below != nil {
			reporter.ReportStatus("Merging personal profiles (below) - detecting duplicates...")
			mergedBelow, belowDupStats = i.merger.merge(existingUnmanaged.Below, backupSchema.Unmanaged.Below)
			logging.Debug.Log("Merged below profiles",
				"duplicates_total", belowDupStats.TotalDuplicates,
				"duplicates_sso", belowDupStats.SsoProfiles,
				"duplicates_iam", belowDupStats.IamProfiles,
			)
		}

		// Generate content from MERGED collections
		if mergedAbove != nil {
			reporter.ReportStatus("Preparing personal profiles (above)...")

			ssoContent, ssoStats, err := generators.GenerateSsoProfiles(ctx, mergedAbove, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate SSO profiles: %w", err)
			}

			iamContent, iamStats, err := generators.GenerateIamProfiles(ctx, mergedAbove, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate IAM profiles: %w", err)
			}

			assumeRoleContent, assumeRoleStats, err := generators.GenerateAssumeRoleProfiles(ctx, mergedAbove, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate AssumeRole profiles: %w", err)
			}

			genericContent, genericStats, err := generators.GenerateGenericProfiles(ctx, mergedAbove, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate Generic profiles: %w", err)
			}

			plan.UnmanagedAboveContent = ssoContent + iamContent + assumeRoleContent + genericContent

			// Store detailed stats for above section (raw stats from merged collection)
			plan.UnmanagedAboveStats = generators.SectionStats{
				ProfilesWritten:    ssoStats.ProfilesWritten + iamStats.ProfilesWritten + assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten,
				SessionsWritten:    ssoStats.SessionsWritten,
				SsoProfiles:        ssoStats.SsoProfiles,
				IamProfiles:        iamStats.IamProfiles,
				AssumeRoleProfiles: assumeRoleStats.AssumeRoleProfiles,
				GenericProfiles:    genericStats.GenericProfiles,
				OrganizationCount:  ssoStats.OrganizationCount,
				PartitionCount:     ssoStats.PartitionCount,
				AccountCount:       ssoStats.AccountCount,
				RoleCount:          ssoStats.RoleCount,
				RegionCount:        ssoStats.RegionCount,
			}
		}

		if mergedBelow != nil {
			reporter.ReportStatus("Preparing personal profiles (below)...")

			ssoContent, ssoStats, err := generators.GenerateSsoProfiles(ctx, mergedBelow, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate SSO profiles: %w", err)
			}

			iamContent, iamStats, err := generators.GenerateIamProfiles(ctx, mergedBelow, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate IAM profiles: %w", err)
			}

			assumeRoleContent, assumeRoleStats, err := generators.GenerateAssumeRoleProfiles(ctx, mergedBelow, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate AssumeRole profiles: %w", err)
			}

			genericContent, genericStats, err := generators.GenerateGenericProfiles(ctx, mergedBelow, reporter)
			if err != nil && err != context.Canceled {
				return nil, fmt.Errorf("failed to generate Generic profiles: %w", err)
			}

			plan.UnmanagedBelowContent = ssoContent + iamContent + assumeRoleContent + genericContent

			// Store detailed stats for below section (raw stats from merged collection)
			plan.UnmanagedBelowStats = generators.SectionStats{
				ProfilesWritten:    ssoStats.ProfilesWritten + iamStats.ProfilesWritten + assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten,
				SessionsWritten:    ssoStats.SessionsWritten,
				SsoProfiles:        ssoStats.SsoProfiles,
				IamProfiles:        iamStats.IamProfiles,
				AssumeRoleProfiles: assumeRoleStats.AssumeRoleProfiles,
				GenericProfiles:    genericStats.GenericProfiles,
				OrganizationCount:  ssoStats.OrganizationCount,
				PartitionCount:     ssoStats.PartitionCount,
				AccountCount:       ssoStats.AccountCount,
				RoleCount:          ssoStats.RoleCount,
				RegionCount:        ssoStats.RegionCount,
			}
		}
	}

	// Store duplicate stats in plan
	plan.UnmanagedAboveDuplicates = aboveDupStats
	plan.UnmanagedBelowDuplicates = belowDupStats
	// Managed section doesn't have duplicates (always replaces), so leave empty

	logging.Debug.Log("PrepareImport completed",
		"managed_profiles", plan.ManagedStats.ProfilesWritten,
		"above_profiles", plan.UnmanagedAboveStats.ProfilesWritten,
		"below_profiles", plan.UnmanagedBelowStats.ProfilesWritten,
		"above_duplicates", plan.UnmanagedAboveDuplicates.TotalDuplicates,
		"below_duplicates", plan.UnmanagedBelowDuplicates.TotalDuplicates,
	)

	return plan, nil
}

// ExecuteImport executes a prepared import by writing the cached content.
//
// This is Phase 2 of a two-phase import. It takes the ImportPlan from PrepareImport
// and writes the pre-generated content to the AWS config file. No re-parsing or
// re-generation happens - content is reused from the plan.
//
// Use this after PrepareImport when the user confirms they want to proceed with
// the import. This is much faster than running Import() again since all the heavy
// work (parsing, generating) was done in PrepareImport.
//
// Parameters:
//   - ctx: Context for cancellation
//   - plan: Prepared import plan from PrepareImport
//   - opts: Import options (same as used in PrepareImport)
//   - reporter: Progress reporter
//
// Returns:
//   - *ImportResult: Import results with actual counts
//   - error: Any error during execution
//
// Example:
//
//	plan, _ := importer.PrepareImport(ctx, opts, reporter)
//	// Show preview to user...
//	// User confirms...
//	result, err := importer.ExecuteImport(ctx, plan, opts, reporter)
func (i *Importer) ExecuteImport(
	ctx context.Context,
	plan *ImportPlan,
	opts ImportOptions,
	reporter task.Reporter,
) (*ImportResult, error) {
	startTime := time.Now()
	reporter.ReportStatus("Executing import...")
	logging.Debug.Log("ExecuteImport started")

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &ImportResult{
		BackupPath: plan.BackupPath,
		Schema:     plan.Schema,
	}

	// Dry run - skip all file writes, just return stats from plan
	if opts.DryRun {
		result.Duration = time.Since(startTime)
		result.ManagedStats = plan.ManagedStats
		result.UnmanagedAboveStats = plan.UnmanagedAboveStats
		result.UnmanagedBelowStats = plan.UnmanagedBelowStats
		result.ManagedDuplicates = plan.ManagedDuplicates
		result.UnmanagedAboveDuplicates = plan.UnmanagedAboveDuplicates
		result.UnmanagedBelowDuplicates = plan.UnmanagedBelowDuplicates

		totalProfiles := plan.ManagedStats.ProfilesWritten + plan.UnmanagedAboveStats.ProfilesWritten + plan.UnmanagedBelowStats.ProfilesWritten
		reporter.ReportStatus(fmt.Sprintf("DRY RUN: Would import %d profiles", totalProfiles))
		logging.Debug.Log("DRY RUN completed",
			"managed", plan.ManagedStats.ProfilesWritten,
			"unmanaged_above", plan.UnmanagedAboveStats.ProfilesWritten,
			"unmanaged_below", plan.UnmanagedBelowStats.ProfilesWritten,
			"total", totalProfiles,
		)

		return result, nil
	}

	// Write managed section using cached content
	if opts.IncludeManaged && plan.ManagedContent != "" {
		reporter.ReportStatus("Writing managed section...")

		// Use writer to write pre-generated content
		profileCount := plan.ManagedStats.ProfilesWritten
		_, err := i.writer.writeConfigWithContent(ctx, plan.ManagedContent, profileCount, reporter)
		if err != nil {
			// Don't wrap cancellation errors
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to write managed section: %w", err)
		}

		logging.Debug.Log("Managed section written", "profiles", profileCount)
	}

	// Write unmanaged sections using cached content
	if (opts.IncludeAbove && plan.UnmanagedAboveContent != "") || (opts.IncludeBelow && plan.UnmanagedBelowContent != "") {
		reporter.ReportStatus("Writing personal profiles...")

		err := i.writeUnmanagedContent(ctx, plan.UnmanagedAboveContent, plan.UnmanagedBelowContent, reporter)
		if err != nil {
			// Don't wrap cancellation errors
			if err == context.Canceled {
				return nil, err
			}
			return nil, fmt.Errorf("failed to write unmanaged sections: %w", err)
		}

		logging.Debug.Log("Unmanaged sections written",
			"above", plan.UnmanagedAboveStats.ProfilesWritten,
			"below", plan.UnmanagedBelowStats.ProfilesWritten,
		)
	}

	result.Duration = time.Since(startTime)

	// Copy stats and duplicates from plan (no re-processing)
	result.ManagedStats = plan.ManagedStats
	result.UnmanagedAboveStats = plan.UnmanagedAboveStats
	result.UnmanagedBelowStats = plan.UnmanagedBelowStats
	result.ManagedDuplicates = plan.ManagedDuplicates
	result.UnmanagedAboveDuplicates = plan.UnmanagedAboveDuplicates
	result.UnmanagedBelowDuplicates = plan.UnmanagedBelowDuplicates

	logging.Debug.Log("ExecuteImport completed",
		"duration", result.Duration,
		"managed_profiles", result.ManagedStats.ProfilesWritten,
		"above_profiles", result.UnmanagedAboveStats.ProfilesWritten,
		"below_profiles", result.UnmanagedBelowStats.ProfilesWritten,
	)

	return result, nil
}

// Import is DEPRECATED - use PrepareImport + ExecuteImport instead.
//
// This function is kept only for backward compatibility with existing tests.
// New code should use the two-phase approach:
//  1. PrepareImport - parse and generate content
//  2. ExecuteImport - write pre-generated content
//
// The two-phase approach enables GUI preview before execution.
func (i *Importer) Import(
	ctx context.Context,
	opts ImportOptions,
	reporter task.Reporter,
) (*ImportResult, error) {
	// Delegate to two-phase approach
	plan, err := i.PrepareImport(ctx, opts, reporter)
	if err != nil {
		return nil, err
	}
	return i.ExecuteImport(ctx, plan, opts, reporter)
}

// writeUnmanagedContent writes pre-generated unmanaged section content.
//
// This is used by ExecuteImport to write cached content from PrepareImport.
// The content is already merged (duplicates removed) and generated.
//
// Parameters:
//   - ctx: Context for cancellation
//   - aboveContent: Pre-generated content for above section
//   - belowContent: Pre-generated content for below section
//   - reporter: Progress reporter
//
// Returns:
//   - error: Any error during writing
func (i *Importer) writeUnmanagedContent(
	ctx context.Context,
	aboveContent string,
	belowContent string,
	reporter task.Reporter,
) error {
	logging.Debug.Log("writeUnmanagedContent started")

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	reporter.ReportStatus("Reading existing config file...")

	// Read existing config to preserve managed section
	lines, err := readConfigFile(i.config.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Detect markers
	markers := detectMarkers(lines, i.config.StartMarker, i.config.EndMarker)

	// Build new config with unmanaged sections
	var newLines []string

	// Add above section (before managed)
	if aboveContent != "" {
		reporter.ReportStatus("Writing personal profiles (above)...")
		aboveLines := strings.Split(strings.TrimSpace(aboveContent), "\n")
		newLines = append(newLines, aboveLines...)
		if markers.Found {
			newLines = append(newLines, "") // Blank line before managed
		}
	}

	// Add managed section (if exists)
	if markers.Found {
		newLines = append(newLines, lines[markers.StartLine:markers.EndLine+1]...)
	}

	// Add below section (after managed)
	if belowContent != "" {
		reporter.ReportStatus("Writing personal profiles (below)...")
		if markers.Found {
			newLines = append(newLines, "") // Blank line after managed
		}
		belowLines := strings.Split(strings.TrimSpace(belowContent), "\n")
		newLines = append(newLines, belowLines...)
	}

	// Write final config
	reporter.ReportStatus("Writing config file...")
	if err := writeConfigFile(i.config.ConfigPath, newLines); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	logging.Debug.Log("writeUnmanagedContent completed")
	return nil
}
