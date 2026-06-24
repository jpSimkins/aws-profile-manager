package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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
// the `exec <shell> --login` tail is omitted so the window closes on completion.
//
// The user's default shell (from $SHELL) is used, respecting shell frameworks
// like oh-my-zsh, fish, nushell, etc.
func launchLinux(terminal string, env map[string]string, command string, exitOnComplete bool) error {
	exports := buildExportStatements(env)
	shell := getUserShell()

	// Build a shell snippet: export vars, optionally run command, then open
	// an interactive login shell. exec <shell> replaces the intermediate shell.
	shellCmd := strings.Join(exports, "; ")
	if command != "" {
		shellCmd += "; " + command
	}
	if !exitOnComplete {
		shellCmd += fmt.Sprintf("; exec %s --login", shell)
	}

	args := terminalArgs(terminal, shellCmd, shell)
	cmd := exec.Command(terminal, args...)
	cmd.Env = os.Environ()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch terminal %q: %w", terminal, err)
	}
	return nil
}

// terminalArgs returns the correct argument list for each known terminal to
// execute a shell command on startup.
//
// Most modern terminals support passing a command via -e followed by separate
// arguments, avoiding complex shell quoting. For terminals that require it,
// the command is properly shell-escaped using strconv.Quote.
//
// The shell parameter is used for terminals that execute shell commands directly,
// and allows respecting the user's shell choice (bash, zsh, fish, etc.).
func terminalArgs(terminal, shellCmd, shell string) []string {
	base := lastSegment(terminal)
	switch base {
	case "ptyxis", "gnome-terminal", "gnome-console":
		// ptyxis (Fedora 41+), gnome-terminal, gnome-console: -- separates terminal args from command
		// Modern GNOME terminals support separate args to avoid quoting issues
		return []string{"--", shell, "-c", shellCmd}
	case "konsole":
		// KDE Konsole: -e runs a command, supports separate arguments
		return []string{"-e", shell, "-c", shellCmd}
	case "xfce4-terminal", "mate-terminal", "xterm":
		// XFCE, MATE, xterm: -e runs a command
		return []string{"-e", shell, "-c", shellCmd}
	case "lxterminal":
		// LXTerminal: -e runs a command
		return []string{"-e", shell, "-c", shellCmd}
	case "kitty":
		// Kitty: supports standard shell invocation
		return []string{"sh", "-c", shellCmd}
	case "alacritty":
		// Alacritty: -e specifies command to run
		return []string{"-e", shell, "-c", shellCmd}
	case "tilix":
		// Tilix: -e runs a command, requires proper quoting
		return []string{"-e", shellCmd}
	case "rxvt-unicode", "urxvt":
		// rxvt-unicode: -e specifies command
		return []string{"-e", shell, "-c", shellCmd}
	default:
		// Fallback for unknown terminals: assume -e accepts command + args
		return []string{"-e", shell, "-c", shellCmd}
	}
}

// buildExportStatements converts an env map into a slice of `export K=V` statements.
// Values are properly escaped to handle special characters and spaces.
func buildExportStatements(env map[string]string) []string {
	stmts := make([]string, 0, len(env))
	for k, v := range env {
		// Use strconv.Quote for proper shell escaping of the value
		// This handles spaces, special chars, and shell metacharacters safely
		escapedValue := strconv.Quote(v)
		stmts = append(stmts, fmt.Sprintf("export %s=%s", k, escapedValue))
	}
	return stmts
}

// lastSegment returns the base name of a path (after the last /).
func lastSegment(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// getUserShell returns the user's default shell, respecting shell frameworks
// like oh-my-zsh, fish, nushell, and others.
//
// First tries the SHELL environment variable, then falls back to bash if
// the environment variable is not set. This allows shell frameworks to work
// transparently — if the user has configured SHELL=/usr/bin/zsh with oh-my-zsh,
// that's what will be used.
func getUserShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		// Extract just the shell name (e.g., /usr/bin/zsh -> zsh)
		return lastSegment(shell)
	}
	// Fallback to bash if SHELL is not set
	return "bash"
}
