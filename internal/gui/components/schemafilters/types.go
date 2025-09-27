package schemafilters

import "aws-profile-manager/internal/schema"

// FilterType identifies a schema filter dimension in canonical UI order.
type FilterType int

const (
	// FilterTypeOrganizations filters by organization alias.
	FilterTypeOrganizations FilterType = iota
	// FilterTypePartitions filters by partition name.
	FilterTypePartitions
	// FilterTypeRegions filters by AWS region.
	FilterTypeRegions
	// FilterTypeRoles filters by IAM role name.
	FilterTypeRoles
	// FilterTypeAccounts filters by account alias.
	FilterTypeAccounts
)

// FilterTypeOrder defines the canonical order for rendering filter controls.
var FilterTypeOrder = []FilterType{
	FilterTypeOrganizations,
	FilterTypePartitions,
	FilterTypeRegions,
	FilterTypeRoles,
	FilterTypeAccounts,
}

// FilterTypeLabels maps each filter type to its user-facing label.
var FilterTypeLabels = map[FilterType]string{
	FilterTypeOrganizations: "Organizations",
	FilterTypePartitions:    "Partitions",
	FilterTypeRegions:       "Regions",
	FilterTypeRoles:         "Roles",
	FilterTypeAccounts:      "Accounts",
}

// FilterItemConfig customizes a filter control instance.
type FilterItemConfig struct {
	AllowSearch       bool
	AllowCheckAll     bool
	MaxHeight         float32
	SearchPlaceholder string
}

// FilterItemProps defines the input contract for a FilterItem.
type FilterItemProps struct {
	Label             string
	AvailableOptions  []string
	SelectedOptions   []string
	OnChange          func([]string)
	AllowSearch       bool
	AllowCheckAll     bool
	MaxHeight         float32
	SearchPlaceholder string
}

// IsEmptyCriteria returns true when no filter fields are active.
func IsEmptyCriteria(criteria schema.FilterCriteria) bool {
	return len(criteria.Organizations) == 0 &&
		len(criteria.Partitions) == 0 &&
		len(criteria.Regions) == 0 &&
		len(criteria.Roles) == 0 &&
		len(criteria.Accounts) == 0 &&
		!criteria.AllRegions
}
