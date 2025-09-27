// Package test provides test schema fixtures for use across the codebase.
//
// This file contains unmanaged section schemas - personal profiles that users
// create and manage themselves, outside the tool-managed section.
//
// Unmanaged profiles can appear in two locations:
//   - Above: Before the managed section markers
//   - Below: After the managed section markers
//
// Most tests use "Below" as this is the more common user pattern.
package test

import (
	"aws-profile-manager/internal/schema"
)

// =============================================================================
// SSO PROFILES (Personal)
// =============================================================================

// NewUnmanagedSsoSingle creates a schema with a single personal SSO profile.
//
// Contains:
//   - 1 Personal SSO profile (below managed section)
//
// Useful for testing personal SSO profile preservation.
func NewUnmanagedSsoSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal AWS organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://personal.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{
										Alias: "personal-main",
										Name:  "Personal Main",
										ID:    "999999999999",
									},
								},
								Roles: []string{"Admin"},
							},
						},
					},
				},
			},
		},
	}
}

// NewUnmanagedSsoMulti creates a schema with multiple personal SSO profiles.
//
// Contains:
//   - 2 Organizations with multiple accounts (below managed section)
//
// Useful for testing preservation of complex personal SSO setups.
func NewUnmanagedSsoMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal AWS organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://personal.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1", "us-west-2"},
								Accounts: []schema.Account{
									{
										Alias: "personal-dev",
										Name:  "Personal Development",
										ID:    "999999999991",
									},
									{
										Alias: "personal-prod",
										Name:  "Personal Production",
										ID:    "999999999992",
									},
								},
								Roles: []string{"Admin", "Developer"},
							},
						},
					},
					"freelance-org": {
						Name:        "Freelance Org",
						Description: "Freelance client organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://freelance.awsapps.com/start",
								DefaultRegion: "us-west-2",
								Regions:       []string{"us-west-2"},
								Accounts: []schema.Account{
									{
										Alias: "client-acct",
										Name:  "Client Account",
										ID:    "888888888888",
									},
								},
								Roles: []string{"Consultant"},
							},
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// IAM PROFILES (Personal)
// =============================================================================

// NewUnmanagedIamSingle creates a schema with a single personal IAM user.
//
// Contains:
//   - 1 Personal IAM user (below managed section)
//
// Useful for testing personal IAM user preservation.
func NewUnmanagedIamSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				IamUsers: []*schema.IamUser{
					{
						ProfileName:    "personal-iam",
						Region:         "us-east-1",
						AwsAccessKeyID: "AKIAIOSFODNN9PERSONAL",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYPERSONAL1",
					},
				},
			},
		},
	}
}

// NewUnmanagedIamMulti creates a schema with multiple personal IAM users.
//
// Contains:
//   - 3 Personal IAM users (below managed section)
//
// Useful for testing preservation of multiple personal IAM users.
func NewUnmanagedIamMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				IamUsers: []*schema.IamUser{
					{
						ProfileName:    "personal-iam",
						Region:         "us-east-1",
						AwsAccessKeyID: "AKIAIOSFODNN9PERSONAL",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYPERSONAL1",
					},
					{
						ProfileName:    "freelance-iam",
						Region:         "us-west-2",
						AwsAccessKeyID: "AKIAIOSFODNN9FREELANC",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYFREELANC1",
					},
					{
						ProfileName:    "legacy-iam",
						Region:         "eu-west-1",
						AwsAccessKeyID: "AKIAIOSFODNN9LEGACYXX",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYLEGACYXX1",
					},
				},
			},
		},
	}
}

// =============================================================================
// ASSUMEROLE PROFILES (Personal)
// =============================================================================

// NewUnmanagedAssumeRoleSingle creates a schema with a single personal AssumeRole chain.
//
// Contains:
//   - 1 Personal AssumeRole profile (below managed section)
//
// Useful for testing personal role chain preservation.
func NewUnmanagedAssumeRoleSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				AssumeRoleChains: []*schema.AssumeRoleChain{
					{
						ProfileName:   "personal-admin",
						RoleArn:       "arn:aws:iam::999999999999:role/AdminRole",
						SourceProfile: "personal-iam",
						Region:        "us-east-1",
					},
				},
			},
		},
	}
}

// NewUnmanagedAssumeRoleMulti creates a schema with multiple personal AssumeRole chains.
//
// Contains:
//   - 3 Personal AssumeRole profiles (below managed section)
//
// Useful for testing preservation of multiple personal role chains.
func NewUnmanagedAssumeRoleMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				AssumeRoleChains: []*schema.AssumeRoleChain{
					{
						ProfileName:   "personal-admin",
						RoleArn:       "arn:aws:iam::999999999999:role/AdminRole",
						SourceProfile: "personal-iam",
						Region:        "us-east-1",
					},
					{
						ProfileName:   "personal-developer",
						RoleArn:       "arn:aws:iam::999999999999:role/DeveloperRole",
						SourceProfile: "personal-iam",
						Region:        "us-west-2",
					},
					{
						ProfileName:   "freelance-consultant",
						RoleArn:       "arn:aws:iam::888888888888:role/ConsultantRole",
						SourceProfile: "freelance-iam",
						Region:        "us-west-2",
						ExternalID:    "freelance-external-id",
					},
				},
			},
		},
	}
}

// =============================================================================
// GENERIC PROFILES (Personal)
// =============================================================================

// NewUnmanagedGenericSingle creates a schema with a single personal generic profile.
//
// Contains:
//   - 1 Personal generic profile (below managed section)
//
// Useful for testing personal generic profile preservation.
func NewUnmanagedGenericSingle() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "personal",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
				},
			},
		},
	}
}

// NewUnmanagedGenericMulti creates a schema with multiple personal generic profiles.
//
// Contains:
//   - 5 Personal generic profiles with various configurations (below managed section)
//
// Useful for testing preservation of multiple personal generic profiles.
func NewUnmanagedGenericMulti() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "personal",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
					{
						ProfileName: "work-laptop",
						Properties: map[string]string{
							"region":    "us-west-2",
							"output":    "yaml",
							"cli_pager": "",
						},
					},
					{
						ProfileName: "home-desktop",
						Properties: map[string]string{
							"region": "us-east-2",
							"output": "text",
						},
					},
					{
						ProfileName: "legacy-config",
						Properties: map[string]string{
							"region":         "eu-west-1",
							"output":         "json",
							"cli_timestamp":  "iso8601",
							"retry_attempts": "3",
						},
					},
					{
						ProfileName: "special-tooling",
						Properties: map[string]string{
							"region":              "ap-southeast-1",
							"output":              "json",
							"duration_seconds":    "7200",
							"s3.use_accelerate":   "true",
							"cloudformation.role": "arn:aws:iam::999999999999:role/CFRole",
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// COMBINED PROFILE TYPES
// =============================================================================

// NewUnmanagedAll creates a schema with all unmanaged profile types.
//
// Contains:
//   - Personal SSO profiles
//   - Personal IAM users
//   - Personal AssumeRole chains
//   - Personal Generic profiles
//   - (All below managed section)
//
// Useful for comprehensive testing of personal profile preservation.
func NewUnmanagedAll() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal AWS organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://personal.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{
										Alias: "personal-main",
										Name:  "Personal Main",
										ID:    "999999999999",
									},
								},
								Roles: []string{"Admin"},
							},
						},
					},
				},
				IamUsers: []*schema.IamUser{
					{
						ProfileName:    "personal-iam",
						Region:         "us-east-1",
						AwsAccessKeyID: "AKIAIOSFODNN9PERSONAL",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYPERSONAL1",
					},
				},
				AssumeRoleChains: []*schema.AssumeRoleChain{
					{
						ProfileName:   "personal-admin",
						RoleArn:       "arn:aws:iam::999999999999:role/AdminRole",
						SourceProfile: "personal-iam",
						Region:        "us-east-1",
					},
				},
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "personal",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
				},
			},
		},
	}
}

// NewUnmanagedMixed creates a schema with mixed personal profile types.
//
// Contains:
//   - Personal SSO profiles
//   - Personal IAM users
//   - Personal Generic profiles
//   - (No AssumeRole chains)
//   - (All below managed section)
//
// Useful for testing common personal profile combinations.
func NewUnmanagedMixed() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal AWS organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://personal.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{
										Alias: "personal-main",
										Name:  "Personal Main",
										ID:    "999999999999",
									},
								},
								Roles: []string{"Admin"},
							},
						},
					},
				},
				IamUsers: []*schema.IamUser{
					{
						ProfileName:    "personal-iam",
						Region:         "us-east-1",
						AwsAccessKeyID: "AKIAIOSFODNN9PERSONAL",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYPERSONAL1",
					},
				},
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "personal",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
					{
						ProfileName: "work-laptop",
						Properties: map[string]string{
							"region": "us-west-2",
							"output": "yaml",
						},
					},
				},
			},
		},
	}
}

// NewUnmanagedAboveBelow creates a schema with personal profiles in both locations.
//
// Contains:
//   - Above: 2 Generic profiles
//   - Below: 3 Generic profiles
//
// Useful for testing that both above and below sections are preserved correctly.
func NewUnmanagedAboveBelow() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Unmanaged: &schema.UnmanagedProfiles{
			Above: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "legacy-above",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
					{
						ProfileName: "tools-above",
						Properties: map[string]string{
							"region": "us-west-2",
							"output": "yaml",
						},
					},
				},
			},
			Below: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "personal-below",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
					{
						ProfileName: "work-below",
						Properties: map[string]string{
							"region": "us-west-2",
							"output": "yaml",
						},
					},
					{
						ProfileName: "home-below",
						Properties: map[string]string{
							"region": "us-east-2",
							"output": "text",
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// SPECIAL CASES
// =============================================================================

// NewUnmanagedEmpty creates a schema with an empty unmanaged section.
//
// Returns a schema with an empty UnmanagedProfiles (no Above or Below).
// Useful for testing that operations work when no personal profiles exist.
func NewUnmanagedEmpty() *schema.Schema {
	return &schema.Schema{
		Version:   "1.0",
		Unmanaged: &schema.UnmanagedProfiles{},
	}
}
