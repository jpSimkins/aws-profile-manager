package sync

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestSync_LocalSuccess tests successful sync from local file.
func TestSync_LocalSuccess(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config file
	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{
		"version": "1.0",
		"managed": {
			"organizations": {}
		}
	}`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config
	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  10 * time.Minute,
		CacheDir:  test.GetTestCacheDir(t),
	}

	// Sync
	result, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Sync() failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result")
	}
	if result.Data == nil {
		t.Fatal("Expected data in result")
	}
	if result.Data.Version != "1.0" {
		t.Errorf("Expected version 1.0, got: %s", result.Data.Version)
	}
	if result.CacheHit {
		t.Error("Expected cache miss on first sync")
	}
}

// TestSync_CacheHit tests cache hit behavior.
func TestSync_CacheHit(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config file
	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{
		"version": "1.0",
		"managed": {"organizations": {}}
	}`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  10 * time.Minute,
		CacheDir:  test.GetTestCacheDir(t),
	}

	// First sync - should miss cache
	result1, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("First Sync() failed: %v", err)
	}
	if result1.CacheHit {
		t.Error("Expected cache miss on first sync")
	}

	// Second sync - should hit cache
	result2, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Second Sync() failed: %v", err)
	}
	if !result2.CacheHit {
		t.Error("Expected cache hit on second sync")
	}
	if result2.Source != "cache" {
		t.Errorf("Expected source 'cache', got: %s", result2.Source)
	}
}

// TestSync_ForceRefresh tests cache bypass.
func TestSync_ForceRefresh(t *testing.T) {
	test.SetupTestEnvironment(t)

	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{
		"version": "1.0",
		"managed": {"organizations": {}}
	}`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  10 * time.Minute,
		CacheDir:  test.GetTestCacheDir(t),
	}

	// First sync - populate cache
	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("First Sync() failed: %v", err)
	}

	// Second sync with force refresh - should bypass cache
	result, err := Sync(context.Background(), cfg, Options{ForceRefresh: true}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Second Sync() failed: %v", err)
	}
	if result.CacheHit {
		t.Error("Expected cache miss with force refresh")
	}
}

// TestSync_InvalidConfig tests configuration validation.
func TestSync_InvalidConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: "", // Invalid - empty path
	}

	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error for invalid config")
	}
}

// TestSync_InvalidJSON tests JSON parse error handling.
func TestSync_InvalidJSON(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create file with invalid JSON
	testFile := filepath.Join(test.GetTestConfigDir(t), "bad-config.json")
	if err := os.WriteFile(testFile, []byte("not valid json"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  0, // Disable cache
		CacheDir:  test.GetTestCacheDir(t),
	}

	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

// TestSync_ContextCancellation tests context cancellation.
func TestSync_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  0,
		CacheDir:  test.GetTestCacheDir(t),
	}

	_, err := Sync(ctx, cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected cancellation error")
	}
}

// TestSync_UnknownStrategy tests unknown strategy handling.
func TestSync_UnknownStrategy(t *testing.T) {
	test.SetupTestEnvironment(t)

	cfg := SyncConfig{
		Strategy: Strategy("unknown"),
	}

	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error for unknown strategy")
	}
}

// TestClearCache tests cache clearing function.
func TestClearCache(t *testing.T) {
	test.SetupTestEnvironment(t)

	cfg := SyncConfig{
		CacheTTL: 10 * time.Minute,
		CacheDir: test.GetTestCacheDir(t),
	}

	// Clear should not error even if cache doesn't exist
	if err := ClearCache(cfg); err != nil {
		t.Errorf("ClearCache() failed: %v", err)
	}
}

// TestConfigFromSettings tests settings conversion.
func TestConfigFromSettings(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test settings
	syncSettings := settings.GetDefaultSync()
	syncSettings.Strategy = "local"
	syncSettings.Local.Path = "/test/path"
	syncSettings.HTTP.URL = "https://example.com"
	syncSettings.S3.Bucket = "test-bucket"
	syncSettings.S3.Key = "test-key"

	cfg := ConfigFromSettings(&syncSettings)

	if cfg.Strategy != StrategyLocal {
		t.Errorf("Expected strategy local, got: %s", cfg.Strategy)
	}
	if cfg.LocalPath != "/test/path" {
		t.Errorf("Expected local path /test/path, got: %s", cfg.LocalPath)
	}
	if cfg.HTTPUrl != "https://example.com" {
		t.Errorf("Expected HTTP URL https://example.com, got: %s", cfg.HTTPUrl)
	}
	if cfg.S3Bucket != "test-bucket" {
		t.Errorf("Expected S3 bucket test-bucket, got: %s", cfg.S3Bucket)
	}
}

// TestSync_ProgressReporting tests progress reporting.
func TestSync_ProgressReporting(t *testing.T) {
	test.SetupTestEnvironment(t)

	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{
		"version": "1.0",
		"managed": {"organizations": {}}
	}`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  0, // Disable cache to force fetch
		CacheDir:  test.GetTestCacheDir(t),
	}

	reporter := task.NewChannelReporter()
	defer reporter.Close()

	// Sync in background
	done := make(chan struct{})
	go func() {
		_, _ = Sync(context.Background(), cfg, Options{}, reporter)
		close(done)
	}()

	// Should receive at least one status update
	select {
	case status := <-reporter.Status():
		if status == "" {
			t.Error("Expected non-empty status")
		}
	case <-done:
		t.Fatal("Sync completed without status update")
	}

	<-done
}

// TestSync_InvalidSchemaRejected tests that invalid schemas are rejected before caching.
func TestSync_InvalidSchemaRejected(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config file with INVALID schema (missing required fields)
	testFile := filepath.Join(test.GetTestConfigDir(t), "invalid-config.json")
	testData := []byte(`{
		"invalid": "data",
		"missing": "required fields"
	}`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config
	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  10 * time.Minute,
		CacheDir:  test.GetTestCacheDir(t),
	}

	// Sync should fail with validation error
	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected validation error for invalid schema")
	}

	// Error should mention validation
	if !strings.Contains(err.Error(), "validation") && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Error should mention validation failure, got: %v", err)
	}

	// Verify cache was NOT created (invalid data should not be cached)
	cacheDir, _ := settings.GetCacheDir()
	cacheFile := filepath.Join(cacheDir, "sync-config.json")
	if _, err := os.Stat(cacheFile); err == nil {
		t.Error("Cache file should not exist for invalid schema")
	}
}

// TestSync_MalformedJSONRejected tests that malformed JSON is rejected.
func TestSync_MalformedJSONRejected(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test config file with malformed JSON
	testFile := filepath.Join(test.GetTestConfigDir(t), "malformed-config.json")
	testData := []byte(`{
		"invalid json": "missing closing brace"
	`)
	if err := os.WriteFile(testFile, testData, 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config
	cfg := SyncConfig{
		Strategy:  StrategyLocal,
		LocalPath: testFile,
		CacheTTL:  10 * time.Minute,
		CacheDir:  test.GetTestCacheDir(t),
	}

	// Sync should fail with parse error
	_, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected parse error for malformed JSON")
	}

	// Error should mention parsing
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Error should mention parse failure, got: %v", err)
	}

	// Verify cache was NOT created
	cacheDir, _ := settings.GetCacheDir()
	cacheFile := filepath.Join(cacheDir, "sync-config.json")
	if _, err := os.Stat(cacheFile); err == nil {
		t.Error("Cache file should not exist for malformed JSON")
	}
}

// TestSync_OptionalFields tests that optional fields are handled correctly.
func TestSync_OptionalFields(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
		checkFn  func(*testing.T, *Result)
	}{
		{
			name:     "minimal schema (only managed)",
			jsonData: `{"version": "1.0", "managed": {"organizations": {}}}`,
			wantErr:  false,
			checkFn: func(t *testing.T, r *Result) {
				if r.Data.Managed == nil {
					t.Error("Expected managed section")
				}
				if r.Data.Unmanaged != nil {
					t.Error("Expected nil unmanaged section")
				}
			},
		},
		{
			name: "with unmanaged section",
			jsonData: `{
				"version": "1.0",
				"managed": {"organizations": {}},
				"unmanaged": {
					"above_managed": {"profiles": []},
					"below_managed": {"profiles": []}
				}
			}`,
			wantErr: false,
			checkFn: func(t *testing.T, r *Result) {
				if r.Data.Managed == nil {
					t.Error("Expected managed section")
				}
				if r.Data.Unmanaged == nil {
					t.Error("Expected unmanaged section")
				}
			},
		},
		{
			name: "with settings section",
			jsonData: `{
				"version": "1.0",
				"managed": {"organizations": {}},
				"settings": {"theme": "dark"}
			}`,
			wantErr: false,
			checkFn: func(t *testing.T, r *Result) {
				if r.Data.Managed == nil {
					t.Error("Expected managed section")
				}
				if r.Data.Settings == nil {
					t.Error("Expected settings section")
				}
			},
		},
		{
			name: "with metadata section",
			jsonData: `{
				"version": "1.0",
				"managed": {"organizations": {}},
				"metadata": {
					"exported_at": "2025-10-28T12:00:00Z",
					"exported_by": "test-user",
					"source": "test"
				}
			}`,
			wantErr: false,
			checkFn: func(t *testing.T, r *Result) {
				if r.Data.Managed == nil {
					t.Error("Expected managed section")
				}
				if r.Data.Metadata == nil {
					t.Error("Expected metadata section")
				}
			},
		},
		{
			name: "all optional fields present",
			jsonData: `{
				"version": "1.0",
				"managed": {"organizations": {}},
				"unmanaged": {
					"above_managed": {"profiles": []},
					"below_managed": {"profiles": []}
				},
				"settings": {"theme": "dark"},
				"metadata": {
					"exported_at": "2025-10-28T12:00:00Z",
					"exported_by": "test-user",
					"source": "test"
				}
			}`,
			wantErr: false,
			checkFn: func(t *testing.T, r *Result) {
				if r.Data.Managed == nil {
					t.Error("Expected managed section")
				}
				if r.Data.Unmanaged == nil {
					t.Error("Expected unmanaged section")
				}
				if r.Data.Settings == nil {
					t.Error("Expected settings section")
				}
				if r.Data.Metadata == nil {
					t.Error("Expected metadata section")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(test.GetTestConfigDir(t), "config.json")
			if err := os.WriteFile(testFile, []byte(tt.jsonData), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			cfg := SyncConfig{
				Strategy:  StrategyLocal,
				LocalPath: testFile,
				CacheTTL:  0, // Disable cache
				CacheDir:  test.GetTestCacheDir(t),
			}

			result, err := Sync(context.Background(), cfg, Options{}, task.NoOpReporter{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil || result.Data == nil {
					t.Fatal("Expected valid result")
				}
				if result.Data.Version != "1.0" {
					t.Errorf("Expected version 1.0, got: %s", result.Data.Version)
				}

				// Run custom checks
				if tt.checkFn != nil {
					tt.checkFn(t, result)
				}
			}
		})
	}
}
