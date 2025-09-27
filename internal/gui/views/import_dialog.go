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

// GuiProgressReporter updates progress dialog with task status.
type GuiProgressReporter struct {
	progressDialog *components.ProgressDialog
}

// ReportStatus updates the progress dialog with current status.
func (r *GuiProgressReporter) ReportStatus(status string) {
	if r.progressDialog != nil {
		fyne.Do(func() {
			r.progressDialog.UpdateDetails(status)
		})
	}
}

// ReportProgress updates progress (not used in simple progress dialog).
func (r *GuiProgressReporter) ReportProgress(current, total int64) {}

// ReportError logs non-fatal warnings during operation.
func (r *GuiProgressReporter) ReportError(err error) {
	logging.Log.Warn("Operation warning", "error", err)
}

// ShowImportDialog is the entry point for the import workflow.
//
// Opens the file picker, then proceeds to showImportPreview once a file is selected.
func ShowImportDialog(window fyne.Window, footer *components.Footer) {
	logging.Debug.Log("Creating import dialog")

	if footer != nil {
		footer.SetStatus("Opening import dialog...")
	}

	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			_ = logging.Log.Error("Failed to open file dialog", "error", err)
			dialog.ShowError(err, window)
			if footer != nil {
				footer.SetStatus("Import failed")
			}
			return
		}

		if reader == nil {
			logging.Debug.Log("Import cancelled by user")
			if footer != nil {
				footer.SetStatus("Import cancelled")
			}
			return
		}

		backupPath := reader.URI().Path()
		_ = reader.Close()

		logging.Debug.Log("Backup file selected", "path", backupPath)

		showImportPreview(window, footer, backupPath)
	}, window)

	openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))

	guiSettings := settings.Get().GUI
	openDialog.Resize(fyne.NewSize(float32(guiSettings.DialogWidth), float32(guiSettings.DialogHeight)))

	// Update footer status
	if footer != nil {
		footer.SetStatus("Pending user input")
	}

	openDialog.Show()
}

// showImportPreview loads a preview of the backup file and then shows the options dialog.
//
// Uses ViewModel.StartPreview (task package) for async loading, then hands off
// to showImportOptionsDialog on the main thread.
func showImportPreview(window fyne.Window, footer *components.Footer, backupPath string) {
	if footer != nil {
		footer.SetStatus("Processing backup file...")
	}

	ctx, cancel := context.WithCancel(context.Background())

	progressDialog := components.ShowProgressDialog(window, "Processing Backup", "Analyzing backup file...", cancel)
	progressDialog.Show()

	viewModel := viewmodels.NewImportViewModel()

	previewOpts := viewmodels.ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
		IgnoreSettings: false,
	}

	viewModel.StartPreview(ctx, previewOpts, &GuiProgressReporter{progressDialog: progressDialog}, func(preview *viewmodels.ImportPreview) {
		fyne.Do(func() {
			progressDialog.Hide()

			if !preview.Success {
				_ = logging.Log.Error("Failed to get import preview", "error", preview.Error)
				dialog.ShowError(preview.Error, window)
				if footer != nil {
					footer.SetStatus("Import failed")
				}
				return
			}

			if footer != nil {
				footer.SetStatus("Pending user input")
			}

			showImportOptionsDialog(window, footer, viewModel, backupPath, preview)
		})
	})
}

// showImportOptionsDialog presents import options and the file preview to the user.
//
// On confirm it calls executeImport with the chosen options.
func showImportOptionsDialog(window fyne.Window, footer *components.Footer, viewModel *viewmodels.ImportViewModel, backupPath string, preview *viewmodels.ImportPreview) {
	includeManagedCheck := widget.NewCheck("Include managed profiles", nil)
	includeManagedCheck.SetChecked(true)
	includeAboveCheck := widget.NewCheck("Include personal profiles above", nil)
	includeAboveCheck.SetChecked(true)
	includeBelowCheck := widget.NewCheck("Include personal profiles below", nil)
	includeBelowCheck.SetChecked(true)
	ignoreSettingsCheck := widget.NewCheck("Don't restore application settings", nil)
	backupCurrentSettingsCheck := widget.NewCheck("Backup current settings before restoring", nil)
	backupCurrentSettingsCheck.SetChecked(true)
	generateCheatSheetCheck := widget.NewCheck("Generate cheat sheet after import", nil)

	previewOpts := viewmodels.ImportOptions{
		BackupPath:     backupPath,
		IncludeManaged: true,
		IncludeAbove:   true,
		IncludeBelow:   true,
	}

	optionsList := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown("**Import Options:**"),
		includeManagedCheck,
		includeAboveCheck,
		includeBelowCheck,
	}
	if preview.HasSettings {
		optionsList = append(optionsList, ignoreSettingsCheck, backupCurrentSettingsCheck)
	}
	optionsList = append(optionsList,
		generateCheatSheetCheck,
	)

	fileInfo := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown("## Import AWS Profiles\n\n_Restore profiles from a backup file._"),
		widget.NewRichTextFromMarkdown("**File:** " + filepath.Base(backupPath)),
	}
	if preview.Description != "" {
		fileInfo = append(fileInfo, widget.NewRichTextFromMarkdown("_"+preview.Description+"_"))
	}
	fileInfo = append(fileInfo, widget.NewSeparator())

	form := container.NewVBox(
		append(fileInfo,
			widget.NewRichTextFromMarkdown("## Preview:"),
			viewModel.FormatPreview(preview, previewOpts),
			widget.NewSeparator(),
			container.NewVBox(optionsList...),
		)...,
	)

	var optionsDialog *dialog.CustomDialog

	buttons := components.CreateStandardButtons("Cancel", "Import",
		func() {
			optionsDialog.Hide()
			if footer != nil {
				footer.SetStatus("Import cancelled")
			}
		},
		func() {
			optionsDialog.Hide()

			importOpts := viewmodels.ImportOptions{
				BackupPath:            backupPath,
				IncludeManaged:        includeManagedCheck.Checked,
				IncludeAbove:          includeAboveCheck.Checked,
				IncludeBelow:          includeBelowCheck.Checked,
				IgnoreSettings:        ignoreSettingsCheck.Checked,
				BackupCurrentSettings: backupCurrentSettingsCheck.Checked,
				GenerateCheatSheet:    generateCheatSheetCheck.Checked,
			}

			executeImport(window, footer, viewModel, importOpts)
		})

	optionsDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:       "Import Options",
		Content:     form,
		Buttons:     buttons,
		Window:      window,
		Scrollable:  true,
		UseSettings: true,
	})

	optionsDialog.Show()
}

// executeImport runs the import operation and shows the result dialog.
//
// Delegates async execution to ViewModel.StartImport (task package), then
// presents success or error on the main thread via fyne.Do.
func executeImport(window fyne.Window, footer *components.Footer, viewModel *viewmodels.ImportViewModel, opts viewmodels.ImportOptions) {
	if footer != nil {
		footer.SetStatus("Importing profiles...")
	}

	ctx, cancel := context.WithCancel(context.Background())

	progressDialog := components.ShowProgressDialog(window, "Import in Progress", "Importing profiles...", cancel)
	progressDialog.Show()

	viewModel.StartImport(ctx, opts, &GuiProgressReporter{progressDialog: progressDialog}, func(result *viewmodels.ImportResult) {
		fyne.Do(func() {
			progressDialog.Hide()

			if !result.Success {
				if result.Error == context.Canceled {
					logging.Log.Warn("Import cancelled by user")
					if footer != nil {
						footer.SetStatus("Import cancelled")
					}
					dialog.ShowInformation("Cancelled", "Import operation was cancelled.", window)
					return
				}

				_ = logging.Log.Error("Import failed", "error", result.Error)
				if footer != nil {
					footer.SetStatus("Import failed")
				}
				dialog.ShowError(result.Error, window)
				return
			}

			if footer != nil {
				footer.SetStatus("Import completed successfully")
			}

			totalWritten := result.ManagedStats.ProfilesWritten +
				result.UnmanagedAboveStats.ProfilesWritten +
				result.UnmanagedBelowStats.ProfilesWritten -
				result.ManagedDuplicates.TotalDuplicates -
				result.UnmanagedAboveDuplicates.TotalDuplicates -
				result.UnmanagedBelowDuplicates.TotalDuplicates

			totalDuplicates := result.ManagedDuplicates.TotalDuplicates +
				result.UnmanagedAboveDuplicates.TotalDuplicates +
				result.UnmanagedBelowDuplicates.TotalDuplicates

			logging.Log.Success("Import completed",
				"total_written", totalWritten,
				"managed_profiles", result.ManagedStats.ProfilesWritten,
				"above_profiles", result.UnmanagedAboveStats.ProfilesWritten,
				"below_profiles", result.UnmanagedBelowStats.ProfilesWritten,
				"duplicates_skipped", totalDuplicates,
			)

			showImportSuccessDialog(window, footer, viewModel, result, opts, totalWritten)
		})
	})
}

// showImportSuccessDialog presents the import result summary to the user.
func showImportSuccessDialog(window fyne.Window, footer *components.Footer, viewModel *viewmodels.ImportViewModel, result *viewmodels.ImportResult, opts viewmodels.ImportOptions, totalWritten int) {
	var statusText string
	if totalWritten == 0 && !result.SettingsRestored {
		statusText = "## 🛈 No Changes Needed\n\n**All profiles already exist.**"
	} else {
		statusText = "## ✅ Import Successful!"
	}

	successItems := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown(statusText),
		widget.NewSeparator(),
		viewModel.FormatResult(result, opts),
	}

	if result.CheatSheetGenerated && result.CheatSheetPath != "" {
		savedName := filepath.Base(result.CheatSheetPath)
		if savedName == "" || savedName == "." {
			savedName = result.CheatSheetPath
		}
		cheatSheetURL := &url.URL{Scheme: "file", Path: result.CheatSheetPath}
		successItems = append(successItems,
			widget.NewSeparator(),
			container.NewHBox(
				widget.NewRichTextFromMarkdown("### Cheat Sheet:"),
				widget.NewHyperlink(savedName, cheatSheetURL),
			),
		)
	}

	var successDialog *dialog.CustomDialog
	successDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:   "Import Complete",
		Content: container.NewVBox(successItems...),
		Buttons: []components.DialogButton{
			{
				Label: "OK",
				OnTapped: func() {
					successDialog.Hide()
					if footer != nil {
						footer.SetStatus("Import complete")
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
