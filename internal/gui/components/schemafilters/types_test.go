package schemafilters

import (
	"testing"

	"aws-profile-manager/internal/schema"
)

func TestIsEmptyCriteria_TrueWhenNoFilters(t *testing.T) {
	criteria := schema.FilterCriteria{}
	if !IsEmptyCriteria(criteria) {
		t.Fatal("expected empty criteria")
	}
}

func TestIsEmptyCriteria_FalseWhenAnyFilterSet(t *testing.T) {
	tests := []struct {
		name     string
		criteria schema.FilterCriteria
	}{
		{name: "organizations", criteria: schema.FilterCriteria{Organizations: []string{"org"}}},
		{name: "partitions", criteria: schema.FilterCriteria{Partitions: []string{"commercial"}}},
		{name: "regions", criteria: schema.FilterCriteria{Regions: []string{"us-east-1"}}},
		{name: "roles", criteria: schema.FilterCriteria{Roles: []string{"Admin"}}},
		{name: "accounts", criteria: schema.FilterCriteria{Accounts: []string{"dev"}}},
		{name: "all regions", criteria: schema.FilterCriteria{AllRegions: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsEmptyCriteria(tt.criteria) {
				t.Fatalf("expected non-empty criteria for %s", tt.name)
			}
		})
	}
}

func TestFilterTypeOrderAndLabels_AreComplete(t *testing.T) {
	if len(FilterTypeOrder) != 5 {
		t.Fatalf("expected 5 filter types in order, got %d", len(FilterTypeOrder))
	}

	for _, filterType := range FilterTypeOrder {
		label, exists := FilterTypeLabels[filterType]
		if !exists {
			t.Fatalf("missing label for filter type %v", filterType)
		}
		if label == "" {
			t.Fatalf("label should not be empty for filter type %v", filterType)
		}
	}
}
