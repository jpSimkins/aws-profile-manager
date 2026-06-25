package sync

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"aws-profile-manager/internal/task"
)

// HttpFetcher fetches configuration via HTTP/HTTPS.
type HttpFetcher struct {
	url        string
	headers    map[string]string
	timeout    time.Duration
	retries    int
	retryDelay time.Duration
	tlsVerify  bool
	bypassSSRF bool
	bypassTLS  bool
}

// NewHttpFetcher creates a new HTTP fetcher.
//
// Parameters:
//   - url: HTTP/HTTPS URL to fetch
//   - headers: Custom HTTP headers (optional)
//   - timeout: Request timeout
//   - retries: Number of retry attempts
//   - retryDelay: Delay between retries
//   - tlsVerify: Whether to verify TLS certificates
//   - bypassSSRF: Whether to bypass SSRF protection
//   - bypassTLS: Whether to bypass TLS verification
//
// Returns:
//   - *HttpFetcher: New HTTP fetcher instance
func NewHttpFetcher(
	url string,
	headers map[string]string,
	timeout time.Duration,
	retries int,
	retryDelay time.Duration,
	tlsVerify bool,
	bypassSSRF bool,
	bypassTLS bool,
) *HttpFetcher {
	return &HttpFetcher{
		url:        url,
		headers:    headers,
		timeout:    timeout,
		retries:    retries,
		retryDelay: retryDelay,
		tlsVerify:  tlsVerify,
		bypassSSRF: bypassSSRF,
		bypassTLS:  bypassTLS,
	}
}

// Fetch retrieves configuration via HTTP/HTTPS with retry logic.
//
// Features:
//   - Context cancellation support
//   - Automatic retries on transient failures
//   - Progress reporting for each attempt
//   - TLS certificate verification (unless bypassed)
//   - Custom headers support
func (h *HttpFetcher) Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error) {
	reporter.ReportStatus(fmt.Sprintf("Fetching from HTTP: %s", h.url))

	var lastErr error
	maxAttempts := h.retries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check for cancellation before each attempt
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Report attempt
		if attempt > 1 {
			reporter.ReportStatus(fmt.Sprintf("Retry attempt %d/%d", attempt-1, h.retries))
		}

		// Try to fetch
		data, err := h.fetchOnce(ctx)
		if err == nil {
			reporter.ReportStatus(fmt.Sprintf("Successfully fetched %d bytes", len(data)))
			return data, nil
		}

		lastErr = err
		reporter.ReportError(fmt.Errorf("attempt %d failed: %w", attempt, err))

		// Don't sleep after last attempt
		if attempt < maxAttempts {
			// Sleep with cancellation support
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(h.retryDelay):
			}
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// fetchOnce performs a single HTTP fetch attempt.
func (h *HttpFetcher) fetchOnce(ctx context.Context) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: h.timeout,
	}

	// Configure TLS if needed
	if !h.tlsVerify || h.bypassTLS {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- User explicitly requested TLS verification bypass.
			},
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// String returns human-readable description.
func (h *HttpFetcher) String() string {
	return fmt.Sprintf("http: %s", h.url)
}

// Validate checks if the HTTP fetcher configuration is valid.
func (h *HttpFetcher) Validate() error {
	if h.url == "" {
		return fmt.Errorf("http url is required")
	}
	if h.timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	if h.retries < 0 {
		return fmt.Errorf("retries cannot be negative")
	}
	return nil
}
