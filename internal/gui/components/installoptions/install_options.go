// Package installoptions provides a component for configuring profile installation options.
//
// The install options component displays form inputs for:
//   - Output file path (AWS config file)
//   - Cheat sheet file path
//   - EOL format selection (Native, LF, CRLF)
//   - Generate cheatsheet checkbox
//   - Cheatsheet only checkbox
package installoptions

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/settings"
)

// EolFormat represents end-of-line format options.
type EolFormat string

const (
	// EolNative uses the OS-native line endings
	EolNative EolFormat = "Native"

	// EolLF uses Unix-style line endings (\n)
	EolLF EolFormat = "LF"

	// EolCRLF uses Windows-style line endings (\r\n)
	EolCRLF EolFormat = "CRLF"
)

// InstallOptions provides a UI component for installation options.
//
// The component displays:
//   - Output file input with browse button (default: ~/.aws/config)
//   - Cheat sheet file input with browse button (default: ~/Desktop/aws-profile-cheatsheet.md)
//   - EOL format dropdown (default: Native)
//   - Generate cheatsheet checkbox
//   - Cheatsheet only checkbox
type InstallOptions struct {
	window              fyne.Window
	outputFileEntry     *widget.Entry
	cheatsheetFileEntry *widget.Entry
	eolFormatSelect     *widget.Select
	generateCheatCheck  *widget.Check
	cheatsheetOnlyCheck *widget.Check
}

// NewInstallOptions creates a new install options component.
//
// Initializes with defaults from settings:
//   - Output file: ~/.aws/config
//   - Cheat sheet file: ~/Desktop/aws-profile-cheatsheet.md
//   - EOL format: Native (OS default at write time)
//   - Generate cheatsheet: Checked
//   - Cheatsheet only: Unchecked
//
// The window parameter is used to attach file-picker dialogs. Pass nil when
// running headless (tests, non-GUI mode) — browse buttons will be hidden.
func NewInstallOptions(window fyne.Window) *InstallOptions {
	io := &InstallOptions{window: window}

	// Output file
	io.outputFileEntry = widget.NewEntry()
	io.outputFileEntry.SetText(filepath.Join(settings.GetAwsDir(), "config"))
	io.outputFileEntry.SetPlaceHolder("Path to AWS config file")

	// Cheat sheet file
	desktopDir := settings.GetDesktopDir()
	defaultCheatsheet := filepath.Join(desktopDir, "aws-profile-cheatsheet.md")
	io.cheatsheetFileEntry = widget.NewEntry()
	io.cheatsheetFileEntry.SetText(defaultCheatsheet)
	io.cheatsheetFileEntry.SetPlaceHolder("Path to cheat sheet file")

	// EOL format — Native by default so the OS decides at write time
	io.eolFormatSelect = widget.NewSelect(
		[]string{string(EolNative), string(EolLF), string(EolCRLF)},
		nil,
	)
	io.eolFormatSelect.SetSelected(string(EolNative))

	// Checkboxes
	io.generateCheatCheck = widget.NewCheck("Generate cheatsheet", nil)
	io.generateCheatCheck.SetChecked(true)
	io.cheatsheetOnlyCheck = widget.NewCheck("Cheatsheet only (skip profile installation)", nil)

	return io
}

// GetContent returns the UI content for install options.
//
// Each file path row shows the (shortened) path and a Browse button that opens
// a file-save dialog. When window is nil the Browse buttons are hidden.
func (io *InstallOptions) GetContent() fyne.CanvasObject {
	header := widget.NewRichTextFromMarkdown("## Installation Options")

	outputRow := io.filePickerRow(io.outputFileEntry, "Output file:")
	cheatsheetRow := io.filePickerRow(io.cheatsheetFileEntry, "Cheat sheet file:")

	eolLabel := widget.NewLabel("EOL Format:")

	return container.NewPadded(container.NewVBox(
		header,
		outputRow,
		cheatsheetRow,
		eolLabel,
		io.eolFormatSelect,
		io.generateCheatCheck,
		io.cheatsheetOnlyCheck,
	))
}

// filePickerRow builds a label + entry + Browse button row for a file path.
//
// Opens a file-save dialog so the user can both pick an existing file and type
// a new filename. The entry is updated when the user confirms.
func (io *InstallOptions) filePickerRow(entry *widget.Entry, label string) fyne.CanvasObject {
	labelWidget := widget.NewLabel(label)

	// Show the path relative to the home directory when possible.
	// The backing entry always holds the absolute path — only displayEntry shows ~/...
	displayEntry := widget.NewEntry()
	displayEntry.SetText(shortenPath(entry.Text))
	displayEntry.OnChanged = func(v string) {
		entry.SetText(expandPath(v))
	}
	entry.OnChanged = func(v string) {
		displayEntry.SetText(shortenPath(v))
	}

	if io.window == nil {
		return container.NewVBox(labelWidget, displayEntry)
	}

	browseBtn := widget.NewButton("Browse...", func() {
		d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil || w == nil {
				return
			}
			// Grab the path then immediately remove the empty placeholder file
			// that the OS file dialog creates on confirmation — the real file
			// will be written by the installer when it runs.
			path := w.URI().Path()
			_ = w.Close()
			_ = os.Remove(path)
			entry.SetText(path)
		}, io.window)
		_ = setDialogDir(d, entry.Text)
		d.Show()
	})
	browseBtn.Importance = widget.LowImportance

	return container.NewVBox(
		labelWidget,
		container.NewBorder(nil, nil, nil, browseBtn, displayEntry),
	)
}

// GetOutputFile returns the output file path.
func (io *InstallOptions) GetOutputFile() string {
	return io.outputFileEntry.Text
}

// SetOutputFile sets the output file path.
func (io *InstallOptions) SetOutputFile(path string) {
	io.outputFileEntry.SetText(path)
}

// GetCheatsheetFile returns the cheat sheet file path.
func (io *InstallOptions) GetCheatsheetFile() string {
	return io.cheatsheetFileEntry.Text
}

// SetCheatsheetFile sets the cheat sheet file path.
func (io *InstallOptions) SetCheatsheetFile(path string) {
	io.cheatsheetFileEntry.SetText(path)
}

// GetEolFormat returns the selected EOL format.
func (io *InstallOptions) GetEolFormat() EolFormat {
	return EolFormat(io.eolFormatSelect.Selected)
}

// SetEolFormat sets the EOL format selection.
func (io *InstallOptions) SetEolFormat(format EolFormat) {
	io.eolFormatSelect.SetSelected(string(format))
}

// ShouldGenerateCheatsheet returns whether to generate cheat sheet.
func (io *InstallOptions) ShouldGenerateCheatsheet() bool {
	return io.generateCheatCheck.Checked
}

// SetGenerateCheatsheet sets the generate cheatsheet checkbox.
func (io *InstallOptions) SetGenerateCheatsheet(generate bool) {
	io.generateCheatCheck.SetChecked(generate)
}

// IsCheatsheetOnly returns whether to skip profile installation.
func (io *InstallOptions) IsCheatsheetOnly() bool {
	return io.cheatsheetOnlyCheck.Checked
}

// SetCheatsheetOnly sets the cheatsheet only checkbox.
func (io *InstallOptions) SetCheatsheetOnly(cheatsheetOnly bool) {
	io.cheatsheetOnlyCheck.SetChecked(cheatsheetOnly)
}

// Reset resets all options to defaults.
//
// Restores:
//   - Output file to ~/.aws/config
//   - Cheat sheet file to ~/Desktop/aws-profile-cheatsheet.md
//   - EOL format to Native
//   - All checkboxes unchecked
func (io *InstallOptions) Reset() {
	io.outputFileEntry.SetText(filepath.Join(settings.GetAwsDir(), "config"))
	desktopDir := settings.GetDesktopDir()
	io.cheatsheetFileEntry.SetText(filepath.Join(desktopDir, "aws-profile-cheatsheet.md"))
	io.eolFormatSelect.SetSelected(string(EolNative))
	io.generateCheatCheck.SetChecked(false)
	io.cheatsheetOnlyCheck.SetChecked(false)
}

// shortenPath returns the path relative to the user's home directory when
// possible, prefixing it with "~/". Returns the original path otherwise.
func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~/" + strings.TrimPrefix(path, home+"/")
	}
	return path
}

// expandPath converts a "~/"-prefixed path back to an absolute path.
// Returns the original string unchanged if it doesn't start with "~/".
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	return filepath.Join(home, path[2:])
}

// setDialogDir positions a file dialog to start in the directory containing
// the current path. Silently ignores errors (the dialog opens at its default).
func setDialogDir(d interface{ SetLocation(fyne.ListableURI) }, path string) error {
	dir := filepath.Dir(path)
	u, err := storage.ListerForURI(storage.NewFileURI(dir))
	if err != nil {
		return err
	}
	d.SetLocation(u)
	return nil
}
