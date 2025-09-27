// Package test provides test schema fixtures for use across the codebase.
//
// This file contains mixed schemas - schemas with both managed and unmanaged sections.
// These are useful for testing backup/export operations that need to handle both
// tool-managed profiles and user's personal profiles.
package test

import (
	"aws-profile-manager/internal/schema"
)

// =============================================================================
// SAME PROFILE TYPES IN BOTH SECTIONS
// =============================================================================

// NewMixedSsoSso creates a schema with SSO profiles in both sections.
//
// Contains:
//   - Managed: 1 SSO organization (work)
//   - Unmanaged: 1 SSO organization (personal)
//
// Useful for testing that both work and personal SSO profiles are preserved.
func NewMixedSsoSso() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal organization",
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

// NewMixedIamIam creates a schema with IAM profiles in both sections.
//
// Contains:
//   - Managed: 1 IAM user (work)
//   - Unmanaged: 1 IAM user (personal)
//
// Useful for testing IAM profile preservation in both sections.
func NewMixedIamIam() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "work-iam",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7WORKXXXX",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYWORKXXXX1",
				},
			},
		},
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

// NewMixedAllAll creates a schema with all profile types in both sections.
//
// Contains:
//   - Managed: SSO + IAM + AssumeRole + Generic
//   - Unmanaged: SSO + IAM + AssumeRole + Generic
//
// The most comprehensive mixed schema for full integration testing.
func NewMixedAllAll() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
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
					ProfileName:    "work-iam",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7WORKXXXX",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYWORKXXXX1",
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "work-assume",
					RoleArn:       "arn:aws:iam::123456789012:role/WorkRole",
					SourceProfile: "work-iam",
					Region:        "us-east-1",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "work-generic",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "json",
					},
				},
			},
		},
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal organization",
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
						ProfileName:   "personal-assume",
						RoleArn:       "arn:aws:iam::999999999999:role/PersonalRole",
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

// =============================================================================
// DIFFERENT PROFILE TYPES IN EACH SECTION
// =============================================================================

// NewMixedSsoIam creates a schema with SSO (managed) and IAM (personal).
//
// Contains:
//   - Managed: SSO profiles
//   - Unmanaged: IAM users
//
// Useful for testing common scenario where work uses SSO, personal uses IAM.
func NewMixedSsoIam() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
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

// NewMixedAllSso creates a schema with all managed types and personal SSO.
//
// Contains:
//   - Managed: SSO + IAM + AssumeRole + Generic
//   - Unmanaged: SSO only
//
// Useful for testing when managed section has everything, personal has only SSO.
func NewMixedAllSso() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
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
					ProfileName:    "work-iam",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7WORKXXXX",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYWORKXXXX1",
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "work-assume",
					RoleArn:       "arn:aws:iam::123456789012:role/WorkRole",
					SourceProfile: "work-iam",
					Region:        "us-east-1",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "work-generic",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "json",
					},
				},
			},
		},
		Unmanaged: &schema.UnmanagedProfiles{
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal organization",
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

// NewMixedSsoGeneric creates a schema with SSO (managed) and Generic (personal).
//
// Contains:
//   - Managed: SSO profiles
//   - Unmanaged: Generic profiles
//
// Useful for testing when personal profiles are just custom configurations.
func NewMixedSsoGeneric() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
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
							"region": "us-west-2",
							"output": "yaml",
						},
					},
				},
			},
		},
	}
}

// =============================================================================
// COMMON TEST SCENARIOS
// =============================================================================

// NewMixedSimple creates a simple mixed schema for basic testing.
//
// Contains:
//   - Managed: 1 SSO organization
//   - Unmanaged: 2 Generic profiles
//
// The most common test scenario - work SSO + personal generic configs.
func NewMixedSimple() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org": {
					Name:        "Work Org",
					Description: "Work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "work-dev",
									Name:  "Work Development",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
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
							"region": "us-west-2",
							"output": "yaml",
						},
					},
				},
			},
		},
	}
}

// NewMixedComplex creates a complex mixed schema for stress testing.
//
// Contains:
//   - Managed: Multiple SSO orgs + IAM + AssumeRole + Generic
//   - Unmanaged: Multiple SSO orgs + IAM + Generic (above and below)
//
// Comprehensive schema for testing all combinations and edge cases.
func NewMixedComplex() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"work-org-1": {
					Name:        "Work Org 1",
					Description: "First work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work1.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2", "us-east-1"},
							Accounts: []schema.Account{
								{Alias: "work1-dev", Name: "Work 1 Development", ID: "111111111111"},
								{Alias: "work1-prod", Name: "Work 1 Production", ID: "111111111112"},
							},
							Roles: []string{"Developer", "Admin"},
						},
					},
				},
				"work-org-2": {
					Name:        "Work Org 2",
					Description: "Second work organization",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://work2.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-east-1"},
							Accounts: []schema.Account{
								{Alias: "work2-main", Name: "Work 2 Main", ID: "222222222222"},
							},
							Roles: []string{"ReadOnly"},
						},
					},
				},
			},
			IamUsers: []*schema.IamUser{
				{
					ProfileName:    "work-iam-1",
					Region:         "us-east-1",
					AwsAccessKeyID: "AKIAIOSFODNN7WORK1XXX",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYWORK1XXX1",
				},
				{
					ProfileName:    "work-iam-2",
					Region:         "us-west-2",
					AwsAccessKeyID: "AKIAIOSFODNN7WORK2XXX",
					AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYWORK2XXX1",
				},
			},
			AssumeRoleChains: []*schema.AssumeRoleChain{
				{
					ProfileName:   "work-assume-1",
					RoleArn:       "arn:aws:iam::111111111111:role/WorkRole1",
					SourceProfile: "work-iam-1",
					Region:        "us-east-1",
				},
				{
					ProfileName:   "work-assume-2",
					RoleArn:       "arn:aws:iam::222222222222:role/WorkRole2",
					SourceProfile: "work-iam-2",
					Region:        "us-west-2",
				},
			},
			GenericProfiles: []*schema.GenericProfile{
				{
					ProfileName: "work-generic",
					Properties: map[string]string{
						"region": "us-west-2",
						"output": "json",
					},
				},
			},
		},
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
				},
			},
			Below: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"personal-org": {
						Name:        "Personal Org",
						Description: "Personal organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://personal.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{Alias: "personal-main", Name: "Personal Main", ID: "999999999999"},
								},
								Roles: []string{"Admin"},
							},
						},
					},
					"freelance-org": {
						Name:        "Freelance Org",
						Description: "Freelance organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://freelance.awsapps.com/start",
								DefaultRegion: "us-west-2",
								Regions:       []string{"us-west-2"},
								Accounts: []schema.Account{
									{Alias: "client-acct", Name: "Client Account", ID: "888888888888"},
								},
								Roles: []string{"Consultant"},
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
					{
						ProfileName:    "freelance-iam",
						Region:         "us-west-2",
						AwsAccessKeyID: "AKIAIOSFODNN9FREELANC",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYFREELANC1",
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
					{
						ProfileName: "home-desktop",
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
