package schema

import (
	"fmt"

	"aws-profile-manager/internal/logging"
)

// GenerateSsoSessionName creates an SSO session name from organization and partition.
//
// SSO session names follow a hierarchical naming convention that groups profiles
// by organization and partition (commercial/govcloud). Multiple profiles can share
// the same SSO session, reducing the number of authentication prompts.
//
// Naming Pattern:
//
//	<organization-alias>-<partition>
//
// Parameters:
//   - orgAlias: Organization alias (e.g., "my-org", "company")
//   - partitionName: Partition name (e.g., "commercial", "govcloud")
//
// Returns:
//   - string: SSO session name
//
// Example:
//
//	GenerateSsoSessionName("company-org", "commercial")
//	// Returns: "company-org-commercial"
func GenerateSsoSessionName(orgAlias, partitionName string) string {
	logging.Debug.Log("GenerateSessionName called",
		"orgAlias", orgAlias,
		"partitionName", partitionName,
	)

	sessionName := fmt.Sprintf("%s-%s", orgAlias, partitionName)

	logging.Debug.Log("GenerateSessionName completed",
		"sessionName", sessionName,
	)

	return sessionName
}

// GenerateProfileName creates an AWS CLI profile name from components.
//
// Profile names follow a hierarchical naming convention that makes profiles
// easy to identify, filter, and organize. The region is only included in the
// name when it differs from the partition's default region.
//
// Naming Patterns:
//
//	Default region:     <partition>-<account-alias>-<role>
//	Non-default region: <partition>-<account-alias>-<role>--<region>
//
// Parameters:
//   - partition: Partition name (e.g., "commercial", "govcloud")
//   - accountAlias: Account alias (e.g., "prod", "dev")
//   - role: IAM role name (e.g., "Administrator", "Developer")
//   - region: AWS region (e.g., "us-east-1")
//   - defaultRegion: Partition's default region
//
// Returns:
//   - string: AWS CLI profile name
//
// Examples:
//
//	GenerateProfileName("commercial", "prod", "Administrator", "us-east-1", "us-east-1")
//	// Returns: "commercial-prod-Administrator"
//
//	GenerateProfileName("commercial", "prod", "Administrator", "us-west-2", "us-east-1")
//	// Returns: "commercial-prod-Administrator--us-west-2"
//
//	GenerateProfileName("govcloud", "prod", "SystemAdmin", "us-gov-east-1", "us-gov-east-1")
//	// Returns: "govcloud-prod-SystemAdmin"
func GenerateProfileName(partition, accountAlias, role, region, defaultRegion string) string {
	logging.Debug.Log("GenerateProfileName called",
		"partition", partition,
		"accountAlias", accountAlias,
		"role", role,
		"region", region,
		"defaultRegion", defaultRegion,
	)

	// Base profile name
	profileName := fmt.Sprintf("%s-%s-%s", partition, accountAlias, role)

	// Add region suffix if non-default (double dash separator)
	if region != "" && region != defaultRegion {
		profileName = fmt.Sprintf("%s--%s", profileName, region)
	}

	logging.Debug.Log("GenerateProfileName completed",
		"profileName", profileName,
	)

	return profileName
}
