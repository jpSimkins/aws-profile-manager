// Package test provides test schema fixtures for use across the codebase.
//
// This file contains managed section schemas - profiles that the tool creates
// and manages within the marked section of AWS config files.
package test

import (
	"aws-profile-manager/internal/schema"
)

// =============================================================================
// SSO PROFILES
// =============================================================================

// NewManagedSsoSingle creates a schema with a single SSO profile.
//
// Contains:
//   - 1 Organization (test-org)
//   - 1 Partition (commercial)
//   - 1 Account (123456789012)
//   - 1 Role (Developer)
//
// This is the most common test scenario for SSO profiles.
func NewManagedSsoSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Test organization",
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
		},
	}
}

// NewManagedSsoMultiAccount creates a schema with multiple AWS accounts.
//
// Contains:
//   - 1 Organization (test-org)
//   - 1 Partition (commercial)
//   - 3 Accounts (dev, staging, prod)
//   - 2 Roles (Developer, Admin)
//
// Useful for testing account filtering and role assignment.
func NewManagedSsoMultiAccount() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Test organization with multiple accounts",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2", "us-east-1"},
							Accounts: []schema.Account{
								{
									Alias: "test-dev",
									Name:  "Development",
									ID:    "111111111111",
								},
								{
									Alias: "test-staging",
									Name:  "Staging",
									ID:    "222222222222",
								},
								{
									Alias: "test-prod",
									Name:  "Production",
									ID:    "333333333333",
								},
							},
							Roles: []string{"Developer", "Admin"},
						},
					},
				},
			},
		},
	}
}

// NewManagedSsoMultiOrg creates a schema with multiple organizations.
//
// Contains:
//   - 2 Organizations (test-org, prod-org)
//   - Multiple partitions (commercial, govcloud)
//   - Multiple accounts per organization
//   - Different roles per organization
//
// Useful for testing organization filtering and complex scenarios.
func NewManagedSsoMultiOrg() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Test organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2", "us-east-1"},
							Accounts: []schema.Account{
								{
									Alias: "test-dev",
									Name:  "Test Development",
									ID:    "111111111111",
								},
								{
									Alias: "test-prod",
									Name:  "Test Production",
									ID:    "222222222222",
								},
							},
							Roles: []string{"Developer", "Admin"},
						},
					},
				},
				"prod-org": {
					Name:        "Production Org",
					Description: "Production organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://prod.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-east-1"},
							Accounts: []schema.Account{
								{
									Alias: "prod-main",
									Name:  "Production Main",
									ID:    "333333333333",
								},
							},
							Roles: []string{"ReadOnly"},
						},
						"govcloud": {
							URL:           "https://prod.awsapps.com/start",
							DefaultRegion: "us-gov-west-1",
							Regions:       []string{"us-gov-west-1"},
							Accounts: []schema.Account{
								{
									Alias: "prod-gov",
									Name:  "Production GovCloud",
									ID:    "444444444444",
								},
							},
							Roles: []string{"Admin"},
						},
					},
				},
			},
		},
	}
}

// NewManagedSsoComplex creates a complex schema with many SSO profiles.
//
// Contains:
//   - 3 Organizations
//   - Mixed commercial and govcloud partitions
//   - Multiple accounts per partition
//   - Various roles
//   - Multiple regions
//
// Useful for stress testing and complex filtering scenarios.
func NewManagedSsoComplex() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"corp": {
					Name:        "Corporate",
					Description: "Corporate accounts",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://corp.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2", "us-east-1", "eu-west-1"},
							Accounts: []schema.Account{
								{Alias: "corp-dev", Name: "Corp Development", ID: "100000000001"},
								{Alias: "corp-staging", Name: "Corp Staging", ID: "100000000002"},
								{Alias: "corp-prod", Name: "Corp Production", ID: "100000000003"},
								{Alias: "corp-shared", Name: "Corp Shared Services", ID: "100000000004"},
							},
							Roles: []string{"Developer", "Admin", "ReadOnly", "PowerUser"},
						},
					},
				},
				"security": {
					Name:        "Security",
					Description: "Security and compliance accounts",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://security.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-east-1"},
							Accounts: []schema.Account{
								{Alias: "sec-audit", Name: "Security Audit", ID: "200000000001"},
								{Alias: "sec-logs", Name: "Security Logs", ID: "200000000002"},
							},
							Roles: []string{"Auditor", "SecurityAdmin"},
						},
						"govcloud": {
							URL:           "https://security.awsapps.com/start",
							DefaultRegion: "us-gov-east-1",
							Regions:       []string{"us-gov-east-1", "us-gov-west-1"},
							Accounts: []schema.Account{
								{Alias: "sec-gov-audit", Name: "Security GovCloud Audit", ID: "200000000003"},
							},
							Roles: []string{"Auditor"},
						},
					},
				},
				"sandbox": {
					Name:        "Sandbox",
					Description: "Development sandbox accounts",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://sandbox.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "sandbox-1", Name: "Sandbox 1", ID: "300000000001"},
								{Alias: "sandbox-2", Name: "Sandbox 2", ID: "300000000002"},
								{Alias: "sandbox-3", Name: "Sandbox 3", ID: "300000000003"},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// IAM PROFILES
// =============================================================================

// NewManagedIamSingle creates a schema with a single IAM user.
//
// Contains:
//   - 1 IAM user (test-iam-user)
//
// Useful for testing IAM profile generation.
func NewManagedIamSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "test-iam-user",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
		},
	}
}

// NewManagedIamMulti creates a schema with multiple IAM users.
//
// Contains:
//   - 3 IAM users with different configurations
//
// Useful for testing IAM profile filtering and generation.
func NewManagedIamMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "test-iam-user",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
				{
					ProfileName:    "ci-cd-user",
					Region:         "us-west-2",
					AwsAccessKeyID: "AKIAIOSFODNN8EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKE2",
				},
				{
					ProfileName:    "backup-user",
					Region:         "us-east-2",
					AwsAccessKeyID: "AKIAIOSFODNN9EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKE3",
				},
			},
		},
	}
}

// =============================================================================
// ASSUMEROLE PROFILES
// =============================================================================

// NewManagedAssumeRoleSingle creates a schema with a single AssumeRole chain.
//
// Contains:
//   - 1 AssumeRole profile (assume-admin)
//
// Useful for testing AssumeRole profile generation.
func NewManagedAssumeRoleSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "assume-admin",
					RoleArn:       "arn:aws:iam::123456789012:role/AdminRole",
					SourceProfile: "test-iam-user",
					Region:        "us-east-1",
				},
			},
		},
	}
}

// NewManagedAssumeRoleMulti creates a schema with multiple AssumeRole chains.
//
// Contains:
//   - 3 AssumeRole profiles with different configurations
//
// Useful for testing AssumeRole profile filtering and chaining.
func NewManagedAssumeRoleMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "assume-admin",
					RoleArn:       "arn:aws:iam::123456789012:role/AdminRole",
					SourceProfile: "test-iam-user",
					Region:        "us-east-1",
				},
				{
					ProfileName:   "assume-developer",
					RoleArn:       "arn:aws:iam::123456789012:role/DeveloperRole",
					SourceProfile: "test-iam-user",
					Region:        "us-west-2",
				},
				{
					ProfileName:   "assume-readonly",
					RoleArn:       "arn:aws:iam::123456789012:role/ReadOnlyRole",
					SourceProfile: "test-iam-user",
					Region:        "us-east-1",
					ExternalID:    "external-id-123",
				},
			},
		},
	}
}

// =============================================================================
// GENERIC PROFILES
// =============================================================================

// NewManagedGenericSingle creates a schema with a single generic profile.
//
// Contains:
//   - 1 Generic profile with common properties
//
// Useful for testing generic profile generation.
func NewManagedGenericSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "custom-profile",
					Properties: map[string]string{
						"region": "us-east-1",
						"output": "json",
					},
				},
			},
		},
	}
}

// NewManagedGenericMulti creates a schema with multiple generic profiles.
//
// Contains:
//   - 3 Generic profiles with various properties
//
// Useful for testing generic profile filtering and generation.
func NewManagedGenericMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "custom-profile",
					Properties: map[string]string{
						"region":              "us-east-1",
						"output":              "json",
						"cli_pager":           "",
						"cli_auto_prompt":     "on",
						"duration_seconds":    "3600",
						"s3.max_bandwidth":    "50MB/s",
						"s3.use_accelerate":   "true",
						"cloudformation.role": "arn:aws:iam::123456789012:role/CFRole",
					},
				},
				{
					ProfileName: "minimal-profile",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "text",
					},
				},
				{
					ProfileName: "legacy-profile",
					Properties: map[string]string{
						"region":         "eu-west-1",
						"output":         "yaml",
						"cli_timestamp":  "iso8601",
						"retry_attempts": "5",
					},
				},
			},
		},
	}
}

// =============================================================================
// COMBINED PROFILE TYPES
// =============================================================================

// NewManagedAll creates a schema with all managed profile types.
//
// Contains:
//   - SSO profiles (via Organization)
//   - IAM user profiles
//   - AssumeRole chains
//   - Generic profiles
//
// This is the most comprehensive managed schema for integration testing.
func NewManagedAll() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Test organization",
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
					ProfileName:    "test-iam",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "test-assume",
					RoleArn:       "arn:aws:iam::123456789012:role/TestRole",
					SourceProfile: "test-iam",
					Region:        "us-east-1",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "test-generic",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "json",
					},
				},
			},
		},
	}
}

// NewManagedSsoIam creates a schema with SSO and IAM profiles.
//
// Contains:
//   - SSO profiles
//   - IAM user profiles
//
// Useful for testing SSO + IAM scenarios.
func NewManagedSsoIam() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test-org": {
					Name:        "Test Org",
					Description: "Test organization",
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
					ProfileName:    "test-iam-user",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
		},
	}
}

// =============================================================================
// SPECIAL CASES
// =============================================================================

// NewManagedEmpty creates a schema with an empty managed section.
//
// Returns a schema with an empty ProfileCollection in the managed section.
// Useful for testing validation and error handling.
func NewManagedEmpty() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{},
	}
}

// =============================================================================
// PRESETS
// =============================================================================

// NewManagedSsoWithPresets creates a schema with SSO profiles and installation presets.
//
// Contains:
//   - 2 Organizations (org1, test-org)
//   - Multiple partitions, accounts, and roles
//   - 4 Presets: developer, devsecops, devsecops-admin, break-glass
//
// This schema demonstrates how presets can filter profiles for common
// installation scenarios. Useful for testing preset functionality and GUI.
func NewManagedSsoWithPresets() *schema.Schema {
	base := NewManagedSsoMultiOrg() // Start with multi-org schema

	// Add more roles for comprehensive preset testing
	if org, exists := base.Managed.Organizations["test-org"]; exists {
		if partition, exists := org.Partitions["commercial"]; exists {
			partition.Roles = append(partition.Roles, "SystemAdmin", "SystemsDeployment", "BreakGlass")
			org.Partitions["commercial"] = partition
		}
	}

	// Add presets
	base.Presets = map[string]*schema.Preset{
		"developer": {
			Label:         "Developer",
			Description:   "Standard developer access to test organization",
			Organizations: []string{"test-org"},
			Roles:         []string{"Developer"},
		},
		"devsecops": {
			Label:         "DevSecOps",
			Description:   "DevSecOps access with deployment permissions to test organization",
			Organizations: []string{"test-org"},
			Roles:         []string{"Developer", "SystemAdmin", "SystemsDeployment"},
		},
		"devsecops-admin": {
			Label:       "DevSecOps Admin",
			Description: "Full DevSecOps access across all organizations",
			Roles:       []string{"Developer", "SystemAdmin", "SystemsDeployment"},
		},
		"break-glass": {
			Label:       "Break Glass",
			Description: "Emergency access for incident response across all organizations",
			Roles:       []string{"BreakGlass"},
		},
	}

	return base
}

// NewManagedSsoWithPresetsComplex creates a complex schema with comprehensive presets.
//
// Contains:
//   - Multiple organizations with different partitions
//   - Presets that filter by organization, partition, account, role, and region
//   - AllRegions preset example
//
// Useful for testing advanced preset filtering scenarios.
func NewManagedSsoWithPresetsComplex() *schema.Schema {
	base := NewManagedSsoComplex() // Start with complex multi-org schema

	base.Presets = map[string]*schema.Preset{
		"dev-all-regions": {
			Label:       "Developer (All Regions)",
			Description: "Developer access across all regions",
			Roles:       []string{"Developer"},
			AllRegions:  true,
		},
		"prod-only": {
			Label:       "Production Only",
			Description: "Access to production accounts only",
			Accounts:    []string{"corp-prod"},
			Roles:       []string{"Admin"},
		},
		"commercial-only": {
			Label:       "Commercial Partition Only",
			Description: "Access to commercial partition accounts",
			Partitions:  []string{"commercial"},
		},
		"specific-regions": {
			Label:       "US East Only",
			Description: "Access limited to us-east-1 region",
			Regions:     []string{"us-east-1"},
			Roles:       []string{"Developer"},
		},
	}

	return base
}
