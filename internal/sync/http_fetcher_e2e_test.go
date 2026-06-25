//go:build e2e

package sync

import (
	"context"
	"os"
	"testing"
	"time"

	"aws-profile-manager/internal/task"
)

// TestHttpFetcher_RealHTTP tests fetching from a real HTTP endpoint.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_HTTP_URL="https://example.com/config.json"
//	go test -tags=e2e -v ./internal/sync -run TestHttpFetcher_RealHTTP
//
// Requirements:
//   - SYNC_E2E_HTTP_URL: Public HTTP/HTTPS URL to a valid config JSON
func TestHttpFetcher_RealHTTP(t *testing.T) {
	url := os.Getenv("SYNC_E2E_HTTP_URL")
	if url == "" {
		t.Skip("SYNC_E2E_HTTP_URL not set - skipping E2E test")
	}

	t.Logf("Testing HTTP fetch from: %s", url)

	fetcher := NewHttpFetcher(
		url,
		nil,
		30*time.Second,
		3,
		2*time.Second,
		false,
		false,
	)

	if err := fetcher.Validate(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully", len(data))
}

// TestHttpFetcher_RealHTTPWithAuth tests fetching from a protected endpoint.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_HTTP_URL="https://example.com/config.json"
//	export SYNC_E2E_HTTP_HEADER_Authorization="Bearer your-token-here"
//	go test -tags=e2e -v ./internal/sync -run TestHttpFetcher_RealHTTPWithAuth
//
// Requirements:
//   - SYNC_E2E_HTTP_URL: URL requiring authentication
//   - SYNC_E2E_HTTP_HEADER_*: Custom headers (e.g., Authorization)
func TestHttpFetcher_RealHTTPWithAuth(t *testing.T) {
	url := os.Getenv("SYNC_E2E_HTTP_URL")
	if url == "" {
		t.Skip("SYNC_E2E_HTTP_URL not set - skipping E2E test")
	}

	// Collect custom headers from environment
	headers := make(map[string]string)
	for _, env := range os.Environ() {
		if len(env) > 23 && env[:23] == "SYNC_E2E_HTTP_HEADER_" {
			parts := splitEnvVar(env)
			if len(parts) == 2 {
				headerName := parts[0][23:] // Remove "SYNC_E2E_HTTP_HEADER_" prefix
				headers[headerName] = parts[1]
			}
		}
	}

	t.Logf("Testing HTTP fetch with auth from: %s", url)
	t.Logf("Using %d custom headers", len(headers))

	fetcher := NewHttpFetcher(
		url,
		headers,
		30*time.Second,
		3,
		2*time.Second,
		false,
		false,
	)

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully", len(data))
}

// splitEnvVar splits "KEY=value" into ["KEY", "value"]
func splitEnvVar(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
