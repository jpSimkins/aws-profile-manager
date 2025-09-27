package profiles

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/task"
)

// Remover removes managed section from AWS config.
//
// This component removes the managed section while preserving personal
// profiles outside the markers. Optionally removes cheat sheet file.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	remover := profiles.NewRemover(config)
//	result, err := remover.Remove(ctx, opts, reporter)
type Remover struct {
	config Config
}

// NewRemover creates a new Remover with injected configuration.
//
// Parameters:
//   - config: Configuration injected by CLI/GUI (from settings)
//
// Returns:
//   - *Remover: Ready to use remover instance
//
// Example:
//
//	config := buildConfigFromSettings()
//	remover := profiles.NewRemover(config)
func NewRemover(config Config) *Remover {
	return &Remover{
		config: config,
	}
}

// Remove removes managed section from AWS config.
//
// Removes managed section while preserving personal profiles.
// Optionally removes cheat sheet file.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Removal options
//   - reporter: Progress reporter
//
// Returns:
//   - *RemoveResult: Indicates what was removed
//   - error: Any error during removal
//
// Example:
//
//	result, err := remover.Remove(ctx, profiles.RemoveOptions{
//	    RemoveCheatSheet: true,
//	}, reporter)
func (r *Remover) Remove(
	ctx context.Context,
	opts RemoveOptions,
	reporter task.Reporter,
) (*RemoveResult, error) {
	startTime := time.Now()
	logging.Debug.Log("Remove started")

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Remove cancelled before start")
		return nil, err
	}

	cheatSheetPath := opts.CheatSheetPath
	if cheatSheetPath == "" {
		cheatSheetPath = filepath.Join(r.config.CheatSheetOutputDir, "AWS_Profile_Cheat_Sheet.md")
	}

	result := &RemoveResult{
		ConfigPath:     r.config.ConfigPath,
		CheatSheetPath: cheatSheetPath,
		Duration:       time.Since(startTime),
	}

	// Dry run - just report what would be removed
	if opts.DryRun {
		reporter.ReportStatus("DRY RUN: Checking for managed profiles...")

		// Read config and detect markers
		lines, err := readConfigFile(r.config.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		markers := detectMarkers(lines, r.config.StartMarker, r.config.EndMarker)
		result.RemovedConfig = markers.Found

		if markers.Found {
			reporter.ReportStatus(fmt.Sprintf("DRY RUN: Would remove managed section (lines %d-%d)", markers.StartLine+1, markers.EndLine+1))
		} else {
			reporter.ReportStatus("DRY RUN: No managed section found")
		}

		if opts.RemoveCheatSheet {
			reporter.ReportStatus("DRY RUN: Would check for cheat sheet")
			result.RemovedCheatSheet = fileExists(cheatSheetPath)
		}

		logging.Debug.Log("DRY RUN completed")
		return result, nil
	}

	// Step 1: Remove managed profiles section
	reporter.ReportStatus("Removing managed profiles...")
	logging.Debug.Log("Removing managed profiles")

	removed, profileCount, err := r.removeManagedProfiles(ctx, reporter)
	if err != nil {
		return nil, fmt.Errorf("failed to remove managed profiles: %w", err)
	}

	result.RemovedConfig = removed
	result.ProfilesRemoved = profileCount

	logging.Debug.Log("Managed profiles removed", "removed", removed, "profiles", profileCount)

	// Check for cancellation before cheat sheet removal
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Remove cancelled before cheat sheet")
		return result, err // Return partial result
	}

	// Step 2: Remove cheat sheet (if requested)
	if opts.RemoveCheatSheet {
		reporter.ReportStatus("Removing cheat sheet...")
		logging.Debug.Log("Removing cheat sheet")

		removedCheatSheet, err := deleteFile(cheatSheetPath)
		if err != nil {
			logging.Log.Warn("Failed to remove cheat sheet", "error", err)
		} else {
			result.RemovedCheatSheet = removedCheatSheet
		}

		logging.Debug.Log("Cheat sheet removal attempted", "removed", removedCheatSheet)
	}

	reporter.ReportStatus("Removal complete")
	logging.Debug.Log("Remove completed",
		"removed_config", result.RemovedConfig,
		"removed_cheatsheet", result.RemovedCheatSheet,
		"profiles", result.ProfilesRemoved,
		"duration", result.Duration,
	)

	return result, nil
}

// removeManagedProfiles removes the managed section from AWS config.
//
// Returns:
//   - removed: true if markers found and removed
//   - profileCount: estimate of profiles removed
//   - error: any error during removal
func (r *Remover) removeManagedProfiles(ctx context.Context, reporter task.Reporter) (bool, int, error) {
	// Read config file
	reporter.ReportStatus("Reading AWS config file...")
	logging.Debug.Log("Reading config file")

	lines, err := readConfigFile(r.config.ConfigPath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read config file: %w", err)
	}

	// Detect markers
	reporter.ReportStatus("Detecting managed section markers...")
	logging.Debug.Log("Detecting markers")

	markers := detectMarkers(lines, r.config.StartMarker, r.config.EndMarker)

	if !markers.Found {
		reporter.ReportStatus("No managed section found")
		logging.Debug.Log("No markers found - nothing to remove")
		return false, 0, nil
	}

	logging.Debug.Log("Markers found",
		"start", markers.StartLine,
		"end", markers.EndLine,
	)

	// Check for cancellation before write
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("removeManagedProfiles cancelled before write")
		return false, 0, err
	}

	// Count profiles being removed (rough estimate based on lines)
	profileCount := markers.EndLine - markers.StartLine - 1 // Rough estimate

	// Build new config without managed section
	reporter.ReportStatus("Removing managed profiles...")
	logging.Debug.Log("Building config without managed section")

	var newLines []string

	// Add content before start marker
	if markers.StartLine > 0 {
		newLines = append(newLines, lines[:markers.StartLine]...)
	}

	// Remove any trailing blank lines
	for len(newLines) > 0 && newLines[len(newLines)-1] == "" {
		newLines = newLines[:len(newLines)-1]
	}

	// Add content after end marker
	if markers.EndLine+1 < len(lines) {
		// Add blank line separator if there's content below
		if len(lines[markers.EndLine+1:]) > 0 && lines[markers.EndLine+1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, lines[markers.EndLine+1:]...)
	}

	// Write updated config
	reporter.ReportStatus("Writing updated config file...")
	logging.Debug.Log("Writing config file", "lines", len(newLines))

	if err := writeConfigFile(r.config.ConfigPath, newLines); err != nil {
		return false, 0, fmt.Errorf("failed to write config file: %w", err)
	}

	reporter.ReportStatus("Managed profiles removed successfully")
	return true, profileCount, nil
}
