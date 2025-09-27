package backup

import (
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// newTestConfig creates a Config for testing.
//
// This helper eliminates duplicate config building in every test.
//
// Parameters:
//   - t: Testing context
//
// Returns:
//   - Config: Test configuration with isolated paths
func newTestConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		ConfigPath:  test.GetTestAwsConfigPath(t),
		AwsDir:      test.GetTestAwsDir(t),
		StartMarker: "# START - Managed by AWS Profile Manager",
		EndMarker:   "# END - Managed by AWS Profile Manager",
	}
}

// newTestExportOptions creates default ExportOptions for testing.
//
// Provides a baseline export configuration that can be customized per test.
//
// Parameters:
//   - t: Testing context
//
// Returns:
//   - ExportOptions: Default export options with test paths
func newTestExportOptions(t *testing.T) ExportOptions {
	t.Helper()
	return ExportOptions{
		OutputPath:     filepath.Join(test.GetTestConfigDir(t), "backup.json"),
		IncludeManaged: true, // Default to including managed profiles
	}
}

// newTestImportOptions creates default ImportOptions for testing.
//
// Provides a baseline import configuration that can be customized per test.
//
// Parameters:
//   - t: Testing context
//   - backupPath: Path to backup file to import
//
// Returns:
//   - ImportOptions: Default import options
func newTestImportOptions(t *testing.T, backupPath string) ImportOptions {
	t.Helper()
	return ImportOptions{
		BackupPath:            backupPath,
		IncludeManaged:        true,  // Default to including managed profiles
		BackupCurrentSettings: false, // Disable in tests by default
	}
}
