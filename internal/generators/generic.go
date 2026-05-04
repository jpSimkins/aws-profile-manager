package generators

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// GenerateGenericProfiles generates generic profiles from ProfileCollection.
//
// This function is a pure content generator that converts schema GenericProfiles
// into AWS CLI config text. Generic profiles are a catch-all for profiles that
// don't fit the SSO, IAM, or AssumeRole patterns - they simply output key-value
// pairs as-is.
//
// Use Cases:
//   - Profiles with custom authentication methods
//   - Profiles with non-standard configuration
//   - Preserving existing profiles during migration
//
// Process:
//  1. Iterate through all generic profiles in the collection
//  2. Generate profile block for each profile
//  3. Sort and output all properties as key-value pairs
//  4. Return formatted content and statistics
//
// Profile Format:
//
//	[profile <profile-name>]
//	key1 = value1
//	key2 = value2
//	...
//
// Parameters:
//   - ctx: Context for cancellation
//   - profiles: ProfileCollection with GenericProfiles to generate from
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - string: Complete AWS CLI config content for generic profiles
//   - *SectionStats: Statistics (profiles written)
//   - error: Context cancellation or other errors
//
// Example:
//
//	collection := schema.ProfileCollection{GenericProfiles: profiles}
//	content, stats, err := GenerateGenericProfiles(ctx, collection, reporter)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("Generated %d generic profiles\n", stats.ProfilesWritten)
func GenerateGenericProfiles(ctx context.Context, profiles *schema.ProfileCollection, reporter task.Reporter) (string, *SectionStats, error) {
	logging.Debug.Log("GenerateGenericProfiles called", "profileCount", len(profiles.GenericProfiles))

	stats := &SectionStats{}
	var content strings.Builder

	// Iterate through generic profiles
	for _, profile := range profiles.GenericProfiles {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logging.Debug.Log("Generic generation cancelled", "error", err)
			return "", stats, err
		}

		// Report progress
		reporter.ReportStatus(fmt.Sprintf("Generating generic profile for %s...", profile.ProfileName))

		// Throttle for manual testing of cancellation (only if AWS_PROFILE_MANAGER_THROTTLE is set)
		if err := core.Throttle(ctx); err != nil {
			logging.Debug.Log("Generic generation cancelled during throttle", "error", err)
			return "", stats, err
		}

		logging.Debug.Log("Processing generic profile",
			"profileName", profile.ProfileName,
			"propertyCount", len(profile.Properties),
		)

		// Generate profile
		profileContent := generateGenericProfile(profile)
		content.WriteString(profileContent)
		content.WriteString("\n")
		stats.ProfilesWritten++

		logging.Debug.Log("Wrote generic profile", "profile", profile.ProfileName)
	}

	// Set type-specific count
	stats.GenericProfiles = stats.ProfilesWritten

	logging.Debug.Log("GenerateGenericProfiles completed",
		"profilesWritten", stats.ProfilesWritten,
		"genericProfiles", stats.GenericProfiles,
	)

	return content.String(), stats, nil
}

// generateGenericProfile generates a single generic profile configuration block.
//
// This helper function creates the AWS CLI config text for one generic profile,
// outputting all properties in alphabetically sorted order for consistency.
//
// Parameters:
//   - profile: Generic profile configuration to generate from
//
// Returns:
//   - string: Formatted generic profile config block
func generateGenericProfile(profile *schema.GenericProfile) string {
	logging.Debug.Log("generateGenericProfile called", "profile", profile.ProfileName)

	var content strings.Builder

	// Profile header
	content.WriteString(fmt.Sprintf("[profile %s]\n", profile.ProfileName))

	// Sort keys for consistent output
	keys := make([]string, 0, len(profile.Properties))
	for key := range profile.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Generate properties in sorted order
	for _, key := range keys {
		value := profile.Properties[key]
		content.WriteString(fmt.Sprintf("%s = %s\n", key, value))
	}

	return content.String()
}
