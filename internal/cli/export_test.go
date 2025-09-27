package cli

import (
	"testing"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/test"
)

func TestParseExportFlags_Defaults(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "", "Output path")
	cmd.Flags().Bool("include-managed", false, "Include managed profiles")
	cmd.Flags().Bool("include-above", false, "Include profiles above")
	cmd.Flags().Bool("include-below", false, "Include profiles below")
	cmd.Flags().String("description", "", "Export description")
	cmd.Flags().Bool("exclude-settings", false, "Exclude settings")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	flags, err := parseExportFlags(cmd)
	if err != nil {
		t.Fatalf("parseExportFlags failed: %v", err)
	}

	// Default behavior: include everything if no flags specified
	if !flags.IncludeManaged || !flags.IncludeAbove || !flags.IncludeBelow {
		t.Error("All sections should be included by default")
	}
}
