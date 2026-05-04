package schemalist

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/test"
)

// --- Construction ---

func TestNewAccountListView_WithSchema(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedSsoSingle())
	if v == nil {
		t.Fatal("NewAccountListView should return a view")
	}
	if v.GetContent() == nil {
		t.Fatal("GetContent should return a canvas object")
	}

	records := v.GetAccountRecords()
	if len(records) != 1 {
		t.Fatalf("expected 1 account record, got %d", len(records))
	}
	if records[0].AccountID != "123456789012" {
		t.Fatalf("expected account id 123456789012, got %s", records[0].AccountID)
	}
}

func TestNewAccountListView_NilSchema(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	// Should not panic on nil input.
	v := NewAccountListView(nil)
	if v == nil {
		t.Fatal("NewAccountListView should return a view even for nil schema")
	}
	if v.GetContent() == nil {
		t.Fatal("GetContent should return a canvas object for nil schema")
	}
	if len(v.GetAccountRecords()) != 0 {
		t.Error("expected no records for nil schema")
	}
}

func TestNewAccountListView_EmptySchema(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedEmpty())
	if v == nil {
		t.Fatal("NewAccountListView should return a view for empty schema")
	}
	if len(v.GetAccountRecords()) != 0 {
		t.Error("expected no records for empty managed schema")
	}
}

// --- SetSchema ---

func TestSetSchema_RefreshesRecords(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedSsoSingle())
	v.SetSchema(schematest.NewManagedSsoMultiAccount())

	records := v.GetAccountRecords()
	if len(records) != 3 {
		t.Fatalf("expected 3 account records after schema swap, got %d", len(records))
	}
}

func TestSetSchema_ResetsSelection(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedSsoSingle())
	v.selectedIndex = 0 // simulate a prior selection
	v.SetSchema(schematest.NewManagedSsoMultiAccount())

	if v.selectedIndex != -1 {
		t.Errorf("expected selectedIndex to reset to -1 after SetSchema, got %d", v.selectedIndex)
	}
}

func TestSetSchema_ToNil(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedSsoSingle())
	v.SetSchema(nil) // should not panic

	if len(v.GetAccountRecords()) != 0 {
		t.Error("expected no records after setting nil schema")
	}
}

// --- GetAccountRecords ---

func TestGetAccountRecords_ReturnsCopy(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewAccountListView(schematest.NewManagedSsoSingle())
	records := v.GetAccountRecords()

	// Mutating the returned slice should not affect the view's internal state.
	if len(records) > 0 {
		records[0].AccountID = "mutated"
	}

	records2 := v.GetAccountRecords()
	if len(records2) > 0 && records2[0].AccountID == "mutated" {
		t.Error("GetAccountRecords should return a copy, not a reference")
	}
}

// --- applySearch ---

func TestApplySearch_EmptyQuery_ShowsAll(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoMultiAccount())
	v.searchText = ""
	v.applySearch()

	if len(v.filteredRecords) != len(v.accountRecords) {
		t.Errorf("empty search should show all %d records, got %d", len(v.accountRecords), len(v.filteredRecords))
	}
}

func TestApplySearch_MatchesByAccountAlias(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoMultiAccount())
	v.searchText = "test-dev"
	v.applySearch()

	if len(v.filteredRecords) == 0 {
		t.Fatal("expected at least one result matching alias 'test-dev'")
	}
	for _, r := range v.filteredRecords {
		if r.AccountAlias != "test-dev" {
			t.Errorf("unexpected record in results: alias=%q", r.AccountAlias)
		}
	}
}

func TestApplySearch_MatchesByAccountID(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoSingle())
	v.searchText = "123456789012"
	v.applySearch()

	if len(v.filteredRecords) != 1 {
		t.Fatalf("expected 1 result for account ID search, got %d", len(v.filteredRecords))
	}
}

func TestApplySearch_CaseInsensitive(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoSingle())
	v.searchText = "TEST-ACCT" // upper-case version of "test-acct"
	v.applySearch()

	if len(v.filteredRecords) == 0 {
		t.Error("search should be case-insensitive")
	}
}

func TestApplySearch_NoMatch_ReturnsEmpty(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoSingle())
	v.searchText = "zzznomatchzzz"
	v.applySearch()

	if len(v.filteredRecords) != 0 {
		t.Errorf("expected 0 results for non-matching search, got %d", len(v.filteredRecords))
	}
}

func TestApplySearch_NilAccountRecords(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = nil
	v.searchText = "anything"
	// Should not panic.
	v.applySearch()

	if len(v.filteredRecords) != 0 {
		t.Error("expected empty filtered records for nil account records")
	}
}

func TestApplySearch_MatchesByPartition(t *testing.T) {
	test.SetupTestEnvironment(t)

	v := &AccountListView{}
	v.accountRecords = FlattenManagedAccounts(schematest.NewManagedSsoMultiOrg())
	v.searchText = "govcloud"
	v.applySearch()

	for _, r := range v.filteredRecords {
		if r.PartitionName != "govcloud" {
			t.Errorf("unexpected partition in govcloud-filtered results: %q", r.PartitionName)
		}
	}
}
