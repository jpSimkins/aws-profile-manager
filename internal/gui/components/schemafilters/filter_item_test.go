package schemafilters

import (
	"strings"
	"testing"

	fyneTest "fyne.io/fyne/v2/test"

	"aws-profile-manager/internal/test"
)

func TestFilterItem_SetDisplayTotalCount_UpdatesCountLabel(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	item := NewFilterItem(FilterItemProps{
		Label:            "Accounts",
		AvailableOptions: []string{"a", "b"},
		AllowSearch:      true,
		AllowCheckAll:    true,
		MaxHeight:        120,
	})

	item.SetDisplayTotalCount(5)

	if item.countLabel == nil {
		t.Fatal("count label should be initialized")
	}
	if !strings.Contains(item.countLabel.Text, "( 5 total)") {
		t.Fatalf("expected count label to include overridden total, got %q", item.countLabel.Text)
	}
}

func TestFilterItem_SetIsLastInRow_AffectsHeightCapping(t *testing.T) {
	t.Skip("Last-row expansion disabled to prevent click event capture bugs")
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	item := NewFilterItem(FilterItemProps{
		Label:            "Roles",
		AvailableOptions: []string{"r1", "r2", "r3", "r4", "r5"},
		AllowSearch:      false,
		AllowCheckAll:    false,
		MaxHeight:        40,
	})

	item.SetIsLastInRow(false)
	nonLastHeight := item.listScroll.MinSize().Height
	if nonLastHeight > 40 {
		t.Fatalf("expected non-last-row height to be capped at 40, got %.2f", nonLastHeight)
	}

	item.SetIsLastInRow(true)
	lastHeight := item.listScroll.MinSize().Height
	if lastHeight <= nonLastHeight {
		t.Fatalf("expected last-row height to exceed capped height; non-last %.2f last %.2f", nonLastHeight, lastHeight)
	}
}

func TestFilterItem_CheckAllAndClear_EmitOnChange(t *testing.T) {
	test.SetupTestEnvironment(t)
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	calls := 0
	item := NewFilterItem(FilterItemProps{
		Label:            "Accounts",
		AvailableOptions: []string{"a", "b", "c"},
		AllowSearch:      true,
		AllowCheckAll:    true,
		MaxHeight:        120,
		OnChange: func(_ []string) {
			calls++
		},
	})

	if item.checkAllBtn == nil {
		t.Fatal("check all button should exist")
	}
	if item.clearBtn == nil {
		t.Fatal("clear button should exist")
	}

	item.checkAllBtn.OnTapped()
	if calls == 0 {
		t.Fatal("expected OnChange to be called by Check All")
	}

	selected := item.GetSelectedOptions()
	if len(selected) != 3 {
		t.Fatalf("expected all options selected, got %v", selected)
	}

	callsBeforeClear := calls
	item.clearBtn.OnTapped()
	if calls <= callsBeforeClear {
		t.Fatal("expected OnChange to be called by Clear")
	}

	if len(item.GetSelectedOptions()) != 0 {
		t.Fatalf("expected no selections after clear, got %v", item.GetSelectedOptions())
	}
}
