package sync

import (
	"context"
	"testing"

	"aws-profile-manager/internal/task"
)

// TestS3Fetcher_Validate tests S3 fetcher validation.
func TestS3Fetcher_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bucket  string
		key     string
		wantErr bool
	}{
		{"valid", "my-bucket", "config.json", false},
		{"missing bucket", "", "config.json", true},
		{"missing key", "my-bucket", "", true},
		{"both missing", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewS3Fetcher(
				tt.bucket,
				tt.key,
				"us-east-1",
				"",
				"aws",
				nil,
			)
			err := fetcher.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestS3Fetcher_String tests String() method.
func TestS3Fetcher_String(t *testing.T) {
	fetcher := NewS3Fetcher(
		"my-bucket",
		"path/to/config.json",
		"us-east-1",
		"",
		"aws",
		nil,
	)
	str := fetcher.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
	if str != "s3: s3://my-bucket/path/to/config.json" {
		t.Errorf("Unexpected string: %s", str)
	}
}

// TestS3Fetcher_DefaultAWSPath tests default AWS path.
func TestS3Fetcher_DefaultAWSPath(t *testing.T) {
	fetcher := NewS3Fetcher(
		"my-bucket",
		"config.json",
		"",
		"",
		"", // Empty path
		nil,
	)

	if fetcher.awsPath != "aws" {
		t.Errorf("Expected default awsPath 'aws', got: %s", fetcher.awsPath)
	}
}

// TestS3Fetcher_ContextCancellation tests context cancellation support.
func TestS3Fetcher_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	fetcher := NewS3Fetcher(
		"test-bucket",
		"test-key",
		"us-east-1",
		"",
		"aws",
		nil,
	)

	// This will fail because context is canceled before execution
	_, err := fetcher.Fetch(ctx, task.NoOpReporter{})
	if err == nil {
		t.Fatal("Expected error due to cancellation")
	}
}

// Note: Full integration test for S3Fetcher.Fetch() would require AWS credentials
// or mocking the AWS CLI, which is beyond unit test scope. The task package
// already tests subprocess execution thoroughly.
