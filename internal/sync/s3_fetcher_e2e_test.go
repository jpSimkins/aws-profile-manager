//go:build e2e

package sync

import (
	"context"
	"os"
	"testing"

	"aws-profile-manager/internal/task"
)

// TestS3Fetcher_RealS3 tests fetching from a real S3 bucket.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_S3_BUCKET="my-config-bucket"
//	export SYNC_E2E_S3_KEY="configs/aws-profiles.json"
//	export SYNC_E2E_S3_REGION="us-east-1"
//	export SYNC_E2E_S3_PROFILE="my-sso-profile"  # Optional
//	go test -tags=e2e -v ./internal/sync -run TestS3Fetcher_RealS3
//
// Requirements:
//   - SYNC_E2E_S3_BUCKET: S3 bucket name
//   - SYNC_E2E_S3_KEY: S3 object key
//   - SYNC_E2E_S3_REGION: AWS region (optional, defaults to us-east-1)
//   - SYNC_E2E_S3_PROFILE: AWS CLI profile (optional)
//   - Valid AWS credentials (SSO, IAM role, or env vars)
//   - AWS CLI installed
func TestS3Fetcher_RealS3(t *testing.T) {
	bucket := os.Getenv("SYNC_E2E_S3_BUCKET")
	key := os.Getenv("SYNC_E2E_S3_KEY")

	if bucket == "" || key == "" {
		t.Skip("SYNC_E2E_S3_BUCKET or SYNC_E2E_S3_KEY not set - skipping E2E test")
	}

	region := os.Getenv("SYNC_E2E_S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	profile := os.Getenv("SYNC_E2E_S3_PROFILE")

	t.Logf("Testing S3 fetch from: s3://%s/%s", bucket, key)
	if profile != "" {
		t.Logf("Using AWS profile: %s", profile)
	}

	fetcher := NewS3Fetcher(
		bucket,
		key,
		region,
		profile,
		"aws",
		nil,
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

	t.Logf("Fetched %d bytes successfully from S3", len(data))
}

// TestS3Fetcher_RealS3WithSSO tests fetching using SSO profile.
//
// This test is SKIPPED by default. To run:
//
//	aws sso login --profile my-sso-profile
//	export SYNC_E2E_S3_BUCKET="my-config-bucket"
//	export SYNC_E2E_S3_KEY="configs/aws-profiles.json"
//	export SYNC_E2E_S3_PROFILE="my-sso-profile"
//	go test -tags=e2e -v ./internal/sync -run TestS3Fetcher_RealS3WithSSO
//
// Requirements:
//   - Valid SSO session (run `aws sso login` first)
//   - SYNC_E2E_S3_BUCKET: S3 bucket name
//   - SYNC_E2E_S3_KEY: S3 object key
//   - SYNC_E2E_S3_PROFILE: SSO profile name
func TestS3Fetcher_RealS3WithSSO(t *testing.T) {
	bucket := os.Getenv("SYNC_E2E_S3_BUCKET")
	key := os.Getenv("SYNC_E2E_S3_KEY")
	profile := os.Getenv("SYNC_E2E_S3_PROFILE")

	if bucket == "" || key == "" || profile == "" {
		t.Skip("SYNC_E2E_S3_BUCKET, SYNC_E2E_S3_KEY, or SYNC_E2E_S3_PROFILE not set - skipping E2E test")
	}

	t.Logf("Testing S3 fetch with SSO profile '%s' from: s3://%s/%s", profile, bucket, key)

	fetcher := NewS3Fetcher(
		bucket,
		key,
		"", // Will use profile's region
		profile,
		"aws",
		nil,
	)

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed (ensure SSO session is valid with 'aws sso login --profile %s'): %v", profile, err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully using SSO", len(data))
}

// TestS3Fetcher_PublicBucket tests fetching from a public S3 bucket.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_S3_BUCKET="public-bucket-name"
//	export SYNC_E2E_S3_KEY="path/to/public/file.json"
//	go test -tags=e2e -v ./internal/sync -run TestS3Fetcher_PublicBucket
//
// Requirements:
//   - SYNC_E2E_S3_BUCKET: Public S3 bucket name
//   - SYNC_E2E_S3_KEY: S3 object key (must have public-read access)
//   - No credentials needed (public access)
func TestS3Fetcher_PublicBucket(t *testing.T) {
	bucket := os.Getenv("SYNC_E2E_S3_BUCKET")
	key := os.Getenv("SYNC_E2E_S3_KEY")

	if bucket == "" || key == "" {
		t.Skip("SYNC_E2E_S3_BUCKET or SYNC_E2E_S3_KEY not set - skipping E2E test")
	}

	t.Logf("Testing S3 fetch from public bucket: s3://%s/%s", bucket, key)

	fetcher := NewS3Fetcher(
		bucket,
		key,
		"us-east-1",
		"", // No profile (public access)
		"aws",
		nil,
	)

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed (ensure object has public-read access): %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully from public bucket", len(data))
}
