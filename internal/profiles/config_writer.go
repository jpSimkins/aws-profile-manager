package profiles

import (
	"context"
	"fmt"
	"strings"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// configWriter handles writing profiles to AWS CLI config file.
//
// This is an internal component used by Installer and Importer.
// It manages the three-section structure of AWS CLI config file:
//  1. Content above managed section (user's personal profiles/settings)
//  2. Managed section (between markers - completely replaced on each write)
//  3. Content below managed section (more user content)
//
// The writer delegates content generation to the generators package and
// handles all file I/O operations with proper error handling and change detection.
type configWriter struct {
	config Config
}

// newConfigWriter creates a new configWriter instance.
//
// Parameters:
//   - config: Configuration injected by component (contains paths and markers)
//
// Returns:
//   - *configWriter: Writer instance ready for use
func newConfigWriter(config Config) *configWriter {
	return &configWriter{
		config: config,
	}
}

// writeConfig writes profiles from schema to AWS CLI config file.
//
// This orchestrates the complete write process:
//  1. Reads existing AWS config file (if exists)
//  2. Detects managed section markers (start/end positions)
//  3. Generates managed section content (delegates to generators)
//  4. Builds final config (above + managed + below sections)
//  5. Writes to file (skips if no changes detected)
//
// The reporter provides detailed progress updates for each step.
// The context enables cancellation before expensive operations.
//
// Parameters:
//   - ctx: Context for cancellation
//   - s: Schema containing profiles to write
//   - reporter: Progress reporter for detailed status updates
//
// Returns:
//   - profilesWritten: Number of profiles written
//   - sessionsWritten: Number of SSO sessions written
//   - stats: Detailed section statistics by profile type
//   - createdMarkers: Whether markers were created (vs replaced)
//   - error: Any error during writing
func (w *configWriter) writeConfig(
	ctx context.Context,
	s *schema.Schema,
	reporter task.Reporter,
) (int, int, sectionStats, bool, error) {
	logging.Debug.Log("writeConfig started", "config", w.config.ConfigPath)

	// Validate inputs
	if s == nil {
		return 0, 0, sectionStats{}, false, fmt.Errorf("schema is required")
	}
	if s.Managed == nil {
		return 0, 0, sectionStats{}, false, fmt.Errorf("schema.Managed is required")
	}

	// Check for cancellation before starting
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("writeConfig cancelled before start")
		return 0, 0, sectionStats{}, false, err
	}

	// Step 1: Read existing config file
	reporter.ReportStatus("Reading existing AWS config file...")
	logging.Debug.Log("Reading config file", "path", w.config.ConfigPath)

	existingLines, err := readConfigFile(w.config.ConfigPath)
	if err != nil {
		return 0, 0, sectionStats{}, false, fmt.Errorf("failed to read config file: %w", err)
	}

	logging.Debug.Log("Config file read",
		"lines", len(existingLines),
		"existed", len(existingLines) > 0,
	)

	// Step 2: Detect existing markers
	reporter.ReportStatus("Detecting managed section markers...")
	logging.Debug.Log("Detecting markers")

	markers := detectMarkers(existingLines, w.config.StartMarker, w.config.EndMarker)
	createdMarkers := !markers.Found

	logging.Debug.Log("Markers detected",
		"start", markers.StartLine,
		"end", markers.EndLine,
		"found", markers.Found,
	)

	// Check for cancellation before expensive generation
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("writeConfig cancelled before generation")
		return 0, 0, sectionStats{}, false, err
	}

	// Step 3: Generate managed section content
	reporter.ReportStatus("Generating profile content...")
	logging.Debug.Log("Generating managed section")

	managedContent, stats, err := w.generateManagedSection(ctx, s.Managed, reporter)
	if err != nil {
		// Don't wrap cancellation errors - pass them through
		if err == context.Canceled {
			return 0, 0, sectionStats{}, false, err
		}
		return 0, 0, sectionStats{}, false, fmt.Errorf("failed to generate managed section: %w", err)
	}

	logging.Debug.Log("Content generated",
		"total_profiles", stats.TotalProfiles,
		"sso_sessions", stats.SsoSessions,
		"sso_profiles", stats.SsoProfiles,
		"iam_profiles", stats.IamProfiles,
		"assume_role_profiles", stats.AssumeRoleProfiles,
		"generic_profiles", stats.GenericProfiles,
	)

	// Step 4: Build final config
	reporter.ReportStatus("Building final configuration...")
	logging.Debug.Log("Building final config")

	finalLines := buildFinalConfig(
		existingLines,
		managedContent,
		markers,
		w.config.StartMarker,
		w.config.EndMarker,
		w.config,
	)

	// Step 5: Check if content changed
	contentChanged := hasContentChanged(existingLines, finalLines)
	logging.Debug.Log("Content change detected", "changed", contentChanged)

	if !contentChanged {
		reporter.ReportStatus("No changes detected, skipping write")
		logging.Debug.Log("writeConfig skipped - no changes")
		return stats.TotalProfiles, stats.SsoSessions, *stats, createdMarkers, nil
	}

	// Check for cancellation before file write
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("writeConfig cancelled before file write")
		return 0, 0, sectionStats{}, false, err
	}

	// Step 6: Write to file
	reporter.ReportStatus(fmt.Sprintf("Writing %d profiles to config file...", stats.TotalProfiles))
	logging.Debug.Log("Writing config file", "lines", len(finalLines))

	if err := writeConfigFile(w.config.ConfigPath, finalLines); err != nil {
		return 0, 0, sectionStats{}, false, fmt.Errorf("failed to write config file: %w", err)
	}

	reporter.ReportStatus("Config file written successfully")
	logging.Debug.Log("writeConfig completed",
		"profiles", stats.TotalProfiles,
		"sessions", stats.SsoSessions,
		"created_markers", createdMarkers,
	)

	return stats.TotalProfiles, stats.SsoSessions, *stats, createdMarkers, nil
}

// writeConfigWithContent writes pre-generated content to AWS CLI config file.
//
// This is used by ExecuteImport to write cached content from PrepareImport
// without re-generating. It follows the same flow as writeConfig but skips
// the generation step since content is already provided.
//
// Parameters:
//   - ctx: Context for cancellation
//   - content: Pre-generated managed section content
//   - profileCount: Number of profiles in content (for reporting)
//   - reporter: Progress reporter
//
// Returns:
//   - createdMarkers: Whether markers were created (vs replaced)
//   - error: Any error during writing
func (w *configWriter) writeConfigWithContent(
	ctx context.Context,
	content string,
	profileCount int,
	reporter task.Reporter,
) (bool, error) {
	logging.Debug.Log("writeConfigWithContent started", "profiles", profileCount)

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return false, err
	}

	// Step 1: Read existing config file
	reporter.ReportStatus("Reading existing AWS config file...")

	existingLines, err := readConfigFile(w.config.ConfigPath)
	if err != nil {
		return false, fmt.Errorf("failed to read config file: %w", err)
	}

	// Step 2: Detect existing markers
	reporter.ReportStatus("Detecting managed section markers...")

	markers := detectMarkers(existingLines, w.config.StartMarker, w.config.EndMarker)
	createdMarkers := !markers.Found

	// Step 3: Build final config with pre-generated content
	reporter.ReportStatus("Building final configuration...")

	finalLines := buildFinalConfig(
		existingLines,
		content,
		markers,
		w.config.StartMarker,
		w.config.EndMarker,
		w.config,
	)

	// Step 4: Write to file
	reporter.ReportStatus(fmt.Sprintf("Writing %d profiles to config file...", profileCount))

	if err := writeConfigFile(w.config.ConfigPath, finalLines); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}

	reporter.ReportStatus("Config file written successfully")
	logging.Debug.Log("writeConfigWithContent completed",
		"profiles", profileCount,
		"created_markers", createdMarkers,
	)

	return createdMarkers, nil
}

// generateManagedSection generates content for the managed section.
//
// This delegates to generators package for each profile type and combines
// the output into a single string.
//
// Parameters:
//   - ctx: Context for cancellation
//   - profiles: ProfileCollection to generate from
//   - reporter: Progress reporter for detailed updates
//
// Returns:
//   - content: Complete managed section content (without markers)
//   - stats: Statistics about generated content
//   - error: Any error during generation (including cancellation)
func (w *configWriter) generateManagedSection(
	ctx context.Context,
	profiles *schema.ProfileCollection,
	reporter task.Reporter,
) (string, *sectionStats, error) {
	logging.Debug.Log("generateManagedSection started")

	stats := &sectionStats{}
	var content strings.Builder

	// SSO profiles
	if len(profiles.Organizations) > 0 {
		orgCount := len(profiles.Organizations)
		reporter.ReportStatus(fmt.Sprintf("Generating SSO profiles for %d organizations...", orgCount))
		logging.Debug.Log("Generating SSO profiles", "organizations", orgCount)

		ssoContent, ssoStats, err := generators.GenerateSsoProfiles(ctx, profiles, reporter)
		if err != nil {
			return "", nil, err
		}
		content.WriteString(ssoContent)
		stats.SsoProfiles = ssoStats.ProfilesWritten
		stats.SsoSessions = ssoStats.SessionsWritten
		stats.TotalProfiles += ssoStats.ProfilesWritten
		stats.OrganizationCount = ssoStats.OrganizationCount
		stats.PartitionCount = ssoStats.PartitionCount
		stats.AccountCount = ssoStats.AccountCount
		stats.RoleCount = ssoStats.RoleCount
		stats.RegionCount = ssoStats.RegionCount
	}

	// IAM profiles
	if len(profiles.IamUsers) > 0 {
		userCount := len(profiles.IamUsers)
		reporter.ReportStatus(fmt.Sprintf("Generating IAM profiles for %d users...", userCount))
		logging.Debug.Log("Generating IAM profiles", "users", userCount)

		iamContent, iamStats, err := generators.GenerateIamProfiles(ctx, profiles, reporter)
		if err != nil {
			return "", nil, err
		}
		content.WriteString(iamContent)
		stats.IamProfiles = iamStats.ProfilesWritten
		stats.TotalProfiles += iamStats.ProfilesWritten
	}

	// AssumeRole profiles
	if len(profiles.AssumeRoleChains) > 0 {
		chainCount := len(profiles.AssumeRoleChains)
		reporter.ReportStatus(fmt.Sprintf("Generating AssumeRole profiles for %d chains...", chainCount))
		logging.Debug.Log("Generating AssumeRole profiles", "chains", chainCount)

		assumeRoleContent, assumeRoleStats, err := generators.GenerateAssumeRoleProfiles(ctx, profiles, reporter)
		if err != nil {
			return "", nil, err
		}
		content.WriteString(assumeRoleContent)
		stats.AssumeRoleProfiles = assumeRoleStats.ProfilesWritten
		stats.TotalProfiles += assumeRoleStats.ProfilesWritten
	}

	// Generic profiles
	if len(profiles.GenericProfiles) > 0 {
		profileCount := len(profiles.GenericProfiles)
		reporter.ReportStatus(fmt.Sprintf("Generating %d generic profiles...", profileCount))
		logging.Debug.Log("Generating generic profiles", "profiles", profileCount)

		genericContent, genericStats, err := generators.GenerateGenericProfiles(ctx, profiles, reporter)
		if err != nil {
			return "", nil, err
		}
		content.WriteString(genericContent)
		stats.GenericProfiles = genericStats.ProfilesWritten
		stats.TotalProfiles += genericStats.ProfilesWritten
	}

	logging.Debug.Log("generateManagedSection completed", "total_profiles", stats.TotalProfiles)

	return content.String(), stats, nil
}
