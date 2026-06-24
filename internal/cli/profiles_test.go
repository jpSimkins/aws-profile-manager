package cli

import (
	"os"
	"testing"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/test"
)

func TestRunProfiles_Empty(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create empty AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `# Empty config`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor with test config
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createProfilesCommand()

	// Run profiles command - should handle empty config gracefully
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor failed: %v", err)
	}
}

func TestRunProfiles_WithSsoProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config with SSO profiles
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1

[profile test-dev-Developer]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile test-prod-PowerUser]
sso_session = test-commercial
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createProfilesCommand()

	// Run profiles command
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor failed: %v", err)
	}
}

func TestRunProfiles_WithAccountFilter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config with multiple profiles
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1

[profile test-dev-Developer]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile test-prod-PowerUser]
sso_session = test-commercial
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with account filter
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("account-id", "123456789012")

	// Run profiles command - should only show dev profile
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with account filter failed: %v", err)
	}
}

func TestRunProfiles_WithRoleFilter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1

[profile test-dev-Developer]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile test-prod-PowerUser]
sso_session = test-commercial
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with role filter
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("role", "Developer")

	// Run profiles command - should only show Developer profile
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with role filter failed: %v", err)
	}
}

func TestRunProfiles_WithRegionFilter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile us-east]
region = us-east-1

[profile us-west]
region = us-west-2

[profile eu-west]
region = eu-west-1
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with region filter
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("region", "us-east-1")

	// Run profiles command - should only show us-east profile
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with region filter failed: %v", err)
	}
}

func TestRunProfiles_WithPatternFilter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile prod-app1]
region = us-east-1

[profile prod-app2]
region = us-west-2

[profile dev-app1]
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with pattern filter
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("pattern", "prod.*")

	// Run profiles command - should only show prod profiles
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with pattern filter failed: %v", err)
	}
}

func TestRunProfiles_WithSessionFilter(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config with multiple sessions
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session session1]
sso_start_url = https://session1.awsapps.com/start
sso_region = us-east-1

[sso-session session2]
sso_start_url = https://session2.awsapps.com/start
sso_region = us-west-2

[profile profile1]
sso_session = session1
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile profile2]
sso_session = session2
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with session filter
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("session", "session1")

	// Run profiles command - should only show profile1
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with session filter failed: %v", err)
	}
}

func TestRunProfiles_VerboseMode(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile test]
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with verbose flag
	cmd := createProfilesCommand()
	_ = cmd.Flags().Set("verbose", "true")

	// Run profiles command in verbose mode
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor in verbose mode failed: %v", err)
	}
}

func TestRunProfiles_WithIamProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config with IAM profiles
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile iam-user]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
region = us-east-1

[profile credential-proc]
credential_process = /usr/local/bin/get-creds
region = us-west-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createProfilesCommand()

	// Run profiles command - should show IAM profiles
	err := runProfilesWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runProfilesWithExtractor with IAM profiles failed: %v", err)
	}
}
