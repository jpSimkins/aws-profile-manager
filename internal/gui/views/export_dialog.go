package views

import (
	"context"
	"net/url"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// ShowExportDialog displays the Export Profiles options dialog.
//
// This is the entry point for the export workflow. It collects export options
// from the user, then proceeds to showExportSaveDialog for file selection.
func ShowExportDialog(window fyne.Window, footer *components.Footer) {
	logging.Debug.Log("Creating export dialog")

	// Update footer status
	if footer != nil {
		footer.SetStatus("Opening export dialog...")
	}

	includeManagedCheck := widget.NewCheck("Include managed profiles", nil)
	includeManagedCheck.SetChecked(true)
	includeAboveCheck := widget.NewCheck("Include personal profiles above", nil)
	includeAboveCheck.SetChecked(true)
	includeBelowCheck := widget.NewCheck("Include personal profiles below", nil)
	includeBelowCheck.SetChecked(true)
	excludeSettingsCheck := widget.NewCheck("Exclude application settings from backup", nil)
	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Optional: Backup description (e.g., 'Pre-OS-wipe backup')")

	form := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Export AWS Profiles\n\nExport your AWS profiles to a Schema JSON file."),
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Export Sections:"),
		includeManagedCheck,
		includeAboveCheck,
		includeBelowCheck,
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Settings:"),
		excludeSettingsCheck,
		widget.NewLabel("Default: Application settings are included in backup"),
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("### Description (Optional):"),
		descriptionEntry,
	)

	var exportDialog *dialog.CustomDialog

	buttons := components.CreateStandardButtons("Cancel", "Export",
		func() {
			exportDialog.Hide()
			if footer != nil {
				footer.SetStatus("Export cancelled")
			}
		},
		func() {
			exportDialog.Hide()

			opts := viewmodels.ExportOptions{
				IncludeManaged:  includeManagedCheck.Checked,
				IncludeAbove:    includeAboveCheck.Checked,
				IncludeBelow:    includeBelowCheck.Checked,
				Description:     descriptionEntry.Text,
				ExcludeSettings: excludeSettingsCheck.Checked,
			}

			showExportSaveDialog(window, footer, opts)
		})

	exportDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:       "Export Profiles",
		Content:     form,
		Buttons:     buttons,
		Window:      window,
		Scrollable:  true,
		UseSettings: true,
	})

	// Update footer status
	if footer != nil {
		footer.SetStatus("Pending user input")
	}

	exportDialog.Show()
}

// showExportSaveDialog opens the file save picker after export options are confirmed.
//
// On file selection it calls executeExport with the final output path.
func showExportSaveDialog(window fyne.Window, footer *components.Footer, opts viewmodels.ExportOptions) {
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			_ = logging.Log.Error("Failed to open save dialog", "error", err)
			dialog.ShowError(err, window)
			if footer != nil {
				footer.SetStatus("Export failed")
			}
			return
		}

		if writer == nil {
			logging.Debug.Log("Export cancelled by user")
			if footer != nil {
				footer.SetStatus("Export cancelled")
			}
			return
		}

		outputPath := writer.URI().Path()
		_ = writer.Close()

		logging.Debug.Log("Export file selected", "path", outputPath)

		opts.OutputPath = outputPath
		executeExport(window, footer, opts)
	}, window)

	saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
	saveDialog.SetFileName("aws-profiles-backup.json")

	guiSettings := settings.Get().GUI
	saveDialog.Resize(fyne.NewSize(float32(guiSettings.DialogWidth), float32(guiSettings.DialogHeight)))

	saveDialog.Show()
}

// executeExport runs the export operation and shows the result dialog.
//
// Delegates async execution to ViewModel.StartExport (task package), then
// presents success or error on the main thread via fyne.Do.
func executeExport(window fyne.Window, footer *components.Footer, opts viewmodels.ExportOptions) {
	// Update footer status
	if footer != nil {
		footer.SetStatus("Exporting profiles...")
	}

	ctx, cancel := context.WithCancel(context.Background())

	progressDialog := components.ShowProgressDialog(window, "Export in Progress", "Exporting profiles...", cancel)
	progressDialog.Show()

	viewModel := viewmodels.NewExportViewModel()

	viewModel.StartExport(ctx, opts, &GuiProgressReporter{progressDialog: progressDialog}, func(result *viewmodels.ExportResult) {
		fyne.Do(func() {
			progressDialog.Hide()

			if !result.Success {
				if result.Error == context.Canceled {
					logging.Log.Warn("Export cancelled by user")
					if footer != nil {
						footer.SetStatus("Export cancelled")
					}
					dialog.ShowInformation("Cancelled", "Export operation was cancelled.", window)
					return
				}

				_ = logging.Log.Error("Export failed", "error", result.Error)
				if footer != nil {
					footer.SetStatus("Export failed")
				}
				dialog.ShowError(result.Error, window)
				return
			}

			if footer != nil {
				footer.SetStatus("Export completed successfully")
			}

			logging.Log.Success("Export completed",
				"total", result.TotalProfiles,
				"managed", result.ManagedProfiles,
				"above", result.UnmanagedAbove,
				"below", result.UnmanagedBelow,
			)

			showExportSuccessDialog(window, footer, viewModel, result, opts)
		})
	})
}

// showExportSuccessDialog presents the export result summary to the user.
func showExportSuccessDialog(window fyne.Window, footer *components.Footer, viewModel *viewmodels.ExportViewModel, result *viewmodels.ExportResult, opts viewmodels.ExportOptions) {
	savedName := filepath.Base(result.OutputPath)
	if savedName == "" || savedName == "." {
		savedName = result.OutputPath
	}

	folderURL := &url.URL{Scheme: "file", Path: result.OutputPath}

	successContent := container.NewVBox(
		widget.NewRichTextFromMarkdown("## ✅ Export Successful!"),
		widget.NewSeparator(),
		viewModel.FormatResult(result, opts),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewRichTextFromMarkdown("### Saved File:"),
			widget.NewHyperlink(savedName, folderURL),
		),
	)

	var successDialog *dialog.CustomDialog
	successDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:   "Export Complete",
		Content: successContent,
		Buttons: []components.DialogButton{
			{
				Label: "OK",
				OnTapped: func() {
					successDialog.Hide()
					if footer != nil {
						footer.SetStatus("Export complete")
					}
				},
				Importance: widget.HighImportance,
			},
		},
		Window:      window,
		Scrollable:  true,
		UseSettings: true,
	})
	successDialog.Show()
}
