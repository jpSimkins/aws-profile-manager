package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"aws-profile-manager/internal/security"
)

// ReadBackupFile reads and parses a backup file.
//
// This function reads the backup file from disk, parses the JSON, and
// returns the structured BackupFile. It performs basic validation to
// ensure the file format is correct.
//
// Parameters:
//   - path: Path to backup file
//
// Returns:
//   - *BackupFile: Parsed backup file
//   - error: Any error encountered during reading or parsing
//
// Example:
//
//	backup, err := backup.ReadBackupFile("/path/to/backup.json")
//	if err != nil {
//	    return fmt.Errorf("failed to read backup: %w", err)
//	}
func ReadBackupFile(path string) (*BackupFile, error) {
	// Read file
	data, err := security.ReadFile(path, security.ReadOptions{
		AllowedExtensions: []string{".json"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read backup file: %w", err)
	}

	// Parse JSON
	var backup BackupFile
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("failed to parse backup file: %w", err)
	}

	// Validate structure
	if err := ValidateBackupFile(&backup); err != nil {
		return nil, fmt.Errorf("invalid backup file: %w", err)
	}

	return &backup, nil
}

// WriteBackupFile writes a backup file to disk.
//
// This function validates the backup structure, serializes it to JSON
// with indentation for readability, and writes it to the specified path.
// The parent directory is created if it doesn't exist.
//
// Parameters:
//   - path: Where to write file
//   - backup: Backup file to write
//
// Returns:
//   - error: Any error encountered during validation or writing
//
// Example:
//
//	backup := &backup.BackupFile{
//	    Version: "2.0",
//	    // ... other fields
//	}
//	if err := backup.WriteBackupFile("/path/to/backup.json", backup); err != nil {
//	    return fmt.Errorf("failed to write backup: %w", err)
//	}
func WriteBackupFile(path string, backup *BackupFile) error {
	// Validate before writing
	if err := ValidateBackupFile(backup); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// ValidateBackupFile validates backup file structure.
//
// This function performs validation checks on the backup file to ensure
// it has all required fields and valid structure. It checks:
//   - Version is specified
//   - At least one of Schema or Settings is present
//   - Metadata timestamp is not zero
//
// Parameters:
//   - backup: Backup file to validate
//
// Returns:
//   - error: Validation errors, or nil if valid
//
// Example:
//
//	if err := backup.ValidateBackupFile(backupFile); err != nil {
//	    return fmt.Errorf("invalid backup: %w", err)
//	}
func ValidateBackupFile(backup *BackupFile) error {
	if backup == nil {
		return fmt.Errorf("backup file is nil")
	}

	// Check version
	if backup.Version == "" {
		return fmt.Errorf("backup version is required")
	}

	// Ensure at least one section is present
	if backup.Data == nil && backup.Settings == nil {
		return fmt.Errorf("backup must contain either data or settings (or both)")
	}

	// Check metadata
	if backup.Metadata.ExportedAt.IsZero() {
		return fmt.Errorf("backup metadata timestamp is required")
	}

	if backup.Metadata.ToolVersion == "" {
		return fmt.Errorf("backup metadata tool version is required")
	}

	return nil
}
