package profiles

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/task"
)

// CheatSheet generates markdown reference guide for profiles.
//
// This component creates user-friendly markdown documentation with SSO login
// commands and profile listings organized by organization and partition.
//
// Usage:
//
//	config := buildConfigFromSettings()  // In CLI/GUI
//	cheatsheet := profiles.NewCheatSheet(config)
//	result, err := cheatsheet.Generate(ctx, opts, reporter)
type CheatSheet struct {
	config Config
}

// NewCheatSheet creates a new CheatSheet with injected configuration.
//
// Parameters:
//   - config: Configuration injected by CLI/GUI (from settings)
//
// Returns:
//   - *CheatSheet: Ready to use cheat sheet generator
//
// Example:
//
//	config := buildConfigFromSettings()
//	cheatsheet := profiles.NewCheatSheet(config)
func NewCheatSheet(config Config) *CheatSheet {
	return &CheatSheet{
		config: config,
	}
}

// Generate generates markdown cheat sheet.
//
// Creates markdown file with SSO login commands and profile listings.
//
// Parameters:
//   - ctx: Context for cancellation
//   - opts: Cheat sheet options
//   - reporter: Progress reporter
//
// Returns:
//   - *CheatSheetResult: Information about generated file
//   - error: Any error during generation
//
// Example:
//
//	result, err := cheatsheet.Generate(ctx, profiles.CheatSheetOptions{
//	    Collection: schema.Managed,
//	}, reporter)
func (c *CheatSheet) Generate(
	ctx context.Context,
	opts CheatSheetOptions,
	reporter task.Reporter,
) (*CheatSheetResult, error) {
	startTime := time.Now()
	logging.Debug.Log("Generate cheat sheet started")

	// Validate inputs
	if opts.Collection == nil || len(opts.Collection.Organizations) == 0 {
		return nil, fmt.Errorf("collection must have at least one organization")
	}

	// Determine output path
	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(c.config.CheatSheetOutputDir, "AWS_Profile_Cheat_Sheet.md")
	}

	// Check for cancellation
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Generate cancelled before start")
		return nil, err
	}

	result := &CheatSheetResult{
		OutputPath: outputPath,
		Duration:   time.Since(startTime),
	}

	// Step 1: Ensure output directory exists
	reporter.ReportStatus("Creating output directory...")
	logging.Debug.Log("Creating output directory")

	if err := ensureDirectory(filepath.Dir(outputPath)); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Step 2: Generate markdown content
	orgCount := len(opts.Collection.Organizations)
	reporter.ReportStatus(fmt.Sprintf("Generating cheat sheet content for %d organizations...", orgCount))
	logging.Debug.Log("Generating markdown content", "organizations", orgCount)

	content := c.generateMarkdownContent(opts.Collection, result)

	logging.Debug.Log("Content generated",
		"sessions", result.Sessions,
		"profiles", result.Profiles,
		"size", len(content),
	)

	// Check for cancellation before write
	if err := ctx.Err(); err != nil {
		logging.Debug.Log("Generate cancelled before write")
		return nil, err
	}

	// Step 3: Write to file
	reporter.ReportStatus(fmt.Sprintf("Writing cheat sheet to %s...", filepath.Base(outputPath)))
	logging.Debug.Log("Writing cheat sheet file")

	if err := writeFileContent(outputPath, content); err != nil {
		return nil, fmt.Errorf("failed to write cheat sheet file: %w", err)
	}

	// Get file size
	if size, err := getFileSize(outputPath); err == nil {
		result.FileSize = size
	}

	reporter.ReportStatus("Cheat sheet generated successfully")
	logging.Debug.Log("Generate completed",
		"path", outputPath,
		"size", result.FileSize,
		"duration", result.Duration,
	)

	return result, nil
}

// generateMarkdownContent generates the complete markdown content.
func (c *CheatSheet) generateMarkdownContent(collection *schema.ProfileCollection, result *CheatSheetResult) string {
	var content strings.Builder

	// Header
	content.WriteString("# AWS CLI Profile Cheat Sheet\n\n")
	content.WriteString("This document provides a quick reference for all AWS CLI profiles managed by the AWS Profile Manager.\n\n")
	content.WriteString("## Usage Instructions\n\n")
	content.WriteString("1. **SSO Login**: Run the login command for your organization before using profiles\n")
	content.WriteString("2. **Use Profile**: Export `AWS_PROFILE` or use `--profile` flag with AWS CLI commands\n")
	content.WriteString("3. **Verify**: Run `aws sts get-caller-identity` to verify profile works\n\n")

	// Generate content for organizations
	content.WriteString(c.generateOrganizationsContent(collection, result))

	// Footer — version and timestamp are each controlled by config flags.
	content.WriteString("\n---\n\n")
	if c.config.IncludeVersion {
		// Show the app version so readers know which tool generated the sheet.
		content.WriteString(fmt.Sprintf("*Generated by AWS Profile Manager v%s*\n", core.AppVersion))
	}
	if c.config.IncludeTimestamp {
		// Use the friendly local-timezone format so the reader immediately knows
		// when the sheet was generated without having to parse an ISO timestamp.
		content.WriteString(fmt.Sprintf("*Generated at %s*\n", core.FormatFriendlyDateTimeWithZone(time.Now())))
	}

	return content.String()
}

// generateOrganizationsContent generates markdown for all organizations.
func (c *CheatSheet) generateOrganizationsContent(collection *schema.ProfileCollection, result *CheatSheetResult) string {
	var content strings.Builder

	// Generate content for each organization
	for orgAlias, org := range collection.Organizations {
		// Generate content for each partition
		for partitionName, partition := range org.Partitions {
			content.WriteString(c.generatePartitionSection(orgAlias, org.Name, partitionName, partition, result))
		}
	}

	return content.String()
}

// generatePartitionSection generates markdown for one partition.
func (c *CheatSheet) generatePartitionSection(
	orgAlias string,
	orgName string,
	partitionName string,
	partition schema.Partition,
	result *CheatSheetResult,
) string {
	var content strings.Builder

	// Capitalize first letter of partition name
	partitionTitle := strings.ToUpper(partitionName[:1]) + partitionName[1:]
	content.WriteString(fmt.Sprintf("## %s - %s\n\n", orgName, partitionTitle))

	// SSO Login Command
	sessionName := schema.GenerateSsoSessionName(orgAlias, partitionName)
	content.WriteString("**SSO Login Command:**\n\n")
	content.WriteString("```bash\n")
	content.WriteString(fmt.Sprintf("aws sso login --sso-session %s\n", sessionName))
	content.WriteString("```\n\n")
	content.WriteString(fmt.Sprintf("**SSO URL:** %s\n\n", partition.URL))
	content.WriteString(fmt.Sprintf("**Default Region:** %s\n\n", partition.DefaultRegion))
	result.Sessions++

	// Profile Table
	content.WriteString("**Profiles:**\n\n")
	content.WriteString("| Account | Role | Profile Name | Region |\n")
	content.WriteString("|---------|------|--------------|--------|\n")

	// List profiles for each account and role
	for _, account := range partition.Accounts {
		for _, role := range partition.Roles {
			profileName := schema.GenerateProfileName(partitionName, account.Alias, role, "", partition.DefaultRegion)
			regionDisplay := partition.DefaultRegion

			content.WriteString(fmt.Sprintf("| %s | %s | `%s` | %s |\n",
				account.Alias,
				role,
				profileName,
				regionDisplay,
			))
			result.Profiles++
		}
	}

	content.WriteString("\n")

	return content.String()
}
