package sync

import (
	"context"

	"aws-profile-manager/internal/task"
)

// Fetcher retrieves configuration from a remote source.
//
// All implementations MUST support context cancellation and progress reporting.
// Fetchers are responsible for retrieving raw JSON data, which is then parsed
// by the caller into a schema.Schema.
type Fetcher interface {
	// Fetch retrieves the configuration file.
	//
	// MUST respect ctx.Done() and cancel gracefully.
	// SHOULD report progress via reporter for user feedback.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadlines
	//   - reporter: Progress reporter for status updates
	//
	// Returns:
	//   - []byte: Raw JSON configuration data
	//   - error: Any error encountered during fetch
	Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error)

	// String returns human-readable source description
	String() string

	// Validate checks if the fetcher configuration is valid
	Validate() error
}
