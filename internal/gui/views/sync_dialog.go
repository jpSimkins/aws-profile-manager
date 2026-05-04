package views

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
)

// ShowSyncDialog is the entry point for the sync workflow.
//
// Checks whether sync is enabled before proceeding. If enabled, calls
// executeSyncNow to perform the operation.
func ShowSyncDialog(window fyne.Window, footer *components.Footer) {
	logging.Log.Info("Sync Now triggered")

	viewModel := viewmodels.NewSyncViewModel()

	if !viewModel.IsEnabled() {
		dialog.ShowInformation("Sync Disabled",
			"Configuration sync is not enabled.\n\nPlease enable sync in Settings to use this feature.",
			window)
		if footer != nil {
			footer.SetStatus("Sync is disabled")
		}
		return
	}

	if footer != nil {
		footer.SetStatus("Syncing configuration...")
	}

	executeSyncNow(window, footer, viewModel)
}

// executeSyncNow runs the sync operation and shows the result dialog.
//
// Delegates async execution to ViewModel.StartSync (task package), then
// presents success or error on the main thread via fyne.Do.
func executeSyncNow(window fyne.Window, footer *components.Footer, viewModel *viewmodels.SyncViewModel) {
	ctx, cancel := context.WithCancel(context.Background())

	progressDialog := components.ShowProgressDialog(window, "Sync in Progress", "Syncing configuration...", cancel)
	progressDialog.Show()

	viewModel.StartSync(ctx, &GuiProgressReporter{progressDialog: progressDialog}, func(result *viewmodels.SyncResult) {
		fyne.Do(func() {
			progressDialog.Hide()

			if !result.Success {
				if result.Error == context.Canceled {
					logging.Log.Warn("Sync cancelled by user")
					if footer != nil {
						footer.SetStatus("Sync cancelled")
					}
					dialog.ShowInformation("Cancelled", "Sync operation was cancelled.", window)
					return
				}

				_ = logging.Log.Error("Sync failed", "error", result.Error)
				if footer != nil {
					footer.SetStatus("Sync failed")
				}
				dialog.ShowError(result.Error, window)
				return
			}

			if footer != nil {
				if result.CacheHit {
					footer.SetStatus("Sync complete (from cache)")
				} else {
					footer.SetStatus("Sync complete")
				}
			}

			logging.Log.Success("Sync completed",
				"source", result.Source,
				"size", result.BytesTransferred,
				"cache_hit", result.CacheHit,
			)

			showSyncSuccessDialog(window, footer, viewModel, result)
		})
	})
}

// showSyncSuccessDialog presents the sync result summary to the user.
func showSyncSuccessDialog(window fyne.Window, footer *components.Footer, viewModel *viewmodels.SyncViewModel, result *viewmodels.SyncResult) {
	successMessage := widget.NewRichTextFromMarkdown(viewModel.FormatResult(result))

	var successDialog *dialog.CustomDialog
	successDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:   "Sync Complete",
		Content: successMessage,
		Buttons: []components.DialogButton{
			{
				Label: "OK",
				OnTapped: func() {
					successDialog.Hide()
					if footer != nil {
						footer.SetStatus("Sync complete")
					}
				},
				Importance: widget.MediumImportance,
			},
		},
		Window:      window,
		UseSettings: false,
	})
	successDialog.Show()
}
