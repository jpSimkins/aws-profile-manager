package backup

import (
	"context"
	"fmt"
	"time"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// ExportProfiles exports AWS CLI profiles and settings to a backup file.
//
// This function orchestrates:
//  1. Calling profiles.Export() for profile extraction
//  2. Getting current settings if requested
//  3. Creating unified backup file
//  4. Writing to disk
//
// Parameters:
//   - ctx: Context for cancellation
//   - cfg: Configuration (paths)
//   - opts: Export options
//   - reporter: Progress reporter (task.Reporter)
//
// Returns:
//   - *ExportResult: Statistics and file path
//   - error: Any error encountered
//
// Example:
//
//	cfg := backup.Config{
//	    ConfigPath: settings.GetAwsConfigPath(),
//	    AwsDir:     settings.GetAwsDir(),
//	}
//	opts := backup.ExportOptions{
//	    OutputPath:     "/path/to/backup.json",
//	    IncludeManaged: true,
//	    IncludeAbove:   true,
//	    IncludeBelow:   true,
//	}
//	result, err := backup.ExportProfiles(ctx, cfg, opts, reporter)
func ExportProfiles(
	ctx context.Context,
	cfg Config,
	opts ExportOptions,
	reporter task.Reporter,
) (*ExportResult, error) {
	startTime := time.Now()

	reporter.ReportStatus("Starting profile export...")

	// Validate options
	if opts.OutputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	// Check if anything to export
	if !opts.IncludeManaged && !opts.IncludeAbove && !opts.IncludeBelow && opts.ExcludeSettings {
		return nil, fmt.Errorf("no content selected for export (all sections disabled)")
	}

	// Step 1: Export profiles using profiles package
	var exportedSchema *schema.Schema
	var profileStats struct {
		managedProfiles int
		unmanagedAbove  int
		unmanagedBelow  int
		totalProfiles   int
		managedStats    generators.SectionStats
		unmanagedStats  generators.SectionStats
	}

	if opts.IncludeManaged || opts.IncludeAbove || opts.IncludeBelow {
		reporter.ReportStatus("Extracting profiles from AWS config...")

		// Build profiles.Config from backup.Config
		profilesConfig := profiles.Config{
			ConfigPath:          cfg.ConfigPath,
			AwsDir:              cfg.AwsDir,
			StartMarker:         cfg.StartMarker,
			EndMarker:           cfg.EndMarker,
			CheatSheetOutputDir: "", // Not needed for export
			CacheDir:            "", // Not needed for export
		}

		// Create exporter
		exporter := profiles.NewExporter(profilesConfig)

		// Build export options for profiles package
		profilesOpts := profiles.ExportOptions{
			OutputPath:     opts.OutputPath, // Temporary - we'll overwrite with BackupFile
			IncludeManaged: opts.IncludeManaged,
			IncludeAbove:   opts.IncludeAbove,
			IncludeBelow:   opts.IncludeBelow,
			Description:    opts.Description,
		}

		// Export profiles
		profilesResult, err := exporter.Export(ctx, profilesOpts, reporter)
		if err != nil {
			// Check if cancelled
			if err == context.Canceled {
				return nil, err // Return cancellation error
			}
			return nil, fmt.Errorf("failed to export profiles: %w", err)
		}

		exportedSchema = profilesResult.Schema
		profileStats.managedProfiles = profilesResult.ManagedProfiles
		profileStats.unmanagedAbove = profilesResult.UnmanagedAbove
		profileStats.unmanagedBelow = profilesResult.UnmanagedBelow
		profileStats.totalProfiles = profilesResult.TotalProfiles

		// Generate detailed stats from exported schema
		if exportedSchema.Managed != nil {
			_, ssoStats, _ := generators.GenerateSsoProfiles(ctx, exportedSchema.Managed, task.NoOpReporter{})
			_, iamStats, _ := generators.GenerateIamProfiles(ctx, exportedSchema.Managed, task.NoOpReporter{})
			_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, exportedSchema.Managed, task.NoOpReporter{})
			_, genericStats, _ := generators.GenerateGenericProfiles(ctx, exportedSchema.Managed, task.NoOpReporter{})

			// Copy SSO stats
			if ssoStats != nil {
				profileStats.managedStats = *ssoStats
			}
			// Add non-SSO profile types
			if iamStats != nil {
				profileStats.managedStats.IamProfiles = iamStats.IamProfiles
			}
			if assumeRoleStats != nil {
				profileStats.managedStats.AssumeRoleProfiles = assumeRoleStats.AssumeRoleProfiles
			}
			if genericStats != nil {
				profileStats.managedStats.GenericProfiles = genericStats.GenericProfiles
			}
		}

		if exportedSchema.Unmanaged != nil {
			// Combine Above and Below stats
			if exportedSchema.Unmanaged.Above != nil {
				_, aboveStats, _ := generators.GenerateSsoProfiles(ctx, exportedSchema.Unmanaged.Above, task.NoOpReporter{})
				_, iamStats, _ := generators.GenerateIamProfiles(ctx, exportedSchema.Unmanaged.Above, task.NoOpReporter{})
				_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, exportedSchema.Unmanaged.Above, task.NoOpReporter{})
				_, genericStats, _ := generators.GenerateGenericProfiles(ctx, exportedSchema.Unmanaged.Above, task.NoOpReporter{})

				if aboveStats != nil {
					profileStats.unmanagedStats.OrganizationCount += aboveStats.OrganizationCount
					profileStats.unmanagedStats.PartitionCount += aboveStats.PartitionCount
					profileStats.unmanagedStats.AccountCount += aboveStats.AccountCount
					profileStats.unmanagedStats.RoleCount += aboveStats.RoleCount
					profileStats.unmanagedStats.RegionCount += aboveStats.RegionCount
					profileStats.unmanagedStats.SsoProfiles += aboveStats.SsoProfiles
					profileStats.unmanagedStats.SessionsWritten += aboveStats.SessionsWritten
				}
				if iamStats != nil {
					profileStats.unmanagedStats.IamProfiles += iamStats.IamProfiles
				}
				if assumeRoleStats != nil {
					profileStats.unmanagedStats.AssumeRoleProfiles += assumeRoleStats.AssumeRoleProfiles
				}
				if genericStats != nil {
					profileStats.unmanagedStats.GenericProfiles += genericStats.GenericProfiles
				}
			}

			if exportedSchema.Unmanaged.Below != nil {
				_, belowStats, _ := generators.GenerateSsoProfiles(ctx, exportedSchema.Unmanaged.Below, task.NoOpReporter{})
				_, iamStats, _ := generators.GenerateIamProfiles(ctx, exportedSchema.Unmanaged.Below, task.NoOpReporter{})
				_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, exportedSchema.Unmanaged.Below, task.NoOpReporter{})
				_, genericStats, _ := generators.GenerateGenericProfiles(ctx, exportedSchema.Unmanaged.Below, task.NoOpReporter{})

				if belowStats != nil {
					profileStats.unmanagedStats.OrganizationCount += belowStats.OrganizationCount
					profileStats.unmanagedStats.PartitionCount += belowStats.PartitionCount
					profileStats.unmanagedStats.AccountCount += belowStats.AccountCount
					profileStats.unmanagedStats.RoleCount += belowStats.RoleCount
					profileStats.unmanagedStats.RegionCount += belowStats.RegionCount
					profileStats.unmanagedStats.SsoProfiles += belowStats.SsoProfiles
					profileStats.unmanagedStats.SessionsWritten += belowStats.SessionsWritten
				}
				if iamStats != nil {
					profileStats.unmanagedStats.IamProfiles += iamStats.IamProfiles
				}
				if assumeRoleStats != nil {
					profileStats.unmanagedStats.AssumeRoleProfiles += assumeRoleStats.AssumeRoleProfiles
				}
				if genericStats != nil {
					profileStats.unmanagedStats.GenericProfiles += genericStats.GenericProfiles
				}
			}
		}

		logging.Log.Info("Profiles extracted",
			"managed", profileStats.managedProfiles,
			"above", profileStats.unmanagedAbove,
			"below", profileStats.unmanagedBelow,
			"total", profileStats.totalProfiles,
		)
	}

	// Step 2: Get settings if requested
	var exportedSettings *settings.Settings
	settingsExported := false

	if !opts.ExcludeSettings {
		reporter.ReportStatus("Including application settings...")
		currentSettings := settings.Get()
		exportedSettings = currentSettings
		settingsExported = true
		logging.Log.Info("Settings included in backup")
	}

	// Step 3: Create unified backup file
	reporter.ReportStatus("Creating backup file...")

	backupFile := &BackupFile{
		Version: "2.0",
		Metadata: BackupMetadata{
			ExportedAt:  time.Now(),
			ToolVersion: core.AppVersion,
			Description: opts.Description,
		},
		Data:     exportedSchema,
		Settings: exportedSettings,
	}

	// Step 4: Write backup file
	reporter.ReportStatus(fmt.Sprintf("Writing backup file to %s...", opts.OutputPath))

	if err := WriteBackupFile(opts.OutputPath, backupFile); err != nil {
		return nil, fmt.Errorf("failed to write backup file: %w", err)
	}

	duration := time.Since(startTime)

	reporter.ReportStatus("Export complete")
	logging.Log.Success("Backup created",
		"path", opts.OutputPath,
		"profiles", profileStats.totalProfiles,
		"settings", settingsExported,
		"duration", duration,
	)

	// Build result
	result := &ExportResult{
		BackupFile:       backupFile,
		OutputPath:       opts.OutputPath,
		ManagedProfiles:  profileStats.managedProfiles,
		UnmanagedAbove:   profileStats.unmanagedAbove,
		UnmanagedBelow:   profileStats.unmanagedBelow,
		TotalProfiles:    profileStats.totalProfiles,
		SettingsExported: settingsExported,
		Duration:         duration,
		ManagedStats:     profileStats.managedStats,
		UnmanagedStats:   profileStats.unmanagedStats,
	}

	return result, nil
}
