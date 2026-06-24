package sync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aws-profile-manager/internal/task"
)

// GitFetcher fetches configuration from a Git repository.
//
// This fetcher clones or pulls from a Git repository and reads the configuration
// file. It uses the task package's SubprocessTask for git command execution,
// providing consistent subprocess handling, context cancellation, and progress
// reporting.
//
// Features:
//   - Clone on first fetch, pull on subsequent fetches
//   - Configurable branch/ref
//   - SSH and HTTPS authentication support
//   - Shallow clone for performance
//   - Automatic cleanup on context cancellation
//
// Authentication Methods:
//   - SSH: Uses SSH keys from ~/.ssh (default git behavior)
//   - HTTPS: Uses git credential helpers or embedded credentials in URL
//   - Token: Can be embedded in URL (https://token@github.com/...)
//
// Example URLs:
//   - HTTPS: https://github.com/org/repo.git
//   - SSH: git@github.com:org/repo.git
//   - HTTPS with token: https://TOKEN@github.com/org/repo.git
type GitFetcher struct {
	repoURL  string            // Git repository URL (SSH or HTTPS)
	branch   string            // Branch or ref to checkout
	filePath string            // Path to config file within repo (e.g., "config.json")
	workDir  string            // Working directory for git operations
	gitPath  string            // Path to git executable
	gitEnv   map[string]string // Environment variables for git commands
}

// NewGitFetcher creates a new Git fetcher.
//
// The workDir parameter specifies where the repository will be cloned. If empty,
// a temporary directory will be created. The gitPath parameter specifies the
// git executable path (defaults to "git" if empty).
//
// Parameters:
//   - repoURL: Git repository URL (SSH or HTTPS)
//   - branch: Branch or ref to checkout (e.g., "main", "master", "v1.0.0")
//   - filePath: Path to config file within repo (e.g., "config.json", "configs/aws.json")
//   - workDir: Working directory for git operations (empty for temp dir)
//   - gitPath: Path to git executable (empty for "git")
//   - gitEnv: Environment variables for git commands (nil for none)
//
// Returns:
//   - *GitFetcher: Configured Git fetcher instance
func NewGitFetcher(repoURL, branch, filePath, workDir, gitPath string, gitEnv map[string]string) *GitFetcher {
	if gitPath == "" {
		gitPath = "git"
	}
	if branch == "" {
		branch = "main"
	}
	return &GitFetcher{
		repoURL:  repoURL,
		branch:   branch,
		filePath: filePath,
		workDir:  workDir,
		gitPath:  gitPath,
		gitEnv:   gitEnv,
	}
}

// Fetch retrieves the configuration from the Git repository.
//
// This method clones the repository on first fetch, then pulls updates on
// subsequent fetches. It performs a shallow clone (depth=1) for performance.
// After updating the repository, it reads the specified file and returns
// its contents.
//
// The method uses task.SubprocessTask for git command execution, which provides:
//   - Automatic process killing on context cancellation
//   - Consistent subprocess handling
//   - Progress reporting
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Progress reporter for status updates
//
// Returns:
//   - []byte: Configuration file contents
//   - error: Any error encountered (clone failed, file not found, etc.)
func (g *GitFetcher) Fetch(ctx context.Context, reporter task.Reporter) ([]byte, error) {
	// Check for cancellation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	reporter.ReportStatus(fmt.Sprintf("Fetching from Git: %s", g.repoURL))

	// Determine repository directory
	repoDir := g.workDir
	if repoDir == "" {
		// Use temp directory
		tempDir, err := os.MkdirTemp("", "aws-profile-manager-git-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
		repoDir = tempDir
		// Note: We don't defer os.RemoveAll here because the clone might be reused
	}

	repoPath := filepath.Join(repoDir, "repo")

	// Check if repo already exists
	gitDir := filepath.Join(repoPath, ".git")
	repoExists := false
	if _, err := os.Stat(gitDir); err == nil {
		repoExists = true
	}

	if repoExists {
		reporter.ReportStatus("Repository exists, pulling updates...")
		if err := g.gitPull(ctx, repoPath, reporter); err != nil {
			return nil, fmt.Errorf("failed to pull updates: %w", err)
		}
	} else {
		reporter.ReportStatus("Cloning repository...")
		if err := g.gitClone(ctx, repoPath, reporter); err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
	}

	// Checkout specified branch
	reporter.ReportStatus(fmt.Sprintf("Checking out branch: %s", g.branch))
	if err := g.gitCheckout(ctx, repoPath, reporter); err != nil {
		return nil, fmt.Errorf("failed to checkout branch: %w", err)
	}

	// Read the config file
	configPath := filepath.Join(repoPath, g.filePath)
	reporter.ReportStatus(fmt.Sprintf("Reading config file: %s", g.filePath))

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", g.filePath, err)
	}

	reporter.ReportStatus(fmt.Sprintf("Successfully fetched %d bytes from Git", len(data)))
	return data, nil
}

// gitClone clones the repository.
func (g *GitFetcher) gitClone(ctx context.Context, repoPath string, reporter task.Reporter) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(repoPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build clone command: git clone --depth 1 --branch <branch> <url> <path>
	args := []string{
		"clone",
		"--depth", "1", // Shallow clone for performance
		"--branch", g.branch, // Specific branch
		"--single-branch", // Only fetch one branch
		g.repoURL,         // Repository URL
		repoPath,          // Destination path
	}

	// Create subprocess task
	gitTask := &task.SubprocessTask{
		Name:    g.gitPath,
		Args:    args,
		Env:     g.gitEnv,
		Timeout: 5 * time.Minute, // 5 minute timeout for clone
	}

	// Execute (subprocess killed on context cancel)
	_, err := task.Run(ctx, gitTask, reporter)
	if err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// gitPull pulls updates from the repository.
func (g *GitFetcher) gitPull(ctx context.Context, repoPath string, reporter task.Reporter) error {
	// Build pull command: git -C <path> pull --ff-only
	args := []string{
		"-C", repoPath, // Work in repo directory
		"pull",
		"--ff-only", // Fast-forward only (safer)
	}

	// Create subprocess task
	gitTask := &task.SubprocessTask{
		Name:    g.gitPath,
		Args:    args,
		Env:     g.gitEnv,
		Timeout: 2 * time.Minute, // 2 minute timeout for pull
	}

	// Execute (subprocess killed on context cancel)
	_, err := task.Run(ctx, gitTask, reporter)
	if err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

// gitCheckout checks out the specified branch.
func (g *GitFetcher) gitCheckout(ctx context.Context, repoPath string, reporter task.Reporter) error {
	// Build checkout command: git -C <path> checkout <branch>
	args := []string{
		"-C", repoPath, // Work in repo directory
		"checkout",
		g.branch, // Branch to checkout
	}

	// Create subprocess task
	gitTask := &task.SubprocessTask{
		Name:    g.gitPath,
		Args:    args,
		Env:     g.gitEnv,
		Timeout: 30 * time.Second, // 30 second timeout for checkout
	}

	// Execute (subprocess killed on context cancel)
	_, err := task.Run(ctx, gitTask, reporter)
	if err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}

	return nil
}

// String returns a human-readable description of the fetcher.
func (g *GitFetcher) String() string {
	return fmt.Sprintf("git: %s (branch: %s, file: %s)", g.repoURL, g.branch, g.filePath)
}

// Validate checks if the Git configuration is valid.
//
// Validation checks:
//   - Repository URL is not empty
//   - Branch name is not empty
//   - File path is not empty
//
// Returns:
//   - error: Validation error if configuration is invalid, nil if valid
func (g *GitFetcher) Validate() error {
	if g.repoURL == "" {
		return fmt.Errorf("git repository URL is required")
	}
	if g.branch == "" {
		return fmt.Errorf("git branch is required")
	}
	if g.filePath == "" {
		return fmt.Errorf("git file path is required")
	}
	return nil
}
