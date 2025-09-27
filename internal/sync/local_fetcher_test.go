package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestLocalFetcher_Success tests successful local file fetch.
func TestLocalFetcher_Success(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test file
	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{"version": "1.0", "managed": {"organizations": {}}}`)
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create fetcher
	fetcher := NewLocalFetcher(testFile)

	// Fetch
	data, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got: %s", testData, data)
	}
}

// TestLocalFetcher_FileNotFound tests missing file handling.
func TestLocalFetcher_FileNotFound(t *testing.T) {
	test.SetupTestEnvironment(t)

	fetcher := NewLocalFetcher("/nonexistent/file.json")

	_, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

// TestLocalFetcher_ContextCancellation tests context cancellation.
func TestLocalFetcher_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	fetcher := NewLocalFetcher("/any/path")

	_, err := fetcher.Fetch(ctx, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected cancellation error")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// TestLocalFetcher_Validate tests validation.
func TestLocalFetcher_Validate(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/path/to/file", false},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewLocalFetcher(tt.path)
			err := fetcher.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestLocalFetcher_String tests String() method.
func TestLocalFetcher_String(t *testing.T) {
	fetcher := NewLocalFetcher("/path/to/file.json")
	str := fetcher.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
	if str != "local file: /path/to/file.json" {
		t.Errorf("Unexpected string: %s", str)
	}
}

// TestLocalFetcher_ProgressReporting tests progress reporting.
func TestLocalFetcher_ProgressReporting(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test file
	testFile := filepath.Join(test.GetTestConfigDir(t), "test-config.json")
	testData := []byte(`{"version": "1.0"}`)
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create channel reporter
	reporter := task.NewChannelReporter()
	defer reporter.Close()

	fetcher := NewLocalFetcher(testFile)

	// Fetch in background
	done := make(chan struct{})
	go func() {
		_, _ = fetcher.Fetch(context.Background(), reporter)
		close(done)
	}()

	// Should receive at least one status update
	select {
	case status := <-reporter.Status():
		if status == "" {
			t.Error("Expected non-empty status")
		}
	case <-done:
		t.Fatal("Fetch completed without status update")
	}

	<-done
}
