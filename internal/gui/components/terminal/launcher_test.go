package terminal

import (
	"runtime"
	"testing"
)

// --- ResolveTerminal ---

func TestResolveTerminal_UsesOverride(t *testing.T) {
	got, err := ResolveTerminal("/usr/bin/my-terminal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/bin/my-terminal" {
		t.Errorf("expected override unchanged; got %q", got)
	}
}

func TestResolveTerminal_TrimsWhitespace(t *testing.T) {
	got, err := ResolveTerminal("  /usr/bin/xterm  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/usr/bin/xterm" {
		t.Errorf("expected trimmed override; got %q", got)
	}
}

func TestResolveTerminal_WhitespaceOnlyFallsBackToOS(t *testing.T) {
	// Whitespace-only string must be treated the same as empty string.
	terminal, err := ResolveTerminal("   ")
	if runtime.GOOS == "darwin" {
		if err != nil {
			t.Fatalf("expected fallback on macOS, got error: %v", err)
		}
		if terminal != "open" {
			t.Errorf("expected 'open' on macOS fallback, got %q", terminal)
		}
	} else {
		_ = terminal
		_ = err // other platforms may or may not have a terminal in CI
	}
}

func TestResolveTerminal_EmptyOverrideFallsBackToOS(t *testing.T) {
	terminal, err := ResolveTerminal("")
	if runtime.GOOS == "darwin" {
		if err != nil {
			t.Fatalf("expected 'open' to be resolved on macOS, got error: %v", err)
		}
		if terminal != "open" {
			t.Errorf("expected 'open' on macOS, got %q", terminal)
		}
	} else {
		// On Linux/Windows: function must not panic; result may be empty in CI
		_ = terminal
		_ = err
	}
}

// --- linuxTerminalCandidates ---

func TestLinuxTerminalCandidates_NotEmpty(t *testing.T) {
	if len(linuxTerminalCandidates) == 0 {
		t.Fatal("linuxTerminalCandidates must not be empty")
	}
}

func TestLinuxTerminalCandidates_ContainsGnomeTerminal(t *testing.T) {
	for _, c := range linuxTerminalCandidates {
		if c == "gnome-terminal" {
			return
		}
	}
	t.Error("linuxTerminalCandidates must include gnome-terminal as the Ubuntu fallback")
}

func TestLinuxTerminalCandidates_ContainsXTerminalEmulator(t *testing.T) {
	// x-terminal-emulator should be first (highest priority on Debian/Ubuntu).
	if linuxTerminalCandidates[0] != "x-terminal-emulator" {
		t.Errorf("expected x-terminal-emulator to be first candidate, got %q", linuxTerminalCandidates[0])
	}
}

func TestLinuxTerminalCandidates_NoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range linuxTerminalCandidates {
		if seen[c] {
			t.Errorf("duplicate entry in linuxTerminalCandidates: %q", c)
		}
		seen[c] = true
	}
}

// --- terminalArgs ---

func TestTerminalArgs_GnomeTerminal_StartsWithDoubleDash(t *testing.T) {
	args := terminalArgs("/usr/bin/gnome-terminal", "export AWS_PROFILE=prod; exec bash --login")
	if len(args) == 0 {
		t.Fatal("expected non-empty args for gnome-terminal")
	}
	if args[0] != "--" {
		t.Errorf("expected first arg '--' for gnome-terminal, got %q", args[0])
	}
}
