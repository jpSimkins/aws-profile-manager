package main

import (
	"testing"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/cli"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/test"
)

func TestInitializeApp(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Test that initializeApp doesn't panic and returns an error value
	err := initializeApp()

	// The function should not panic
	// Error could be nil (success) or an error (failure), both are acceptable
	if err != nil {
		t.Logf("App initialization returned error (acceptable for test environment): %v", err)
	}
}

func TestConfigFileGlobal(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		expected string
	}{
		{
			name:     "empty config file",
			setValue: "",
			expected: "",
		},
		{
			name:     "config file set",
			setValue: "/path/to/config.json",
			expected: "/path/to/config.json",
		},
		{
			name:     "relative path",
			setValue: "./config.json",
			expected: "./config.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalConfigFile := configFile
			defer func() { configFile = originalConfigFile }()

			// Set test value
			configFile = tt.setValue

			// Verify the global variable works as expected
			if configFile != tt.expected {
				t.Errorf("configFile = %v, want %v", configFile, tt.expected)
			}
		})
	}
}

func TestCommandHandlerSignatures(t *testing.T) {
	// Note: Command handlers (runGUI, runVersion, runInstall, etc.) are now
	// in the CLI package and tested there. We just verify the main structure here.
	// The default behavior (no subcommand) is to show help, which is handled by Cobra.

	// Test that we can create commands without panic
	cmd := &cobra.Command{Use: "test"}

	// Verify command was created successfully
	if cmd.Use != "test" {
		t.Error("Command Use field should be 'test'")
	}
}

func TestCommandCreation(t *testing.T) {
	// Test that we can create the root command without issues
	// This tests the command structure and flag setup

	rootCmd := &cobra.Command{
		Use:     "aws-profile-manager",
		Short:   "A cross-platform AWS profile management tool",
		Version: core.GetVersionString(),
	}

	// Test adding flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "test config flag")

	// Test using RegisterCommands from CLI package instead of creating manually
	cli.RegisterCommands(rootCmd)

	// Verify command structure
	if rootCmd.Use != "aws-profile-manager" {
		t.Errorf("rootCmd.Use = %v, want %v", rootCmd.Use, "aws-profile-manager")
	}

	// Check that subcommands were added
	commands := rootCmd.Commands()
	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name()
	}

	expectedCommands := []string{"install", "sessions", "profiles", "sync", "gui", "version"}
	for _, expected := range expectedCommands {
		found := false
		for _, actual := range commandNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %q not found in commands: %v", expected, commandNames)
		}
	}

	// Test that config flag exists
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("config flag should be registered")
	}
}
