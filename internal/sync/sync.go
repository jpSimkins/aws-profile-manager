package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// Sync retrieves configuration from remote source with caching support.
//
// This is the main entry point for sync operations. It:
//  1. Validates configuration
//  2. Checks cache (unless force refresh)
//  3. Routes to appropriate fetcher based on strategy
//  4. Fetches configuration with progress reporting
//  5. Parses and validates schema
//  6. Updates cache
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - cfg: Sync configuration (strategy, URLs, credentials)
//   - opts: Sync options (force refresh, etc.)
//   - reporter: Progress reporter (use task.NoOpReporter{} for silent)
//
// Returns:
//   - *Result: Sync result with data and metadata
//   - error: Any error encountered
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	cfg := ConfigFromSettings(&settings.Get().Sync)
//	result, err := Sync(ctx, cfg, Options{}, task.CliReporter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use result.Data
func Sync(ctx context.Context, cfg SyncConfig, opts Options, reporter task.Reporter) (*Result, error) {
	start := time.Now()

	// Validate configuration
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize cache
	cache := NewCache(cfg.CacheTTL, cfg.CacheDir)

	// Try cache first (unless force refresh)
	if !opts.ForceRefresh {
		if entry, err := cache.Get(); err == nil && entry != nil {
			reporter.ReportStatus("Using cached configuration")
			return &Result{
				Data:      entry.Data,
				Source:    "cache",
				FetchTime: entry.FetchTime,
				CacheHit:  true,
				Duration:  time.Since(start),
			}, nil
		}
	}

	// Create fetcher based on strategy
	fetcher, err := createFetcher(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetcher: %w", err)
	}

	// Validate fetcher
	if err := fetcher.Validate(); err != nil {
		return nil, fmt.Errorf("invalid fetcher configuration: %w", err)
	}

	// Fetch configuration
	reporter.ReportStatus(fmt.Sprintf("Fetching configuration from %s", fetcher.String()))
	data, err := fetcher.Fetch(ctx, reporter)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	// Parse schema
	var configSchema schema.Schema
	if err := json.Unmarshal(data, &configSchema); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// ALWAYS validate schema before caching (security & correctness)
	reporter.ReportStatus("Validating configuration schema")
	if err := configSchema.Validate(); err != nil {
		return nil, fmt.Errorf("schema validation failed (refusing to cache invalid config): %w", err)
	}

	// Update cache
	fetchTime := time.Now()
	cacheEntry := &CacheEntry{
		Data:      &configSchema,
		FetchTime: fetchTime,
		Source:    fetcher.String(),
		Strategy:  cfg.Strategy,
	}
	if err := cache.Set(cacheEntry); err != nil {
		// Log but don't fail on cache write error
		reporter.ReportError(fmt.Errorf("failed to update cache: %w", err))
	}

	reporter.ReportStatus("Configuration sync complete")

	return &Result{
		Data:             &configSchema,
		Source:           fetcher.String(),
		FetchTime:        fetchTime,
		CacheHit:         false,
		Duration:         time.Since(start),
		BytesTransferred: int64(len(data)),
	}, nil
}

// ConfigFromSettings creates SyncConfig from settings.Sync.
//
// This helper converts application settings to sync configuration,
// providing defaults where needed. It automatically retrieves the
// cache directory from settings and injects it into the configuration,
// following the dependency injection pattern.
//
// Parameters:
//   - syncSettings: Settings from settings.Get().Sync
//
// Returns:
//   - SyncConfig: Configuration ready for Sync()
func ConfigFromSettings(syncSettings *settings.SyncSettings) SyncConfig {
	// Get cache directory from settings
	cacheDir, _ := settings.GetCacheDir() // Ignore error, will be validated later

	cfg := SyncConfig{
		Strategy: Strategy(syncSettings.Strategy),
		CacheTTL: 15 * time.Minute, // Default 15 minute cache
		CacheDir: cacheDir,

		// HTTP settings
		HTTPUrl:        syncSettings.HTTP.URL,
		HTTPHeaders:    make(map[string]string),
		HTTPTimeout:    30 * time.Second,
		HTTPRetries:    3,
		HTTPRetryDelay: 2 * time.Second,
		HTTPTLSVerify:  true,

		// S3 settings
		S3Bucket:  syncSettings.S3.Bucket,
		S3Key:     syncSettings.S3.Key,
		S3Region:  syncSettings.S3.Region,
		S3Profile: syncSettings.S3.Profile,
		S3AWSPath: "aws",

		// Local settings
		LocalPath: syncSettings.Local.Path,

		// Git settings
		GitRepoURL:  syncSettings.Git.RepoURL,
		GitBranch:   syncSettings.Git.Branch,
		GitFilePath: syncSettings.Git.FilePath,
		GitWorkDir:  syncSettings.Git.WorkDir,
		GitPath:     "git",
	}

	// Add custom headers if any
	if len(syncSettings.HTTP.Headers) > 0 {
		for k, v := range syncSettings.HTTP.Headers {
			cfg.HTTPHeaders[k] = v
		}
	}

	return cfg
}

// createFetcher creates the appropriate fetcher based on strategy.
func createFetcher(cfg SyncConfig) (Fetcher, error) {
	switch cfg.Strategy {
	case StrategyHTTP:
		return NewHttpFetcher(
			cfg.HTTPUrl,
			cfg.HTTPHeaders,
			cfg.HTTPTimeout,
			cfg.HTTPRetries,
			cfg.HTTPRetryDelay,
			cfg.HTTPTLSVerify,
			cfg.HTTPBypassSSRF,
			cfg.HTTPBypassTLS,
		), nil

	case StrategyS3:
		return NewS3Fetcher(
			cfg.S3Bucket,
			cfg.S3Key,
			cfg.S3Region,
			cfg.S3Profile,
			cfg.S3AWSPath,
			cfg.S3AWSEnv,
		), nil

	case StrategyLocal:
		return NewLocalFetcher(cfg.LocalPath), nil

	case StrategyGit:
		return NewGitFetcher(
			cfg.GitRepoURL,
			cfg.GitBranch,
			cfg.GitFilePath,
			cfg.GitWorkDir,
			cfg.GitPath,
			cfg.GitEnv,
		), nil

	default:
		return nil, fmt.Errorf("unknown strategy: %s", cfg.Strategy)
	}
}

// ClearCache removes the cached configuration.
//
// Parameters:
//   - cfg: Sync configuration (CacheTTL and CacheDir are used)
//
// Returns:
//   - error: Any error encountered
func ClearCache(cfg SyncConfig) error {
	cache := NewCache(cfg.CacheTTL, cfg.CacheDir)
	return cache.Clear()
}
