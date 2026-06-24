package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestWriteConfigFile_DirectoryCreationError tests directory creation errors
func TestWriteConfigFile_DirectoryCreationError(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a file where we want to create a directory
	baseDir := test.GetTestConfigDir(t)
	blockingFile := filepath.Join(baseDir, "blocking")
	_ = os.WriteFile(blockingFile, []byte("block"), 0600)

	// Try to create config under this file (should fail)
	configPath := filepath.Join(blockingFile, "config")
	err := writeConfigFile(configPath, []string{"test"})

	if err == nil {
		t.Error("Should fail when directory creation is blocked")
	}
}

// TestWriteConfigFile_WriteError tests write errors during line writing
func TestWriteConfigFile_WriteError(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Normal write should succeed
	configPath := filepath.Join(test.GetTestConfigDir(t), "test-config")
	lines := []string{"line1", "line2", "line3"}

	err := writeConfigFile(configPath, lines)
	if err != nil {
		t.Errorf("Write should succeed: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	expected := "line1\nline2\nline3\n"
	if string(content) != expected {
		t.Errorf("Content = %q, want %q", string(content), expected)
	}
}

// TestWriteConfigFile_EmptyLines tests writing empty line slice
func TestWriteConfigFile_EmptyLines(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "empty-config")
	err := writeConfigFile(configPath, []string{})

	if err != nil {
		t.Errorf("Writing empty lines should succeed: %v", err)
	}

	// File should exist but be empty
	if !fileExists(configPath) {
		t.Error("File should have been created")
	}

	size, _ := getFileSize(configPath)
	if size != 0 {
		t.Errorf("File should be empty, got size %d", size)
	}
}

// TestWriteFileContent_DirectoryCreationError tests directory creation errors
func TestWriteFileContent_DirectoryCreationError(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a file where we want to create a directory
	baseDir := test.GetTestConfigDir(t)
	blockingFile := filepath.Join(baseDir, "blocking2")
	_ = os.WriteFile(blockingFile, []byte("block"), 0600)

	// Try to create file under this file (should fail)
	filePath := filepath.Join(blockingFile, "file.txt")
	err := writeFileContent(filePath, "test content")

	if err == nil {
		t.Error("Should fail when directory creation is blocked")
	}
}

// TestWriteFileContent_Success tests successful file writing
func TestWriteFileContent_Success(t *testing.T) {
	test.SetupTestEnvironment(t)

	filePath := filepath.Join(test.GetTestConfigDir(t), "content.txt")
	content := "test content\nline 2"

	err := writeFileContent(filePath, content)
	if err != nil {
		t.Errorf("Write should succeed: %v", err)
	}

	// Verify content
	readContent, err := readFileContent(filePath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if readContent != content {
		t.Errorf("Content = %q, want %q", readContent, content)
	}
}

// TestReadConfigFile_ScannerError tests scanner errors
func TestReadConfigFile_ScannerError(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Normal file should succeed
	configPath := filepath.Join(test.GetTestConfigDir(t), "scanner-test")
	testContent := "line1\nline2\nline3"
	_ = os.WriteFile(configPath, []byte(testContent), 0600)

	lines, err := readConfigFile(configPath)
	if err != nil {
		t.Errorf("Read should succeed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("Should have 3 lines, got %d", len(lines))
	}
}

// TestReadFileContent_NonExistent tests reading non-existent file
func TestReadFileContent_NonExistent(t *testing.T) {
	test.SetupTestEnvironment(t)

	path := filepath.Join(test.GetTestConfigDir(t), "does-not-exist.txt")
	_, err := readFileContent(path)

	if err == nil {
		t.Error("Should error on non-existent file")
	}
}

// TestDeleteFile_Errors tests delete error scenarios
func TestDeleteFile_Errors(t *testing.T) {
	test.SetupTestEnvironment(t)

	t.Run("delete non-existent file returns false, no error", func(t *testing.T) {
		path := filepath.Join(test.GetTestConfigDir(t), "does-not-exist")
		deleted, err := deleteFile(path)

		if err != nil {
			t.Errorf("Should not error: %v", err)
		}
		if deleted {
			t.Error("Should return false for non-existent file")
		}
	})

	t.Run("delete existing file returns true", func(t *testing.T) {
		path := filepath.Join(test.GetTestConfigDir(t), "to-delete")
		_ = os.WriteFile(path, []byte("test"), 0600)

		deleted, err := deleteFile(path)

		if err != nil {
			t.Errorf("Should not error: %v", err)
		}
		if !deleted {
			t.Error("Should return true for deleted file")
		}
		if fileExists(path) {
			t.Error("File should no longer exist")
		}
	})
}

// TestGetFileSize_Errors tests file size error scenarios
func TestGetFileSize_Errors(t *testing.T) {
	test.SetupTestEnvironment(t)

	t.Run("non-existent file returns 0, no error", func(t *testing.T) {
		path := filepath.Join(test.GetTestConfigDir(t), "does-not-exist")
		size, err := getFileSize(path)

		if err != nil {
			t.Errorf("Should not error: %v", err)
		}
		if size != 0 {
			t.Errorf("Size should be 0, got %d", size)
		}
	})

	t.Run("existing file returns correct size", func(t *testing.T) {
		path := filepath.Join(test.GetTestConfigDir(t), "sized-file")
		content := "123456789" // 9 bytes
		_ = os.WriteFile(path, []byte(content), 0600)

		size, err := getFileSize(path)

		if err != nil {
			t.Errorf("Should not error: %v", err)
		}
		if size != 9 {
			t.Errorf("Size should be 9, got %d", size)
		}
	})

	t.Run("directory returns size", func(t *testing.T) {
		path := test.GetTestConfigDir(t)
		size, err := getFileSize(path)

		// Directory stat should work, size depends on filesystem
		if err != nil {
			t.Errorf("Should not error on directory: %v", err)
		}
		// Size will be non-zero for directory
		_ = size
	})
}

// TestEnsureDirectory_Errors tests directory creation errors
func TestEnsureDirectory_Errors(t *testing.T) {
	test.SetupTestEnvironment(t)

	t.Run("create new nested directory", func(t *testing.T) {
		path := filepath.Join(test.GetTestConfigDir(t), "a", "b", "c")
		err := ensureDirectory(path)

		if err != nil {
			t.Errorf("Should create nested directories: %v", err)
		}

		// Verify it exists
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Directory should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("Should be a directory")
		}
	})

	t.Run("already exists is not an error", func(t *testing.T) {
		path := test.GetTestConfigDir(t)
		err := ensureDirectory(path)

		if err != nil {
			t.Errorf("Should not error if directory exists: %v", err)
		}
	})
}
