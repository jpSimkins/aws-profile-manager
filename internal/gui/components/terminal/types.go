// Package terminal provides cross-platform terminal launching with AWS
// environment variables pre-set in the new terminal session.
//
// # Reuse
//
// The LaunchSession function is designed to be reused for other shell-entry
// operations such as SSO login prompts. Callers simply provide a LaunchConfig
// specifying the profile, region, and any additional environment variables.
//
// # Terminal Resolution
//
// The terminal executable is resolved in the following order:
//  1. The value configured in application settings (if non-empty)
//  2. The OS-specific default (see ResolveTerminal)
//
// # Environment Variables
//
// Every launched terminal session receives at minimum:
//
//	AWS_PROFILE           = <profile name>
//	AWS_DEFAULT_REGION    = <region>
//	AWS_SDK_LOAD_CONFIG   = 1   (ensures AWS CLI config is picked up)
//
// Additional variables can be passed via LaunchConfig.ExtraEnv.
package terminal

// LaunchConfig holds everything needed to open a terminal for a specific AWS context.
type LaunchConfig struct {
	// ProfileName is the AWS CLI profile name, e.g. "commercial-acme-prod-Admin".
	// Leave empty when the session does not require a specific profile (e.g. SSO login).
	ProfileName string

	// Region is the AWS region to set as AWS_DEFAULT_REGION.
	// Leave empty when not relevant (e.g. SSO login).
	Region string

	// Command is an optional shell command to run inside the new terminal before
	// dropping into an interactive shell. When empty, the terminal opens directly
	// to an interactive shell with the AWS environment pre-set.
	//
	// Example: "aws sso login --sso-session acme-commercial"
	Command string

	// ExtraEnv contains any additional environment variables to inject, keyed by
	// variable name. These are merged on top of the standard AWS vars.
	ExtraEnv map[string]string

	// TerminalPath overrides the resolved terminal executable for this launch.
	// Leave empty to use settings or the OS default.
	TerminalPath string
}

// envVars builds the map of environment variables to inject into the terminal.
//
// AWS_PROFILE and AWS_DEFAULT_REGION are only included when non-empty -
// injecting an empty value causes the AWS CLI to look up a profile named ""
// and fail with "The config profile () could not be found".
func (c LaunchConfig) envVars() map[string]string {
	env := map[string]string{
		"AWS_SDK_LOAD_CONFIG": "1",
	}
	if c.ProfileName != "" {
		env["AWS_PROFILE"] = c.ProfileName
	}
	if c.Region != "" {
		env["AWS_DEFAULT_REGION"] = c.Region
	}
	for k, v := range c.ExtraEnv {
		env[k] = v
	}
	return env
}
