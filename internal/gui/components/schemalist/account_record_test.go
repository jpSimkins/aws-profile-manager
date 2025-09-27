package schemalist

import (
	"testing"

	schematest "aws-profile-manager/internal/schema/test"
)

// --- FlattenManagedAccounts ---

func TestFlattenManagedAccounts_NilSchema(t *testing.T) {
	records := FlattenManagedAccounts(nil)
	if records != nil {
		t.Fatal("expected nil records for nil schema")
	}
}

func TestFlattenManagedAccounts_NilManagedSection(t *testing.T) {
	schema := schematest.NewEmpty()
	records := FlattenManagedAccounts(schema)
	if records != nil {
		t.Fatal("expected nil records for schema with nil managed section")
	}
}

func TestFlattenManagedAccounts_SortedStableOrder(t *testing.T) {
	records := FlattenManagedAccounts(schematest.NewManagedSsoMultiOrg())
	if len(records) == 0 {
		t.Fatal("expected at least one account record")
	}
	for i := 1; i < len(records); i++ {
		prev := records[i-1]
		curr := records[i]
		if prev.OrganizationAlias > curr.OrganizationAlias {
			t.Fatalf("records not sorted by organization alias at index %d: %q > %q",
				i, prev.OrganizationAlias, curr.OrganizationAlias)
		}
	}
}

func TestFlattenManagedAccounts_AllFieldsMapped(t *testing.T) {
	records := FlattenManagedAccounts(schematest.NewManagedSsoSingle())

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	r := records[0]

	if r.AccountID == "" {
		t.Error("AccountID should not be empty")
	}
	if r.AccountAlias == "" {
		t.Error("AccountAlias should not be empty")
	}
	if r.OrganizationAlias == "" {
		t.Error("OrganizationAlias should not be empty")
	}
	if r.PartitionName == "" {
		t.Error("PartitionName should not be empty")
	}
	if r.DefaultRegion == "" {
		t.Error("DefaultRegion should not be empty")
	}
	if len(r.Roles) == 0 {
		t.Error("Roles should not be empty for SSO single schema")
	}
}

func TestFlattenManagedAccounts_SlicesAreCopies(t *testing.T) {
	records := FlattenManagedAccounts(schematest.NewManagedSsoSingle())
	if len(records) == 0 {
		t.Fatal("expected at least one record")
	}

	r := records[0]

	// Mutating the returned slices should not affect subsequent calls.
	if len(r.Regions) > 0 {
		r.Regions[0] = "mutated"
	}
	if len(r.Roles) > 0 {
		r.Roles[0] = "mutated"
	}

	records2 := FlattenManagedAccounts(schematest.NewManagedSsoSingle())
	if len(records2) == 0 {
		t.Fatal("expected at least one record on second call")
	}
	if len(records2[0].Regions) > 0 && records2[0].Regions[0] == "mutated" {
		t.Error("Regions slice should be a copy, not a shared reference")
	}
	if len(records2[0].Roles) > 0 && records2[0].Roles[0] == "mutated" {
		t.Error("Roles slice should be a copy, not a shared reference")
	}
}

func TestFlattenManagedAccounts_MultiAccount(t *testing.T) {
	records := FlattenManagedAccounts(schematest.NewManagedSsoMultiAccount())
	if len(records) != 3 {
		t.Fatalf("expected 3 records for multi-account schema, got %d", len(records))
	}
}

// --- resolveAccountTitle ---

func TestResolveAccountTitle_UsesAlias(t *testing.T) {
	r := AccountRecord{AccountAlias: "acme-prod", PartitionName: "commercial", AccountID: "123"}
	title := resolveAccountTitle(r)
	if title != "acme-prod commercial (123)" {
		t.Errorf("unexpected title: %q", title)
	}
}

func TestResolveAccountTitle_FallsBackToUnnamed(t *testing.T) {
	r := AccountRecord{AccountAlias: "", PartitionName: "commercial", AccountID: "123"}
	title := resolveAccountTitle(r)
	if title != "Unnamed Account commercial (123)" {
		t.Errorf("unexpected title: %q", title)
	}
}

func TestResolveAccountTitle_WhitespaceAliasIsUnnamed(t *testing.T) {
	r := AccountRecord{AccountAlias: "   ", PartitionName: "commercial", AccountID: "123"}
	title := resolveAccountTitle(r)
	if title != "Unnamed Account commercial (123)" {
		t.Errorf("unexpected title: %q", title)
	}
}

// --- safeValue ---

func TestSafeValue_EmptyString(t *testing.T) {
	if safeValue("") != "n/a" {
		t.Error("expected n/a for empty string")
	}
}

func TestSafeValue_WhitespaceOnly(t *testing.T) {
	if safeValue("   ") != "n/a" {
		t.Error("expected n/a for whitespace-only string")
	}
}

func TestSafeValue_NonEmpty(t *testing.T) {
	if safeValue("  hello  ") != "hello" {
		t.Error("expected trimmed value")
	}
}

// --- joinOrFallback ---

func TestJoinOrFallback_Empty(t *testing.T) {
	if joinOrFallback(nil) != "n/a" {
		t.Error("expected n/a for nil slice")
	}
	if joinOrFallback([]string{}) != "n/a" {
		t.Error("expected n/a for empty slice")
	}
}

func TestJoinOrFallback_SingleItem(t *testing.T) {
	if joinOrFallback([]string{"us-east-1"}) != "us-east-1" {
		t.Error("expected single item unchanged")
	}
}

func TestJoinOrFallback_MultipleItems(t *testing.T) {
	result := joinOrFallback([]string{"us-east-1", "us-west-2"})
	if result != "us-east-1, us-west-2" {
		t.Errorf("unexpected join result: %q", result)
	}
}
