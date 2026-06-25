package sync

import (
	"context"
	"fmt"

	"aws-profile-manager/internal/security"
	"aws-profile-manager/internal/task"
)

// LocalFetcher fetches configuration from local file system.
type LocalFetcher struct {
	path string
}

// NewLocalFetcher creates a new local file fetcher.
//
// Parameters:
//   - path: Local file path to read
//
// Returns:
//   - *LocalFetcher: New local fetcher instance
func NewLocalFetcher(path string) *LocalFetcher {
	return &LocalFetcher{
		path: path,
	}
}

// Fetch retrieves configuration from local file system.
//
// This is a simple file read operation that respects context cancellation.
// Progress is reported before and after the read.
func (l *LocalFetcher) Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error) {
	reporter.ReportStatus(fmt.Sprintf("Reading local file: %s", l.path))

	// Check for cancellation before reading
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Read file
	data, err := security.ReadFile(l.path, security.ReadOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to read local file: %w", err)
	}

	reporter.ReportStatus(fmt.Sprintf("Read %d bytes from local file", len(data)))
	return data, nil
}

// String returns human-readable description.
func (l *LocalFetcher) String() string {
	return fmt.Sprintf("local file: %s", l.path)
}

// Validate checks if the local fetcher configuration is valid.
func (l *LocalFetcher) Validate() error {
	if l.path == "" {
		return fmt.Errorf("local path is required")
	}
	return nil
}
