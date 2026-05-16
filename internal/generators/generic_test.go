package generators

import (
	"aws-profile-manager/internal/task"
	"context"
	"strings"
	"testing"

	"aws-profile-manager/internal/schema"
)

func TestGenerateGenericProfiles(t *testing.T) {
	tests := []struct {
		name         string
		profiles     *schema.ProfileCollection
		wantProfiles int
		checkContent []string
	}{
		{
			name: "Single generic profile",
			profiles: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "custom",
						Properties: map[string]string{
							"region": "us-east-1",
							"output": "json",
						},
					},
				},
			},
			wantProfiles: 1,
			checkContent: []string{
				"[profile custom]",
				"region = us-east-1",
				"output = json",
			},
		},
		{
			name: "Multiple generic profiles",
			profiles: &schema.ProfileCollection{
				GenericProfiles: []*schema.GenericProfile{
					{
						ProfileName: "profile1",
						Properties: map[string]string{
							"region": "us-west-2",
						},
					},
					{
						ProfileName: "profile2",
						Properties: map[string]string{
							"region": "eu-west-1",
						},
					},
				},
			},
			wantProfiles: 2,
			checkContent: []string{
				"[profile profile1]",
				"[profile profile2]",
			},
		},
		{
			name:         "Empty generic profiles",
			profiles:     &schema.ProfileCollection{},
			wantProfiles: 0,
			checkContent: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, stats, _ := GenerateGenericProfiles(context.Background(), tt.profiles, task.NoOpReporter{})

			if stats.ProfilesWritten != tt.wantProfiles {
				t.Errorf("GenerateGenericProfiles() profiles = %v, want %v", stats.ProfilesWritten, tt.wantProfiles)
			}

			for _, check := range tt.checkContent {
				if !strings.Contains(content, check) {
					t.Errorf("GenerateGenericProfiles() content missing: %q", check)
				}
			}
		})
	}
}
