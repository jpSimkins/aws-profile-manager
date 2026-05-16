package test

import (
	"testing"
)

// =============================================================================
// SSO PROFILE TESTS
// =============================================================================

func TestNewManagedSsoSingle(t *testing.T) {
	schema := NewManagedSsoSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schema.Version)
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(schema.Managed.Organizations))
	}
	if schema.Unmanaged != nil {
		t.Error("Unmanaged section should be nil")
	}

	org, exists := schema.Managed.Organizations["test-org"]
	if !exists {
		t.Fatal("test-org should exist")
	}
	if org.Name != "Test Org" {
		t.Errorf("Expected 'Test Org', got %s", org.Name)
	}

	partition, exists := org.Partitions["commercial"]
	if !exists {
		t.Fatal("commercial partition should exist")
	}
	if len(partition.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(partition.Accounts))
	}
	if len(partition.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(partition.Roles))
	}
}

func TestNewManagedSsoMultiAccount(t *testing.T) {
	schema := NewManagedSsoMultiAccount()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	org := schema.Managed.Organizations["test-org"]
	if org == nil {
		t.Fatal("test-org should exist")
	}

	partition := org.Partitions["commercial"]
	if len(partition.Accounts) != 3 {
		t.Errorf("Expected 3 accounts, got %d", len(partition.Accounts))
	}
	if len(partition.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(partition.Roles))
	}
}

func TestNewManagedSsoMultiOrg(t *testing.T) {
	schema := NewManagedSsoMultiOrg()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(schema.Managed.Organizations))
	}

	testOrg := schema.Managed.Organizations["test-org"]
	prodOrg := schema.Managed.Organizations["prod-org"]
	if testOrg == nil || prodOrg == nil {
		t.Fatal("Both test-org and prod-org should exist")
	}

	// Verify prod-org has both commercial and govcloud
	if len(prodOrg.Partitions) != 2 {
		t.Errorf("Expected 2 partitions in prod-org, got %d", len(prodOrg.Partitions))
	}
}

func TestNewManagedSsoComplex(t *testing.T) {
	schema := NewManagedSsoComplex()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 3 {
		t.Errorf("Expected 3 organizations, got %d", len(schema.Managed.Organizations))
	}

	// Verify all three orgs exist
	corp := schema.Managed.Organizations["corp"]
	security := schema.Managed.Organizations["security"]
	sandbox := schema.Managed.Organizations["sandbox"]
	if corp == nil || security == nil || sandbox == nil {
		t.Fatal("All three organizations should exist")
	}

	// Verify corp has 4 accounts
	corpPartition := corp.Partitions["commercial"]
	if len(corpPartition.Accounts) != 4 {
		t.Errorf("Expected 4 accounts in corp, got %d", len(corpPartition.Accounts))
	}

	// Verify security has govcloud
	if len(security.Partitions) != 2 {
		t.Errorf("Expected 2 partitions in security, got %d", len(security.Partitions))
	}
}

// =============================================================================
// IAM PROFILE TESTS
// =============================================================================

func TestNewManagedIamSingle(t *testing.T) {
	schema := NewManagedIamSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.IamUsers) != 1 {
		t.Errorf("Expected 1 IAM user, got %d", len(schema.Managed.IamUsers))
	}

	user := schema.Managed.IamUsers[0]
	if user.ProfileName != "test-iam-user" {
		t.Errorf("Expected 'test-iam-user', got %s", user.ProfileName)
	}
	if user.Region != "us-east-1" {
		t.Errorf("Expected 'us-east-1', got %s", user.Region)
	}
}

func TestNewManagedIamMulti(t *testing.T) {
	schema := NewManagedIamMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.IamUsers) != 3 {
		t.Errorf("Expected 3 IAM users, got %d", len(schema.Managed.IamUsers))
	}

	// Verify all users have unique names
	names := make(map[string]bool)
	for _, user := range schema.Managed.IamUsers {
		if names[user.ProfileName] {
			t.Errorf("Duplicate profile name: %s", user.ProfileName)
		}
		names[user.ProfileName] = true
	}
}

// =============================================================================
// ASSUMEROLE PROFILE TESTS
// =============================================================================

func TestNewManagedAssumeRoleSingle(t *testing.T) {
	schema := NewManagedAssumeRoleSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.AssumeRoleChains) != 1 {
		t.Errorf("Expected 1 AssumeRole chain, got %d", len(schema.Managed.AssumeRoleChains))
	}

	chain := schema.Managed.AssumeRoleChains[0]
	if chain.ProfileName != "assume-admin" {
		t.Errorf("Expected 'assume-admin', got %s", chain.ProfileName)
	}
	if chain.SourceProfile != "test-iam-user" {
		t.Errorf("Expected 'test-iam-user', got %s", chain.SourceProfile)
	}
}

func TestNewManagedAssumeRoleMulti(t *testing.T) {
	schema := NewManagedAssumeRoleMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.AssumeRoleChains) != 3 {
		t.Errorf("Expected 3 AssumeRole chains, got %d", len(schema.Managed.AssumeRoleChains))
	}

	// Verify all chains have unique names
	names := make(map[string]bool)
	for _, chain := range schema.Managed.AssumeRoleChains {
		if names[chain.ProfileName] {
			t.Errorf("Duplicate profile name: %s", chain.ProfileName)
		}
		names[chain.ProfileName] = true
	}
}

// =============================================================================
// GENERIC PROFILE TESTS
// =============================================================================

func TestNewManagedGenericSingle(t *testing.T) {
	schema := NewManagedGenericSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Errorf("Expected 1 generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	profile := schema.Managed.GenericProfiles[0]
	if profile.ProfileName != "custom-profile" {
		t.Errorf("Expected 'custom-profile', got %s", profile.ProfileName)
	}
	if len(profile.Properties) < 2 {
		t.Errorf("Expected at least 2 properties, got %d", len(profile.Properties))
	}
}

func TestNewManagedGenericMulti(t *testing.T) {
	schema := NewManagedGenericMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.GenericProfiles) != 3 {
		t.Errorf("Expected 3 generic profiles, got %d", len(schema.Managed.GenericProfiles))
	}

	// Verify all profiles have unique names
	names := make(map[string]bool)
	for _, profile := range schema.Managed.GenericProfiles {
		if names[profile.ProfileName] {
			t.Errorf("Duplicate profile name: %s", profile.ProfileName)
		}
		names[profile.ProfileName] = true
	}
}

// =============================================================================
// COMBINED PROFILE TYPE TESTS
// =============================================================================

func TestNewManagedAll(t *testing.T) {
	schema := NewManagedAll()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify all profile types exist
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected SSO organizations")
	}
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected IAM users")
	}
	if len(schema.Managed.AssumeRoleChains) == 0 {
		t.Error("Expected AssumeRole chains")
	}
	if len(schema.Managed.GenericProfiles) == 0 {
		t.Error("Expected Generic profiles")
	}
}

func TestNewManagedSsoIam(t *testing.T) {
	schema := NewManagedSsoIam()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify SSO and IAM exist
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected SSO organizations")
	}
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected IAM users")
	}

	// Verify AssumeRole and Generic don't exist
	if len(schema.Managed.AssumeRoleChains) != 0 {
		t.Error("Should not have AssumeRole chains")
	}
	if len(schema.Managed.GenericProfiles) != 0 {
		t.Error("Should not have Generic profiles")
	}
}

// =============================================================================
// SPECIAL CASE TESTS
// =============================================================================

func TestNewManagedEmpty(t *testing.T) {
	schema := NewManagedEmpty()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify everything is empty
	if len(schema.Managed.Organizations) != 0 {
		t.Error("Organizations should be empty")
	}
	if len(schema.Managed.IamUsers) != 0 {
		t.Error("IAM users should be empty")
	}
	if len(schema.Managed.AssumeRoleChains) != 0 {
		t.Error("AssumeRole chains should be empty")
	}
	if len(schema.Managed.GenericProfiles) != 0 {
		t.Error("Generic profiles should be empty")
	}
}

// =============================================================================
// PRESET TESTS
// =============================================================================

func TestNewManagedSsoWithPresets(t *testing.T) {
	schema := NewManagedSsoWithPresets()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if schema.Presets == nil {
		t.Fatal("Presets should not be nil")
	}

	// Verify presets exist
	expectedPresets := []string{"developer", "devsecops", "devsecops-admin", "break-glass"}
	if len(schema.Presets) != len(expectedPresets) {
		t.Errorf("Expected %d presets, got %d", len(expectedPresets), len(schema.Presets))
	}

	for _, presetKey := range expectedPresets {
		preset, exists := schema.Presets[presetKey]
		if !exists {
			t.Errorf("Preset '%s' should exist", presetKey)
			continue
		}
		if preset.Label == "" {
			t.Errorf("Preset '%s' should have a label", presetKey)
		}
	}

	// Verify developer preset
	dev := schema.Presets["developer"]
	if dev == nil {
		t.Fatal("Developer preset should exist")
	}
	if dev.Label != "Developer" {
		t.Errorf("Expected label 'Developer', got '%s'", dev.Label)
	}
	if len(dev.Organizations) != 1 || dev.Organizations[0] != "test-org" {
		t.Errorf("Expected organizations ['test-org'], got %v", dev.Organizations)
	}
	if len(dev.Roles) != 1 || dev.Roles[0] != "Developer" {
		t.Errorf("Expected roles ['Developer'], got %v", dev.Roles)
	}

	// Verify break-glass preset
	breakGlass := schema.Presets["break-glass"]
	if breakGlass == nil {
		t.Fatal("Break-glass preset should exist")
	}
	if len(breakGlass.Organizations) != 0 {
		t.Error("Break-glass should not filter by organization (all orgs)")
	}
	if len(breakGlass.Roles) != 1 || breakGlass.Roles[0] != "BreakGlass" {
		t.Errorf("Expected roles ['BreakGlass'], got %v", breakGlass.Roles)
	}

	// Validate schema
	if err := schema.Validate(); err != nil {
		t.Errorf("Schema validation failed: %v", err)
	}
}

func TestNewManagedSsoWithPresetsComplex(t *testing.T) {
	schema := NewManagedSsoWithPresetsComplex()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Presets == nil {
		t.Fatal("Presets should not be nil")
	}

	// Verify complex presets
	expectedPresets := []string{"dev-all-regions", "prod-only", "commercial-only", "specific-regions"}
	if len(schema.Presets) != len(expectedPresets) {
		t.Errorf("Expected %d presets, got %d", len(expectedPresets), len(schema.Presets))
	}

	// Verify all-regions preset
	allRegions := schema.Presets["dev-all-regions"]
	if allRegions == nil {
		t.Fatal("dev-all-regions preset should exist")
	}
	if !allRegions.AllRegions {
		t.Error("dev-all-regions should have AllRegions=true")
	}

	// Verify partition filter preset
	commercial := schema.Presets["commercial-only"]
	if commercial == nil {
		t.Fatal("commercial-only preset should exist")
	}
	if len(commercial.Partitions) != 1 || commercial.Partitions[0] != "commercial" {
		t.Errorf("Expected partitions ['commercial'], got %v", commercial.Partitions)
	}

	// Verify region filter preset
	usEast := schema.Presets["specific-regions"]
	if usEast == nil {
		t.Fatal("specific-regions preset should exist")
	}
	if len(usEast.Regions) != 1 || usEast.Regions[0] != "us-east-1" {
		t.Errorf("Expected regions ['us-east-1'], got %v", usEast.Regions)
	}

	// Validate schema
	if err := schema.Validate(); err != nil {
		t.Errorf("Schema validation failed: %v", err)
	}
}
