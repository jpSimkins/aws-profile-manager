package sync

import (
	"context"
	"fmt"
	"time"

	"aws-profile-manager/internal/task"
)

// S3Fetcher fetches configuration from AWS S3 using AWS CLI.
type S3Fetcher struct {
	bucket  string
	key     string
	region  string
	profile string
	awsPath string
	awsEnv  map[string]string
}

// NewS3Fetcher creates a new S3 fetcher.
//
// Parameters:
//   - bucket: S3 bucket name
//   - key: S3 object key
//   - region: AWS region (optional)
//   - profile: AWS profile to use (optional)
//   - awsPath: Path to AWS CLI executable (default: "aws")
//   - awsEnv: Environment variables for AWS CLI
//
// Returns:
//   - *S3Fetcher: New S3 fetcher instance
func NewS3Fetcher(
	bucket string,
	key string,
	region string,
	profile string,
	awsPath string,
	awsEnv map[string]string,
) *S3Fetcher {
	if awsPath == "" {
		awsPath = "aws"
	}

	return &S3Fetcher{
		bucket:  bucket,
		key:     key,
		region:  region,
		profile: profile,
		awsPath: awsPath,
		awsEnv:  awsEnv,
	}
}

// Fetch retrieves configuration from S3 using AWS CLI subprocess.
//
// Uses the task package to execute AWS CLI with proper cancellation support.
// The AWS CLI command outputs to stdout, which is captured and returned.
//
// Command format:
//
//	aws s3 cp s3://bucket/key - [--region region] [--profile profile]
func (s *S3Fetcher) Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error) {
	reporter.ReportStatus(fmt.Sprintf("Fetching from S3: s3://%s/%s", s.bucket, s.key))

	// Build AWS CLI arguments
	args := []string{
		"s3", "cp",
		fmt.Sprintf("s3://%s/%s", s.bucket, s.key),
		"-", // Output to stdout
	}

	// Add optional region
	if s.region != "" {
		args = append(args, "--region", s.region)
	}

	// Add optional profile
	if s.profile != "" {
		args = append(args, "--profile", s.profile)
	}

	// Create subprocess task
	awsTask := &task.SubprocessTask{
		Name:    s.awsPath,
		Args:    args,
		Env:     s.awsEnv,
		Timeout: 30 * time.Second,
	}

	// Execute (subprocess killed on context cancel)
	result, err := task.Run(ctx, awsTask, reporter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from S3: %w", err)
	}

	reporter.ReportStatus(fmt.Sprintf("Successfully fetched %d bytes from S3", len(result.Output)))
	return result.Output, nil
}

// String returns human-readable description.
func (s *S3Fetcher) String() string {
	return fmt.Sprintf("s3: s3://%s/%s", s.bucket, s.key)
}

// Validate checks if the S3 fetcher configuration is valid.
func (s *S3Fetcher) Validate() error {
	if s.bucket == "" {
		return fmt.Errorf("s3 bucket is required")
	}
	if s.key == "" {
		return fmt.Errorf("s3 key is required")
	}
	return nil
}
