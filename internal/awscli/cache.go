package awscli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"aws-profile-manager/internal/logging"
)

const (
	// CacheFileName is the filename used for storing cached AWS CLI profile data.
	//
	// The cache file is stored in JSON format and contains both profile data
	// and metadata for cache invalidation based on config file modifications.
	CacheFileName = "aws-cli-profiles-cache.json"
)

// Cache handles caching of AWS CLI profile data with intelligent file modification tracking.
//
// The cache system improves performance by avoiding repeated parsing of AWS CLI config files.
// It automatically invalidates cached data when the source config file is modified,
// ensuring users always get current data while benefiting from caching.
//
// Cache Invalidation Strategy:
//   - Tracks source file modification time (mtime)
//   - Compares current mtime with cached mtime
//   - Reloads automatically when source file changes
//   - Falls back to fresh extraction if cache is invalid or missing
//
// Use Cases:
//   - Repeated profile listings without unnecessary file parsing
//   - GUI applications that query profiles frequently
//   - Improved responsiveness for interactive tools
type Cache struct {
	extractor *Extractor // Extractor for reading AWS CLI config
	cacheDir  string     // Directory for storing cache files
}

// NewCache creates a new cache instance with default extractor and specified cache directory.
//
// The cache uses the default AWS CLI config path (~/.aws/config) and stores cache
// files in the specified directory.
//
// Parameters:
//   - cacheDir: Directory where cache files will be stored (must be writable)
//
// Returns:
//   - *Cache: Configured cache instance ready to use
//
// Example:
//
//	cache := awscli.NewCache("/home/user/.config/aws-profile-manager")
//	data, err := cache.GetData()  // Uses cache if valid
func NewCache(cacheDir string) *Cache {
	return &Cache{
		extractor: NewExtractor(),
		cacheDir:  cacheDir,
	}
}

// NewCacheWithExtractor creates a new cache instance with custom extractor and cache directory.
//
// This allows you to specify both a custom AWS config file location and cache directory,
// useful for testing or processing multiple config files.
//
// Parameters:
//   - extractor: Custom extractor configured with specific config path or test data
//   - cacheDir: Directory where cache files will be stored (must be writable)
//
// Returns:
//   - *Cache: Configured cache instance ready to use
//
// Example:
//
//	extractor := awscli.NewExtractorWithPath("/custom/aws/config")
//	cache := awscli.NewCacheWithExtractor(extractor, "/tmp/cache")
//	data, err := cache.GetData()
func NewCacheWithExtractor(extractor *Extractor, cacheDir string) *Cache {
	return &Cache{
		extractor: extractor,
		cacheDir:  cacheDir,
	}
}

// GetData returns AWS CLI profile data, using cache when valid.
//
// This is the main method for retrieving profile data with intelligent caching.
// It automatically determines whether to use cached data or extract fresh data
// based on file modification times.
//
// Process:
//  1. Check if source config file has been modified since last cache
//  2. If cache is valid: Return cached data
//  3. If cache is invalid: Extract fresh data and update cache
//  4. If cache load fails: Extract fresh data as fallback
//
// Cache Invalidation:
//   - Source file modified: Cache is automatically invalidated
//   - Cache file missing: Fresh data extracted
//   - Cache load error: Fresh data extracted as fallback
//
// Returns:
//   - *ExtractedData: Profile data (from cache or fresh extraction)
//   - error: Any error encountered during extraction (cache errors are logged but don't fail the request)
func (c *Cache) GetData() (*ExtractedData, error) {
	logging.Debug.Log("Getting AWS CLI profile data with caching")

	// Check if we need to reload based on file modification time
	needsReload, err := c.needsReload()
	if err != nil {
		logging.Log.Warnf("Error checking cache validity, will extract fresh data: %v", err)
		needsReload = true
	}

	if !needsReload {
		// Try to load from cache
		if cachedData, err := c.loadFromCache(); err == nil {
			logging.Log.Info("Using cached AWS CLI profile data",
				"profiles", len(cachedData.Profiles),
				"sso_sessions", len(cachedData.SsoSessions))
			return cachedData, nil
		}
		logging.Log.Warnf("Failed to load from cache, will extract fresh data: %v", err)
	}

	// Extract fresh data
	data, err := c.extractor.ExtractFromFile()
	if err != nil {
		return nil, err
	}

	// Cache the fresh data
	if err := c.saveToCache(data); err != nil {
		logging.Log.Warnf("Failed to save AWS CLI data to cache: %v", err)
		// Don't fail the request just because caching failed
	}

	return data, nil
}

// ForceReload forces extraction of fresh AWS CLI data, bypassing cache.
//
// This method ignores any cached data and always extracts fresh profile information
// from the AWS CLI config file. The new data is saved to cache for subsequent requests.
//
// Use Cases:
//   - User explicitly requests refresh
//   - After making manual changes to AWS config
//   - Debugging cache-related issues
//
// Returns:
//   - *ExtractedData: Freshly extracted profile data
//   - error: Any error encountered during extraction
func (c *Cache) ForceReload() (*ExtractedData, error) {
	logging.Log.Info("Force reloading AWS CLI profile data")

	data, err := c.extractor.ExtractFromFile()
	if err != nil {
		return nil, err
	}

	// Update the cache with fresh data
	if err := c.saveToCache(data); err != nil {
		logging.Log.Warnf("Failed to save fresh AWS CLI data to cache: %v", err)
	}

	return data, nil
}

// ClearCache removes the cached AWS CLI profile data from disk.
//
// This method deletes the cache file, forcing subsequent GetData() calls to
// extract fresh data. Useful for cleanup operations or troubleshooting.
//
// Returns:
//   - error: Any error encountered during file deletion (returns nil if cache file doesn't exist)
func (c *Cache) ClearCache() error {
	cachePath, err := c.getCachePath()
	if err != nil {
		return err
	}

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			logging.Debug.Log("No AWS CLI cache to clear")
			return nil
		}
		return logging.Log.ErrorWithDetails("Failed to remove AWS CLI cache file", err)
	}

	logging.Log.Success("AWS CLI cache cleared successfully")
	return nil
}

// HasCache checks if a valid cache file exists on disk.
//
// This method only checks for file existence, not cache validity. Use GetData()
// for automatic cache validation based on modification times.
//
// Returns:
//   - bool: true if cache file exists, false otherwise
func (c *Cache) HasCache() bool {
	cachePath, err := c.getCachePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(cachePath)
	return err == nil
}

// needsReload determines if cache needs to be invalidated based on file modification times.
//
// This method implements the cache invalidation logic by comparing modification times
// of the source AWS CLI config file with the cached modification time.
//
// Invalidation Logic:
//   - Source config doesn't exist: Needs reload
//   - Cache doesn't exist: Needs reload
//   - Cache file corrupted: Needs reload
//   - Source modified after cache: Needs reload
//   - Otherwise: Cache is valid
//
// Returns:
//   - bool: true if reload is needed, false if cache is valid
//   - error: Any error encountered checking files (treated as needs reload)
func (c *Cache) needsReload() (bool, error) {
	// Check if config file exists
	configModTime, err := c.extractor.GetFileModTime()
	if err != nil {
		return true, err // Config file doesn't exist or can't be read
	}

	// Check if cache exists
	cachePath, err := c.getCachePath()
	if err != nil {
		return true, err
	}

	cacheInfo, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug.Log("No cache file exists, reload needed")
			return true, nil // Cache doesn't exist
		}
		return true, err // Can't read cache file
	}

	// Load cache metadata to check source file modification time
	cachedData, err := c.loadCacheMetadata()
	if err != nil {
		logging.Debug.Logf("Failed to load cache metadata, reload needed: %v", err)
		return true, nil // Can't read cache metadata
	}

	// Compare modification times
	if configModTime.After(cachedData.LastModified) {
		logging.Debug.Logf("Config file is newer than cache, reload needed (config: %v, cache: %v)",
			configModTime, cachedData.LastModified)
		return true, nil
	}

	// Check if cache is older than a reasonable threshold (24 hours)
	maxCacheAge := 24 * time.Hour
	if time.Since(cacheInfo.ModTime()) > maxCacheAge {
		logging.Debug.Logf("Cache is too old, reload needed (cache age: %v)", time.Since(cacheInfo.ModTime()))
		return true, nil
	}

	logging.Debug.Log("Cache is valid, no reload needed")
	return false, nil
}

// loadFromCache loads AWS CLI profile data from the cache file.
//
// This method reads and deserializes the cache file, extracting the profile data
// without metadata. Use loadCacheMetadata() if you need full cache metadata.
//
// Returns:
//   - *ExtractedData: Cached profile data
//   - error: Any error encountered reading or parsing cache file
func (c *Cache) loadFromCache() (*ExtractedData, error) {
	cachePath, err := c.getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, logging.Log.ErrorWithDetails("Failed to parse AWS CLI cache file", err)
	}

	logging.Debug.Log("Loaded AWS CLI data from cache",
		"profiles", len(cacheData.Data.Profiles),
		"sso_sessions", len(cacheData.Data.SsoSessions),
		"cached_at", cacheData.CachedAt)

	return &cacheData.Data, nil
}

// loadCacheMetadata loads cache metadata for modification time checking.
//
// This method reads only the metadata portion of the cache file, which includes
// file modification times used for cache validation. More efficient than loading
// full cache data when only metadata is needed.
//
// Returns:
//   - *CacheData: Complete cache data including metadata
//   - error: Any error encountered reading or parsing cache file
func (c *Cache) loadCacheMetadata() (*CacheData, error) {
	cachePath, err := c.getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var cacheData CacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, err
	}

	return &cacheData, nil
}

// saveToCache saves AWS CLI profile data to the cache file with metadata.
//
// This method serializes profile data along with metadata (modification times, source path)
// and writes it to the cache file. The metadata enables intelligent cache invalidation.
//
// Cache Structure:
//   - Data: Complete ExtractedData with profiles and sessions
//   - LastModified: Config file modification time (for invalidation)
//   - CachedAt: Cache creation timestamp
//   - SourcePath: Path to source config file
//
// Parameters:
//   - data: Profile data to cache
//
// Returns:
//   - error: Any error encountered creating cache directory or writing file
func (c *Cache) saveToCache(data *ExtractedData) error {
	logging.Debug.Log("Saving AWS CLI data to cache")

	// Get the current modification time of the config file
	configModTime, err := c.extractor.GetFileModTime()
	if err != nil {
		return err
	}

	cacheData := CacheData{
		Data:         *data,
		LastModified: configModTime,
		CachedAt:     time.Now(),
		SourcePath:   c.extractor.GetConfigPath(),
	}

	jsonData, err := json.MarshalIndent(cacheData, "", "  ")
	if err != nil {
		return logging.Log.ErrorWithDetails("Failed to marshal AWS CLI cache data", err)
	}

	cachePath, err := c.getCachePath()
	if err != nil {
		return err
	}

	// Ensure the config directory exists
	configDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return logging.Log.ErrorWithDetails("Failed to create config directory for AWS CLI cache", err)
	}

	if err := os.WriteFile(cachePath, jsonData, 0600); err != nil {
		return logging.Log.ErrorWithDetails("Failed to write AWS CLI cache file", err)
	}

	logging.Log.Success("AWS CLI data cached successfully",
		"profiles", len(data.Profiles),
		"sso_sessions", len(data.SsoSessions),
		"cache_file", cachePath,
		"size", len(jsonData))

	return nil
}

// getCachePath returns the full path to the cache file.
//
// This method constructs the cache file path from the configured cache directory
// and validates that the cache directory is configured.
//
// Returns:
//   - string: Full path to cache file
//   - error: Error if cache directory is not configured
func (c *Cache) getCachePath() (string, error) {
	if c.cacheDir == "" {
		return "", logging.Log.Error("Cache directory not configured")
	}

	cachePath := filepath.Join(c.cacheDir, CacheFileName)
	logging.Debug.Logf("AWS CLI cache path determined: %s", cachePath)

	return cachePath, nil
}
