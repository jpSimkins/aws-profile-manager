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

// GenerateIamProfiles generates IAM user profiles from ProfileCollection.
//
// This function is a pure content generator that converts schema IAM users
// into properly formatted AWS CLI config text. Supports both static credentials
// and credential_process authentication methods.
//
// Process:
//  1. Iterate through all IAM users in the collection
//  2. Generate profile block for each user
//  3. Include credentials (static or credential_process)
//  4. Return formatted content and statistics
//
// Profile Format (credential_process):
//
//	[profile <profile-name>]
//	credential_process = aws-vault exec my-profile --json
//	region = us-east-1
//
// Profile Format (static credentials):
//
//	[profile <profile-name>]
//	aws_access_key_id = AKIA...
//	aws_secret_access_key = ...
//	region = us-east-1
//
// Parameters:
//   - ctx: Context for cancellation
//   - profiles: ProfileCollection with IamUsers to generate from
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - string: Complete AWS CLI config content for IAM profiles
//   - *SectionStats: Statistics (profiles written)
//   - error: Context cancellation or other errors
//
// Example:
//
//	collection := schema.ProfileCollection{IamUsers: users}
//	content, stats, err := GenerateIamProfiles(ctx, collection, reporter)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Generated %d IAM profiles\n", stats.ProfilesWritten)
func GenerateIamProfiles(ctx context.Context, profiles *schema.ProfileCollection, reporter task.Reporter) (string, *SectionStats, error) {
	logging.Debug.Log("GenerateIamProfiles called", "userCount", len(profiles.IamUsers))

	stats := &SectionStats{}
	var content strings.Builder

	// Iterate through IAM users
	for _, user := range profiles.IamUsers {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logging.Debug.Log("IAM generation cancelled", "error", err)
			return "", stats, err
		}

		// Report progress
		reporter.ReportStatus(fmt.Sprintf("Generating IAM profile for %s...", user.ProfileName))

		// Throttle for manual testing of cancellation (only if AWS_PROFILE_MANAGER_THROTTLE is set)
		if err := core.Throttle(ctx); err != nil {
			logging.Debug.Log("IAM generation cancelled during throttle", "error", err)
			return "", stats, err
		}

		logging.Debug.Log("Processing IAM user",
			"profileName", user.ProfileName,
			"region", user.Region,
			"hasCredentialProcess", user.CredentialProcess != "",
		)

		// Generate profile
		profileContent := generateIamProfile(user)
		content.WriteString(profileContent)
		content.WriteString("\n")
		stats.ProfilesWritten++

		logging.Debug.Log("Wrote IAM profile", "profile", user.ProfileName)
	}

	// Set type-specific count
	stats.IamProfiles = stats.ProfilesWritten

	logging.Debug.Log("GenerateIamProfiles completed",
		"profilesWritten", stats.ProfilesWritten,
		"iamProfiles", stats.IamProfiles,
	)

	return content.String(), stats, nil
}

// generateIamProfile generates a single IAM user profile configuration block.
//
// This helper function creates the AWS CLI config text for one IAM user profile,
// supporting both static credentials and credential_process authentication.
//
// Parameters:
//   - user: IAM user configuration to generate profile from
//
// Returns:
//   - string: Formatted IAM profile config block
func generateIamProfile(user *schema.IamUser) string {
	logging.Debug.Log("generateIamProfile called", "profile", user.ProfileName)

	var content strings.Builder

	// Profile header
	content.WriteString(fmt.Sprintf("[profile %s]\n", user.ProfileName))

	// Static credentials (if specified)
	if user.AwsAccessKeyID != "" {
		content.WriteString(fmt.Sprintf("aws_access_key_id = %s\n", user.AwsAccessKeyID))
	}
	if user.AwsSecretKey != "" {
		content.WriteString(fmt.Sprintf("aws_secret_access_key = %s\n", user.AwsSecretKey))
	}

	// Credential process (if specified)
	if user.CredentialProcess != "" {
		content.WriteString(fmt.Sprintf("credential_process = %s\n", user.CredentialProcess))
	}

	// Region (if specified)
	if user.Region != "" {
		content.WriteString(fmt.Sprintf("region = %s\n", user.Region))
	}

	return content.String()
}
