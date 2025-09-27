package viewmodels

import (
	"strings"
	"testing"

	"aws-profile-manager/internal/generators"
	"aws-profile-manager/internal/profiles"
)

func TestPluralize(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{name: "single", count: 1, want: ""},
		{name: "zero", count: 0, want: "s"},
		{name: "multiple", count: 2, want: "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.want {
				t.Fatalf("pluralize(%d) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

func TestFormatStatsSection_EmptyReturnsNil(t *testing.T) {
	got := formatStatsSection(generators.SectionStats{}, 0, "Managed")
	if got != nil {
		t.Fatal("expected nil for zero total and zero duplicates")
	}
}

func TestFormatStatsSection_NonEmptyReturnsWidget(t *testing.T) {
	stats := generators.SectionStats{SsoProfiles: 1, SessionsWritten: 1}
	got := formatStatsSection(stats, 1, "Managed Section")
	if got == nil {
		t.Fatal("expected non-nil widget for non-zero total")
	}
}

func TestFormatStatsSection_DuplicatesOnlyReturnsWidget(t *testing.T) {
	dups := profiles.SectionDuplicateStats{TotalDuplicates: 1, IamProfiles: 1}
	got := formatStatsSection(generators.SectionStats{}, 0, "Managed", dups)
	if got == nil {
		t.Fatal("expected non-nil widget when only duplicates present")
	}
}

func TestBuildStatsSectionOrgMarkdown_IncludesExpectedFields(t *testing.T) {
	stats := generators.SectionStats{
		OrganizationCount:  2,
		PartitionCount:     3,
		AccountCount:       4,
		RoleCount:          5,
		RegionCount:        6,
		SsoProfiles:        7,
		SessionsWritten:    8,
		IamProfiles:        9,
		AssumeRoleProfiles: 10,
		GenericProfiles:    11,
	}

	got := buildStatsSectionOrgMarkdown(stats)

	expects := []string{
		"**Accounts:** 4",
		"**Organizations:** 2",
		"**Partitions:** 3",
		"**Regions:** 6",
		"**Roles:** 5",
	}

	for _, expected := range expects {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected org markdown to contain %q, got:\n%s", expected, got)
		}
	}
	if strings.Contains(got, "Managed Section") {
		t.Fatal("section title should not be included in org markdown")
	}
}

func TestBuildStatsSectionOrgMarkdown_EmptyWhenNoData(t *testing.T) {
	got := buildStatsSectionOrgMarkdown(generators.SectionStats{})
	if got != "" {
		t.Fatalf("expected empty string when all org stats are zero, got:\n%s", got)
	}
}

func TestBuildStatsSectionOrgMarkdown_SkipsZeroFields(t *testing.T) {
	stats := generators.SectionStats{
		AccountCount: 3,
		// All other fields zero
	}
	got := buildStatsSectionOrgMarkdown(stats)
	if !strings.Contains(got, "**Accounts:** 3") {
		t.Fatalf("expected Accounts line, got:\n%s", got)
	}
	for _, absent := range []string{"Organizations", "Partitions", "Regions", "Roles"} {
		if strings.Contains(got, absent) {
			t.Fatalf("expected %q to be absent when zero, got:\n%s", absent, got)
		}
	}
}

func TestBuildStatsSectionProfileMarkdown_WithoutDuplicates(t *testing.T) {
	stats := generators.SectionStats{
		SsoProfiles:        7,
		SessionsWritten:    8,
		IamProfiles:        9,
		AssumeRoleProfiles: 10,
		GenericProfiles:    11,
	}

	got := buildStatsSectionProfileMarkdown(stats, 37, profiles.SectionDuplicateStats{})

	expects := []string{
		"**SSO:** 7 profiles (8 sessions)",
		"**IAM:** 9 profiles",
		"**AssumeRole:** 10 profiles",
		"**Generic:** 11 profiles",
	}

	for _, expected := range expects {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected profile markdown to contain %q, got:\n%s", expected, got)
		}
	}
}

func TestBuildStatsSectionProfileMarkdown_WithDuplicates(t *testing.T) {
	stats := generators.SectionStats{
		SsoProfiles:        3,
		SessionsWritten:    2,
		IamProfiles:        2,
		AssumeRoleProfiles: 1,
		GenericProfiles:    4,
	}
	dups := profiles.SectionDuplicateStats{
		TotalDuplicates:    4,
		SsoProfiles:        1,
		IamProfiles:        2,
		AssumeRoleProfiles: 0,
		GenericProfiles:    1,
	}

	got := buildStatsSectionProfileMarkdown(stats, 6, dups)

	expects := []string{
		"**SSO:** 3 profiles (2 sessions) [1 duplicate]",
		"**IAM:** 2 profiles [2 duplicates]",
		"**AssumeRole:** 1 profiles",
		"**Generic:** 4 profiles [1 duplicate]",
	}

	for _, expected := range expects {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected profile markdown to contain %q, got:\n%s", expected, got)
		}
	}
}

func TestBuildStatsSectionProfileMarkdown_ShowsTypeWhenOnlyDuplicatesExist(t *testing.T) {
	dups := profiles.SectionDuplicateStats{
		TotalDuplicates: 1,
		IamProfiles:     1,
	}

	got := buildStatsSectionProfileMarkdown(generators.SectionStats{}, 0, dups)

	if !strings.Contains(got, "**IAM:** 0 profiles [1 duplicate]") {
		t.Fatalf("expected IAM duplicate-only line, got:\n%s", got)
	}
}
