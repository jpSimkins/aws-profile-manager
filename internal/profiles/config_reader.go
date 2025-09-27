package profiles

import (
	"context"
	"fmt"
	"os"
	"strings"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// configReader handles reading and parsing AWS CLI config files.
//
// This is an internal component used by Exporter.
// It wraps awscli.Extractor and provides logic to:
//   - Detect managed section markers
//   - Extract specific sections (managed, above, below)
//   - Convert flat AWS CLI profiles back to hierarchical schema
type configReader struct {
	config Config
}

// newConfigReader creates a new configReader instance.
//
// Parameters:
//   - config: Configuration injected by component (contains paths and markers)
//
// Returns:
//   - *configReader: Reader instance ready for use
func newConfigReader(config Config) *configReader {
	return &configReader{
		config: config,
	}
}

// readConfig reads and parses the AWS CLI config file.
//
// This reads the entire file and returns all sections based on the provided options.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Export options controlling which sections to extract
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - *schema.Schema: Schema with requested sections populated
//   - *ExportStats: Statistics about extracted profiles
//   - error: Any error encountered during read or parse
func (r *configReader) readConfig(ctx context.Context, opts ExportOptions, reporter task.Reporter) (*schema.Schema, *ExportStats, error) {
	logging.Debug.Log("readConfig started", "path", r.config.ConfigPath)

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	reporter.ReportStatus("Reading AWS config file...")

	// Read file content
	content, err := readFileContent(r.config.ConfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Detect markers
	lines := strings.Split(content, "\n")
	markers := detectMarkers(lines, r.config.StartMarker, r.config.EndMarker)

	logging.Debug.Log("Markers detected",
		"found", markers.Found,
		"start", markers.StartLine,
		"end", markers.EndLine,
	)

	// Create result schema and stats
	result := &schema.Schema{
		Version: schema.CurrentSchemaVersion,
	}
	stats := &ExportStats{}

	// Extract based on options
	if opts.IncludeManaged {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		reporter.ReportStatus("Extracting managed profiles...")

		if markers.Found {
			result.Managed = r.extractManagedSection(content, markers)

			// Use generators to get accurate counts (single source of truth)
			reporter.ReportStatus("Counting managed profiles...")

			// Run generators to count (discard content, keep stats)
			if result.Managed != nil {
				_, ssoStats, _ := generators.GenerateSsoProfiles(ctx, result.Managed, task.NoOpReporter{})
				_, iamStats, _ := generators.GenerateIamProfiles(ctx, result.Managed, task.NoOpReporter{})
				_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, result.Managed, task.NoOpReporter{})
				_, genericStats, _ := generators.GenerateGenericProfiles(ctx, result.Managed, task.NoOpReporter{})

				// Aggregate stats
				stats.ManagedProfiles = ssoStats.ProfilesWritten + iamStats.ProfilesWritten +
					assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten
				stats.SsoSessions = ssoStats.SessionsWritten
				stats.SsoProfiles = ssoStats.SsoProfiles
				stats.IamProfiles = iamStats.IamProfiles
				stats.AssumeRoleProfiles = assumeRoleStats.AssumeRoleProfiles
				stats.GenericProfiles = genericStats.GenericProfiles
			}
		} else {
			// No markers = no managed section to extract
			logging.Debug.Log("No markers found - no managed section to extract")
			result.Managed = &schema.ProfileCollection{}
		}
	}

	if opts.IncludeAbove || opts.IncludeBelow {
		// Check for cancellation
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		reporter.ReportStatus("Extracting personal profiles...")

		unmanaged := r.extractUnmanagedSections(content, markers, opts.IncludeAbove, opts.IncludeBelow)
		result.Unmanaged = unmanaged

		// Use generators to count unmanaged profiles (single source of truth)
		reporter.ReportStatus("Counting personal profiles...")

		if unmanaged != nil {
			if unmanaged.Above != nil {
				_, ssoStats, _ := generators.GenerateSsoProfiles(ctx, unmanaged.Above, task.NoOpReporter{})
				_, iamStats, _ := generators.GenerateIamProfiles(ctx, unmanaged.Above, task.NoOpReporter{})
				_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, unmanaged.Above, task.NoOpReporter{})
				_, genericStats, _ := generators.GenerateGenericProfiles(ctx, unmanaged.Above, task.NoOpReporter{})
				stats.UnmanagedAbove = ssoStats.ProfilesWritten + iamStats.ProfilesWritten +
					assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten
			}
			if unmanaged.Below != nil {
				_, ssoStats, _ := generators.GenerateSsoProfiles(ctx, unmanaged.Below, task.NoOpReporter{})
				_, iamStats, _ := generators.GenerateIamProfiles(ctx, unmanaged.Below, task.NoOpReporter{})
				_, assumeRoleStats, _ := generators.GenerateAssumeRoleProfiles(ctx, unmanaged.Below, task.NoOpReporter{})
				_, genericStats, _ := generators.GenerateGenericProfiles(ctx, unmanaged.Below, task.NoOpReporter{})
				stats.UnmanagedBelow = ssoStats.ProfilesWritten + iamStats.ProfilesWritten +
					assumeRoleStats.ProfilesWritten + genericStats.ProfilesWritten
			}
		}
	}

	// Calculate total
	stats.TotalProfiles = stats.ManagedProfiles + stats.UnmanagedAbove + stats.UnmanagedBelow

	return result, stats, nil
}

// extractManagedSection extracts profiles from the managed section.
//
// Parameters:
//   - content: Full config file content
//   - markers: Marker position information
//
// Returns:
//   - *schema.ProfileCollection: Profiles from managed section
func (r *configReader) extractManagedSection(content string, markers markerPosition) *schema.ProfileCollection {
	if !markers.Found {
		return &schema.ProfileCollection{}
	}

	lines := strings.Split(content, "\n")

	// Extract lines between markers (exclusive of marker lines)
	if markers.StartLine >= 0 && markers.EndLine >= 0 && markers.StartLine < markers.EndLine {
		managedLines := lines[markers.StartLine+1 : markers.EndLine]
		managedContent := strings.Join(managedLines, "\n")

		logging.Debug.Log("Extracting managed section",
			"start_line", markers.StartLine,
			"end_line", markers.EndLine,
			"lines", len(managedLines),
		)

		// Parse managed section with SSO sessions from full file
		return r.parseSectionWithSessions(managedContent, content)
	}

	return &schema.ProfileCollection{}
}

// extractUnmanagedSections extracts personal profile sections from outside managed markers.
//
// Parameters:
//   - content: Full config file content
//   - markers: Marker position information
//   - includeAbove: Whether to extract Above section
//   - includeBelow: Whether to extract Below section
//
// Returns:
//   - *schema.UnmanagedProfiles: Personal profiles from above and/or below sections
func (r *configReader) extractUnmanagedSections(
	content string,
	markers markerPosition,
	includeAbove, includeBelow bool,
) *schema.UnmanagedProfiles {
	lines := strings.Split(content, "\n")
	unmanaged := &schema.UnmanagedProfiles{}

	// Extract above managed section
	if includeAbove {
		if !markers.Found {
			// No markers = everything is above
			unmanaged.Above = r.parseSection(content)
		} else if markers.StartLine > 0 {
			aboveLines := lines[:markers.StartLine]
			aboveContent := strings.Join(aboveLines, "\n")

			logging.Debug.Log("Extracting above section", "lines", len(aboveLines))
			unmanaged.Above = r.parseSection(aboveContent)
		}
	}

	// Extract below managed section
	if includeBelow && markers.Found && markers.EndLine >= 0 && markers.EndLine < len(lines)-1 {
		belowLines := lines[markers.EndLine+1:]
		belowContent := strings.Join(belowLines, "\n")

		logging.Debug.Log("Extracting below section", "lines", len(belowLines))
		unmanaged.Below = r.parseSection(belowContent)
	}

	return unmanaged
}

// parseSection parses a config section and converts it to ProfileCollection.
//
// This is a convenience wrapper for cases where the section content
// and full content are the same (no external SSO sessions to look up).
//
// Parameters:
//   - content: Config section content to parse
//
// Returns:
//   - *schema.ProfileCollection: Parsed profiles
func (r *configReader) parseSection(content string) *schema.ProfileCollection {
	return r.parseSectionWithSessions(content, content)
}

// parseSectionWithSessions parses a config section with SSO session lookup.
//
// This handles the case where SSO profiles in one section reference SSO
// sessions defined elsewhere in the file.
//
// Parameters:
//   - profileContent: Config section content to extract profiles from
//   - fullContent: Full config file content for SSO session lookup
//
// Returns:
//   - *schema.ProfileCollection: Parsed profiles with SSO sessions resolved (empty on error)
func (r *configReader) parseSectionWithSessions(profileContent, fullContent string) *schema.ProfileCollection {
	if strings.TrimSpace(profileContent) == "" {
		return &schema.ProfileCollection{}
	}

	// Parse profiles from the section
	profileData, err := r.parseContentToExtractedData(profileContent)
	if err != nil {
		logging.Debug.Log("Failed to parse profile content", "error", err)
		return &schema.ProfileCollection{}
	}

	// Parse sessions from the full content
	sessionData, err := r.parseContentToExtractedData(fullContent)
	if err != nil {
		logging.Debug.Log("Failed to parse session content", "error", err)
		// Continue with profiles even if session parsing fails
		sessionData = &awscli.ExtractedData{
			SsoSessions: []awscli.SsoSession{},
		}
	}

	// Combine: use profiles from section, sessions from full content
	profileData.SsoSessions = sessionData.SsoSessions

	// Convert to ProfileCollection
	return r.convertToProfileCollection(profileData)
}

// parseContentToExtractedData parses config content using awscli.Extractor.
//
// This wraps the awscli.Extractor to parse AWS CLI config content.
// Creates a temporary file for the extractor (which expects a file path).
//
// Parameters:
//   - content: AWS CLI config content to parse
//
// Returns:
//   - *awscli.ExtractedData: Parsed profiles and SSO sessions (empty if content empty)
//   - error: Any error during parsing
func (r *configReader) parseContentToExtractedData(content string) (*awscli.ExtractedData, error) {
	if strings.TrimSpace(content) == "" {
		return &awscli.ExtractedData{
			Profiles:    []awscli.AwsCliProfile{},
			SsoSessions: []awscli.SsoSession{},
		}, nil
	}

	// Write content to a temporary file for awscli.Extractor
	tmpFile, err := os.CreateTemp("", "aws-config-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for parsing: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.WriteString(content); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	_ = tmpFile.Close()

	// Use awscli.Extractor to parse
	extractor := awscli.NewExtractorWithPath(tmpFile.Name())
	data, err := extractor.ExtractFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to parse config content: %w", err)
	}

	return data, nil
}

// convertToProfileCollection converts awscli.ExtractedData to schema.ProfileCollection.
//
// This performs the complex task of converting flat AWS CLI profile structure
// into hierarchical schema structure. For SSO profiles, it reconstructs the
// organization/partition/account/role hierarchy.
//
// Parameters:
//   - data: Extracted AWS CLI profiles and sessions
//
// Returns:
//   - *schema.ProfileCollection: Hierarchical profile collection
func (r *configReader) convertToProfileCollection(data *awscli.ExtractedData) *schema.ProfileCollection {
	pc := &schema.ProfileCollection{
		Organizations: make(map[string]*schema.Organization),
	}

	// Group SSO profiles by session for organization reconstruction
	ssoProfilesBySession := make(map[string][]awscli.AwsCliProfile)

	// Convert profiles by type
	for _, profile := range data.Profiles {
		switch profile.Type {
		case awscli.ProfileTypeIAM:
			pc.IamUsers = append(pc.IamUsers, r.convertIamProfile(profile))

		case awscli.ProfileTypeAssumeRole:
			pc.AssumeRoleChains = append(pc.AssumeRoleChains, r.convertAssumeRoleProfile(profile))

		case awscli.ProfileTypeSSO:
			// Group SSO profiles by session for organization reconstruction
			if profile.SsoSession != "" {
				ssoProfilesBySession[profile.SsoSession] = append(ssoProfilesBySession[profile.SsoSession], profile)
			} else {
				// SSO profile without session - treat as generic
				pc.GenericProfiles = append(pc.GenericProfiles, r.convertGenericProfile(profile))
			}

		case awscli.ProfileTypeUnknown:
			fallthrough
		default:
			pc.GenericProfiles = append(pc.GenericProfiles, r.convertGenericProfile(profile))
		}
	}

	// Reconstruct SSO organizations from grouped profiles
	if len(ssoProfilesBySession) > 0 {
		// Convert sessions slice to map for easier lookup
		sessionsMap := make(map[string]awscli.SsoSession)
		for _, session := range data.SsoSessions {
			sessionsMap[session.Name] = session
		}
		r.reconstructSsoOrganizations(pc, ssoProfilesBySession, sessionsMap)
	}

	logging.Debug.Log("Converted profiles",
		"organizations", len(pc.Organizations),
		"iam", len(pc.IamUsers),
		"assume_role", len(pc.AssumeRoleChains),
		"generic", len(pc.GenericProfiles),
	)

	return pc
}

// convertIamProfile converts an AWS CLI IAM profile to schema.IamUser format.
func (r *configReader) convertIamProfile(profile awscli.AwsCliProfile) *schema.IamUser {
	iam := &schema.IamUser{
		ProfileName: profile.Name,
		Region:      profile.Region,
	}

	// Copy static credentials if present
	if accessKey, ok := profile.Properties["aws_access_key_id"]; ok {
		iam.AwsAccessKeyID = accessKey
	}
	if secretKey, ok := profile.Properties["aws_secret_access_key"]; ok {
		iam.AwsSecretKey = secretKey
	}

	if profile.HasCredentialProc {
		iam.CredentialProcess = profile.CredentialProcess
	}

	return iam
}

// convertAssumeRoleProfile converts an AWS CLI assume role profile to schema.AssumeRoleChain format.
func (r *configReader) convertAssumeRoleProfile(profile awscli.AwsCliProfile) *schema.AssumeRoleChain {
	assumeRole := &schema.AssumeRoleChain{
		ProfileName:   profile.Name,
		Region:        profile.Region,
		SourceProfile: profile.Properties["source_profile"],
		RoleArn:       profile.Properties["role_arn"],
	}

	if mfa, ok := profile.Properties["mfa_serial"]; ok {
		assumeRole.MfaSerial = mfa
	}

	if extID, ok := profile.Properties["external_id"]; ok {
		assumeRole.ExternalID = extID
	}

	if sessionName, ok := profile.Properties["session_name"]; ok {
		assumeRole.SessionName = sessionName
	}

	return assumeRole
}

// convertGenericProfile converts an AWS CLI profile to schema.GenericProfile format.
func (r *configReader) convertGenericProfile(profile awscli.AwsCliProfile) *schema.GenericProfile {
	return &schema.GenericProfile{
		ProfileName: profile.Name,
		Properties:  profile.Properties,
	}
}

// reconstructSsoOrganizations rebuilds the organization hierarchy from flattened SSO profiles.
//
// This performs the inverse operation of profile generation - takes flat profiles
// and reconstructs the hierarchical organization/partition/account/role structure.
func (r *configReader) reconstructSsoOrganizations(
	pc *schema.ProfileCollection,
	profilesBySession map[string][]awscli.AwsCliProfile,
	sessions map[string]awscli.SsoSession,
) {
	logging.Debug.Log("reconstructSsoOrganizations called", "sessions", len(profilesBySession))

	// Process each SSO session
	for sessionName, profiles := range profilesBySession {
		logging.Debug.Log("Processing SSO session", "session", sessionName, "profiles", len(profiles))

		// Parse session name to extract org alias and partition
		// Format: <org-alias>-<partition>
		orgAlias, partitionName := r.parseSessionName(sessionName)
		if orgAlias == "" || partitionName == "" {
			logging.Log.Warn("Failed to parse session name, skipping",
				"session", sessionName,
			)
			continue
		}

		logging.Debug.Log("Parsed session name",
			"session", sessionName,
			"org_alias", orgAlias,
			"partition", partitionName,
		)

		// Get SSO session details
		session, hasSession := sessions[sessionName]
		if !hasSession {
			logging.Log.Warn("SSO session not found in sessions map",
				"session", sessionName,
			)
			continue
		}

		// Get or create organization
		org, exists := pc.Organizations[orgAlias]
		if !exists {
			// Try to get organization name from session metadata, fallback to alias
			orgName := session.OrganizationName
			if orgName == "" {
				orgName = orgAlias
			}

			org = &schema.Organization{
				Name:        orgName,
				Description: session.Description,
				Partitions:  make(map[string]schema.Partition),
			}
			pc.Organizations[orgAlias] = org
		}

		// Reconstruct partition from profiles
		partition := r.reconstructPartition(profiles, session)
		org.Partitions[partitionName] = partition

		logging.Debug.Log("Reconstructed partition",
			"org", orgAlias,
			"partition", partitionName,
			"accounts", len(partition.Accounts),
			"roles", len(partition.Roles),
			"regions", len(partition.Regions),
		)
	}
}

// parseSessionName extracts organization alias and partition name from SSO session name.
//
// The session name follows the convention: <org-alias>-<partition>
// where partition is always a single word at the end (commercial or govcloud).
func (r *configReader) parseSessionName(sessionName string) (orgAlias, partition string) {
	// Find last dash - partition is always a single word (commercial/govcloud)
	lastDash := strings.LastIndex(sessionName, "-")
	if lastDash == -1 {
		return "", ""
	}

	orgAlias = sessionName[:lastDash]
	partition = sessionName[lastDash+1:]

	return orgAlias, partition
}

// reconstructPartition rebuilds a partition structure from SSO profiles.
//
// This collects unique accounts, roles, and regions by parsing profile names.
func (r *configReader) reconstructPartition(
	profiles []awscli.AwsCliProfile,
	session awscli.SsoSession,
) schema.Partition {
	partition := schema.Partition{
		URL:           session.StartURL,
		DefaultRegion: session.Region,
		Regions:       []string{},
		Accounts:      []schema.Account{},
		Roles:         []string{},
	}

	// Track unique values
	accountsMap := make(map[string]schema.Account) // Key: account ID
	rolesSet := make(map[string]bool)              // Unique roles
	regionsSet := make(map[string]bool)            // Unique regions

	// Parse each profile
	for _, profile := range profiles {
		// Parse profile name
		parsedProfile := r.parseProfileName(profile.Name)
		if parsedProfile == nil {
			logging.Log.Warn("Failed to parse profile name",
				"profile", profile.Name,
			)
			continue
		}

		logging.Debug.Log("Parsed profile",
			"profile", profile.Name,
			"account_alias", parsedProfile.AccountAlias,
			"role", parsedProfile.Role,
			"region_from_name", parsedProfile.Region,
			"region_from_config", profile.Region,
		)

		// Add account (deduplicated by ID)
		if profile.AccountID != "" {
			if _, exists := accountsMap[profile.AccountID]; !exists {
				// Try to get account name from profile metadata, fallback to alias
				accountName := profile.AccountName
				if accountName == "" {
					accountName = parsedProfile.AccountAlias
				}

				accountsMap[profile.AccountID] = schema.Account{
					Alias: parsedProfile.AccountAlias,
					Name:  accountName,
					ID:    profile.AccountID,
				}
			}
		}

		// Add role
		if profile.RoleName != "" {
			rolesSet[profile.RoleName] = true
		}

		// Add region
		if profile.Region != "" {
			regionsSet[profile.Region] = true
		}
	}

	// Convert maps to slices
	for _, account := range accountsMap {
		partition.Accounts = append(partition.Accounts, account)
	}

	for role := range rolesSet {
		partition.Roles = append(partition.Roles, role)
	}

	for region := range regionsSet {
		partition.Regions = append(partition.Regions, region)
	}

	return partition
}

// parsedProfile holds the parsed components extracted from an AWS CLI profile name.
type parsedProfile struct {
	Partition    string // Partition name (commercial or govcloud)
	AccountAlias string // Account alias (may contain dashes)
	Role         string // Role name (single word)
	Region       string // Region (empty string if using default region)
}

// parseProfileName extracts components from an AWS CLI profile name.
//
// Profile Name Format:
//   - Default region: <partition>-<account-alias>-<role>
//   - Specific region: <partition>-<account-alias>-<role>--<region>
func (r *configReader) parseProfileName(profileName string) *parsedProfile {
	// Check for region suffix (double dash)
	var region string
	baseName := profileName

	if idx := strings.Index(profileName, "--"); idx != -1 {
		baseName = profileName[:idx]
		region = profileName[idx+2:]
	}

	// Split by single dash
	parts := strings.Split(baseName, "-")
	if len(parts) < 3 {
		return nil
	}

	// First part is partition
	partition := parts[0]

	// Last part is role
	role := parts[len(parts)-1]

	// Middle parts form account alias (may contain dashes)
	accountAlias := strings.Join(parts[1:len(parts)-1], "-")

	return &parsedProfile{
		Partition:    partition,
		AccountAlias: accountAlias,
		Role:         role,
		Region:       region,
	}
}
