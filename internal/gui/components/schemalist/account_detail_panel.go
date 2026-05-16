package schemalist

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components/actionbuttons"
)

// buildAccountDetailsContent builds the full detail panel for a selected account.
//
// The panel is organised into four sections separated by visual dividers:
//   - 🪪 Identity:  name, alias, account ID, organization, partition
//   - 🔐 SSO:       session name (copyable) and portal URL (copyable)
//   - 🌎 Regions:   default region; each additional region on its own copyable row
//   - 👤 Roles:     each role shown as its full CLI profile name (copyable)
//
// Copyable fields show a small icon button on the right. Tapping it writes
// the value to the clipboard and briefly shows a checkmark to confirm.
//
// When showCliButtons is false the terminal launch and SSO login buttons are
// hidden. The browser URL button is always shown.
func buildAccountDetailsContent(record AccountRecord, showCliButtons bool) fyne.CanvasObject {
	// SSO session name convention: <org-alias>-<partition>
	ssoSessionName := record.OrganizationAlias + "-" + record.PartitionName

	// --- 🪪 Identity ---
	identitySection := container.NewVBox(
		detailSectionHeader("🪪  Identity"),
		detailRow("Name", safeValue(record.AccountName), ""),
		detailRow("Alias", safeValue(record.AccountAlias), record.AccountAlias),
		detailRow("Account ID", safeValue(record.AccountID), record.AccountID),
		detailRow("Organization", safeValue(record.OrganizationName)+" ("+safeValue(record.OrganizationAlias)+")", ""),
		detailRow("Partition", safeValue(record.PartitionName), record.PartitionName),
	)

	// --- 🔐 SSO ---
	ssoSection := container.NewVBox(
		detailSectionHeader("🔐  SSO"),
		ssoSessionRow(ssoSessionName, showCliButtons),
		ssoURLRow(safeValue(record.SsoURL), record.SsoURL),
	)

	// --- 🌎 Regions ---
	regionsSection := container.NewVBox(
		detailSectionHeader("🌎  Regions"),
		detailRow("Default", safeValue(record.DefaultRegion), record.DefaultRegion),
		buildRegionsSection(record.Regions),
	)

	// --- 👤 Roles ---
	// Each role is rendered as its full CLI profile name: <partition>-<alias>-<role>
	rolesSection := buildRolesSection(record, showCliButtons)

	return container.NewVScroll(container.NewPadded(container.NewVBox(
		identitySection,
		widget.NewSeparator(),
		ssoSection,
		widget.NewSeparator(),
		regionsSection,
		widget.NewSeparator(),
		rolesSection,
	)))
}

// detailSectionHeader returns an h2 RichText widget used as a section title.
func detailSectionHeader(title string) fyne.CanvasObject {
	return widget.NewRichTextFromMarkdown("## " + title)
}

// ssoSessionRow renders the SSO session name with a copy button and, when
// showCliButtons is true, a terminal launch button that runs
// `aws sso login --sso-session <name>`.
func ssoSessionRow(sessionName string, showCliButtons bool) fyne.CanvasObject {
	labelWidget := widget.NewLabel("Session:")
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	valueWidget := widget.NewLabel(sessionName)
	valueWidget.Wrapping = fyne.TextWrapBreak

	var buttons *fyne.Container
	if showCliButtons {
		buttons = container.NewHBox(
			actionbuttons.SsoLogin(sessionName),
			actionbuttons.Copy(sessionName),
		)
	} else {
		buttons = container.NewHBox(actionbuttons.Copy(sessionName))
	}
	valueRow := container.NewBorder(nil, nil, nil, buttons, valueWidget)
	return container.NewVBox(labelWidget, indented(valueRow))
}

// ssoURLRow renders the SSO portal URL with a copy button and a browser
// open button. When the raw URL is empty the browser button is hidden.
func ssoURLRow(displayValue, rawURL string) fyne.CanvasObject {
	labelWidget := widget.NewLabel("URL:")
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	valueWidget := widget.NewLabel(displayValue)
	valueWidget.Wrapping = fyne.TextWrapBreak

	buttons := container.NewHBox(
		actionbuttons.OpenURL(rawURL),
		actionbuttons.Copy(rawURL),
	)
	valueRow := container.NewBorder(nil, nil, nil, buttons, valueWidget)
	return container.NewVBox(labelWidget, indented(valueRow))
}

// detailRow renders a single field as a label → indented value pair.
//
// The label is bold and sits on its own line. The value appears below it,
// indented to give a definition-list appearance. When copyValue is non-empty
// a small copy icon button is shown to the right of the value.
func detailRow(label, displayValue, copyValue string) fyne.CanvasObject {
	labelWidget := widget.NewLabel(label + ":")
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	valueWidget := widget.NewLabel(displayValue)
	valueWidget.Wrapping = fyne.TextWrapBreak

	var valueContent fyne.CanvasObject
	if copyValue == "" {
		valueContent = valueWidget
	} else {
		valueContent = container.NewBorder(nil, nil, nil, actionbuttons.Copy(copyValue), valueWidget)
	}

	return container.NewVBox(labelWidget, indented(valueContent))
}

// indented wraps content with a fixed left margin to produce a definition-list indent.
func indented(content fyne.CanvasObject) fyne.CanvasObject {
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(16, 1))
	return container.NewBorder(nil, nil, spacer, nil, content)
}

// buildRegionsSection renders the "All Regions" list inside the Regions block.
//
// Each region is shown on its own row with a label and a copy icon button.
// Returns a single "n/a" label when the slice is empty.
func buildRegionsSection(regions []string) fyne.CanvasObject {
	if len(regions) == 0 {
		labelWidget := widget.NewLabel("All:")
		labelWidget.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewVBox(labelWidget, indented(widget.NewLabel("n/a")))
	}

	labelWidget := widget.NewLabel("All:")
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	rows := []fyne.CanvasObject{labelWidget}
	for _, region := range regions {
		r := region // capture for closure
		row := container.NewBorder(nil, nil, nil, actionbuttons.Copy(r), widget.NewLabel(r))
		rows = append(rows, indented(row))
	}
	return container.NewVBox(rows...)
}

// buildRolesSection renders the Roles block.
//
// Each role occupies its own row showing the full CLI profile name.
// buildRolesSection renders the Roles block.
//
// Each role occupies its own row showing the full CLI profile name.
// When showCliButtons is true, a terminal launch button appears alongside the
// copy button. When false, only the copy button is shown.
func buildRolesSection(record AccountRecord, showCliButtons bool) fyne.CanvasObject {
	items := []fyne.CanvasObject{detailSectionHeader("👤  Roles")}

	if len(record.Roles) == 0 {
		items = append(items, widget.NewLabel("n/a"))
		return container.NewVBox(items...)
	}

	for _, role := range record.Roles {
		// CLI profile name convention: <partition>-<account-alias>-<role>
		profileName := record.PartitionName + "-" + record.AccountAlias + "-" + role
		items = append(items, roleRow(role, profileName, record.DefaultRegion, showCliButtons))
	}

	return container.NewVBox(items...)
}

// roleRow renders a single role entry with a profile name label, a copy button,
// and, when showCliButtons is true, a terminal launch button.
func roleRow(roleName, profileName, defaultRegion string, showCliButtons bool) fyne.CanvasObject {
	labelWidget := widget.NewLabel(roleName + ":")
	labelWidget.TextStyle = fyne.TextStyle{Bold: true}

	valueWidget := widget.NewLabel(profileName)
	valueWidget.Wrapping = fyne.TextWrapBreak

	var buttons *fyne.Container
	if showCliButtons {
		buttons = container.NewHBox(
			actionbuttons.Terminal(profileName, defaultRegion),
			actionbuttons.Copy(profileName),
		)
	} else {
		buttons = container.NewHBox(actionbuttons.Copy(profileName))
	}
	valueRow := container.NewBorder(nil, nil, nil, buttons, valueWidget)

	return container.NewVBox(labelWidget, indented(valueRow))
}
