package layouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// NewMaxWidthCentered wraps content so it never exceeds maxWidth pixels wide
// and is horizontally centred within the available space.
//
// This is useful for settings-style pages where stretching inputs to the full
// window width looks poor on wide displays.
//
// Parameters:
//   - content: The canvas object to constrain.
//   - maxWidth: Maximum width in pixels. Ignored if the container is narrower.
//
// Returns:
//   - fyne.CanvasObject: Wrapped content with max-width centering applied.
func NewMaxWidthCentered(content fyne.CanvasObject, maxWidth float32) fyne.CanvasObject {
	return container.New(&maxWidthCenteredLayout{maxWidth: maxWidth}, content)
}

type maxWidthCenteredLayout struct {
	maxWidth float32
}

func (l *maxWidthCenteredLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, obj := range objects {
		w := containerSize.Width
		if l.maxWidth > 0 && w > l.maxWidth {
			w = l.maxWidth
		}
		x := (containerSize.Width - w) / 2
		obj.Move(fyne.NewPos(x, 0))
		obj.Resize(fyne.NewSize(w, containerSize.Height))
	}
}

func (l *maxWidthCenteredLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, obj := range objects {
		s := obj.MinSize()
		if s.Width > min.Width {
			min.Width = s.Width
		}
		if s.Height > min.Height {
			min.Height = s.Height
		}
	}
	return min
}
