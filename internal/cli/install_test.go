package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestParseInstallFlags_Defaults(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := createInstallCommand()

	flags, err := parseInstallFlags(cmd)
	if err != nil {
		t.Fatalf("parseInstallFlags failed: %v", err)
	}

	if len(flags.Organizations) != 0 {
		t.Error("Organizations should be empty by default")
	}

	if len(flags.Partitions) != 0 {
		t.Error("Partitions should be empty by default")
	}

	if flags.DryRun {
		t.Error("DryRun should be false by default")
	}

	if flags.Verbose {
		t.Error("Verbose should be false by default")
	}

	if flags.Remove {
		t.Error("Remove should be false by default")
	}

	if flags.CheatSheetOnly {
		t.Error("CheatSheetOnly should be false by default")
	}

	if flags.AllRegions {
		t.Error("AllRegions should be false by default")
	}
}

func TestParseInstallFlags_WithFilters(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := createInstallCommand()
	_ = cmd.Flags().Set("organizations", "org1,org2")
	_ = cmd.Flags().Set("partitions", "commercial")
	_ = cmd.Flags().Set("accounts", "dev,prod")
	_ = cmd.Flags().Set("roles", "Developer")
	_ = cmd.Flags().Set("regions", "us-east-1")

	flags, err := parseInstallFlags(cmd)
	if err != nil {
		t.Fatalf("parseInstallFlags failed: %v", err)
	}

	if len(flags.Organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(flags.Organizations))
	}

	if len(flags.Partitions) != 1 {
		t.Errorf("Expected 1 partition, got %d", len(flags.Partitions))
	}

	if len(flags.Accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(flags.Accounts))
	}

	if len(flags.Roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(flags.Roles))
	}

	if len(flags.Regions) != 1 {
		t.Errorf("Expected 1 region, got %d", len(flags.Regions))
	}
}

func TestParseInstallFlags_WithCheatSheet(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := createInstallCommand()
	_ = cmd.Flags().Set("cheat-sheet", "/path/to/cheatsheet.md")

	flags, err := parseInstallFlags(cmd)
	if err != nil {
		t.Fatalf("parseInstallFlags failed: %v", err)
	}

	if flags.CheatSheet != "/path/to/cheatsheet.md" {
		t.Errorf("Expected cheat sheet path /path/to/cheatsheet.md, got %s", flags.CheatSheet)
	}
}

func TestParseInstallFlags_WithAllOptions(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := createInstallCommand()
	_ = cmd.Flags().Set("organizations", "org1")
	_ = cmd.Flags().Set("dry-run", "true")
	_ = cmd.Flags().Set("verbose", "true")
	_ = cmd.Flags().Set("all-regions", "true")
	_ = cmd.Flags().Set("cheat-sheet-only", "true")

	flags, err := parseInstallFlags(cmd)
	if err != nil {
		t.Fatalf("parseInstallFlags failed: %v", err)
	}

	if !flags.DryRun {
		t.Error("DryRun should be true")
	}

	if !flags.Verbose {
		t.Error("Verbose should be true")
	}

	if !flags.AllRegions {
		t.Error("AllRegions should be true")
	}

	if !flags.CheatSheetOnly {
		t.Error("CheatSheetOnly should be true")
	}
}

func TestRunInstall_WithSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test schema file
	configDir := test.GetTestConfigDir(t)
	schemaFile := filepath.Join(configDir, "test-schema.json")

	testSchema := &schema.Schema{
		Version: "2.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test": {
					Name: "test",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{
									Alias: "dev",
									Name:  "Development",
									ID:    "123456789012",
								},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}

	schemaJSON, _ := json.Marshal(testSchema)
	if err := os.WriteFile(schemaFile, schemaJSON, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create command
	cmd := createInstallCommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", schemaFile)
	_ = cmd.Flags().Set("dry-run", "true")

	// Run install
	err := runInstall(cmd, []string{})
	if err != nil {
		t.Fatalf("runInstall failed: %v", err)
	}
}

func TestRunInstall_DryRun(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test schema
	configDir := test.GetTestConfigDir(t)
	schemaFile := filepath.Join(configDir, "test-schema.json")

	testSchema := &schema.Schema{
		Version: "2.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test": {
					Name: "test",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "dev", Name: "Dev", ID: "123456789012"},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}

	schemaJSON, _ := json.Marshal(testSchema)
	if err := os.WriteFile(schemaFile, schemaJSON, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create command with dry-run
	cmd := createInstallCommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", schemaFile)
	_ = cmd.Flags().Set("dry-run", "true")

	// Run install in dry-run mode
	err := runInstall(cmd, []string{})
	if err != nil {
		t.Fatalf("runInstall in dry-run mode failed: %v", err)
	}

	// Verify no AWS config was actually written
	awsConfigPath := filepath.Join(settings.GetAwsDir(), "config")
	if _, err := os.Stat(awsConfigPath); err == nil {
		// If file exists, check it wasn't modified by this test
		content, _ := os.ReadFile(awsConfigPath)
		if len(content) > 0 {
			t.Fatal("Config files should not be created in dry-run mode")
		}
	}
}

func TestRunInstall_WithFiltering(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test schema with multiple orgs/accounts
	configDir := test.GetTestConfigDir(t)
	schemaFile := filepath.Join(configDir, "test-schema.json")

	testSchema := &schema.Schema{
		Version: "2.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"org1": {
					Name: "org1",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://org1.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "dev", Name: "Dev", ID: "111111111111"},
								{Alias: "prod", Name: "Prod", ID: "222222222222"},
							},
							Roles: []string{"Developer", "PowerUser"},
						},
					},
				},
				"org2": {
					Name: "org2",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://org2.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "staging", Name: "Staging", ID: "333333333333"},
							},
							Roles: []string{"ReadOnly"},
						},
					},
				},
			},
		},
	}

	schemaJSON, _ := json.Marshal(testSchema)
	if err := os.WriteFile(schemaFile, schemaJSON, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create command with filters
	cmd := createInstallCommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", schemaFile)
	_ = cmd.Flags().Set("organizations", "org1")
	_ = cmd.Flags().Set("accounts", "dev")
	_ = cmd.Flags().Set("dry-run", "true")

	// Run install with filters
	err := runInstall(cmd, []string{})
	if err != nil {
		t.Fatalf("runInstall with filters failed: %v", err)
	}
}

func TestRunInstall_VerboseMode(t *testing.T) {
	test.SetupTestEnvironment(t)

	// Create test schema
	configDir := test.GetTestConfigDir(t)
	schemaFile := filepath.Join(configDir, "test-schema.json")

	testSchema := &schema.Schema{
		Version: "2.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test": {
					Name: "test",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "dev", Name: "Dev", ID: "123456789012"},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}

	schemaJSON, _ := json.Marshal(testSchema)
	if err := os.WriteFile(schemaFile, schemaJSON, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	// Create command with verbose
	cmd := createInstallCommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", schemaFile)
	_ = cmd.Flags().Set("verbose", "true")
	_ = cmd.Flags().Set("dry-run", "true")

	// Run install in verbose mode
	err := runInstall(cmd, []string{})
	if err != nil {
		t.Fatalf("runInstall in verbose mode failed: %v", err)
	}
}

func TestRunInstall_CheatSheetOnly_DefaultPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	configDir := test.GetTestConfigDir(t)
	schemaFile := filepath.Join(configDir, "test-schema.json")

	testSchema := &schema.Schema{
		Version: "2.0",
		Managed: &schema.ProfileCollection{
			Organizations: map[string]*schema.Organization{
				"test": {
					Name: "test",
					Partitions: map[string]schema.Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-west-2"},
							Accounts: []schema.Account{
								{Alias: "dev", Name: "Dev", ID: "123456789012"},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}

	schemaJSON, _ := json.Marshal(testSchema)
	if err := os.WriteFile(schemaFile, schemaJSON, 0644); err != nil {
		t.Fatalf("Failed to write schema file: %v", err)
	}

	cmd := createInstallCommand()
	cmd.Root().PersistentFlags().String("config", "", "Config file")
	_ = cmd.Root().PersistentFlags().Set("config", schemaFile)
	_ = cmd.Flags().Set("cheat-sheet-only", "true")

	err := runInstall(cmd, []string{})
	if err != nil {
		t.Fatalf("runInstall in cheat-sheet-only mode failed: %v", err)
	}

	awsConfigPath := filepath.Join(settings.GetAwsDir(), "config")
	if _, err := os.Stat(awsConfigPath); !os.IsNotExist(err) {
		t.Fatal("Config file should not be created in cheat-sheet-only mode")
	}

	defaultCheatSheetPath := filepath.Join(settings.GetDesktopDir(), "AWS_Profile_Cheat_Sheet.md")
	if _, err := os.Stat(defaultCheatSheetPath); err != nil {
		t.Fatalf("Expected cheat sheet at default path, got error: %v", err)
	}
}
