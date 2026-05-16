package layouts

import (
	"testing"

	"fyne.io/fyne/v2"
	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestNewMaxWidthCentered_ReturnsNonNil(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("test")
	result := NewMaxWidthCentered(label, 600)
	if result == nil {
		t.Fatal("NewMaxWidthCentered should not return nil")
	}
}

func TestMaxWidthCenteredLayout_ConstrainsWidthWhenContainerIsWider(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("content")
	l := &maxWidthCenteredLayout{maxWidth: 400}

	objects := []fyne.CanvasObject{label}
	containerSize := fyne.NewSize(800, 200)

	l.Layout(objects, containerSize)

	if label.Size().Width != 400 {
		t.Fatalf("expected child width 400, got %f", label.Size().Width)
	}
}

func TestMaxWidthCenteredLayout_CentresChildHorizontally(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("content")
	l := &maxWidthCenteredLayout{maxWidth: 400}

	objects := []fyne.CanvasObject{label}
	containerSize := fyne.NewSize(800, 200)

	l.Layout(objects, containerSize)

	expectedX := float32((800 - 400) / 2) // 200
	if label.Position().X != expectedX {
		t.Fatalf("expected child X position %f, got %f", expectedX, label.Position().X)
	}
}

func TestMaxWidthCenteredLayout_DoesNotConstrainWhenContainerIsNarrower(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("content")
	l := &maxWidthCenteredLayout{maxWidth: 800}

	objects := []fyne.CanvasObject{label}
	containerSize := fyne.NewSize(400, 200)

	l.Layout(objects, containerSize)

	if label.Size().Width != 400 {
		t.Fatalf("expected child width 400 (full container width), got %f", label.Size().Width)
	}
	if label.Position().X != 0 {
		t.Fatalf("expected child X position 0 when not constrained, got %f", label.Position().X)
	}
}

func TestMaxWidthCenteredLayout_SetsFullHeight(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("content")
	l := &maxWidthCenteredLayout{maxWidth: 400}

	objects := []fyne.CanvasObject{label}
	containerSize := fyne.NewSize(800, 300)

	l.Layout(objects, containerSize)

	if label.Size().Height != 300 {
		t.Fatalf("expected child height 300, got %f", label.Size().Height)
	}
}

func TestMaxWidthCenteredLayout_MinSize_ReturnsChildMinSize(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	label := widget.NewLabel("some content")
	l := &maxWidthCenteredLayout{maxWidth: 800}

	objects := []fyne.CanvasObject{label}
	minSize := l.MinSize(objects)

	childMin := label.MinSize()
	if minSize.Width != childMin.Width {
		t.Fatalf("expected min width %f, got %f", childMin.Width, minSize.Width)
	}
	if minSize.Height != childMin.Height {
		t.Fatalf("expected min height %f, got %f", childMin.Height, minSize.Height)
	}
}

func TestMaxWidthCenteredLayout_HandlesEmptyObjects(t *testing.T) {
	testApp := fyneTest.NewApp()
	defer testApp.Quit()

	l := &maxWidthCenteredLayout{maxWidth: 400}

	// Should not panic with empty objects
	l.Layout(nil, fyne.NewSize(800, 200))
	minSize := l.MinSize(nil)
	if minSize.Width != 0 || minSize.Height != 0 {
		t.Fatalf("expected zero min size for empty objects, got %v", minSize)
	}
}
