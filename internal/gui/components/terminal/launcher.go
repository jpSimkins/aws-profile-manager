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
//
// Order prioritizes:
//  1. x-terminal-emulator (Debian/Ubuntu standard alias)
//  2. Modern desktop environments (ptyxis, gnome-console, gnome-terminal, konsole)
//  3. Lightweight terminals (Tilix, kitty, alacritty)
//  4. Legacy fallbacks (xterm)
//
// This order works across Ubuntu, Fedora, Debian, openSUSE, and other major distros.
// Shell frameworks (oh-my-zsh, zsh, fish, etc.) work automatically within any terminal.
var linuxTerminalCandidates = []string{
	"x-terminal-emulator", // Debian/Ubuntu standard (highest priority)
	"ptyxis",              // GNOME (Fedora 41+ default)
	"gnome-console",       // GNOME 43+ (alternative modern GNOME terminal)
	"gnome-terminal",      // GNOME (older versions, Ubuntu fallback)
	"konsole",             // KDE/Plasma (Fedora KDE Spin)
	"xfce4-terminal",      // XFCE (lightweight)
	"mate-terminal",       // MATE (lightweight)
	"lxterminal",          // LXDE (lightweight)
	"tilix",               // Tilix (modern, feature-rich)
	"kitty",               // kitty (modern, GPU-based)
	"alacritty",           // Alacritty (modern, GPU-based)
	"rxvt-unicode",        // rxvt-unicode (lightweight)
	"urxvt",               // urxvt alias
	"xterm",               // xterm (fallback: requires X11)
}

// resolveLinuxTerminal finds the first available terminal in linuxTerminalCandidates.
//
// The function tries each candidate in order and returns the full path to the
// first one found in the system PATH. If none are found, an error is returned
// with installation instructions.
func resolveLinuxTerminal() (string, error) {
	for _, candidate := range linuxTerminalCandidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no supported terminal emulator found\n\n" +
		"Install one of the following:\n" +
		"  Ubuntu/Debian:  sudo apt install gnome-terminal (or konsole, xfce4-terminal)\n" +
		"  Fedora 41+:     sudo dnf install ptyxis (or gnome-terminal, konsole)\n" +
		"  Fedora <41:     sudo dnf install gnome-terminal (or konsole, xfce4-terminal)\n" +
		"  openSUSE:       sudo zypper install gnome-terminal (or konsole, xfce4-terminal)\n" +
		"  Arch/Manjaro:   sudo pacman -S gnome-terminal (or konsole, xfce4-terminal)\n\n" +
		"Or configure a custom terminal in Settings.",
	)
}
