package sync

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestNewGitFetcher tests Git fetcher creation.
func TestNewGitFetcher(t *testing.T) {
	fetcher := NewGitFetcher(
		"https://github.com/org/repo.git",
		"main",
		"config.json",
		"",
		"",
		nil,
	)

	if fetcher.repoURL != "https://github.com/org/repo.git" {
		t.Errorf("Expected repoURL 'https://github.com/org/repo.git', got %s", fetcher.repoURL)
	}

	if fetcher.branch != "main" {
		t.Errorf("Expected branch 'main', got %s", fetcher.branch)
	}

	if fetcher.filePath != "config.json" {
		t.Errorf("Expected filePath 'config.json', got %s", fetcher.filePath)
	}

	if fetcher.gitPath != "git" {
		t.Errorf("Expected default gitPath 'git', got %s", fetcher.gitPath)
	}
}

// TestGitFetcher_Validate tests Git fetcher validation.
func TestGitFetcher_Validate(t *testing.T) {
	tests := []struct {
		name      string
		fetcher   *GitFetcher
		expectErr bool
	}{
		{
			name: "valid config",
			fetcher: NewGitFetcher(
				"https://github.com/org/repo.git",
				"main",
				"config.json",
				"",
				"",
				nil,
			),
			expectErr: false,
		},
		{
			name: "empty repo URL",
			fetcher: NewGitFetcher(
				"",
				"main",
				"config.json",
				"",
				"",
				nil,
			),
			expectErr: true,
		},
		{
			name: "empty branch (gets default)",
			fetcher: NewGitFetcher(
				"https://github.com/org/repo.git",
				"",
				"config.json",
				"",
				"",
				nil,
			),
			expectErr: false, // Gets "main" as default
		},
		{
			name: "empty file path",
			fetcher: NewGitFetcher(
				"https://github.com/org/repo.git",
				"main",
				"",
				"",
				"",
				nil,
			),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fetcher.Validate()
			if tt.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestGitFetcher_String tests string representation.
func TestGitFetcher_String(t *testing.T) {
	fetcher := NewGitFetcher(
		"https://github.com/org/repo.git",
		"main",
		"config.json",
		"",
		"",
		nil,
	)

	str := fetcher.String()
	if str != "git: https://github.com/org/repo.git (branch: main, file: config.json)" {
		t.Errorf("Unexpected string representation: %s", str)
	}
}

// TestGitFetcher_LocalRepo tests fetching from a local git repo.
func TestGitFetcher_LocalRepo(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create a test git repository
	repoDir := filepath.Join(test.GetTestConfigDir(t), "test-repo")
	if err := os.MkdirAll(repoDir, 0700); err != nil {
		t.Fatalf("Failed to create repo directory: %v", err)
	}

	// Initialize git repo
	gitTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "init"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	_, err := task.Run(context.Background(), gitTask, task.NoOpReporter{})
	if err != nil {
		t.Skipf("Git not available or failed to init: %v", err)
	}

	// Detect the default branch for this git installation (main/master).
	branchTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "symbolic-ref", "--short", "HEAD"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	branchResult, err := task.Run(context.Background(), branchTask, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to detect default branch: %v", err)
	}
	defaultBranch := strings.TrimSpace(string(branchResult.Output))
	if defaultBranch == "" {
		t.Fatal("Detected empty default branch")
	}

	// Create test config file
	configContent := []byte(`{"version":"1.0","managed":{"organizations":{}}}`)
	configPath := filepath.Join(repoDir, "config.json")
	if err := os.WriteFile(configPath, configContent, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Git add and commit
	addTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "add", "config.json"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	if _, err := task.Run(context.Background(), addTask, task.NoOpReporter{}); err != nil {
		t.Fatalf("Failed to add config file: %v", err)
	}

	// Configure git user for commit (local to test repo only)
	userTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "config", "--local", "user.email", "test@example.com"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	if _, err := task.Run(context.Background(), userTask, task.NoOpReporter{}); err != nil {
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	nameTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "config", "--local", "user.name", "Test User"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	if _, err := task.Run(context.Background(), nameTask, task.NoOpReporter{}); err != nil {
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	commitTask := &task.SubprocessTask{
		Name: "git",
		Args: []string{"-C", repoDir, "commit", "-m", "Initial commit"},
		Env:  map[string]string{"GIT_CONFIG_GLOBAL": "/dev/null"},
	}
	if _, err := task.Run(context.Background(), commitTask, task.NoOpReporter{}); err != nil {
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	// Create work directory for fetcher
	workDir := filepath.Join(test.GetTestConfigDir(t), "git-work")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	// Create fetcher using local repo path
	fetcher := NewGitFetcher(
		repoDir, // Local repo path
		defaultBranch,
		"config.json",
		workDir,
		"git",
		nil,
	)

	// Validate
	if err := fetcher.Validate(); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Fetch
	data, err := fetcher.Fetch(context.Background(), task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify content
	if string(data) != string(configContent) {
		t.Errorf("Expected content %s, got %s", string(configContent), string(data))
	}
}

// TestGitFetcher_ContextCancellation tests context cancellation.
func TestGitFetcher_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	workDir := filepath.Join(test.GetTestConfigDir(t), "git-work")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewGitFetcher(
		"https://github.com/very-large-repo/huge.git", // Non-existent repo
		"main",
		"config.json",
		workDir,
		"git",
		nil,
	)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Fetch should fail with context error
	_, err := fetcher.Fetch(ctx, task.NoOpReporter{})
	if err == nil {
		t.Error("Expected error with cancelled context")
	}
}
