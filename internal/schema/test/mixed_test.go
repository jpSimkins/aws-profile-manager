package test

import (
	"testing"
)

// =============================================================================
// SAME PROFILE TYPES IN BOTH SECTIONS
// =============================================================================

func TestNewMixedSsoSso(t *testing.T) {
	schema := NewMixedSsoSso()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schema.Version)
	}

	// Verify managed section
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 managed organization, got %d", len(schema.Managed.Organizations))
	}
	if _, exists := schema.Managed.Organizations["work-org"]; !exists {
		t.Error("work-org should exist in managed section")
	}

	// Verify unmanaged section
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.Organizations) != 1 {
		t.Errorf("Expected 1 unmanaged organization, got %d", len(schema.Unmanaged.Below.Organizations))
	}
	if _, exists := schema.Unmanaged.Below.Organizations["personal-org"]; !exists {
		t.Error("personal-org should exist in unmanaged section")
	}
}

func TestNewMixedIamIam(t *testing.T) {
	schema := NewMixedIamIam()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed IAM
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.IamUsers) != 1 {
		t.Errorf("Expected 1 managed IAM user, got %d", len(schema.Managed.IamUsers))
	}
	if schema.Managed.IamUsers[0].ProfileName != "work-iam" {
		t.Errorf("Expected 'work-iam', got %s", schema.Managed.IamUsers[0].ProfileName)
	}

	// Verify unmanaged IAM
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.IamUsers) != 1 {
		t.Errorf("Expected 1 unmanaged IAM user, got %d", len(schema.Unmanaged.Below.IamUsers))
	}
	if schema.Unmanaged.Below.IamUsers[0].ProfileName != "personal-iam" {
		t.Errorf("Expected 'personal-iam', got %s", schema.Unmanaged.Below.IamUsers[0].ProfileName)
	}
}

func TestNewMixedAllAll(t *testing.T) {
	schema := NewMixedAllAll()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has all types
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected managed SSO organizations")
	}
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected managed IAM users")
	}
	if len(schema.Managed.AssumeRoleChains) == 0 {
		t.Error("Expected managed AssumeRole chains")
	}
	if len(schema.Managed.GenericProfiles) == 0 {
		t.Error("Expected managed Generic profiles")
	}

	// Verify unmanaged has all types
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.Organizations) == 0 {
		t.Error("Expected unmanaged SSO organizations")
	}
	if len(schema.Unmanaged.Below.IamUsers) == 0 {
		t.Error("Expected unmanaged IAM users")
	}
	if len(schema.Unmanaged.Below.AssumeRoleChains) == 0 {
		t.Error("Expected unmanaged AssumeRole chains")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) == 0 {
		t.Error("Expected unmanaged Generic profiles")
	}
}

// =============================================================================
// DIFFERENT PROFILE TYPES IN EACH SECTION
// =============================================================================

func TestNewMixedSsoIam(t *testing.T) {
	schema := NewMixedSsoIam()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has SSO
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected managed SSO organizations")
	}
	if len(schema.Managed.IamUsers) != 0 {
		t.Error("Should not have managed IAM users")
	}

	// Verify unmanaged has IAM
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.IamUsers) == 0 {
		t.Error("Expected unmanaged IAM users")
	}
	if len(schema.Unmanaged.Below.Organizations) != 0 {
		t.Error("Should not have unmanaged SSO organizations")
	}
}

func TestNewMixedAllSso(t *testing.T) {
	schema := NewMixedAllSso()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has all types
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected managed SSO organizations")
	}
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected managed IAM users")
	}
	if len(schema.Managed.AssumeRoleChains) == 0 {
		t.Error("Expected managed AssumeRole chains")
	}
	if len(schema.Managed.GenericProfiles) == 0 {
		t.Error("Expected managed Generic profiles")
	}

	// Verify unmanaged has only SSO
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.Organizations) == 0 {
		t.Error("Expected unmanaged SSO organizations")
	}
	if len(schema.Unmanaged.Below.IamUsers) != 0 {
		t.Error("Should not have unmanaged IAM users")
	}
	if len(schema.Unmanaged.Below.AssumeRoleChains) != 0 {
		t.Error("Should not have unmanaged AssumeRole chains")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) != 0 {
		t.Error("Should not have unmanaged Generic profiles")
	}
}

func TestNewMixedSsoGeneric(t *testing.T) {
	schema := NewMixedSsoGeneric()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has SSO
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected managed SSO organizations")
	}

	// Verify unmanaged has Generic
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) != 2 {
		t.Errorf("Expected 2 unmanaged generic profiles, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}
}

// =============================================================================
// COMMON TEST SCENARIOS
// =============================================================================

func TestNewMixedSimple(t *testing.T) {
	schema := NewMixedSimple()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has SSO
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 managed organization, got %d", len(schema.Managed.Organizations))
	}

	// Verify unmanaged has Generic
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}
	if len(schema.Unmanaged.Below.GenericProfiles) != 2 {
		t.Errorf("Expected 2 unmanaged generic profiles, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}
}

func TestNewMixedComplex(t *testing.T) {
	schema := NewMixedComplex()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed has multiple orgs and all types
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if len(schema.Managed.Organizations) != 2 {
		t.Errorf("Expected 2 managed organizations, got %d", len(schema.Managed.Organizations))
	}
	if len(schema.Managed.IamUsers) != 2 {
		t.Errorf("Expected 2 managed IAM users, got %d", len(schema.Managed.IamUsers))
	}
	if len(schema.Managed.AssumeRoleChains) != 2 {
		t.Errorf("Expected 2 managed AssumeRole chains, got %d", len(schema.Managed.AssumeRoleChains))
	}
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Errorf("Expected 1 managed generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	// Verify unmanaged has both Above and Below
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
	if len(schema.Unmanaged.Above.GenericProfiles) != 1 {
		t.Errorf("Expected 1 generic profile in Above, got %d", len(schema.Unmanaged.Above.GenericProfiles))
	}

	// Verify Below has multiple types
	if len(schema.Unmanaged.Below.Organizations) != 2 {
		t.Errorf("Expected 2 unmanaged organizations, got %d", len(schema.Unmanaged.Below.Organizations))
	}
	if len(schema.Unmanaged.Below.IamUsers) != 2 {
		t.Errorf("Expected 2 unmanaged IAM users, got %d", len(schema.Unmanaged.Below.IamUsers))
	}
	if len(schema.Unmanaged.Below.GenericProfiles) != 3 {
		t.Errorf("Expected 3 unmanaged generic profiles, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}
}
