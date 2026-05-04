package terminal

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ResolveTerminal returns the terminal executable to use.
//
// Resolution order:
//  1. override — used as-is when non-empty (e.g. from settings or a caller)
//  2. OS-specific default (see resolveOSDefault)
func ResolveTerminal(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	return resolveOSDefault()
}

// resolveOSDefault returns the best available terminal for the current OS.
//
// macOS:   open -a Terminal (always available)
// Windows: cmd.exe (always available; Windows Terminal used when present)
// Linux:   checks a prioritised candidate list, falls back to gnome-terminal
func resolveOSDefault() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// macOS: "open -a Terminal" is always available
		return "open", nil
	case "windows":
		// Prefer Windows Terminal if installed, fall back to cmd.exe
		if wt, err := exec.LookPath("wt.exe"); err == nil {
			return wt, nil
		}
		if cmd, err := exec.LookPath("cmd.exe"); err == nil {
			return cmd, nil
		}
		return "", fmt.Errorf("no terminal found on Windows")
	default:
		// Linux and other Unix-likes: try candidates in preference order
		return resolveLinuxTerminal()
	}
}

// linuxTerminalCandidates is the ordered list of terminals to try on Linux.
// gnome-terminal is last because it is the guaranteed fallback for Ubuntu users.
var linuxTerminalCandidates = []string{
	"x-terminal-emulator", // Debian/Ubuntu alias pointing to the user's default
	"konsole",             // KDE
	"xfce4-terminal",      // XFCE
	"mate-terminal",       // MATE
	"lxterminal",          // LXDE
	"tilix",               // Tilix
	"kitty",               // kitty
	"alacritty",           // Alacritty
	"gnome-terminal",      // GNOME — explicit fallback for Ubuntu
	"xterm",               // Last resort: always available if X11 is installed
}

// resolveLinuxTerminal finds the first available terminal in linuxTerminalCandidates.
func resolveLinuxTerminal() (string, error) {
	for _, candidate := range linuxTerminalCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no supported terminal emulator found; install gnome-terminal or set a terminal in Settings")
}
