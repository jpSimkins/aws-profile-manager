package cli

import (
	"runtime"
	"testing"

	"github.com/spf13/cobra"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/test"
)

func TestRunVersion(t *testing.T) {
	test.SetupTestEnvironment(t)

	cmd := &cobra.Command{}

	// Version command should not panic
	runVersion(cmd, []string{})

	// Verify version info is accessible
	versionInfo := core.GetVersion()

	// Verify version is set (should match AppVersion or be overridden)
	if versionInfo.Version == "" {
		t.Error("Version should not be empty")
	}

	if versionInfo.Version != core.AppVersion && versionInfo.Version != core.Version {
		t.Errorf("Version mismatch: got %s, expected %s or %s",
			versionInfo.Version, core.AppVersion, core.Version)
	}

	// Verify GoVersion is set
	if versionInfo.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if versionInfo.GoVersion != runtime.Version() {
		t.Errorf("GoVersion mismatch: got %s, expected %s",
			versionInfo.GoVersion, runtime.Version())
	}

	// Verify Platform is set
	if versionInfo.Platform == "" {
		t.Error("Platform should not be empty")
	}

	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if versionInfo.Platform != expectedPlatform {
		t.Errorf("Platform mismatch: got %s, expected %s",
			versionInfo.Platform, expectedPlatform)
	}
}

func TestGetVersionString(t *testing.T) {
	test.SetupTestEnvironment(t)

	versionString := core.GetVersionString()

	// Version string should not be empty
	if versionString == "" {
		t.Error("Version string should not be empty")
	}

	// Should contain the version number
	versionInfo := core.GetVersion()
	if versionInfo.Version == "" {
		t.Error("Version info should have a version")
	}
}
