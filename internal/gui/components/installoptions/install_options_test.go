package installoptions

import (
	"path/filepath"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/test"
)

func TestNewInstallOptions(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	if io == nil {
		t.Fatal("InstallOptions should not be nil")
	}

	if io.outputFileEntry == nil {
		t.Fatal("Output file entry should not be nil")
	}

	if io.cheatsheetFileEntry == nil {
		t.Fatal("Cheat sheet file entry should not be nil")
	}

	if io.eolFormatSelect == nil {
		t.Fatal("EOL format select should not be nil")
	}

	if io.generateCheatCheck == nil {
		t.Fatal("Generate cheat checkbox should not be nil")
	}

	if io.cheatsheetOnlyCheck == nil {
		t.Fatal("Cheatsheet only checkbox should not be nil")
	}
}

func TestInstallOptions_GetContent(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)
	content := io.GetContent()

	if content == nil {
		t.Fatal("GetContent should not return nil")
	}
}

func TestInstallOptions_DefaultValues(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	// Check output file default
	expectedOutput := filepath.Join(settings.GetAwsDir(), "config")
	if io.GetOutputFile() != expectedOutput {
		t.Errorf("Expected output file '%s', got '%s'", expectedOutput, io.GetOutputFile())
	}

	// Check cheat sheet file default
	expectedCheatsheet := filepath.Join(settings.GetDesktopDir(), "aws-profile-cheatsheet.md")
	if io.GetCheatsheetFile() != expectedCheatsheet {
		t.Errorf("Expected cheat sheet file '%s', got '%s'", expectedCheatsheet, io.GetCheatsheetFile())
	}

	// Check EOL format default (Native — OS decides at write time)
	if io.GetEolFormat() != EolNative {
		t.Errorf("Expected EOL format '%s', got '%s'", EolNative, io.GetEolFormat())
	}

	// Check checkboxes default
	// Generate cheatsheet should be checked by default
	if !io.ShouldGenerateCheatsheet() {
		t.Error("Generate cheatsheet should be checked by default")
	}

	// Cheatsheet only should be unchecked by default
	if io.IsCheatsheetOnly() {
		t.Error("Cheatsheet only should be unchecked by default")
	}
}

func TestInstallOptions_GetSetOutputFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	testPath := "/test/path/to/config"
	io.SetOutputFile(testPath)

	if io.GetOutputFile() != testPath {
		t.Errorf("Expected output file '%s', got '%s'", testPath, io.GetOutputFile())
	}
}

func TestInstallOptions_GetSetCheatsheetFile(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	testPath := "/test/path/to/cheatsheet.md"
	io.SetCheatsheetFile(testPath)

	if io.GetCheatsheetFile() != testPath {
		t.Errorf("Expected cheat sheet file '%s', got '%s'", testPath, io.GetCheatsheetFile())
	}
}

func TestInstallOptions_GetSetEolFormat(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	tests := []struct {
		name   string
		format EolFormat
	}{
		{"Native", EolNative},
		{"LF", EolLF},
		{"CRLF", EolCRLF},
	}

	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			io.SetEolFormat(tt.format)

			if io.GetEolFormat() != tt.format {
				t.Errorf("Expected EOL format '%s', got '%s'", tt.format, io.GetEolFormat())
			}
		})
	}
}

func TestInstallOptions_GenerateCheatsheet(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	// Should be checked by default
	if !io.ShouldGenerateCheatsheet() {
		t.Error("Should be checked by default")
	}

	// Set to unchecked
	io.SetGenerateCheatsheet(false)
	if io.ShouldGenerateCheatsheet() {
		t.Error("Should be unchecked after SetGenerateCheatsheet(false)")
	}

	// Set back to checked
	io.SetGenerateCheatsheet(true)
	if !io.ShouldGenerateCheatsheet() {
		t.Error("Should be checked after SetGenerateCheatsheet(true)")
	}
}

func TestInstallOptions_CheatsheetOnly(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	// Initially unchecked
	if io.IsCheatsheetOnly() {
		t.Error("Should be unchecked initially")
	}

	// Set to checked
	io.SetCheatsheetOnly(true)
	if !io.IsCheatsheetOnly() {
		t.Error("Should be checked after SetCheatsheetOnly(true)")
	}

	// Set to unchecked
	io.SetCheatsheetOnly(false)
	if io.IsCheatsheetOnly() {
		t.Error("Should be unchecked after SetCheatsheetOnly(false)")
	}
}

func TestInstallOptions_Reset(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	io := NewInstallOptions(nil)

	// Change all values
	io.SetOutputFile("/custom/path/config")
	io.SetCheatsheetFile("/custom/path/cheatsheet.md")
	io.SetEolFormat(EolCRLF)
	io.SetGenerateCheatsheet(true)
	io.SetCheatsheetOnly(true)

	// Verify changes
	if io.GetOutputFile() == filepath.Join(settings.GetAwsDir(), "config") {
		t.Error("Output file should have been changed")
	}

	// Reset
	io.Reset()

	// Verify defaults restored
	expectedOutput := filepath.Join(settings.GetAwsDir(), "config")
	if io.GetOutputFile() != expectedOutput {
		t.Errorf("Expected output file '%s' after reset, got '%s'", expectedOutput, io.GetOutputFile())
	}

	expectedCheatsheet := filepath.Join(settings.GetDesktopDir(), "aws-profile-cheatsheet.md")
	if io.GetCheatsheetFile() != expectedCheatsheet {
		t.Errorf("Expected cheat sheet file '%s' after reset, got '%s'", expectedCheatsheet, io.GetCheatsheetFile())
	}

	if io.GetEolFormat() != EolNative {
		t.Errorf("Expected EOL format '%s' after reset, got '%s'", EolNative, io.GetEolFormat())
	}

	if io.ShouldGenerateCheatsheet() {
		t.Error("Generate cheatsheet should be unchecked after reset")
	}

	if io.IsCheatsheetOnly() {
		t.Error("Cheatsheet only should be unchecked after reset")
	}
}

func TestDefaultEolIsNative(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	io := NewInstallOptions(nil)
	if io.GetEolFormat() != EolNative {
		t.Errorf("default EOL format should be EolNative, got %q", io.GetEolFormat())
	}
}

func TestEolFormatConstants(t *testing.T) {
	// Verify constant values match expected strings
	if EolNative != "Native" {
		t.Errorf("EolNative should be 'Native', got '%s'", EolNative)
	}

	if EolLF != "LF" {
		t.Errorf("EolLF should be 'LF', got '%s'", EolLF)
	}

	if EolCRLF != "CRLF" {
		t.Errorf("EolCRLF should be 'CRLF', got '%s'", EolCRLF)
	}
}
