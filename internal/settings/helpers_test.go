package settings

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestGetAwsDir(t *testing.T) {
	t.Run("with test environment", func(t *testing.T) {
		// Setup test environment (sets AWS_PROFILE_MANAGER_AWS_DIR)
		test.SetupTestEnvironment(t)

		awsDir := GetAwsDir()
		if awsDir == "" {
			t.Error("GetAwsDir should not return empty string")
		}

		// Should be absolute path in test environment
		if !filepath.IsAbs(awsDir) {
			t.Error("GetAwsDir should return absolute path")
		}
	})
}

func TestGetDesktopDir(t *testing.T) {
	t.Run("with test environment", func(t *testing.T) {
		// Setup test environment (sets AWS_PROFILE_MANAGER_DESKTOP_DIR)
		test.SetupTestEnvironment(t)

		desktopDir := GetDesktopDir()
		if desktopDir == "" {
			t.Error("GetDesktopDir should not return empty string")
		}

		// Should be absolute path in test environment
		if !filepath.IsAbs(desktopDir) {
			t.Error("GetDesktopDir should return absolute path")
		}
	})
}

func TestGetConfigDir(t *testing.T) {
	t.Run("with test environment", func(t *testing.T) {
		// Setup test environment (sets AWS_PROFILE_MANAGER_CONFIG_DIR)
		test.SetupTestEnvironment(t)

		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("GetConfigDir failed: %v", err)
		}
		if configDir == "" {
			t.Error("GetConfigDir should not return empty string")
		}
		if !filepath.IsAbs(configDir) {
			t.Error("GetConfigDir should return absolute path")
		}
	})
}

func TestGetCacheDir(t *testing.T) {
	t.Run("creates cache directory", func(t *testing.T) {
		// Setup test environment
		test.SetupTestEnvironment(t)
		configDir, err := GetConfigDir()
		if err != nil {
			t.Fatalf("Failed to get config dir: %v", err)
		}

		cacheDir, err := GetCacheDir()
		if err != nil {
			t.Fatalf("GetCacheDir failed: %v", err)
		}

		// Verify it's under config directory
		expectedCacheDir := filepath.Join(configDir, "cache")
		if cacheDir != expectedCacheDir {
			t.Errorf("Expected cache dir %s, got %s", expectedCacheDir, cacheDir)
		}

		// Verify directory was created
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			t.Error("Cache directory should have been created")
		}
	})

	t.Run("cache directory path structure", func(t *testing.T) {
		// Setup test environment
		test.SetupTestEnvironment(t)

		cacheDir, err := GetCacheDir()
		if err != nil {
			t.Fatalf("GetCacheDir failed: %v", err)
		}

		// Should end with "cache"
		if filepath.Base(cacheDir) != "cache" {
			t.Errorf("Expected cache directory to end with 'cache', got %s", filepath.Base(cacheDir))
		}
	})
}
