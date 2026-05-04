package terminal

import (
	"runtime"
	"testing"
)

// LaunchSession cannot be tested end-to-end without spawning a real terminal
// window, which is not appropriate in automated tests. The tests here verify:
//   - Error propagation when terminal resolution fails
//   - Correct OS routing (no panic on any platform)
//   - Individual platform helper functions where testable in isolation

// --- LaunchSession ---

func TestLaunchSession_BadTerminalPathReturnsError(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("macOS uses osascript which ignores TerminalPath; skipping")
	}

	cfg := LaunchConfig{
		ProfileName:  "commercial-acme-prod-Admin",
		Region:       "us-east-1",
		TerminalPath: "/nonexistent/terminal-that-does-not-exist",
	}

	// LaunchSession should return an error when the given terminal cannot be started.
	err := LaunchSession(cfg)
	if err == nil {
		t.Error("expected error when terminal path does not exist")
	}
}

func TestLaunchSession_EmptyProfileAndRegion_DoesNotPanic(t *testing.T) {
	// We cannot assert success here (no real terminal in CI), but we can
	// assert the function does not panic on an empty config.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LaunchSession panicked: %v", r)
		}
	}()

	cfg := LaunchConfig{
		ProfileName:  "",
		Region:       "",
		TerminalPath: "/nonexistent/terminal",
	}
	// Error is expected; panic is not.
	_ = LaunchSession(cfg)
}

// --- launchLinux (white-box, non-spawning) ---

func TestLaunchLinux_BadTerminalReturnsError(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("skipping launchLinux test on non-Unix platform")
	}

	env := map[string]string{"AWS_PROFILE": "prod", "AWS_DEFAULT_REGION": "us-east-1"}
	err := launchLinux("/nonexistent/terminal-xyz", env, "")
	if err == nil {
		t.Error("expected error for nonexistent terminal on Linux")
	}
}

// --- launchMacOS (white-box, non-spawning) ---

func TestLaunchMacOS_DoesNotPanicOnEmptyEnv(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping launchMacOS test on non-macOS platform")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("launchMacOS panicked: %v", r)
		}
	}()

	// We pass an empty env map — it will produce an empty AppleScript command.
	// osascript may or may not succeed depending on the environment, but it
	// must not panic.
	_ = launchMacOS("open", map[string]string{}, "")
}

// --- Command field ---

func TestLaunchSession_WithCommand_BadTerminalReturnsError(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("macOS uses osascript which ignores TerminalPath; skipping")
	}

	cfg := LaunchConfig{
		Command:      "aws sso login --sso-session test-session",
		TerminalPath: "/nonexistent/terminal-that-does-not-exist",
	}
	if err := LaunchSession(cfg); err == nil {
		t.Error("expected error when terminal path does not exist (with Command set)")
	}
}

func TestLaunchLinux_WithCommand_DoesNotPanic(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("skipping launchLinux test on non-Unix platform")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("launchLinux panicked with command: %v", r)
		}
	}()

	env := map[string]string{"AWS_PROFILE": "prod"}
	// A nonexistent terminal will error, not panic.
	_ = launchLinux("/nonexistent/terminal-xyz", env, "aws sso login --sso-session test")
}
