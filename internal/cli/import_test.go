package cli

import (
	"testing"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/test"
)

func TestParseImportFlags_Defaults(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := &cobra.Command{}
	cmd.Flags().String("backup", "/tmp/backup.json", "Backup file")
	cmd.Flags().Bool("include-managed", false, "Include managed profiles")
	cmd.Flags().Bool("include-above", false, "Include profiles above")
	cmd.Flags().Bool("include-below", false, "Include profiles below")
	cmd.Flags().Bool("dry-run", false, "Dry run")
	cmd.Flags().Bool("ignore-settings", false, "Ignore settings")
	cmd.Flags().Bool("backup-current-settings", false, "Backup current settings")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	flags, err := parseImportFlags(cmd, []string{})
	if err != nil {
		t.Fatalf("parseImportFlags failed: %v", err)
	}

	// Default behavior: include everything if no flags specified
	if !flags.IncludeManaged || !flags.IncludeAbove || !flags.IncludeBelow {
		t.Error("All sections should be included by default")
	}
}
