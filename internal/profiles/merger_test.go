package profiles

import (
	"testing"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/test"
)

// TestNewMerger tests merger constructor
func TestNewMerger(t *testing.T) {

	merger := newMerger()

	if merger == nil {
		t.Fatal("newMerger should not return nil")
	}
}

// TestMerger_Merge_NoDuplicates tests merging with no duplicates
func TestMerger_Merge_NoDuplicates(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create existing and new collections with different profiles
	existing := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{
				ProfileName: "existing-profile",
				Properties: map[string]string{
					"region": "us-east-1",
				},
			},
		},
	}

	new := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{
				ProfileName: "new-profile",
				Properties: map[string]string{
					"region": "us-west-2",
				},
			},
		},
	}

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	if duplicates.TotalDuplicates != 0 {
		t.Errorf("Should have 0 duplicates, got %d", duplicates)
	}

	if len(result.GenericProfiles) != 2 {
		t.Errorf("Should have 2 profiles, got %d", len(result.GenericProfiles))
	}
}

// TestMerger_Merge_WithGenericDuplicates tests duplicate detection for generic profiles
func TestMerger_Merge_WithGenericDuplicates(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create existing and new collections with duplicate names (case insensitive)
	existing := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{
				ProfileName: "my-profile",
				Properties: map[string]string{
					"region": "us-east-1",
				},
			},
		},
	}

	new := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{
				ProfileName: "MY-PROFILE", // Different case
				Properties: map[string]string{
					"region": "us-west-2",
				},
			},
			{
				ProfileName: "another-profile",
				Properties: map[string]string{
					"region": "eu-west-1",
				},
			},
		},
	}

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	if duplicates.TotalDuplicates != 1 {
		t.Errorf("Should have detected 1 duplicate, got %d", duplicates)
	}

	// Should only have 2 profiles (existing + non-duplicate new)
	if len(result.GenericProfiles) != 2 {
		t.Errorf("Should have 2 profiles after deduplication, got %d", len(result.GenericProfiles))
	}

	// Original should be preserved
	found := false
	for _, p := range result.GenericProfiles {
		if p.ProfileName == "my-profile" && p.Properties["region"] == "us-east-1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Original profile should be preserved")
	}
}

// TestMerger_Merge_WithIamDuplicates tests IAM user duplicate detection
func TestMerger_Merge_WithIamDuplicates(t *testing.T) {
	test.SetupTestEnvironment(t)

	existing := &schema.ProfileCollection{
		IamUsers: []*schema.IamUser{
			{
				ProfileName: "iam-user-1",

				Region: "us-east-1",
			},
		},
	}

	new := &schema.ProfileCollection{
		IamUsers: []*schema.IamUser{
			{
				ProfileName: "IAM-USER-1", // Different case

				Region: "us-west-2",
			},
			{
				ProfileName: "iam-user-2",

				Region: "us-east-1",
			},
		},
	}

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	if duplicates.TotalDuplicates != 1 {
		t.Errorf("Should have detected 1 IAM duplicate, got %d", duplicates)
	}

	if len(result.IamUsers) != 2 {
		t.Errorf("Should have 2 IAM users after deduplication, got %d", len(result.IamUsers))
	}
}

// TestMerger_Merge_WithAssumeRoleDuplicates tests AssumeRole duplicate detection
func TestMerger_Merge_WithAssumeRoleDuplicates(t *testing.T) {
	test.SetupTestEnvironment(t)

	existing := &schema.ProfileCollection{
		AssumeRoleChains: []*schema.AssumeRoleChain{
			{
				ProfileName: "assume-role-1",
				RoleArn:     "arn:aws:iam::123456789012:role/MyRole",
			},
		},
	}

	new := &schema.ProfileCollection{
		AssumeRoleChains: []*schema.AssumeRoleChain{
			{
				ProfileName: "ASSUME-ROLE-1", // Different case
				RoleArn:     "arn:aws:iam::123456789012:role/MyRole",
			},
			{
				ProfileName: "assume-role-2",
				RoleArn:     "arn:aws:iam::123456789012:role/OtherRole",
			},
		},
	}

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	if duplicates.TotalDuplicates != 1 {
		t.Errorf("Should have detected 1 AssumeRole duplicate, got %d", duplicates)
	}

	if len(result.AssumeRoleChains) != 2 {
		t.Errorf("Should have 2 AssumeRole chains after deduplication, got %d", len(result.AssumeRoleChains))
	}
}

// TestMerger_Merge_MixedTypes tests merging all profile types
func TestMerger_Merge_MixedTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Use test schemas
	existing := schematest.NewUnmanagedGenericMulti().Unmanaged.Above
	new := schematest.NewUnmanagedIamMulti().Unmanaged.Above

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	// Should merge without errors
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	// Duplicates depend on the schemas - just verify it ran
	_ = duplicates
}

// TestMerger_Merge_NilCollections tests handling nil collections
func TestMerger_Merge_NilCollections(t *testing.T) {
	test.SetupTestEnvironment(t)

	merger := newMerger()

	// Test nil existing
	result, duplicates := merger.merge(nil, &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{ProfileName: "test", Properties: map[string]string{"region": "us-east-1"}},
		},
	})

	if result == nil {
		t.Error("Result should not be nil when existing is nil")
	}
	if duplicates.TotalDuplicates != 0 {
		t.Error("Should have no duplicates when existing is nil")
	}

	// Test nil new
	result, duplicates = merger.merge(&schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{ProfileName: "test", Properties: map[string]string{"region": "us-east-1"}},
		},
	}, nil)

	if result == nil {
		t.Error("Result should not be nil when new is nil")
	}
	if duplicates.TotalDuplicates != 0 {
		t.Error("Should have no duplicates when new is nil")
	}

	// Test both nil
	result, duplicates = merger.merge(nil, nil)
	if result == nil {
		t.Error("Result should not be nil when both are nil")
	}
	if duplicates.TotalDuplicates != 0 {
		t.Error("Should have no duplicates when both are nil")
	}
}

// TestMerger_Merge_EmptyCollections tests merging empty collections
func TestMerger_Merge_EmptyCollections(t *testing.T) {
	test.SetupTestEnvironment(t)

	merger := newMerger()

	existing := &schema.ProfileCollection{}
	new := &schema.ProfileCollection{}

	result, duplicates := merger.merge(existing, new)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if duplicates.TotalDuplicates != 0 {
		t.Errorf("Should have 0 duplicates with empty collections, got %d", duplicates)
	}
}

// TestMerger_Merge_PreservesExisting tests that existing profiles are preserved
func TestMerger_Merge_PreservesExisting(t *testing.T) {
	test.SetupTestEnvironment(t)

	existingProfile := &schema.GenericProfile{
		ProfileName: "important-profile",
		Properties: map[string]string{
			"region":      "us-east-1",
			"output":      "json",
			"custom_prop": "important_value",
		},
	}

	existing := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{existingProfile},
	}

	// Try to add duplicate with different properties
	new := &schema.ProfileCollection{
		GenericProfiles: []*schema.GenericProfile{
			{
				ProfileName: "IMPORTANT-PROFILE", // Same name, different case
				Properties: map[string]string{
					"region": "us-west-2", // Different region
					"output": "yaml",      // Different output
				},
			},
		},
	}

	merger := newMerger()
	result, duplicates := merger.merge(existing, new)

	if duplicates.TotalDuplicates != 1 {
		t.Errorf("Should have detected duplicate, got %d", duplicates)
	}

	// Find the preserved profile
	var preserved *schema.GenericProfile
	for _, p := range result.GenericProfiles {
		if p.ProfileName == "important-profile" {
			preserved = p
			break
		}
	}

	if preserved == nil {
		t.Fatal("Original profile should be preserved")
	}

	// Verify original properties are unchanged
	if preserved.Properties["region"] != "us-east-1" {
		t.Error("Original region should be preserved")
	}
	if preserved.Properties["output"] != "json" {
		t.Error("Original output should be preserved")
	}
	if preserved.Properties["custom_prop"] != "important_value" {
		t.Error("Original custom properties should be preserved")
	}
}
