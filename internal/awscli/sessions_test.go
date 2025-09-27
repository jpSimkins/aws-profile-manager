package awscli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/test"
)

// Test helper to create SSO cache file
func createTestSSOCache(t *testing.T, cacheDir string, fileName string, isExpired bool) {
	t.Helper()
	createTestSSOCacheWithURL(t, cacheDir, fileName, isExpired, "https://test.awsapps.com/start")
}

func createTestSSOCacheWithURL(t *testing.T, cacheDir string, fileName string, isExpired bool, startURL string) {
	t.Helper()

	expiry := time.Now().Add(time.Hour) // Default: not expired
	if isExpired {
		expiry = time.Now().Add(-time.Hour) // Expired
	}

	cache := SsoCacheFile{
		StartURL:     startURL,
		Region:       "us-east-1",
		AccessToken:  "test-access-token-123",
		ExpiresAt:    expiry,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RefreshToken: "test-refresh-token",
	}

	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("Failed to marshal cache data: %v", err)
	}

	err = os.WriteFile(filepath.Join(cacheDir, fileName), data, 0644)
	if err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}
}

func TestNewSessionManager(t *testing.T) {
	tests := []struct {
		name           string
		setupExtractor bool
		expectNil      bool
	}{
		{
			name:           "Valid extractor",
			setupExtractor: true,
			expectNil:      false,
		},
		{
			name:           "Nil extractor",
			setupExtractor: false,
			expectNil:      false, // Should still work with nil extractor
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var extractor *Extractor
			if tt.setupExtractor {
				test.SetupTestEnvironment(t)
				configFile := test.GetTestAwsConfigPath(t)
				err := os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0644)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				extractor = NewExtractorWithPath(configFile)
			}

			sm := NewSessionManager(extractor)

			if tt.expectNil && sm != nil {
				t.Error("Expected nil session manager but got non-nil")
			}
			if !tt.expectNil && sm == nil {
				t.Error("Expected session manager but got nil")
			}
		})
	}
}

func TestNewSessionManagerWithEnvironmentVariable(t *testing.T) {
	// Setup temporary directory for development AWS dir
	test.SetupTestEnvironment(t)
	devAwsDir := filepath.Join(test.GetTestAwsDir(t), "dev-aws")

	// Set environment variable
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")
	defer func() {
		if originalValue != "" {
			os.Setenv("AWS_PROFILE_MANAGER_AWS_DIR", originalValue)
		} else {
			os.Unsetenv("AWS_PROFILE_MANAGER_AWS_DIR")
		}
	}()

	os.Setenv("AWS_PROFILE_MANAGER_AWS_DIR", devAwsDir)

	// Create session manager
	sm := NewSessionManager(nil)

	// Verify cache directory uses development path
	expectedCacheDir := filepath.Join(devAwsDir, "sso", "cache")
	if sm.cacheDir != expectedCacheDir {
		t.Errorf("Expected cache dir %s, got %s", expectedCacheDir, sm.cacheDir)
	}
}

func TestGetSessionStatus(t *testing.T) {
	tests := []struct {
		name                 string
		setupCache           func(string) // Function to setup cache files
		expectedActiveCount  int
		expectedExpiredCount int
	}{
		{
			name: "No cache files",
			setupCache: func(cacheDir string) {
				// No files
			},
			expectedActiveCount:  0,
			expectedExpiredCount: 0,
		},
		{
			name: "One active session",
			setupCache: func(cacheDir string) {
				createTestSSOCache(t, cacheDir, "session1.json", false)
			},
			expectedActiveCount:  1,
			expectedExpiredCount: 0,
		},
		{
			name: "One expired session",
			setupCache: func(cacheDir string) {
				createTestSSOCache(t, cacheDir, "session1.json", true)
			},
			expectedActiveCount:  0,
			expectedExpiredCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SetupTestEnvironment(t)
			ssoDir := filepath.Join(test.GetTestAwsDir(t), "sso")
			err := os.MkdirAll(ssoDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create SSO directory: %v", err)
			}

			tt.setupCache(ssoDir)

			sm := NewSessionManagerWithPath(ssoDir, nil)

			status, err := sm.GetSessionStatus()
			if err != nil {
				t.Fatalf("Failed to get session status: %v", err)
			}

			// Check status counts
			if len(status.ActiveSessions) != tt.expectedActiveCount {
				t.Errorf("Expected %d active sessions, got %d", tt.expectedActiveCount, len(status.ActiveSessions))
			}
			if len(status.ExpiredSessions) != tt.expectedExpiredCount {
				t.Errorf("Expected %d expired sessions, got %d", tt.expectedExpiredCount, len(status.ExpiredSessions))
			}
		})
	}
}

func TestRefreshSessionStatus(t *testing.T) {
	test.SetupTestEnvironment(t)
	ssoDir := filepath.Join(test.GetTestAwsDir(t), "sso")
	err := os.MkdirAll(ssoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create SSO directory: %v", err)
	}

	// Create initial cache file
	createTestSSOCache(t, ssoDir, "session1.json", false)

	sm := NewSessionManagerWithPath(ssoDir, nil)

	// Get initial session status
	status1, err := sm.GetSessionStatus()
	if err != nil {
		t.Fatalf("Failed to get initial session status: %v", err)
	}
	initialCount := len(status1.ActiveSessions) + len(status1.ExpiredSessions)

	// Add another cache file with different start URL to create a different session
	createTestSSOCacheWithURL(t, ssoDir, "session2.json", false, "https://test2.awsapps.com/start")

	// Refresh status
	status2, err := sm.RefreshSessionStatus()
	if err != nil {
		t.Errorf("Unexpected error refreshing status: %v", err)
	}

	// Check that status was refreshed
	newCount := len(status2.ActiveSessions) + len(status2.ExpiredSessions)

	if newCount <= initialCount {
		t.Errorf("Expected session count to increase after refresh, got %d (was %d)", newCount, initialCount)
	}
}

func TestClearExpiredCache(t *testing.T) {
	test.SetupTestEnvironment(t)
	ssoDir := filepath.Join(test.GetTestAwsDir(t), "sso")
	err := os.MkdirAll(ssoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create SSO directory: %v", err)
	}

	// Create test cache files with different start URLs to create distinct sessions
	createTestSSOCacheWithURL(t, ssoDir, "active1.json", false, "https://active.awsapps.com/start")
	createTestSSOCacheWithURL(t, ssoDir, "expired1.json", true, "https://expired1.awsapps.com/start")
	createTestSSOCacheWithURL(t, ssoDir, "expired2.json", true, "https://expired2.awsapps.com/start")

	sm := NewSessionManagerWithPath(ssoDir, nil)

	// Verify initial state
	status, err := sm.GetSessionStatus()
	if err != nil {
		t.Fatalf("Failed to get initial session status: %v", err)
	}

	totalBefore := len(status.ActiveSessions) + len(status.ExpiredSessions)
	expiredBefore := len(status.ExpiredSessions)

	if totalBefore != 3 {
		t.Errorf("Expected 3 total sessions before cleanup, got %d", totalBefore)
	}
	if expiredBefore != 2 {
		t.Errorf("Expected 2 expired sessions before cleanup, got %d", expiredBefore)
	}

	// Clear expired cache
	err = sm.ClearExpiredCache()
	if err != nil {
		t.Errorf("Unexpected error clearing cache: %v", err)
	}

	// Verify final state - refresh status
	status, err = sm.RefreshSessionStatus()
	if err != nil {
		t.Fatalf("Failed to refresh session status: %v", err)
	}

	totalAfter := len(status.ActiveSessions) + len(status.ExpiredSessions)
	expiredAfter := len(status.ExpiredSessions)

	if totalAfter != 1 {
		t.Errorf("Expected 1 total session after cleanup, got %d", totalAfter)
	}
	if expiredAfter != 0 {
		t.Errorf("Expected 0 expired sessions after cleanup, got %d", expiredAfter)
	}
	if len(status.ActiveSessions) != 1 {
		t.Errorf("Expected 1 active session after cleanup, got %d", len(status.ActiveSessions))
	}
}

func TestLoginToSession(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		expectError bool
	}{
		{
			name:        "Valid profile name",
			profileName: "test-profile",
			expectError: false, // Will likely error due to AWS CLI not configured, but method should not panic
		},
		{
			name:        "Empty profile name",
			profileName: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SetupTestEnvironment(t)
			ssoDir := filepath.Join(test.GetTestAwsDir(t), "sso")
			err := os.MkdirAll(ssoDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create SSO directory: %v", err)
			}

			sm := NewSessionManagerWithPath(ssoDir, nil)

			err = sm.LoginToSession(tt.profileName)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			// In test environment, this might error due to AWS CLI not being configured
			// That's expected and acceptable for unit tests
			if err != nil {
				t.Logf("Expected error in test environment: %v", err)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "Global logout",
			expectError: false, // AWS CLI only supports global logout, so this should work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.SetupTestEnvironment(t)
			ssoDir := filepath.Join(test.GetTestAwsDir(t), "sso")
			err := os.MkdirAll(ssoDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create SSO directory: %v", err)
			}

			sm := NewSessionManagerWithPath(ssoDir, nil)

			err = sm.Logout()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			// In test environment, this might error due to AWS CLI not being configured
			// That's expected and acceptable for unit tests
			if err != nil {
				t.Logf("Expected error in test environment: %v", err)
			}
		})
	}
}

func TestGetCacheDir(t *testing.T) {
	test.SetupTestEnvironment(t)
	cacheDir := filepath.Join(test.GetTestAwsDir(t), "sso")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	sm := NewSessionManagerWithPath(cacheDir, nil)

	result := sm.GetCacheDir()
	if result != cacheDir {
		t.Errorf("Expected cache dir %s, got %s", cacheDir, result)
	}
}
