package viewmodels

import (
	"context"
	"fmt"
	"sync"

	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
	syncpkg "aws-profile-manager/internal/sync"
	"aws-profile-manager/internal/task"
)

// SyncViewModel manages the state and business logic for sync operations.
//
// It follows the same MVVM pattern as ExportViewModel and InstallViewModel:
//   - Builds SyncConfig from settings (dependency injection)
//   - Delegates async execution to the task package via StartSync
//   - Keeps all concurrency out of the view layer
type SyncViewModel struct {
	IsSyncing bool // Is a sync operation currently in progress?
	mu        sync.RWMutex
}

// SyncResult contains the outcome of a sync operation for GUI display.
type SyncResult struct {
	Success          bool
	Source           string
	CacheHit         bool
	BytesTransferred int64
	Error            error
}

// NewSyncViewModel creates a new sync view model.
func NewSyncViewModel() *SyncViewModel {
	logging.Debug.Log("\t🔹 Creating sync view model")
	return &SyncViewModel{}
}

// IsEnabled reports whether sync is configured and enabled in settings.
func (vm *SyncViewModel) IsEnabled() bool {
	return settings.Get().Sync.Enabled
}

// StartSync executes the sync operation asynchronously using the task package.
//
// Builds SyncConfig from the current settings, then runs sync.Sync via
// task.RunAsync. The onComplete callback is invoked on the task goroutine —
// callers must use fyne.Do for any UI updates.
//
// Parameters:
//   - ctx: Context for cancellation
//   - reporter: Progress reporter (typically GuiProgressReporter)
//   - onComplete: Callback invoked with the sync result
func (vm *SyncViewModel) StartSync(
	ctx context.Context,
	reporter task.Reporter,
	onComplete func(*SyncResult),
) {
	logging.Debug.Log("Sync view model: StartSync triggered")

	vm.mu.Lock()
	vm.IsSyncing = true
	vm.mu.Unlock()

	var syncResult *SyncResult

	asyncTask := &task.FunctionTask{
		Name: "sync-config-async",
		Fn: func(runCtx context.Context, runReporter task.Reporter) ([]byte, error) {
			syncSettings := settings.Get().Sync
			cfg := syncpkg.ConfigFromSettings(&syncSettings)
			opts := syncpkg.Options{ForceRefresh: true}

			runReporter.ReportStatus("Fetching remote configuration...")
			result, err := syncpkg.Sync(runCtx, cfg, opts, runReporter)
			if err != nil {
				syncResult = &SyncResult{Success: false, Error: err}
				return nil, err
			}

			syncResult = &SyncResult{
				Success:          true,
				Source:           result.Source,
				CacheHit:         result.CacheHit,
				BytesTransferred: result.BytesTransferred,
			}
			return []byte("ok"), nil
		},
	}

	task.RunAsync(ctx, asyncTask, reporter, func(_ *task.Result, err error) {
		vm.mu.Lock()
		vm.IsSyncing = false
		vm.mu.Unlock()

		logging.Debug.Log("Sync view model: StartSync complete", "error", err)
		if syncResult == nil {
			syncResult = &SyncResult{Success: false, Error: err}
		}
		if onComplete != nil {
			onComplete(syncResult)
		}
	})
}

// FormatResult returns a markdown status text for the sync result dialog.
func (vm *SyncViewModel) FormatResult(result *SyncResult) string {
	text := "## ✅ Sync Successful!\n\n"
	if result.CacheHit {
		text += "**Note:** Configuration was retrieved from cache.\n\n"
	}
	text += "**Source:**\n" + result.Source + "\n\n"
	if result.BytesTransferred > 0 {
		text += "**Size:**\n" + formatSyncBytes(int(result.BytesTransferred)) + "\n"
	}
	text += "\n**Next step:** Run Install to apply profiles"
	return text
}

// formatSyncBytes formats a byte count as a human-readable string.
func formatSyncBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
