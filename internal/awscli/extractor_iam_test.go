package awscli

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestExtractIamFields_AccessKeys(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with IAM access keys
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile iam-user]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
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

	// Verify IAM profile detection
	if profile.Name != "iam-user" {
		t.Errorf("Expected name 'iam-user', got '%s'", profile.Name)
	}
	if profile.Type != ProfileTypeIAM {
		t.Errorf("Expected type 'iam', got '%s'", profile.Type)
	}
	if !profile.HasAccessKey {
		t.Error("Expected HasAccessKey to be true")
	}
	if !profile.HasSecretKey {
		t.Error("Expected HasSecretKey to be true")
	}
	if profile.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", profile.Region)
	}
}

func TestExtractIamFields_CredentialProcess(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with credential_process
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile credential-proc]
credential_process = /usr/local/bin/aws-vault export my-profile
region = us-west-2
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

	// Verify credential_process detection
	if profile.Type != ProfileTypeIAM {
		t.Errorf("Expected type 'iam', got '%s'", profile.Type)
	}
	if !profile.HasCredentialProc {
		t.Error("Expected HasCredentialProc to be true")
	}
	if profile.CredentialProcess != "/usr/local/bin/aws-vault export my-profile" {
		t.Errorf("Expected credential process path, got '%s'", profile.CredentialProcess)
	}
}

func TestExtractIamFields_AssumeRole(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with assume role
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile base-profile]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-1

[profile assume-role-profile]
role_arn = arn:aws:iam::123456789012:role/MyRole
source_profile = base-profile
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

	if len(data.Profiles) != 2 {
		t.Fatalf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	// Verify base profile is IAM
	baseProfile := data.Profiles[0]
	if baseProfile.Type != ProfileTypeIAM {
		t.Errorf("Base profile should be IAM type, got '%s'", baseProfile.Type)
	}

	// Verify assume role profile
	assumeProfile := data.Profiles[1]
	if assumeProfile.Name != "assume-role-profile" {
		t.Errorf("Expected name 'assume-role-profile', got '%s'", assumeProfile.Name)
	}
	if assumeProfile.Type != ProfileTypeAssumeRole {
		t.Errorf("Expected type 'assume_role', got '%s'", assumeProfile.Type)
	}
}

func TestExtractIamFields_MixedProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with mixed profile types
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile sso-profile]
sso_session = my-session
sso_account_id = 123456789012
sso_role_name = Admin
region = us-east-1

[profile iam-profile]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-west-2

[profile credential-proc-profile]
credential_process = /usr/local/bin/credential-helper
region = eu-west-1

[sso-session my-session]
sso_start_url = https://my-org.awsapps.com/start
sso_region = us-east-1
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

	if len(data.Profiles) != 3 {
		t.Fatalf("Expected 3 profiles, got %d", len(data.Profiles))
	}

	// Verify profile types
	expectedTypes := map[string]ProfileType{
		"sso-profile":             ProfileTypeSSO,
		"iam-profile":             ProfileTypeIAM,
		"credential-proc-profile": ProfileTypeIAM,
	}

	for _, profile := range data.Profiles {
		expectedType, exists := expectedTypes[profile.Name]
		if !exists {
			t.Errorf("Unexpected profile name: %s", profile.Name)
			continue
		}
		if profile.Type != expectedType {
			t.Errorf("Profile '%s' expected type '%s', got '%s'",
				profile.Name, expectedType, profile.Type)
		}
	}
}

func TestExtractIamFields_PartialAccessKey(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with only access key (no secret key)
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile partial-iam]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
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

	// Verify partial IAM profile
	if profile.Type != ProfileTypeIAM {
		t.Errorf("Expected type 'iam', got '%s'", profile.Type)
	}
	if !profile.HasAccessKey {
		t.Error("Expected HasAccessKey to be true")
	}
	if profile.HasSecretKey {
		t.Error("Expected HasSecretKey to be false")
	}
}

func TestExtractIamFields_DefaultProfile(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with default IAM profile
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
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

	// Verify default IAM profile
	if profile.Name != "default" {
		t.Errorf("Expected name 'default', got '%s'", profile.Name)
	}
	if profile.Type != ProfileTypeIAM {
		t.Errorf("Expected type 'iam', got '%s'", profile.Type)
	}
}

func TestExtractIamFields_AssumeRoleWithWebIdentity(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config with web identity token
	configPath := test.GetTestAwsConfigPath(t)
	configContent := `[profile web-identity]
role_arn = arn:aws:iam::123456789012:role/MyRole
web_identity_token_file = /path/to/token
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

	// Verify assume role profile
	if profile.Type != ProfileTypeAssumeRole {
		t.Errorf("Expected type 'assume_role', got '%s'", profile.Type)
	}
}
