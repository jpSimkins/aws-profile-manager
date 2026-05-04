package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/test"
)

// TestCache_GetSet tests basic cache operations.
func TestCache_GetSet(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	cache := NewCache(10*time.Minute, cacheDir)

	// Initially empty
	entry, err := cache.Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if entry != nil {
		t.Error("Expected nil entry for empty cache")
	}

	// Set entry
	testEntry := &CacheEntry{
		Data: &schema.Schema{
			Version: "1.0",
		},
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  StrategyLocal,
	}

	if err := cache.Set(testEntry); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Get entry
	entry, err = cache.Get()
	if err != nil {
		t.Fatalf("Get() after Set() failed: %v", err)
	}
	if entry == nil {
		t.Fatal("Expected entry after Set()")
	}
	if entry.Data.Version != "1.0" {
		t.Errorf("Expected version 1.0, got: %s", entry.Data.Version)
	}
}

// TestCache_Expiration tests cache TTL expiration.
func TestCache_Expiration(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	// Very short TTL for testing
	cache := NewCache(100*time.Millisecond, cacheDir)

	// Set entry
	testEntry := &CacheEntry{
		Data: &schema.Schema{
			Version: "1.0",
		},
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  StrategyLocal,
	}

	if err := cache.Set(testEntry); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Should be available immediately
	entry, err := cache.Get()
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if entry == nil {
		t.Fatal("Expected entry immediately after Set()")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	entry, err = cache.Get()
	if err != nil {
		t.Fatalf("Get() after expiration failed: %v", err)
	}
	if entry != nil {
		t.Error("Expected nil entry after expiration")
	}
}

// TestCache_Disabled tests cache with TTL = 0.
func TestCache_Disabled(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	cache := NewCache(0, cacheDir) // Disabled

	// Set should be no-op
	testEntry := &CacheEntry{
		Data:      &schema.Schema{Version: "1.0"},
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  StrategyLocal,
	}

	if err := cache.Set(testEntry); err != nil {
		t.Fatalf("Set() on disabled cache failed: %v", err)
	}

	// Get should return nil
	entry, err := cache.Get()
	if err != nil {
		t.Fatalf("Get() on disabled cache failed: %v", err)
	}
	if entry != nil {
		t.Error("Expected nil entry for disabled cache")
	}
}

// TestCache_Clear tests cache clearing.
func TestCache_Clear(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	cache := NewCache(10*time.Minute, cacheDir)

	// Set entry
	testEntry := &CacheEntry{
		Data:      &schema.Schema{Version: "1.0"},
		FetchTime: time.Now(),
		Source:    "test",
		Strategy:  StrategyLocal,
	}

	if err := cache.Set(testEntry); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Verify it exists
	entry, err := cache.Get()
	if err != nil || entry == nil {
		t.Fatal("Expected entry after Set()")
	}

	// Clear cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	// Should be empty
	entry, err = cache.Get()
	if err != nil {
		t.Fatalf("Get() after Clear() failed: %v", err)
	}
	if entry != nil {
		t.Error("Expected nil entry after Clear()")
	}

	// Clearing again should not error
	if err := cache.Clear(); err != nil {
		t.Errorf("Clear() on empty cache failed: %v", err)
	}
}

// TestCache_CacheFilePath tests cache file location.
func TestCache_CacheFilePath(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	cache := NewCache(10*time.Minute, cacheDir)

	// Get cache file path
	cacheFile, err := cache.getCacheFilePath()
	if err != nil {
		t.Fatalf("getCacheFilePath() failed: %v", err)
	}

	// Should end with sync-config.json
	if filepath.Base(cacheFile) != "sync-config.json" {
		t.Errorf("Expected cache file named sync-config.json, got: %s", filepath.Base(cacheFile))
	}

	// Should be in cache directory
	expectedPath := filepath.Join(cacheDir, "sync-config.json")
	if cacheFile != expectedPath {
		t.Errorf("Expected cache path %s, got: %s", expectedPath, cacheFile)
	}
}

// TestCache_CorruptedCache tests handling of corrupted cache file.
func TestCache_CorruptedCache(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir := test.GetTestCacheDir(t)
	cache := NewCache(10*time.Minute, cacheDir)

	// Write corrupted cache file
	cacheFile, err := cache.getCacheFilePath()
	if err != nil {
		t.Fatalf("getCacheFilePath() failed: %v", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	if err := os.WriteFile(cacheFile, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write corrupted cache: %v", err)
	}

	// Get should return error
	_, err = cache.Get()
	if err == nil {
		t.Error("Expected error for corrupted cache")
	}
}
