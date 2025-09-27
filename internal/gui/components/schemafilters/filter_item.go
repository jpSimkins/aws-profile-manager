package schemafilters

import (
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FilterItem renders one selectable filter dimension.
//
// Programmatic updates via SetAvailableOptions and SetSelectedOptions are
// silent and do not trigger OnChange. OnChange is invoked only for direct
// user actions (checkbox interaction, Check All, Clear).
type FilterItem struct {
	props                  FilterItemProps
	available              []string
	selectedSet            map[string]struct{}
	searchText             string
	isSyncingCheckboxState bool
	displayTotalCount      int
	// Set by grid layout; last-row cards can opt out of per-card max-height cap.
	isLastInRow bool

	searchEntry *widget.Entry
	countLabel  *widget.Label
	checkAllBtn *widget.Button
	clearBtn    *widget.Button
	list        *fyne.Container
	listScroll  *container.Scroll
	root        *fyne.Container
}

// NewFilterItem creates a new FilterItem from the given props.
func NewFilterItem(props FilterItemProps) *FilterItem {
	item := &FilterItem{
		props:             props,
		available:         sortedUniqueStrings(props.AvailableOptions),
		selectedSet:       toStringSet(props.SelectedOptions),
		displayTotalCount: len(sortedUniqueStrings(props.AvailableOptions)),
	}
	item.pruneSelectionsToAvailable()
	item.build()
	item.renderOptionList(false)
	item.refreshControlState()
	return item
}

// GetContent returns the filter control canvas object.
func (f *FilterItem) GetContent() fyne.CanvasObject {
	return f.root
}

// GetSelectedOptions returns a sorted snapshot of selected options.
func (f *FilterItem) GetSelectedOptions() []string {
	return sortedSetKeys(f.selectedSet)
}

// SetSelectedOptions replaces selected options silently.
func (f *FilterItem) SetSelectedOptions(selected []string) {
	f.selectedSet = toStringSet(selected)
	f.pruneSelectionsToAvailable()
	f.renderOptionList(false)
	f.refreshControlState()
}

// SetAvailableOptions replaces available options silently.
//
// Any selected values that are no longer available are pruned without
// triggering OnChange.
func (f *FilterItem) SetAvailableOptions(options []string) {
	f.available = sortedUniqueStrings(options)
	f.pruneSelectionsToAvailable()
	f.renderOptionList(false)
	f.refreshControlState()
}

// SetDisplayTotalCount overrides the denominator shown in the count label.
//
// This is useful for keeping totals stable against dynamic cascading option
// updates, so users can understand filtered availability vs original counts.
func (f *FilterItem) SetDisplayTotalCount(totalCount int) {
	if totalCount < 0 {
		totalCount = 0
	}
	f.displayTotalCount = totalCount
	f.refreshControlState()
}

// SetIsLastInRow marks this filter as the last item in its row.
//
// Last-row filters are allowed to expand beyond MaxHeight since they don't
// need to align with items below them.
func (f *FilterItem) SetIsLastInRow(isLast bool) {
	if f.isLastInRow == isLast {
		// Avoid unnecessary relayout/refresh churn.
		return
	}
	f.isLastInRow = isLast
	f.refreshListScrollHeight()
}

func (f *FilterItem) build() {
	header := widget.NewRichTextFromMarkdown("### " + f.props.Label)

	controls := []fyne.CanvasObject{}
	if f.props.AllowSearch {
		f.searchEntry = widget.NewEntry()
		f.searchEntry.SetPlaceHolder(f.searchPlaceholder())
		f.searchEntry.OnChanged = func(value string) {
			f.searchText = strings.TrimSpace(value)
			f.renderOptionList(false)
			f.refreshControlState()
		}
		controls = append(controls, f.searchEntry)
	}

	f.countLabel = widget.NewLabel("")
	controls = append(controls, f.countLabel)

	buttonRow := []fyne.CanvasObject{}
	if f.props.AllowCheckAll {
		f.checkAllBtn = widget.NewButton("Check All", func() {
			for _, option := range f.filteredAvailableOptions() {
				f.selectedSet[option] = struct{}{}
			}
			f.renderOptionList(true)
			f.refreshControlState()
		})
		buttonRow = append(buttonRow, f.checkAllBtn)
	}

	f.clearBtn = widget.NewButton("Clear", func() {
		f.selectedSet = map[string]struct{}{}
		f.renderOptionList(true)
		f.refreshControlState()
	})
	buttonRow = append(buttonRow, f.clearBtn)

	if len(buttonRow) > 0 {
		controls = append(controls, container.NewHBox(buttonRow...))
	}

	// Use a simple VBox for the checkboxes - no internal scroll
	// The parent scroll container will handle all scrolling
	// This prevents click event capture issues
	f.list = container.NewVBox()

	controls = append(controls, f.list)
	f.root = container.NewVBox(append([]fyne.CanvasObject{header}, controls...)...)
}

func (f *FilterItem) searchPlaceholder() string {
	if strings.TrimSpace(f.props.SearchPlaceholder) != "" {
		return f.props.SearchPlaceholder
	}
	return "Search..."
}

func (f *FilterItem) renderOptionList(emitOnChange bool) {
	visible := f.filteredAvailableOptions()
	// Rebuild visible checkbox list from current search + option set.
	objects := make([]fyne.CanvasObject, 0, len(visible))
	for _, option := range visible {
		optionName := option
		_, selected := f.selectedSet[optionName]
		check := widget.NewCheck(optionName, func(isChecked bool) {
			if f.isSyncingCheckboxState {
				return
			}
			if isChecked {
				f.selectedSet[optionName] = struct{}{}
			} else {
				delete(f.selectedSet, optionName)
			}
			f.refreshControlState()
			if f.props.OnChange != nil {
				f.props.OnChange(f.GetSelectedOptions())
			}
		})
		f.isSyncingCheckboxState = true
		check.SetChecked(selected)
		f.isSyncingCheckboxState = false
		objects = append(objects, check)
	}

	f.list.Objects = objects
	f.list.Refresh()
	// Recompute desired card height whenever visible options change.
	f.refreshListScrollHeight()

	if emitOnChange && f.props.OnChange != nil {
		f.props.OnChange(f.GetSelectedOptions())
	}
}

// refreshListScrollHeight sets MinSize based on actual content height capped by MaxHeight.
//
// This tells the adaptive layout how tall this filter naturally wants to be,
// which is used to calculate row heights where all items match the tallest in the row.
// Last-row items are allowed to exceed MaxHeight since they don't need row alignment.
func (f *FilterItem) refreshListScrollHeight() {
	if f.listScroll == nil {
		return
	}

	// Calculate actual content height from rendered list
	contentHeight := f.list.MinSize().Height
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Cap by MaxHeight for aligned rows; allow full height on last row.
	displayHeight := contentHeight
	if !f.isLastInRow && f.props.MaxHeight > 0 && displayHeight > f.props.MaxHeight {
		displayHeight = f.props.MaxHeight
	}

	// Report this height to the layout so row alignment works correctly
	f.listScroll.SetMinSize(fyne.NewSize(1, displayHeight))
}

func (f *FilterItem) filteredAvailableOptions() []string {
	if f.searchText == "" {
		return append([]string{}, f.available...)
	}

	query := strings.ToLower(f.searchText)
	out := make([]string, 0, len(f.available))
	for _, option := range f.available {
		if strings.Contains(strings.ToLower(option), query) {
			out = append(out, option)
		}
	}
	return out
}

func (f *FilterItem) pruneSelectionsToAvailable() {
	availableSet := toStringSet(f.available)
	for selected := range f.selectedSet {
		if _, exists := availableSet[selected]; !exists {
			delete(f.selectedSet, selected)
		}
	}
}

func (f *FilterItem) refreshControlState() {
	visibleCount := len(f.filteredAvailableOptions())
	totalCount := len(f.available)
	totalCountForDisplay := totalCount
	if f.displayTotalCount > 0 {
		totalCountForDisplay = f.displayTotalCount
	}
	f.countLabel.SetText(strings.TrimSpace(strings.Join([]string{
		strconv.Itoa(visibleCount),
		"available",
		"(",
		strconv.Itoa(totalCountForDisplay),
		"total)",
	}, " ")))

	hasOptions := totalCount > 0
	hasVisible := visibleCount > 0
	hasSelection := len(f.selectedSet) > 0

	if f.checkAllBtn != nil {
		if hasVisible {
			f.checkAllBtn.Enable()
		} else {
			f.checkAllBtn.Disable()
		}
	}

	if hasSelection {
		f.clearBtn.Enable()
	} else {
		f.clearBtn.Disable()
	}

	if !hasOptions {
		if f.searchEntry != nil {
			f.searchEntry.Disable()
		}
		if f.checkAllBtn != nil {
			f.checkAllBtn.Disable()
		}
		f.clearBtn.Disable()
		return
	}

	if f.searchEntry != nil {
		f.searchEntry.Enable()
	}
}

func toStringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		set[value] = struct{}{}
	}
	return set
}

func sortedUniqueStrings(values []string) []string {
	set := toStringSet(values)
	return sortedSetKeys(set)
}

func sortedSetKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
