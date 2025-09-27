package viewmodels

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/profiles"
)

// formatStatsSection returns a styled widget showing section statistics.
//
// This shared helper is used by export, import preview, and import result dialogs
// to ensure consistent presentation across all flows.
//
// An optional SectionDuplicateStats may be passed as the last argument to display
// per-type duplicate counts inline. When omitted, duplicates are not shown.
//
// The widget is split into two RichText blocks separated by a visual divider:
//   - Org-level stats (accounts, orgs, partitions, regions, roles)
//   - Profile-type breakdown with optional duplicate counts and total
//
// Returns nil when both totalProfiles and TotalDuplicates are zero so callers
// can conditionally append it to a container.
func formatStatsSection(stats generators.SectionStats, totalProfiles int, sectionTitle string, duplicates ...profiles.SectionDuplicateStats) fyne.CanvasObject {
	dups := profiles.SectionDuplicateStats{}
	if len(duplicates) > 0 {
		dups = duplicates[0]
	}

	if totalProfiles == 0 && dups.TotalDuplicates == 0 {
		return nil
	}

	orgMarkdown := buildStatsSectionOrgMarkdown(stats)
	profileMarkdown := buildStatsSectionProfileMarkdown(stats, totalProfiles, dups)

	items := []fyne.CanvasObject{
		widget.NewRichTextFromMarkdown(fmt.Sprintf("## %s", sectionTitle)),
	}
	if orgMarkdown != "" {
		items = append(items, widget.NewRichTextFromMarkdown(orgMarkdown))
	}
	if profileMarkdown != "" {
		items = append(items, widget.NewRichTextFromMarkdown(profileMarkdown))
	}

	if len(items) == 0 {
		return nil
	}

	return container.NewVBox(items...)
}

// buildStatsSectionOrgMarkdown builds the org-level stats markdown for a section.
// Only non-zero fields are included. Returns an empty string if no fields have data.
// The section title heading is handled separately by the caller.
func buildStatsSectionOrgMarkdown(stats generators.SectionStats) string {
	lines := ""
	if stats.AccountCount > 0 {
		lines += fmt.Sprintf("- **Accounts:** %d\n", stats.AccountCount)
	}
	if stats.OrganizationCount > 0 {
		lines += fmt.Sprintf("- **Organizations:** %d\n", stats.OrganizationCount)
	}
	if stats.PartitionCount > 0 {
		lines += fmt.Sprintf("- **Partitions:** %d\n", stats.PartitionCount)
	}
	if stats.RegionCount > 0 {
		lines += fmt.Sprintf("- **Regions:** %d\n", stats.RegionCount)
	}
	if stats.RoleCount > 0 {
		lines += fmt.Sprintf("- **Roles:** %d\n", stats.RoleCount)
	}
	return lines
}

// buildStatsSectionProfileMarkdown builds the profile-type breakdown markdown for a section,
// including optional inline duplicate counts.
func buildStatsSectionProfileMarkdown(stats generators.SectionStats, totalProfiles int, dups profiles.SectionDuplicateStats) string {
	text := "### Profile Types\n"

	// AssumeRole profiles
	if stats.AssumeRoleProfiles > 0 || dups.AssumeRoleProfiles > 0 {
		if dups.AssumeRoleProfiles > 0 {
			text += fmt.Sprintf("- **AssumeRole:** %d profiles [%d duplicate%s]\n",
				stats.AssumeRoleProfiles, dups.AssumeRoleProfiles, pluralize(dups.AssumeRoleProfiles))
		} else {
			text += fmt.Sprintf("- **AssumeRole:** %d profiles\n", stats.AssumeRoleProfiles)
		}
	}

	// Generic profiles
	if stats.GenericProfiles > 0 || dups.GenericProfiles > 0 {
		if dups.GenericProfiles > 0 {
			text += fmt.Sprintf("- **Generic:** %d profiles [%d duplicate%s]\n",
				stats.GenericProfiles, dups.GenericProfiles, pluralize(dups.GenericProfiles))
		} else {
			text += fmt.Sprintf("- **Generic:** %d profiles\n", stats.GenericProfiles)
		}
	}

	// IAM profiles
	if stats.IamProfiles > 0 || dups.IamProfiles > 0 {
		if dups.IamProfiles > 0 {
			text += fmt.Sprintf("- **IAM:** %d profiles [%d duplicate%s]\n",
				stats.IamProfiles, dups.IamProfiles, pluralize(dups.IamProfiles))
		} else {
			text += fmt.Sprintf("- **IAM:** %d profiles\n", stats.IamProfiles)
		}
	}

	// SSO profiles
	if stats.SsoProfiles > 0 || dups.SsoProfiles > 0 {
		if dups.SsoProfiles > 0 {
			text += fmt.Sprintf("- **SSO:** %d profiles (%d sessions) [%d duplicate%s]\n",
				stats.SsoProfiles, stats.SessionsWritten, dups.SsoProfiles, pluralize(dups.SsoProfiles))
		} else {
			text += fmt.Sprintf("- **SSO:** %d profiles (%d sessions)\n",
				stats.SsoProfiles, stats.SessionsWritten)
		}
	}

	return text
}

// appendStatsSection appends a separator and a stats section widget to items if the section
// has any content to display. It is a convenience wrapper around formatStatsSection that
// removes the nil-check boilerplate from callers.
//
// An optional SectionDuplicateStats may be passed as the last argument (same as formatStatsSection).
// Returns the (potentially extended) items slice.
func appendStatsSection(items []fyne.CanvasObject, stats generators.SectionStats, totalProfiles int, sectionTitle string, duplicates ...profiles.SectionDuplicateStats) []fyne.CanvasObject {
	if w := formatStatsSection(stats, totalProfiles, sectionTitle, duplicates...); w != nil {
		return append(items, widget.NewSeparator(), w)
	}
	return items
}

// pluralize returns "s" if count != 1, otherwise empty string.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
