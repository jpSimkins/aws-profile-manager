package sync

import (
	"net"
	"testing"
	"time"
)

// TestValidateConfig_HTTP tests HTTP configuration validation.
func TestValidateConfig_HTTP(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SyncConfig
		wantErr bool
	}{
		{
			name: "valid http config",
			cfg: SyncConfig{
				Strategy: StrategyHTTP,
				HTTPUrl:  "https://example.com/config.json",
			},
			wantErr: false,
		},
		{
			name: "valid http config with bypass",
			cfg: SyncConfig{
				Strategy:       StrategyHTTP,
				HTTPUrl:        "http://192.168.1.1/config.json",
				HTTPBypassSSRF: true,
			},
			wantErr: false,
		},
		{
			name: "missing url",
			cfg: SyncConfig{
				Strategy: StrategyHTTP,
				HTTPUrl:  "",
			},
			wantErr: true,
		},
		{
			name: "invalid url scheme",
			cfg: SyncConfig{
				Strategy: StrategyHTTP,
				HTTPUrl:  "ftp://example.com/config.json",
			},
			wantErr: true,
		},
		{
			name: "localhost blocked",
			cfg: SyncConfig{
				Strategy: StrategyHTTP,
				HTTPUrl:  "http://localhost:8080/config.json",
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			cfg: SyncConfig{
				Strategy:    StrategyHTTP,
				HTTPUrl:     "https://example.com/config.json",
				HTTPTimeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative retries",
			cfg: SyncConfig{
				Strategy:    StrategyHTTP,
				HTTPUrl:     "https://example.com/config.json",
				HTTPRetries: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateConfig_S3 tests S3 configuration validation.
func TestValidateConfig_S3(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SyncConfig
		wantErr bool
	}{
		{
			name: "valid s3 config",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Bucket: "my-bucket",
				S3Key:    "config.json",
				S3Region: "us-east-1",
			},
			wantErr: false,
		},
		{
			name: "valid s3 config without region",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Bucket: "my-bucket",
				S3Key:    "config.json",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Key:    "config.json",
			},
			wantErr: true,
		},
		{
			name: "missing key",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Bucket: "my-bucket",
			},
			wantErr: true,
		},
		{
			name: "invalid bucket name",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Bucket: "INVALID",
				S3Key:    "config.json",
			},
			wantErr: true,
		},
		{
			name: "invalid region",
			cfg: SyncConfig{
				Strategy: StrategyS3,
				S3Bucket: "my-bucket",
				S3Key:    "config.json",
				S3Region: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateConfig_Local tests local configuration validation.
func TestValidateConfig_Local(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SyncConfig
		wantErr bool
	}{
		{
			name: "valid local config",
			cfg: SyncConfig{
				Strategy:  StrategyLocal,
				LocalPath: "/path/to/config.json",
			},
			wantErr: false,
		},
		{
			name: "missing path",
			cfg: SyncConfig{
				Strategy: StrategyLocal,
			},
			wantErr: true,
		},
		{
			name: "directory traversal attempt",
			cfg: SyncConfig{
				Strategy:  StrategyLocal,
				LocalPath: "/path/../../../etc/passwd",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateConfig_UnknownStrategy tests unknown strategy rejection.
func TestValidateConfig_UnknownStrategy(t *testing.T) {
	cfg := SyncConfig{
		Strategy: Strategy("unknown"),
	}

	err := ValidateConfig(cfg)
	if err == nil {
		t.Fatal("Expected error for unknown strategy")
	}
}

// TestValidateS3BucketName tests S3 bucket name validation.
func TestValidateS3BucketName(t *testing.T) {
	tests := []struct {
		name    string
		bucket  string
		wantErr bool
	}{
		{"valid bucket", "my-bucket", false},
		{"valid with numbers", "my-bucket-123", false},
		{"valid with periods", "my.bucket.name", false},
		{"too short", "ab", true},
		{"too long", "this-is-a-very-long-bucket-name-that-exceeds-the-sixty-three-character-limit", true},
		{"uppercase", "MyBucket", true},
		{"starts with hyphen", "-my-bucket", true},
		{"ends with hyphen", "my-bucket-", true},
		{"consecutive periods", "my..bucket", true},
		{"invalid character", "my_bucket", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateS3BucketName(tt.bucket)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateS3BucketName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateAWSRegion tests AWS region validation.
func TestValidateAWSRegion(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		wantErr bool
	}{
		{"valid us-east-1", "us-east-1", false},
		{"valid eu-west-2", "eu-west-2", false},
		{"valid ap-southeast-1", "ap-southeast-1", false},
		{"valid us-gov-east-1", "us-gov-east-1", false},
		{"invalid format", "invalid", true},
		{"too few parts", "us-east", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAWSRegion(tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAWSRegion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsPrivateIP tests private IP detection.
func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"public ip", "8.8.8.8", false},
		{"private 10.x", "10.0.0.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"loopback", "127.0.0.1", true},
		{"link-local", "169.254.1.1", true},
		{"ipv6 loopback", "::1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			got := isPrivateIP(ip)
			if got != tt.want {
				t.Errorf("isPrivateIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
