package settings

import (
	"os"
	"path/filepath"
	"runtime"

	"aws-profile-manager/internal/logging"
)

// GetAwsDir returns the AWS CLI directory path.
//
// This function respects the AWS_PROFILE_MANAGER_AWS_DIR environment variable,
// falling back to the standard ~/.aws/ directory if not set.
//
// Returns:
//   - string: Absolute path to AWS CLI directory
func GetAwsDir() string {
	if envAwsDir := os.Getenv("AWS_PROFILE_MANAGER_AWS_DIR"); envAwsDir != "" {
		absPath, err := filepath.Abs(envAwsDir)
		if err == nil {
			return absPath
		}
	}

	// Default to standard AWS directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".aws")
}

// GetDesktopDir returns the Desktop directory path.
//
// This function respects the AWS_PROFILE_MANAGER_DESKTOP_DIR environment
// variable, falling back to platform-specific desktop directories.
//
// Returns:
//   - string: Absolute path to Desktop directory
func GetDesktopDir() string {
	if envDesktopDir := os.Getenv("AWS_PROFILE_MANAGER_DESKTOP_DIR"); envDesktopDir != "" {
		absPath, err := filepath.Abs(envDesktopDir)
		if err == nil {
			return absPath
		}
	}

	// Default to platform-specific desktop directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Platform-specific desktop paths
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(homeDir, "Desktop")
	case "darwin":
		return filepath.Join(homeDir, "Desktop")
	default: // Linux and others
		return filepath.Join(homeDir, "Desktop")
	}
}

// GetCacheDir returns the cache directory path.
//
// The cache directory is located within the config directory as a subdirectory.
// This function ensures the cache directory exists, creating it if necessary.
//
// Returns:
//   - string: Absolute path to cache directory (e.g., ~/.config/aws-profile-manager/cache)
//   - error: Any error encountered
func GetCacheDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", logging.Log.ErrorfWithDetails("failed to get config directory", err)
	}

	cacheDir := filepath.Join(configDir, "cache")

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", logging.Log.ErrorfWithDetails("failed to create cache directory", err,
			"dir", cacheDir)
	}

	logging.Debug.Log("Cache directory resolved", "path", cacheDir)
	return cacheDir, nil
}

// GetConfigDir returns the application configuration directory path.
//
// This function respects the AWS_PROFILE_MANAGER_CONFIG_DIR environment variable,
// falling back to platform-specific config directories:
//   - Linux/macOS: ~/.config/aws-profile-manager/
//   - Windows: %APPDATA%\aws-profile-manager\
//
// Returns:
//   - string: Absolute path to config directory
//   - error: Any error encountered
func GetConfigDir() (string, error) {
	// Check for environment variable override
	if envConfigDir := os.Getenv("AWS_PROFILE_MANAGER_CONFIG_DIR"); envConfigDir != "" {
		absPath, err := filepath.Abs(envConfigDir)
		if err != nil {
			return "", logging.Log.ErrorfWithDetails("invalid config directory path", err,
				"path", envConfigDir)
		}
		logging.Debug.Log("Config directory from environment", "path", absPath)
		return absPath, nil
	}

	// Use standard cross-platform config directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", logging.Log.ErrorfWithDetails("failed to get user config directory", err)
	}

	configDir := filepath.Join(userConfigDir, "aws-profile-manager")
	logging.Debug.Log("Config directory using default", "path", configDir)
	return configDir, nil
}
