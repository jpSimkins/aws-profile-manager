package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aws-profile-manager/internal/schema"
)

// CacheEntry represents a cached configuration with metadata.
type CacheEntry struct {
	// Data is the cached schema
	Data *schema.Schema

	// FetchTime is when the data was originally fetched
	FetchTime time.Time

	// Source indicates where the data came from
	Source string

	// Strategy is the sync strategy used
	Strategy Strategy
}

// Cache provides TTL-aware caching for sync operations.
type Cache struct {
	ttl      time.Duration
	cacheDir string // Directory where cache file is stored
}

// NewCache creates a new cache with the specified TTL and cache directory.
//
// Parameters:
//   - ttl: Time-to-live for cache entries (0 = disabled)
//   - cacheDir: Directory where cache file will be stored
//
// Returns:
//   - *Cache: New cache instance
func NewCache(ttl time.Duration, cacheDir string) *Cache {
	return &Cache{
		ttl:      ttl,
		cacheDir: cacheDir,
	}
}

// Get retrieves a cached entry if it exists and is valid.
//
// Returns:
//   - *CacheEntry: Cached entry if valid, nil otherwise
//   - error: Any error encountered
func (c *Cache) Get() (*CacheEntry, error) {
	// Cache disabled
	if c.ttl == 0 {
		return nil, nil
	}

	// Get cache file path
	cacheFile, err := c.getCacheFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache path: %w", err)
	}

	// Check if cache file exists
	info, err := os.Stat(cacheFile)
	if os.IsNotExist(err) {
		return nil, nil // No cache
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat cache file: %w", err)
	}

	// Check if cache is expired
	age := time.Since(info.ModTime())
	if age > c.ttl {
		return nil, nil // Expired
	}

	// Read cache file
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache entry
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &entry, nil
}

// Set stores a cache entry.
//
// Parameters:
//   - entry: Cache entry to store
//
// Returns:
//   - error: Any error encountered
func (c *Cache) Set(entry *CacheEntry) error {
	// Cache disabled
	if c.ttl == 0 {
		return nil
	}

	// Get cache file path
	cacheFile, err := c.getCacheFilePath()
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	// Create cache directory if needed
	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal cache entry
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	// Write cache file
	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes the cache file.
//
// Returns:
//   - error: Any error encountered (nil if file doesn't exist)
func (c *Cache) Clear() error {
	cacheFile, err := c.getCacheFilePath()
	if err != nil {
		return fmt.Errorf("failed to get cache path: %w", err)
	}

	// Remove cache file (ignore error if it doesn't exist)
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	return nil
}

// getCacheFilePath returns the path to the cache file.
func (c *Cache) getCacheFilePath() (string, error) {
	if c.cacheDir == "" {
		return "", fmt.Errorf("cache directory not configured")
	}

	return filepath.Join(c.cacheDir, "sync-config.json"), nil
}
