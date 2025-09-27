package profiles

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/task"
)

// Exporter exports AWS CLI profiles to JSON.
//
// This component reads AWS config file and exports selected sections to
// portable JSON format for backup or distribution.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	exporter := profiles.NewExporter(config)
//	result, err := exporter.Export(opts)
type Exporter struct {
	config Config
	reader *configReader
}

// NewExporter creates a new Exporter with injected configuration.
//
// Parameters:
//   - config: Configuration injected by CLI/GUI (from settings)
//
// Returns:
//   - *Exporter: Ready to use exporter instance
//
// Example:
//
//	config := buildConfigFromSettings()
//	exporter := profiles.NewExporter(config)
func NewExporter(config Config) *Exporter {
	return &Exporter{
		config: config,
		reader: newConfigReader(config),
	}
}

// Export exports profiles to JSON backup.
//
// Reads AWS config and exports selected sections. Use granular
// Include* flags to control exactly what gets exported.
//
// Common Scenarios:
//   - Team config: IncludeManaged=true
//   - Full backup: All three =true
//   - Personal only: IncludeAbove=true, IncludeBelow=true
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Export options
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - *ExportResult: Statistics and schema
//   - error: Any error during export (including cancellation)
//
// Example:
//
//	result, err := exporter.Export(ctx, profiles.ExportOptions{
//	    OutputPath:     "/path/to/backup.json",
//	    IncludeManaged: true,
//	    IncludeAbove:   true,
//	    IncludeBelow:   true,
//	}, reporter)
func (e *Exporter) Export(ctx context.Context, opts ExportOptions, reporter task.Reporter) (*ExportResult, error) {
	startTime := time.Now()
	logging.Debug.Log("Export started", "output", opts.OutputPath)

	// Validate output path
	if opts.OutputPath == "" {
		return nil, fmt.Errorf("output path is required")
	}

	// Check for cancellation before starting
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Export cancelled before start")
		return nil, err
	}

	// Read and parse config file (includes stats, no double processing!)
	s, stats, err := e.reader.readConfig(ctx, opts, reporter)
	if err != nil {
		// Check if cancelled
		if err == context.Canceled {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Stats already calculated during read - no double processing!
	logging.Debug.Log("Export counts",
		"managed", stats.ManagedProfiles,
		"unmanaged_above", stats.UnmanagedAbove,
		"unmanaged_below", stats.UnmanagedBelow,
		"total", stats.TotalProfiles,
		"sso_sessions", stats.SsoSessions,
	)

	reporter.ReportStatus(fmt.Sprintf("Writing backup file (%d profiles)...", stats.TotalProfiles))

	// Write to JSON file
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := writeFileContent(opts.OutputPath, string(jsonData)); err != nil {
		return nil, fmt.Errorf("failed to write output file: %w", err)
	}

	reporter.ReportStatus("Export complete")

	result := &ExportResult{
		TotalProfiles:   stats.TotalProfiles,
		ManagedProfiles: stats.ManagedProfiles,
		UnmanagedAbove:  stats.UnmanagedAbove,
		UnmanagedBelow:  stats.UnmanagedBelow,
		SsoSessions:     stats.SsoSessions,
		OutputPath:      opts.OutputPath,
		Schema:          s,
		Duration:        time.Since(startTime),
		Timestamp:       time.Now(),
	}

	logging.Debug.Log("Export completed",
		"output", opts.OutputPath,
		"total_profiles", stats.TotalProfiles,
		"duration", result.Duration,
	)

	return result, nil
}
