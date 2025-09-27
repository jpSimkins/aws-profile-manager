package profiles

import (
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/test"
)

// TestReadConfigFile tests reading config file into lines
func TestReadConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name      string
		content   string
		wantLines int
		wantErr   bool
	}{
		{
			name:      "normal file",
			content:   "[profile test]\nregion = us-east-1\n",
			wantLines: 2,
			wantErr:   false,
		},
		{
			name:      "empty file",
			content:   "",
			wantLines: 0,
			wantErr:   false,
		},
		{
			name:      "single line no newline",
			content:   "[profile test]",
			wantLines: 1,
			wantErr:   false,
		},
		{
			name:      "multiple empty lines",
			content:   "\n\n\n",
			wantLines: 3,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "test.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			lines, err := readConfigFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(lines) != tt.wantLines {
				t.Errorf("readConfigFile() got %d lines, want %d", len(lines), tt.wantLines)
			}
		})
	}
}

// TestReadConfigFile_NonExistent tests reading non-existent file returns empty slice
func TestReadConfigFile_NonExistent(t *testing.T) {
	test.SetupTestEnvironment(t)

	nonExistentPath := filepath.Join(test.GetTestConfigDir(t), "does-not-exist.txt")

	lines, err := readConfigFile(nonExistentPath)
	if err != nil {
		t.Errorf("Should not error on non-existent file, got: %v", err)
	}
	if lines == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(lines) != 0 {
		t.Errorf("Expected empty slice, got %d lines", len(lines))
	}
} // TestReadFileContent tests reading file content as string
func TestReadFileContent(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "normal content",
			content: "test content\nline 2",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "content-test.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := readFileContent(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.content {
				t.Errorf("readFileContent() = %q, want %q", got, tt.content)
			}
		})
	}
}

// TestWriteConfigFile tests writing lines to config file
func TestWriteConfigFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name    string
		lines   []string
		wantErr bool
	}{
		{
			name: "normal write",
			lines: []string{
				"[profile test]",
				"region = us-east-1",
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			lines:   []string{},
			wantErr: false,
		},
		{
			name: "with empty lines",
			lines: []string{
				"[profile test]",
				"",
				"region = us-east-1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "output.txt")

			err := writeConfigFile(tmpFile, tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was written
				content, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read written file: %v", err)
				}

				// Verify content
				expected := ""
				for _, line := range tt.lines {
					expected += line + "\n"
				}

				if string(content) != expected {
					t.Errorf("File content mismatch\nGot:\n%s\nWant:\n%s", content, expected)
				}
			}
		})
	}
}

// TestWriteConfigFile_CreateDirectory tests directory creation
func TestWriteConfigFile_CreateDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)

	// File in non-existent subdirectory
	tmpFile := filepath.Join(test.GetTestConfigDir(t), "subdir", "nested", "file.txt")
	lines := []string{"test content"}

	err := writeConfigFile(tmpFile, lines)
	if err != nil {
		t.Errorf("writeConfigFile() should create directories, got error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("File was not created")
	}
}

// TestWriteFileContent tests writing content as string
func TestWriteFileContent(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "normal content",
			content: "test content\nline 2",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "write-content.txt")

			err := writeFileContent(tmpFile, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read written file: %v", err)
				}

				if string(got) != tt.content {
					t.Errorf("Content mismatch\nGot: %q\nWant: %q", got, tt.content)
				}
			}
		})
	}
}

// TestDeleteFile tests file deletion
func TestDeleteFile(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name        string
		createFile  bool
		wantDeleted bool
		wantErr     bool
	}{
		{
			name:        "delete existing file",
			createFile:  true,
			wantDeleted: true,
			wantErr:     false,
		},
		{
			name:        "delete non-existent file",
			createFile:  false,
			wantDeleted: false,
			wantErr:     false, // Should not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "delete-test.txt")

			if tt.createFile {
				if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			deleted, err := deleteFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if deleted != tt.wantDeleted {
				t.Errorf("deleteFile() deleted = %v, want %v", deleted, tt.wantDeleted)
			}

			// Verify file doesn't exist
			if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
				t.Error("File still exists after deletion")
			}
		})
	}
}

// TestFileExists tests file existence check
func TestFileExists(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name       string
		createFile bool
		createDir  bool
		want       bool
	}{
		{
			name:       "existing file",
			createFile: true,
			want:       true,
		},
		{
			name:       "non-existent file",
			createFile: false,
			want:       false,
		},
		{
			name:      "directory exists",
			createDir: true,
			want:      true, // fileExists returns true for directories too (checks os.Stat only)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := filepath.Join(test.GetTestConfigDir(t), "exists-test-"+tt.name)

			if tt.createFile {
				if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			if tt.createDir {
				if err := os.MkdirAll(testPath, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			}

			got := fileExists(testPath)
			if got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestEnsureDirectory tests directory creation
func TestEnsureDirectory(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name      string
		createDir bool
		wantErr   bool
	}{
		{
			name:      "create new directory",
			createDir: false,
			wantErr:   false,
		},
		{
			name:      "directory already exists",
			createDir: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(test.GetTestConfigDir(t), "ensuredir", tt.name)

			if tt.createDir {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("Failed to create test directory: %v", err)
				}
			}

			err := ensureDirectory(dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify directory exists
			if info, err := os.Stat(dir); err != nil || !info.IsDir() {
				t.Error("Directory was not created or is not a directory")
			}
		})
	}
}

// TestEnsureDirectory_Nested tests nested directory creation
func TestEnsureDirectory_Nested(t *testing.T) {
	test.SetupTestEnvironment(t)

	nestedDir := filepath.Join(test.GetTestConfigDir(t), "level1", "level2", "level3")

	err := ensureDirectory(nestedDir)
	if err != nil {
		t.Errorf("ensureDirectory() failed for nested path: %v", err)
	}

	// Verify all levels exist
	if info, err := os.Stat(nestedDir); err != nil || !info.IsDir() {
		t.Error("Nested directory was not created")
	}
}

// TestGetFileSize tests file size retrieval
func TestGetFileSize(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		content  string
		wantSize int64
		wantErr  bool
	}{
		{
			name:     "normal file",
			content:  "test content",
			wantSize: 12,
			wantErr:  false,
		},
		{
			name:     "empty file",
			content:  "",
			wantSize: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(test.GetTestConfigDir(t), "size-test.txt")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			size, err := getFileSize(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && size != tt.wantSize {
				t.Errorf("getFileSize() = %d, want %d", size, tt.wantSize)
			}
		})
	}
}

// TestGetFileSize_NonExistent tests non-existent file returns 0
func TestGetFileSize_NonExistent(t *testing.T) {
	test.SetupTestEnvironment(t)

	nonExistentPath := filepath.Join(test.GetTestConfigDir(t), "does-not-exist.txt")

	size, err := getFileSize(nonExistentPath)
	if err != nil {
		t.Errorf("Should not error on non-existent file, got: %v", err)
	}
	if size != 0 {
		t.Errorf("Expected size 0 for non-existent file, got: %d", size)
	}
}
