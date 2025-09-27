package layouts

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewHorizontalSplitWithLeftRightPaneMinWidths_ConfiguresOffset(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	split := NewHorizontalSplitWithLeftRightPaneMinWidths(
		widget.NewLabel("left"),
		widget.NewLabel("right"),
		260,
		320,
		0.35,
	)

	if split == nil {
		t.Fatal("split should not be nil")
	}
	if split.Offset != 0.35 {
		t.Fatalf("expected split offset 0.35, got %f", split.Offset)
	}
}

func TestNewHorizontalSplitWithLeftRightPaneMinWidths_AppliesPaneMinWidths(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	split := NewHorizontalSplitWithLeftRightPaneMinWidths(
		widget.NewLabel("left"),
		widget.NewLabel("right"),
		260,
		320,
		0.35,
	)

	if split.Leading.MinSize().Width < 260 {
		t.Fatalf("expected left pane min width >= 260, got %f", split.Leading.MinSize().Width)
	}
	if split.Trailing.MinSize().Width < 320 {
		t.Fatalf("expected right pane min width >= 320, got %f", split.Trailing.MinSize().Width)
	}
}
