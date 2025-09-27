package viewmodels

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	syncpkg "aws-profile-manager/internal/sync"
	"aws-profile-manager/internal/task"
)

// InstallViewModel manages all business logic for the Install view.
//
// Owns schema loading, preset selection, filter state, and installation execution.
// The view layer only handles UI wiring and rendering.
type InstallViewModel struct {
	IsLoading      bool
	selectedPreset string
	mu             sync.RWMutex
}

// NewInstallViewModel creates and registers an install view model.
func NewInstallViewModel() *InstallViewModel {
	logging.Debug.Log("\t🔹 Creating install view model")

	vm := &InstallViewModel{
		IsLoading: false,
	}
	core.App.RegisterState("install-view", vm)

	logging.Debug.Log("\t🔹 Install view model created")
	return vm
}

// GetSyncCachePath returns the sync cache file path.
func (vm *InstallViewModel) GetSyncCachePath() (string, error) {
	cacheDir, err := settings.GetCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "sync-config.json"), nil
}

// CacheExists checks if the sync cache file exists.
func (vm *InstallViewModel) CacheExists() bool {
	cachePath, err := vm.GetSyncCachePath()
	if err != nil {
		return false
	}
	_, err = os.Stat(cachePath)
	return err == nil
}

// FormatLoadedAt formats a timestamp for display.
func (vm *InstallViewModel) FormatLoadedAt(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// EmptyDisplaySchema returns an empty schema for initial display.
func (vm *InstallViewModel) EmptyDisplaySchema() *schema.Schema {
	return &schema.Schema{
		Version: "1.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{},
		},
	}
}

// LoadDisplaySchema loads schema from sync cache.
//
// This loads from the sync cache file (not AWS config). The sync cache contains
// the centralized configuration fetched from the remote source.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Task reporter for progress updates
//
// Returns: (displaySchema, sourcePath, error)
func (vm *InstallViewModel) LoadDisplaySchema(ctx context.Context, reporter task.Reporter) (*schema.Schema, string, error) {
	cachePath, err := vm.GetSyncCachePath()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get cache path: %w", err)
	}

	var displaySchema *schema.Schema

	logging.Debug.Log("Install view model: Loading schema from sync cache",
		"cachePath", cachePath,
	)

	loadTask := &task.FunctionTask{
		Name: "load-install-schema",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			runReporter.ReportStatus("Reading sync cache")

			// Read cache file
			data, err := os.ReadFile(cachePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read cache file: %w", err)
			}

			// Parse cache entry
			var cacheEntry syncpkg.CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				return nil, fmt.Errorf("failed to parse cache file: %w", err)
			}

			if cacheEntry.Data == nil {
				return nil, fmt.Errorf("cache entry has no schema data")
			}

			displaySchema = cacheEntry.Data

			runReporter.ReportStatus("Schema loaded")
			return []byte("ok"), nil
		},
	}

	if _, err := task.Run(ctx, loadTask, reporter); err != nil {
		logging.Debug.Log("Install view model: Schema load failed",
			"cachePath", cachePath,
			"error", err,
		)
		return nil, cachePath, err
	}

	logging.Debug.Log("Install view model: Schema loaded successfully",
		"cachePath", cachePath,
		"hasPresets", displaySchema.Presets != nil,
		"presetCount", len(displaySchema.Presets),
	)
	return displaySchema, cachePath, nil
}

// LoadSchemaFromFile loads schema from a user-selected JSON file.
//
// Parameters:
//   - ctx: Context for cancellation
//   - filePath: Path to JSON file containing schema
//   - reporter: Task reporter for progress updates
//
// Returns: (displaySchema, sourcePath, error)
func (vm *InstallViewModel) LoadSchemaFromFile(ctx context.Context, filePath string, reporter task.Reporter) (*schema.Schema, string, error) {
	var displaySchema *schema.Schema

	logging.Debug.Log("Install view model: Loading schema from file",
		"filePath", filePath,
	)

	loadTask := &task.FunctionTask{
		Name: "load-schema-from-file",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			runReporter.ReportStatus("Reading file")

			// Read file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}

			// Try parsing as cache entry first (with Data wrapper)
			var cacheEntry syncpkg.CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err == nil && cacheEntry.Data != nil {
				displaySchema = cacheEntry.Data
				runReporter.ReportStatus("Schema loaded (cache format)")
				return []byte("ok"), nil
			}

			// Try parsing as direct schema (without Data wrapper)
			var directSchema schema.Schema
			if err := json.Unmarshal(data, &directSchema); err != nil {
				return nil, fmt.Errorf("failed to parse file as schema: %w", err)
			}

			displaySchema = &directSchema
			runReporter.ReportStatus("Schema loaded (direct format)")
			return []byte("ok"), nil
		},
	}

	if _, err := task.Run(ctx, loadTask, reporter); err != nil {
		logging.Debug.Log("Install view model: Schema load from file failed",
			"filePath", filePath,
			"error", err,
		)
		return nil, filePath, err
	}

	logging.Debug.Log("Install view model: Schema loaded from file successfully",
		"filePath", filePath,
		"hasPresets", displaySchema.Presets != nil,
		"presetCount", len(displaySchema.Presets),
	)
	return displaySchema, filePath, nil
}

// StartLoad loads schema asynchronously with callback.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Task reporter for progress updates
//   - onLoadComplete: Callback receives (displaySchema, configPath, error, loadedAt)
func (vm *InstallViewModel) StartLoad(
	ctx context.Context,
	reporter task.Reporter,
	onLoadComplete func(*schema.Schema, string, error, time.Time),
) {
	logging.Debug.Log("Install view model: StartLoad triggered")

	vm.mu.Lock()
	vm.IsLoading = true
	vm.mu.Unlock()

	var displaySchema *schema.Schema
	var configPath string

	asyncTask := &task.FunctionTask{
		Name: "load-install-schema-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			var err error
			displaySchema, configPath, err = vm.LoadDisplaySchema(runCtx, runReporter)
			if err != nil {
				return nil, err
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		vm.mu.Lock()
		vm.IsLoading = false
		vm.mu.Unlock()

		logging.Debug.Log("Install view model: StartLoad complete",
			"error", err,
		)

		if onLoadComplete != nil {
			onLoadComplete(displaySchema, configPath, err, time.Now())
		}
	})
}

// GetPresets returns presets from the loaded schema.
//
// This is a convenience method for the view to access presets.
// Returns nil if no presets are available.
func (vm *InstallViewModel) GetPresets(displaySchema *schema.Schema) map[string]*schema.Preset {
	if displaySchema == nil {
		return nil
	}
	return displaySchema.Presets
}

// SelectPreset stores the currently selected preset key.
//
// The view can use this to highlight the active preset selection.
func (vm *InstallViewModel) SelectPreset(presetKey string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.selectedPreset = presetKey
}

// GetSelectedPreset returns the currently selected preset key.
func (vm *InstallViewModel) GetSelectedPreset() string {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.selectedPreset
}

// Install executes profile installation using profiles.Installer.
//
// This method:
//  1. Installs filtered profiles using profiles.Installer
//  2. Updates sync cache with original (full) schema for future use
//  3. Updates sync settings to point to local cache
//
// Parameters:
//   - ctx: Context for cancellation
//   - originalSchema: Full unfiltered schema (saved to cache for future loads)
//   - filteredSchema: Filtered schema (what gets installed)
//   - opts: Installation options (filters, paths, flags)
//   - reporter: Task reporter for progress updates
//
// Returns: (result, error)
func (vm *InstallViewModel) Install(
	ctx context.Context,
	originalSchema *schema.Schema,
	filteredSchema *schema.Schema,
	opts profiles.InstallOptions,
	reporter task.Reporter,
) (*profiles.InstallResult, error) {
	logging.Debug.Log("Install view model: Starting installation",
		"hasOriginalSchema", originalSchema != nil,
		"hasFilteredSchema", filteredSchema != nil,
		"dryRun", opts.DryRun,
		"generateCheatSheet", opts.GenerateCheatSheet,
	)

	// Ensure filtered schema is set in options (this is what gets installed)
	opts.Schema = filteredSchema

	// Build config from settings, then apply any UI overrides from opts.
	currentSettings := settings.Get()
	appSettings := currentSettings.Application
	cacheDir, _ := settings.GetCacheDir()

	configPath := filepath.Join(settings.GetAwsDir(), "config")
	if opts.OutputFilePath != "" {
		configPath = opts.OutputFilePath
	}

	config := profiles.Config{
		ConfigPath:          configPath,
		CheatSheetOutputDir: settings.GetDesktopDir(),
		CacheDir:            cacheDir,
		AwsDir:              settings.GetAwsDir(),
		StartMarker:         appSettings.GetFormattedStartMarker(),
		EndMarker:           appSettings.GetFormattedEndMarker(),
		IncludeTimestamp:    appSettings.IncludeTimestamp,
		IncludeVersion:      appSettings.IncludeVersion,
	}

	// Create installer and execute
	installer := profiles.NewInstaller(config)
	result, err := installer.Install(ctx, opts, reporter)

	if err != nil {
		logging.Debug.Log("Install view model: Installation failed",
			"error", err,
		)
		return nil, err
	}

	logging.Debug.Log("Install view model: Installation complete",
		"totalProfiles", result.TotalProfiles,
		"cheatSheetPath", result.CheatSheetPath,
	)

	// Save ORIGINAL (unfiltered) schema to sync cache for future loads
	// This ensures all filter options remain available after a filtered install
	if !opts.DryRun && originalSchema != nil {
		cacheDir, cacheErr := settings.GetCacheDir()
		if cacheErr != nil {
			logging.Log.Warn("Failed to get cache dir for sync cache update", "error", cacheErr)
		} else {
			cache := syncpkg.NewCache(24*time.Hour, cacheDir)
			cacheEntry := &syncpkg.CacheEntry{
				Data:      originalSchema,
				FetchTime: time.Now(),
				Source:    "install",
				Strategy:  syncpkg.StrategyLocal,
			}
			if writeErr := cache.Set(cacheEntry); writeErr != nil {
				logging.Log.Warn("Failed to update sync cache", "error", writeErr)
			} else {
				logging.Debug.Log("Install view model: Updated sync cache", "cacheDir", cacheDir)

				// Update sync settings to use local strategy pointing to cache dir
				currentSettings.Sync.Strategy = string(syncpkg.StrategyLocal)
				currentSettings.Sync.Local.Path = cacheDir
				if setErr := settings.Set(currentSettings); setErr != nil {
					logging.Log.Warn("Failed to update sync settings", "error", setErr)
				} else {
					logging.Debug.Log("Install view model: Updated sync settings to use local cache")
				}
			}
		}
	}

	return result, nil
}

// StartInstall executes installation asynchronously with callback.
//
// Parameters:
//   - ctx: Context for cancellation
//   - originalSchema: Full unfiltered schema (saved to cache)
//   - filteredSchema: Filtered schema (what gets installed)
//   - opts: Installation options (filters, paths, flags)
//   - reporter: Task reporter for progress updates
//   - onInstallComplete: Callback receives (result, error)
func (vm *InstallViewModel) StartInstall(
	ctx context.Context,
	originalSchema *schema.Schema,
	filteredSchema *schema.Schema,
	opts profiles.InstallOptions,
	reporter task.Reporter,
	onInstallComplete func(*profiles.InstallResult, error),
) {
	logging.Debug.Log("Install view model: StartInstall triggered")

	var installResult *profiles.InstallResult

	asyncTask := &task.FunctionTask{
		Name: "install-profiles-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			var err error
			installResult, err = vm.Install(runCtx, originalSchema, filteredSchema, opts, runReporter)
			if err != nil {
				return nil, err
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		logging.Debug.Log("Install view model: StartInstall complete",
			"error", err,
		)

		if onInstallComplete != nil {
			onInstallComplete(installResult, err)
		}
	})
}

// FormatInstallResult returns a styled widget summarising the installation result.
//
// Uses the same formatStatsSection helper as the import view for consistent formatting.
// Returns nil when no profiles were installed (e.g. all already existed).
func (vm *InstallViewModel) FormatInstallResult(result *profiles.InstallResult) fyne.CanvasObject {
	return formatStatsSection(result.ManagedStats, result.TotalProfiles, "📦 Managed Data")
}

// FormatInstallStatusHeader returns the markdown heading shown above the detail
// message in the post-installation dialog.
//
// The heading communicates the overall outcome — preview, no-op, or success —
// so the view only needs to call this rather than inspect the result itself.
func (vm *InstallViewModel) FormatInstallStatusHeader(result *profiles.InstallResult, isDryRun bool) string {
	if isDryRun {
		return "## 🔍 Preview Complete!\n\n**⚠️ No changes were made.**"
	}
	if result.TotalProfiles == 0 {
		return "## 🛈 No Changes Needed\n\n**All profiles already exist.**"
	}
	return "## ✅ Installation Successful!"
}

// BuildInstallOptions constructs a profiles.InstallOptions from UI-supplied values.
//
// Centralising this here keeps the view free of direct profiles package imports
// and makes the mapping between UI state and install behaviour explicit.
func (vm *InstallViewModel) BuildInstallOptions(isDryRun, generateCheatSheet, cheatSheetOnly bool, outputFilePath, cheatSheetPath string) profiles.InstallOptions {
	return profiles.InstallOptions{
		DryRun:             isDryRun,
		GenerateCheatSheet: generateCheatSheet,
		CheatSheetOnly:     cheatSheetOnly,
		CheatSheetPath:     cheatSheetPath,
		OutputFilePath:     outputFilePath,
	}
}
