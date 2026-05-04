package schema

import (
	"strings"

	"aws-profile-manager/internal/logging"
)

// FilterCriteria defines criteria for filtering profiles.
//
// Empty fields mean no filtering for that dimension. Multiple criteria
// are combined with AND logic (profiles must match all specified criteria).
type FilterCriteria struct {
	Organizations []string // Filter by organization names (exact match)
	Partitions    []string // Filter by partition names (e.g., "commercial", "govcloud")
	Accounts      []string // Filter by account aliases (exact match)
	Roles         []string // Filter by role names (exact match)
	Regions       []string // Filter by regions (exact match)
	AllRegions    bool     // Include all regions (overrides Regions filter)
}

// FilterSchema applies filtering criteria to a Schema and returns a new filtered Schema.
//
// This is the main entry point for filtering profiles. It creates a new Schema
// with only the profiles that match the specified criteria. The unmanaged section
// is never filtered (always copied as-is).
//
// Filtering Logic:
//   - Empty criteria: Returns original schema pointer unchanged (no filtering)
//   - Organizations: Keep only matching organizations
//   - Partitions: Keep only matching partitions within organizations
//   - Accounts: Keep only matching accounts within partitions
//   - Roles: Keep only matching roles within partitions
//   - Regions: Keep only matching regions (or all if AllRegions=true)
//
// Multiple filters are combined with AND logic - profiles must match all
// specified criteria to be included.
//
// Parameters:
//   - s: Schema to filter
//   - criteria: Filter criteria
//
// Returns:
//   - *Schema: New Schema with filtered managed section
//   - error: Any error encountered during filtering
//
// Example:
//
//	filtered, err := schema.FilterSchema(original, schema.FilterCriteria{
//	    Accounts: []string{"prod", "staging"},
//	    Roles:    []string{"Administrator"},
//	})
func FilterSchema(s *Schema, criteria FilterCriteria) (*Schema, error) {
	logging.Debug.Log("FilterSchema called",
		"organizations", len(criteria.Organizations),
		"partitions", len(criteria.Partitions),
		"accounts", len(criteria.Accounts),
		"roles", len(criteria.Roles),
		"regions", len(criteria.Regions),
		"allRegions", criteria.AllRegions,
	)

	// If no filters, return copy of original
	if !hasFilters(criteria) {
		logging.Debug.Log("No filters specified, returning original schema")
		return s, nil
	}

	// Create filtered schema
	filtered := &Schema{
		Version: s.Version,
		Managed: &ProfileCollection{
			Organizations: make(map[string]*Organization),
		},
	}

	// Copy unmanaged sections as-is (no filtering)
	filtered.Unmanaged = s.Unmanaged

	// Filter managed section
	if s.Managed != nil {
		filterManagedSection(s.Managed, filtered.Managed, criteria)
	}

	logging.Debug.Log("FilterSchema completed",
		"originalOrgs", len(s.Managed.Organizations),
		"filteredOrgs", len(filtered.Managed.Organizations),
	)

	return filtered, nil
}

// hasFilters checks if any filter criteria are specified.
//
// This helper function determines if filtering should be applied by checking
// if any filter fields are set in the criteria.
//
// Parameters:
//   - criteria: Filter criteria to check
//
// Returns:
//   - bool: true if any filter is specified
func hasFilters(criteria FilterCriteria) bool {
	return len(criteria.Organizations) > 0 ||
		len(criteria.Partitions) > 0 ||
		len(criteria.Accounts) > 0 ||
		len(criteria.Roles) > 0 ||
		len(criteria.Regions) > 0 ||
		criteria.AllRegions
}

// filterManagedSection filters the managed ProfileCollection.
//
// This internal function applies filter criteria to SSO organizations,
// creating a new filtered organization structure.
//
// Parameters:
//   - source: Source ProfileCollection to filter
//   - dest: Destination ProfileCollection for filtered results
//   - criteria: Filter criteria to apply
func filterManagedSection(source, dest *ProfileCollection, criteria FilterCriteria) {
	// Filter SSO organizations
	for orgName, org := range source.Organizations {
		// Check organization filter
		if len(criteria.Organizations) > 0 && !containsString(criteria.Organizations, orgName) {
			logging.Debug.Log("Skipping organization", "org", orgName)
			continue
		}

		filteredOrg := &Organization{
			Name:       org.Name,
			Partitions: make(map[string]Partition),
		}

		// Filter partitions within organization
		for partName, part := range org.Partitions {
			// Check partition filter
			if len(criteria.Partitions) > 0 && !containsString(criteria.Partitions, partName) {
				logging.Debug.Log("Skipping partition", "org", orgName, "partition", partName)
				continue
			}

			filteredPart := Partition{
				URL:           part.URL,
				DefaultRegion: part.DefaultRegion,
				Accounts:      []Account{},
				Roles:         []string{},
				Regions:       []string{},
			}

			// Filter accounts
			for _, account := range part.Accounts {
				if len(criteria.Accounts) > 0 && !containsString(criteria.Accounts, account.Alias) {
					logging.Debug.Log("Skipping account", "account", account.Alias)
					continue
				}
				filteredPart.Accounts = append(filteredPart.Accounts, account)
			}

			// Filter roles
			for _, role := range part.Roles {
				if len(criteria.Roles) > 0 && !containsString(criteria.Roles, role) {
					logging.Debug.Log("Skipping role", "role", role)
					continue
				}
				filteredPart.Roles = append(filteredPart.Roles, role)
			}

			// Filter regions
			if criteria.AllRegions {
				// Include all regions
				filteredPart.Regions = part.Regions
			} else if len(criteria.Regions) > 0 {
				// Include only specified regions
				for _, region := range part.Regions {
					if containsString(criteria.Regions, region) {
						filteredPart.Regions = append(filteredPart.Regions, region)
					}
				}
			} else {
				// No region filter - include all
				filteredPart.Regions = part.Regions
			}

			// Only add partition if it has accounts and roles
			if len(filteredPart.Accounts) > 0 && len(filteredPart.Roles) > 0 {
				filteredOrg.Partitions[partName] = filteredPart
			}
		}

		// Only add organization if it has partitions
		if len(filteredOrg.Partitions) > 0 {
			dest.Organizations[orgName] = filteredOrg
		}
	}

	// Copy IAM users, AssumeRole chains, and Generic profiles as-is
	// These don't have organization/partition/account/role hierarchy
	dest.IamUsers = source.IamUsers
	dest.AssumeRoleChains = source.AssumeRoleChains
	dest.GenericProfiles = source.GenericProfiles
}

// containsString checks if a slice contains a string (case-insensitive).
//
// This helper function performs case-insensitive string matching for filter criteria.
//
// Parameters:
//   - slice: Slice of strings to search
//   - item: Item to search for
//
// Returns:
//   - bool: true if item is found (case-insensitive match)
func containsString(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}
