package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRegisterCommands_InstallCommand(t *testing.T) {
	// Create a root command
	rootCmd := &cobra.Command{
		Use:   "aws-profile-manager",
		Short: "Test root command",
	}

	// Register CLI commands
	RegisterCommands(rootCmd)

	// Verify install command was added
	installCmd := findSubcommand(rootCmd, "install")
	if installCmd == nil {
		t.Error("install command was not registered")
		return
	}

	// Verify install command properties
	if installCmd.Use != "install" {
		t.Errorf("expected install command use 'install', got '%s'", installCmd.Use)
	}

	if installCmd.Short == "" {
		t.Error("install command should have a short description")
	}

	if installCmd.Long == "" {
		t.Error("install command should have a long description")
	}

	// Verify install command has the expected flags
	expectedFlags := []string{
		"organizations",
		"partitions",
		"accounts",
		"roles",
		"regions",
		"all-regions",
	}

	for _, flagName := range expectedFlags {
		flag := installCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("install command missing expected flag: %s", flagName)
		}
	}
}

func TestRegisterCommandsMultipleCalls(t *testing.T) {
	// Create a root command
	rootCmd := &cobra.Command{
		Use:   "aws-profile-manager",
		Short: "Test root command",
	}

	// Register commands multiple times
	RegisterCommands(rootCmd)
	RegisterCommands(rootCmd)

	// Should have duplicate subcommands (not ideal but acceptable)
	commands := rootCmd.Commands()
	if len(commands) < 6 {
		t.Errorf("Expected at least 6 commands after registration, got %d", len(commands))
	}
}

func TestGetAvailableCommands(t *testing.T) {
	rootCmd := &cobra.Command{
		Use:   "aws-profile-manager",
		Short: "Test root command",
	}

	RegisterCommands(rootCmd)

	commands := rootCmd.Commands()
	commandNames := make([]string, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name()
	}

	// Expected commands
	expectedCommands := []string{"install", "sessions", "profiles", "sync", "export", "import", "gui", "version"}

	for _, expected := range expectedCommands {
		found := false
		for _, actual := range commandNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command %q not found in: %v", expected, commandNames)
		}
	}
}

// Helper function to find a subcommand by name
func findSubcommand(rootCmd *cobra.Command, name string) *cobra.Command {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
