package awscli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestNewCache(t *testing.T) {
	test.SetupTestEnvironment(t)

	cacheDir, err := settings.GetCacheDir()
	if err != nil {
		t.Fatalf("Failed to get cache directory: %v", err)
	}

	cache := NewCache(cacheDir)

	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}

	if cache.cacheDir != cacheDir {
		t.Errorf("Expected cache directory %s, got %s", cacheDir, cache.cacheDir)
	}

	if cache.extractor == nil {
		t.Error("Expected extractor to be initialized")
	}
}

func TestNewCacheWithExtractor(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)
	cacheDir := test.GetTestCacheDir(t)

	// Create a test config file
	testConfig := `[default]
region = us-east-1

[profile test-profile]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-west-2
sso_session = test-session

[sso-session test-session]
sso_start_url = https://example.awsapps.com/start
sso_region = us-east-1
`
	err := os.WriteFile(configFile, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	cache := NewCacheWithExtractor(extractor, cacheDir)

	if cache == nil {
		t.Fatal("NewCacheWithExtractor() returned nil")
	}

	if cache.cacheDir != cacheDir {
		t.Errorf("Expected cache directory %s, got %s", cacheDir, cache.cacheDir)
	}

	if cache.extractor != extractor {
		t.Error("Expected extractor to match the provided one")
	}
}

func TestGetData_FirstTime(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)
	cacheDir := test.GetTestCacheDir(t)

	// Create a test config file
	testConfig := `[default]
region = us-east-1

[profile test-profile]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-west-2
sso_session = test-session

[sso-session test-session]
sso_start_url = https://example.awsapps.com/start
sso_region = us-east-1
`
	err := os.WriteFile(configFile, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	cache := NewCacheWithExtractor(extractor, cacheDir)

	// First call should extract from file and cache it
	data, err := cache.GetData()
	if err != nil {
		t.Fatalf("GetData() failed: %v", err)
	}

	if data == nil {
		t.Fatal("GetData() returned nil data")
	}

	// Verify extracted data
	if len(data.Profiles) != 2 { // default + test-profile
		t.Errorf("Expected 2 profiles, got %d", len(data.Profiles))
	}

	if len(data.SsoSessions) != 1 {
		t.Errorf("Expected 1 SSO session, got %d", len(data.SsoSessions))
	}

	// Verify cache file was created
	cachePath := filepath.Join(cacheDir, CacheFileName)
	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("Cache file was not created: %v", err)
	}
}

func TestGetData_FromCache(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create a test config file
	testConfig := `[profile test-profile]
sso_account_id = 123456789012
sso_role_name = AdminRole
`
	err := os.WriteFile(tempConfigFile, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	// First call - creates cache
	_, err = cache.GetData()
	if err != nil {
		t.Fatalf("First GetData() failed: %v", err)
	}

	// Second call - should use cache
	data, err := cache.GetData()
	if err != nil {
		t.Fatalf("Second GetData() failed: %v", err)
	}

	if len(data.Profiles) != 1 {
		t.Errorf("Expected 1 profile from cache, got %d", len(data.Profiles))
	}
}

func TestForceReload(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create initial config
	initialConfig := `[profile initial]
sso_account_id = 111111111111
`
	err := os.WriteFile(tempConfigFile, []byte(initialConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create initial config: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	// Get initial data
	data1, err := cache.GetData()
	if err != nil {
		t.Fatalf("Initial GetData() failed: %v", err)
	}

	if len(data1.Profiles) != 1 {
		t.Errorf("Expected 1 initial profile, got %d", len(data1.Profiles))
	}

	// Update config file
	updatedConfig := `[profile initial]
sso_account_id = 111111111111

[profile new-profile]
sso_account_id = 222222222222
`
	err = os.WriteFile(tempConfigFile, []byte(updatedConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Force reload
	data2, err := cache.ForceReload()
	if err != nil {
		t.Fatalf("ForceReload() failed: %v", err)
	}

	if len(data2.Profiles) != 2 {
		t.Errorf("Expected 2 profiles after reload, got %d", len(data2.Profiles))
	}
}

func TestClearCache(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create test config
	testConfig := `[profile test]
region = us-east-1
`
	err := os.WriteFile(tempConfigFile, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	// Create cache
	_, err = cache.GetData()
	if err != nil {
		t.Fatalf("GetData() failed: %v", err)
	}

	// Verify cache exists
	if !cache.HasCache() {
		t.Error("Cache should exist before clearing")
	}

	// Clear cache
	err = cache.ClearCache()
	if err != nil {
		t.Fatalf("ClearCache() failed: %v", err)
	}

	// Verify cache is gone
	if cache.HasCache() {
		t.Error("Cache should not exist after clearing")
	}
}

func TestHasCache(t *testing.T) {
	test.SetupTestEnvironment(t)
	cacheDir := test.GetTestCacheDir(t)

	cache := NewCache(cacheDir)

	// Initially no cache
	if cache.HasCache() {
		t.Error("HasCache() should return false when no cache exists")
	}

	// Create cache directory first
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	// Create a dummy cache file
	cachePath := filepath.Join(cacheDir, CacheFileName)
	err := os.WriteFile(cachePath, []byte("{}"), 0600)
	if err != nil {
		t.Fatalf("Failed to create dummy cache file: %v", err)
	}

	// Now should have cache
	if !cache.HasCache() {
		t.Error("HasCache() should return true when cache file exists")
	}
}

func TestNeedsReload_NoCache(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create config file
	err := os.WriteFile(tempConfigFile, []byte("[default]\nregion = us-east-1"), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	needsReload, err := cache.needsReload()
	if err != nil {
		t.Fatalf("needsReload() failed: %v", err)
	}

	if !needsReload {
		t.Error("needsReload() should return true when no cache exists")
	}
}

func TestNeedsReload_CacheNewer(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create config file
	configContent := "[default]\nregion = us-east-1"
	err := os.WriteFile(tempConfigFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	// Create cache
	_, err = cache.GetData()
	if err != nil {
		t.Fatalf("GetData() failed: %v", err)
	}

	// Check if reload needed (should be false since cache is current)
	needsReload, err := cache.needsReload()
	if err != nil {
		t.Fatalf("needsReload() failed: %v", err)
	}

	if needsReload {
		t.Error("needsReload() should return false when cache is current")
	}
}

func TestNeedsReload_ConfigNewer(t *testing.T) {
	test.SetupTestEnvironment(t)
	tempConfigFile := test.GetTestAwsConfigPath(t)

	// Create initial config
	err := os.WriteFile(tempConfigFile, []byte("[default]\nregion = us-east-1"), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(tempConfigFile)
	cache := NewCacheWithExtractor(extractor, test.GetTestCacheDir(t))

	// Create cache
	_, err = cache.GetData()
	if err != nil {
		t.Fatalf("GetData() failed: %v", err)
	}

	// Wait a moment and update config file
	time.Sleep(10 * time.Millisecond)
	err = os.WriteFile(tempConfigFile, []byte("[default]\nregion = us-west-2"), 0600)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Check if reload needed (should be true since config is newer)
	needsReload, err := cache.needsReload()
	if err != nil {
		t.Fatalf("needsReload() failed: %v", err)
	}

	if !needsReload {
		t.Error("needsReload() should return true when config is newer than cache")
	}
}

func TestGetCachePath_EmptyDirectory(t *testing.T) {
	cache := NewCache("")

	_, err := cache.getCachePath()
	if err == nil {
		t.Error("getCachePath() should fail with empty cache directory")
	}
}

func TestGetCachePath_ValidDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)
	cacheDir := test.GetTestCacheDir(t)

	cache := NewCache(cacheDir)

	path, err := cache.getCachePath()
	if err != nil {
		t.Fatalf("getCachePath() failed: %v", err)
	}

	expectedPath := filepath.Join(cacheDir, CacheFileName)
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
}

func TestCacheDataSerialization(t *testing.T) {
	test.SetupTestEnvironment(t)
	configFile := test.GetTestAwsConfigPath(t)
	cacheDir := test.GetTestCacheDir(t)

	// Create test config
	testConfig := `[profile test]
sso_account_id = 123456789012
sso_role_name = AdminRole
region = us-east-1
`
	err := os.WriteFile(configFile, []byte(testConfig), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	cache := NewCacheWithExtractor(extractor, cacheDir)

	// Get data to create cache
	originalData, err := cache.GetData()
	if err != nil {
		t.Fatalf("GetData() failed: %v", err)
	}

	// Read cache file directly
	cachePath := filepath.Join(cacheDir, CacheFileName)
	cacheContent, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("Failed to read cache file: %v", err)
	}

	// Parse cache data
	var cacheData CacheData
	err = json.Unmarshal(cacheContent, &cacheData)
	if err != nil {
		t.Fatalf("Failed to parse cache data: %v", err)
	}

	// Verify cache structure
	if len(cacheData.Data.Profiles) != len(originalData.Profiles) {
		t.Errorf("Cache profiles count mismatch: expected %d, got %d",
			len(originalData.Profiles), len(cacheData.Data.Profiles))
	}

	if cacheData.SourcePath != configFile {
		t.Errorf("Cache source path mismatch: expected %s, got %s",
			configFile, cacheData.SourcePath)
	}

	if cacheData.CachedAt.IsZero() {
		t.Error("Cache timestamp should be set")
	}

	if cacheData.LastModified.IsZero() {
		t.Error("Last modified timestamp should be set")
	}
}

func TestCacheWithCorruptedFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	cacheDir := test.GetTestCacheDir(t)

	cache := NewCache(cacheDir)

	// Create corrupted cache file
	cachePath := filepath.Join(cacheDir, CacheFileName)
	err := os.WriteFile(cachePath, []byte("invalid json content"), 0600)
	if err != nil {
		t.Fatalf("Failed to create corrupted cache: %v", err)
	}

	// Try to load from corrupted cache
	_, err = cache.loadFromCache()
	if err == nil {
		t.Error("loadFromCache() should fail with corrupted cache file")
	}
}

func TestCacheDirectoryCreation(t *testing.T) {
	test.SetupTestEnvironment(t)
	// Create a custom nested cache directory path for this test
	nestedCacheDir := filepath.Join(test.GetTestCacheDir(t), "nested", "cache", "dir")
	configFile := test.GetTestAwsConfigPath(t)

	// Create config file
	err := os.WriteFile(configFile, []byte("[default]\nregion = us-east-1"), 0600)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	extractor := NewExtractorWithPath(configFile)
	cache := NewCacheWithExtractor(extractor, nestedCacheDir)

	// This should create the nested directory structure
	_, err = cache.GetData()
	if err != nil {
		t.Fatalf("GetData() should create nested directories: %v", err)
	}

	// Verify the cache file was created in the nested directory
	cachePath := filepath.Join(nestedCacheDir, CacheFileName)
	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("Cache file should exist in nested directory: %v", err)
	}
}
