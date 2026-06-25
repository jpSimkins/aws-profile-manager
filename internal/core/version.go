package core

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"aws-profile-manager/internal/logging"
)

// Application version constants.
//
// These constants define the application metadata and should be updated
// for each release. The Version variable can be overridden at build time
// via ldflags.
const (
	AppName    = "AWS Profile Manager"                              // Application display name
	AppVersion = "0.3.0"                                            // Semantic version (update for releases)
	AppAuthor  = "jpSimkins"                                        // Application author
	AppURL     = "https://github.com/jpsimkins/aws-profile-manager" // Project URL
	Framework  = "Fyne"                                             // GUI framework name
)

// Build information variables.
//
// These variables can be set via ldflags during the build process to inject
// build metadata into the binary.
//
// Example build command:
//
//	go build -ldflags "-X aws-profile-manager/internal/core.Version=1.0.0 \
//	  -X aws-profile-manager/internal/core.Commit=$(git rev-parse HEAD) \
//	  -X aws-profile-manager/internal/core.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
var (
	Version   = AppVersion        // Version defaults to AppVersion, overridable at build time
	Commit    = ""                // Git commit hash (set via ldflags if available)
	Date      = ""                // Build date in ISO 8601 format (set via ldflags)
	GoVersion = runtime.Version() // Go version used to build the binary
)

// Info contains complete version and build information.
//
// This struct provides all version metadata in a structured format,
// useful for JSON serialization and programmatic access.
type Info struct {
	Version          string `json:"version"`           // Application version
	Commit           string `json:"commit"`            // Git commit hash
	Date             string `json:"date"`              // Build date
	GoVersion        string `json:"go_version"`        // Go version used for build
	Platform         string `json:"platform"`          // OS/architecture (e.g., "linux/amd64")
	Framework        string `json:"framework"`         // GUI framework name
	FrameworkVersion string `json:"framework_version"` // GUI framework version
}

// GetVersion returns complete version and build information.
//
// This function gathers all version metadata including runtime information
// (OS, architecture) and returns it in a structured format.
//
// Returns:
//   - Info: Complete version information structure
//
// Example:
//
//	info := core.GetVersion()
//	fmt.Printf("Version: %s\n", info.Version)
//	fmt.Printf("Platform: %s\n", info.Platform)
func GetVersion() Info {
	info := Info{
		Version:          Version,
		Commit:           Commit,
		Date:             FormatBuildDate(Date),
		GoVersion:        GoVersion,
		Platform:         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Framework:        Framework,
		FrameworkVersion: GetFrameworkVersion(),
	}

	logging.Debug.Log("Version info",
		"version", info.Version,
		"platform", info.Platform,
		"framework", info.Framework,
		"frameworkVersion", info.FrameworkVersion,
	)

	if info.Commit != "" {
		logging.Debug.Log("Build info",
			"commit", info.Commit,
			"date", info.Date,
		)
	}

	return info
}

// GetVersionString returns a formatted version string with build information.
//
// This function creates a human-readable version string that includes the
// version number and, if available, the short commit hash and build date.
//
// Returns:
//   - string: Formatted version string
//
// Examples:
//   - With build info: "0.0.1 (commit: 1a2b3c4) built on 2025-10-27T12:00:00Z"
//   - Without build info: "0.0.1"
func GetVersionString() string {
	info := GetVersion()

	// Build version string based on available information
	versionStr := info.Version

	if info.Commit != "" && len(info.Commit) > 7 {
		// Show short commit hash
		versionStr += fmt.Sprintf(" (commit: %s)", info.Commit[:7])
	} else if info.Commit != "" {
		versionStr += fmt.Sprintf(" (commit: %s)", info.Commit)
	}

	if info.Date != "" {
		versionStr += fmt.Sprintf(" built on %s", info.Date)
	}

	return versionStr
}

// GetSimpleVersion returns just the version number without build information.
//
// This function returns the bare version string without any additional metadata,
// useful when you need just the semantic version number.
//
// Returns:
//   - string: Version number (e.g., "0.0.1")
func GetSimpleVersion() string {
	return Version
}

// GetFullVersionString returns detailed multi-line version information.
//
// This function creates a comprehensive version display including application name,
// version, Go version, platform, and build information. Suitable for --version
// command output or about dialogs.
//
// Returns:
//   - string: Multi-line formatted version information
//
// Example output:
//
//	AWS Profile Manager
//	Version:    0.0.1
//	Go version: go1.21.0
//	Platform:   linux/amd64
//	Commit:     1a2b3c4d
//	Date:       2025-10-27T12:00:00Z
func GetFullVersionString() string {
	logging.Debug.Log("Generating full version string")
	info := GetVersion()

	result := fmt.Sprintf("%s\nVersion:    %s\nGo version: %s\nPlatform:   %s",
		AppName,
		info.Version,
		info.GoVersion,
		info.Platform)

	if info.FrameworkVersion != "unknown" {
		result += fmt.Sprintf("\nFramework:  %s %s", info.Framework, info.FrameworkVersion)
	} else {
		result += fmt.Sprintf("\nFramework:  %s", info.Framework)
	}

	if info.Commit != "" {
		result += fmt.Sprintf("\nCommit:     %s", info.Commit)
	}

	if info.Date != "" {
		result += fmt.Sprintf("\nBuilt:      %s", info.Date)
	}

	result += fmt.Sprintf("\nAuthor:     %s\nURL:        %s", AppAuthor, AppURL)

	logging.Debug.Log("Full version string generated")
	return result
}

// GetAppName returns the application name
func GetAppName() string {
	return AppName
}

// GetAppURL returns the application URL
func GetAppURL() string {
	return AppURL
}

// GetFrameworkVersion returns the detected GUI framework version.
//
// This function reads module build information embedded in the binary and
// returns the resolved Fyne dependency version when available.
//
// Returns:
//   - string: Framework dependency version, or "unknown" when unavailable
func GetFrameworkVersion() string {
	return getDependencyVersion("fyne.io/fyne/v2")
}

// getDependencyVersion returns the resolved version for a module dependency.
//
// This helper inspects build metadata for the current binary. It prefers the
// replacement version when the dependency is replaced, then falls back to the
// declared dependency version.
//
// Parameters:
//   - modulePath: Module path to locate in the embedded build info
//
// Returns:
//   - string: Detected module version, or "unknown" when unavailable
func getDependencyVersion(modulePath string) string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	for _, dependency := range buildInfo.Deps {
		if dependency.Path != modulePath {
			continue
		}

		if dependency.Replace != nil && dependency.Replace.Version != "" {
			return dependency.Replace.Version
		}

		if dependency.Version != "" {
			return dependency.Version
		}

		return "unknown"
	}

	return "unknown"
}
