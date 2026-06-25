package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadOptions controls path validation behavior for read operations.
type ReadOptions struct {
	// BaseDir constrains file access to this directory tree when set.
	BaseDir string

	// AllowedExtensions constrains file extensions (for example: ".json", ".txt").
	// Extension matching is case-insensitive.
	AllowedExtensions []string

	// AllowNonRegularFile permits reading non-regular files (directories, devices, etc.).
	// Default behavior requires regular files.
	AllowNonRegularFile bool
}

// ValidateReadPath normalizes and validates a file path for secure read operations.
//
// The returned path is absolute and symlink-resolved when possible.
func ValidateReadPath(path string, options ReadOptions) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

	absolutePath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}

	resolvedPath := absolutePath
	if symlinkPath, symlinkErr := filepath.EvalSymlinks(absolutePath); symlinkErr == nil {
		resolvedPath = symlinkPath
	} else if !os.IsNotExist(symlinkErr) {
		return "", symlinkErr
	}

	if options.BaseDir != "" {
		absoluteBaseDir, baseErr := filepath.Abs(filepath.Clean(options.BaseDir))
		if baseErr != nil {
			return "", baseErr
		}

		resolvedBaseDir := absoluteBaseDir
		if symlinkBaseDir, symlinkBaseErr := filepath.EvalSymlinks(absoluteBaseDir); symlinkBaseErr == nil {
			resolvedBaseDir = symlinkBaseDir
		} else if !os.IsNotExist(symlinkBaseErr) {
			return "", symlinkBaseErr
		}

		relPath, relErr := filepath.Rel(resolvedBaseDir, resolvedPath)
		if relErr != nil {
			return "", relErr
		}
		if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			return "", fmt.Errorf("path escapes base directory")
		}
	}

	if len(options.AllowedExtensions) > 0 {
		ext := strings.ToLower(filepath.Ext(resolvedPath))
		allowed := false
		for _, candidate := range options.AllowedExtensions {
			normalizedCandidate := strings.ToLower(candidate)
			if !strings.HasPrefix(normalizedCandidate, ".") {
				normalizedCandidate = "." + normalizedCandidate
			}
			if ext == normalizedCandidate {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("file extension %q is not allowed", ext)
		}
	}

	fileInfo, statErr := os.Stat(resolvedPath)
	if statErr != nil {
		return "", statErr
	}

	if !options.AllowNonRegularFile && !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("path is not a regular file")
	}

	return resolvedPath, nil
}

// ReadFile validates a path and reads file bytes.
func ReadFile(path string, options ReadOptions) ([]byte, error) {
	validatedPath, err := ValidateReadPath(path, options)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(validatedPath) // #nosec G304 -- validated by ValidateReadPath.
	if err != nil {
		return nil, err
	}

	return data, nil
}

// OpenFile validates a path and opens it for reading.
func OpenFile(path string, options ReadOptions) (*os.File, error) {
	validatedPath, err := ValidateReadPath(path, options)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(validatedPath) // #nosec G304 -- validated by ValidateReadPath.
	if err != nil {
		return nil, err
	}

	return file, nil
}
