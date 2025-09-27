package cli

import (
	"testing"

	"aws-profile-manager/internal/test"
)

func TestRunGUI_InTestEnvironment(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create command
	cmd := createGUICommand()

	// Run GUI command - should detect test environment and not launch actual GUI
	runGUI(cmd, []string{})

	// If we get here without hanging, the test passed
	// The runGUI function detects test environment and returns early
}

func TestRunGUI_WithConfigFlag(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create command
	cmd := createGUICommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", "/test/config.json")

	// Run GUI command
	runGUI(cmd, []string{})

	// If we get here without hanging, the test passed
}

func TestIsTestEnvironment(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Should detect we're in a test environment
	if !isTestEnvironment() {
		t.Error("isTestEnvironment should return true when running tests")
	}
}
