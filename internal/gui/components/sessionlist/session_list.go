// Package sessionlist provides a self-contained, reusable AWS SSO session status component.
//
// The component owns all business logic: loading sessions via the awscli package,
// CLI login/logout, browser launch for the Console button, auto-refresh every
// 5 minutes, and the full UI including explanation text and bottom action bar.
//
// The parent view only needs to embed Content() and optionally surface
// GetLoadInfo() in its header.
//
// Usage:
//
// comp := sessionlist.New(window)
// header := viewheader.New("# Sessions", "...").
//
//	WithInfo(comp.GetLoadInfo().GetContent())
//
// return container.NewBorder(header.GetContent(), nil, nil, nil, comp.Content())
package sessionlist

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/components/actionbuttons"
	"aws-profile-manager/internal/gui/components/loadinfo"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// SessionList is a self-contained reusable component for displaying AWS SSO sessions.
//
// It owns the viewmodel, load info, refresh/logout buttons, auto-refresh ticker,
// and all login/logout/browser logic. Drop Content() anywhere you need sessions.
type SessionList struct {
	vm         *viewmodels.SessionsViewModel
	loadInfo   *loadinfo.LoadInfo
	window     fyne.Window
	lastResult *viewmodels.SessionsResult

	refreshButton   *widget.Button
	logoutButton    *widget.Button
	logoutDescLabel *widget.Label
	sessionRows     *fyne.Container
	rootContent     *fyne.Container
}

// New creates a fully self-contained SessionList component.
//
// Auto-refresh behaviour and interval are read from settings.AwsCLI at
// construction time. When AutoRefresh is disabled the refresh button label
// changes to "Check Sessions Manually" and no background ticker is started.
func New(window fyne.Window) *SessionList {
	slc := &SessionList{
		vm:          viewmodels.NewSessionsViewModel(),
		loadInfo:    loadinfo.NewLoadInfo(window),
		window:      window,
		sessionRows: container.NewVBox(),
	}

	// Read AWS CLI settings once — used for button labels and auto-refresh.
	awsCLISettings := settings.Get().AwsCLI
	autoRefresh := awsCLISettings.AutoRefresh
	refreshInterval := time.Duration(awsCLISettings.RefreshIntervalMins) * time.Minute

	// Refresh button label reflects whether auto-refresh is active.
	refreshButtonLabel := "Check Sessions"
	if !autoRefresh {
		refreshButtonLabel = "Check Sessions Manually"
	}

	// Refresh button
	slc.refreshButton = widget.NewButton(refreshButtonLabel, nil)
	slc.refreshButton.Importance = widget.HighImportance
	slc.refreshButton.OnTapped = func() {
		originalText := slc.refreshButton.Text
		slc.refreshButton.SetText("Checking...")
		slc.refreshButton.Disable()
		slc.load(true, func() {
			slc.refreshButton.SetText(originalText)
			slc.refreshButton.Enable()
		})
	}
	if window == nil {
		slc.refreshButton.Disable()
	}

	// Logout button — hidden until there are active sessions
	slc.logoutButton = widget.NewButton("Logout", nil)
	slc.logoutButton.OnTapped = func() {
		originalText := slc.logoutButton.Text
		slc.logoutButton.SetText("Logging out...")
		slc.logoutButton.Disable()
		slc.vm.LogoutAllSessions(func(err error) {
			if err != nil {
				_ = logging.Log.Error("Logout failed", "error", err)
			}
			fyne.Do(func() {
				slc.logoutButton.SetText(originalText)
				slc.logoutButton.Enable()
				slc.load(true, nil)
			})
		})
	}
	slc.logoutButton.Hide()

	// Bottom action bar — stacked vertically so labels wrap in narrow left panes.
	var autoCheckLabel string
	if autoRefresh {
		autoCheckLabel = fmt.Sprintf("Auto-check is %d minutes", awsCLISettings.RefreshIntervalMins)
	} else {
		autoCheckLabel = "Auto-check is disabled"
	}

	autoCheckLabelWidget := widget.NewLabel(autoCheckLabel)
	autoCheckLabelWidget.Wrapping = fyne.TextWrapWord

	logoutDescLabel := widget.NewLabel("Logout affects all AWS SSO sessions")
	logoutDescLabel.Wrapping = fyne.TextWrapWord
	slc.logoutDescLabel = logoutDescLabel
	slc.logoutDescLabel.Hide()

	autoCheckLabelWidget.Alignment = fyne.TextAlignCenter
	logoutDescLabel.Alignment = fyne.TextAlignCenter

	bottomBar := container.NewVBox(
		container.NewCenter(slc.refreshButton),
		autoCheckLabelWidget,
		container.NewCenter(slc.logoutButton),
		logoutDescLabel,
	)

	explanation := widget.NewLabel("These sessions are the currently installed AWS SSO sessions.")
	explanation.Wrapping = fyne.TextWrapWord

	slc.rootContent = container.NewPadded(container.NewVBox(
		explanation,
		slc.sessionRows,
		widget.NewSeparator(),
		bottomBar,
	))

	// Initial load
	slc.load(false, nil)

	// Auto-refresh: only start ticker when enabled in settings.
	if autoRefresh && refreshInterval > 0 {
		go func() {
			ticker := time.NewTicker(refreshInterval)
			defer ticker.Stop()
			for range ticker.C {
				fyne.Do(func() { slc.load(true, nil) })
			}
		}()
	}

	return slc
}

// Content returns the full component UI for embedding in any container.
func (slc *SessionList) Content() fyne.CanvasObject {
	return slc.rootContent
}

// GetLoadInfo returns the LoadInfo widget for use in a parent view's header.
func (slc *SessionList) GetLoadInfo() *loadinfo.LoadInfo {
	return slc.loadInfo
}

// load fetches sessions asynchronously, shows a progress dialog, and updates the UI.
// onDone is called on the main thread after the update (may be nil).
func (slc *SessionList) load(forceRefresh bool, onDone func()) {
	if slc.window == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	prog := components.ShowProgressDialog(slc.window, "Loading Sessions", "Checking SSO session status...", cancel)
	prog.Show()

	slc.vm.LoadSessions(ctx, forceRefresh, task.NoOpReporter{}, func(result *viewmodels.SessionsResult) {
		fyne.Do(func() {
			prog.Hide()

			if result.Error != nil {
				_ = logging.Log.Error("Failed to load sessions", "error", result.Error)
				slc.updateRows(nil)
				slc.logoutButton.Hide()
				slc.logoutDescLabel.Hide()
				if onDone != nil {
					onDone()
				}
				return
			}

			slc.lastResult = result
			slc.updateRows(result.Sessions)
			slc.loadInfo.SetSource(result.ConfigPath)
			slc.loadInfo.SetLoadedAt(result.LastChecked)

			if result.HasActiveSessions() {
				slc.logoutButton.Show()
				slc.logoutDescLabel.Show()
			} else {
				slc.logoutButton.Hide()
				slc.logoutDescLabel.Hide()
			}

			if onDone != nil {
				onDone()
			}
		})
	})
}

// updateRows rebuilds the session row list from the provided sessions.
func (slc *SessionList) updateRows(sessions []awscli.ActiveSessionInfo) {
	slc.sessionRows.Objects = nil

	if len(sessions) == 0 {
		slc.sessionRows.Add(container.NewCenter(widget.NewLabel("No AWS SSO sessions configured in AWS CLI")))
		slc.sessionRows.Refresh()
		return
	}

	for i, s := range sessions {
		slc.sessionRows.Add(slc.createSessionItem(s, i))
	}
	slc.sessionRows.Refresh()
}

// createSessionItem builds a single session row with alternating background.
func (slc *SessionList) createSessionItem(session awscli.ActiveSessionInfo, index int) fyne.CanvasObject {
	// Thin vertical status bar — same instance reused on both sides
	statusBar := canvas.NewRectangle(slc.getStatusColor(session))
	statusBar.Resize(fyne.NewSize(4, 50))

	statusLabel := widget.NewLabel(session.SessionName)
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	datetimeLabel := widget.NewLabel(slc.getDateTimeText(session))

	cliLoginButton := actionbuttons.SsoLogin(session.SessionName)
	consoleButton := actionbuttons.OpenURL(session.StartURL)

	if !session.IsExpired {
		cliLoginButton.Disable()
		consoleButton.Enable()
	} else {
		cliLoginButton.Enable()
		consoleButton.Enable()
	}

	actionsContainer := container.NewVBox(
		layout.NewSpacer(),
		container.NewHBox(cliLoginButton, consoleButton),
		layout.NewSpacer(),
	)

	sessionInfoContainer := container.NewVBox(statusLabel, datetimeLabel)

	mainLayout := container.NewHBox(
		container.NewPadded(statusBar),
		sessionInfoContainer,
		layout.NewSpacer(),
		actionsContainer,
		container.NewPadded(statusBar),
	)

	background := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	if index%2 != 0 {
		background.FillColor = theme.Color(theme.ColorNameOverlayBackground)
	}

	return container.NewStack(background, container.NewPadded(mainLayout))
}

// handleCLILogin and handleWebLogin are intentionally removed — buttons are
// created directly via actionbuttons.SsoLogin and actionbuttons.OpenURL.

// getStatusColor returns the theme colour appropriate for the session state.
func (slc *SessionList) getStatusColor(session awscli.ActiveSessionInfo) color.Color {
	if !session.IsExpired {
		return theme.Color(theme.ColorNameSuccess) // Green — active
	} else if session.ExpiresAt.IsZero() {
		return theme.Color(theme.ColorNamePrimary) // Blue — unknown/N/A
	}
	return theme.Color(theme.ColorNameError) // Red — expired
}

// getDateTimeText returns formatted expiry information for a session.
func (slc *SessionList) getDateTimeText(session awscli.ActiveSessionInfo) string {
	if session.ExpiresAt.IsZero() {
		return "Expiry: N/A"
	}
	local := session.ExpiresAt.Local()
	if !session.IsExpired {
		return fmt.Sprintf("SSO expires: %s", local.Format("2006-01-02 15:04 MST"))
	}
	return fmt.Sprintf("SSO expired: %s", local.Format("2006-01-02 15:04 MST"))
}
