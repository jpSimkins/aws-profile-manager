//go:build e2e

package sync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestGitFetcher_RealGitHTTPS tests fetching from a real Git repository via HTTPS.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_GIT_REPO_URL="https://github.com/org/repo.git"
//	export SYNC_E2E_GIT_BRANCH="main"  # Optional, defaults to "main"
//	export SYNC_E2E_GIT_FILE_PATH="config.json"
//	go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitHTTPS
//
// Requirements:
//   - SYNC_E2E_GIT_REPO_URL: Git repository URL (HTTPS)
//   - SYNC_E2E_GIT_FILE_PATH: Path to config file within repo
//   - SYNC_E2E_GIT_BRANCH: Branch/tag/commit (optional, defaults to "main")
//   - Public repo or token in URL: https://token@github.com/org/repo.git
//   - Git installed
func TestGitFetcher_RealGitHTTPS(t *testing.T) {
	test.SetupTestEnvironment(t)

	repoURL := os.Getenv("SYNC_E2E_GIT_REPO_URL")
	filePath := os.Getenv("SYNC_E2E_GIT_FILE_PATH")

	if repoURL == "" || filePath == "" {
		t.Skip("SYNC_E2E_GIT_REPO_URL or SYNC_E2E_GIT_FILE_PATH not set - skipping E2E test")
	}

	branch := os.Getenv("SYNC_E2E_GIT_BRANCH")
	if branch == "" {
		branch = "main"
	}

	t.Logf("Testing Git fetch from: %s @ %s", repoURL, branch)
	t.Logf("Fetching file: %s", filePath)

	workDir := filepath.Join(test.GetTestConfigDir(t), "git-e2e")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewGitFetcher(
		repoURL,
		branch,
		filePath,
		workDir,
		"git",
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

	t.Logf("Fetched %d bytes successfully from Git", len(data))
}

// TestGitFetcher_RealGitSSH tests fetching from a real Git repository via SSH.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_GIT_REPO_URL="git@github.com:org/repo.git"
//	export SYNC_E2E_GIT_FILE_PATH="config.json"
//	export SYNC_E2E_GIT_BRANCH="main"  # Optional
//	go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitSSH
//
// Requirements:
//   - SYNC_E2E_GIT_REPO_URL: Git repository URL (SSH)
//   - SYNC_E2E_GIT_FILE_PATH: Path to config file within repo
//   - Valid SSH key in ~/.ssh (must be configured for git)
//   - SSH key must have access to the repository
//   - Git installed
func TestGitFetcher_RealGitSSH(t *testing.T) {
	test.SetupTestEnvironment(t)

	repoURL := os.Getenv("SYNC_E2E_GIT_REPO_URL")
	filePath := os.Getenv("SYNC_E2E_GIT_FILE_PATH")

	if repoURL == "" || filePath == "" {
		t.Skip("SYNC_E2E_GIT_REPO_URL or SYNC_E2E_GIT_FILE_PATH not set - skipping E2E test")
	}

	// Verify SSH URL format
	if len(repoURL) < 4 || repoURL[:4] != "git@" {
		t.Skipf("SYNC_E2E_GIT_REPO_URL is not an SSH URL (expected git@...): %s", repoURL)
	}

	branch := os.Getenv("SYNC_E2E_GIT_BRANCH")
	if branch == "" {
		branch = "main"
	}

	t.Logf("Testing Git fetch via SSH from: %s @ %s", repoURL, branch)
	t.Logf("Fetching file: %s", filePath)

	workDir := filepath.Join(test.GetTestConfigDir(t), "git-e2e-ssh")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewGitFetcher(
		repoURL,
		branch,
		filePath,
		workDir,
		"git",
		nil,
	)

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed (ensure SSH key is configured and has access): %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully via SSH", len(data))
}

// TestGitFetcher_RealGitTag tests fetching a specific tag from a Git repository.
//
// This test is SKIPPED by default. To run:
//
//	export SYNC_E2E_GIT_REPO_URL="https://github.com/org/repo.git"
//	export SYNC_E2E_GIT_BRANCH="v1.0.0"  # Tag name
//	export SYNC_E2E_GIT_FILE_PATH="config.json"
//	go test -tags=e2e -v ./internal/sync -run TestGitFetcher_RealGitTag
//
// Requirements:
//   - SYNC_E2E_GIT_REPO_URL: Git repository URL
//   - SYNC_E2E_GIT_BRANCH: Tag name (e.g., "v1.0.0")
//   - SYNC_E2E_GIT_FILE_PATH: Path to config file within repo
func TestGitFetcher_RealGitTag(t *testing.T) {
	test.SetupTestEnvironment(t)

	repoURL := os.Getenv("SYNC_E2E_GIT_REPO_URL")
	tag := os.Getenv("SYNC_E2E_GIT_BRANCH")
	filePath := os.Getenv("SYNC_E2E_GIT_FILE_PATH")

	if repoURL == "" || tag == "" || filePath == "" {
		t.Skip("SYNC_E2E_GIT_REPO_URL, SYNC_E2E_GIT_BRANCH, or SYNC_E2E_GIT_FILE_PATH not set - skipping E2E test")
	}

	t.Logf("Testing Git fetch of tag: %s from %s", tag, repoURL)
	t.Logf("Fetching file: %s", filePath)

	workDir := filepath.Join(test.GetTestConfigDir(t), "git-e2e-tag")
	if err := os.MkdirAll(workDir, 0700); err != nil {
		t.Fatalf("Failed to create work directory: %v", err)
	}

	fetcher := NewGitFetcher(
		repoURL,
		tag,
		filePath,
		workDir,
		"git",
		nil,
	)

	ctx := context.Background()
	data, err := fetcher.Fetch(ctx, task.CliReporter{})
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	t.Logf("Fetched %d bytes successfully from tag '%s'", len(data), tag)
}
