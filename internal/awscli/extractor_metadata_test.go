package awscli

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestExtractor_ParseMetadataComments verifies metadata comments are extracted correctly
func TestExtractor_ParseMetadataComments(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with metadata comments
	configContent := `# Organization: Gov Production
# Description: Production environment for Government services
[sso-session gov-prod-commercial]
sso_start_url = https://gov-prod.awsapps.com/start
sso_region = us-west-2
sso_registration_scopes = sso:account:access

# Organization: Gov Production
# Account: Gov Production - Main
[profile commercial-gov-prod-SystemAdmin]
sso_session = gov-prod-commercial
sso_account_id = 123456789012
sso_role_name = SystemAdmin
region = us-west-2

# Organization: Gov Production
# Account: Gov Production - DevOps
[profile commercial-gov-prod-devops-SystemAdmin]
sso_session = gov-prod-commercial
sso_account_id = 987654321012
sso_role_name = SystemAdmin
region = us-west-2
`

	// Write test config
	awsDir := test.GetTestAwsDir(t)
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Extract data
	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Verify SSO session metadata
	if len(data.SsoSessions) != 1 {
		t.Fatalf("Expected 1 SSO session, got %d", len(data.SsoSessions))
	}

	session := data.SsoSessions[0]
	if session.OrganizationName != "Gov Production" {
		t.Errorf("Session organization name = %q, want %q", session.OrganizationName, "Gov Production")
	}
	if session.Description != "Production environment for Government services" {
		t.Errorf("Session description = %q, want %q", session.Description, "Production environment for Government services")
	}

	// Verify profile metadata
	if len(data.Profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	// First profile
	profile1 := data.Profiles[0]
	if profile1.OrganizationName != "Gov Production" {
		t.Errorf("Profile 1 organization name = %q, want %q", profile1.OrganizationName, "Gov Production")
	}
	if profile1.AccountName != "Gov Production - Main" {
		t.Errorf("Profile 1 account name = %q, want %q", profile1.AccountName, "Gov Production - Main")
	}

	// Second profile
	profile2 := data.Profiles[1]
	if profile2.OrganizationName != "Gov Production" {
		t.Errorf("Profile 2 organization name = %q, want %q", profile2.OrganizationName, "Gov Production")
	}
	if profile2.AccountName != "Gov Production - DevOps" {
		t.Errorf("Profile 2 account name = %q, want %q", profile2.AccountName, "Gov Production - DevOps")
	}
}

// TestExtractor_MetadataWithoutComments verifies extraction works without metadata comments
func TestExtractor_MetadataWithoutComments(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config without metadata comments
	configContent := `[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Administrator
region = us-east-1
`

	// Write test config
	awsDir := test.GetTestAwsDir(t)
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Extract data
	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Verify extraction works (metadata should be empty)
	if len(data.SsoSessions) != 1 {
		t.Fatalf("Expected 1 SSO session, got %d", len(data.SsoSessions))
	}

	session := data.SsoSessions[0]
	if session.OrganizationName != "" {
		t.Errorf("Session organization name should be empty, got %q", session.OrganizationName)
	}
	if session.Description != "" {
		t.Errorf("Session description should be empty, got %q", session.Description)
	}

	// Verify profile
	if len(data.Profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(data.Profiles))
	}

	profile := data.Profiles[0]
	if profile.OrganizationName != "" {
		t.Errorf("Profile organization name should be empty, got %q", profile.OrganizationName)
	}
	if profile.AccountName != "" {
		t.Errorf("Profile account name should be empty, got %q", profile.AccountName)
	}
}

// TestExtractor_MetadataCommentsNotAppliedToWrongSection verifies metadata is only applied to next section
func TestExtractor_MetadataCommentsNotAppliedToWrongSection(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with metadata before wrong section type
	configContent := `# Organization: Should Not Apply
# Account: Should Not Apply
[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

# Organization: Correct Org
# Account: Correct Account
[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Administrator
region = us-east-1
`

	// Write test config
	awsDir := test.GetTestAwsDir(t)
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Extract data
	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Verify SSO session did NOT get Account metadata (only Organization and Description apply to sessions)
	session := data.SsoSessions[0]
	if session.OrganizationName != "Should Not Apply" {
		t.Errorf("Session got Organization from profile metadata")
	}

	// Verify profile got correct metadata
	profile := data.Profiles[0]
	if profile.OrganizationName != "Correct Org" {
		t.Errorf("Profile organization name = %q, want %q", profile.OrganizationName, "Correct Org")
	}
	if profile.AccountName != "Correct Account" {
		t.Errorf("Profile account name = %q, want %q", profile.AccountName, "Correct Account")
	}
}

// TestExtractor_MetadataWithRegularComments verifies regular comments are ignored
func TestExtractor_MetadataWithRegularComments(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with mix of metadata and regular comments
	configContent := `# This is a regular comment
# Organization: Test Org
# Another regular comment
# Description: Test Description
[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
# Comment inside session
sso_registration_scopes = sso:account:access

# Regular comment before profile
# Organization: Test Org
# Account: Test Account
[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
# Comment inside profile
sso_role_name = Administrator
region = us-east-1
`

	// Write test config
	awsDir := test.GetTestAwsDir(t)
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Extract data
	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Verify metadata was extracted correctly (regular comments ignored)
	session := data.SsoSessions[0]
	if session.OrganizationName != "Test Org" {
		t.Errorf("Session organization name = %q, want %q", session.OrganizationName, "Test Org")
	}
	if session.Description != "Test Description" {
		t.Errorf("Session description = %q, want %q", session.Description, "Test Description")
	}

	profile := data.Profiles[0]
	if profile.OrganizationName != "Test Org" {
		t.Errorf("Profile organization name = %q, want %q", profile.OrganizationName, "Test Org")
	}
	if profile.AccountName != "Test Account" {
		t.Errorf("Profile account name = %q, want %q", profile.AccountName, "Test Account")
	}
}

// TestExtractor_MetadataEdgeCases tests various edge cases
func TestExtractor_MetadataEdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		wantSessionOrg  string
		wantSessionDesc string
		wantProfileOrg  string
		wantProfileAcct string
	}{
		{
			name: "Metadata with extra spaces",
			content: `#   Organization:   Test Org   
#  Description:    Test Description  
[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

#   Organization:   Test Org   
# Account:  Test Account  
[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Administrator
region = us-east-1
`,
			wantSessionOrg:  "Test Org",
			wantSessionDesc: "Test Description",
			wantProfileOrg:  "Test Org",
			wantProfileAcct: "Test Account",
		},
		{
			name: "Metadata with colons in value",
			content: `# Organization: Test: Production: Main
# Description: For: Testing: Purposes
[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

# Organization: Test: Production
# Account: Main: Account: 001
[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Administrator
region = us-east-1
`,
			wantSessionOrg:  "Test: Production: Main",
			wantSessionDesc: "For: Testing: Purposes",
			wantProfileOrg:  "Test: Production",
			wantProfileAcct: "Main: Account: 001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SetupTestEnvironment(t)

			// Write test config
			awsDir := test.GetTestAwsDir(t)
			configPath := filepath.Join(awsDir, "config")
			if err := os.WriteFile(configPath, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Extract data
			extractor := NewExtractorWithPath(configPath)
			data, err := extractor.ExtractFromFile()
			if err != nil {
				t.Fatalf("Failed to extract: %v", err)
			}

			// Verify session metadata
			if len(data.SsoSessions) != 1 {
				t.Fatalf("Expected 1 SSO session, got %d", len(data.SsoSessions))
			}
			session := data.SsoSessions[0]
			if session.OrganizationName != tt.wantSessionOrg {
				t.Errorf("Session org = %q, want %q", session.OrganizationName, tt.wantSessionOrg)
			}
			if session.Description != tt.wantSessionDesc {
				t.Errorf("Session desc = %q, want %q", session.Description, tt.wantSessionDesc)
			}

			// Verify profile metadata
			if len(data.Profiles) != 1 {
				t.Fatalf("Expected 1 profile, got %d", len(data.Profiles))
			}
			profile := data.Profiles[0]
			if profile.OrganizationName != tt.wantProfileOrg {
				t.Errorf("Profile org = %q, want %q", profile.OrganizationName, tt.wantProfileOrg)
			}
			if profile.AccountName != tt.wantProfileAcct {
				t.Errorf("Profile account = %q, want %q", profile.AccountName, tt.wantProfileAcct)
			}
		})
	}
}

// TestExtractor_MetadataClearedBetweenSections verifies pending metadata is cleared
func TestExtractor_MetadataClearedBetweenSections(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with metadata that shouldn't carry over
	configContent := `# Organization: Org 1
# Description: Description 1
[sso-session test-org-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

[profile commercial-test-Administrator]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Administrator
region = us-east-1
`

	// Write test config
	awsDir := test.GetTestAwsDir(t)
	configPath := filepath.Join(awsDir, "config")
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Extract data
	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Profile should NOT have metadata from previous section
	profile := data.Profiles[0]
	if profile.OrganizationName != "" {
		t.Errorf("Profile should not have organization name from previous section, got %q", profile.OrganizationName)
	}
	if profile.AccountName != "" {
		t.Errorf("Profile should not have account name, got %q", profile.AccountName)
	}
}
