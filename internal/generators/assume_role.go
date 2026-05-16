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

// GenerateAssumeRoleProfiles generates AssumeRole profiles from ProfileCollection.
//
// This function is a pure content generator that converts schema AssumeRole chains
// into properly formatted AWS CLI config text for role assumption.
//
// Process:
//  1. Iterate through all AssumeRole chains in the collection
//  2. Generate profile block for each chain
//  3. Include required fields (source_profile, role_arn)
//  4. Include optional fields (mfa_serial, external_id, etc.)
//  5. Return formatted content and statistics
//
// Profile Format:
//
//	[profile <profile-name>]
//	source_profile = <source-profile>
//	role_arn = arn:aws:iam::123456789012:role/MyRole
//	mfa_serial = arn:aws:iam::123456789012:mfa/user (optional)
//	external_id = external-id-value (optional)
//	role_session_name = my-session (optional)
//	region = us-east-1 (optional)
//
// Parameters:
//   - ctx: Context for cancellation
//   - profiles: ProfileCollection with AssumeRoleChains to generate from
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - string: Complete AWS CLI config content for AssumeRole profiles
//   - *SectionStats: Statistics (profiles written)
//   - error: Context cancellation or other errors
//
// Example:
//
//	collection := schema.ProfileCollection{AssumeRoleChains: chains}
//	content, stats, err := GenerateAssumeRoleProfiles(ctx, collection, reporter)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Generated %d AssumeRole profiles\n", stats.ProfilesWritten)
func GenerateAssumeRoleProfiles(ctx context.Context, profiles *schema.ProfileCollection, reporter task.Reporter) (string, *SectionStats, error) {
	logging.Debug.Log("GenerateAssumeRoleProfiles called", "chainCount", len(profiles.AssumeRoleChains))

	stats := &SectionStats{}
	var content strings.Builder

	// Iterate through assume role chains
	for _, chain := range profiles.AssumeRoleChains {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logging.Debug.Log("AssumeRole generation cancelled", "error", err)
			return "", stats, err
		}

		// Report progress
		reporter.ReportStatus(fmt.Sprintf("Generating AssumeRole profile for %s...", chain.ProfileName))

		// Throttle for manual testing of cancellation (only if AWS_PROFILE_MANAGER_THROTTLE is set)
		if err := core.Throttle(ctx); err != nil {
			logging.Debug.Log("AssumeRole generation cancelled during throttle", "error", err)
			return "", stats, err
		}

		logging.Debug.Log("Processing AssumeRole chain",
			"profileName", chain.ProfileName,
			"sourceProfile", chain.SourceProfile,
			"roleArn", chain.RoleArn,
		)

		// Generate profile
		profileContent := generateAssumeRoleProfile(chain)
		content.WriteString(profileContent)
		content.WriteString("\n")
		stats.ProfilesWritten++

		logging.Debug.Log("Wrote AssumeRole profile", "profile", chain.ProfileName)
	}

	// Set type-specific count
	stats.AssumeRoleProfiles = stats.ProfilesWritten

	logging.Debug.Log("GenerateAssumeRoleProfiles completed",
		"profilesWritten", stats.ProfilesWritten,
		"assumeRoleProfiles", stats.AssumeRoleProfiles,
	)

	return content.String(), stats, nil
}

// generateAssumeRoleProfile generates a single AssumeRole profile configuration block.
//
// This helper function creates the AWS CLI config text for one AssumeRole profile,
// including all required and optional fields.
//
// Parameters:
//   - chain: AssumeRole chain configuration to generate profile from
//
// Returns:
//   - string: Formatted AssumeRole profile config block
func generateAssumeRoleProfile(chain *schema.AssumeRoleChain) string {
	logging.Debug.Log("generateAssumeRoleProfile called", "profile", chain.ProfileName)

	var content strings.Builder

	// Profile header
	content.WriteString(fmt.Sprintf("[profile %s]\n", chain.ProfileName))

	// Required fields
	content.WriteString(fmt.Sprintf("source_profile = %s\n", chain.SourceProfile))
	content.WriteString(fmt.Sprintf("role_arn = %s\n", chain.RoleArn))

	// Optional fields
	if chain.MfaSerial != "" {
		content.WriteString(fmt.Sprintf("mfa_serial = %s\n", chain.MfaSerial))
	}

	if chain.ExternalID != "" {
		content.WriteString(fmt.Sprintf("external_id = %s\n", chain.ExternalID))
	}

	if chain.SessionName != "" {
		content.WriteString(fmt.Sprintf("role_session_name = %s\n", chain.SessionName))
	}

	if chain.Region != "" {
		content.WriteString(fmt.Sprintf("region = %s\n", chain.Region))
	}

	return content.String()
}
