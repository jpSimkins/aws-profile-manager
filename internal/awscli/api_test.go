package awscli

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// =============================================================================
// HIGH-LEVEL API TESTS
// =============================================================================

func TestListProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test AWS config
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testConfig := `[default]
region = us-east-1

[profile dev-profile]
sso_session = test-org-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile prod-profile]
sso_session = test-org-commercial
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2

[sso-session test-org-commercial]
sso_start_url = https://test-org.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	tests := []struct {
		name           string
		criteria       FilterCriteria
		expectedCount  int
		expectedNames  []string
		shouldHaveData bool
	}{
		{
			name:           "No filters - returns all profiles",
			criteria:       FilterCriteria{},
			expectedCount:  3, // default + 2 profiles
			expectedNames:  []string{"default", "dev-profile", "prod-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter by account ID",
			criteria: FilterCriteria{
				AccountIDs: []string{"123456789012"},
			},
			expectedCount:  1,
			expectedNames:  []string{"dev-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter by role name",
			criteria: FilterCriteria{
				RoleNames: []string{"PowerUser"},
			},
			expectedCount:  1,
			expectedNames:  []string{"prod-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter by region",
			criteria: FilterCriteria{
				Regions: []string{"us-west-2"},
			},
			expectedCount:  1,
			expectedNames:  []string{"prod-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter by SSO session",
			criteria: FilterCriteria{
				SsoSessions: []string{"test-org-commercial"},
			},
			expectedCount:  2, // Both SSO profiles
			expectedNames:  []string{"dev-profile", "prod-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter by name pattern",
			criteria: FilterCriteria{
				NamePattern: "dev.*",
			},
			expectedCount:  1,
			expectedNames:  []string{"dev-profile"},
			shouldHaveData: true,
		},
		{
			name: "Multiple filters - AND logic",
			criteria: FilterCriteria{
				RoleNames: []string{"Developer"},
				Regions:   []string{"us-east-1"},
			},
			expectedCount:  1,
			expectedNames:  []string{"dev-profile"},
			shouldHaveData: true,
		},
		{
			name: "Filter excludes all",
			criteria: FilterCriteria{
				AccountIDs: []string{"000000000000"}, // Non-existent account
			},
			expectedCount:  0,
			expectedNames:  []string{},
			shouldHaveData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ListProfiles(tt.criteria)
			if err != nil {
				t.Fatalf("ListProfiles() error = %v", err)
			}

			if result == nil {
				t.Fatal("Expected result, got nil")
			}

			// Verify profile count
			if len(result.Profiles) != tt.expectedCount {
				t.Errorf("Expected %d profiles, got %d", tt.expectedCount, len(result.Profiles))
			}

			// Verify profile names
			if len(tt.expectedNames) > 0 {
				profileNames := make(map[string]bool)
				for _, profile := range result.Profiles {
					profileNames[profile.Name] = true
				}

				for _, expectedName := range tt.expectedNames {
					if !profileNames[expectedName] {
						t.Errorf("Expected profile %s not found in results", expectedName)
					}
				}
			}

			// Verify result completeness
			if tt.shouldHaveData {
				// Should have SSO sessions
				if len(result.SsoSessions) == 0 {
					t.Error("Expected SSO sessions in result")
				}

				// Should have session status
				if result.SessionStatus.LastChecked.IsZero() {
					t.Error("Expected session status to be checked")
				}

				// Should have config path
				if result.ConfigPath == "" {
					t.Error("Expected config path in result")
				}
			}

			// Verify profiles are sorted by name
			for i := 1; i < len(result.Profiles); i++ {
				if result.Profiles[i-1].Name > result.Profiles[i].Name {
					t.Error("Profiles are not sorted by name")
					break
				}
			}
		})
	}
}

func TestListProfiles_NoConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Don't create config file - should error
	_, err := ListProfiles(FilterCriteria{})
	if err == nil {
		t.Error("Expected error when config file doesn't exist, got nil")
	}
}

func TestListProfilesWithExtractor(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test AWS config
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testConfig := `[profile test-profile]
sso_session = test-session
sso_account_id = 123456789012
sso_role_name = TestRole
region = us-east-1

[sso-session test-session]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Create custom extractor
	extractor := NewExtractorWithPath(configPath)

	// Test with custom extractor
	result, err := ListProfilesWithExtractor(FilterCriteria{}, extractor)
	if err != nil {
		t.Fatalf("ListProfilesWithExtractor() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if len(result.Profiles) == 0 {
		t.Error("Expected profiles in result")
	}

	if len(result.SsoSessions) == 0 {
		t.Error("Expected SSO sessions in result")
	}

	if result.ConfigPath != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, result.ConfigPath)
	}
}

func TestAPIGetSessionStatus(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test AWS config with SSO sessions
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testConfig := `[profile test-profile]
sso_session = test-session
sso_account_id = 123456789012
sso_role_name = TestRole
region = us-east-1

[sso-session test-session]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	tests := []struct {
		name         string
		forceRefresh bool
		wantError    bool
	}{
		{
			name:         "Get status without force refresh",
			forceRefresh: false,
			wantError:    false,
		},
		{
			name:         "Get status with force refresh",
			forceRefresh: true,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := GetSessionStatus(tt.forceRefresh)

			if (err != nil) != tt.wantError {
				t.Errorf("GetSessionStatus() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err == nil {
				// Verify status structure
				if status.LastChecked.IsZero() {
					t.Error("Expected LastChecked to be set")
				}

				// Should have checked CLI availability
				// Note: May be true or false depending on environment
				_ = status.CLIAvailable
			}
		})
	}
}

func TestAPIGetSessionStatus_NoConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Don't create config file
	// Note: GetSessionStatus() gracefully handles missing config by returning empty status
	status, err := GetSessionStatus(false)

	// Should not error - just returns empty status
	if err != nil {
		t.Errorf("GetSessionStatus() should handle missing config gracefully, got error: %v", err)
	}

	// Status should have empty session lists
	if len(status.ActiveSessions) > 0 {
		t.Error("Expected no active sessions for missing config")
	}

	if len(status.ExpiredSessions) > 0 {
		t.Error("Expected no expired sessions for missing config")
	}
}

func TestGetFilterOptions(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test AWS config with various profiles
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testConfig := `[profile dev-profile]
sso_session = test-session
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile prod-profile]
sso_session = test-session
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2

[profile another-dev]
sso_session = test-session
sso_account_id = 111111111111
sso_role_name = Developer
region = us-east-1

[sso-session test-session]
sso_start_url = https://test.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	result, err := GetFilterOptions()
	if err != nil {
		t.Fatalf("GetFilterOptions() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify account IDs
	if len(result.AccountIDs) != 3 {
		t.Errorf("Expected 3 account IDs, got %d", len(result.AccountIDs))
	}

	expectedAccounts := map[string]bool{
		"123456789012": true,
		"987654321098": true,
		"111111111111": true,
	}

	for _, accountID := range result.AccountIDs {
		if !expectedAccounts[accountID] {
			t.Errorf("Unexpected account ID: %s", accountID)
		}
	}

	// Verify role names (should have 2 unique roles)
	if len(result.RoleNames) != 2 {
		t.Errorf("Expected 2 role names, got %d", len(result.RoleNames))
	}

	expectedRoles := map[string]bool{
		"Developer": true,
		"PowerUser": true,
	}

	for _, roleName := range result.RoleNames {
		if !expectedRoles[roleName] {
			t.Errorf("Unexpected role name: %s", roleName)
		}
	}

	// Verify regions (should have 2 unique regions)
	if len(result.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(result.Regions))
	}

	expectedRegions := map[string]bool{
		"us-east-1": true,
		"us-west-2": true,
	}

	for _, region := range result.Regions {
		if !expectedRegions[region] {
			t.Errorf("Unexpected region: %s", region)
		}
	}

	// Verify SSO sessions
	if len(result.SsoSessions) != 1 {
		t.Errorf("Expected 1 SSO session, got %d", len(result.SsoSessions))
	}

	if len(result.SsoSessions) > 0 && result.SsoSessions[0] != "test-session" {
		t.Errorf("Expected SSO session 'test-session', got '%s'", result.SsoSessions[0])
	}
}

func TestGetFilterOptions_NoConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Don't create config file - should error
	_, err := GetFilterOptions()
	if err == nil {
		t.Error("Expected error when config file doesn't exist, got nil")
	}
}

// =============================================================================
// ADVANCED API TESTS
// =============================================================================

func TestNewSessionManagerDefault(t *testing.T) {
	test.SetupTestEnvironment(t)

	manager := NewSessionManagerDefault()
	if manager == nil {
		t.Fatal("Expected session manager, got nil")
	}

	// Verify manager has extractor
	if manager.extractor == nil {
		t.Error("Expected session manager to have extractor")
	}

	// Verify cache directory is set
	if manager.GetCacheDir() == "" {
		t.Error("Expected session manager to have cache directory")
	}
}

func TestNewSessionManagerWithExtractor(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := test.GetTestAwsConfigPath(t)
	extractor := NewExtractorWithPath(configPath)

	manager := NewSessionManagerWithExtractor(extractor)
	if manager == nil {
		t.Fatal("Expected session manager, got nil")
	}

	// Verify manager has the custom extractor
	if manager.extractor != extractor {
		t.Error("Expected session manager to have custom extractor")
	}
}

func TestNewExtractorDefault(t *testing.T) {
	test.SetupTestEnvironment(t)

	extractor := NewExtractorDefault()
	if extractor == nil {
		t.Fatal("Expected extractor, got nil")
	}

	// Verify extractor has config path set
	if extractor.GetConfigPath() == "" {
		t.Error("Expected extractor to have config path")
	}
}

func TestNewFilterDefault(t *testing.T) {
	filter := NewFilterDefault()
	if filter == nil {
		t.Fatal("Expected filter, got nil")
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestListProfiles_Integration(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create comprehensive test config
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	testConfig := `[default]
region = us-east-1

[profile sso-dev]
sso_session = org-commercial
sso_account_id = 123456789012
sso_role_name = Developer
region = us-east-1

[profile sso-prod]
sso_session = org-commercial
sso_account_id = 987654321098
sso_role_name = PowerUser
region = us-west-2

[profile iam-user]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
region = us-east-1

[profile assume-role]
role_arn = arn:aws:iam::123456789012:role/MyRole
source_profile = iam-user
region = us-east-1

[sso-session org-commercial]
sso_start_url = https://org.awsapps.com/start
sso_region = us-east-1
sso_registration_scopes = sso:account:access
`

	if err := os.WriteFile(configPath, []byte(testConfig), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test: List all profiles
	allResult, err := ListProfiles(FilterCriteria{})
	if err != nil {
		t.Fatalf("ListProfiles() error = %v", err)
	}

	if len(allResult.Profiles) != 5 { // default + 4 named profiles
		t.Errorf("Expected 5 total profiles, got %d", len(allResult.Profiles))
	}

	// Test: Filter by profile type (SSO only)
	ssoResult, err := ListProfiles(FilterCriteria{
		ProfileTypes: []ProfileType{ProfileTypeSSO},
	})
	if err != nil {
		t.Fatalf("ListProfiles() with SSO filter error = %v", err)
	}

	if len(ssoResult.Profiles) != 2 { // 2 SSO profiles
		t.Errorf("Expected 2 SSO profiles, got %d", len(ssoResult.Profiles))
	}

	// Test: Get filter options
	filterOptions, err := GetFilterOptions()
	if err != nil {
		t.Fatalf("GetFilterOptions() error = %v", err)
	}

	if len(filterOptions.AccountIDs) != 2 {
		t.Errorf("Expected 2 unique account IDs, got %d", len(filterOptions.AccountIDs))
	}

	if len(filterOptions.SsoSessions) != 1 {
		t.Errorf("Expected 1 SSO session, got %d", len(filterOptions.SsoSessions))
	}

	// Test: Get session status
	sessionStatus, err := GetSessionStatus(false)
	if err != nil {
		t.Fatalf("GetSessionStatus() error = %v", err)
	}

	if sessionStatus.LastChecked.IsZero() {
		t.Error("Expected session status to have LastChecked set")
	}
}

func TestListProfiles_EmptyConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create empty config
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(configPath, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to write empty config: %v", err)
	}

	result, err := ListProfiles(FilterCriteria{})
	if err != nil {
		t.Fatalf("ListProfiles() with empty config error = %v", err)
	}

	if len(result.Profiles) != 0 {
		t.Errorf("Expected 0 profiles for empty config, got %d", len(result.Profiles))
	}

	if len(result.SsoSessions) != 0 {
		t.Errorf("Expected 0 SSO sessions for empty config, got %d", len(result.SsoSessions))
	}
}
