// Package test provides utilities for test isolation and environment setup.
//
// This package contains helpers for creating isolated test environments with
// temporary directories and environment variables. It ensures tests don't interfere
// with each other or with the user's actual files.
//
// Key Features:
//   - Automatic temporary directory creation
//   - Environment variable isolation and restoration
//   - Logger suppression for cleaner test output
//   - Helper functions for accessing test directories
//
// Usage Pattern:
//
//	func TestMyFunction(t *testing.T) {
//	    test.SetupTestEnvironment(t)
//	    // Test uses isolated environment
//	    // Automatic cleanup via t.Cleanup()
//	}
package test

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestEnvironment configures a completely isolated test environment.
//
// This function is the primary entry point for test isolation. It creates temporary
// directories, sets environment variables, suppresses logger output, and ensures
// automatic cleanup when the test completes.
//
// Directory Structure Created:
//   - config/ - Application configuration directory
//   - config/cache/ - Cache subdirectory
//   - .aws/ - AWS CLI directory
//   - Desktop/ - Desktop directory for exports
//
// Environment Variables Set:
//   - AWS_PROFILE_MANAGER_DEBUG=0 (unless AWS_PROFILE_MANAGER_TEST_DEBUG=1)
//   - AWS_PROFILE_MANAGER_CONFIG_DIR=<temp>/config
//   - AWS_PROFILE_MANAGER_AWS_DIR=<temp>/.aws
//   - AWS_PROFILE_MANAGER_DESKTOP_DIR=<temp>/Desktop
//   - AWS_PROFILE_MANAGER_SILENCE_LOGGER=1
//
// All original environment variables are automatically restored via t.Cleanup().
//
// Parameters:
//   - t: Testing instance (used for t.TempDir() and t.Cleanup())
//
// Example:
//
//	func TestMyFunction(t *testing.T) {
//	    test.SetupTestEnvironment(t)
//
//	    // All file operations are now isolated
//	    configDir := test.GetTestConfigDir(t)
//	    awsDir := test.GetTestAwsDir(t)
//
//	    // Test with isolated directories...
//	}
//
// Debug Mode:
//
//	Set AWS_PROFILE_MANAGER_TEST_DEBUG=1 to enable debug logging during tests:
//	    AWS_PROFILE_MANAGER_TEST_DEBUG=1 go test ./...
func SetupTestEnvironment(t *testing.T) {
	t.Helper()

	// Suppress logger output during tests
	SuppressLogger(t)

	// Create base temp directory for this test
	tempDir := t.TempDir() // Used for testing on tmp files

	// Save all original environment variables
	originalDebug := os.Getenv("AWS_PROFILE_MANAGER_DEBUG")
	originalSilenceLogger := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
	originalConfigDir := os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")
	originalAwsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")
	originalDesktopDir := os.Getenv("AWS_PROFILE_MANAGER_DESKTOP_DIR")

	// Set all environment variables to temp subdirectories
	testConfigDir := filepath.Join(tempDir, "config")
	testAwsDir := filepath.Join(tempDir, ".aws")
	testDesktopDir := filepath.Join(tempDir, "Desktop")
	testCacheDir := filepath.Join(testConfigDir, "cache")

	// Create all directories including cache
	for _, dir := range []string{testConfigDir, testCacheDir, testAwsDir, testDesktopDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Assign environment variables for testing
	if err := os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "0"); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_DEBUG: %v", err)
	}
	if err := os.Setenv("AWS_PROFILE_MANAGER_TEST_HOME_DIR", tempDir); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_TEST_HOME: %v", err)
	}
	if err := os.Setenv("AWS_PROFILE_MANAGER_CONFIG_DIR", testConfigDir); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_CONFIG_DIR: %v", err)
	}
	if err := os.Setenv("AWS_PROFILE_MANAGER_AWS_DIR", testAwsDir); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_AWS_DIR: %v", err)
	}
	if err := os.Setenv("AWS_PROFILE_MANAGER_DESKTOP_DIR", testDesktopDir); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_DESKTOP_DIR: %v", err)
	}
	if err := os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "1"); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_SILENCE_LOGGER: %v", err)
	}

	// Testing overrides
	if os.Getenv("AWS_PROFILE_MANAGER_TEST_DEBUG") == "1" {
		if err := os.Setenv("AWS_PROFILE_MANAGER_DEBUG", "1"); err != nil {
			t.Fatalf("Failed to set AWS_PROFILE_MANAGER_DEBUG: %v", err)
		}
		os.Unsetenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")
		t.Log("Forcing debug to true for testing")
	}

	// Restore all environment variables on cleanup
	t.Cleanup(func() {
		restoreEnv("AWS_PROFILE_MANAGER_DEBUG", originalDebug)
		restoreEnv("AWS_PROFILE_MANAGER_CONFIG_DIR", originalConfigDir)
		restoreEnv("AWS_PROFILE_MANAGER_AWS_DIR", originalAwsDir)
		restoreEnv("AWS_PROFILE_MANAGER_DESKTOP_DIR", originalDesktopDir)
		restoreEnv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalSilenceLogger)
		os.Unsetenv("AWS_PROFILE_MANAGER_TEST_HOME_DIR")
	})
}

// restoreEnv restores an environment variable to its original value.
//
// If the original value was empty, the variable is unset to maintain
// the original environment state.
//
// Parameters:
//   - key: Environment variable name
//   - originalValue: Original value to restore (empty string means unset)
func restoreEnv(key, originalValue string) {
	if originalValue == "" {
		_ = os.Unsetenv(key)
	} else {
		_ = os.Setenv(key, originalValue)
	}
}

// SuppressLogger silences all logger output for the duration of the test.
//
// This function is useful for tests that don't use SetupTestEnvironment() but
// still need to suppress logger output for cleaner test results. The logger
// is automatically restored to its original state after the test completes.
//
// Note: SetupTestEnvironment() calls this automatically, so no need to call
// it separately when using SetupTestEnvironment().
//
// Parameters:
//   - t: Testing instance (used for t.Cleanup())
//
// Example:
//
//	func TestMyFunction(t *testing.T) {
//	    test.SuppressLogger(t)
//	    // Logger output is now silenced
//	    // Automatic restoration via t.Cleanup()
//	}
func SuppressLogger(t *testing.T) {
	t.Helper()

	// Save original value
	originalValue := os.Getenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER")

	// Restore after test
	t.Cleanup(func() {
		restoreEnv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", originalValue)
	})

	// Silence logger output during test
	if err := os.Setenv("AWS_PROFILE_MANAGER_SILENCE_LOGGER", "1"); err != nil {
		t.Fatalf("Failed to set AWS_PROFILE_MANAGER_SILENCE_LOGGER: %v", err)
	}
}

// GetTestHomeDir returns the test home directory path.
//
// This function retrieves the base temporary directory created by
// SetupTestEnvironment(). Must be called after SetupTestEnvironment().
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test home directory
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestHomeDir(t *testing.T) string {
	t.Helper()
	homeDir := os.Getenv("AWS_PROFILE_MANAGER_TEST_HOME_DIR")
	if homeDir == "" {
		t.Fatal("AWS_PROFILE_MANAGER_TEST_HOME_DIR not set - did you call SetupTestEnvironment first?")
	}
	return homeDir
}

// GetTestConfigDir returns the test config directory path.
//
// This function retrieves the temporary config directory path created by
// SetupTestEnvironment(). This is where settings.json and other config
// files should be stored during tests.
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test config directory (e.g., /tmp/test123/config)
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestConfigDir(t *testing.T) string {
	t.Helper()
	configDir := os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR")
	if configDir == "" {
		t.Fatal("AWS_PROFILE_MANAGER_CONFIG_DIR not set - did you call SetupTestEnvironment first?")
	}
	return configDir
}

// GetTestAwsDir returns the test AWS directory path.
//
// This function retrieves the temporary AWS directory path created by
// SetupTestEnvironment(). This is where AWS CLI config and credentials
// files should be stored during tests.
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test AWS directory (e.g., /tmp/test123/.aws)
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestAwsDir(t *testing.T) string {
	t.Helper()
	awsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR")
	if awsDir == "" {
		t.Fatal("AWS_PROFILE_MANAGER_AWS_DIR not set - did you call SetupTestEnvironment first?")
	}
	return awsDir
}

// GetTestAwsConfigPath returns the test AWS config file path.
//
// This is a convenience function that combines GetTestAwsDir() with the
// standard AWS config filename.
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test AWS config file (e.g., /tmp/test123/.aws/config)
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestAwsConfigPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(GetTestAwsDir(t), "config")
}

// GetTestDesktopDir returns the test desktop directory path.
//
// This function retrieves the temporary desktop directory path created by
// SetupTestEnvironment(). This is where exported files (like cheat sheets)
// should be stored during tests.
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test desktop directory (e.g., /tmp/test123/Desktop)
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestDesktopDir(t *testing.T) string {
	t.Helper()
	desktopDir := os.Getenv("AWS_PROFILE_MANAGER_DESKTOP_DIR")
	if desktopDir == "" {
		t.Fatal("AWS_PROFILE_MANAGER_DESKTOP_DIR not set - did you call SetupTestEnvironment first?")
	}
	return desktopDir
}

// GetTestCacheDir returns the test cache directory path.
//
// This is a convenience function that returns the cache subdirectory within
// the test config directory. Cache files should be stored here during tests.
//
// Parameters:
//   - t: Testing instance
//
// Returns:
//   - string: Path to test cache directory (e.g., /tmp/test123/config/cache)
//
// Panics if SetupTestEnvironment() was not called first.
func GetTestCacheDir(t *testing.T) string {
	t.Helper()
	configDir := GetTestConfigDir(t)
	return filepath.Join(configDir, "cache")
}
