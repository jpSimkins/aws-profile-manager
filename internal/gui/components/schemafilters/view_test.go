package schemafilters

import (
	"testing"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/test"
)

func TestFilterItem_SetAvailableOptions_SilentPrune(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	called := 0
	item := NewFilterItem(FilterItemProps{
		Label:            "Accounts",
		AvailableOptions: []string{"a", "b", "c"},
		SelectedOptions:  []string{"a", "b"},
		OnChange: func(_ []string) {
			called++
		},
		AllowSearch:   true,
		AllowCheckAll: true,
		MaxHeight:     120,
	})

	item.SetAvailableOptions([]string{"a"})
	selected := item.GetSelectedOptions()

	if called != 0 {
		t.Fatalf("SetAvailableOptions should be silent, got %d calls", called)
	}
	if len(selected) != 1 || selected[0] != "a" {
		t.Fatalf("expected selection to be pruned to [a], got %v", selected)
	}
}

func TestAvailableOptionsProvider_CascadesFromCriteria(t *testing.T) {
	s := schematest.NewManagedSsoMultiOrg()
	provider := NewAvailableOptionsProvider(s)

	options := provider.GetAvailableOptions(schema.FilterCriteria{
		Organizations: []string{"prod-org"},
	})

	found := false
	for _, account := range options[FilterTypeAccounts] {
		if account == "test-dev" {
			found = true
			break
		}
	}

	if found {
		t.Fatalf("expected account test-dev to be unavailable after org filter")
	}
}

func TestSchemaFiltersView_UpdatesCriteriaFromFilterChange(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewSchemaFiltersView(schematest.NewManagedSsoMultiOrg())
	v.OnFilterChange(func(_ *schema.Schema) {})

	item := v.filterItemsByType[FilterTypeOrganizations]
	if item == nil {
		t.Fatal("organizations filter item should exist")
	}

	item.SetSelectedOptions([]string{"prod-org"})
	v.handleFilterItemSelectionChanged(FilterTypeOrganizations)([]string{"prod-org"})

	criteria := v.GetCriteria()
	if len(criteria.Organizations) != 1 || criteria.Organizations[0] != "prod-org" {
		t.Fatalf("expected criteria organizations [prod-org], got %v", criteria.Organizations)
	}
}

func TestSchemaFiltersView_ResetClearsCriteria(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()
	test.SetupTestEnvironment(t)

	v := NewSchemaFiltersView(schematest.NewManagedSsoSingle())
	v.handleFilterItemSelectionChanged(FilterTypeOrganizations)([]string{"test-org"})

	v.Reset()
	criteria := v.GetCriteria()
	if !IsEmptyCriteria(criteria) {
		t.Fatalf("expected empty criteria after reset, got %+v", criteria)
	}
}

func TestSchemaFiltersAdaptiveLayout_MinSizeTracksColumns(t *testing.T) {
	layout := &schemaFiltersAdaptiveColumnsLayout{}
	objects := []fyne.CanvasObject{
		widget.NewLabel("one"),
		widget.NewLabel("two"),
		widget.NewLabel("three"),
		widget.NewLabel("four"),
		widget.NewLabel("five"),
	}

	// Narrow width => 1 column => tallest stacked height.
	layout.Layout(objects, fyne.NewSize(400, 800))
	narrowMin := layout.MinSize(objects)

	// Wide width => more columns => fewer rows => shorter total height.
	layout.Layout(objects, fyne.NewSize(2000, 800))
	wideMin := layout.MinSize(objects)

	if wideMin.Height >= narrowMin.Height {
		t.Fatalf("expected wide min height (%f) to be less than narrow (%f)", wideMin.Height, narrowMin.Height)
	}
}
