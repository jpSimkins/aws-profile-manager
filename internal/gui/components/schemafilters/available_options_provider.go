package schemafilters

import (
	"sort"

	"aws-profile-manager/internal/schema"
)

// AvailableOptionsProvider computes option sets for each filter control.
//
// It evaluates current criteria to determine which options remain available
// for each dimension, while leaving filtering semantics to schema.FilterSchema.
type AvailableOptionsProvider struct {
	sourceSchema *schema.Schema
}

// NewAvailableOptionsProvider creates an options provider for a schema.
func NewAvailableOptionsProvider(original *schema.Schema) *AvailableOptionsProvider {
	return &AvailableOptionsProvider{sourceSchema: original}
}

// SetSchema updates the source schema used for option calculations.
func (p *AvailableOptionsProvider) SetSchema(original *schema.Schema) {
	p.sourceSchema = original
}

// GetAvailableOptions returns available options by filter type.
func (p *AvailableOptionsProvider) GetAvailableOptions(criteria schema.FilterCriteria) map[FilterType][]string {
	result := map[FilterType][]string{
		FilterTypeOrganizations: {},
		FilterTypePartitions:    {},
		FilterTypeRegions:       {},
		FilterTypeRoles:         {},
		FilterTypeAccounts:      {},
	}

	if p.sourceSchema == nil || p.sourceSchema.Managed == nil {
		return result
	}

	for _, filterType := range FilterTypeOrder {
		criteriaIgnoringDimension := cloneCriteriaIgnoringDimension(criteria, filterType)
		filteredSchema, err := schema.FilterSchema(p.sourceSchema, criteriaIgnoringDimension)
		if err != nil || filteredSchema == nil || filteredSchema.Managed == nil {
			continue
		}

		result[filterType] = collectOptionsForFilterType(filteredSchema.Managed, filterType)
	}

	return result
}

func cloneCriteriaIgnoringDimension(criteria schema.FilterCriteria, filterType FilterType) schema.FilterCriteria {
	clone := schema.FilterCriteria{
		Organizations: append([]string{}, criteria.Organizations...),
		Partitions:    append([]string{}, criteria.Partitions...),
		Regions:       append([]string{}, criteria.Regions...),
		Roles:         append([]string{}, criteria.Roles...),
		Accounts:      append([]string{}, criteria.Accounts...),
		AllRegions:    criteria.AllRegions,
	}

	switch filterType {
	case FilterTypeOrganizations:
		clone.Organizations = nil
	case FilterTypePartitions:
		clone.Partitions = nil
	case FilterTypeRegions:
		clone.Regions = nil
	case FilterTypeRoles:
		clone.Roles = nil
	case FilterTypeAccounts:
		clone.Accounts = nil
	}

	return clone
}

func collectOptionsForFilterType(collection *schema.ProfileCollection, filterType FilterType) []string {
	if collection == nil {
		return nil
	}

	set := map[string]struct{}{}
	for orgName, organization := range collection.Organizations {
		if filterType == FilterTypeOrganizations {
			set[orgName] = struct{}{}
		}

		for partitionName, partition := range organization.Partitions {
			if filterType == FilterTypePartitions {
				set[partitionName] = struct{}{}
			}

			if filterType == FilterTypeRegions {
				for _, region := range partition.Regions {
					set[region] = struct{}{}
				}
			}

			if filterType == FilterTypeRoles {
				for _, role := range partition.Roles {
					set[role] = struct{}{}
				}
			}

			if filterType == FilterTypeAccounts {
				for _, account := range partition.Accounts {
					set[account.Alias] = struct{}{}
				}
			}
		}
	}

	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
