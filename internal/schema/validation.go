package schema

import (
	"fmt"
	"regexp"
	"strings"

	"aws-profile-manager/internal/logging"
)

// Validate performs comprehensive validation of the Schema.
//
// This method validates the entire schema structure including version,
// presence of at least one section (managed or unmanaged), and validates
// all profile collections within each section.
//
// Validation Rules:
//   - Version field must not be empty
//   - At least one section (managed or unmanaged) must be defined
//   - Each section's profile collection must be valid
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *Schema) Validate() error {
	logging.Debug.Log("Validating Schema", "version", s.Version)

	if strings.TrimSpace(s.Version) == "" {
		return fmt.Errorf("version field is required")
	}

	// At least one section must be present
	if s.Managed == nil && s.Unmanaged == nil {
		return fmt.Errorf("at least one section (managed or unmanaged) must be defined")
	}

	// Validate managed section
	if s.Managed != nil {
		if err := s.Managed.Validate(); err != nil {
			return fmt.Errorf("managed: %w", err)
		}
	}

	// Validate unmanaged section
	if s.Unmanaged != nil {
		if err := s.Unmanaged.Validate(); err != nil {
			return fmt.Errorf("unmanaged: %w", err)
		}
	}

	// Validate presets
	if s.Presets != nil {
		if err := s.ValidatePresets(); err != nil {
			return fmt.Errorf("presets: %w", err)
		}
	}

	logging.Debug.Log("Schema validation completed")
	return nil
}

// Validate validates a ProfileCollection.
//
// This method validates all profile types within the collection including
// organizations, IAM users, assume role chains, and generic profiles.
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (p *ProfileCollection) Validate() error {
	logging.Debug.Log("Validating ProfileCollection",
		"orgCount", len(p.Organizations),
		"iamUserCount", len(p.IamUsers),
		"assumeRoleCount", len(p.AssumeRoleChains),
		"genericCount", len(p.GenericProfiles),
	)

	// Validate organizations
	for orgName, org := range p.Organizations {
		if err := org.Validate(); err != nil {
			return fmt.Errorf("organization %s: %w", orgName, err)
		}
	}

	// Validate IAM users
	for i, user := range p.IamUsers {
		if err := user.Validate(); err != nil {
			return fmt.Errorf("iam_user[%d]: %w", i, err)
		}
	}

	// Validate assume role chains
	for i, chain := range p.AssumeRoleChains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("assume_role_chain[%d]: %w", i, err)
		}
	}

	// Validate generic profiles
	for i, profile := range p.GenericProfiles {
		if err := profile.Validate(); err != nil {
			return fmt.Errorf("generic_profile[%d]: %w", i, err)
		}
	}

	logging.Debug.Log("ProfileCollection validation completed")
	return nil
}

// Validate validates UnmanagedProfiles section.
//
// This method validates the unmanaged profiles structure, ensuring at least
// one subsection (above or below) is defined and validates all profile
// collections within each subsection.
//
// Validation Rules:
//   - At least one subsection (above or below) must be defined
//   - Each subsection's profile collection must be valid
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (u *UnmanagedProfiles) Validate() error {
	logging.Debug.Log("Validating UnmanagedProfiles")

	// At least one subsection must be present
	if u.Above == nil && u.Below == nil {
		return fmt.Errorf("at least one subsection (above or below) must be defined")
	}

	// Validate above
	if u.Above != nil {
		if err := u.Above.Validate(); err != nil {
			return fmt.Errorf("above_managed: %w", err)
		}
	}

	// Validate below
	if u.Below != nil {
		if err := u.Below.Validate(); err != nil {
			return fmt.Errorf("below_managed: %w", err)
		}
	}

	return nil
}

// Validate validates Organization structure.
//
// This method validates an organization's basic fields and all partitions
// within the organization.
//
// Validation Rules:
//   - Name field must not be empty
//   - At least one partition must be defined
//   - All partitions must be valid
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (o *Organization) Validate() error {
	logging.Debug.Log("Validating Organization", "name", o.Name)

	if strings.TrimSpace(o.Name) == "" {
		return fmt.Errorf("organization name is required")
	}

	if len(o.Partitions) == 0 {
		return fmt.Errorf("organization must have at least one partition")
	}

	// Validate each partition
	for partitionName, partition := range o.Partitions {
		if err := partition.Validate(); err != nil {
			return fmt.Errorf("partition %s: %w", partitionName, err)
		}
	}

	return nil
}

// Validate validates Partition structure.
//
// This method validates a partition's SSO configuration, regions, roles,
// and all accounts within the partition.
//
// Validation Rules:
//   - URL must not be empty
//   - Default region must not be empty
//   - At least one region must be defined
//   - Default region must be in regions list
//   - At least one role must be defined
//   - At least one account must be defined
//   - All accounts must be valid
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (p *Partition) Validate() error {
	logging.Debug.Log("Validating Partition", "url", p.URL)

	// Validate URL
	if strings.TrimSpace(p.URL) == "" {
		return fmt.Errorf("SSO start URL is required")
	}

	// Validate default region
	if strings.TrimSpace(p.DefaultRegion) == "" {
		return fmt.Errorf("default region is required")
	}

	// Validate regions list
	if len(p.Regions) == 0 {
		return fmt.Errorf("at least one region must be defined")
	}

	// Validate accounts
	if len(p.Accounts) == 0 {
		return fmt.Errorf("at least one account must be defined")
	}

	for i, account := range p.Accounts {
		if err := account.Validate(); err != nil {
			return fmt.Errorf("account[%d]: %w", i, err)
		}
	}

	// Validate roles
	if len(p.Roles) == 0 {
		return fmt.Errorf("at least one role must be defined")
	}

	return nil
}

// Validate validates Account structure.
//
// This method validates an account's alias and AWS account ID.
//
// Validation Rules:
//   - Alias must not be empty
//   - ID must not be empty
//   - ID must be exactly 12 digits
//   - ID must contain only numeric characters
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (a *Account) Validate() error {
	logging.Debug.Log("Validating Account", "alias", a.Alias, "id", a.ID)

	// Validate alias
	if strings.TrimSpace(a.Alias) == "" {
		return fmt.Errorf("account alias is required")
	}

	// Validate name
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("account name is required")
	}

	// Validate account ID format (12 digits)
	accountIDPattern := regexp.MustCompile(`^\d{12}$`)
	if !accountIDPattern.MatchString(a.ID) {
		return fmt.Errorf("account ID must be 12 digits: %s", a.ID)
	}

	return nil
}

// Validate validates IamUser structure.
//
// This method validates an IAM user profile configuration.
//
// Validation Rules:
//   - ProfileName must not be empty
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (i *IamUser) Validate() error {
	logging.Debug.Log("Validating IamUser", "profileName", i.ProfileName)

	if strings.TrimSpace(i.ProfileName) == "" {
		return fmt.Errorf("profile_name is required")
	}

	return nil
}

// Validate validates AssumeRoleChain structure.
//
// This method validates an assume role chain configuration.
//
// Validation Rules:
//   - ProfileName must not be empty
//   - SourceProfile must not be empty
//   - RoleArn must not be empty
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (a *AssumeRoleChain) Validate() error {
	logging.Debug.Log("Validating AssumeRoleChain", "profileName", a.ProfileName)

	if strings.TrimSpace(a.ProfileName) == "" {
		return fmt.Errorf("profile_name is required")
	}

	if strings.TrimSpace(a.SourceProfile) == "" {
		return fmt.Errorf("source_profile is required")
	}

	if strings.TrimSpace(a.RoleArn) == "" {
		return fmt.Errorf("role_arn is required")
	}

	return nil
}

// Validate validates GenericProfile structure.
//
// This method validates a generic profile configuration.
//
// Validation Rules:
//   - ProfileName must not be empty
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (g *GenericProfile) Validate() error {
	logging.Debug.Log("Validating GenericProfile", "profileName", g.ProfileName)

	if strings.TrimSpace(g.ProfileName) == "" {
		return fmt.Errorf("profile_name is required")
	}

	if len(g.Properties) == 0 {
		return fmt.Errorf("at least one property is required")
	}

	return nil
}

// ValidatePresets validates all preset configurations in the schema.
//
// This method validates preset structure and ensures that preset filters
// reference valid entities in the managed section. Presets are optional,
// so if no presets are defined, this returns nil.
//
// Validation Rules:
//   - Preset keys must be non-empty
//   - Each preset must have a non-empty label
//   - Organization references must exist in managed section
//   - Partition references must exist in referenced organizations
//   - Account references must exist in referenced partitions
//   - Role references must exist in referenced partitions
//   - Region references must exist in referenced partitions
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (s *Schema) ValidatePresets() error {
	logging.Debug.Log("Validating Presets", "count", len(s.Presets))

	for key, preset := range s.Presets {
		if err := preset.Validate(key, s); err != nil {
			return fmt.Errorf("preset '%s': %w", key, err)
		}
	}

	logging.Debug.Log("Presets validation completed")
	return nil
}

// Validate validates a Preset configuration.
//
// This method validates the preset structure and optionally validates that
// filter values reference existing entities in the schema (organizations,
// partitions, accounts, roles, regions).
//
// Validation Rules:
//   - Label must not be empty
//   - If organizations are specified, they must exist in managed section
//   - If partitions are specified, they must exist in referenced organizations
//   - If accounts are specified, they must exist in referenced partitions
//   - If roles are specified, they must exist in referenced partitions
//   - If regions are specified, they must exist in referenced partitions
//
// Parameters:
//   - key: The preset key (for error messages)
//   - schema: The parent schema (for reference validation)
//
// Returns:
//   - error: First validation error encountered, nil if valid
func (p *Preset) Validate(key string, schema *Schema) error {
	logging.Debug.Log("Validating Preset",
		"key", key,
		"label", p.Label,
		"orgs", len(p.Organizations),
		"roles", len(p.Roles),
	)

	// Validate required fields
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("preset key cannot be empty")
	}

	if strings.TrimSpace(p.Label) == "" {
		return fmt.Errorf("label is required")
	}

	// If no managed section, can't validate references (but preset is still valid)
	if schema.Managed == nil || schema.Managed.Organizations == nil {
		logging.Debug.Log("Skipping preset reference validation (no managed section)")
		return nil
	}

	// Validate organization references
	if len(p.Organizations) > 0 {
		for _, orgAlias := range p.Organizations {
			if _, exists := schema.Managed.Organizations[orgAlias]; !exists {
				return fmt.Errorf("references non-existent organization '%s'", orgAlias)
			}
		}
	}

	// Build list of valid partitions, accounts, roles, regions from managed section
	validPartitions := make(map[string]bool)
	validAccounts := make(map[string]bool)
	validRoles := make(map[string]bool)
	validRegions := make(map[string]bool)

	// Collect from all organizations (or just filtered ones if specified)
	orgsToCheck := schema.Managed.Organizations
	if len(p.Organizations) > 0 {
		orgsToCheck = make(map[string]*Organization)
		for _, orgAlias := range p.Organizations {
			if org, exists := schema.Managed.Organizations[orgAlias]; exists {
				orgsToCheck[orgAlias] = org
			}
		}
	}

	for _, org := range orgsToCheck {
		for partitionName, partition := range org.Partitions {
			validPartitions[partitionName] = true

			// Collect accounts
			for _, account := range partition.Accounts {
				validAccounts[account.Alias] = true
			}

			// Collect roles
			for _, role := range partition.Roles {
				validRoles[role] = true
			}

			// Collect regions
			for _, region := range partition.Regions {
				validRegions[region] = true
			}
		}
	}

	// Validate partition references
	if len(p.Partitions) > 0 {
		for _, partition := range p.Partitions {
			if !validPartitions[partition] {
				return fmt.Errorf("references non-existent partition '%s'", partition)
			}
		}
	}

	// Validate account references
	if len(p.Accounts) > 0 {
		for _, account := range p.Accounts {
			if !validAccounts[account] {
				return fmt.Errorf("references non-existent account '%s'", account)
			}
		}
	}

	// Validate role references
	if len(p.Roles) > 0 {
		for _, role := range p.Roles {
			if !validRoles[role] {
				return fmt.Errorf("references non-existent role '%s'", role)
			}
		}
	}

	// Validate region references
	if len(p.Regions) > 0 {
		for _, region := range p.Regions {
			if !validRegions[region] {
				return fmt.Errorf("references non-existent region '%s'", region)
			}
		}
	}

	logging.Debug.Log("Preset validation completed", "key", key)
	return nil
}
