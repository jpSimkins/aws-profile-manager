// Package generators provides AWS CLI configuration file content generation.
//
// This package contains generators for different types of AWS profiles and sessions,
// converting schema models into properly formatted AWS CLI config file content.
//
// Generators:
//   - SSO: Profiles using AWS SSO authentication
//   - IAM: Profiles using IAM user credentials (static or credential_process)
//   - AssumeRole: Profiles that assume roles using source profiles
//   - Generic: Catch-all for profiles with unrecognized authentication types
//
// All generators follow a common pattern:
//  1. Accept a schema.ProfileCollection as input
//  2. Generate formatted AWS CLI config content
//  3. Return content string and statistics
//  4. Pure functions with no file I/O (caller handles writing)
//
// Example Usage:
//
//	collection := schema.ProfileCollection{...}
//	content, stats := generators.GenerateSsoProfiles(collection)
//	// Write content to ~/.aws/config managed section
package generators

// SectionStats tracks statistics for generated profiles and sessions.
//
// This struct provides comprehensive counts of what was generated, useful for
// reporting to users, logging, and validation. Provides detailed breakdown by
// profile type and organizational structure.
type SectionStats struct {
	// Totals
	ProfilesWritten int // Total number of profiles generated
	SessionsWritten int // Number of SSO sessions generated (SSO only)

	// Breakdown by profile type
	SsoProfiles        int // Number of SSO profiles generated
	IamProfiles        int // Number of IAM profiles generated
	AssumeRoleProfiles int // Number of AssumeRole profiles generated
	GenericProfiles    int // Number of Generic profiles generated

	// SSO-specific organizational details
	OrganizationCount int // Number of organizations processed (SSO only)
	PartitionCount    int // Number of partitions across all orgs (SSO only)
	AccountCount      int // Number of accounts across all partitions (SSO only)
	RoleCount         int // Number of unique roles (SSO only)
	RegionCount       int // Number of regions including defaults (SSO only)
}
