package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/bundled"
	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/gui/components"
	"aws-profile-manager/internal/logging"
)

// ShowAboutDialog displays the About dialog.
//
// Entry point for the about workflow. Delegates content building to helpers,
// then constructs and shows the dialog.
func ShowAboutDialog(window fyne.Window, footer *components.Footer) {
	logging.Log.Info("About dialog triggered")

	if footer != nil {
		footer.SetStatus("Opening about...")
	}

	versionInfo := core.GetVersion()
	detailsText, copyDetails := buildAboutVersionText(versionInfo)
	aboutContent := buildAboutContent(window, footer, detailsText, copyDetails)

	var aboutDialog *dialog.CustomDialog
	aboutDialog = components.ShowCustomDialog(components.DialogOptions{
		Title:   "About",
		Content: aboutContent,
		Buttons: []components.DialogButton{
			{
				Label: "Close",
				OnTapped: func() {
					aboutDialog.Hide()
					if footer != nil {
						footer.SetStatus("Exited About")
					}
				},
				Importance: widget.MediumImportance,
			},
		},
		Window:      window,
		Scrollable:  false,
		UseSettings: false,
	})

	aboutDialog.Show()

	if footer != nil {
		footer.SetStatus("About dialog shown")
	}
}

// buildAboutVersionText constructs the markdown details text and plain-text copy string
// from version information.
//
// Returns:
//   - detailsText: Markdown formatted for display
//   - copyDetails: Plain text suitable for clipboard copy
func buildAboutVersionText(versionInfo core.Info) (detailsText, copyDetails string) {
	frameworkDisplay := versionInfo.Framework
	if versionInfo.FrameworkVersion != "unknown" {
		frameworkDisplay = fmt.Sprintf("%s %s", versionInfo.Framework, versionInfo.FrameworkVersion)
	}

	detailsText = fmt.Sprintf("**Version:** %s\n\n**Platform:** %s\n\n**Go Version:** %s\n\n**Framework:** %s",
		versionInfo.Version,
		versionInfo.Platform,
		versionInfo.GoVersion,
		frameworkDisplay,
	)

	copyDetails = fmt.Sprintf(
		"App: %s\nVersion: `%s`\nPlatform: `%s`\nGo Version: `%s`\nFramework: `%s`",
		core.AppName,
		versionInfo.Version,
		versionInfo.Platform,
		versionInfo.GoVersion,
		frameworkDisplay,
	)

	if versionInfo.Commit != "" {
		commitLen := len(versionInfo.Commit)
		if commitLen > 8 {
			commitLen = 8
		}
		detailsText += fmt.Sprintf("\n\n**Commit:** %s", versionInfo.Commit[:commitLen])
		copyDetails += fmt.Sprintf("\nCommit: `%s`", versionInfo.Commit[:commitLen])
	}
	if versionInfo.Date != "" {
		detailsText += fmt.Sprintf("\n\n**Built:** %s", versionInfo.Date)
		copyDetails += fmt.Sprintf("\nBuilt: %s", versionInfo.Date)
	}

	return detailsText, copyDetails
}

// buildAboutContent assembles the full dialog content: branding row, version details, and footer links.
func buildAboutContent(window fyne.Window, footer *components.Footer, detailsText, copyDetails string) fyne.CanvasObject {
	footerText := fmt.Sprintf("\n---\n\n**Author:** %s\n\n**Repository:** [%s](%s)\n\nA tool for managing AWS CLI profiles with SSO support.\n",
		core.AppAuthor,
		core.AppURL,
		core.AppURL,
	)

	copyButton := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		fyne.CurrentApp().Clipboard().SetContent(copyDetails)
		dialog.ShowInformation("Copied!", "Debug info copied to clipboard.", window)
		if footer != nil {
			footer.SetStatus("Debug info copied")
		}
	})
	copyButton.Importance = widget.LowImportance

	logo := canvas.NewImageFromResource(getAboutLogoResourceForTheme())
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(71, 50))
	title := widget.NewRichTextFromMarkdown("# " + core.AppName)

	return container.NewVBox(
		container.NewCenter(container.NewHBox(logo, title)), // Branding
		container.NewBorder(nil, nil, nil, copyButton, widget.NewRichTextFromMarkdown(detailsText)),
		widget.NewRichTextFromMarkdown(footerText),
	)
}

// getAboutLogoResourceForTheme returns the same theme-aware logo used in the header.
func getAboutLogoResourceForTheme() fyne.Resource {
	bgColor := theme.Color(theme.ColorNameBackground)
	r, g, b, _ := bgColor.RGBA()
	brightness := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)

	if brightness < 128 {
		return bundled.ResourceLogoDarkMode
	}

	return bundled.ResourceLogo
}
