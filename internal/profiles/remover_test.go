package profiles

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestNewRemover tests Remover constructor
func TestNewRemover(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	remover := NewRemover(config)

	if remover == nil {
		t.Fatal("NewRemover should not return nil")
	}
}

// TestRemover_Remove_Success tests successful profile removal
func TestRemover_Remove_Success(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath

	// First install some profiles
	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()
	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to install profiles: %v", err)
	}

	// Verify profiles exist
	if !fileExists(configPath) {
		t.Fatal("Config file should exist after install")
	}

	// Remove profiles
	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if result.ProfilesRemoved == 0 {
		t.Error("Should have removed profiles")
	}

	// Verify config no longer has markers
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "# START - Test") {
		t.Error("Config should not contain markers after removal")
	}
}

// TestRemover_Remove_NoManagedSection tests removing when no managed section exists
func TestRemover_Remove_NoManagedSection(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath

	// Create config without managed section
	configContent := `[profile personal]
region = us-east-1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Remove should succeed (no-op)
	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Remove should succeed when no managed section exists, got error: %v", err)
	}

	if result.ProfilesRemoved != 0 {
		t.Errorf("ProfilesRemoved should be 0, got %d", result.ProfilesRemoved)
	}

	// Verify personal profiles preserved
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[profile personal]") {
		t.Error("Personal profiles should be preserved")
	}
}

// TestRemover_Remove_NonExistentConfig tests removing from non-existent config
func TestRemover_Remove_NonExistentConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	configPath := filepath.Join(test.GetTestConfigDir(t), "nonexistent-config")
	config := Config{
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		ConfigPath:          configPath,
		StartMarker:         "# START - Test",
		EndMarker:           "# END - Test",
	}

	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})

	// Should succeed (nothing to remove)
	if err != nil {
		t.Fatalf("Remove should succeed on non-existent config, got error: %v", err)
	}

	if result.ProfilesRemoved != 0 {
		t.Errorf("ProfilesRemoved should be 0, got %d", result.ProfilesRemoved)
	}
}

// TestRemover_Remove_DryRun tests dry-run mode
func TestRemover_Remove_DryRun(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath

	// Install profiles first
	installer := NewInstaller(config)
	schema := schematest.NewManagedSsoSingle()
	_, err := installer.Install(context.Background(), InstallOptions{
		Schema: schema,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Failed to install profiles: %v", err)
	}

	// Read original content
	originalContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Dry run removal
	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{
		DryRun: true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Should report profiles that would be removed
	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !result.RemovedConfig {
		t.Error("Dry-run should report managed config would be removed")
	}

	// Config should not be modified
	currentContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if string(currentContent) != string(originalContent) {
		t.Error("Config should not be modified in dry-run mode")
	}
}

// TestRemover_Remove_PreservesPersonalProfiles tests that personal profiles are preserved
func TestRemover_Remove_PreservesPersonalProfiles(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	configPath := config.ConfigPath

	// Create config with managed and personal profiles
	configContent := `[profile personal-above]
region = us-east-1

# START - Test
[profile work-dev]
region = us-west-2
# END - Test

[profile personal-below]
region = us-west-1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Remove managed section
	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if result.ProfilesRemoved == 0 {
		t.Error("Should have removed profiles")
	}

	// Verify personal profiles preserved
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "personal-above") {
		t.Error("Above personal profiles should be preserved")
	}
	if !strings.Contains(contentStr, "personal-below") {
		t.Error("Below personal profiles should be preserved")
	}
	if strings.Contains(contentStr, "work-dev") {
		t.Error("Managed profiles should be removed")
	}
	if strings.Contains(contentStr, "# START - Test") {
		t.Error("Markers should be removed")
	}
}

// TestRemover_Remove_ContextCancellation tests context cancellation
func TestRemover_Remove_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Install profiles first
	installer := NewInstaller(config)
	_, _ = installer.Install(context.Background(), InstallOptions{
		Schema: schematest.NewManagedSsoSingle(),
	}, task.NoOpReporter{})

	// Cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	remover := NewRemover(config)
	_, err := remover.Remove(ctx, RemoveOptions{}, task.NoOpReporter{})

	if err == nil {
		t.Error("Should return error when context is cancelled")
	}
}

// TestRemover_Remove_WithCheatsheet tests removing cheatsheet
func TestRemover_Remove_WithCheatsheet(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)

	// Install with cheatsheet
	installer := NewInstaller(config)
	_, _ = installer.Install(context.Background(), InstallOptions{
		Schema:             schematest.NewManagedSsoSingle(),
		GenerateCheatSheet: true,
	}, task.NoOpReporter{})

	// Remove including cheatsheet
	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{
		RemoveCheatSheet: true,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Check result
	_ = result
}

// TestRemover_Remove_WithCustomCheatSheetPath tests cleanup with custom path.
func TestRemover_Remove_WithCustomCheatSheetPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	customCheatSheetPath := filepath.Join(test.GetTestDesktopDir(t), "custom-cheat-sheet.md")

	installer := NewInstaller(config)
	_, err := installer.Install(context.Background(), InstallOptions{
		Schema:             schematest.NewManagedSsoSingle(),
		GenerateCheatSheet: true,
		CheatSheetPath:     customCheatSheetPath,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if !fileExists(customCheatSheetPath) {
		t.Fatalf("Expected custom cheat sheet at %s", customCheatSheetPath)
	}

	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{
		RemoveCheatSheet: true,
		CheatSheetPath:   customCheatSheetPath,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if !result.RemovedCheatSheet {
		t.Error("Expected custom cheat sheet to be removed")
	}

	if fileExists(customCheatSheetPath) {
		t.Error("Custom cheat sheet should be removed")
	}
}

// TestRemover_Remove_DryRun_WithCheatSheet tests dry-run cheat sheet detection.
func TestRemover_Remove_DryRun_WithCheatSheet(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	customCheatSheetPath := filepath.Join(test.GetTestDesktopDir(t), "dry-run-cheat-sheet.md")

	if err := os.WriteFile(customCheatSheetPath, []byte("content"), 0600); err != nil {
		t.Fatalf("Failed to create cheat sheet: %v", err)
	}

	remover := NewRemover(config)
	result, err := remover.Remove(context.Background(), RemoveOptions{
		RemoveCheatSheet: true,
		CheatSheetPath:   customCheatSheetPath,
		DryRun:           true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if !result.RemovedCheatSheet {
		t.Error("Dry-run should report cheat sheet would be removed")
	}

	if !fileExists(customCheatSheetPath) {
		t.Error("Dry-run should not delete the cheat sheet")
	}
}
