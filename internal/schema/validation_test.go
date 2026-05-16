package schema

import (
	"testing"
)

func TestSchema_Validate(t *testing.T) {
	tests := []struct {
		name    string
		schema  *Schema
		wantErr bool
	}{
		{
			name: "Valid schema with managed section",
			schema: &Schema{
				Version: CurrentSchemaVersion,
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Organization",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-east-1",
									Regions:       []string{"us-east-1"},
									Accounts: []Account{
										{Alias: "dev", Name: "Development", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid schema with unmanaged section",
			schema: &Schema{
				Version: CurrentSchemaVersion,
				Unmanaged: &UnmanagedProfiles{
					Above: &ProfileCollection{
						IamUsers: []*IamUser{
							{ProfileName: "personal-iam", Region: "us-west-2"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid - missing version",
			schema: &Schema{
				Version: "",
				Managed: &ProfileCollection{},
			},
			wantErr: true,
		},
		{
			name: "Invalid - no sections",
			schema: &Schema{
				Version: CurrentSchemaVersion,
			},
			wantErr: true,
		},
		{
			name: "Invalid - bad account ID",
			schema: &Schema{
				Version: CurrentSchemaVersion,
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Organization",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-east-1",
									Regions:       []string{"us-east-1"},
									Accounts: []Account{
										{Alias: "dev", Name: "Development", ID: "invalid"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Schema.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileCollection_Validate(t *testing.T) {
	tests := []struct {
		name       string
		collection *ProfileCollection
		wantErr    bool
	}{
		{
			name: "Valid SSO organization",
			collection: &ProfileCollection{
				Organizations: map[string]*Organization{
					"test-org": {
						Name: "Test Organization",
						Partitions: map[string]Partition{
							"commercial": {
								URL:           "https://test.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []Account{
									{Alias: "dev", Name: "Development", ID: "123456789012"},
								},
								Roles: []string{"Developer"},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid IAM user",
			collection: &ProfileCollection{
				IamUsers: []*IamUser{
					{ProfileName: "my-iam", Region: "us-west-2", CredentialProcess: "aws-vault exec my-user"},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid AssumeRole chain",
			collection: &ProfileCollection{
				AssumeRoleChains: []*AssumeRoleChain{
					{
						ProfileName:   "assume-admin",
						SourceProfile: "my-iam",
						RoleArn:       "arn:aws:iam::123456789012:role/AdminRole",
						Region:        "us-east-1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid Generic profile",
			collection: &ProfileCollection{
				GenericProfiles: []*GenericProfile{
					{
						ProfileName: "custom",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid - missing SSO URL",
			collection: &ProfileCollection{
				Organizations: map[string]*Organization{
					"test-org": {
						Name: "Test Organization",
						Partitions: map[string]Partition{
							"commercial": {
								URL:           "",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []Account{
									{Alias: "dev", Name: "Development", ID: "123456789012"},
								},
								Roles: []string{"Developer"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid - missing IAM profile name",
			collection: &ProfileCollection{
				IamUsers: []*IamUser{
					{ProfileName: "", Region: "us-west-2"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.collection.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProfileCollection.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
