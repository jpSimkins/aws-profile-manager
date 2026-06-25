package security

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestReadFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "config.json")

	if err := os.WriteFile(filePath, []byte("{}"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	data, err := ReadFile(filePath, ReadOptions{})
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "{}" {
		t.Fatalf("unexpected file content: %s", string(data))
	}
}

func TestValidateReadPath_RejectsPathOutsideBaseDir(t *testing.T) {
	test.SetupTestEnvironment(t)

	baseDir := t.TempDir()
	outsideDir := t.TempDir()
	outsidePath := filepath.Join(outsideDir, "config.json")

	if err := os.WriteFile(outsidePath, []byte("{}"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ValidateReadPath(outsidePath, ReadOptions{BaseDir: baseDir})
	if err == nil {
		t.Fatal("expected path outside base dir to be rejected")
	}
}

func TestValidateReadPath_RejectsInvalidExtension(t *testing.T) {
	test.SetupTestEnvironment(t)

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "config.txt")

	if err := os.WriteFile(filePath, []byte("{}"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := ValidateReadPath(filePath, ReadOptions{AllowedExtensions: []string{".json"}})
	if err == nil {
		t.Fatal("expected non-json extension to be rejected")
	}
}

func TestValidateReadPath_RejectsDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)

	tempDir := t.TempDir()

	_, err := ValidateReadPath(tempDir, ReadOptions{})
	if err == nil {
		t.Fatal("expected directory path to be rejected")
	}
}

func TestValidateReadPath_RejectsSymlinkEscape(t *testing.T) {
	test.SetupTestEnvironment(t)

	baseDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.json")
	linkPath := filepath.Join(baseDir, "link.json")

	if err := os.WriteFile(outsideFile, []byte("{}"), 0600); err != nil {
		t.Fatalf("failed to write outside file: %v", err)
	}

	if err := os.Symlink(outsideFile, linkPath); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	_, err := ValidateReadPath(linkPath, ReadOptions{BaseDir: baseDir, AllowedExtensions: []string{".json"}})
	if err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}
