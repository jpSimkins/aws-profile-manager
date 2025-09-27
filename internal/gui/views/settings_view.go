package views

import (
	"image/color"
	"os"
	"strconv"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components"
	guiLayouts "aws-profile-manager/internal/gui/layouts"
	"aws-profile-manager/internal/gui/viewmodels"
	"aws-profile-manager/internal/logging"
	"aws-profile-manager/internal/settings"
)

// dependentField tracks a field and its container for visibility management
type dependentField struct {
	container        *fyne.Container
	fieldName        string
	dependency       *settings.FieldDependency
	updateVisibility func()
}

// NewSettingsView builds the Settings as a full-screen canvas object.
//
// The layout is a horizontal split: a narrow navigation sidebar on the left
// lists every settings section; clicking a section replaces the right pane
// with that section's fields. onClose is invoked when the user clicks Cancel
// or after a successful save, allowing the caller to pop the overlay.
func NewSettingsView(window fyne.Window, footer *components.Footer, onClose func(), refreshCallback func()) fyne.CanvasObject {
	logging.Debug.Log("Building settings view")

	vm := viewmodels.NewSettingsViewModel()
	vm.InitializeValues()
	currentValues := vm.GetValues() // direct reference — mutations propagate to VM
	dependentFields := &[]*dependentField{}

	// Build one scrollable content pane per section.
	sections := vm.GetSectionOrder()
	allSchemas := settings.Get().GetAllSchemas()

	sectionScrollers := make([]*container.Scroll, len(sections))
	for i, key := range sections {
		schema, exists := allSchemas[key]
		if !exists {
			// Section registered in order but not yet in schemas — render empty pane.
			sectionScrollers[i] = container.NewVScroll(container.NewVBox())
			continue
		}
		displayName := vm.GetSectionDisplayName(key)
		content := buildSectionContent(displayName, schema, vm, currentValues, dependentFields, window)
		sectionScrollers[i] = container.NewVScroll(
			guiLayouts.NewMaxWidthCentered(container.NewPadded(content), 700),
		)
	}

	// Stack all panes; only the selected one is visible at any time.
	stackObjects := make([]fyne.CanvasObject, len(sectionScrollers))
	for i, s := range sectionScrollers {
		stackObjects[i] = s
	}
	rightContent := container.NewStack(stackObjects...)

	// Start with the first section visible, all others hidden.
	for i, s := range sectionScrollers {
		if i == 0 {
			s.Show()
		} else {
			s.Hide()
		}
	}

	// Navigation list — clicking a row swaps the visible section pane.
	navList := widget.NewList(
		func() int { return len(sections) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(vm.GetSectionDisplayName(sections[id]))
		},
	)
	navList.OnSelected = func(id widget.ListItemID) {
		// Hide every pane then show only the chosen one.
		for i, s := range sectionScrollers {
			if i == int(id) {
				s.Show()
			} else {
				s.Hide()
			}
		}
		rightContent.Refresh()
	}
	navList.Select(0)

	// Wire initial visibility for dependency-controlled fields.
	setupVisibilityManagement(rightContent, dependentFields)

	// Split: narrow nav on the left, scrollable section content on the right.
	split := guiLayouts.NewHorizontalSplitWithLeftRightPaneMinWidths(
		navList,
		rightContent,
		150,
		400,
		0.20,
	)

	cancelBtn := widget.NewButton("Cancel", func() {
		vm.Cleanup()
		if footer != nil {
			footer.SetStatus("Settings cancelled")
		}
		if onClose != nil {
			onClose()
		}
	})

	saveBtn := widget.NewButton("Save", func() {
		requiresRestart, err := vm.SaveSettings(refreshCallback)
		if err != nil {
			_ = logging.Log.Error("Failed to save settings", "error", err)
			if footer != nil {
				footer.SetStatus("Failed to save settings")
			}
			dialog.ShowError(err, window)
			return
		}
		vm.Cleanup()
		// Close the overlay first, then update the footer so the success
		// message is not overwritten by the onClose callback.
		if onClose != nil {
			onClose()
		}
		logging.Log.Success("Settings saved")
		if footer != nil {
			footer.SetStatus("Settings saved")
		}
		if requiresRestart {
			showRestartPrompt(window, footer)
		}
	})
	saveBtn.Importance = widget.HighImportance

	actionBar := container.NewCenter(container.NewHBox(cancelBtn, saveBtn))
	actionBarPadded := container.NewPadded(actionBar)

	header := widget.NewRichTextFromMarkdown("# Settings")

	page := container.NewBorder(
		container.NewVBox(container.NewPadded(header), widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), actionBarPadded),
		nil,
		nil,
		split,
	)

	// Opaque background so the page completely covers whatever is beneath it
	// in the overlay stack (e.g. the tabs).
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	return container.NewStack(bg, page)
}

// buildCompleteForm creates the complete form container with all settings sections.\n//\n// Sections are built via buildSectionContent and appended in canonical order.\n// This is retained for tests and tooling that need a flat single-container\n// representation of the full settings form without the split-pane nav.\nfunc buildCompleteForm(vm *viewmodels.SettingsViewModel, currentValues map[string]interface{}, dependentFields *[]*dependentField, window fyne.Window) *fyne.Container {\n\tformContainer := container.NewVBox()\n\tallSchemas := settings.Get().GetAllSchemas()\n\tfor _, key := range vm.GetSectionOrder() {\n\t\tschema, exists := allSchemas[key]\n\t\tif !exists {\n\t\t\tcontinue\n\t\t}\n\t\tdisplayName := vm.GetSectionDisplayName(key)\n\t\tsectionContent := buildSectionContent(displayName, schema, vm, currentValues, dependentFields, window)\n\t\tformContainer.Add(sectionContent)\n\t}\n\treturn formContainer\n}

// buildSectionContent builds the field widgets for a single settings section.
//
// The returned container holds a ## heading followed by all field widgets for
// that section. It is intended to be wrapped in a VScroll by the caller.
func buildSectionContent(
	displayName string,
	schema settings.Schema,
	vm *viewmodels.SettingsViewModel,
	currentValues map[string]interface{},
	dependentFields *[]*dependentField,
	window fyne.Window,
) *fyne.Container {
	content := container.NewVBox()

	// Section heading mirrors the nav sidebar label so users know where they are.
	header := widget.NewRichTextFromMarkdown("## " + displayName)
	header.Wrapping = fyne.TextWrapWord
	content.Add(header)

	// Optional section description rendered as italicised text below the heading.
	if schema.Description != "" {
		content.Add(createFieldDescription(schema.Description))
	}

	buildFieldsFromSchema(content, displayName, schema, vm, currentValues, dependentFields, window, "")
	return content
}

// setupVisibilityManagement performs an initial visibility pass over all dependent
// fields and refreshes the form container.
//
// This must be called once after the form is fully built so that fields whose
// dependencies are already unsatisfied start hidden, rather than visible until
// the user first interacts with the controlling field.
func setupVisibilityManagement(refreshTarget fyne.CanvasObject, dependentFields *[]*dependentField) {
	updateAllVisibility := func() {
		for _, df := range *dependentFields {
			if df.updateVisibility != nil {
				df.updateVisibility()
			}
		}
		// Refresh triggers a layout recalculation so hidden fields don't leave gaps.
		refreshTarget.Refresh()
	}
	updateAllVisibility()
}

// buildFieldsFromSchema recursively builds form fields from a schema.
//
// Object-typed fields are treated as nested sections: their child fields are
// built into a sub-container, which is then shown/hidden as a unit when the
// object has a DependsOn rule. All other field types get an individual widget
// plus optional description and help text.
//
// The prefix parameter accumulates the dot-notation path for nested objects,
// e.g. prefix="s3." when building fields inside the S3 sub-section of Sync.
func buildFieldsFromSchema(
	formContainer *fyne.Container,
	sectionName string,
	schema settings.Schema,
	vm *viewmodels.SettingsViewModel,
	currentValues map[string]interface{},
	dependentFields *[]*dependentField,
	window fyne.Window,
	prefix string,
) {
	// Sort fields by their schema Order so the form is deterministic.
	fieldNames := vm.GetSortedFieldNames(schema.Fields)

	for _, fieldName := range fieldNames {
		fieldSchema := schema.Fields[fieldName]
		// Build the full flat key, e.g. "sync.s3.bucket".
		fullKey := vm.BuildFieldKey(sectionName, fieldName, prefix)

		if fieldSchema.Type == "object" {
			// Object fields group related child fields under a sub-section heading.
			nestedContainer := container.NewVBox()

			// Add a ### sub-heading only for top-level objects (prefix == "").
			// Deeply nested objects (prefix != "") intentionally have no extra heading.
			if fieldSchema.Label != "" && prefix == "" {
				header := widget.NewRichTextFromMarkdown("### " + fieldSchema.Label)
				header.Wrapping = fyne.TextWrapWord
				nestedContainer.Add(header)
				if fieldSchema.Description != "" {
					nestedContainer.Add(createFieldDescription(fieldSchema.Description))
				}
			}

			// Build the prefix for child fields:
			//   top-level object:  prefix=""     + "s3" → nestedPrefix="s3."
			//   already nested:    prefix="s3."  + "advanced" → nestedPrefix="s3.advanced."
			var nestedPrefix string
			if prefix == "" {
				nestedPrefix = fieldName + "."
			} else {
				nestedPrefix = prefix + fieldName + "."
			}

			if fieldSchema.Nested != nil {
				buildFieldsFromSchema(nestedContainer, sectionName, *fieldSchema.Nested, vm, currentValues, dependentFields, window, nestedPrefix)
			}

			// Wire up show/hide for the entire sub-section if it has a dependency.
			if fieldSchema.DependsOn != nil {
				updateVisibility := func() {
					visible := vm.EvaluateDependency(fieldSchema.DependsOn, sectionName)
					logging.Debug.Logf("Nested field %s visibility: %v (depends on %s)", fullKey, visible, fieldSchema.DependsOn.Field)
					if visible {
						nestedContainer.Show()
					} else {
						nestedContainer.Hide()
					}
				}
				*dependentFields = append(*dependentFields, &dependentField{
					container:        nestedContainer,
					fieldName:        fullKey,
					dependency:       fieldSchema.DependsOn,
					updateVisibility: updateVisibility,
				})
				updateVisibility()
			}

			formContainer.Add(nestedContainer)
			continue
		}

		// Non-object fields each get their own VBox container holding the label
		// (if applicable), the widget, and optional description/help text.
		fieldContainer := container.NewVBox()

		// Boolean fields carry their label inside the checkbox widget itself,
		// so we don't add a separate ### heading above them.
		if fieldSchema.Type != "bool" {
			fieldContainer.Add(createFieldLabel(fieldSchema))
		}

		var fieldWidget fyne.CanvasObject
		switch fieldSchema.Type {
		case "bool":
			fieldWidget = createBoolField(fullKey, fieldSchema, currentValues, dependentFields)
		case "string":
			if len(fieldSchema.Enum) > 0 {
				// Enum strings become a dropdown select.
				fieldWidget = createEnumField(fullKey, fieldSchema, currentValues, dependentFields)
			} else {
				fieldWidget = createStringField(fullKey, fieldSchema, currentValues, dependentFields)
			}
		case "int":
			fieldWidget = createIntField(fullKey, currentValues, dependentFields)
		case "file":
			fieldWidget = createFileField(fullKey, fieldSchema, currentValues, window, dependentFields)
		default:
			logging.Log.Warn("Unknown field type", "type", fieldSchema.Type, "field", fieldName)
			continue
		}

		fieldContainer.Add(fieldWidget)

		// Description appears below the widget as italicised text.
		if fieldSchema.Description != "" {
			fieldContainer.Add(createFieldDescription(fieldSchema.Description))
		}
		// Help text is additional guidance shown below the description.
		if fieldSchema.HelpText != "" {
			help := widget.NewRichTextFromMarkdown("*" + fieldSchema.HelpText + "*")
			help.Wrapping = fyne.TextWrapWord
			fieldContainer.Add(help)
		}

		// Small transparent spacer so fields aren't cramped together.
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(1, 5))
		fieldContainer.Add(spacer)

		// Wire up show/hide for this individual field if it has a dependency.
		if fieldSchema.DependsOn != nil {
			updateVisibility := func() {
				visible := vm.EvaluateDependency(fieldSchema.DependsOn, sectionName)
				logging.Debug.Logf("Field %s visibility: %v (depends on %s)", fullKey, visible, fieldSchema.DependsOn.Field)
				if visible {
					fieldContainer.Show()
				} else {
					fieldContainer.Hide()
				}
			}
			*dependentFields = append(*dependentFields, &dependentField{
				container:        fieldContainer,
				fieldName:        fullKey,
				dependency:       fieldSchema.DependsOn,
				updateVisibility: updateVisibility,
			})
			updateVisibility()
		}

		formContainer.Add(fieldContainer)
	}
}

// createBoolField creates a checkbox widget for boolean fields.
//
// The checkbox label is taken from the field schema; if absent it is derived
// from the human-readable form of the key. Toggling the checkbox immediately
// updates currentValues and re-evaluates all dependent field visibilities.
func createBoolField(key string, fieldSchema settings.FieldSchema, currentValues map[string]interface{}, dependentFields *[]*dependentField) fyne.CanvasObject {
	value, _ := currentValues[key].(bool)

	// Prefer the schema label; fall back to a humanised version of the key.
	label := fieldSchema.Label
	if label == "" {
		label = humanizeName(key)
	}

	check := widget.NewCheck(label, func(checked bool) {
		currentValues[key] = checked
		// Re-evaluate visibility of any fields that depend on this checkbox.
		updateDependentFieldsVisibility(dependentFields)
	})
	check.Checked = value
	return check
}

// createStringField creates a single-line text entry for free-form string fields.
//
// An optional placeholder from the schema is shown when the entry is empty.
// Every keystroke updates currentValues and re-evaluates dependent visibilities
// so that fields controlled by a text value react immediately.
func createStringField(key string, fieldSchema settings.FieldSchema, currentValues map[string]interface{}, dependentFields *[]*dependentField) fyne.CanvasObject {
	value, _ := currentValues[key].(string)
	entry := widget.NewEntry()
	entry.SetText(value)
	if fieldSchema.Placeholder != "" {
		entry.SetPlaceHolder(fieldSchema.Placeholder)
	}
	entry.OnChanged = func(text string) {
		currentValues[key] = text
		updateDependentFieldsVisibility(dependentFields)
	}
	return entry
}

// createIntField creates a text entry that accepts only integer values.
//
// Non-numeric input is silently ignored: the entry text updates but currentValues
// is only changed when strconv.Atoi succeeds, preventing invalid state.
func createIntField(key string, currentValues map[string]interface{}, dependentFields *[]*dependentField) fyne.CanvasObject {
	value, _ := currentValues[key].(int)
	entry := widget.NewEntry()
	entry.SetText(strconv.Itoa(value))
	entry.OnChanged = func(text string) {
		// Only commit the value when the text parses as a valid integer.
		if intVal, err := strconv.Atoi(text); err == nil {
			currentValues[key] = intVal
			updateDependentFieldsVisibility(dependentFields)
		}
	}
	return entry
}

// createEnumField creates a dropdown select widget for fields with a fixed set of options.
//
// The options slice is copied from the schema so mutations to the schema after
// construction don't affect the displayed list. The currently saved value is
// pre-selected. Changing the selection immediately updates currentValues.
func createEnumField(key string, fieldSchema settings.FieldSchema, currentValues map[string]interface{}, dependentFields *[]*dependentField) fyne.CanvasObject {
	value, _ := currentValues[key].(string)

	// Copy options so the widget's internal slice is independent of the schema.
	options := make([]string, len(fieldSchema.Enum))
	copy(options, fieldSchema.Enum)

	selector := widget.NewSelect(options, func(selected string) {
		currentValues[key] = selected
		// Re-evaluate visibility for any fields that depend on this selector.
		updateDependentFieldsVisibility(dependentFields)
	})
	selector.SetSelected(value)
	return selector
}

// createFileField creates a file-path field with an editable entry and a Browse button.
//
// The entry is fully editable so users can type a path directly. The Browse
// button opens a file-open dialog pre-seeded at the parent of the current value
// (if any). Selecting a file via the dialog updates both currentValues and the
// visible entry text. The dialog is sized according to the GUI settings.
func createFileField(key string, fieldSchema settings.FieldSchema, currentValues map[string]interface{}, window fyne.Window, dependentFields *[]*dependentField) fyne.CanvasObject {
	value, _ := currentValues[key].(string)

	pathEntry := widget.NewEntry()
	pathEntry.SetText(value)
	if fieldSchema.Placeholder != "" {
		pathEntry.SetPlaceHolder(fieldSchema.Placeholder)
	}

	// Keep currentValues in sync as the user types.
	pathEntry.OnChanged = func(text string) {
		currentValues[key] = text
		updateDependentFieldsVisibility(dependentFields)
	}

	button := widget.NewButton("Browse...", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			path := reader.URI().Path()
			currentValues[key] = path
			pathEntry.SetText(path)
			logging.Debug.Logf("File selected for %s: %s", key, path)
			updateDependentFieldsVisibility(dependentFields)
		}, window)

		// Pre-navigate the dialog to the parent directory of the current value.
		if value != "" {
			if uri := storage.NewFileURI(value); uri != nil {
				parentURI, err := storage.Parent(uri)
				if err == nil {
					if lister, err := storage.ListerForURI(parentURI); err == nil && lister != nil {
						fileDialog.SetLocation(lister)
					}
				}
			}
		}

		// Size the file dialog according to the configured dialog dimensions.
		guiSettings := settings.Get().GUI
		width, height := guiSettings.DialogWidth, guiSettings.DialogHeight
		fileDialog.Resize(fyne.NewSize(float32(width), float32(height)))

		fileDialog.Show()
	})

	// Place the entry and browse button side-by-side, entry taking all remaining width.
	filePicker := container.NewBorder(nil, nil, nil, button, pathEntry)

	return filePicker
}

// updateDependentFieldsVisibility calls the updateVisibility function on every
// tracked dependent field, causing each one to show or hide itself based on the
// current form values. This is called after any widget value changes.
func updateDependentFieldsVisibility(dependentFields *[]*dependentField) {
	if dependentFields == nil {
		return
	}
	for _, df := range *dependentFields {
		if df != nil && df.updateVisibility != nil {
			df.updateVisibility()
		}
	}
}

// createFieldLabel creates a ### Markdown heading for a non-boolean form field.
//
// If the schema has no Label the key is humanised via humanizeName as a fallback.
func createFieldLabel(fieldSchema settings.FieldSchema) *widget.RichText {
	label := fieldSchema.Label
	if label == "" {
		// humanizeName converts e.g. "managed_section_start" to "Managed Section Start".
		label = humanizeName(fieldSchema.Label)
	}
	richText := widget.NewRichTextFromMarkdown("### " + label)
	richText.Wrapping = fyne.TextWrapWord
	return richText
}

// createFieldDescription creates an italic Markdown description
func createFieldDescription(description string) *widget.RichText {
	richText := widget.NewRichTextFromMarkdown("*" + description + "*")
	richText.Wrapping = fyne.TextWrapWord
	return richText
}

// humanizeName converts snake_case to Human Readable.
func humanizeName(name string) string {
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// showRestartPrompt shows a dialog asking the user to restart the application.
//
// One or more saved settings require a restart to take effect. The user can
// choose to restart immediately (via syscall.Exec) or dismiss and restart later.
func showRestartPrompt(window fyne.Window, footer *components.Footer) {
	var restartDialog *dialog.CustomDialog

	content := container.NewVBox(
		widget.NewRichTextFromMarkdown("## Restart Required"),
		widget.NewLabel("Some settings require a restart to take effect. Would you like to restart now?"),
	)

	restartDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:   "Restart Required",
		Content: content,
		Buttons: []components.DialogButton{
			{
				Label: "Later",
				OnTapped: func() {
					restartDialog.Hide()
					if footer != nil {
						footer.SetStatus("Restart required to apply some settings")
					}
				},
			},
			{
				Label:      "Restart Now",
				Importance: widget.HighImportance,
				OnTapped: func() {
					restartDialog.Hide()
					restartApp()
				},
			},
		},
		Window: window,
	})
	restartDialog.Show()
}

// restartApp re-executes the current binary with the same arguments and environment.
func restartApp() {
	exe, err := os.Executable()
	if err != nil {
		_ = logging.Log.Error("Failed to get executable path for restart", "error", err)
		return
	}
	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		_ = logging.Log.Error("Failed to restart application", "error", err)
	}
}
