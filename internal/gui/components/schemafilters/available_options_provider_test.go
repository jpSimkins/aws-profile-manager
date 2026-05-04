package schemafilters

import (
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
)

func TestAvailableOptionsProvider_NilSchema_ReturnsEmptySlices(t *testing.T) {
	provider := NewAvailableOptionsProvider(nil)
	options := provider.GetAvailableOptions(schema.FilterCriteria{})

	for _, filterType := range FilterTypeOrder {
		values, exists := options[filterType]
		if !exists {
			t.Fatalf("expected options key for filter type %v", filterType)
		}
		if len(values) != 0 {
			t.Fatalf("expected empty options for %v, got %v", filterType, values)
		}
	}
}

func TestAvailableOptionsProvider_SetSchema_UpdatesResults(t *testing.T) {
	provider := NewAvailableOptionsProvider(nil)
	provider.SetSchema(schematest.NewManagedSsoSingle())

	options := provider.GetAvailableOptions(schema.FilterCriteria{})

	if len(options[FilterTypeOrganizations]) == 0 {
		t.Fatal("expected organizations to be available after setting schema")
	}
	if len(options[FilterTypeAccounts]) == 0 {
		t.Fatal("expected accounts to be available after setting schema")
	}
}

func TestAvailableOptionsProvider_CriteriaBypassPerDimension(t *testing.T) {
	s := schematest.NewManagedSsoMultiOrg()
	provider := NewAvailableOptionsProvider(s)

	criteria := schema.FilterCriteria{
		Organizations: []string{"prod-org"},
	}
	options := provider.GetAvailableOptions(criteria)

	// Organization list should still include all organizations because provider
	// computes each dimension while ignoring that dimension's active criteria.
	if len(options[FilterTypeOrganizations]) < 2 {
		t.Fatalf("expected multiple organizations while org criteria is active, got %v", options[FilterTypeOrganizations])
	}

	// Account options should be constrained by organization filter.
	for _, account := range options[FilterTypeAccounts] {
		if account == "test-dev" {
			t.Fatalf("expected account test-dev to be excluded by org criteria, got %v", options[FilterTypeAccounts])
		}
	}
}
