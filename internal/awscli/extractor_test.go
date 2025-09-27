package awscli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/test"
)

func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()

	if extractor == nil {
		t.Fatal("NewExtractor() returned nil")
	}

	// Should default to ~/.aws/config path
	expectedSuffix := filepath.Join(".aws", "config")
	if !filepath.IsAbs(extractor.configPath) {
		t.Error("Config path should be absolute")
	}

	if !contains(extractor.configPath, expectedSuffix) {
		t.Errorf("Config path should contain %s, got %s", expectedSuffix, extractor.configPath)
	}
}

func TestNewExtractorWithPath(t *testing.T) {
	testPath := "/test/path/config"
	extractor := NewExtractorWithPath(testPath)

	if extractor == nil {
		t.Fatal("NewExtractorWithPath() returned nil")
	}

	if extractor.configPath != testPath {
		t.Errorf("Expected config path %s, got %s", testPath, extractor.configPath)
	}
}

func TestNewExtractorWithEnvOverride(t *testing.T) {
	// Use the test helper to set up an isolated environment and get the test AWS dir
	test.SetupTestEnvironment(t)

	extractor := NewExtractor()

	if extractor == nil {
		t.Fatal("NewExtractor() returned nil")
	}

	expectedPath := filepath.Join(test.GetTestAwsDir(t), "config")
	if extractor.configPath != expectedPath {
		t.Errorf("Expected config path %s, got %s", expectedPath, extractor.configPath)
	}
}

func TestGetConfigPath(t *testing.T) {
	testPath := "/test/path/config"
	extractor := NewExtractorWithPath(testPath)

	if extractor.GetConfigPath() != testPath {
		t.Errorf("Expected config path %s, got %s", testPath, extractor.GetConfigPath())
	}
}

func TestExtractFromFile_SimpleProfile(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create a simple AWS config
	configContent := `[default]
region = us-east-1
output = json

[profile test-profile]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-west-2
sso_session = my-sso
sso_start_url = https://example.awsapps.com/start
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	if data == nil {
		t.Fatal("ExtractFromFile() returned nil data")
	}

	// Verify basic structure
	if len(data.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	if len(data.SsoSessions) != 0 {
		t.Errorf("Expected 0 SSO sessions, got %d", len(data.SsoSessions))
	}

	if data.SourceFile != configFile {
		t.Errorf("Expected source file %s, got %s", configFile, data.SourceFile)
	}

	if data.ExtractedAt.IsZero() {
		t.Error("ExtractedAt timestamp should be set")
	}

	// Verify default profile
	var defaultProfile *AwsCliProfile
	for _, profile := range data.Profiles {
		if profile.Name == "default" {
			defaultProfile = &profile
			break
		}
	}

	if defaultProfile == nil {
		t.Error("Default profile not found")
	} else {
		if defaultProfile.Region != "us-east-1" {
			t.Errorf("Expected default region us-east-1, got %s", defaultProfile.Region)
		}
		if defaultProfile.Properties["output"] != "json" {
			t.Errorf("Expected output property json, got %s", defaultProfile.Properties["output"])
		}
	}

	// Verify test profile
	var testProfile *AwsCliProfile
	for _, profile := range data.Profiles {
		if profile.Name == "test-profile" {
			testProfile = &profile
			break
		}
	}

	if testProfile == nil {
		t.Error("Test profile not found")
	} else {
		if testProfile.AccountID != "123456789012" {
			t.Errorf("Expected account ID 123456789012, got %s", testProfile.AccountID)
		}
		if testProfile.RoleName != "AdminRole" {
			t.Errorf("Expected role name AdminRole, got %s", testProfile.RoleName)
		}
		if testProfile.Region != "us-west-2" {
			t.Errorf("Expected region us-west-2, got %s", testProfile.Region)
		}
		if testProfile.SsoSession != "my-sso" {
			t.Errorf("Expected SSO session my-sso, got %s", testProfile.SsoSession)
		}
		if testProfile.SsoStartURL != "https://example.awsapps.com/start" {
			t.Errorf("Expected SSO start URL https://example.awsapps.com/start, got %s", testProfile.SsoStartURL)
		}
	}
}

func TestExtractFromFile_WithSsoSessions(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create AWS config with SSO sessions
	configContent := `[profile prod-admin]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-east-1
sso_session = prod-session

[profile dev-readonly]
sso_account_id = 123456789013
sso_role_name = ReadOnlyRole
region = us-west-2
sso_session = dev-session

[sso-session prod-session]
sso_start_url = https://prod.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

[sso-session dev-session]
sso_start_url = https://dev.awsapps.com/start
sso_region = us-west-2
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	// Verify profiles
	if len(data.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	// Verify SSO sessions
	if len(data.SsoSessions) != 2 {
		t.Errorf("Expected 2 SSO sessions, got %d", len(data.SsoSessions))
	}

	// Verify prod session
	var prodSession *SsoSession
	for _, session := range data.SsoSessions {
		if session.Name == "prod-session" {
			prodSession = &session
			break
		}
	}

	if prodSession == nil {
		t.Error("Prod SSO session not found")
	} else {
		if prodSession.StartURL != "https://prod.awsapps.com/start" {
			t.Errorf("Expected prod start URL https://prod.awsapps.com/start, got %s", prodSession.StartURL)
		}
		if prodSession.Region != "us-east-1" {
			t.Errorf("Expected prod region us-east-1, got %s", prodSession.Region)
		}
		if prodSession.RegistrationScopes != "sso:account:access" {
			t.Errorf("Expected registration scopes sso:account:access, got %s", prodSession.RegistrationScopes)
		}
	}

	// Verify dev session
	var devSession *SsoSession
	for _, session := range data.SsoSessions {
		if session.Name == "dev-session" {
			devSession = &session
			break
		}
	}

	if devSession == nil {
		t.Error("Dev SSO session not found")
	} else {
		if devSession.StartURL != "https://dev.awsapps.com/start" {
			t.Errorf("Expected dev start URL https://dev.awsapps.com/start, got %s", devSession.StartURL)
		}
		if devSession.Region != "us-west-2" {
			t.Errorf("Expected dev region us-west-2, got %s", devSession.Region)
		}
	}
}

func TestExtractFromFile_EmptyFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create empty config file
	err := os.WriteFile(configFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	if len(data.Profiles) != 0 {
		t.Errorf("Expected 0 profiles from empty file, got %d", len(data.Profiles))
	}

	if len(data.SsoSessions) != 0 {
		t.Errorf("Expected 0 SSO sessions from empty file, got %d", len(data.SsoSessions))
	}
}

func TestExtractFromFile_WithComments(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create config with comments and empty lines
	configContent := `# This is a comment
# AWS CLI Configuration

[default]
# Default region
region = us-east-1

# Production profile
[profile prod]
sso_account_id = 123456789012
# Admin role for production
sso_role_name = AdminRole
region = us-east-1

# Empty lines should be ignored

[sso-session prod-session]
# SSO configuration
sso_start_url = https://example.awsapps.com/start
sso_region = us-east-1
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	if len(data.Profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	if len(data.SsoSessions) != 1 {
		t.Errorf("Expected 1 SSO session, got %d", len(data.SsoSessions))
	}
}

func TestExtractFromFile_UnknownSections(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create config with unknown sections that should be ignored
	configContent := `[default]
region = us-east-1

[unknown-section]
some_property = some_value

[profile test]
sso_account_id = 123456789012

[another-unknown]
property = value

[sso-session test-session]
sso_start_url = https://example.awsapps.com/start
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	// Should only parse known sections
	if len(data.Profiles) != 2 {
		t.Errorf("Expected 2 profiles (ignoring unknown sections), got %d", len(data.Profiles))
	}

	if len(data.SsoSessions) != 1 {
		t.Errorf("Expected 1 SSO session (ignoring unknown sections), got %d", len(data.SsoSessions))
	}
}

func TestExtractFromFile_NonExistentFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	// Use a non-existent file path
	configFile := filepath.Join(test.GetTestAwsDir(t), "nonexistent-config")

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err == nil {
		t.Error("ExtractFromFile() should fail for non-existent file")
	}

	if data != nil {
		t.Error("ExtractFromFile() should return nil data for non-existent file")
	}
}

func TestExtractFromPath(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)
	// Create a different config file in the same test AWS directory
	differentFile := filepath.Join(test.GetTestAwsDir(t), "different-config")

	// Create two different config files
	config1 := `[default]
region = us-east-1
`
	config2 := `[profile test]
region = us-west-2
`

	err := os.WriteFile(configFile, []byte(config1), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = os.WriteFile(differentFile, []byte(config2), 0644)
	if err != nil {
		t.Fatalf("Failed to create different config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)

	// Extract from different path
	data, err := extractor.ExtractFromPath(differentFile)
	if err != nil {
		t.Fatalf("ExtractFromPath() failed: %v", err)
	}

	// Should extract from the specified path, not the configured one
	if data.SourceFile != differentFile {
		t.Errorf("Expected source file %s, got %s", differentFile, data.SourceFile)
	}

	// Should find the test profile from different file
	if len(data.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(data.Profiles))
	}

	if data.Profiles[0].Name != "test" {
		t.Errorf("Expected profile name 'test', got %s", data.Profiles[0].Name)
	}
}

func TestValidateConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	testAwsDir := test.GetTestAwsDir(t)

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "Valid file",
			setup: func() string {
				configFile := filepath.Join(testAwsDir, "valid-config")
				_ = os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0644)
				return configFile
			},
			expectError: false,
		},
		{
			name: "Non-existent file",
			setup: func() string {
				return filepath.Join(testAwsDir, "nonexistent")
			},
			expectError: true,
		},
		{
			name: "Directory instead of file",
			setup: func() string {
				dirPath := filepath.Join(testAwsDir, "directory")
				_ = os.Mkdir(dirPath, 0755)
				return dirPath
			},
			expectError: true,
		},
		{
			name: "Unreadable file",
			setup: func() string {
				configFile := filepath.Join(testAwsDir, "unreadable")
				_ = os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0000)
				return configFile
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setup()
			extractor := NewExtractorWithPath(configPath)

			err := extractor.ValidateConfigFile()

			if tt.expectError && err == nil {
				t.Error("ValidateConfigFile() should have failed")
			}

			if !tt.expectError && err != nil {
				t.Errorf("ValidateConfigFile() should not have failed: %v", err)
			}
		})
	}
}

func TestGetFileModTime(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create config file
	err := os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)

	modTime, err := extractor.GetFileModTime()
	if err != nil {
		t.Fatalf("GetFileModTime() failed: %v", err)
	}

	if modTime.IsZero() {
		t.Error("GetFileModTime() should return non-zero time")
	}

	// Modification time should be recent
	if time.Since(modTime) > time.Minute {
		t.Error("GetFileModTime() should return recent modification time")
	}
}

func TestGetFileModTime_NonExistentFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	// Use a non-existent file path
	configFile := filepath.Join(test.GetTestAwsDir(t), "nonexistent")

	extractor := NewExtractorWithPath(configFile)

	_, err := extractor.GetFileModTime()
	if err == nil {
		t.Error("GetFileModTime() should fail for non-existent file")
	}
}

func TestComplexAWSConfig(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)

	// Create a complex, realistic AWS config
	configContent := `# AWS CLI Configuration for Multiple Environments
[default]
region = us-east-1
output = json
cli_pager = 

[profile gov-prod-admin]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-east-1
sso_session = gov-prod
output = json

[profile gov-prod-readonly]
sso_account_id = 123456789012
sso_role_name = ReadOnlyRole
region = us-east-1
sso_session = gov-prod

[profile dev-admin]
sso_account_id = 987654321098
sso_role_name = AdminRole
region = us-west-2
sso_session = dev

[profile legacy-access-key]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-west-1

[sso-session gov-prod]
sso_start_url = https://gov-prod.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

[sso-session dev]
sso_start_url = https://dev.awsapps.com/start
sso_region = us-west-2
sso_registration_scopes = sso:account:access
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	data, err := extractor.ExtractFromFile()

	if err != nil {
		t.Fatalf("ExtractFromFile() failed: %v", err)
	}

	// Verify profiles count
	expectedProfiles := 5 // default + 4 named profiles
	if len(data.Profiles) != expectedProfiles {
		t.Errorf("Expected %d profiles, got %d", expectedProfiles, len(data.Profiles))
	}

	// Verify SSO sessions count
	expectedSessions := 2
	if len(data.SsoSessions) != expectedSessions {
		t.Errorf("Expected %d SSO sessions, got %d", expectedSessions, len(data.SsoSessions))
	}

	// Verify specific profile details
	profileTests := map[string]struct {
		accountID  string
		roleName   string
		region     string
		ssoSession string
	}{
		"gov-prod-admin": {
			accountID:  "123456789012",
			roleName:   "AdminRole",
			region:     "us-east-1",
			ssoSession: "gov-prod",
		},
		"dev-admin": {
			accountID:  "987654321098",
			roleName:   "AdminRole",
			region:     "us-west-2",
			ssoSession: "dev",
		},
		"legacy-access-key": {
			region: "us-west-1",
		},
	}

	for profileName, expected := range profileTests {
		var profile *AwsCliProfile
		for _, p := range data.Profiles {
			if p.Name == profileName {
				profile = &p
				break
			}
		}

		if profile == nil {
			t.Errorf("Profile %s not found", profileName)
			continue
		}

		if expected.accountID != "" && profile.AccountID != expected.accountID {
			t.Errorf("Profile %s: expected account ID %s, got %s", profileName, expected.accountID, profile.AccountID)
		}

		if expected.roleName != "" && profile.RoleName != expected.roleName {
			t.Errorf("Profile %s: expected role name %s, got %s", profileName, expected.roleName, profile.RoleName)
		}

		if profile.Region != expected.region {
			t.Errorf("Profile %s: expected region %s, got %s", profileName, expected.region, profile.Region)
		}

		if expected.ssoSession != "" && profile.SsoSession != expected.ssoSession {
			t.Errorf("Profile %s: expected SSO session %s, got %s", profileName, expected.ssoSession, profile.SsoSession)
		}
	}

	// Verify properties are captured
	var legacyProfile *AwsCliProfile
	for _, p := range data.Profiles {
		if p.Name == "legacy-access-key" {
			legacyProfile = &p
			break
		}
	}

	if legacyProfile != nil {
		if legacyProfile.Properties["aws_access_key_id"] != "AKIAIOSFODNN7EXAMPLE" {
			t.Errorf("Legacy profile should capture aws_access_key_id property")
		}
		if legacyProfile.Properties["aws_secret_access_key"] != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
			t.Errorf("Legacy profile should capture aws_secret_access_key property")
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-len(substr)] == string(os.PathSeparator) && s[len(s)-len(substr):] == substr
}
