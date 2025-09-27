package generators

import (
	"aws-profile-manager/internal/task"
	"context"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
)

func TestGenerateIamProfiles(t *testing.T) {
	tests := []struct {
		name         string
		profiles     *schema.ProfileCollection
		wantProfiles int
		checkContent []string
	}{
		{
			name: "Single IAM user with credential process",
			profiles: &schema.ProfileCollection{
				IamUsers: []*schema.IamUser{
					{
						ProfileName:       "my-iam",
						Region:            "us-west-2",
						CredentialProcess: "aws-vault exec my-user --json",
					},
				},
			},
			wantProfiles: 1,
			checkContent: []string{
				"[profile my-iam]",
				"credential_process = aws-vault exec my-user --json",
				"region = us-west-2",
			},
		},
		{
			name: "Multiple IAM users",
			profiles: &schema.ProfileCollection{
				IamUsers: []*schema.IamUser{
					{ProfileName: "user1", Region: "us-east-1", CredentialProcess: "cmd1"},
					{ProfileName: "user2", Region: "eu-west-1", CredentialProcess: "cmd2"},
				},
			},
			wantProfiles: 2,
			checkContent: []string{
				"[profile user1]",
				"[profile user2]",
			},
		},
		{
			name:         "Empty IAM users",
			profiles:     &schema.ProfileCollection{},
			wantProfiles: 0,
			checkContent: []string{},
		},
		{
			name: "IAM user with static credentials",
			profiles: &schema.ProfileCollection{
				IamUsers: []*schema.IamUser{
					{
						ProfileName:    "static-creds",
						Region:         "us-east-1",
						AwsAccessKeyID: "AKIAIOSFODNN7EXAMPLE",
						AwsSecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					},
				},
			},
			wantProfiles: 1,
			checkContent: []string{
				"[profile static-creds]",
				"aws_access_key_id = AKIAIOSFODNN7EXAMPLE",
				"aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				"region = us-east-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, stats, _ := GenerateIamProfiles(context.Background(), tt.profiles, task.NoOpReporter{})

			if stats.ProfilesWritten != tt.wantProfiles {
				t.Errorf("GenerateIamProfiles() profiles = %v, want %v", stats.ProfilesWritten, tt.wantProfiles)
			}

			for _, check := range tt.checkContent {
				if !strings.Contains(content, check) {
					t.Errorf("GenerateIamProfiles() content missing: %q", check)
				}
			}
		})
	}
}
