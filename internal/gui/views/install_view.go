// Package views provides the GUI views for the AWS Profile Manager application.
package views

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/components/installoptions"
	"aws-profile-manager/internal/gui/components/loadinfo"
	"aws-profile-manager/internal/gui/components/presetselect"
	"aws-profile-manager/internal/gui/components/schemafilters"
	"aws-profile-manager/internal/gui/components/schemalist"
	"aws-profile-manager/internal/gui/components/viewheader"
	guiLayouts "aws-profile-manager/internal/gui/layouts"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/profiles"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

const (
	installLeftPaneMinWidth  float32 = 260
	installRightPaneMinWidth float32 = 320
	installSplitOffset       float64 = 0.35 // left pane takes 35% of available width
)

// NewInstallView creates the Install tab content.
//
// The Install view allows users to install AWS profiles from centralized
// configuration (sync cache) with preset support, filters, and installation options.
//
// Component Structure:
//   - LEFT PANE: PresetSelector → SchemaFilters → InstallOptions → Buttons
//   - RIGHT PANE: AccountListView
func NewInstallView(window fyne.Window) fyne.CanvasObject {
	logging.Debug.Log("\t🔹 Creating install view")

	viewModel := viewmodels.NewInstallViewModel()
	initialSchema := viewModel.EmptyDisplaySchema()

	// Store the currently loaded schema for installation
	var currentLoadedSchema *schema.Schema

	// Store the currently filtered schema (what's shown in account list)
	// This is what gets installed - not the full schema with filters
	var currentFilteredSchema *schema.Schema

	// Create components
	filtersView := schemafilters.NewSchemaFiltersView(initialSchema)
	accountListView := schemalist.NewAccountListView(initialSchema)
	installOptionsView := installoptions.NewInstallOptions(window)

	// Preset selector starts with nil (will be populated when schema loads)
	presetSelector := presetselect.NewPresetSelector(nil, func(preset *schema.Preset) {
		logging.Debug.Log("Install view: preset selected",
			"preset", preset,
		)
		filtersView.ApplyPreset(preset)
	})

	// Header — info button reveals source path and load time on tap
	loadInfo := loadinfo.NewLoadInfo(window)

	reloadButton := widget.NewButton("Reload Cache", func() {
		if window == nil {
			return
		}
		startInstallSchemaLoadUI(window, viewModel, loadInfo, presetSelector, filtersView, accountListView, &currentLoadedSchema, &currentFilteredSchema)
	})
	reloadButton.Importance = widget.MediumImportance
	if window == nil {
		reloadButton.Disable()
	}

	// Select file button - allows user to manually select a schema file
	selectFileButton := widget.NewButton("Select File", func() {
		if window == nil {
			return
		}
		showFileSelectionDialog(window, viewModel, loadInfo, presetSelector, filtersView, accountListView, &currentLoadedSchema, &currentFilteredSchema)
	})
	selectFileButton.Importance = widget.MediumImportance
	if window == nil {
		selectFileButton.Disable()
	}

	header := viewheader.New(
		"# Install Profiles",
		"Install AWS profiles via the sync settings or select a file. The accounts shown will be installed.",
	).WithInfo(loadInfo.GetContent()).WithButtons(selectFileButton, reloadButton)

	// Wire up filter changes
	filtersView.OnFilterChange(func(filteredSchema *schema.Schema) {
		logging.Debug.Log("Install view: filter changed")
		if filteredSchema == nil {
			accountListView.SetSchema(viewModel.EmptyDisplaySchema())
			currentFilteredSchema = viewModel.EmptyDisplaySchema()
			return
		}
		// Store the filtered schema - this is what will be installed
		currentFilteredSchema = filteredSchema
		accountListView.SetSchema(filteredSchema)
	})

	// Install button
	installButton := widget.NewButton("Install Profiles", func() {
		if currentFilteredSchema == nil {
			dialog.ShowError(fmt.Errorf("no schema loaded - please load a configuration first"), window)
			return
		}
		// Pass both original (with presets) and filtered schemas
		executeInstallation(window, viewModel, currentLoadedSchema, currentFilteredSchema, installOptionsView)
	})
	installButton.Importance = widget.HighImportance
	if window == nil {
		installButton.Disable()
	}

	// Clear filters button
	clearFiltersButton := widget.NewButton("Clear All Filters", func() {
		filtersView.Reset()
		presetSelector.Reset()
	})

	// Create filter section description
	filterDescription := widget.NewLabel("Use the filters below to fine-tune which AWS profiles will be installed. The account list on the right shows the AWS profiles that will be generated based on your current filter selection, this is what will be installed.")
	filterDescription.Wrapping = fyne.TextWrapWord

	// Accordion sections — both open by default so first-time users see everything.
	// The accordion allows users to collapse sections they don't need.
	presetsItem := widget.NewAccordionItem("Quick Presets", presetSelector.GetContent())
	presetsItem.Open = true

	filtersContent := container.NewVBox(
		filterDescription,
		filtersView.GetContent(),
		clearFiltersButton,
	)
	filtersItem := widget.NewAccordionItem("Filters", filtersContent)
	filtersItem.Open = true

	accordion := widget.NewAccordion(presetsItem, filtersItem)
	accordion.MultiOpen = true

	// Left pane: scrollable accordion (Presets + Filters) on top, fixed
	// InstallOptions + Install button pinned at the bottom.
	//
	// Keeping the install options outside the scroll area prevents filter
	// checkboxes from visually overlapping with them when the window is small,
	// and ensures the Install button is always visible without scrolling.
	accordionScroll := container.NewVScroll(accordion)

	bottomSection := container.NewVBox(
		widget.NewSeparator(),
		installOptionsView.GetContent(),
		widget.NewSeparator(),
		installButton,
	)

	leftPaneScroll := container.NewBorder(nil, bottomSection, nil, nil, accordionScroll)

	// Split layout
	split := guiLayouts.NewHorizontalSplitWithLeftRightPaneMinWidths(
		leftPaneScroll,
		accountListView.GetContent(),
		installLeftPaneMinWidth,
		installRightPaneMinWidth,
		installSplitOffset,
	)

	// Load schema on initialization
	if window == nil {
		// Headless mode: load synchronously if cache exists
		if viewModel.CacheExists() {
			displaySchema, sourcePath, loadErr := viewModel.LoadDisplaySchema(context.Background(), task.NoOpReporter{})
			if loadErr != nil {
				logging.Debug.Log("Install view: failed to load schema in headless mode", "error", loadErr)
			} else {
				currentLoadedSchema = displaySchema
				filtersView.SetSchema(displaySchema)
				accountListView.SetSchema(displaySchema)
				presetSelector.SetPresets(viewModel.GetPresets(displaySchema))
				if currentFilteredSchema == nil {
					currentFilteredSchema = displaySchema
				}
				loadInfo.SetSource(sourcePath)
				loadInfo.SetLoadedAt(time.Now())
			}
		}
	} else {
		// GUI mode: check if cache exists
		if viewModel.CacheExists() {
			// Load from cache
			startInstallSchemaLoadUI(window, viewModel, loadInfo, presetSelector, filtersView, accountListView, &currentLoadedSchema, &currentFilteredSchema)
		} else {
			// Prompt user to select a file
			showFileSelectionDialog(window, viewModel, loadInfo, presetSelector, filtersView, accountListView, &currentLoadedSchema, &currentFilteredSchema)
		}
	}

	return container.NewBorder(header.GetContent(), nil, nil, nil, split)
}

// showFileSelectionDialog shows a file picker for selecting a schema JSON file.
func showFileSelectionDialog(
	window fyne.Window,
	viewModel *viewmodels.InstallViewModel,
	loadInfo *loadinfo.LoadInfo,
	presetSelector *presetselect.PresetSelector,
	filtersView *schemafilters.SchemaFiltersView,
	accountListView *schemalist.AccountListView,
	currentLoadedSchema **schema.Schema,
	currentFilteredSchema **schema.Schema,
) {
	logging.Debug.Log("Install view: Showing file selection dialog")

	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			logging.Debug.Log("Install view: file selection error", "error", err)
			dialog.ShowError(err, window)
			return
		}

		if reader == nil {
			logging.Debug.Log("Install view: file selection cancelled")
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		logging.Debug.Log("Install view: file selected", "path", filePath)

		// Load schema from selected file
		progressDialog := components.ShowProgressDialog(window, "Loading Schema", "Reading file...", nil)
		progressDialog.Show()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_ = cancel // Intentionally not deferred - would cancel before async goroutine completes

		go func() {
			displaySchema, sourcePath, loadErr := viewModel.LoadSchemaFromFile(ctx, filePath, task.NoOpReporter{})

			fyne.Do(func() {
				progressDialog.Hide()

				if loadErr != nil {
					logging.Debug.Log("Install view: schema load from file failed", "error", loadErr)
					dialog.ShowError(loadErr, window)
					return
				}

				logging.Debug.Log("Install view: schema loaded from file successfully")

				// Set the filters schema first — see startInstallSchemaLoadUI for explanation.
				filtersView.SetSchema(displaySchema)
				accountListView.SetSchema(displaySchema)

				presets := viewModel.GetPresets(displaySchema)
				presetSelector.SetPresets(presets)

				if *currentFilteredSchema == nil {
					*currentFilteredSchema = displaySchema
				}
				*currentLoadedSchema = displaySchema

				loadInfo.SetSource(sourcePath)
				loadInfo.SetLoadedAt(time.Now())
			})
		}()
	}, window)

	// Set filter to JSON files
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))

	// Set dialog to configured size
	guiSettings := settings.Get().GUI
	width, height := guiSettings.DialogWidth, guiSettings.DialogHeight
	fileDialog.Resize(fyne.NewSize(float32(width), float32(height)))

	fileDialog.Show()
}

// startInstallSchemaLoadUI initiates async schema loading from sync cache.
func startInstallSchemaLoadUI(
	window fyne.Window,
	viewModel *viewmodels.InstallViewModel,
	loadInfo *loadinfo.LoadInfo,
	presetSelector *presetselect.PresetSelector,
	filtersView *schemafilters.SchemaFiltersView,
	accountListView *schemalist.AccountListView,
	currentLoadedSchema **schema.Schema,
	currentFilteredSchema **schema.Schema,
) {
	logging.Debug.Log("Install view: Starting schema load from cache")

	progressDialog := components.ShowProgressDialog(window, "Loading Schema", "Reading sync cache...", nil)
	progressDialog.Show()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = cancel // Intentionally not deferred - would cancel before async callback fires

	viewModel.StartLoad(ctx, task.NoOpReporter{}, func(displaySchema *schema.Schema, sourcePath string, err error, loadedAt time.Time) {
		fyne.Do(func() {
			progressDialog.Hide()

			if err != nil {
				logging.Debug.Log("Install view: schema load failed", "error", err)

				// Show error and prompt for file selection
				dialog.ShowError(err, window)
				showFileSelectionDialog(window, viewModel, loadInfo, presetSelector, filtersView, accountListView, currentLoadedSchema, currentFilteredSchema)
				return
			}

			logging.Debug.Log("Install view: schema loaded successfully")

			// Set the filters schema first so that when presetSelector.SetPresets
			// fires its Reset callback (→ filtersView.Reset → emitCurrentFilteredSchema)
			// the filters already know about the new schema.
			filtersView.SetSchema(displaySchema)
			accountListView.SetSchema(displaySchema)

			// Now apply presets — this triggers emitCurrentFilteredSchema which
			// sets currentFilteredSchema via OnFilterChange.
			presets := viewModel.GetPresets(displaySchema)
			presetSelector.SetPresets(presets)

			// Belt-and-suspenders: ensure filteredSchema is set even if no preset
			// callback fired (e.g. schema has no presets).
			if *currentFilteredSchema == nil {
				*currentFilteredSchema = displaySchema
			}
			*currentLoadedSchema = displaySchema
			loadInfo.SetLoadedAt(loadedAt)
		})
	})
}

// executeInstallation executes the profile installation using the ViewModel.
//
// This is a thin presentation layer that:
//  1. Shows progress dialog
//  2. Calls ViewModel.Install() (which contains all business logic)
//  3. Displays results or errors
//
// The ViewModel handles:
//   - Profile installation using profiles package
//   - Saving cache file with full schema
//   - Updating sync settings
func executeInstallation(
	window fyne.Window,
	viewModel *viewmodels.InstallViewModel,
	originalSchema *schema.Schema,
	filteredSchema *schema.Schema,
	installOptionsView *installoptions.InstallOptions,
) {
	logging.Debug.Log("Install view: Starting installation")

	progressDialog := components.ShowProgressDialog(window, "Installing Profiles", "Preparing installation...", nil)
	progressDialog.Show()

	// Build install options via viewmodel — keeps profiles package out of the view layer.
	installOpts := viewModel.BuildInstallOptions(
		installOptionsView.IsCheatsheetOnly(),
		installOptionsView.ShouldGenerateCheatsheet(),
		installOptionsView.IsCheatsheetOnly(),
		installOptionsView.GetOutputFile(),
		installOptionsView.GetCheatsheetFile(),
	)

	// Create context with timeout - NOTE: do NOT defer cancel() here as StartInstall is async.
	// The context will be canceled naturally when the timeout expires.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	_ = cancel // Intentionally not deferred - would cancel before async goroutine completes

	// Call ViewModel using StartInstall (handles goroutine and context)
	viewModel.StartInstall(
		ctx,
		originalSchema,
		filteredSchema,
		installOpts,
		task.NoOpReporter{},
		func(result *profiles.InstallResult, err error) {
			fyne.Do(func() {
				progressDialog.Hide()

				if err != nil {
					logging.Debug.Log("Install view: installation failed", "error", err)
					dialog.ShowError(fmt.Errorf("installation failed: %w", err), window)
					return
				}

				logging.Debug.Log("Install view: installation complete",
					"profiles", result.TotalProfiles,
				)

				// Status header and detail widget are both produced by the viewmodel.
				statusHeader := viewModel.FormatInstallStatusHeader(result, installOpts.DryRun)
				resultWidget := viewModel.FormatInstallResult(result)
				// Build success items
				successItems := []fyne.CanvasObject{
					widget.NewRichTextFromMarkdown(statusHeader),
				}
				if resultWidget != nil {
					successItems = append(successItems, widget.NewSeparator(), resultWidget)
				}

				if result.CheatSheetPath != "" {
					savedName := filepath.Base(result.CheatSheetPath)
					cheatSheetURL := &url.URL{Scheme: "file", Path: result.CheatSheetPath}
					cheatSheetSection := container.NewHBox(
						widget.NewRichTextFromMarkdown("### Cheat Sheet:"),
						widget.NewHyperlink(savedName, cheatSheetURL),
					)
					successItems = append(successItems,
						widget.NewSeparator(),
						cheatSheetSection,
					)
				}

				successContent := container.NewVBox(successItems...)

				// Show success dialog
				var successDialog *dialog.CustomDialog
				successDialog = components.ShowCustomDialog(components.DialogOptions{
					Title:   "Installation Complete",
					Content: successContent,
					Buttons: []components.DialogButton{
						{
							Label: "OK",
							OnTapped: func() {
								successDialog.Hide()
							},
							Importance: widget.HighImportance,
						},
					},
					Window:      window,
					Scrollable:  true,
					UseSettings: true,
				})
				successDialog.Show()

				logging.Log.Success("Installation completed",
					"profiles_installed", result.TotalProfiles,
					"cheat_sheet", result.CheatSheetPath != "",
				)
			})
		},
	)
}
