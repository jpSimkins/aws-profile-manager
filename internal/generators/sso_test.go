package generators

import (
	"aws-profile-manager/internal/task"
	"context"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
)

func TestGenerateSsoProfiles(t *testing.T) {
	tests := []struct {
		name         string
		profiles     *schema.ProfileCollection
		wantProfiles int
		wantSessions int
		checkContent []string // Strings that should appear in output
	}{
		{
			name: "Single organization, single account, single role",
			profiles: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"test-org": {
						Name: "Test Organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://test.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{Alias: "dev", Name: "Development", ID: "123456789012"},
								},
								Roles: []string{"Developer"},
							},
						},
					},
				},
			},
			wantProfiles: 1,
			wantSessions: 1,
			checkContent: []string{
				"[sso-session test-org-commercial]",
				"sso_start_url = https://test.awsapps.com/start",
				"sso_region = us-east-1",
				"[profile commercial-dev-Developer]",
				"sso_session = test-org-commercial",
				"sso_account_id = 123456789012",
				"sso_role_name = Developer",
				"region = us-east-1",
			},
		},
		{
			name: "Multiple regions",
			profiles: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"test-org": {
						Name: "Test Organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://test.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1", "us-west-2"},
								Accounts: []schema.Account{
									{Alias: "prod", Name: "Production", ID: "123456789012"},
								},
								Roles: []string{"Administrator"},
							},
						},
					},
				},
			},
			wantProfiles: 2, // One for each region
			wantSessions: 1,
			checkContent: []string{
				"[profile commercial-prod-Administrator]",
				"[profile commercial-prod-Administrator--us-west-2]",
			},
		},
		{
			name: "Multiple accounts and roles",
			profiles: &schema.ProfileCollection{
				Organizations: map[string]*schema.Organization{
					"test-org": {
						Name: "Test Organization",
						Partitions: map[string]schema.Partition{
							"commercial": {
								URL:           "https://test.awsapps.com/start",
								DefaultRegion: "us-east-1",
								Regions:       []string{"us-east-1"},
								Accounts: []schema.Account{
									{Alias: "dev", Name: "Development", ID: "111111111111"},
									{Alias: "prod", Name: "Production", ID: "222222222222"},
								},
								Roles: []string{"Developer", "Administrator"},
							},
						},
					},
				},
			},
			wantProfiles: 4, // 2 accounts * 2 roles
			wantSessions: 1,
			checkContent: []string{
				"[profile commercial-dev-Developer]",
				"[profile commercial-dev-Administrator]",
				"[profile commercial-prod-Developer]",
				"[profile commercial-prod-Administrator]",
			},
		},
		{
			name:         "Empty profile collection",
			profiles:     &schema.ProfileCollection{},
			wantProfiles: 0,
			wantSessions: 0,
			checkContent: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, stats, _ := GenerateSsoProfiles(context.Background(), tt.profiles, task.NoOpReporter{})

			if stats.ProfilesWritten != tt.wantProfiles {
				t.Errorf("GenerateSsoProfiles() profiles = %v, want %v", stats.ProfilesWritten, tt.wantProfiles)
			}

			if stats.SessionsWritten != tt.wantSessions {
				t.Errorf("GenerateSsoProfiles() sessions = %v, want %v", stats.SessionsWritten, tt.wantSessions)
			}

			for _, check := range tt.checkContent {
				if !strings.Contains(content, check) {
					t.Errorf("GenerateSsoProfiles() content missing: %q", check)
				}
			}
		})
	}
}
