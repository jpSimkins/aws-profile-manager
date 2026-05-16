package settings

// TerminalSettings holds configuration for the integrated terminal launcher.
//
// The terminal launcher opens a new OS terminal window pre-configured with
// AWS environment variables for a selected profile and region.
type TerminalSettings struct {
	// ExecutablePath is the path to the terminal emulator to use.
	// Leave empty to use the OS default (x-terminal-emulator on Linux,
	// Terminal.app on macOS, cmd.exe on Windows).
	ExecutablePath string `json:"executable_path"`
}

// GetDefaultTerminal returns the runtime default for terminal settings.
//
// Consumed by GetDefaults() to build the live *Settings struct on first launch.
// The Default values in GetSchema() are separate — those are UI metadata only.
// The default has no executable path set, which causes the launcher to detect
// the OS default terminal at runtime.
func GetDefaultTerminal() TerminalSettings {
	return TerminalSettings{
		ExecutablePath: "",
	}
}

// Validate validates terminal settings.
//
// ExecutablePath is optional — an empty value is valid and means "use OS default".
func (s *TerminalSettings) Validate() error {
	// No validation required; empty path is valid.
	return nil
}

// GetSchema returns the schema definition for terminal settings.
//
// The Default value in each FieldSchema is UI metadata for the settings form
// renderer. Runtime defaults live in GetDefaultTerminal() and must stay in sync.
func (s *TerminalSettings) GetSchema() Schema {
	return Schema{
		Version:     "1.0",
		Description: "Configure the terminal emulator used when opening an AWS shell session from the profile list.",
		Fields: map[string]FieldSchema{
			"executable_path": {
				Type:        "file",
				Label:       "Terminal Executable",
				Description: "Path to the terminal emulator (leave empty to use the OS default). Examples: /usr/bin/gnome-terminal, /usr/bin/kitty, C:\\Windows\\System32\\cmd.exe",
				Required:    false,
				Default:     "",
				Order:       1,
			},
		},
	}
}
