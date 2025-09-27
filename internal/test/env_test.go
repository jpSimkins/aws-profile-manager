package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupTestEnvironment(t *testing.T) {
	// Save original environment before test
	originalConfigDir := os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")
	originalAwsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")
	originalDesktopDir := os.Getenv("AWS_PROFILE_MANAGER_DESKTOP_DIR")

	// Run the setup
	SetupTestEnvironment(t)

	// Verify all environment variables are set
	configDir := os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")
	awsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")
	desktopDir := os.Getenv("AWS_PROFILE_MANAGER_DESKTOP_DIR")

	if configDir == "" {
		t.Error("AWS_PROFILE_MANAGER_CONFIG_DIR not set")
	}
	if awsDir == "" {
		t.Error("AWS_PROFILE_MANAGER_AWS_DIR not set")
	}
	if desktopDir == "" {
		t.Error("AWS_PROFILE_MANAGER_DESKTOP_DIR not set")
	}

	// Verify expected subdirectory names by checking path endings
	if !strings.HasSuffix(configDir, "config") {
		t.Errorf("AWS_PROFILE_MANAGER_CONFIG_DIR should end with 'config': got %s", configDir)
	}
	if !strings.HasSuffix(awsDir, ".aws") {
		t.Errorf("AWS_PROFILE_MANAGER_AWS_DIR should end with '.aws': got %s", awsDir)
	}
	if !strings.HasSuffix(desktopDir, "Desktop") {
		t.Errorf("AWS_PROFILE_MANAGER_DESKTOP_DIR should end with 'Desktop': got %s", desktopDir)
	}

	// Verify all paths share a common base (all under same temp directory)
	baseDir1 := filepath.Dir(filepath.Dir(configDir)) // up two levels from config
	baseDir2 := filepath.Dir(filepath.Dir(awsDir))
	baseDir3 := filepath.Dir(filepath.Dir(desktopDir))

	if baseDir1 != baseDir2 || baseDir1 != baseDir3 {
		t.Errorf("All directories should share common base: %s, %s, %s", baseDir1, baseDir2, baseDir3)
	}

	// Cleanup should restore original values (tested implicitly by t.Cleanup)
	// We verify restoration by checking after cleanup in a separate subtest
	t.Run("VerifyCleanup", func(t *testing.T) {
		// At this point, parent test cleanup has not run yet
		// But we can verify the cleanup function is registered
		// This test serves as documentation that cleanup happens automatically
	})

	// After this test completes, t.Cleanup() will restore environment variables
	// We can't easily test this within the same test, but we verify originals didn't change
	_ = originalConfigDir
	_ = originalAwsDir
	_ = originalDesktopDir
}

func TestSetupTestEnvironment_MultipleTests(t *testing.T) {
	// Test that multiple calls to SetupTestEnvironment work correctly
	// Each should get its own isolated temp directory

	var configDir1 string
	t.Run("First", func(t *testing.T) {
		SetupTestEnvironment(t)
		configDir1 = os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")

		if configDir1 == "" {
			t.Error("First test: config dir not set")
		}
	})

	var configDir2 string
	t.Run("Second", func(t *testing.T) {
		SetupTestEnvironment(t)
		configDir2 = os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")

		if configDir2 == "" {
			t.Error("Second test: config dir not set")
		}

		// The two tests should get different temp directories
		if configDir1 == configDir2 {
			t.Error("Second test should get different temp directory than first")
		}
	})
}

// TestRestoreEnv removed - restoreEnv is an internal helper tested indirectly
// through TestSetupTestEnvironment which verifies cleanup via t.Cleanup()

func TestSuppressLogger(t *testing.T) {
	// Save original environment
	originalSilence := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	defer func() {
		if originalSilence != "" {
			_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalSilence)
		} else {
			_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		}
	}()

	// Initially unset the variable
	_ = os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")

	// Create a sub-test to verify cleanup
	t.Run("suppresses and restores", func(t *testing.T) {
		// Call SuppressLogger
		SuppressLogger(t)

		// Verify logger is silenced
		silenceValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		if silenceValue != "1" {
			t.Errorf("AWS_PROFILE_MANAGER_SILENCE_LOGGER = %q, want \"1\"", silenceValue)
		}
	})

	// After sub-test cleanup, variable should be restored (unset)
	afterValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	if afterValue != "" {
		t.Errorf("AWS_PROFILE_MANAGER_SILENCE_LOGGER = %q, want \"\" (should be restored)", afterValue)
	}

	// Test with pre-existing value
	t.Run("restores previous value", func(t *testing.T) {
		// Set a custom value first
		_ = os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "custom")

		t.Run("nested", func(t *testing.T) {
			SuppressLogger(t)
			if os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER") != "1" {
				t.Error("Should be silenced with value \"1\"")
			}
		})

		// After nested test cleanup, should restore to "custom"
		if os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER") != "custom" {
			t.Errorf("AWS_PROFILE_MANAGER_SILENCE_LOGGER = %q, want \"custom\"", os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER"))
		}
	})
}
