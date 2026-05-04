package schema

import (
	"testing"
)

func TestFilterSchema_NoFilters(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1", "us-west-2"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{} // No filters

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should return original (all organizations present)
	if len(filtered.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(filtered.Managed.Organizations))
	}
}

func TestFilterSchema_OrganizationFilter(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"org1": {
					Name: "Org 1",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://org1.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1"},
						},
					},
				},
				"org2": {
					Name: "Org 2",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://org2.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Accounts:      []Account{{Alias: "prod", Name: "Prod", ID: "222222222222"}},
							Roles:         []string{"Admin"},
							Regions:       []string{"us-west-2"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Organizations: []string{"org1"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have org1
	if len(filtered.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(filtered.Managed.Organizations))
	}

	if _, exists := filtered.Managed.Organizations["org1"]; !exists {
		t.Error("Expected org1 to be present")
	}

	if _, exists := filtered.Managed.Organizations["org2"]; exists {
		t.Error("Expected org2 to be filtered out")
	}
}

func TestFilterSchema_PartitionFilter(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1"},
						},
						"govcloud": {
							URL:           "https://test-gov.awsapps.com/start",
							DefaultRegion: "us-gov-west-1",
							Accounts:      []Account{{Alias: "gov-dev", Name: "Gov Dev", ID: "333333333333"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-gov-west-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Partitions: []string{"commercial"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	org := filtered.Managed.Organizations["test-org"]
	if org == nil {
		t.Fatal("Expected test-org to be present")
	}

	// Should only have commercial partition
	if len(org.Partitions) != 1 {
		t.Errorf("Expected 1 partition, got %d", len(org.Partitions))
	}

	if _, exists := org.Partitions["commercial"]; !exists {
		t.Error("Expected commercial partition to be present")
	}

	if _, exists := org.Partitions["govcloud"]; exists {
		t.Error("Expected govcloud partition to be filtered out")
	}
}

func TestFilterSchema_AccountFilter(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts: []Account{
								{Alias: "dev", Name: "Dev", ID: "111111111111"},
								{Alias: "staging", Name: "Staging", ID: "222222222222"},
								{Alias: "prod", Name: "Prod", ID: "333333333333"},
							},
							Roles:   []string{"Developer"},
							Regions: []string{"us-east-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Accounts: []string{"dev", "staging"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	part := filtered.Managed.Organizations["test-org"].Partitions["commercial"]

	// Should only have dev and staging accounts
	if len(part.Accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(part.Accounts))
	}

	foundDev := false
	foundStaging := false
	foundProd := false

	for _, acc := range part.Accounts {
		if acc.Alias == "dev" {
			foundDev = true
		}
		if acc.Alias == "staging" {
			foundStaging = true
		}
		if acc.Alias == "prod" {
			foundProd = true
		}
	}

	if !foundDev {
		t.Error("Expected dev account to be present")
	}
	if !foundStaging {
		t.Error("Expected staging account to be present")
	}
	if foundProd {
		t.Error("Expected prod account to be filtered out")
	}
}

func TestFilterSchema_RoleFilter(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer", "ReadOnly", "Admin"},
							Regions:       []string{"us-east-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Roles: []string{"Developer", "Admin"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	part := filtered.Managed.Organizations["test-org"].Partitions["commercial"]

	// Should only have Developer and Admin roles
	if len(part.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(part.Roles))
	}

	foundDev := false
	foundAdmin := false
	foundReadOnly := false

	for _, role := range part.Roles {
		if role == "Developer" {
			foundDev = true
		}
		if role == "Admin" {
			foundAdmin = true
		}
		if role == "ReadOnly" {
			foundReadOnly = true
		}
	}

	if !foundDev {
		t.Error("Expected Developer role to be present")
	}
	if !foundAdmin {
		t.Error("Expected Admin role to be present")
	}
	if foundReadOnly {
		t.Error("Expected ReadOnly role to be filtered out")
	}
}

func TestFilterSchema_RegionFilter(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Regions: []string{"us-east-1", "us-west-2"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	part := filtered.Managed.Organizations["test-org"].Partitions["commercial"]

	// Should only have us-east-1 and us-west-2
	if len(part.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(part.Regions))
	}
}

func TestFilterSchema_AllRegions(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		AllRegions: true,
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	part := filtered.Managed.Organizations["test-org"].Partitions["commercial"]

	// Should have all regions
	if len(part.Regions) != 4 {
		t.Errorf("Expected 4 regions, got %d", len(part.Regions))
	}
}

func TestFilterSchema_CombinedFilters(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"org1": {
					Name: "Org 1",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://org1.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts: []Account{
								{Alias: "dev", Name: "Dev", ID: "111111111111"},
								{Alias: "prod", Name: "Prod", ID: "222222222222"},
							},
							Roles:   []string{"Developer", "Admin"},
							Regions: []string{"us-east-1", "us-west-2"},
						},
					},
				},
				"org2": {
					Name: "Org 2",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://org2.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Accounts:      []Account{{Alias: "test", Name: "Test", ID: "333333333333"}},
							Roles:         []string{"ReadOnly"},
							Regions:       []string{"us-west-2"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Organizations: []string{"org1"},
		Accounts:      []string{"dev"},
		Roles:         []string{"Developer"},
		Regions:       []string{"us-east-1"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have org1
	if len(filtered.Managed.Organizations) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(filtered.Managed.Organizations))
	}

	org := filtered.Managed.Organizations["org1"]
	if org == nil {
		t.Fatal("Expected org1 to be present")
	}

	part := org.Partitions["commercial"]

	// Should only have dev account
	if len(part.Accounts) != 1 || part.Accounts[0].Alias != "dev" {
		t.Error("Expected only dev account")
	}

	// Should only have Developer role
	if len(part.Roles) != 1 || part.Roles[0] != "Developer" {
		t.Error("Expected only Developer role")
	}

	// Should only have us-east-1 region
	if len(part.Regions) != 1 || part.Regions[0] != "us-east-1" {
		t.Error("Expected only us-east-1 region")
	}
}

func TestFilterSchema_CaseInsensitive(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"TestOrg": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"Commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "Dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1"},
						},
					},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Organizations: []string{"testorg"},    // Lowercase
		Partitions:    []string{"commercial"}, // Lowercase
		Accounts:      []string{"dev"},        // Lowercase
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should match despite case differences
	if len(filtered.Managed.Organizations) != 1 {
		t.Error("Expected case-insensitive matching for organizations")
	}
}

func TestFilterSchema_PreservesUnmanaged(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Accounts:      []Account{{Alias: "dev", Name: "Dev", ID: "111111111111"}},
							Roles:         []string{"Developer"},
							Regions:       []string{"us-east-1"},
						},
					},
				},
			},
		},
		Unmanaged: &UnmanagedProfiles{
			Above: &ProfileCollection{
				IamUsers: []*IamUser{
					{ProfileName: "personal-iam", Region: "us-west-2"},
				},
			},
		},
	}

	criteria := FilterCriteria{
		Organizations: []string{"test-org"},
	}

	filtered, err := FilterSchema(schema, criteria)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Unmanaged section should be preserved
	if filtered.Unmanaged == nil {
		t.Fatal("Expected unmanaged section to be preserved")
	}

	if len(filtered.Unmanaged.Above.IamUsers) != 1 {
		t.Error("Expected unmanaged IAM users to be preserved")
	}
}
