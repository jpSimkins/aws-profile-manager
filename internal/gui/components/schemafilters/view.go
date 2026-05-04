package schemafilters

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"aws-profile-manager/internal/schema"
	"aws-profile-manager/internal/settings"
)

const (
	schemaFiltersMinColumnWidth float32 = 120
	schemaFiltersMaxColumnCount int     = 5
	// Inter-card spacing inside the filters grid.
	schemaFiltersColumnGap float32 = 12
	schemaFiltersRowGap    float32 = 12
	// Outer gutter applied on both left and right edges of the grid.
	schemaFiltersHorizontalPadding           float32 = 12
	schemaFiltersWidthBreakpointOneColumn    float32 = 520
	schemaFiltersWidthBreakpointTwoColumns   float32 = 860
	schemaFiltersWidthBreakpointThreeColumns float32 = 1200
	schemaFiltersWidthBreakpointFourColumns  float32 = 1540
)

// SchemaFiltersView is a GUI container for schema filter controls.
//
// It owns only UI state (schema.FilterCriteria) and delegates filtering to
// schema.FilterSchema. Consumers decide how to render filtered results.
type SchemaFiltersView struct {
	originalSchema          *schema.Schema
	activeCriteria          schema.FilterCriteria
	onFilteredSchemaChanged func(*schema.Schema)
	availableOptions        *AvailableOptionsProvider
	originalOptionsByType   map[FilterType][]string
	filterItemsByType       map[FilterType]*FilterItem
	filterConfigByType      map[FilterType]FilterItemConfig
	filtersContainer        *fyne.Container
	rootContainer           *fyne.Container
	lastComputedColumns     int
}

// NewSchemaFiltersView initializes a new SchemaFiltersView.
func NewSchemaFiltersView(original *schema.Schema) *SchemaFiltersView {
	gui := settings.Get().GUI

	v := &SchemaFiltersView{
		originalSchema:        original,
		availableOptions:      NewAvailableOptionsProvider(original),
		originalOptionsByType: map[FilterType][]string{},
		filterItemsByType:     map[FilterType]*FilterItem{},
		rootContainer:         container.NewVBox(), // Changed from NewMax() to NewVBox() to prevent overflow
		filterConfigByType: map[FilterType]FilterItemConfig{
			FilterTypeOrganizations: {AllowSearch: gui.FilterSearchOrganizations, AllowCheckAll: true, MaxHeight: 180, SearchPlaceholder: "Search organizations"},
			FilterTypePartitions:    {AllowSearch: gui.FilterSearchPartitions, AllowCheckAll: true, MaxHeight: 180, SearchPlaceholder: "Search partitions"},
			FilterTypeRegions:       {AllowSearch: gui.FilterSearchRegions, AllowCheckAll: true, MaxHeight: 180, SearchPlaceholder: "Search regions"},
			FilterTypeRoles:         {AllowSearch: gui.FilterSearchRoles, AllowCheckAll: true, MaxHeight: 180, SearchPlaceholder: "Search roles"},
			FilterTypeAccounts:      {AllowSearch: gui.FilterSearchAccounts, AllowCheckAll: true, MaxHeight: 300, SearchPlaceholder: "Search accounts"},
		},
	}

	v.buildFilterControlsContainer()
	v.refreshOriginalOptionsCache()
	v.recalculateAvailableOptionsAndPruneSelections()
	return v
}

// GetContent returns the root canvas object for embedding in a layout.
func (v *SchemaFiltersView) GetContent() fyne.CanvasObject {
	return v.rootContainer
}

// OnFilterChange sets the callback invoked with filtered schema results.
func (v *SchemaFiltersView) OnFilterChange(callback func(*schema.Schema)) {
	v.onFilteredSchemaChanged = callback
}

// SetSchema replaces the source schema and rebuilds filter state.
//
// Existing criteria is cleared because available options can change
// substantially with a new source schema.
func (v *SchemaFiltersView) SetSchema(sourceSchema *schema.Schema) {
	v.originalSchema = sourceSchema
	v.availableOptions.SetSchema(sourceSchema)
	v.activeCriteria = schema.FilterCriteria{}
	v.buildFilterControlsContainer()
	v.refreshOriginalOptionsCache()
	v.recalculateAvailableOptionsAndPruneSelections()
}

// Reset clears criteria and emits the default-state schema.
func (v *SchemaFiltersView) Reset() {
	v.activeCriteria = schema.FilterCriteria{}
	for _, filterType := range FilterTypeOrder {
		item := v.filterItemsByType[filterType]
		if item == nil {
			continue
		}
		item.SetSelectedOptions(nil)
	}
	v.recalculateAvailableOptionsAndPruneSelections()
	v.emitCurrentFilteredSchema()
}

// GetCriteria returns a snapshot of current filter criteria.
func (v *SchemaFiltersView) GetCriteria() schema.FilterCriteria {
	return schema.FilterCriteria{
		Organizations: append([]string{}, v.activeCriteria.Organizations...),
		Partitions:    append([]string{}, v.activeCriteria.Partitions...),
		Regions:       append([]string{}, v.activeCriteria.Regions...),
		Roles:         append([]string{}, v.activeCriteria.Roles...),
		Accounts:      append([]string{}, v.activeCriteria.Accounts...),
		AllRegions:    v.activeCriteria.AllRegions,
	}
}

// ConfigureFilter updates configuration for a specific filter and rebuilds it.
func (v *SchemaFiltersView) ConfigureFilter(filterType FilterType, config FilterItemConfig) {
	v.filterConfigByType[filterType] = config
	v.buildFilterControlsContainer()
	v.refreshOriginalOptionsCache()
	v.recalculateAvailableOptionsAndPruneSelections()
}

// ApplyPreset applies filter criteria from a preset to the view.
//
// This programmatically sets the filter selections based on the preset
// configuration. Empty arrays in the preset mean "no filter" (include all).
// The AllRegions flag is also applied if set in the preset.
//
// After applying the preset, the filtered schema callback is triggered.
func (v *SchemaFiltersView) ApplyPreset(preset *schema.Preset) {
	if preset == nil {
		v.Reset()
		return
	}

	// Build criteria from preset
	criteria := schema.FilterCriteria{
		Organizations: append([]string{}, preset.Organizations...),
		Partitions:    append([]string{}, preset.Partitions...),
		Accounts:      append([]string{}, preset.Accounts...),
		Roles:         append([]string{}, preset.Roles...),
		Regions:       append([]string{}, preset.Regions...),
		AllRegions:    preset.AllRegions,
	}

	// Apply criteria to filter items
	v.activeCriteria = criteria

	// Update UI to reflect new criteria
	for _, filterType := range FilterTypeOrder {
		item := v.filterItemsByType[filterType]
		if item == nil {
			continue
		}
		item.SetSelectedOptions(v.criteriaValuesForFilterType(filterType))
	}

	// Emit filtered schema without recalculating available options
	// Presets should just check the boxes, not narrow down the available options list
	v.emitCurrentFilteredSchema()
}

func (v *SchemaFiltersView) buildFilterControlsContainer() {
	// Force row-flag recomputation after any rebuild (new schema/config).
	v.lastComputedColumns = 0
	objects := make([]fyne.CanvasObject, 0, len(FilterTypeOrder))
	for _, filterType := range FilterTypeOrder {
		config := v.filterConfigByType[filterType]
		item := NewFilterItem(FilterItemProps{
			Label:             FilterTypeLabels[filterType],
			AvailableOptions:  nil,
			SelectedOptions:   v.criteriaValuesForFilterType(filterType),
			OnChange:          v.handleFilterItemSelectionChanged(filterType),
			AllowSearch:       config.AllowSearch,
			AllowCheckAll:     config.AllowCheckAll,
			MaxHeight:         config.MaxHeight,
			SearchPlaceholder: config.SearchPlaceholder,
		})
		v.filterItemsByType[filterType] = item
		objects = append(objects, item.GetContent())
	}

	v.filtersContainer = container.New(&schemaFiltersAdaptiveColumnsLayout{
		// Layout reports effective column count on each pass so the view can
		// update last-row behavior while resizing.
		onColumnsChanged: v.updateLastRowFlagsForColumns,
	}, objects...)
	v.rootContainer.Objects = []fyne.CanvasObject{v.filtersContainer}
	v.rootContainer.Refresh()
}

// updateLastRowFlagsForColumns marks all FilterItems in the last row so they can expand freely.
//
// This is called from layout during resize so row behavior always matches the
// current column count.
func (v *SchemaFiltersView) updateLastRowFlagsForColumns(columns int) {
	if len(v.filterItemsByType) == 0 {
		return
	}
	if columns < 1 {
		columns = 1
	}
	if v.lastComputedColumns == columns {
		// Skip redundant work if width changes did not alter column count.
		return
	}
	v.lastComputedColumns = columns

	totalItems := len(v.filterItemsByType)
	rows := (totalItems + columns - 1) / columns
	if rows <= 0 {
		return
	}

	lastRowStart := (rows - 1) * columns

	// Mark items in last row
	itemIndex := 0
	for _, filterType := range FilterTypeOrder {
		item := v.filterItemsByType[filterType]
		if item == nil {
			continue
		}
		isLastRow := itemIndex >= lastRowStart
		item.SetIsLastInRow(isLastRow)
		itemIndex++
	}
}

func (v *SchemaFiltersView) handleFilterItemSelectionChanged(filterType FilterType) func([]string) {
	return func(selected []string) {
		switch filterType {
		case FilterTypeOrganizations:
			v.activeCriteria.Organizations = append([]string{}, selected...)
		case FilterTypePartitions:
			v.activeCriteria.Partitions = append([]string{}, selected...)
		case FilterTypeRegions:
			v.activeCriteria.Regions = append([]string{}, selected...)
		case FilterTypeRoles:
			v.activeCriteria.Roles = append([]string{}, selected...)
		case FilterTypeAccounts:
			v.activeCriteria.Accounts = append([]string{}, selected...)
		}

		v.recalculateAvailableOptionsAndPruneSelections()
		v.emitCurrentFilteredSchema()
	}
}

func (v *SchemaFiltersView) recalculateAvailableOptionsAndPruneSelections() {
	availableOptionsByType := v.availableOptions.GetAvailableOptions(v.activeCriteria)
	for _, filterType := range FilterTypeOrder {
		item := v.filterItemsByType[filterType]
		if item == nil {
			continue
		}
		item.SetDisplayTotalCount(len(v.originalOptionsByType[filterType]))
		item.SetAvailableOptions(availableOptionsByType[filterType])

		// Keep criteria in sync with silent pruning in SetAvailableOptions.
		selected := item.GetSelectedOptions()
		switch filterType {
		case FilterTypeOrganizations:
			v.activeCriteria.Organizations = selected
		case FilterTypePartitions:
			v.activeCriteria.Partitions = selected
		case FilterTypeRegions:
			v.activeCriteria.Regions = selected
		case FilterTypeRoles:
			v.activeCriteria.Roles = selected
		case FilterTypeAccounts:
			v.activeCriteria.Accounts = selected
		}
	}
}

func (v *SchemaFiltersView) refreshOriginalOptionsCache() {
	v.originalOptionsByType = v.availableOptions.GetAvailableOptions(schema.FilterCriteria{})
}

func (v *SchemaFiltersView) emitCurrentFilteredSchema() {
	if v.onFilteredSchemaChanged == nil {
		return
	}

	filteredSchema, err := schema.FilterSchema(v.originalSchema, v.activeCriteria)
	if err != nil {
		v.onFilteredSchemaChanged(v.originalSchema)
		return
	}
	v.onFilteredSchemaChanged(filteredSchema)
}

func (v *SchemaFiltersView) criteriaValuesForFilterType(filterType FilterType) []string {
	switch filterType {
	case FilterTypeOrganizations:
		return append([]string{}, v.activeCriteria.Organizations...)
	case FilterTypePartitions:
		return append([]string{}, v.activeCriteria.Partitions...)
	case FilterTypeRegions:
		return append([]string{}, v.activeCriteria.Regions...)
	case FilterTypeRoles:
		return append([]string{}, v.activeCriteria.Roles...)
	case FilterTypeAccounts:
		return append([]string{}, v.activeCriteria.Accounts...)
	default:
		return nil
	}
}

type schemaFiltersAdaptiveColumnsLayout struct {
	onColumnsChanged func(columns int)
	// Cached from Layout() so MinSize() can estimate for current width mode.
	lastColumns int
}

func (l *schemaFiltersAdaptiveColumnsLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	columns := schemaFiltersColumnCountForWidth(size.Width)
	if columns < 1 {
		columns = 1
	}
	l.lastColumns = columns
	if l.onColumnsChanged != nil {
		// Keep view-level row metadata synchronized with responsive columns.
		l.onColumnsChanged(columns)
	}

	rows := (len(objects) + columns - 1) / columns
	// Width available to cards after removing outer gutters and column gaps.
	usableWidth := size.Width - (2 * schemaFiltersHorizontalPadding) - (float32(columns-1) * schemaFiltersColumnGap)
	if usableWidth < 0 {
		usableWidth = 0
	}
	itemWidth := usableWidth / float32(columns)

	// Each row height is driven by the tallest card in that row so cards align.
	rowHeights := make([]float32, rows)
	for index, object := range objects {
		row := index / columns
		minHeight := object.MinSize().Height
		if minHeight > rowHeights[row] {
			rowHeights[row] = minHeight
		}
	}

	currentY := float32(0)
	for row := 0; row < rows; row++ {
		rowHeight := rowHeights[row]
		for column := 0; column < columns; column++ {
			index := (row * columns) + column
			if index >= len(objects) {
				break
			}

			object := objects[index]
			x := schemaFiltersHorizontalPadding + float32(column)*(itemWidth+schemaFiltersColumnGap)
			object.Move(fyne.NewPos(x, currentY))
			object.Resize(fyne.NewSize(itemWidth, rowHeight))
		}
		currentY += rowHeight
		if row < rows-1 {
			currentY += schemaFiltersRowGap
		}
	}
}

func (l *schemaFiltersAdaptiveColumnsLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	// MinSize is computed using the last responsive column count determined by
	// Layout(). This keeps outer scroll extents aligned with current width mode.
	columns := l.lastColumns
	if columns < 1 {
		columns = 1
	}
	if columns > len(objects) {
		columns = len(objects)
	}

	rows := (len(objects) + columns - 1) / columns
	// Mirror row-height logic from Layout() so MinSize matches arranged content.
	rowHeights := make([]float32, rows)
	for index, object := range objects {
		row := index / columns
		minHeight := object.MinSize().Height
		if minHeight > rowHeights[row] {
			rowHeights[row] = minHeight
		}
	}

	totalHeight := float32(0)
	for _, rowHeight := range rowHeights {
		totalHeight += rowHeight
	}
	if rows > 1 {
		totalHeight += float32(rows-1) * schemaFiltersRowGap
	}

	// Minimum width for the current column mode, including gaps and gutters.
	minWidth := (float32(columns) * schemaFiltersMinColumnWidth) +
		(float32(columns-1) * schemaFiltersColumnGap) +
		(2 * schemaFiltersHorizontalPadding)

	return fyne.NewSize(minWidth, totalHeight)
}

func schemaFiltersColumnCountForWidth(width float32) int {
	switch {
	case width < schemaFiltersWidthBreakpointOneColumn:
		return 1
	case width < schemaFiltersWidthBreakpointTwoColumns:
		return 2
	case width < schemaFiltersWidthBreakpointThreeColumns:
		return 3
	case width < schemaFiltersWidthBreakpointFourColumns:
		return 4
	default:
		return 5
	}
}
