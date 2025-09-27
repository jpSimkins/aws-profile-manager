package generators

import (
	"context"
	"fmt"
	"strings"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// GenerateSsoProfiles generates AWS CLI config for SSO-based profiles.
//
// Creates profiles and SSO sessions from ProfileCollection. Generates one
// SSO session per organization and profiles for each account/role combination.
//
// Parameters:
//   - ctx: Context for cancellation
//   - profiles: ProfileCollection with Organizations to generate from
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - string: Complete AWS CLI config content for SSO profiles
//   - *SectionStats: Statistics (profiles written, sessions written)
//   - error: Context cancellation or other errors
//
// Example:
//
//	collection := schema.ProfileCollection{Organizations: orgs}
//	content, stats, err := GenerateSsoProfiles(ctx, collection, reporter)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Generated %d profiles and %d sessions\n",
//	    stats.ProfilesWritten, stats.SessionsWritten)
func GenerateSsoProfiles(ctx context.Context, profiles *schema.ProfileCollection, reporter task.Reporter) (string, *SectionStats, error) {
	logging.Debug.Log("GenerateSsoProfiles called", "orgCount", len(profiles.Organizations))

	stats := &SectionStats{}
	var content strings.Builder

	// Track written sessions to avoid duplicates
	writtenSessions := make(map[string]bool)

	// Track unique roles and regions for stats
	uniqueRoles := make(map[string]bool)
	uniqueRegions := make(map[string]bool)

	// Iterate through organizations
	for orgAlias, org := range profiles.Organizations {
		stats.OrganizationCount++
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logging.Debug.Log("SSO generation cancelled", "error", err)
			return "", stats, err
		}

		// Throttle for manual testing of cancellation (only if AWS_PROFILE_MANAGER_THROTTLE is set)
		if err := core.Throttle(ctx); err != nil {
			logging.Debug.Log("SSO generation cancelled during throttle", "error", err)
			return "", stats, err
		}

		logging.Debug.Log("Processing organization",
			"alias", orgAlias,
			"name", org.Name,
			"partitionCount", len(org.Partitions),
		)

		// Iterate through partitions
		for partitionName, partition := range org.Partitions {
			stats.PartitionCount++

			logging.Debug.Log("Processing partition",
				"org", orgAlias,
				"partition", partitionName,
				"accountCount", len(partition.Accounts),
			)

			// Track accounts
			stats.AccountCount += len(partition.Accounts)

			// Track roles
			for _, role := range partition.Roles {
				uniqueRoles[role] = true
			}

			// Track regions (including default)
			uniqueRegions[partition.DefaultRegion] = true
			for _, region := range partition.Regions {
				uniqueRegions[region] = true
			}

			// Generate session name
			sessionName := schema.GenerateSsoSessionName(orgAlias, partitionName)

			// Write SSO session (once per org-partition)
			if !writtenSessions[sessionName] {
				sessionContent := generateSsoSession(sessionName, partition.URL, partition.DefaultRegion, org.Name, org.Description)
				content.WriteString(sessionContent)
				content.WriteString("\n")
				writtenSessions[sessionName] = true
				stats.SessionsWritten++

				logging.Debug.Log("Wrote SSO session", "session", sessionName)
			}

			// Generate profiles for each account + role + region combination
			for _, account := range partition.Accounts {
				// Check for cancellation at account level
				if err := ctx.Err(); err != nil {
					logging.Debug.Log("SSO generation cancelled at account level", "error", err)
					return "", stats, err
				}

				// Report progress
				reporter.ReportStatus(fmt.Sprintf("Generating SSO profiles for %s/%s...", org.Name, account.Name))

				for _, role := range partition.Roles {
					// Default region profile
					profileName := schema.GenerateProfileName(partitionName, account.Alias, role, partition.DefaultRegion, partition.DefaultRegion)
					profileContent := generateSsoProfile(profileName, sessionName, account.ID, role, partition.DefaultRegion, org.Name, account.Name)
					content.WriteString(profileContent)
					content.WriteString("\n")
					stats.ProfilesWritten++

					logging.Debug.Log("Wrote SSO profile (default region)",
						"profile", profileName,
						"account", account.Alias,
						"role", role,
					)

					// Additional regions (if any)
					for _, region := range partition.Regions {
						if region != partition.DefaultRegion {
							profileName := schema.GenerateProfileName(partitionName, account.Alias, role, region, partition.DefaultRegion)
							profileContent := generateSsoProfile(profileName, sessionName, account.ID, role, region, org.Name, account.Name)
							content.WriteString(profileContent)
							content.WriteString("\n")
							stats.ProfilesWritten++

							logging.Debug.Log("Wrote SSO profile (additional region)",
								"profile", profileName,
								"account", account.Alias,
								"role", role,
								"region", region,
							)
						}
					}
				}
			}
		}
	}

	// Set type-specific count
	stats.SsoProfiles = stats.ProfilesWritten

	// Set final unique counts
	stats.RoleCount = len(uniqueRoles)
	stats.RegionCount = len(uniqueRegions)

	logging.Debug.Log("generateSsoProfiles completed",
		"sessionsWritten", stats.SessionsWritten,
		"profilesWritten", stats.ProfilesWritten,
		"ssoProfiles", stats.SsoProfiles,
		"organizationCount", stats.OrganizationCount,
		"partitionCount", stats.PartitionCount,
		"accountCount", stats.AccountCount,
		"roleCount", stats.RoleCount,
		"regionCount", stats.RegionCount,
	)

	return content.String(), stats, nil
}

// generateSsoSession generates a single SSO session configuration block.
//
// This helper function creates the AWS CLI config text for one SSO session,
// including optional metadata comments for organization name and description.
//
// Parameters:
//   - sessionName: Name of the SSO session (e.g., "my-org-commercial")
//   - startUrl: SSO portal URL
//   - region: AWS region for the SSO session
//   - orgName: Organization name (for metadata comment)
//   - description: Organization description (for metadata comment)
//
// Returns:
//   - string: Formatted SSO session config block
func generateSsoSession(sessionName, startUrl, region, orgName, description string) string {
	logging.Debug.Log("generateSsoSession called", "session", sessionName)

	var content strings.Builder

	// Add metadata comments if available
	if orgName != "" {
		content.WriteString(fmt.Sprintf("# Organization: %s\n", orgName))
	}
	if description != "" {
		content.WriteString(fmt.Sprintf("# Description: %s\n", description))
	}

	content.WriteString(fmt.Sprintf(`[sso-session %s]
sso_start_url = %s
sso_region = %s
sso_registration_scopes = sso:account:access
`, sessionName, startUrl, region))

	return content.String()
}

// generateSsoProfile generates a single SSO profile configuration block.
//
// This helper function creates the AWS CLI config text for one SSO profile,
// including optional metadata comments for organization and account names.
//
// Parameters:
//   - profileName: Name of the profile (e.g., "commercial-dev-Developer")
//   - sessionName: SSO session name to reference
//   - accountID: AWS account ID
//   - roleName: IAM role name to assume
//   - region: AWS region for the profile
//   - orgName: Organization name (for metadata comment)
//   - accountName: Account name (for metadata comment)
//
// Returns:
//   - string: Formatted SSO profile config block
func generateSsoProfile(profileName, sessionName, accountID, roleName, region, orgName, accountName string) string {
	logging.Debug.Log("generateSsoProfile called", "profile", profileName)

	var content strings.Builder

	// Add metadata comments if available
	if orgName != "" {
		content.WriteString(fmt.Sprintf("# Organization: %s\n", orgName))
	}
	if accountName != "" {
		content.WriteString(fmt.Sprintf("# Account: %s\n", accountName))
	}

	content.WriteString(fmt.Sprintf(`[profile %s]
sso_session = %s
sso_account_id = %s
sso_role_name = %s
region = %s
`, profileName, sessionName, accountID, roleName, region))

	return content.String()
}
