// Package sync provides centralized configuration synchronization from remote sources.
//
// The sync package supports multiple fetch strategies (HTTP, S3, Git, local) with
// security features, caching, and progress reporting. All fetching operations
// use the task package for subprocess execution and cancellation support.
//
// Key features:
//   - Multiple fetch strategies (HTTP, S3, Git, local)
//   - Security validation (SSRF protection, TLS verification)
//   - TTL-aware caching with automatic refresh
//   - Progress reporting for GUI/CLI
//   - Context cancellation support
//   - Dependency injection pattern (no direct settings usage)
//
// Example usage:
//
//	cfg := sync.ConfigFromSettings(&settings.Get().Sync)
//	result, err := sync.Sync(ctx, cfg, sync.Options{}, task.CliReporter{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use result.Data
package sync

import (
	"time"

	"aws-profile-manager/internal/schema"
)

// Strategy represents the sync strategy type.
type Strategy string

const (
	// StrategyHTTP fetches configuration via HTTP/HTTPS.
	StrategyHTTP Strategy = "http"

	// StrategyS3 fetches configuration from AWS S3.
	StrategyS3 Strategy = "s3"

	// StrategyLocal reads configuration from local file system.
	StrategyLocal Strategy = "local"

	// StrategyGit fetches configuration from Git repository.
	StrategyGit Strategy = "git"
)

// SyncConfig contains all configuration needed for syncing.
//
// This struct is populated from settings.Sync and passed to the
// Sync() function. It contains strategy-specific configuration
// and security settings.
type SyncConfig struct {
	// Strategy determines which fetcher to use
	Strategy Strategy

	// HTTP Strategy Configuration
	HTTPUrl        string            // HTTP/HTTPS URL
	HTTPHeaders    map[string]string // Custom HTTP headers
	HTTPTimeout    time.Duration     // Request timeout
	HTTPRetries    int               // Number of retries
	HTTPRetryDelay time.Duration     // Delay between retries
	HTTPBypassSSRF bool              // Bypass SSRF protection (dangerous!)
	HTTPBypassTLS  bool              // Bypass TLS verification (dangerous!)

	// S3 Strategy Configuration
	S3Bucket  string            // S3 bucket name
	S3Key     string            // S3 object key
	S3Region  string            // AWS region
	S3Profile string            // AWS profile to use
	S3AWSPath string            // Path to AWS CLI (default: "aws")
	S3AWSEnv  map[string]string // Environment for AWS CLI via task package

	// Local Strategy Configuration
	LocalPath string // Local file path

	// Git Strategy Configuration
	GitRepoURL  string            // Git repository URL (SSH or HTTPS)
	GitBranch   string            // Branch or ref to checkout (e.g., "main", "v1.0.0")
	GitFilePath string            // Path to config file within repo (e.g., "config.json")
	GitWorkDir  string            // Working directory for git operations (empty for temp)
	GitPath     string            // Path to git executable (default: "git")
	GitEnv      map[string]string // Environment for git commands via task package

	// Cache Configuration
	CacheTTL time.Duration // Cache time-to-live (0 = disabled)
	CacheDir string        // Directory for cache files
}

// Options contains optional parameters for sync operations.
type Options struct {
	// ForceRefresh bypasses cache and fetches fresh data
	ForceRefresh bool

	// Note: Schema validation is ALWAYS performed before caching
	// to ensure only valid configurations are cached
}

// Result contains the outcome of a sync operation.
type Result struct {
	// Data is the fetched configuration schema
	Data *schema.Schema

	// Source indicates where the data came from
	Source string // "cache", "http", "s3", "local"

	// FetchTime is when the data was fetched
	FetchTime time.Time

	// CacheHit indicates if data was served from cache
	CacheHit bool

	// Duration is how long the operation took
	Duration time.Duration

	// BytesTransferred is the size of fetched data
	BytesTransferred int64
}
