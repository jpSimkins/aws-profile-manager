package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/test"
)

func TestRunSessions_NoSessions(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create empty AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile test]
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor with test config
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createSessionsCommand()
	_ = cmd.Flags().Set("verbose", "true")

	// Run sessions command
	err := runSessionsWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runSessionsWithExtractor failed: %v", err)
	}
}

func TestRunSessions_WithSsoSessions(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config with SSO sessions
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access

[profile test-dev]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor with test config
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createSessionsCommand()

	// Run sessions command
	err := runSessionsWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runSessionsWithExtractor failed: %v", err)
	}
}

func TestRunSessions_WithRefresh(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1

[profile test-dev]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with refresh flag
	cmd := createSessionsCommand()
	_ = cmd.Flags().Set("refresh", "true")

	// Run sessions command with refresh
	err := runSessionsWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runSessionsWithExtractor failed: %v", err)
	}
}

func TestRunSessions_WithExpiredCache(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[sso-session test-commercial]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1

[profile test-dev]
sso_session = test-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create expired SSO cache file
	awsDir := test.GetTestAwsDir(t)
	ssoCacheDir := filepath.Join(awsDir, "sso", "cache")
	if err := os.MkdirAll(ssoCacheDir, 0755); err != nil {
		t.Fatalf("Failed to create SSO cache dir: %v", err)
	}

	// Create an expired cache file (1 hour ago)
	expiredTime := time.Now().Add(-1 * time.Hour)
	cacheFile := filepath.Join(ssoCacheDir, "expired.json")
	cacheContent := `{
		"startUrl": "https://test.awsapps.com/start",
		"region": "us-east-1",
		"accessToken": "expired-token",
		"expiresAt": "` + expiredTime.Format(time.RFC3339) + `"
	}`
	if err := os.WriteFile(cacheFile, []byte(cacheContent), 0644); err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command
	cmd := createSessionsCommand()

	// Run sessions command - should handle expired cache gracefully
	err := runSessionsWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runSessionsWithExtractor failed: %v", err)
	}
}

func TestRunSessions_VerboseMode(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create AWS config
	configPath := test.GetTestAwsConfigPath(t)
	content := `[profile test]
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create extractor
	extractor := awscli.NewExtractorWithPath(configPath)

	// Create command with verbose flag
	cmd := createSessionsCommand()
	_ = cmd.Flags().Set("verbose", "true")

	// Run sessions command in verbose mode
	err := runSessionsWithExtractor(cmd, []string{}, extractor)
	if err != nil {
		t.Fatalf("runSessionsWithExtractor failed: %v", err)
	}
}
