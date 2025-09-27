package viewmodels

import (
	"context"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"aws-profile-manager/internal/awscli"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

// SessionsViewModel manages state and business logic for the Sessions view.
//
// Follows MVVM: the view only handles UI wiring; all data fetching and
// formatting lives here. Session status is loaded asynchronously via the
// task package and the viewmodel is safe to call from multiple goroutines.
type SessionsViewModel struct {
	IsLoading bool
	mu        sync.RWMutex
}

// SessionsResult contains the loaded session data ready for display.
//
// Sessions contains ALL configured SSO sessions, with IsExpired reflecting
// whether a valid token exists. Active sessions come first (sorted by name),
// followed by expired/unknown sessions (also sorted by name).
type SessionsResult struct {
	Sessions     []awscli.ActiveSessionInfo // All configured sessions, active first
	CLIAvailable bool
	CLIVersion   string
	LastChecked  time.Time
	ConfigPath   string // AWS CLI config file that was read
	Error        error
}

// HasActiveSessions reports whether any session in the result is currently active.
func (r *SessionsResult) HasActiveSessions() bool {
	for _, s := range r.Sessions {
		if !s.IsExpired {
			return true
		}
	}
	return false
}

// NewSessionsViewModel creates and registers a sessions view model.
func NewSessionsViewModel() *SessionsViewModel {
	logging.Debug.Log("\t🔹 Creating sessions view model")

	vm := &SessionsViewModel{}
	core.App.RegisterState("sessions-view", vm)

	logging.Debug.Log("\t🔹 Sessions view model created")
	return vm
}

// LoadSessions fetches current SSO session status asynchronously.
//
// It reads ALL configured [sso-session …] entries from ~/.aws/config, then
// overlays live token status from the SSO cache so that sessions without a
// valid cache token still appear (defaulting to expired). Active sessions are
// sorted before expired ones; within each group sessions are sorted by name.
//
// Parameters:
//   - ctx: Context for cancellation
//   - forceRefresh: When true, bypasses the 30-second awscli cache
//   - reporter: Progress reporter
//   - onComplete: Callback invoked on the calling goroutine with the result
func (vm *SessionsViewModel) LoadSessions(
	ctx context.Context,
	forceRefresh bool,
	reporter task.Reporter,
	onComplete func(*SessionsResult),
) {
	vm.mu.Lock()
	vm.IsLoading = true
	vm.mu.Unlock()

	var result *SessionsResult

	asyncTask := &task.FunctionTask{
		Name: "load-sessions-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			runReporter.ReportStatus("Reading configured SSO sessions...")

			// Step 1: get ALL configured sessions from the config file
			configuredSessions, err := awscli.GetConfiguredSessions()
			if err != nil {
				result = &SessionsResult{Error: err}
				return nil, err
			}

			// Step 2: get live token status
			runReporter.ReportStatus("Checking SSO session status...")
			status, err := awscli.GetSessionStatus(forceRefresh)
			if err != nil {
				// Non-fatal — we can still show configured sessions as expired
				logging.Log.Warnf("Could not get session status: %v", err)
				status = awscli.SessionStatus{}
			}

			// Step 3: build a status lookup map
			statusMap := make(map[string]awscli.ActiveSessionInfo)
			for _, s := range status.ActiveSessions {
				statusMap[s.SessionName] = s
			}
			for _, s := range status.ExpiredSessions {
				statusMap[s.SessionName] = s
			}

			// Step 4: merge — every configured session gets an entry
			seen := make(map[string]bool)
			var active, expired []awscli.ActiveSessionInfo

			for _, cfg := range configuredSessions {
				if seen[cfg.Name] {
					continue
				}
				seen[cfg.Name] = true

				info := awscli.ActiveSessionInfo{
					SessionName: cfg.Name,
					StartURL:    cfg.StartURL,
					Region:      cfg.Region,
					IsExpired:   true, // default until proven otherwise
				}

				if live, exists := statusMap[cfg.Name]; exists {
					info = live
				}

				if info.IsExpired {
					expired = append(expired, info)
				} else {
					active = append(active, info)
				}
			}

			// Sort each group by name
			sort.Slice(active, func(i, j int) bool { return active[i].SessionName < active[j].SessionName })
			sort.Slice(expired, func(i, j int) bool { return expired[i].SessionName < expired[j].SessionName })

			result = &SessionsResult{
				Sessions:     append(active, expired...),
				CLIAvailable: status.CLIAvailable,
				CLIVersion:   status.CLIVersion,
				LastChecked:  status.LastChecked,
				ConfigPath:   filepath.Join(settings.GetAwsDir(), "config"),
			}
			if result.LastChecked.IsZero() {
				result.LastChecked = time.Now()
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		vm.mu.Lock()
		vm.IsLoading = false
		vm.mu.Unlock()

		if err != nil && result == nil {
			result = &SessionsResult{Error: err}
		}
		if onComplete != nil {
			onComplete(result)
		}
	})
}

// LoginSession runs `aws sso login --sso-session <name>` asynchronously.
//
// onComplete is called on the background goroutine with any error. The caller
// must use fyne.Do to update the UI from onComplete.
func (vm *SessionsViewModel) LoginSession(sessionName string, onComplete func(error)) {
	go func() {
		err := awscli.LoginSession(sessionName)
		if onComplete != nil {
			onComplete(err)
		}
	}()
}

// LogoutAllSessions runs `aws sso logout` asynchronously.
//
// onComplete is called on the background goroutine with any error.
func (vm *SessionsViewModel) LogoutAllSessions(onComplete func(error)) {
	go func() {
		err := awscli.LogoutAllSessions()
		if onComplete != nil {
			onComplete(err)
		}
	}()
}

// FormatLastChecked formats the last-checked timestamp for display in the header.
func (vm *SessionsViewModel) FormatLastChecked(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
