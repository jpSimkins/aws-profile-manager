// Package layouts provides reusable GUI layout primitives.
package layouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// NewHorizontalSplitWithLeftRightPaneMinWidths creates a horizontal split view with
// explicit minimum widths for both panes.
//
// This helper isolates split-resize behavior from view/business logic. It
// constrains how far the divider can be moved by wrapping each pane in a
// layout that reports the desired minimum width.
//
// Parameters:
//   - leftPane: Content displayed on the left side
//   - rightPane: Content displayed on the right side
//   - leftPaneMinWidth: Minimum width for the left pane
//   - rightPaneMinWidth: Minimum width for the right pane
//   - initialOffset: Initial split offset in [0.0, 1.0]
//
// Returns:
//   - *container.Split: Configured horizontal split container
func NewHorizontalSplitWithLeftRightPaneMinWidths(
	leftPane fyne.CanvasObject,
	rightPane fyne.CanvasObject,
	leftPaneMinWidth float32,
	rightPaneMinWidth float32,
	initialOffset float64,
) *container.Split {
	split := container.NewHSplit(
		wrapPaneWithMinWidth(leftPane, leftPaneMinWidth),
		wrapPaneWithMinWidth(rightPane, rightPaneMinWidth),
	)
	split.Offset = initialOffset
	return split
}

func wrapPaneWithMinWidth(content fyne.CanvasObject, minWidth float32) fyne.CanvasObject {
	return container.New(paneMinWidthLayout{minWidth: minWidth}, content)
}

type paneMinWidthLayout struct {
	minWidth float32
}

func (l paneMinWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Move(fyne.NewPos(0, 0))
	objects[0].Resize(size)
}

func (l paneMinWidthLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	if l.minWidth < 1 {
		return fyne.NewSize(1, 1)
	}
	return fyne.NewSize(l.minWidth, 1)
}
