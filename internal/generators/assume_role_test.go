package generators

import (
	"aws-profile-manager/internal/task"
	"context"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
)

func TestGenerateAssumeRoleProfiles(t *testing.T) {
	tests := []struct {
		name         string
		profiles     *schema.ProfileCollection
		wantProfiles int
		checkContent []string
	}{
		{
			name: "Single assume role chain",
			profiles: &schema.ProfileCollection{
				AssumeRoleChains: []*schema.AssumeRoleChain{
					{
						ProfileName:   "assume-admin",
						SourceProfile: "my-iam",
						RoleArn:       "arn:aws:iam::123456789012:role/AdminRole",
						Region:        "us-east-1",
					},
				},
			},
			wantProfiles: 1,
			checkContent: []string{
				"[profile assume-admin]",
				"source_profile = my-iam",
				"role_arn = arn:aws:iam::123456789012:role/AdminRole",
				"region = us-east-1",
			},
		},
		{
			name: "With MFA serial",
			profiles: &schema.ProfileCollection{
				AssumeRoleChains: []*schema.AssumeRoleChain{
					{
						ProfileName:   "assume-with-mfa",
						SourceProfile: "my-iam",
						RoleArn:       "arn:aws:iam::123456789012:role/AdminRole",
						MfaSerial:     "arn:aws:iam::999999999999:mfa/user",
						Region:        "us-east-1",
					},
				},
			},
			wantProfiles: 1,
			checkContent: []string{
				"[profile assume-with-mfa]",
				"mfa_serial = arn:aws:iam::999999999999:mfa/user",
			},
		},
		{
			name:         "Empty chains",
			profiles:     &schema.ProfileCollection{},
			wantProfiles: 0,
			checkContent: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, stats, _ := GenerateAssumeRoleProfiles(context.Background(), tt.profiles, task.NoOpReporter{})

			if stats.ProfilesWritten != tt.wantProfiles {
				t.Errorf("GenerateAssumeRoleProfiles() profiles = %v, want %v", stats.ProfilesWritten, tt.wantProfiles)
			}

			for _, check := range tt.checkContent {
				if !strings.Contains(content, check) {
					t.Errorf("GenerateAssumeRoleProfiles() content missing: %q", check)
				}
			}
		})
	}
}
