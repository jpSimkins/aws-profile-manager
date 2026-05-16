package schemalist

import (
	"sort"
	"strings"

	"aws-profile-manager/internal/schema"
)

// AccountRecord is the flattened, display-ready representation of one AWS account.
//
// It is derived from the schema by FlattenManagedAccounts and contains all
// information needed to render both the list row and the detail panel.
type AccountRecord struct {
	OrganizationAlias string
	OrganizationName  string
	PartitionName     string
	AccountAlias      string
	AccountName       string
	AccountID         string
	DefaultRegion     string
	Regions           []string
	Roles             []string
	SsoURL            string
}

// FlattenManagedAccounts extracts AccountRecords from the managed section of a
// schema in a stable, deterministic sort order.
//
// Records are sorted by: organization alias → partition → account name → account ID.
// Returns nil when the schema or its managed section is nil.
func FlattenManagedAccounts(sourceSchema *schema.Schema) []AccountRecord {
	if sourceSchema == nil || sourceSchema.Managed == nil {
		return nil
	}

	// Collect and sort organization aliases for deterministic iteration.
	orgAliases := make([]string, 0, len(sourceSchema.Managed.Organizations))
	for alias := range sourceSchema.Managed.Organizations {
		orgAliases = append(orgAliases, alias)
	}
	sort.Strings(orgAliases)

	results := make([]AccountRecord, 0)
	for _, orgAlias := range orgAliases {
		org := sourceSchema.Managed.Organizations[orgAlias]
		if org == nil {
			continue
		}

		// Collect and sort partition names for deterministic iteration.
		partitionNames := make([]string, 0, len(org.Partitions))
		for name := range org.Partitions {
			partitionNames = append(partitionNames, name)
		}
		sort.Strings(partitionNames)

		for _, partitionName := range partitionNames {
			partition := org.Partitions[partitionName]
			for _, account := range partition.Accounts {
				results = append(results, AccountRecord{
					OrganizationAlias: orgAlias,
					OrganizationName:  org.Name,
					PartitionName:     partitionName,
					AccountAlias:      account.Alias,
					AccountName:       account.Name,
					AccountID:         account.ID,
					DefaultRegion:     partition.DefaultRegion,
					Regions:           append([]string{}, partition.Regions...),
					Roles:             append([]string{}, partition.Roles...),
					SsoURL:            partition.URL,
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		l, r := results[i], results[j]
		if l.OrganizationAlias != r.OrganizationAlias {
			return l.OrganizationAlias < r.OrganizationAlias
		}
		if l.PartitionName != r.PartitionName {
			return l.PartitionName < r.PartitionName
		}
		if l.AccountName != r.AccountName {
			return l.AccountName < r.AccountName
		}
		return l.AccountID < r.AccountID
	})

	return results
}

// resolveAccountTitle returns a display string for a list row.
//
// Format: "<alias> <partition> (<id>)"
// Falls back to "Unnamed Account" when alias is blank.
func resolveAccountTitle(record AccountRecord) string {
	name := strings.TrimSpace(record.AccountAlias)
	if name == "" {
		name = "Unnamed Account"
	}
	return name + " " + strings.TrimSpace(record.PartitionName) + " (" + record.AccountID + ")"
}

// safeValue returns the trimmed value, or "n/a" if it is empty.
func safeValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "n/a"
	}
	return trimmed
}

// joinOrFallback joins a slice with ", " or returns "n/a" when empty.
func joinOrFallback(values []string) string {
	if len(values) == 0 {
		return "n/a"
	}
	return strings.Join(values, ", ")
}
