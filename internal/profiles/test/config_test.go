package test

import (
	"os"
	"strings"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestNewConfigWithSsoSingle verifies SSO single config generation.
func TestNewConfigWithSsoSingle(t *testing.T) {
	content := NewConfigWithSsoSingle()
	if content == "" {
		t.Fatal("NewConfigWithSsoSingle() returned empty content")
	}
	if !strings.Contains(content, "# START - Managed by AWS Profile Manager") {
		t.Error("Content should contain start marker")
	}
	if !strings.Contains(content, "# END - Managed by AWS Profile Manager") {
		t.Error("Content should contain end marker")
	}
	if !strings.Contains(content, "[profile ") {
		t.Error("Content should contain profile section")
	}
}

// TestNewConfigWithSsoMultiAccount verifies SSO multi-account config generation.
func TestNewConfigWithSsoMultiAccount(t *testing.T) {
	content := NewConfigWithSsoMultiAccount()
	if content == "" {
		t.Fatal("NewConfigWithSsoMultiAccount() returned empty content")
	}
	// Should have multiple profiles
	profileCount := strings.Count(content, "[profile ")
	if profileCount < 2 {
		t.Errorf("Expected multiple profiles, got %d", profileCount)
	}
}

// TestNewConfigWithIamSingle verifies IAM single config generation.
func TestNewConfigWithIamSingle(t *testing.T) {
	content := NewConfigWithIamSingle()
	if content == "" {
		t.Fatal("NewConfigWithIamSingle() returned empty content")
	}
	if !strings.Contains(content, "[profile ") {
		t.Error("Content should contain profile section")
	}
}

// TestNewConfigWithIamMulti verifies IAM multi config generation.
func TestNewConfigWithIamMulti(t *testing.T) {
	content := NewConfigWithIamMulti()
	if content == "" {
		t.Fatal("NewConfigWithIamMulti() returned empty content")
	}
	// Should have multiple profiles
	profileCount := strings.Count(content, "[profile ")
	if profileCount < 2 {
		t.Errorf("Expected multiple profiles, got %d", profileCount)
	}
}

// TestNewConfigWithAssumeRoleSingle verifies AssumeRole config generation.
func TestNewConfigWithAssumeRoleSingle(t *testing.T) {
	content := NewConfigWithAssumeRoleSingle()
	if content == "" {
		t.Fatal("NewConfigWithAssumeRoleSingle() returned empty content")
	}
	if !strings.Contains(content, "[profile ") {
		t.Error("Content should contain profile section")
	}
}

// TestNewConfigWithAllTypes verifies all profile types config generation.
func TestNewConfigWithAllTypes(t *testing.T) {
	content := NewConfigWithAllTypes()
	if content == "" {
		t.Fatal("NewConfigWithAllTypes() returned empty content")
	}
	// Should have many profiles
	profileCount := strings.Count(content, "[profile ")
	if profileCount < 3 {
		t.Errorf("Expected many profiles, got %d", profileCount)
	}
}

// TestNewConfigMixed verifies mixed config generation.
func TestNewConfigMixed(t *testing.T) {
	content := NewConfigMixed()
	if content == "" {
		t.Fatal("NewConfigMixed() returned empty content")
	}
	if !strings.Contains(content, "# START - Managed by AWS Profile Manager") {
		t.Error("Content should contain start marker")
	}
	if !strings.Contains(content, "# END - Managed by AWS Profile Manager") {
		t.Error("Content should contain end marker")
	}
}

// TestNewConfigEmpty verifies empty config generation.
func TestNewConfigEmpty(t *testing.T) {
	content := NewConfigEmpty()
	if content != "" {
		t.Error("NewConfigEmpty() should return empty content")
	}
}

// TestNewConfigLarge verifies large config generation.
func TestNewConfigLarge(t *testing.T) {
	content := NewConfigLarge()
	if content == "" {
		t.Fatal("NewConfigLarge() returned empty content")
	}
	// Should have many profiles (100+)
	profileCount := strings.Count(content, "[profile ")
	if profileCount < 100 {
		t.Errorf("Expected large number of profiles, got %d", profileCount)
	}
}

// TestWriteConfig verifies writing config to test environment.
func TestWriteConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	content := NewConfigWithSsoSingle()
	WriteConfig(t, content)

	// Verify file was written
	configPath := test.GetTestAwsConfigPath(t)
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read written config: %v", err)
	}

	if string(data) != content {
		t.Error("Written content doesn't match expected content")
	}
}
