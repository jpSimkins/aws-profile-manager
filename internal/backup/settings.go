package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"aws-profile-manager/internal/settings"
)

// BackupSettings creates a backup of current settings.
//
// This function saves the current application settings to a single backup
// file. Each backup overwrites the previous one to avoid cluttering the
// config directory.
//
// The backup file is: settings-backup.json (in config directory)
//
// Parameters:
//   - cfg: Configuration (paths)
//
// Returns:
//   - backupPath: Path to settings backup file
//   - error: Any error encountered
//
// Example:
//
//	cfg := backup.Config{
//	    ConfigPath: settings.GetAwsConfigPath(),
//	    AwsDir:     settings.GetAwsDir(),
//	}
//	backupPath, err := backup.BackupSettings(cfg)
//	if err != nil {
//	    return fmt.Errorf("failed to backup settings: %w", err)
//	}
func BackupSettings(cfg Config) (string, error) {
	// Get current settings
	currentSettings := settings.Get()

	// Get config directory for backup
	configDir, err := settings.GetConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Create single backup filename (overwrites previous backup)
	backupFilename := "settings-backup.json"
	backupPath := filepath.Join(configDir, backupFilename)

	// Marshal settings to JSON
	data, err := json.MarshalIndent(currentSettings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to backup file
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write settings backup: %w", err)
	}

	return backupPath, nil
}

// RestoreSettings restores settings from a Settings object.
//
// This function validates and applies the provided settings. If validation
// fails, the current settings are not modified and an error is returned.
//
// Parameters:
//   - settings: Settings to restore
//
// Returns:
//   - error: Any error encountered during validation or restoration
//
// Example:
//
//	// Read backup file
//	backup, err := backup.ReadBackupFile("/path/to/backup.json")
//	if err != nil {
//	    return err
//	}
//
//	// Restore settings if present
//	if backup.Settings != nil {
//	    if err := backup.RestoreSettings(backup.Settings); err != nil {
//	        return fmt.Errorf("failed to restore settings: %w", err)
//	    }
//	}
func RestoreSettings(s *settings.Settings) error {
	if s == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Set settings (this validates automatically)
	if err := settings.Set(s); err != nil {
		return fmt.Errorf("failed to set settings: %w", err)
	}

	// Save settings to disk
	configDir, err := settings.GetConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}
	settingsPath := filepath.Join(configDir, "settings.json")

	if err := settings.Save(settingsPath); err != nil {
		return fmt.Errorf("failed to save settings to disk: %w", err)
	}

	return nil
}
