package schema

import (
	"testing"
)

func TestGenerateSessionName(t *testing.T) {
	tests := []struct {
		name          string
		orgAlias      string
		partitionName string
		want          string
	}{
		{
			name:          "Commercial partition",
			orgAlias:      "test-org",
			partitionName: "commercial",
			want:          "test-org-commercial",
		},
		{
			name:          "GovCloud partition",
			orgAlias:      "gov-org",
			partitionName: "govcloud",
			want:          "gov-org-govcloud",
		},
		{
			name:          "Hyphenated org name",
			orgAlias:      "my-company-org",
			partitionName: "commercial",
			want:          "my-company-org-commercial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSsoSessionName(tt.orgAlias, tt.partitionName)
			if got != tt.want {
				t.Errorf("GenerateSsoSessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateProfileName(t *testing.T) {
	tests := []struct {
		name          string
		partition     string
		accountAlias  string
		role          string
		region        string
		defaultRegion string
		want          string
	}{
		{
			name:          "Default region - no suffix",
			partition:     "commercial",
			accountAlias:  "prod",
			role:          "Administrator",
			region:        "us-east-1",
			defaultRegion: "us-east-1",
			want:          "commercial-prod-Administrator",
		},
		{
			name:          "Non-default region - with suffix",
			partition:     "commercial",
			accountAlias:  "prod",
			role:          "Administrator",
			region:        "us-west-2",
			defaultRegion: "us-east-1",
			want:          "commercial-prod-Administrator--us-west-2",
		},
		{
			name:          "GovCloud partition",
			partition:     "govcloud",
			accountAlias:  "gov-prod",
			role:          "Developer",
			region:        "us-gov-west-1",
			defaultRegion: "us-gov-west-1",
			want:          "govcloud-gov-prod-Developer",
		},
		{
			name:          "Empty region - no suffix",
			partition:     "commercial",
			accountAlias:  "dev",
			role:          "PowerUser",
			region:        "",
			defaultRegion: "us-east-1",
			want:          "commercial-dev-PowerUser",
		},
		{
			name:          "Different non-default region",
			partition:     "commercial",
			accountAlias:  "staging",
			role:          "ReadOnly",
			region:        "eu-west-1",
			defaultRegion: "us-east-1",
			want:          "commercial-staging-ReadOnly--eu-west-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateProfileName(tt.partition, tt.accountAlias, tt.role, tt.region, tt.defaultRegion)
			if got != tt.want {
				t.Errorf("GenerateProfileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
