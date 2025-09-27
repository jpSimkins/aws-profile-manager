package sync

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateConfig validates a SyncConfig for security and correctness.
//
// Performs validation based on strategy:
//   - HTTP: URL format, SSRF protection (unless bypassed)
//   - S3: Bucket/key format, region validation
//   - Local: Path validation (non-empty)
//
// Returns error if validation fails.
func ValidateConfig(cfg SyncConfig) error {
	switch cfg.Strategy {
	case StrategyHTTP:
		return validateHTTPConfig(cfg)
	case StrategyS3:
		return validateS3Config(cfg)
	case StrategyLocal:
		return validateLocalConfig(cfg)
	default:
		return fmt.Errorf("unknown strategy: %s", cfg.Strategy)
	}
}

// validateHTTPConfig validates HTTP strategy configuration.
func validateHTTPConfig(cfg SyncConfig) error {
	if cfg.HTTPUrl == "" {
		return fmt.Errorf("http url is required")
	}

	// Parse URL
	parsedURL, err := url.Parse(cfg.HTTPUrl)
	if err != nil {
		return fmt.Errorf("invalid http url: %w", err)
	}

	// Require HTTPS or HTTP
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("url must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	// SSRF protection (unless explicitly bypassed)
	if !cfg.HTTPBypassSSRF {
		if err := checkSSRF(parsedURL); err != nil {
			return fmt.Errorf("ssrf check failed (use HTTPBypassSSRF=true to override): %w", err)
		}
	}

	// Validate timeout
	if cfg.HTTPTimeout < 0 {
		return fmt.Errorf("http timeout cannot be negative")
	}

	// Validate retries
	if cfg.HTTPRetries < 0 {
		return fmt.Errorf("http retries cannot be negative")
	}

	return nil
}

// validateS3Config validates S3 strategy configuration.
func validateS3Config(cfg SyncConfig) error {
	if cfg.S3Bucket == "" {
		return fmt.Errorf("s3 bucket is required")
	}

	if cfg.S3Key == "" {
		return fmt.Errorf("s3 key is required")
	}

	// Validate bucket name format (basic check)
	if err := validateS3BucketName(cfg.S3Bucket); err != nil {
		return fmt.Errorf("invalid s3 bucket name: %w", err)
	}

	// Validate region (if specified)
	if cfg.S3Region != "" {
		if err := validateAWSRegion(cfg.S3Region); err != nil {
			return fmt.Errorf("invalid s3 region: %w", err)
		}
	}

	return nil
}

// validateLocalConfig validates local strategy configuration.
func validateLocalConfig(cfg SyncConfig) error {
	if cfg.LocalPath == "" {
		return fmt.Errorf("local path is required")
	}

	// Basic path validation (no directory traversal attempts)
	if strings.Contains(cfg.LocalPath, "..") {
		return fmt.Errorf("local path cannot contain '..' (directory traversal)")
	}

	return nil
}

// checkSSRF performs SSRF (Server-Side Request Forgery) protection checks.
//
// Blocks requests to:
//   - Private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
//   - Loopback addresses (127.0.0.0/8, ::1)
//   - Link-local addresses (169.254.0.0/16, fe80::/10)
//   - Localhost hostname
func checkSSRF(parsedURL *url.URL) error {
	hostname := parsedURL.Hostname()

	// Check for localhost
	if hostname == "localhost" {
		return fmt.Errorf("localhost access blocked for security")
	}

	// Resolve hostname to IP
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	// Check each resolved IP
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("private ip address blocked for security: %s", ip)
		}
	}

	return nil
}

// isPrivateIP checks if an IP is in a private range.
func isPrivateIP(ip net.IP) bool {
	// Private IPv4 ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Private network
		"172.16.0.0/12",  // Private network
		"192.168.0.0/16", // Private network
		"127.0.0.0/8",    // Loopback
		"169.254.0.0/16", // Link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// validateS3BucketName validates S3 bucket name format.
//
// Rules:
//   - 3-63 characters
//   - Lowercase letters, numbers, hyphens, periods
//   - Start and end with letter or number
//   - No consecutive periods
func validateS3BucketName(bucket string) error {
	if len(bucket) < 3 || len(bucket) > 63 {
		return fmt.Errorf("bucket name must be 3-63 characters")
	}

	// Check valid characters
	for i, c := range bucket {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.') {
			return fmt.Errorf("bucket name contains invalid character: %c", c)
		}

		// First and last must be letter or number
		if i == 0 || i == len(bucket)-1 {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
				return fmt.Errorf("bucket name must start and end with letter or number")
			}
		}
	}

	// Check for consecutive periods
	if strings.Contains(bucket, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive periods")
	}

	return nil
}

// validateAWSRegion validates AWS region format.
//
// Basic check: region should match pattern like "us-east-1", "eu-west-2", etc.
func validateAWSRegion(region string) error {
	// Basic format check (region-direction-number)
	parts := strings.Split(region, "-")
	if len(parts) < 3 {
		return fmt.Errorf("region must have format like 'us-east-1'")
	}

	return nil
}
