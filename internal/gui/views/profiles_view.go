// Package views provides the GUI views for the AWS Profile Manager application.
package views

import (
	"time"

	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/gui/components/loadinfo"
	"aws-profile-manager/internal/gui/components/schemafilters"
	"aws-profile-manager/internal/gui/components/schemalist"
	"aws-profile-manager/internal/gui/components/sessionlist"
	"aws-profile-manager/internal/gui/components/viewheader"
	guiLayouts "aws-profile-manager/internal/gui/layouts"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
	"aws-profile-manager/internal/task"
)

const (
	profilesLeftPaneMinWidth  float32 = 260
	profilesRightPaneMinWidth float32 = 320
	profilesSplitOffset       float64 = 0.35 // left pane takes 35% of available width
)

// NewProfilesView creates the Profiles tab content.
//
// The Profiles view displays the list of AWS CLI profiles extracted from
// ~/.aws/config, with filtering and sorting capabilities.
func NewProfilesView(window fyne.Window) fyne.CanvasObject {
	logging.Debug.Log("\t🔹 Creating profiles view")

	viewModel := viewmodels.NewProfilesViewModel()
	configPath := viewModel.ConfigPath()
	initialSchema := viewModel.EmptyDisplaySchema()
	filtersView := schemafilters.NewSchemaFiltersView(initialSchema)
	accountListView := schemalist.NewAccountListView(initialSchema).WithCliButtons(true)

	// Header — info button reveals source path and load time on tap
	loadInfo := loadinfo.NewLoadInfo(window)

	refreshButton := widget.NewButton("Refresh", func() {
		if window == nil {
			return
		}
		startProfilesSchemaLoadUI(window, viewModel, loadInfo, filtersView, accountListView)
	})
	refreshButton.Importance = widget.MediumImportance
	if window == nil {
		refreshButton.Disable()
	}

	header := viewheader.New(
		"# Profiles",
		"Browse your currently configured AWS CLI profiles. Profiles are loaded from your AWS config file.",
	).WithInfo(loadInfo.GetContent()).WithButtons(refreshButton)

	filtersView.OnFilterChange(func(filteredSchema *schema.Schema) {
		logging.Debug.Log("Profiles view: filter changed")
		if filteredSchema == nil {
			accountListView.SetSchema(viewModel.EmptyDisplaySchema())
			return
		}
		accountListView.SetSchema(filteredSchema)
	})

	// Create filter section description
	filterDescription := widget.NewLabel("Use the filters below to fine-tune which AWS profiles are shown. The account list on the right displays profiles currently configured in your AWS CLI.")
	filterDescription.Wrapping = fyne.TextWrapWord

	filtersItem := widget.NewAccordionItem("Filters", container.NewVBox(
		filterDescription,
		filtersView.GetContent(),
	))
	filtersItem.Open = false
	filtersAccordion := widget.NewAccordion(filtersItem)
	filtersAccordion.MultiOpen = true

	// Left pane is a VBox with an optional sessions section above the filters accordion.
	// Sessions are rendered at full natural height (no accordion wrapper) so they
	// are never clipped. The outer VScroll on the left pane handles overflow.
	leftPaneContent := container.NewVBox(filtersAccordion)

	if settings.Get().AwsCLI.ShowSsoSessions {
		sessionsComp := sessionlist.New(window)
		sessionsHeader := widget.NewRichTextFromMarkdown("## SSO Sessions")
		sessionsSection := container.NewVBox(sessionsHeader, sessionsComp.Content(), widget.NewSeparator())
		leftPaneContent = container.NewVBox(sessionsSection, filtersAccordion)
	}

	leftPaneScroll := container.NewVScroll(leftPaneContent)

	split := guiLayouts.NewHorizontalSplitWithLeftRightPaneMinWidths(
		leftPaneScroll,
		accountListView.GetContent(),
		profilesLeftPaneMinWidth,
		profilesRightPaneMinWidth,
		profilesSplitOffset,
	)

	if window == nil {
		displaySchema, _, loadErr := viewModel.LoadDisplaySchema(context.Background(), task.NoOpReporter{})
		if loadErr != nil {
			logging.Log.Warn("Profiles view: failed to load schema in headless mode", "error", loadErr)
		} else {
			loadInfo.SetSource(configPath)
			loadInfo.SetLoadedAt(time.Now())
		}

		if displaySchema != nil {
			filtersView.SetSchema(displaySchema)
			accountListView.SetSchema(displaySchema)
		}
	} else {
		startProfilesSchemaLoadUI(window, viewModel, loadInfo, filtersView, accountListView)
	}

	return container.NewBorder(
		header.GetContent(),
		nil,
		nil,
		nil,
		split,
	)
}

func startProfilesSchemaLoadUI(
	window fyne.Window,
	viewModel *viewmodels.ProfilesViewModel,
	loadInfo *loadinfo.LoadInfo,
	filtersView *schemafilters.SchemaFiltersView,
	accountListView *schemalist.AccountListView,
) {
	logging.Debug.Log("Profiles view: starting schema load UI")

	progressDialog := components.ShowProgressDialog(window, "Loading Profiles", "Reading AWS config...", nil)
	progressDialog.UpdateDetails("Building schema from current environment")
	progressDialog.Show()

	viewModel.StartLoad(context.Background(), task.NoOpReporter{}, func(displaySchema *schema.Schema, configPath string, loadErr error, loadedAt time.Time) {
		fyne.Do(func() {
			progressDialog.Hide()

			logging.Debug.Log("Profiles view: schema load callback received",
				"configPath", configPath,
				"error", loadErr,
			)

			if loadErr != nil {
				logging.Log.Warn("Failed to load schema from AWS config",
					"path", configPath,
					"error", loadErr,
				)
				filtersView.SetSchema(viewModel.EmptyDisplaySchema())
				accountListView.SetSchema(viewModel.EmptyDisplaySchema())
				return
			}

			loadInfo.SetSource(configPath)
			loadInfo.SetLoadedAt(loadedAt)
			filtersView.SetSchema(displaySchema)
			accountListView.SetSchema(displaySchema)
		})
	})
}
