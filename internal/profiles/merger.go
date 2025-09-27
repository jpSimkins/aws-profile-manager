package profiles

import (
	"strings"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
)

// merger handles duplicate detection when merging ProfileCollections.
//
// This is an internal component used by Importer for combining existing
// personal profiles with incoming profiles from a backup file.
//
// Duplicate Detection:
//   - Uses case-insensitive profile name matching
//   - Applies to all profile types (IAM, AssumeRole, Generic, SSO)
//   - Existing profiles take precedence (incoming duplicates are skipped)
type merger struct{}

// newMerger creates a new merger instance.
//
// The merger is stateless and can be reused for multiple merge operations.
//
// Returns:
//   - *merger: Configured merger ready for use
func newMerger() *merger {
	return &merger{}
}

// merge combines existing and incoming ProfileCollections with duplicate detection.
//
// This merges two ProfileCollections, preserving all existing profiles
// and adding only non-duplicate profiles from the incoming collection.
//
// Merge Strategy:
//   - Existing profiles are preserved
//   - Incoming profiles are added only if no duplicate exists
//   - Duplicates are counted for reporting
//
// Duplicate Detection Rules:
//   - IAM Users: Match by ProfileName (case-insensitive)
//   - AssumeRole Chains: Match by ProfileName (case-insensitive)
//   - Generic Profiles: Match by ProfileName (case-insensitive)
//   - SSO Organizations: Match by organization name (case-sensitive)
//
// Parameters:
//   - existing: Current profiles (nil-safe)
//   - incoming: Profiles to merge in (nil-safe)
//
// Returns:
//   - *schema.ProfileCollection: Merged collection (existing + new incoming)
//   - int: Number of duplicate profiles skipped
func (m *merger) merge(existing, incoming *schema.ProfileCollection) (*schema.ProfileCollection, SectionDuplicateStats) {
	logging.Debug.Log("merge called",
		"existing_nil", existing == nil,
		"incoming_nil", incoming == nil,
	)

	// Handle nil cases
	emptyStats := SectionDuplicateStats{}
	if existing == nil && incoming == nil {
		return &schema.ProfileCollection{}, emptyStats
	}
	if existing == nil {
		return incoming, emptyStats
	}
	if incoming == nil {
		return existing, emptyStats
	}

	dupStats := SectionDuplicateStats{}
	result := &schema.ProfileCollection{
		Organizations: make(map[string]*schema.Organization),
	}

	// Copy existing organizations (SSO profiles - usually not in unmanaged, but handle it)
	for orgName, org := range existing.Organizations {
		result.Organizations[orgName] = org
	}

	// Merge incoming organizations (SSO profiles)
	for orgName, incomingOrg := range incoming.Organizations {
		if _, exists := result.Organizations[orgName]; !exists {
			result.Organizations[orgName] = incomingOrg
		} else {
			// Count SSO profiles in the duplicate org
			dupStats.SsoProfiles++
			dupStats.TotalDuplicates++
			logging.Debug.Log("Skipping duplicate SSO organization", "org", orgName)
		}
	}

	// Merge each profile type with typed duplicate tracking
	var dup int
	result.IamUsers, dup = m.mergeIamUsers(existing.IamUsers, incoming.IamUsers)
	dupStats.IamProfiles += dup
	dupStats.TotalDuplicates += dup

	result.AssumeRoleChains, dup = m.mergeAssumeRoles(existing.AssumeRoleChains, incoming.AssumeRoleChains)
	dupStats.AssumeRoleProfiles += dup
	dupStats.TotalDuplicates += dup

	result.GenericProfiles, dup = m.mergeGeneric(existing.GenericProfiles, incoming.GenericProfiles)
	dupStats.GenericProfiles += dup
	dupStats.TotalDuplicates += dup

	logging.Debug.Log("merge completed",
		"iam_profiles", len(result.IamUsers),
		"assume_role_profiles", len(result.AssumeRoleChains),
		"generic_profiles", len(result.GenericProfiles),
		"duplicates_total", dupStats.TotalDuplicates,
		"duplicates_sso", dupStats.SsoProfiles,
		"duplicates_iam", dupStats.IamProfiles,
		"duplicates_assume_role", dupStats.AssumeRoleProfiles,
		"duplicates_generic", dupStats.GenericProfiles,
	)

	return result, dupStats
}

// mergeIamUsers merges IAM user profiles with duplicate detection.
func (m *merger) mergeIamUsers(existing, incoming []*schema.IamUser) ([]*schema.IamUser, int) {
	result := append([]*schema.IamUser{}, existing...) // Copy existing
	duplicates := 0

	for _, incomingUser := range incoming {
		if !m.iamUserExists(result, incomingUser) {
			result = append(result, incomingUser)
			logging.Debug.Log("Added IAM user", "profile", incomingUser.ProfileName)
		} else {
			duplicates++
			logging.Debug.Log("Skipping duplicate IAM user", "profile", incomingUser.ProfileName)
		}
	}

	return result, duplicates
}

// mergeAssumeRoles merges assume role chain profiles with duplicate detection.
func (m *merger) mergeAssumeRoles(existing, incoming []*schema.AssumeRoleChain) ([]*schema.AssumeRoleChain, int) {
	result := append([]*schema.AssumeRoleChain{}, existing...) // Copy existing
	duplicates := 0

	for _, incomingRole := range incoming {
		if !m.assumeRoleExists(result, incomingRole) {
			result = append(result, incomingRole)
			logging.Debug.Log("Added AssumeRole profile", "profile", incomingRole.ProfileName)
		} else {
			duplicates++
			logging.Debug.Log("Skipping duplicate AssumeRole profile", "profile", incomingRole.ProfileName)
		}
	}

	return result, duplicates
}

// mergeGeneric merges generic profiles with duplicate detection.
func (m *merger) mergeGeneric(existing, incoming []*schema.GenericProfile) ([]*schema.GenericProfile, int) {
	result := append([]*schema.GenericProfile{}, existing...) // Copy existing
	duplicates := 0

	for _, incomingGeneric := range incoming {
		if !m.genericExists(result, incomingGeneric) {
			result = append(result, incomingGeneric)
			logging.Debug.Log("Added generic profile", "profile", incomingGeneric.ProfileName)
		} else {
			duplicates++
			logging.Debug.Log("Skipping duplicate generic profile", "profile", incomingGeneric.ProfileName)
		}
	}

	return result, duplicates
}

// iamUserExists checks if an IAM user profile already exists in the collection.
func (m *merger) iamUserExists(users []*schema.IamUser, user *schema.IamUser) bool {
	for _, existing := range users {
		if strings.EqualFold(existing.ProfileName, user.ProfileName) {
			return true
		}
	}
	return false
}

// assumeRoleExists checks if an assume role profile already exists in the collection.
func (m *merger) assumeRoleExists(roles []*schema.AssumeRoleChain, role *schema.AssumeRoleChain) bool {
	for _, existing := range roles {
		if strings.EqualFold(existing.ProfileName, role.ProfileName) {
			return true
		}
	}
	return false
}

// genericExists checks if a generic profile already exists in the collection.
func (m *merger) genericExists(profiles []*schema.GenericProfile, profile *schema.GenericProfile) bool {
	for _, existing := range profiles {
		if strings.EqualFold(existing.ProfileName, profile.ProfileName) {
			return true
		}
	}
	return false
}
