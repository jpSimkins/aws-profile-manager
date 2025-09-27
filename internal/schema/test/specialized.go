// Package test provides test schema fixtures for use across the codebase.
//
// This file contains specialized schemas for edge cases, error testing,
// and performance testing scenarios.
package test

import (
	"aws-profile-manager/internal/schema"
	"fmt"
)

// =============================================================================
// EMPTY SCHEMAS
// =============================================================================

// NewEmpty creates a completely empty schema.
//
// Returns a schema with no managed or unmanaged sections.
// Useful for testing error handling when schema has no data.
func NewEmpty() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
	}
}

// NewManagedOnlyEmpty creates a schema with only an empty managed section.
//
// Returns a schema with an empty managed ProfileCollection but no unmanaged section.
// Useful for testing validation when managed section exists but is empty.
//
// Note: This is different from NewManagedEmpty() which is in managed.go
// and includes the empty ProfileCollection struct.
func NewManagedOnlyEmpty() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{},
	}
}

// NewUnmanagedOnlyEmpty creates a schema with only an empty unmanaged section.
//
// Returns a schema with an empty unmanaged section but no managed section.
// Useful for testing when only personal profiles exist (but there are none).
func NewUnmanagedOnlyEmpty() *schema.Schema {
	return &schema.Schema{
		Version:   "1.0",
		Unmanaged: &schema.UnmanagedProfiles{},
	}
}

// =============================================================================
// INVALID SCHEMAS
// =============================================================================

// NewInvalid creates an intentionally invalid schema for error testing.
//
// Returns a schema that will fail validation:
//   - Invalid version string
//   - SSO account with invalid ID format
//   - IAM user with empty required fields
//   - AssumeRole with malformed ARN
//   - Generic profile with empty name
//
// Useful for testing error handling, validation logic, and error reporting.
func NewInvalid() *schema.Schema {
	return &schema.Schema{
		Version: "invalid-version-format",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"invalid-org": {
					Name:        "", // Empty name (invalid)
					Description: "Invalid organization for testing",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "not-a-valid-url", // Invalid URL format
							DefaultRegion: "invalid-region",  // Invalid region
							Regions:       []string{"invalid-region"},
							Accounts: []schema.Account{
								{
									Alias: "", // Empty alias (invalid)
									Name:  "Invalid Account",
									ID:    "not-a-valid-account-id", // Invalid format
								},
							},
							Roles: []string{""}, // Empty role (invalid)
						},
					},
				},
			},
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "", // Empty profile name (invalid)
					Region:         "invalid-region",
					AwsAccessKeyID: "INVALID", // Too short
					AwsSecretKey:   "",        // Empty (invalid)
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "", // Empty profile name (invalid)
					RoleArn:       "not-a-valid-arn",
					SourceProfile: "", // Empty source (invalid)
					Region:        "invalid-region",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "", // Empty profile name (invalid)
					Properties: map[string]string{
						"": "value", // Empty key (invalid)
					},
				},
			},
		},
	}
}

// NewInvalidMissingRequired creates a schema missing required fields.
//
// Returns a schema with profiles that are missing required fields:
//   - SSO partition with no URL
//   - IAM user with no credentials
//   - AssumeRole with no role ARN
//
// Useful for testing required field validation.
func NewInvalidMissingRequired() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"missing-fields": {
					Name:        "Missing Fields Org",
					Description: "Organization with missing required fields",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "", // Missing (required)
							DefaultRegion: "", // Missing (required)
							Regions:       []string{},
							Accounts:      []schema.Account{}, // Empty (invalid)
							Roles:         []string{},         // Empty (invalid)
						},
					},
				},
			},
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "missing-creds",
					Region:         "", // Missing (required)
					AwsAccessKeyID: "", // Missing (required)
					AwsSecretKey:   "", // Missing (required)
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "missing-arn",
					RoleArn:       "", // Missing (required)
					SourceProfile: "", // Missing (required)
					Region:        "", // Missing (required)
				},
			},
		},
	}
}

// =============================================================================
// LARGE SCALE SCHEMAS
// =============================================================================

// NewLargeScale creates a large schema for performance and stress testing.
//
// Contains:
//   - 10 Organizations
//   - 100+ AWS accounts
//   - 5 roles per partition = 500+ profiles
//   - 20 IAM users
//   - 30 AssumeRole chains
//   - 50 Generic profiles
//   - Total: 600+ profiles
//
// Useful for testing:
//   - Performance with large datasets
//   - Memory usage
//   - File I/O with large configs
//   - Filtering performance
//   - UI responsiveness with many items
func NewLargeScale() *schema.Schema {
	orgs := make(map[string]*schema.Organization)

	// Create 10 organizations with 10 accounts each
	for orgNum := 1; orgNum <= 10; orgNum++ {
		accounts := make([]schema.Account, 10)
		for accNum := 1; accNum <= 10; accNum++ {
			// Generate proper 12-digit account ID
			accountID := fmt.Sprintf("%012d", 100000000000+(orgNum*1000)+accNum)
			accounts[accNum-1] = schema.Account{
				Alias: fmt.Sprintf("account-%d-%d", orgNum, accNum),
				Name:  fmt.Sprintf("Account %d-%d", orgNum, accNum),
				ID:    accountID,
			}
		}

		orgAlias := fmt.Sprintf("org-%d", orgNum)
		orgs[orgAlias] = &schema.Organization{
			Name:        fmt.Sprintf("Organization %d", orgNum),
			Description: fmt.Sprintf("Large scale test organization %d", orgNum),
			Partitions: map[string]schema.Partition{
				"commercial": {
					URL:           fmt.Sprintf("https://%s.awsapps.com/start", orgAlias),
					DefaultRegion: "us-west-2",
					Regions:       []string{"us-west-2", "us-east-1", "eu-west-1", "ap-southeast-1"},
					Accounts:      accounts,
					Roles:         []string{"Developer", "Admin", "ReadOnly", "PowerUser", "Auditor"},
				},
			},
		}
	}

	// Create 20 IAM users
	iamUsers := make([]*schema.IamUser, 20)
	for i := 0; i < 20; i++ {
		iamUsers[i] = &schema.IamUser{
			ProfileName:    fmt.Sprintf("iam-user-%d", i+1),
			Region:         "us-east-1",
			AwsAccessKeyID: fmt.Sprintf("AKIAIOSFODNN7EXAMP%02d", i),
			AwsSecretKey:   fmt.Sprintf("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMP%02d", i),
		}
	}

	// Create 30 AssumeRole chains
	assumeRoles := make([]*schema.AssumeRoleChain, 30)
	for i := 0; i < 30; i++ {
		assumeRoles[i] = &schema.AssumeRoleChain{
			ProfileName:   fmt.Sprintf("assume-role-%d", i+1),
			RoleArn:       fmt.Sprintf("arn:aws:iam::123456789012:role/Role%d", i+1),
			SourceProfile: "iam-user-1",
			Region:        "us-east-1",
		}
	}

	// Create 50 Generic profiles
	genericProfiles := make([]*schema.GenericProfile, 50)
	for i := 0; i < 50; i++ {
		genericProfiles[i] = &schema.GenericProfile{
			ProfileName: fmt.Sprintf("generic-%d", i+1),
			Properties: map[string]string{
				"region": "us-west-2",
				"output": "json",
			},
		}
	}

	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations:    orgs,
			IamUsers:         iamUsers,
			AssumeRoleChains: assumeRoles,
			GenericProfiles:  genericProfiles,
		},
	}
} // NewLargeScaleUnmanaged creates a large unmanaged schema for testing personal profile limits.
// Contains:
//   - 100 Personal generic profiles (Below section)
//
// Useful for testing personal profile preservation at scale.
func NewLargeScaleUnmanaged() *schema.Schema {
	genericProfiles := make([]*schema.GenericProfile, 100)
	for i := 0; i < 100; i++ {
		genericProfiles[i] = &schema.GenericProfile{
			ProfileName: "personal-" + string(rune('0'+i+1)),
			Properties: map[string]string{
				"region": "us-east-1",
				"output": "json",
			},
		}
	}

	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				GenericProfiles: genericProfiles,
			},
		},
	}
}

// =============================================================================
// EDGE CASES
// =============================================================================

// NewSingleProfileAllTypes creates a schema with exactly one profile of each type.
//
// Contains:
//   - 1 SSO profile (minimal: 1 org, 1 partition, 1 account, 1 role)
//   - 1 IAM user
//   - 1 AssumeRole chain
//   - 1 Generic profile
//
// Useful for testing that each profile type is handled correctly in isolation.
func NewSingleProfileAllTypes() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Single profile test",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "test-acct",
									Name:  "Test Account",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "single-iam",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "single-assume",
					RoleArn:       "arn:aws:iam::123456789012:role/TestRole",
					SourceProfile: "single-iam",
					Region:        "us-east-1",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "single-generic",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "json",
					},
				},
			},
		},
	}
}

// NewMinimal creates the absolute minimum valid schema.
//
// Contains:
//   - 1 Generic profile with minimal properties
//
// The smallest possible valid schema for testing baseline functionality.
func NewMinimal() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "minimal",
					Properties: map[string]string{
						"region": "us-east-1",
					},
				},
			},
		},
	}
}

// =============================================================================
// PARTIAL/MISSING DATA SCHEMAS (for validation testing)
// =============================================================================

// NewPartialSsoNoAccounts creates a schema with SSO org but no accounts.
//
// Contains:
//   - 1 Organization with valid partition but empty accounts list
//
// Useful for testing validation when SSO partition has no accounts.
func NewPartialSsoNoAccounts() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"no-accounts": {
					Name:        "No Accounts Org",
					Description: "Organization with no accounts",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts:      []schema.Account{}, // Empty accounts
							Roles:         []string{"Developer"},
						},
					},
				},
			},
		},
	}
}

// NewPartialSsoNoRoles creates a schema with SSO org but no roles.
//
// Contains:
//   - 1 Organization with valid accounts but empty roles list
//
// Useful for testing validation when SSO partition has no roles.
func NewPartialSsoNoRoles() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"no-roles": {
					Name:        "No Roles Org",
					Description: "Organization with no roles",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "test-acct",
									Name:  "Test Account",
									ID:    "123456789012",
								},
							},
							Roles: []string{}, // Empty roles
						},
					},
				},
			},
		},
	}
}

// NewPartialSsoNoPartitions creates a schema with SSO org but no partitions.
//
// Contains:
//   - 1 Organization with empty partitions map
//
// Useful for testing validation when organization has no partitions.
func NewPartialSsoNoPartitions() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"no-partitions": {
					Name:        "No Partitions Org",
					Description: "Organization with no partitions",
					Partitions:  map[string]schema.Partition{}, // Empty partitions
				},
			},
		},
	}
}

// NewPartialIamMissingCredentials creates a schema with IAM user missing credentials.
//
// Contains:
//   - 1 IAM user with profile name but no access key or secret
//
// Useful for testing validation of required IAM fields.
func NewPartialIamMissingCredentials() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "missing-creds",
					Region:         "us-east-1",
					AwsAccessKeyID: "", // Missing
					AwsSecretKey:   "", // Missing
				},
			},
		},
	}
}

// NewPartialAssumeRoleMissingSource creates a schema with AssumeRole missing source profile.
//
// Contains:
//   - 1 AssumeRole chain with no source profile
//
// Useful for testing validation of AssumeRole source profile requirement.
func NewPartialAssumeRoleMissingSource() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "missing-source",
					RoleArn:       "arn:aws:iam::123456789012:role/TestRole",
					SourceProfile: "", // Missing source
					Region:        "us-east-1",
				},
			},
		},
	}
}

// NewPartialGenericEmptyName creates a schema with Generic profile missing profile name.
//
// Contains:
//   - 1 Generic profile with empty name
//
// Useful for testing validation of required profile name.
func NewPartialGenericEmptyName() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "", // Empty name
					Properties: map[string]string{
						"region": "us-east-1",
						"output": "json",
					},
				},
			},
		},
	}
}

// NewPartialGenericEmptyProperties creates a schema with Generic profile with no properties.
//
// Contains:
//   - 1 Generic profile with empty properties map
//
// Useful for testing validation when generic profile has no configuration.
func NewPartialGenericEmptyProperties() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "empty-props",
					Properties:  map[string]string{}, // Empty properties
				},
			},
		},
	}
}

// NewPartialGenericNoProperties creates a schema with Generic profile with nil properties.
//
// Contains:
//   - 1 Generic profile with nil properties map
//
// Useful for testing validation when generic profile has no configuration at all.
func NewPartialGenericNoProperties() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "no-props",
					Properties:  nil, // Nil properties
				},
			},
		},
	}
}

// NewPartialIamNoCredentials creates a schema with IAM user missing credentials.
//
// Contains:
//   - 1 IAM user with empty access key and secret
//
// Useful for testing validation of required IAM credentials.
func NewPartialIamNoCredentials() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "no-creds",
					Region:         "us-east-1",
					AwsAccessKeyID: "", // Missing
					AwsSecretKey:   "", // Missing
				},
			},
		},
	}
}

// NewPartialAssumeRoleNoArn creates a schema with AssumeRole missing ARN.
//
// Contains:
//   - 1 AssumeRole profile with empty role ARN
//
// Useful for testing validation of required role ARN.
func NewPartialAssumeRoleNoArn() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "no-arn",
					RoleArn:       "", // Missing ARN
					SourceProfile: "source-profile",
					Region:        "us-east-1",
				},
			},
		},
	}
}
