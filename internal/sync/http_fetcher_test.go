package sync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aws-profile-manager/internal/task"
)

// TestHttpFetcher_Success tests successful HTTP fetch.
func TestHttpFetcher_Success(t *testing.T) {
	// Create test server
	testData := []byte(`{"version": "1.0", "managed": {"organizations": {}}}`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(testData)
	}))
	defer server.Close()

	// Create fetcher
	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		5*time.Second,
		0, // No retries for success case
		0,
		true, // Bypass SSRF for test server
		false,
	)

	// Fetch
	data, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Expected data %s, got: %s", testData, data)
	}
}

// TestHttpFetcher_CustomHeaders tests custom headers.
func TestHttpFetcher_CustomHeaders(t *testing.T) {
	// Create test server that checks headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "test-value" {
			t.Error("Custom header not received")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	headers := map[string]string{
		"X-Custom-Header": "test-value",
	}

	fetcher := NewHttpFetcher(
		server.URL,
		headers,
		5*time.Second,
		0,
		0,
		true,
		false,
	)

	_, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}
}

// TestHttpFetcher_StatusError tests non-200 status code handling.
func TestHttpFetcher_StatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		5*time.Second,
		0,
		0,
		true,
		false,
	)

	_, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error for 404 status")
	}
}

// TestHttpFetcher_Retry tests retry logic.
func TestHttpFetcher_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		5*time.Second,
		3,                   // 3 retries
		10*time.Millisecond, // Short delay for testing
		true,
		false,
	)

	data, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Fetch() failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}

	if string(data) != `{"success": true}` {
		t.Errorf("Unexpected data: %s", data)
	}
}

// TestHttpFetcher_RetryExhaustion tests retry exhaustion.
func TestHttpFetcher_RetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		5*time.Second,
		2, // 2 retries = 3 total attempts
		10*time.Millisecond,
		true,
		false,
	)

	_, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error after retry exhaustion")
	}
}

// TestHttpFetcher_ContextCancellation tests context cancellation.
func TestHttpFetcher_ContextCancellation(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Reduced from 5s to 200ms
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		10*time.Second,
		0,
		0,
		true,
		false,
	)

	// Cancel after 50ms (before server responds)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := fetcher.Fetch(ctx, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected cancellation error")
	}
}

// TestHttpFetcher_Timeout tests request timeout.
func TestHttpFetcher_Timeout(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Reduced from 5s to 200ms
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	fetcher := NewHttpFetcher(
		server.URL,
		nil,
		50*time.Millisecond, // Timeout before server responds
		0,
		0,
		true,
		false,
	)

	_, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected timeout error")
	}
}

// TestHttpFetcher_Validate tests validation.
func TestHttpFetcher_Validate(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		timeout time.Duration
		retries int
		wantErr bool
	}{
		{"valid", "https://example.com", 5 * time.Second, 3, false},
		{"empty url", "", 5 * time.Second, 3, true},
		{"negative timeout", "https://example.com", -1 * time.Second, 3, true},
		{"negative retries", "https://example.com", 5 * time.Second, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewHttpFetcher(
				tt.url,
				nil,
				tt.timeout,
				tt.retries,
				0,
				false,
				false,
			)
			err := fetcher.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHttpFetcher_String tests String() method.
func TestHttpFetcher_String(t *testing.T) {
	fetcher := NewHttpFetcher(
		"https://example.com/config.json",
		nil,
		5*time.Second,
		3,
		0,
		false,
		false,
	)
	str := fetcher.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}
