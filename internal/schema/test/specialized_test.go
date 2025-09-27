package test

import (
	"testing"
)

// =============================================================================
// EMPTY SCHEMA TESTS
// =============================================================================

func TestNewEmpty(t *testing.T) {
	schema := NewEmpty()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", schema.Version)
	}
	if schema.Managed != nil {
		t.Error("Managed section should be nil")
	}
	if schema.Unmanaged != nil {
		t.Error("Unmanaged section should be nil")
	}
}

func TestNewManagedOnlyEmpty(t *testing.T) {
	schema := NewManagedOnlyEmpty()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}
	if schema.Unmanaged != nil {
		t.Error("Unmanaged section should be nil")
	}

	// Verify managed section is empty
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

func TestNewUnmanagedOnlyEmpty(t *testing.T) {
	schema := NewUnmanagedOnlyEmpty()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed != nil {
		t.Error("Managed section should be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}

	// Above and Below should both be nil
	if schema.Unmanaged.Above != nil {
		t.Error("Above section should be nil")
	}
	if schema.Unmanaged.Below != nil {
		t.Error("Below section should be nil")
	}
}

// =============================================================================
// INVALID SCHEMA TESTS
// =============================================================================

func TestNewInvalid(t *testing.T) {
	schema := NewInvalid()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify version is invalid
	if schema.Version == "1.0" {
		t.Error("Version should be invalid for testing")
	}

	// Verify managed section exists with invalid data
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify invalid organization exists
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected invalid organization")
	}

	// Verify invalid IAM user exists
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected invalid IAM user")
	}

	// Verify invalid AssumeRole exists
	if len(schema.Managed.AssumeRoleChains) == 0 {
		t.Error("Expected invalid AssumeRole chain")
	}

	// Verify invalid Generic profile exists
	if len(schema.Managed.GenericProfiles) == 0 {
		t.Error("Expected invalid Generic profile")
	}
}

func TestNewInvalidMissingRequired(t *testing.T) {
	schema := NewInvalidMissingRequired()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}

	// Verify managed section exists
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify organization with missing fields exists
	if len(schema.Managed.Organizations) == 0 {
		t.Error("Expected organization with missing fields")
	}

	// Verify IAM user with missing fields exists
	if len(schema.Managed.IamUsers) == 0 {
		t.Error("Expected IAM user with missing fields")
	}

	// Verify AssumeRole with missing fields exists
	if len(schema.Managed.AssumeRoleChains) == 0 {
		t.Error("Expected AssumeRole chain with missing fields")
	}
}

// =============================================================================
// LARGE SCALE SCHEMA TESTS
// =============================================================================

func TestNewLargeScale(t *testing.T) {
	schema := NewLargeScale()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify large number of organizations
	if len(schema.Managed.Organizations) < 5 {
		t.Errorf("Expected at least 5 organizations, got %d", len(schema.Managed.Organizations))
	}

	// Verify large number of IAM users
	if len(schema.Managed.IamUsers) < 10 {
		t.Errorf("Expected at least 10 IAM users, got %d", len(schema.Managed.IamUsers))
	}

	// Verify large number of AssumeRole chains
	if len(schema.Managed.AssumeRoleChains) < 20 {
		t.Errorf("Expected at least 20 AssumeRole chains, got %d", len(schema.Managed.AssumeRoleChains))
	}

	// Verify large number of Generic profiles
	if len(schema.Managed.GenericProfiles) < 30 {
		t.Errorf("Expected at least 30 Generic profiles, got %d", len(schema.Managed.GenericProfiles))
	}
}

func TestNewLargeScaleUnmanaged(t *testing.T) {
	schema := NewLargeScaleUnmanaged()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Unmanaged == nil {
		t.Fatal("Unmanaged section should not be nil")
	}
	if schema.Unmanaged.Below == nil {
		t.Fatal("Below section should not be nil")
	}

	// Verify large number of generic profiles
	if len(schema.Unmanaged.Below.GenericProfiles) < 50 {
		t.Errorf("Expected at least 50 generic profiles, got %d", len(schema.Unmanaged.Below.GenericProfiles))
	}
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

func TestNewSingleProfileAllTypes(t *testing.T) {
	schema := NewSingleProfileAllTypes()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify exactly 1 of each type
	if len(schema.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(schema.Managed.Organizations))
	}
	if len(schema.Managed.IamUsers) != 1 {
		t.Errorf("Expected 1 IAM user, got %d", len(schema.Managed.IamUsers))
	}
	if len(schema.Managed.AssumeRoleChains) != 1 {
		t.Errorf("Expected 1 AssumeRole chain, got %d", len(schema.Managed.AssumeRoleChains))
	}
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Errorf("Expected 1 Generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	// Verify the SSO org has exactly 1 account and 1 role
	for _, org := range schema.Managed.Organizations {
		for _, partition := range org.Partitions {
			if len(partition.Accounts) != 1 {
				t.Errorf("Expected 1 account, got %d", len(partition.Accounts))
			}
			if len(partition.Roles) != 1 {
				t.Errorf("Expected 1 role, got %d", len(partition.Roles))
			}
		}
	}
}

func TestNewMinimal(t *testing.T) {
	schema := NewMinimal()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Verify only Generic profiles exist
	if len(schema.Managed.Organizations) != 0 {
		t.Error("Should not have organizations")
	}
	if len(schema.Managed.IamUsers) != 0 {
		t.Error("Should not have IAM users")
	}
	if len(schema.Managed.AssumeRoleChains) != 0 {
		t.Error("Should not have AssumeRole chains")
	}
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Errorf("Expected 1 Generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	// Verify minimal properties
	profile := schema.Managed.GenericProfiles[0]
	if profile.ProfileName != "minimal" {
		t.Errorf("Expected 'minimal', got %s", profile.ProfileName)
	}
	if len(profile.Properties) < 1 {
		t.Errorf("Expected at least 1 property, got %d", len(profile.Properties))
	}
}

// =============================================================================
// PARTIAL/MISSING DATA SCHEMA TESTS
// =============================================================================

func TestNewPartialSsoNoAccounts(t *testing.T) {
	schema := NewPartialSsoNoAccounts()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 organization
	if len(schema.Managed.Organizations) != 1 {
		t.Fatalf("Expected 1 organization, got %d", len(schema.Managed.Organizations))
	}

	org := schema.Managed.Organizations["no-accounts"]
	if org == nil {
		t.Fatal("Organization 'no-accounts' should exist")
	}

	// Should have 1 partition
	if len(org.Partitions) != 1 {
		t.Fatalf("Expected 1 partition, got %d", len(org.Partitions))
	}

	partition := org.Partitions["commercial"]
	if len(partition.Accounts) != 0 {
		t.Errorf("Expected 0 accounts, got %d", len(partition.Accounts))
	}
	if len(partition.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(partition.Roles))
	}
}

func TestNewPartialSsoNoRoles(t *testing.T) {
	schema := NewPartialSsoNoRoles()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	org := schema.Managed.Organizations["no-roles"]
	if org == nil {
		t.Fatal("Organization 'no-roles' should exist")
	}

	partition := org.Partitions["commercial"]
	if len(partition.Accounts) != 1 {
		t.Errorf("Expected 1 account, got %d", len(partition.Accounts))
	}
	if len(partition.Roles) != 0 {
		t.Errorf("Expected 0 roles, got %d", len(partition.Roles))
	}
}

func TestNewPartialSsoNoPartitions(t *testing.T) {
	schema := NewPartialSsoNoPartitions()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	org := schema.Managed.Organizations["no-partitions"]
	if org == nil {
		t.Fatal("Organization 'no-partitions' should exist")
	}

	if len(org.Partitions) != 0 {
		t.Errorf("Expected 0 partitions, got %d", len(org.Partitions))
	}
}

func TestNewPartialIamMissingCredentials(t *testing.T) {
	schema := NewPartialIamMissingCredentials()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 IAM user
	if len(schema.Managed.IamUsers) != 1 {
		t.Fatalf("Expected 1 IAM user, got %d", len(schema.Managed.IamUsers))
	}

	user := schema.Managed.IamUsers[0]
	if user.ProfileName != "missing-creds" {
		t.Errorf("Expected profile name 'missing-creds', got %s", user.ProfileName)
	}
	if user.AwsAccessKeyID != "" {
		t.Errorf("Expected empty access key, got %s", user.AwsAccessKeyID)
	}
	if user.AwsSecretKey != "" {
		t.Errorf("Expected empty secret key, got %s", user.AwsSecretKey)
	}
}

func TestNewPartialAssumeRoleMissingSource(t *testing.T) {
	schema := NewPartialAssumeRoleMissingSource()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 AssumeRole chain
	if len(schema.Managed.AssumeRoleChains) != 1 {
		t.Fatalf("Expected 1 AssumeRole chain, got %d", len(schema.Managed.AssumeRoleChains))
	}

	chain := schema.Managed.AssumeRoleChains[0]
	if chain.ProfileName != "missing-source" {
		t.Errorf("Expected profile name 'missing-source', got %s", chain.ProfileName)
	}
	if chain.SourceProfile != "" {
		t.Errorf("Expected empty source profile, got %s", chain.SourceProfile)
	}
	if chain.RoleArn == "" {
		t.Error("RoleArn should not be empty")
	}
}

func TestNewPartialGenericEmptyName(t *testing.T) {
	schema := NewPartialGenericEmptyName()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 generic profile
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Fatalf("Expected 1 generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	profile := schema.Managed.GenericProfiles[0]
	if profile.ProfileName != "" {
		t.Errorf("Expected empty profile name, got %s", profile.ProfileName)
	}
	if len(profile.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(profile.Properties))
	}
}

func TestNewPartialGenericEmptyProperties(t *testing.T) {
	schema := NewPartialGenericEmptyProperties()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 generic profile
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Fatalf("Expected 1 generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	profile := schema.Managed.GenericProfiles[0]
	if profile.ProfileName != "empty-props" {
		t.Errorf("Expected profile name 'empty-props', got %s", profile.ProfileName)
	}
	if len(profile.Properties) != 0 {
		t.Errorf("Expected 0 properties, got %d", len(profile.Properties))
	}
}

func TestNewPartialGenericNoProperties(t *testing.T) {
	schema := NewPartialGenericNoProperties()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 generic profile
	if len(schema.Managed.GenericProfiles) != 1 {
		t.Fatalf("Expected 1 generic profile, got %d", len(schema.Managed.GenericProfiles))
	}

	profile := schema.Managed.GenericProfiles[0]
	if profile.ProfileName != "no-props" {
		t.Errorf("Expected profile name 'no-props', got %s", profile.ProfileName)
	}
	if profile.Properties != nil {
		t.Errorf("Expected nil properties, got non-nil")
	}
}

func TestNewPartialIamNoCredentials(t *testing.T) {
	schema := NewPartialIamNoCredentials()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 IAM user
	if len(schema.Managed.IamUsers) != 1 {
		t.Fatalf("Expected 1 IAM user, got %d", len(schema.Managed.IamUsers))
	}

	user := schema.Managed.IamUsers[0]
	if user.ProfileName != "no-creds" {
		t.Errorf("Expected profile name 'no-creds', got %s", user.ProfileName)
	}
	if user.AwsAccessKeyID != "" {
		t.Errorf("Expected empty access key, got %s", user.AwsAccessKeyID)
	}
	if user.AwsSecretKey != "" {
		t.Errorf("Expected empty secret key, got %s", user.AwsSecretKey)
	}
}

func TestNewPartialAssumeRoleNoArn(t *testing.T) {
	schema := NewPartialAssumeRoleNoArn()

	if schema == nil {
		t.Fatal("Schema should not be nil")
	}
	if schema.Managed == nil {
		t.Fatal("Managed section should not be nil")
	}

	// Should have 1 AssumeRole profile
	if len(schema.Managed.AssumeRoleChains) != 1 {
		t.Fatalf("Expected 1 AssumeRole profile, got %d", len(schema.Managed.AssumeRoleChains))
	}

	chain := schema.Managed.AssumeRoleChains[0]
	if chain.ProfileName != "no-arn" {
		t.Errorf("Expected profile name 'no-arn', got %s", chain.ProfileName)
	}
	if chain.RoleArn != "" {
		t.Errorf("Expected empty role ARN, got %s", chain.RoleArn)
	}
	if chain.SourceProfile != "source-profile" {
		t.Errorf("Expected source profile 'source-profile', got %s", chain.SourceProfile)
	}
}
