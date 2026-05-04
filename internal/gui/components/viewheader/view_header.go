// Package viewheader provides a standardized header component for views and tabs.
//
// ViewHeader gives every tab a consistent look: a markdown title, a descriptive
// paragraph, an optional info widget (e.g. LoadInfo), and action buttons — all
// in a single reusable component.
package viewheader

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// ViewHeader is a standardized header for views and tabs.
//
// Layout:
//
//	[Title (markdown)]              [info?] [btn1] [btn2]
//	[Description — word-wrapped paragraph]
//
// Usage:
//
//	loadInfo := loadinfo.NewLoadInfo(window)
//	header := viewheader.New(
//	    "# Install Profiles",
//	    "Install AWS profiles from centralized configuration.",
//	).WithInfo(loadInfo.GetContent()).WithButtons(reloadBtn, selectBtn)
//
//	return container.NewBorder(header.GetContent(), nil, nil, nil, content)
type ViewHeader struct {
	title       string            // Markdown string, e.g. "# Install Profiles"
	description string            // Plain text shown as a wrapped paragraph
	infoWidget  fyne.CanvasObject // Optional — e.g. a LoadInfo button
	buttons     []*widget.Button  // Optional action buttons shown right of title
}

// New creates a ViewHeader with the given title and description.
//
// title should be a markdown string (e.g. "# My View").
// description is plain text displayed below the title as a word-wrapped paragraph.
func New(title, description string) *ViewHeader {
	return &ViewHeader{
		title:       title,
		description: description,
	}
}

// WithInfo adds an optional info widget to the right side of the title row.
//
// Typically a loadinfo.LoadInfo button. Returns the ViewHeader for chaining.
func (h *ViewHeader) WithInfo(info fyne.CanvasObject) *ViewHeader {
	h.infoWidget = info
	return h
}

// WithButtons appends action buttons to the right side of the title row.
//
// Returns the ViewHeader for chaining.
func (h *ViewHeader) WithButtons(buttons ...*widget.Button) *ViewHeader {
	h.buttons = append(h.buttons, buttons...)
	return h
}

// GetContent builds and returns the header as a canvas object.
//
// Call this once and embed the result in a border layout.
func (h *ViewHeader) GetContent() fyne.CanvasObject {
	// Build title row: title on left, optional info + buttons on right.
	titleItems := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown(h.title),
		layout.NewSpacer(),
	}
	if h.infoWidget != nil {
		titleItems = append(titleItems, h.infoWidget)
	}
	for _, btn := range h.buttons {
		titleItems = append(titleItems, btn)
	}

	desc := widget.NewLabel(h.description)
	desc.Wrapping = fyne.TextWrapWord

	return container.NewPadded(container.NewVBox(
		container.NewHBox(titleItems...),
		desc,
	))
}
