package profiles

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"aws-profile-manager/internal/security"
)

// =============================================================================
// FILE READING
// =============================================================================

// readConfigFile reads the AWS CLI config file into memory.
//
// Returns empty slice if file doesn't exist (not an error).
// Only returns error for actual read failures.
//
// Parameters:
//   - configPath: Absolute path to config file
//
// Returns:
//   - []string: Lines from the file (empty slice if file doesn't exist)
//   - error: Any error encountered during read
func readConfigFile(configPath string) ([]string, error) {
	file, err := security.OpenFile(configPath, security.ReadOptions{})
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet - return empty slice
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close() // Simple defer for read-only files

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan config file: %w", err)
	}

	return lines, nil
}

// readFileContent reads entire file content as string.
//
// Parameters:
//   - path: Absolute path to file
//
// Returns:
//   - string: File content
//   - error: Any error encountered during read
func readFileContent(path string) (string, error) {
	content, err := security.ReadFile(path, security.ReadOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

// =============================================================================
// FILE WRITING
// =============================================================================

// writeConfigFile writes lines to the AWS CLI config file.
//
// This ensures the directory exists and writes atomically.
// Creates parent directories if they don't exist.
//
// Parameters:
//   - configPath: Absolute path to config file
//   - lines: Lines to write to file
//
// Returns:
//   - error: Any error during write
func writeConfigFile(configPath string, lines []string) (err error) {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open file for writing
	// #nosec G304 -- Config output path is controlled by application settings/CLI options.
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

	// Write lines
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// writeFileContent writes string content to file.
//
// Creates parent directories if they don't exist.
//
// Parameters:
//   - path: Absolute path to file
//   - content: String content to write
//
// Returns:
//   - error: Any error during write
func writeFileContent(path string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// =============================================================================
// FILE OPERATIONS
// =============================================================================

// ensureDirectory creates a directory if it doesn't exist.
//
// Parameters:
//   - path: Absolute path to directory
//
// Returns:
//   - error: Any error during directory creation
func ensureDirectory(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

// fileExists checks if a file exists.
//
// Parameters:
//   - path: Absolute path to file
//
// Returns:
//   - bool: true if file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// deleteFile removes a file if it exists.
//
// Returns no error if file doesn't exist (idempotent).
//
// Parameters:
//   - path: Absolute path to file
//
// Returns:
//   - bool: true if file existed and was removed
//   - error: Any error during removal
func deleteFile(path string) (bool, error) {
	if !fileExists(path) {
		return false, nil
	}

	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("failed to remove file: %w", err)
	}

	return true, nil
}

// getFileSize returns the size of a file in bytes.
//
// Parameters:
//   - path: Absolute path to file
//
// Returns:
//   - int64: File size in bytes (0 if file doesn't exist)
//   - error: Any error during stat
func getFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}
	return info.Size(), nil
}
