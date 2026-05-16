package awscli

import (
	"regexp"
	"sort"

	"aws-profile-manager/internal/logging"
)

// Filter handles filtering of AWS CLI profiles based on various criteria.
//
// The Filter provides flexible profile filtering with support for multiple
// criteria types. All criteria within a FilterCriteria are combined with AND logic,
// while multiple values within a single criterion use OR logic.
//
// Filtering Strategy:
//   - Empty criteria: Matches all profiles
//   - Multiple criteria: All must match (AND)
//   - Multiple values in criterion: Any can match (OR)
//   - Name pattern: Supports regex matching
//
// Example:
//
//	AccountIDs: ["123", "456"] AND RoleNames: ["Admin"]
//	→ Matches profiles with (account 123 OR 456) AND role Admin
type Filter struct{}

// NewFilter creates a new profile filter instance.
//
// Returns:
//   - *Filter: Filter instance ready to use
func NewFilter() *Filter {
	return &Filter{}
}

// FilterProfiles filters AWS CLI profiles based on specified criteria.
//
// This is the main filtering function that applies all filter criteria and
// returns matching profiles sorted by name.
//
// Filter Logic:
//   - Account IDs: Profile must match one of the specified account IDs (OR)
//   - Role Names: Profile must match one of the specified role names (OR)
//   - Regions: Profile must match one of the specified regions (OR)
//   - SSO Sessions: Profile must match one of the specified sessions (OR)
//   - Profile Types: Profile must match one of the specified types (OR)
//   - Name Pattern: Profile name must match the regex pattern
//   - All criteria are combined with AND logic
//
// Parameters:
//   - profiles: List of profiles to filter
//   - criteria: Filter criteria to apply
//
// Returns:
//   - []AwsCliProfile: Filtered profiles sorted alphabetically by name
func (f *Filter) FilterProfiles(profiles []AwsCliProfile, criteria FilterCriteria) []AwsCliProfile {
	logging.Debug.Log("Filtering AWS CLI profiles",
		"total_profiles", len(profiles),
		"account_ids", len(criteria.AccountIDs),
		"role_names", len(criteria.RoleNames),
		"regions", len(criteria.Regions),
		"sso_sessions", len(criteria.SsoSessions),
		"profile_types", len(criteria.ProfileTypes),
		"name_pattern", criteria.NamePattern)

	var filtered []AwsCliProfile
	for _, profile := range profiles {
		if f.matchesFilters(profile, criteria) {
			filtered = append(filtered, profile)
		}
	}

	// Sort by profile name for consistent ordering
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	logging.Debug.Log("Profile filtering completed",
		"filtered_count", len(filtered),
		"original_count", len(profiles))

	return filtered
}

// matchesFilters checks if a single profile matches all filter criteria.
//
// This method implements the AND logic for combining criteria - a profile must
// match ALL specified criteria to pass the filter.
//
// Parameters:
//   - profile: Profile to check
//   - criteria: Filter criteria to apply
//
// Returns:
//   - bool: true if profile matches all criteria, false otherwise
func (f *Filter) matchesFilters(profile AwsCliProfile, criteria FilterCriteria) bool {
	// Filter by profile types
	if len(criteria.ProfileTypes) > 0 {
		if !f.containsProfileType(criteria.ProfileTypes, profile.Type) {
			return false
		}
	}

	// Filter by account IDs
	if len(criteria.AccountIDs) > 0 {
		if !f.containsString(criteria.AccountIDs, profile.AccountID) {
			return false
		}
	}

	// Filter by role names
	if len(criteria.RoleNames) > 0 {
		if !f.containsString(criteria.RoleNames, profile.RoleName) {
			return false
		}
	}

	// Filter by regions
	if len(criteria.Regions) > 0 {
		if !f.containsString(criteria.Regions, profile.Region) {
			return false
		}
	}

	// Filter by SSO sessions
	if len(criteria.SsoSessions) > 0 {
		if !f.containsString(criteria.SsoSessions, profile.SsoSession) {
			return false
		}
	}

	// Filter by profile name pattern
	if criteria.NamePattern != "" {
		matched, err := regexp.MatchString(criteria.NamePattern, profile.Name)
		if err != nil {
			logging.Log.Warn("Invalid regex pattern for profile name filter",
				"pattern", criteria.NamePattern,
				"error", err)
			return false
		}
		if !matched {
			return false
		}
	}

	return true
}

// containsString checks if a string slice contains a specific string (case-sensitive).
//
// Parameters:
//   - slice: String slice to search
//   - target: String to find
//
// Returns:
//   - bool: true if target is found, false otherwise
func (f *Filter) containsString(slice []string, target string) bool {
	if target == "" {
		return false
	}

	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// containsProfileType checks if a ProfileType slice contains a specific ProfileType.
//
// Parameters:
//   - slice: ProfileType slice to search
//   - target: ProfileType to find
//
// Returns:
//   - bool: true if target is found, false otherwise
func (f *Filter) containsProfileType(slice []ProfileType, target ProfileType) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// GetAvailableFilterOptions extracts all unique filter values from profiles and sessions.
//
// This method scans through all profiles and sessions to build lists of unique values
// for each filterable field. Useful for populating filter UI dropdowns.
//
// Parameters:
//   - profiles: List of profiles to extract values from
//   - sessions: List of SSO sessions to extract values from
//
// Returns:
//   - FilterOptions: Unique values for each filter field, sorted alphabetically
func (f *Filter) GetAvailableFilterOptions(profiles []AwsCliProfile, sessions []SsoSession) FilterOptions {
	logging.Debug.Log("Generating available filter options",
		"profiles", len(profiles),
		"sessions", len(sessions))

	accountIDs := make(map[string]bool)
	roleNames := make(map[string]bool)
	regions := make(map[string]bool)
	ssoSessions := make(map[string]bool)

	// Extract unique values from profiles
	for _, profile := range profiles {
		if profile.AccountID != "" {
			accountIDs[profile.AccountID] = true
		}
		if profile.RoleName != "" {
			roleNames[profile.RoleName] = true
		}
		if profile.Region != "" {
			regions[profile.Region] = true
		}
		if profile.SsoSession != "" {
			ssoSessions[profile.SsoSession] = true
		}
	}

	// Add sessions that might not be referenced by profiles
	for _, session := range sessions {
		if session.Name != "" {
			ssoSessions[session.Name] = true
		}
	}

	options := FilterOptions{
		AccountIDs:  f.mapKeysToSortedSlice(accountIDs),
		RoleNames:   f.mapKeysToSortedSlice(roleNames),
		Regions:     f.mapKeysToSortedSlice(regions),
		SsoSessions: f.mapKeysToSortedSlice(ssoSessions),
	}

	logging.Debug.Log("Filter options generated",
		"account_ids", len(options.AccountIDs),
		"role_names", len(options.RoleNames),
		"regions", len(options.Regions),
		"sso_sessions", len(options.SsoSessions))

	return options
}

// FilterByAccountIDs creates a FilterCriteria with only account ID filters
func (f *Filter) FilterByAccountIDs(accountIDs ...string) FilterCriteria {
	return FilterCriteria{
		AccountIDs: accountIDs,
	}
}

// FilterByRoleNames creates a FilterCriteria with only role name filters
func (f *Filter) FilterByRoleNames(roleNames ...string) FilterCriteria {
	return FilterCriteria{
		RoleNames: roleNames,
	}
}

// FilterByRegions creates a FilterCriteria with only region filters
func (f *Filter) FilterByRegions(regions ...string) FilterCriteria {
	return FilterCriteria{
		Regions: regions,
	}
}

// FilterBySsoSessions creates a FilterCriteria with only SSO session filters
func (f *Filter) FilterBySsoSessions(ssoSessions ...string) FilterCriteria {
	return FilterCriteria{
		SsoSessions: ssoSessions,
	}
}

// FilterByNamePattern creates a FilterCriteria with only name pattern filter
func (f *Filter) FilterByNamePattern(pattern string) FilterCriteria {
	return FilterCriteria{
		NamePattern: pattern,
	}
}

// FilterByProfileTypes creates a FilterCriteria with only profile type filters
func (f *Filter) FilterByProfileTypes(profileTypes ...ProfileType) FilterCriteria {
	return FilterCriteria{
		ProfileTypes: profileTypes,
	}
}

// CombineFilters combines multiple FilterCriteria into one
func (f *Filter) CombineFilters(filters ...FilterCriteria) FilterCriteria {
	var combined FilterCriteria

	for _, filter := range filters {
		combined.AccountIDs = append(combined.AccountIDs, filter.AccountIDs...)
		combined.RoleNames = append(combined.RoleNames, filter.RoleNames...)
		combined.Regions = append(combined.Regions, filter.Regions...)
		combined.SsoSessions = append(combined.SsoSessions, filter.SsoSessions...)
		combined.ProfileTypes = append(combined.ProfileTypes, filter.ProfileTypes...)

		// For name pattern, use the last non-empty one
		if filter.NamePattern != "" {
			combined.NamePattern = filter.NamePattern
		}
	}

	// Remove duplicates and sort
	combined.AccountIDs = f.removeDuplicatesAndSort(combined.AccountIDs)
	combined.RoleNames = f.removeDuplicatesAndSort(combined.RoleNames)
	combined.Regions = f.removeDuplicatesAndSort(combined.Regions)
	combined.SsoSessions = f.removeDuplicatesAndSort(combined.SsoSessions)
	combined.ProfileTypes = f.removeDuplicateProfileTypes(combined.ProfileTypes)

	return combined
}

// ValidateFilterCriteria validates that the filter criteria is properly formed
func (f *Filter) ValidateFilterCriteria(criteria FilterCriteria) error {
	// Validate regex pattern if provided
	if criteria.NamePattern != "" {
		if _, err := regexp.Compile(criteria.NamePattern); err != nil {
			return logging.Log.ErrorWithDetails("Invalid regex pattern in filter criteria", err)
		}
	}

	logging.Debug.Log("Filter criteria validation passed")
	return nil
}

// mapKeysToSortedSlice converts map keys to a sorted slice
func (f *Filter) mapKeysToSortedSlice(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// removeDuplicatesAndSort removes duplicates from a string slice and sorts it
func (f *Filter) removeDuplicatesAndSort(slice []string) []string {
	uniqueMap := make(map[string]bool)
	for _, item := range slice {
		if item != "" {
			uniqueMap[item] = true
		}
	}
	return f.mapKeysToSortedSlice(uniqueMap)
}

// removeDuplicateProfileTypes removes duplicates from a ProfileType slice
func (f *Filter) removeDuplicateProfileTypes(slice []ProfileType) []ProfileType {
	uniqueMap := make(map[ProfileType]bool)
	for _, item := range slice {
		uniqueMap[item] = true
	}

	// Convert map keys to slice
	result := make([]ProfileType, 0, len(uniqueMap))
	for k := range uniqueMap {
		result = append(result, k)
	}
	return result
}
