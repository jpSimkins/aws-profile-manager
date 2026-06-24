// Package test provides test fixtures for profiles package.package test

// This package is for use by OTHER packages that need profiles test data
// (e.g., backup tests, CLI tests, installer tests). It CANNOT be used by the
// profiles package itself due to import cycles.
//
// The profiles package has its own internal test fixtures in *_test.go files.
//
// This package provides pre-built AWS config file content that can be written
// to test environments for testing profile extraction, import, export, etc.
//
// Usage:
//
//	import profilestest "aws-profile-manager/internal/profiles/test"
//
//	func TestMyFeature(t *testing.T) {
//	    test.SetupTestEnvironment(t)
//	    profilestest.WriteConfig(t, profilestest.NewConfigWithSso())
//	    // ... test with AWS config file
//	}
package test

import (
	"aws-profile-manager/internal/task"
	"context"
	"os"
	"testing"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/test"
)

// NewConfigWithSsoSingle returns AWS config content with single SSO profile.
//
// Uses schema/test's NewManagedSsoSingle for consistency.
// Includes managed section markers and SSO session.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithSsoSingle() string {
	schema := schematest.NewManagedSsoSingle()
	return generateConfigContent(schema)
}

// NewConfigWithSsoMultiAccount returns AWS config content with multiple SSO accounts.
//
// Uses schema/test's NewManagedSsoMultiAccount for consistency.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithSsoMultiAccount() string {
	schema := schematest.NewManagedSsoMultiAccount()
	return generateConfigContent(schema)
}

// NewConfigWithIamSingle returns AWS config content with single IAM profile.
//
// Uses schema/test's NewManagedIamSingle for consistency.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithIamSingle() string {
	schema := schematest.NewManagedIamSingle()
	return generateConfigContent(schema)
}

// NewConfigWithIamMulti returns AWS config content with multiple IAM profiles.
//
// Uses schema/test's NewManagedIamMulti for consistency.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithIamMulti() string {
	schema := schematest.NewManagedIamMulti()
	return generateConfigContent(schema)
}

// NewConfigWithAssumeRoleSingle returns AWS config content with single AssumeRole chain.
//
// Uses schema/test's NewManagedAssumeRoleSingle for consistency.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithAssumeRoleSingle() string {
	schema := schematest.NewManagedAssumeRoleSingle()
	return generateConfigContent(schema)
}

// NewConfigWithAllTypes returns AWS config content with all profile types.
//
// Uses schema/test's NewManagedAll for consistency.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigWithAllTypes() string {
	schema := schematest.NewManagedAll()
	return generateConfigContent(schema)
}

// NewConfigMixed returns AWS config content with managed and unmanaged sections.
//
// Uses schema/test's NewMixedSimple for consistency.
// Includes personal profiles above and below managed section.
//
// Returns:
//   - string: Complete AWS config file content
func NewConfigMixed() string {
	schema := schematest.NewMixedSimple()
	return generateConfigContentWithUnmanaged(schema)
}

// NewConfigEmpty returns empty AWS config content.
//
// Returns:
//   - string: Empty string
func NewConfigEmpty() string {
	return ""
}

// NewConfigLarge returns AWS config content with many profiles.
//
// Uses schema/test's NewLargeScale for performance testing.
//
// Returns:
//   - string: Complete AWS config file content with 2100+ profiles
func NewConfigLarge() string {
	schema := schematest.NewLargeScale()
	return generateConfigContent(schema)
}

// WriteConfig writes AWS config content to the test environment's AWS config file.
//
// This is a convenience function that writes to ~/.aws/config in the test environment
// created by test.SetupTestEnvironment(t).
//
// Parameters:
//   - t: Testing context
//   - content: AWS config file content to write
//
// Example:
//
//	test.SetupTestEnvironment(t)
//	profilestest.WriteConfig(t, profilestest.NewConfigWithSsoSingle())
func WriteConfig(t *testing.T, content string) {
	t.Helper()
	configPath := test.GetTestAwsConfigPath(t)
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
}

// generateConfigContent generates AWS config file content from schema.
//
// Generates managed section only with markers.
//
// Parameters:
//   - s: Schema to generate from
//
// Returns:
//   - string: Complete AWS config file content
func generateConfigContent(s *schema.Schema) string {
	if s.Managed == nil {
		return ""
	}

	var content string

	// Add start marker
	content += "# START - Managed by AWS Profile Manager\n\n"

	// Generate profiles (SSO includes sessions automatically)
	if len(s.Managed.Organizations) > 0 {
		profileContent, _, _ := generators.GenerateSsoProfiles(context.Background(), s.Managed, task.NoOpReporter{})
		content += profileContent
	}
	if len(s.Managed.IamUsers) > 0 {
		profileContent, _, _ := generators.GenerateIamProfiles(context.Background(), s.Managed, task.NoOpReporter{})
		content += profileContent
	}
	if len(s.Managed.AssumeRoleChains) > 0 {
		profileContent, _, _ := generators.GenerateAssumeRoleProfiles(context.Background(), s.Managed, task.NoOpReporter{})
		content += profileContent
	}
	if len(s.Managed.GenericProfiles) > 0 {
		profileContent, _, _ := generators.GenerateGenericProfiles(context.Background(), s.Managed, task.NoOpReporter{})
		content += profileContent
	}

	// Add end marker
	content += "\n# END - Managed by AWS Profile Manager\n"

	return content
}

// generateConfigContentWithUnmanaged generates AWS config file content with unmanaged sections.
//
// Generates content above managed section, managed section, and content below.
//
// Parameters:
//   - s: Schema to generate from
//
// Returns:
//   - string: Complete AWS config file content
func generateConfigContentWithUnmanaged(s *schema.Schema) string {
	var content string

	// Generate above section
	if s.Unmanaged != nil && s.Unmanaged.Above != nil {
		if len(s.Unmanaged.Above.Organizations) > 0 {
			profileContent, _, _ := generators.GenerateSsoProfiles(context.Background(), s.Unmanaged.Above, task.NoOpReporter{})
			content += profileContent + "\n"
		}
		if len(s.Unmanaged.Above.IamUsers) > 0 {
			profileContent, _, _ := generators.GenerateIamProfiles(context.Background(), s.Unmanaged.Above, task.NoOpReporter{})
			content += profileContent + "\n"
		}
		if len(s.Unmanaged.Above.GenericProfiles) > 0 {
			profileContent, _, _ := generators.GenerateGenericProfiles(context.Background(), s.Unmanaged.Above, task.NoOpReporter{})
			content += profileContent + "\n"
		}
	}

	// Generate managed section
	if s.Managed != nil {
		content += "# START - Managed by AWS Profile Manager\n\n"

		if len(s.Managed.Organizations) > 0 {
			profileContent, _, _ := generators.GenerateSsoProfiles(context.Background(), s.Managed, task.NoOpReporter{})
			content += profileContent
		}
		if len(s.Managed.IamUsers) > 0 {
			profileContent, _, _ := generators.GenerateIamProfiles(context.Background(), s.Managed, task.NoOpReporter{})
			content += profileContent
		}
		if len(s.Managed.AssumeRoleChains) > 0 {
			profileContent, _, _ := generators.GenerateAssumeRoleProfiles(context.Background(), s.Managed, task.NoOpReporter{})
			content += profileContent
		}
		if len(s.Managed.GenericProfiles) > 0 {
			profileContent, _, _ := generators.GenerateGenericProfiles(context.Background(), s.Managed, task.NoOpReporter{})
			content += profileContent
		}

		content += "\n# END - Managed by AWS Profile Manager\n"
	}

	// Generate below section
	if s.Unmanaged != nil && s.Unmanaged.Below != nil {
		content += "\n"
		if len(s.Unmanaged.Below.Organizations) > 0 {
			profileContent, _, _ := generators.GenerateSsoProfiles(context.Background(), s.Unmanaged.Below, task.NoOpReporter{})
			content += profileContent
		}
		if len(s.Unmanaged.Below.IamUsers) > 0 {
			profileContent, _, _ := generators.GenerateIamProfiles(context.Background(), s.Unmanaged.Below, task.NoOpReporter{})
			content += profileContent
		}
		if len(s.Unmanaged.Below.GenericProfiles) > 0 {
			profileContent, _, _ := generators.GenerateGenericProfiles(context.Background(), s.Unmanaged.Below, task.NoOpReporter{})
			content += profileContent
		}
	}

	return content
}
