package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// LaunchSession opens a new terminal window with the AWS environment variables
// from cfg pre-set in the shell's environment.
//
// This function is the primary reuse point — it is also called for SSO login
// prompts and any other operation that needs a pre-configured AWS shell session.
//
// The terminal executable is resolved from cfg.TerminalPath, then the
// application settings, then the OS default (see ResolveTerminal).
func LaunchSession(cfg LaunchConfig) error {
	terminal, err := ResolveTerminal(cfg.TerminalPath)
	if err != nil {
		return fmt.Errorf("could not resolve terminal: %w", err)
	}

	env := cfg.envVars()

	switch runtime.GOOS {
	case "darwin":
		return launchMacOS(terminal, env, cfg.Command, cfg.ExitOnComplete)
	case "windows":
		return launchWindows(terminal, env, cfg.Command, cfg.ExitOnComplete)
	default:
		return launchLinux(terminal, env, cfg.Command, cfg.ExitOnComplete)
	}
}

// launchMacOS opens a new Terminal.app window using AppleScript so that the
// environment variables are visible to the interactive shell inside.
//
// When command is non-empty it is appended after the export statements so it
// runs immediately when the window opens. Unless ExitOnComplete is true the
// terminal stays open after the command exits (Terminal.app default behaviour).
func launchMacOS(_ string, env map[string]string, command string, exitOnComplete bool) error {
	exports := buildExportStatements(env)
	parts := make([]string, len(exports))
	copy(parts, exports)
	if command != "" {
		parts = append(parts, command)
		if exitOnComplete {
			parts = append(parts, "exit")
		}
	}
	script := fmt.Sprintf(
		`tell application "Terminal" to do script "%s"`,
		strings.Join(parts, "; "),
	)
	cmd := exec.Command("osascript", "-e", script)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch terminal via osascript: %w", err)
	}
	return nil
}

// launchWindows opens a new cmd.exe (or Windows Terminal) window with the
// AWS environment variables set via the `set` command.
//
// When command is non-empty it is run after the env vars are set. Unless
// ExitOnComplete is true an interactive cmd session opens afterwards so the
// user can inspect the output.
func launchWindows(terminal string, env map[string]string, command string, exitOnComplete bool) error {
	var sets []string
	for k, v := range env {
		sets = append(sets, fmt.Sprintf("set %s=%s", k, v))
	}
	if command != "" {
		sets = append(sets, command)
	}
	if !exitOnComplete {
		sets = append(sets, "cmd /k") // drop into interactive shell when done
	}

	args := []string{"/c", "start", terminal, "/k", strings.Join(sets, " & ")}
	cmd := exec.Command("cmd.exe", args...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch terminal on Windows: %w", err)
	}
	return nil
}

// launchLinux opens a new terminal window, passing the AWS environment via
// the terminal's -e flag to execute a shell with the variables exported.
//
// When command is non-empty it is inserted before the interactive shell so it
// runs immediately and the user sees its output. When ExitOnComplete is true
// the `exec bash --login` tail is omitted so the window closes on completion.
func launchLinux(terminal string, env map[string]string, command string, exitOnComplete bool) error {
	exports := buildExportStatements(env)
	// Build a shell snippet: export vars, optionally run command, then open
	// an interactive login shell. exec bash replaces the intermediate shell.
	shellCmd := strings.Join(exports, "; ")
	if command != "" {
		shellCmd += "; " + command
	}
	if !exitOnComplete {
		shellCmd += "; exec bash --login"
	}

	args := terminalArgs(terminal, shellCmd)
	cmd := exec.Command(terminal, args...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch terminal %q: %w", terminal, err)
	}
	return nil
}

// terminalArgs returns the correct argument list for each known terminal to
// execute a shell command on startup.
func terminalArgs(terminal, shellCmd string) []string {
	base := lastSegment(terminal)
	switch base {
	case "gnome-terminal":
		// gnome-terminal uses -- to separate its own args from the command
		return []string{"--", "bash", "-c", shellCmd}
	case "konsole":
		return []string{"-e", "bash", "-c", shellCmd}
	case "tilix":
		return []string{"-e", "bash -c '" + shellCmd + "'"}
	default:
		// xterm, xfce4-terminal, mate-terminal, lxterminal, kitty, alacritty,
		// x-terminal-emulator all accept -e <cmd>
		return []string{"-e", "bash -c '" + shellCmd + "'"}
	}
}

// buildExportStatements converts an env map into a slice of `export K=V` statements.
func buildExportStatements(env map[string]string) []string {
	stmts := make([]string, 0, len(env))
	for k, v := range env {
		stmts = append(stmts, fmt.Sprintf("export %s=%s", k, v))
	}
	return stmts
}

// lastSegment returns the base name of a path (after the last /).
func lastSegment(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
