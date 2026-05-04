package test

import (
	"testing"
)

// =============================================================================
// SSO PROFILE TESTS
// =============================================================================

func TestNewUnmanagedSsoSingle(t *testing.T) {
	schema := NewUnmanagedSsoSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schema.Version)
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if schema.Managed != nil {
		t.Error("Managed section should be nil")
	}

	if len(schema.Unmanaged.Below.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(schema.Unmanaged.Below.Organizations))
	}

	org, exists := schema.Unmanaged.Below.Organizations["personal-org"]
	if !exists {
		t.Fatal("personal-org should exist")
	}
	if org.Name != "Personal Org" {
		t.Errorf("Expected 'Personal Org', got %s", org.Name)
	}
}

func TestNewUnmanagedSsoMulti(t *testing.T) {
	schema := NewUnmanagedSsoMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.Organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(schema.Unmanaged.Below.Organizations))
	}

	personalOrg := schema.Unmanaged.Below.Organizations["personal-org"]
	freelanceOrg := schema.Unmanaged.Below.Organizations["freelance-org"]
	if personalOrg == nil || freelanceOrg == nil {
		t.Fatal("Both personal-org and freelance-org should exist")
	}
}

// =============================================================================
// IAM PROFILE TESTS
// =============================================================================

func TestNewUnmanagedIamSingle(t *testing.T) {
	schema := NewUnmanagedIamSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.IamUsers) != 1 {
		t.Errorf("Expected 1 IAM user, got %d", len(schema.Unmanaged.Below.IamUsers))
	}

	user := schema.Unmanaged.Below.IamUsers[0]
	if user.ProfileName != "personal-iam" {
		t.Errorf("Expected 'personal-iam', got %s", user.ProfileName)
	}
}

func TestNewUnmanagedIamMulti(t *testing.T) {
	schema := NewUnmanagedIamMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.IamUsers) != 3 {
		t.Errorf("Expected 3 IAM users, got %d", len(schema.Unmanaged.Below.IamUsers))
	}

	// Verify all users have unique names
	names := make(map[string]bool)
	for _, user := range schema.Unmanaged.Below.IamUsers {
		if names[user.ProfileName] {
			t.Errorf("Duplicate profile name: %s", user.ProfileName)
		}
		names[user.ProfileName] = true
	}
}

// =============================================================================
// ASSUMEROLE PROFILE TESTS
// =============================================================================

func TestNewUnmanagedAssumeRoleSingle(t *testing.T) {
	schema := NewUnmanagedAssumeRoleSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.AssumeRoleChains) != 1 {
		t.Errorf("Expected 1 AssumeRole chain, got %d", len(schema.Unmanaged.Below.AssumeRoleChains))
	}

	chain := schema.Unmanaged.Below.AssumeRoleChains[0]
	if chain.ProfileName != "personal-admin" {
		t.Errorf("Expected 'personal-admin', got %s", chain.ProfileName)
	}
}

func TestNewUnmanagedAssumeRoleMulti(t *testing.T) {
	schema := NewUnmanagedAssumeRoleMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.AssumeRoleChains) != 3 {
		t.Errorf("Expected 3 AssumeRole chains, got %d", len(schema.Unmanaged.Below.AssumeRoleChains))
	}

	// Verify all chains have unique names
	names := make(map[string]bool)
	for _, chain := range schema.Unmanaged.Below.AssumeRoleChains {
		if names[chain.ProfileName] {
			t.Errorf("Duplicate profile name: %s", chain.ProfileName)
		}
		names[chain.ProfileName] = true
	}
}

// =============================================================================
// GENERIC PROFILE TESTS
// =============================================================================

func TestNewUnmanagedGenericSingle(t *testing.T) {
	schema := NewUnmanagedGenericSingle()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.GenericProfiles) != 1 {
		t.Errorf("Expected 1 generic profile, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}

	profile := schema.Unmanaged.Below.GenericProfiles[0]
	if profile.ProfileName != "personal" {
		t.Errorf("Expected 'personal', got %s", profile.ProfileName)
	}
}

func TestNewUnmanagedGenericMulti(t *testing.T) {
	schema := NewUnmanagedGenericMulti()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	if len(schema.Unmanaged.Below.GenericProfiles) != 5 {
		t.Errorf("Expected 5 generic profiles, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}

	// Verify all profiles have unique names
	names := make(map[string]bool)
	for _, profile := range schema.Unmanaged.Below.GenericProfiles {
		if names[profile.ProfileName] {
			t.Errorf("Duplicate profile name: %s", profile.ProfileName)
		}
		names[profile.ProfileName] = true
	}
}

// =============================================================================
// COMBINED PROFILE TYPE TESTS
// =============================================================================

func TestNewUnmanagedAll(t *testing.T) {
	schema := NewUnmanagedAll()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	// Verify all profile types exist
	if len(schema.Unmanaged.Below.Organizations) == 0 {
		t.Error("Expected SSO organizations")
	}
	if len(schema.Unmanaged.Below.IamUsers) == 0 {
		t.Error("Expected IAM users")
	}
	if len(schema.Unmanaged.Below.AssumeRoleChains) == 0 {
		t.Error("Expected AssumeRole chains")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) == 0 {
		t.Error("Expected Generic profiles")
	}
}

func TestNewUnmanagedMixed(t *testing.T) {
	schema := NewUnmanagedMixed()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	// Verify SSO, IAM, and Generic exist
	if len(schema.Unmanaged.Below.Organizations) == 0 {
		t.Error("Expected SSO organizations")
	}
	if len(schema.Unmanaged.Below.IamUsers) == 0 {
		t.Error("Expected IAM users")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) == 0 {
		t.Error("Expected Generic profiles")
	}

	// Verify AssumeRole doesn't exist
	if len(schema.Unmanaged.Below.AssumeRoleChains) != 0 {
		t.Error("Should not have AssumeRole chains")
	}
}

func TestNewUnmanagedAboveBelow(t *testing.T) {
	schema := NewUnmanagedAboveBelow()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Above == nil {
		t.Fatal("Above section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	// Verify Above has profiles
	if len(schema.Unmanaged.Above.GenericProfiles) != 2 {
		t.Errorf("Expected 2 generic profiles in Above, got %d", len(schema.Unmanaged.Above.GenericProfiles))
	}

	// Verify Below has profiles
	if len(schema.Unmanaged.Below.GenericProfiles) != 3 {
		t.Errorf("Expected 3 generic profiles in Below, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}
}

// =============================================================================
// SPECIAL CASE TESTS
// =============================================================================

func TestNewUnmanagedEmpty(t *testing.T) {
	schema := NewUnmanagedEmpty()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}

	// Above and Below should both be nil (empty unmanaged)
	if schema.Unmanaged.Above != nil {
		t.Error("Above section should be nil")
	}
	if schema.Unmanaged.Below != nil {
		t.Error("Below section should be nil")
	}
}
