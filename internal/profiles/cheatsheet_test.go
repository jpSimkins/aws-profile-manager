package profiles

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aws-profile-manager/internal/core"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

// TestNewCheatSheet tests CheatSheet constructor
func TestNewCheatSheet(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)

	if writer == nil {
		t.Fatal("NewCheatSheet should not return nil")
	}
}

// TestCheatSheet_Generate_Success tests successful cheat sheet generation
func TestCheatSheet_Generate_Success(t *testing.T) {
	test.SetupTestEnvironment(t)

	outputPath := filepath.Join(test.GetTestConfigDir(t), "cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	// Create cheat sheet
	writer := NewCheatSheet(config)
	result, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: schematest.NewManagedSsoSingle().Managed,
		OutputPath: outputPath,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.Profiles == 0 {
		t.Error("Should have generated profiles")
	}

	// Verify file was created
	if !fileExists(outputPath) {
		t.Fatal("Cheat sheet file should have been created")
	}

	// Verify content is markdown
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read cheat sheet: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# AWS") || !strings.Contains(contentStr, "Cheat Sheet") {
		t.Errorf("Cheat sheet should contain title, got: %s", contentStr[:min(len(contentStr), 200)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestCheatSheet_Generate_EmptyCollection tests empty collection
func TestCheatSheet_Generate_EmptyCollection(t *testing.T) {
	test.SetupTestEnvironment(t)

	outputPath := filepath.Join(test.GetTestConfigDir(t), "empty-cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)
	_, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: schematest.NewEmpty().Managed,
		OutputPath: outputPath,
	}, task.NoOpReporter{})

	// Should fail with empty collection (needs at least one org)
	if err == nil {
		t.Error("Generate should fail with empty collection")
	}
}

// TestCheatSheet_Generate_InvalidOutputPath tests invalid output path
func TestCheatSheet_Generate_InvalidOutputPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)
	_, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: schematest.NewManagedSsoSingle().Managed,
		OutputPath: "/invalid/path/that/cannot/exist/cheatsheet.md",
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Generate should fail with invalid output path")
	}
}

// TestCheatSheet_Generate_AllTypes tests all profile types
func TestCheatSheet_Generate_AllTypes(t *testing.T) {
	test.SetupTestEnvironment(t)

	outputPath := filepath.Join(test.GetTestConfigDir(t), "all-types-cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)
	result, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: schematest.NewManagedAll().Managed,
		OutputPath: outputPath,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.Profiles == 0 {
		t.Error("Should have generated profiles")
	}

	// Verify file contains information
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read cheat sheet: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Cheat sheet should not be empty")
	}
}

// TestCheatSheet_Generate_NilCollection tests nil collection handling
func TestCheatSheet_Generate_NilCollection(t *testing.T) {
	test.SetupTestEnvironment(t)

	outputPath := filepath.Join(test.GetTestConfigDir(t), "nil-cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)
	_, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: nil,
		OutputPath: outputPath,
	}, task.NoOpReporter{})

	// Should fail with nil collection
	if err == nil {
		t.Error("Generate should fail with nil collection")
	}
}

// TestCheatSheet_Generate_ContextCancellation tests context cancellation
func TestCheatSheet_Generate_ContextCancellation(t *testing.T) {
	test.SetupTestEnvironment(t)

	customPath := filepath.Join(test.GetTestDesktopDir(t), "custom-cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	// Cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	writer := NewCheatSheet(config)
	_, err := writer.Generate(ctx, CheatSheetOptions{
		Collection: schematest.NewManagedSsoSingle().Managed,
		OutputPath: customPath,
	}, task.NoOpReporter{})

	if err == nil {
		t.Error("Should return error when context is cancelled")
	}
}

// TestCheatSheet_Generate_CustomOutputPath tests custom output path
func TestCheatSheet_Generate_CustomOutputPath(t *testing.T) {
	test.SetupTestEnvironment(t)

	customPath := filepath.Join(test.GetTestDesktopDir(t), "custom-cheatsheet.md")
	config := Config{
		ConfigPath:          test.GetTestAwsConfigPath(t),
		CheatSheetOutputDir: test.GetTestDesktopDir(t),
		StartMarker:         "# START",
		EndMarker:           "# END",
	}

	writer := NewCheatSheet(config)
	result, err := writer.Generate(context.Background(), CheatSheetOptions{
		Collection: schematest.NewManagedSsoSingle().Managed,
		OutputPath: customPath,
	}, task.NoOpReporter{})

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.OutputPath != customPath {
		t.Errorf("Output path = %s, want %s", result.OutputPath, customPath)
	}

	// Verify file was created at custom path
	if !fileExists(customPath) {
		t.Error("File should exist at custom path")
	}
}

// TestCheatSheet_Footer_IncludeVersion checks that the app version appears in
// the cheat sheet footer when IncludeVersion is true and is absent when false.
func TestCheatSheet_Footer_IncludeVersion(t *testing.T) {
	test.SetupTestEnvironment(t)

	wantVersion := fmt.Sprintf("*Generated by AWS Profile Manager v%s*", core.AppVersion)

	t.Run("version present when IncludeVersion=true", func(t *testing.T) {
		outputPath := filepath.Join(test.GetTestConfigDir(t), "version-on.md")
		cfg := Config{
			StartMarker:    "# START",
			EndMarker:      "# END",
			IncludeVersion: true,
		}
		_, err := NewCheatSheet(cfg).Generate(context.Background(), CheatSheetOptions{
			Collection: schematest.NewManagedSsoSingle().Managed,
			OutputPath: outputPath,
		}, task.NoOpReporter{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		content, _ := os.ReadFile(outputPath)
		if !strings.Contains(string(content), wantVersion) {
			t.Errorf("expected %q in footer\n---\n%s", wantVersion, string(content))
		}
	})

	t.Run("version absent when IncludeVersion=false", func(t *testing.T) {
		outputPath := filepath.Join(test.GetTestConfigDir(t), "version-off.md")
		cfg := Config{
			StartMarker:    "# START",
			EndMarker:      "# END",
			IncludeVersion: false,
		}
		_, err := NewCheatSheet(cfg).Generate(context.Background(), CheatSheetOptions{
			Collection: schematest.NewManagedSsoSingle().Managed,
			OutputPath: outputPath,
		}, task.NoOpReporter{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		content, _ := os.ReadFile(outputPath)
		// Just check the version string prefix, not the full line, in case appversion changes.
		if strings.Contains(string(content), "*Generated by AWS Profile Manager v") {
			t.Errorf("did not expect version line in footer when disabled\n---\n%s", string(content))
		}
	})
}

// TestCheatSheet_Footer_IncludeTimestamp checks that a timestamp line appears
// in the cheat sheet footer when IncludeTimestamp is true and is absent when false.
func TestCheatSheet_Footer_IncludeTimestamp(t *testing.T) {
	test.SetupTestEnvironment(t)

	const timestampPrefix = "*Generated at "

	t.Run("timestamp present when IncludeTimestamp=true", func(t *testing.T) {
		outputPath := filepath.Join(test.GetTestConfigDir(t), "ts-on.md")
		cfg := Config{
			StartMarker:      "# START",
			EndMarker:        "# END",
			IncludeTimestamp: true,
		}
		_, err := NewCheatSheet(cfg).Generate(context.Background(), CheatSheetOptions{
			Collection: schematest.NewManagedSsoSingle().Managed,
			OutputPath: outputPath,
		}, task.NoOpReporter{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		content, _ := os.ReadFile(outputPath)
		if !strings.Contains(string(content), timestampPrefix) {
			t.Errorf("expected timestamp line starting with %q in footer\n---\n%s", timestampPrefix, string(content))
		}
	})

	t.Run("timestamp absent when IncludeTimestamp=false", func(t *testing.T) {
		outputPath := filepath.Join(test.GetTestConfigDir(t), "ts-off.md")
		cfg := Config{
			StartMarker:      "# START",
			EndMarker:        "# END",
			IncludeTimestamp: false,
		}
		_, err := NewCheatSheet(cfg).Generate(context.Background(), CheatSheetOptions{
			Collection: schematest.NewManagedSsoSingle().Managed,
			OutputPath: outputPath,
		}, task.NoOpReporter{})
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		content, _ := os.ReadFile(outputPath)
		if strings.Contains(string(content), timestampPrefix) {
			t.Errorf("did not expect timestamp line in footer when disabled\n---\n%s", string(content))
		}
	})
}

// TestCheatSheet_Footer_BothMetadataFlags exercises all four combinations of
// IncludeVersion and IncludeTimestamp to prevent regression.
func TestCheatSheet_Footer_BothMetadataFlags(t *testing.T) {
	test.SetupTestEnvironment(t)

	versionPrefix := "*Generated by AWS Profile Manager v"
	timestampPrefix := "*Generated at "

	tests := []struct {
		name             string
		includeVersion   bool
		includeTimestamp bool
		wantVersion      bool
		wantTimestamp    bool
	}{
		{"both off", false, false, false, false},
		{"version only", true, false, true, false},
		{"timestamp only", false, true, false, true},
		{"both on", true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(test.GetTestConfigDir(t), tt.name+".md")
			cfg := Config{
				StartMarker:      "# START",
				EndMarker:        "# END",
				IncludeVersion:   tt.includeVersion,
				IncludeTimestamp: tt.includeTimestamp,
			}
			_, err := NewCheatSheet(cfg).Generate(context.Background(), CheatSheetOptions{
				Collection: schematest.NewManagedSsoSingle().Managed,
				OutputPath: outputPath,
			}, task.NoOpReporter{})
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			content := string(func() []byte { b, _ := os.ReadFile(outputPath); return b }())

			hasVersion := strings.Contains(content, versionPrefix)
			hasTimestamp := strings.Contains(content, timestampPrefix)

			if hasVersion != tt.wantVersion {
				t.Errorf("version present=%v, want %v\n---\n%s", hasVersion, tt.wantVersion, content)
			}
			if hasTimestamp != tt.wantTimestamp {
				t.Errorf("timestamp present=%v, want %v\n---\n%s", hasTimestamp, tt.wantTimestamp, content)
			}
		})
	}
}
