// Package schemalist provides GUI components for rendering schema results.
package schemalist

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/schema"
)

// accountListSplitOffset controls the list/detail split ratio.
// The list pane takes this fraction of the total width; the detail panel gets the rest.
const accountListSplitOffset float64 = 0.42

// AccountListView is a master-detail component for browsing AWS accounts.
//
// Layout:
//
// ┌─────────────────────┬─────────────────────────────────────┐
// │  Account list       │  Detail panel                       │
// │  (compact rows)     │  (identity, SSO, regions, roles)    │
// │                     │                                     │
// │  > acme commercial  │  🪪 Identity                        │
// │    acme govcloud    │    Name:    Acme Prod                │
// │    widget prod      │    Alias:   acme-prod      [copy]   │
// │    ...              │    ...                              │
// └─────────────────────┴─────────────────────────────────────┘
//
// The left list stays compact; tapping a row reveals full account details
// in the right panel without expanding into the list or obscuring other rows.
//
// Public API is intentionally minimal — callers only need SetSchema and GetContent.
type AccountListView struct {
	sourceSchema   *schema.Schema
	accountRecords []AccountRecord

	// filteredRecords is the subset of accountRecords matching the current search.
	// When searchText is empty this equals accountRecords.
	filteredRecords []AccountRecord

	// searchText is the current lower-cased search query.
	searchText string

	// showCliButtons controls whether terminal/SSO-login buttons are rendered
	// in the detail panel. Set to false on the Install tab where profiles do
	// not yet exist in the AWS config. Browser buttons are unaffected.
	showCliButtons bool

	// list is the compact left-side account list.
	list *widget.List

	// detailContainer holds the currently displayed detail panel (right side).
	// Swapped out on each selection change.
	detailContainer *fyne.Container

	// selectedIndex tracks the currently selected list row (-1 = none).
	selectedIndex int

	// rootContainer is the top-level widget returned by GetContent.
	rootContainer fyne.CanvasObject
}

// NewAccountListView creates a new AccountListView populated from sourceSchema.
func NewAccountListView(sourceSchema *schema.Schema) *AccountListView {
	v := &AccountListView{selectedIndex: -1}
	v.buildUI()
	v.SetSchema(sourceSchema)
	return v
}

// GetContent returns the root canvas object to embed in a parent layout.
func (v *AccountListView) GetContent() fyne.CanvasObject {
	return v.rootContainer
}

// SetSchema replaces the current schema and refreshes the view.
//
// The selected account is cleared and the detail panel returns to its placeholder.
func (v *AccountListView) SetSchema(sourceSchema *schema.Schema) {
	v.sourceSchema = sourceSchema
	v.accountRecords = FlattenManagedAccounts(sourceSchema)
	v.selectedIndex = -1
	v.applySearch()
	v.showDetailPlaceholder()
	v.list.Refresh()
}

// GetAccountRecords returns a copy of the currently displayed account records.
func (v *AccountListView) GetAccountRecords() []AccountRecord {
	cp := make([]AccountRecord, len(v.accountRecords))
	copy(cp, v.accountRecords)
	return cp
}

// WithCliButtons enables or disables the terminal and SSO-login CLI buttons in
// the detail panel. Call before the view is displayed; returns the receiver for
// fluent chaining.
//
// Defaults to false — callers that want CLI buttons must opt in explicitly.
func (v *AccountListView) WithCliButtons(show bool) *AccountListView {
	v.showCliButtons = show
	return v
}

// buildUI constructs the master-detail split layout.
//
// Called once during NewAccountListView. SetSchema handles subsequent data changes.
func (v *AccountListView) buildUI() {
	v.detailContainer = container.NewStack()
	v.filteredRecords = nil

	listPane := v.buildListPane()
	detailPane := v.buildDetailPane()

	// 35% list / 65% detail — detail panel gets the most space since it's content-heavy.
	split := container.NewHSplit(listPane, detailPane)
	split.Offset = accountListSplitOffset

	v.rootContainer = split
}

// buildListPane constructs the left-side account list with a search entry above it.
func (v *AccountListView) buildListPane() fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search accounts...")
	searchEntry.OnChanged = func(value string) {
		v.searchText = strings.ToLower(strings.TrimSpace(value))
		v.selectedIndex = -1
		v.showDetailPlaceholder()
		v.applySearch()
		v.list.Refresh()
	}

	v.list = widget.NewList(
		// Item count — use filtered slice.
		func() int { return len(v.filteredRecords) },

		// Template — a single label is enough; UpdateItem sets the text.
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Account")
			lbl.Truncation = fyne.TextTruncateEllipsis
			return lbl
		},

		// Update — populate the label for each row.
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(v.filteredRecords) {
				return
			}
			item.(*widget.Label).SetText(resolveAccountTitle(v.filteredRecords[id]))
		},
	)

	// On selection, render the chosen account's details in the right panel.
	v.list.OnSelected = func(id widget.ListItemID) {
		if id >= len(v.filteredRecords) {
			return
		}
		v.selectedIndex = id
		v.showAccountDetail(v.filteredRecords[id])
	}

	// Wrap in a border: header on top, search entry below it, list fills remaining space.
	header := widget.NewRichTextFromMarkdown("## Accounts")
	return container.NewBorder(
		container.NewVBox(header, searchEntry),
		nil, nil, nil,
		v.list,
	)
}

// buildDetailPane constructs the right-side detail panel container.
//
// Starts with a placeholder and is populated on list selection.
func (v *AccountListView) buildDetailPane() fyne.CanvasObject {
	v.showDetailPlaceholder()
	return v.detailContainer
}

// applySearch rebuilds filteredRecords from accountRecords using the current
// searchText. When searchText is empty all records are visible.
//
// A record matches when the search text appears (case-insensitive) in any of:
// account alias, account name, account ID, organization alias, or partition.
func (v *AccountListView) applySearch() {
	if v.searchText == "" {
		v.filteredRecords = v.accountRecords
		return
	}
	query := strings.ToLower(v.searchText)
	filtered := make([]AccountRecord, 0, len(v.accountRecords))
	for _, r := range v.accountRecords {
		if strings.Contains(strings.ToLower(r.AccountAlias), query) ||
			strings.Contains(strings.ToLower(r.AccountName), query) ||
			strings.Contains(strings.ToLower(r.AccountID), query) ||
			strings.Contains(strings.ToLower(r.OrganizationAlias), query) ||
			strings.Contains(strings.ToLower(r.PartitionName), query) {
			filtered = append(filtered, r)
		}
	}
	v.filteredRecords = filtered
}

// showDetailPlaceholder replaces the detail panel with a prompt to select an account.
func (v *AccountListView) showDetailPlaceholder() {
	placeholder := widget.NewRichTextFromMarkdown("*Select an account from the list to view details.*")
	v.detailContainer.Objects = []fyne.CanvasObject{
		container.NewCenter(placeholder),
	}
	v.detailContainer.Refresh()
}

// showAccountDetail replaces the detail panel with the full detail view for record.
func (v *AccountListView) showAccountDetail(record AccountRecord) {
	v.detailContainer.Objects = []fyne.CanvasObject{
		buildAccountDetailsContent(record, v.showCliButtons),
	}
	v.detailContainer.Refresh()
}
