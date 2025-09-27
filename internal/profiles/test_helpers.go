package profiles

import (
	"testing"

	"aws-profile-manager/internal/test"
)

// newTestConfig creates a Config for testing with all paths pointing to temp directories.
//
// This ensures tests are completely isolated and don't create files in the repository.
// All tests MUST use this function instead of creating Config structures manually.
//
// Parameters:
//   - t: Testing context
//
// Returns:
//   - Config: Test configuration with temp directory paths
//
// Example:
//
//	config := newTestConfig(t)
//	installer := NewInstaller(config)
func newTestConfig(t *testing.T) Config {
	t.Helper()

	return Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
	}
}
