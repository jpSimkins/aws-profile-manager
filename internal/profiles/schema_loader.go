package profiles

import (
	"context"
	"fmt"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// SchemaReader reads schema data from an AWS config file.
//
// This component is intended for read-only consumers (for example GUI views)
// that need schema data from the local AWS config without writing an export
// backup file.
type SchemaReader struct {
	config Config
	reader *configReader
}

// NewSchemaReader creates a new SchemaReader with injected configuration.
//
// Parameters:
//   - config: Configuration with AWS config path and marker settings
//
// Returns:
//   - *SchemaReader: Ready-to-use schema reader component
func NewSchemaReader(config Config) *SchemaReader {
	return &SchemaReader{
		config: config,
		reader: newConfigReader(config),
	}
}

// Read loads schema data from the configured AWS config file.
//
// Parameters:
//   - ctx: Context for cancellation
//   - options: Section include options
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - *schema.Schema: Extracted schema from requested sections
//   - error: Any error encountered while reading/parsing config
func (r *SchemaReader) Read(
	ctx context.Context,
	options SchemaReadOptions,
	reporter task.Reporter,
) (*schema.Schema, error) {
	if !options.IncludeManaged && !options.IncludeUnmanagedAbove && !options.IncludeUnmanagedBelow {
		return nil, fmt.Errorf("at least one schema section must be included")
	}

	parsedSchema, _, err := r.reader.readConfig(ctx, ExportOptions{
		IncludeManaged: options.IncludeManaged,
		IncludeAbove:   options.IncludeUnmanagedAbove,
		IncludeBelow:   options.IncludeUnmanagedBelow,
	}, reporter)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema from aws config: %w", err)
	}

	return parsedSchema, nil
}
