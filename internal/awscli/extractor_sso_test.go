package awscli

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestExtractSsoFields_AllFields(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with all SSO fields
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile sso-complete]
sso_session = my-session
sso_account_id = 123456789012
sso_role_name = Administrator
sso_start_url = https://my-org.awsapps.com/start
region = us-east-1

[sso-session my-session]
sso_start_url = https://my-org.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(data.Profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(data.Profiles))
	}

	profile := data.Profiles[0]

	// Verify all SSO fields were extracted
	if profile.Name != "sso-complete" {
		t.Errorf("Expected name 'sso-complete', got '%s'", profile.Name)
	}
	if profile.Type != ProfileTypeSSO {
		t.Errorf("Expected type 'sso', got '%s'", profile.Type)
	}
	if profile.SsoSession != "my-session" {
		t.Errorf("Expected sso_session 'my-session', got '%s'", profile.SsoSession)
	}
	if profile.AccountID != "123456789012" {
		t.Errorf("Expected account_id '123456789012', got '%s'", profile.AccountID)
	}
	if profile.RoleName != "Administrator" {
		t.Errorf("Expected role 'Administrator', got '%s'", profile.RoleName)
	}
	if profile.SsoStartURL != "https://my-org.awsapps.com/start" {
		t.Errorf("Expected start URL, got '%s'", profile.SsoStartURL)
	}
	if profile.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", profile.Region)
	}
}

func TestExtractSsoFields_MinimalSso(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with minimal SSO fields
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile sso-minimal]
sso_session = minimal-session
region = us-west-2

[sso-session minimal-session]
sso_start_url = https://minimal.awsapps.com/start
sso_region = us-west-2
`

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(data.Profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(data.Profiles))
	}

	profile := data.Profiles[0]

	// Verify minimal SSO fields
	if profile.Type != ProfileTypeSSO {
		t.Errorf("Expected type 'sso', got '%s'", profile.Type)
	}
	if profile.SsoSession != "minimal-session" {
		t.Errorf("Expected sso_session 'minimal-session', got '%s'", profile.SsoSession)
	}
	// Account ID and role should be empty for minimal config
	if profile.AccountID != "" {
		t.Errorf("Expected empty account_id, got '%s'", profile.AccountID)
	}
	if profile.RoleName != "" {
		t.Errorf("Expected empty role, got '%s'", profile.RoleName)
	}
}

func TestExtractSsoFields_LegacySsoFormat(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with legacy SSO format (no sso-session, inline fields)
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile legacy-sso]
sso_start_url = https://legacy.awsapps.com/start
sso_region = us-east-1
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-east-1
`

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(data.Profiles) != 1 {
		t.Fatalf("Expected 1 profile, got %d", len(data.Profiles))
	}

	profile := data.Profiles[0]

	// Verify legacy SSO format is detected
	if profile.Type != ProfileTypeSSO {
		t.Errorf("Expected type 'sso', got '%s'", profile.Type)
	}
	if profile.SsoStartURL != "https://legacy.awsapps.com/start" {
		t.Errorf("Expected legacy start URL, got '%s'", profile.SsoStartURL)
	}
	if profile.AccountID != "987654321098" {
		t.Errorf("Expected account_id '987654321098', got '%s'", profile.AccountID)
	}
}

func TestExtractSsoFields_MultipleSsoProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with multiple SSO profiles
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile prod-sso]
sso_session = prod-session
sso_account_id = 111111111111
sso_role_name = Admin
region = us-east-1

[profile dev-sso]
sso_session = dev-session
sso_account_id = 222222222222
sso_role_name = Developer
region = us-west-2

[sso-session prod-session]
sso_start_url = https://prod.awsapps.com/start
sso_region = us-east-1

[sso-session dev-session]
sso_start_url = https://dev.awsapps.com/start
sso_region = us-west-2
`

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(data.Profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	// Verify both profiles are SSO type
	for _, profile := range data.Profiles {
		if profile.Type != ProfileTypeSSO {
			t.Errorf("Profile '%s' should be SSO type, got '%s'", profile.Name, profile.Type)
		}
	}

	// Verify prod profile
	prodProfile := data.Profiles[0]
	if prodProfile.AccountID != "111111111111" {
		t.Errorf("Prod profile account mismatch")
	}

	// Verify dev profile
	devProfile := data.Profiles[1]
	if devProfile.AccountID != "222222222222" {
		t.Errorf("Dev profile account mismatch")
	}
}

func TestExtractSsoFields_SsoSession(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with SSO session
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[sso-session test-session]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	extractor := NewExtractorWithPath(configPath)
	data, err := extractor.ExtractFromFile()
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if len(data.SsoSessions) != 1 {
		t.Fatalf("Expected 1 SSO session, got %d", len(data.SsoSessions))
	}

	session := data.SsoSessions[0]
	if session.Name != "test-session" {
		t.Errorf("Expected session name 'test-session', got '%s'", session.Name)
	}
	if session.StartURL != "https://test.awsapps.com/start" {
		t.Errorf("Expected start URL, got '%s'", session.StartURL)
	}
	if session.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", session.Region)
	}
}
