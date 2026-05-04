package profiles

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// Installer installs AWS CLI profiles from schema.
//
// This component orchestrates the complete installation process including
// validation, filtering, profile generation, and optional cheat sheet creation.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	installer := profiles.NewInstaller(config)
//	result, err := installer.Install(ctx, opts, reporter)
type Installer struct {
	config Config
	writer *configWriter
}

// NewInstaller creates a new Installer with injected configuration.
//
// Parameters:
//   - config: Configuration injected by CLI/GUI (from settings)
//
// Returns:
//   - *Installer: Ready to use installer instance
//
// Example:
//
//	config := buildConfigFromSettings()
//	installer := profiles.NewInstaller(config)
func NewInstaller(config Config) *Installer {
	return &Installer{
		config: config,
		writer: newConfigWriter(config),
	}
}

// Install installs profiles to AWS config.
//
// Installs managed section and optionally generates cheat sheet.
// Applies filters before installation.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Installation options
//   - reporter: Progress reporter
//
// Returns:
//   - *InstallResult: Statistics and paths
//   - error: Any error including context.Canceled
//
// Example:
//
//	result, err := installer.Install(ctx, profiles.InstallOptions{
//	    Schema:             schema,
//	    Roles:              []string{"Developer"},
//	    GenerateCheatSheet: true,
//	}, reporter)
func (i *Installer) Install(
	ctx context.Context,
	opts InstallOptions,
	reporter task.Reporter,
) (*InstallResult, error) {
	startTime := time.Now()
	logging.Debug.Log("Install started")

	// Validate schema
	if err := validateSchema(opts.Schema); err != nil {
		return nil, err
	}

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Install cancelled before start")
		return nil, err
	}

	// Step 1: Apply filters to schema
	reporter.ReportStatus("Applying filters to schema...")
	logging.Debug.Log("Applying filters")

	criteria := schema.FilterCriteria{
		Organizations: opts.Organizations,
		Partitions:    opts.Partitions,
		Accounts:      opts.Accounts,
		Roles:         opts.Roles,
		Regions:       opts.Regions,
		AllRegions:    opts.AllRegions,
	}

	filteredSchema, err := schema.FilterSchema(opts.Schema, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to filter schema: %w", err)
	}

	// Calculate profile counts
	profileCount, sessionCount := calculateProfileCounts(filteredSchema)

	result := &InstallResult{
		TotalProfiles: profileCount,
		SsoSessions:   sessionCount,
		ConfigPath:    i.config.ConfigPath,
		Duration:      time.Since(startTime),
	}

	// Dry run
	if opts.DryRun {
		reporter.ReportStatus(fmt.Sprintf("DRY RUN: Would install %d profiles and %d SSO sessions", profileCount, sessionCount))
		logging.Debug.Log("DRY RUN completed",
			"profiles", profileCount,
			"sessions", sessionCount,
		)

		if opts.GenerateCheatSheet {
			reporter.ReportStatus("DRY RUN: Would generate cheat sheet")
		}

		return result, nil
	}

	// Check for cancellation before expensive operations
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Install cancelled before write")
		return nil, err
	}

	// Step 2: Write profiles to config file unless this is cheat-sheet-only mode.
	if !opts.CheatSheetOnly {
		reporter.ReportStatus("Installing profiles to AWS config...")
		logging.Debug.Log("Writing profiles to config")

		profilesWritten, sessionsWritten, writeStats, createdMarkers, err := i.writer.writeConfig(ctx, filteredSchema, reporter)
		if err != nil {
			return nil, fmt.Errorf("failed to write config: %w", err)
		}

		result.TotalProfiles = profilesWritten
		result.SsoSessions = sessionsWritten
		result.CreatedMarkers = createdMarkers
		result.Duration = time.Since(startTime)
		result.ManagedStats = generators.SectionStats{
			ProfilesWritten:    writeStats.TotalProfiles,
			SessionsWritten:    writeStats.SsoSessions,
			SsoProfiles:        writeStats.SsoProfiles,
			IamProfiles:        writeStats.IamProfiles,
			AssumeRoleProfiles: writeStats.AssumeRoleProfiles,
			GenericProfiles:    writeStats.GenericProfiles,
			OrganizationCount:  writeStats.OrganizationCount,
			PartitionCount:     writeStats.PartitionCount,
			AccountCount:       writeStats.AccountCount,
			RoleCount:          writeStats.RoleCount,
			RegionCount:        writeStats.RegionCount,
		}

		logging.Debug.Log("Profiles written",
			"profiles", profilesWritten,
			"sessions", sessionsWritten,
		)
	}

	// Step 3: Generate cheat sheet if requested
	if opts.GenerateCheatSheet {
		if err := ctx.Err(); err != nil {
			logging.Debug.Log("Install cancelled before cheat sheet")
			return result, err // Return partial result
		}

		reporter.ReportStatus("Generating cheat sheet...")
		logging.Debug.Log("Generating cheat sheet")

		// Determine output path
		cheatSheetPath := opts.CheatSheetPath
		if cheatSheetPath == "" {
			cheatSheetPath = filepath.Join(i.config.CheatSheetOutputDir, "AWS_Profile_Cheat_Sheet.md")
		}

		// Create cheat sheet component with path
		cheatSheetConfig := i.config
		cheatSheet := NewCheatSheet(cheatSheetConfig)

		cheatSheetOpts := CheatSheetOptions{
			Collection: filteredSchema.Managed,
			OutputPath: cheatSheetPath,
		}

		cheatSheetResult, err := cheatSheet.Generate(ctx, cheatSheetOpts, reporter)
		if err != nil {
			return result, fmt.Errorf("failed to generate cheat sheet: %w", err)
		}

		result.CheatSheetPath = cheatSheetResult.OutputPath
		result.Duration = time.Since(startTime)

		logging.Debug.Log("Cheat sheet generated",
			"sessions", cheatSheetResult.Sessions,
			"profiles", cheatSheetResult.Profiles,
			"path", cheatSheetResult.OutputPath,
		)
	}

	reporter.ReportStatus("Installation complete")
	logging.Debug.Log("Install completed",
		"profiles", result.TotalProfiles,
		"sessions", result.SsoSessions,
		"cheat_sheet", result.CheatSheetPath != "",
		"duration", result.Duration,
	)

	return result, nil
}
