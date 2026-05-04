// Package actionbuttons provides small, reusable icon-only action buttons used
// throughout the GUI — terminal launch, SSO login, browser open, and clipboard copy.
//
// All buttons use widget.LowImportance so they appear as lightweight icon controls
// without a filled background, keeping rows visually clean.
//
// Usage:
//
//	container.NewHBox(
//	    actionbuttons.SsoLogin(sessionName),
//	    actionbuttons.OpenURL(startURL),
//	    actionbuttons.Copy(profileName),
//	)
package actionbuttons

import (
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components/terminal"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// Terminal returns a small icon button (theme.ComputerIcon) that opens a new
// terminal window with AWS_PROFILE and AWS_DEFAULT_REGION pre-set.
//
// The terminal executable is read from application settings at tap time so
// that settings changes take effect without restarting the app.
func Terminal(profileName, region string) *widget.Button {
	btn := widget.NewButtonWithIcon("", theme.ComputerIcon(), func() {
		cfg := terminal.LaunchConfig{
			ProfileName:  profileName,
			Region:       region,
			TerminalPath: settings.Get().Terminal.ExecutablePath,
		}
		if err := terminal.LaunchSession(cfg); err != nil {
			_ = logging.Log.Error("Failed to launch terminal",
				"profile", profileName,
				"error", err,
			)
		}
	})
	btn.Importance = widget.LowImportance
	return btn
}

// SsoLogin returns a small icon button (theme.LoginIcon) that opens a terminal
// and runs `aws sso login --sso-session <sessionName>`.
//
// The terminal stays open after the command completes so the user can see
// any output or interact with the browser authentication prompt.
func SsoLogin(sessionName string) *widget.Button {
	btn := widget.NewButtonWithIcon("", theme.LoginIcon(), func() {
		cfg := terminal.LaunchConfig{
			Command:      "aws sso login --sso-session " + sessionName,
			TerminalPath: settings.Get().Terminal.ExecutablePath,
		}
		if err := terminal.LaunchSession(cfg); err != nil {
			_ = logging.Log.Error("Failed to launch SSO login terminal",
				"session", sessionName,
				"error", err,
			)
		}
	})
	btn.Importance = widget.LowImportance
	return btn
}

// OpenURL returns a small icon button (theme.HomeIcon) that opens rawURL in
// the OS default browser. The button is disabled when rawURL is empty.
func OpenURL(rawURL string) *widget.Button {
	btn := widget.NewButtonWithIcon("", theme.HomeIcon(), func() {
		u, err := url.Parse(rawURL)
		if err != nil {
			_ = logging.Log.Error("Failed to parse URL",
				"url", rawURL,
				"error", err,
			)
			return
		}
		if err := fyne.CurrentApp().OpenURL(u); err != nil {
			_ = logging.Log.Error("Failed to open URL in browser",
				"url", rawURL,
				"error", err,
			)
		}
	})
	btn.Importance = widget.LowImportance
	if rawURL == "" {
		btn.Disable()
	}
	return btn
}

// Copy returns a small icon button (theme.ContentCopyIcon) that writes value
// to the clipboard. On tap it briefly shows theme.ConfirmIcon for 500ms as
// visual confirmation, then resets to the copy icon.
func Copy(value string) *widget.Button {
	var btn *widget.Button
	btn = widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Clipboard().SetContent(value)
		btn.SetIcon(theme.ConfirmIcon())
		go func() {
			time.Sleep(500 * time.Millisecond)
			fyne.Do(func() { btn.SetIcon(theme.ContentCopyIcon()) })
		}()
	})
	btn.Importance = widget.LowImportance
	return btn
}
