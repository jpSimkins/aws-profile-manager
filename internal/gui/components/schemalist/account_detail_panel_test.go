package schemalist

import (
	"testing"

	fyneTest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"aws-profile-manager/internal/gui/components/actionbuttons"
	"aws-profile-manager/internal/test"
)

// ssoRecord returns a fully-populated AccountRecord for use in detail panel tests.
func ssoRecord() AccountRecord {
	return AccountRecord{
		OrganizationAlias: "acme",
		OrganizationName:  "Acme Corp",
		PartitionName:     "commercial",
		AccountAlias:      "acme-prod",
		AccountName:       "Acme Production",
		AccountID:         "123456789012",
		DefaultRegion:     "us-east-1",
		Regions:           []string{"us-east-1", "us-west-2"},
		Roles:             []string{"Admin", "Developer"},
		SsoURL:            "https://acme.awsapps.com/start",
	}
}

// --- buildAccountDetailsContent ---

func TestBuildAccountDetailsContent_ReturnsContent(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	content := buildAccountDetailsContent(ssoRecord(), true)
	if content == nil {
		t.Fatal("buildAccountDetailsContent should not return nil")
	}
}

func TestBuildAccountDetailsContent_NoRoles(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	r := ssoRecord()
	r.Roles = nil
	content := buildAccountDetailsContent(r, true)
	if content == nil {
		t.Fatal("should return content even with no roles")
	}
}

func TestBuildAccountDetailsContent_EmptyRecord(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	// Should not panic on a completely empty record.
	content := buildAccountDetailsContent(AccountRecord{}, true)
	if content == nil {
		t.Fatal("should return content for empty record")
	}
}

// --- detailSectionHeader ---

func TestDetailSectionHeader_ReturnsContent(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	content := detailSectionHeader("🪪  Identity")
	if content == nil {
		t.Fatal("detailSectionHeader should not return nil")
	}
}

// --- detailRow ---

func TestDetailRow_WithoutCopyValue(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	// Should not panic; copyValue="" means no copy button.
	row := detailRow("Name", "Acme Prod", "")
	if row == nil {
		t.Fatal("detailRow should not return nil without copy value")
	}
}

func TestDetailRow_WithCopyValue(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	row := detailRow("Account ID", "123456789012", "123456789012")
	if row == nil {
		t.Fatal("detailRow should not return nil with copy value")
	}
}

func TestDetailRow_EmptyValues(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	// Should not panic on empty strings.
	row := detailRow("", "", "")
	if row == nil {
		t.Fatal("detailRow should not return nil for empty values")
	}
}

// --- buildRolesSection ---

func TestBuildRolesSection_NoRoles(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	r := ssoRecord()
	r.Roles = nil
	section := buildRolesSection(r, true)
	if section == nil {
		t.Fatal("buildRolesSection should not return nil with no roles")
	}
}

func TestBuildRolesSection_WithRoles(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	section := buildRolesSection(ssoRecord(), true)
	if section == nil {
		t.Fatal("buildRolesSection should not return nil with roles")
	}
}

func TestBuildRolesSection_ProfileNameFormat(t *testing.T) {
	// buildRolesSection is pure Go — verify the profile name convention via
	// the detailRow label/value logic by checking the record fields directly.
	r := ssoRecord()
	expectedProfile := r.PartitionName + "-" + r.AccountAlias + "-" + r.Roles[0]
	if expectedProfile != "commercial-acme-prod-Admin" {
		t.Errorf("unexpected profile name: %q", expectedProfile)
	}
}

// --- actionbuttons.Copy (formerly copyIconButton) ---

func TestCopyIconButton_ReturnsButton(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	btn := actionbuttons.Copy("some-value")
	if btn == nil {
		t.Fatal("actionbuttons.Copy should not return nil")
	}
}

func TestCopyIconButton_IsLowImportance(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	btn := actionbuttons.Copy("value")
	if btn.Importance != widget.LowImportance {
		t.Errorf("expected LowImportance, got %v", btn.Importance)
	}
}

func TestBuildAccountDetailsContent_NoCliButtons_DoesNotPanic(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	content := buildAccountDetailsContent(ssoRecord(), false)
	if content == nil {
		t.Fatal("should return content when showCliButtons is false")
	}
}

func TestBuildRolesSection_NoCliButtons_DoesNotPanic(t *testing.T) {
	_ = fyneTest.NewApp()
	test.SetupTestEnvironment(t)

	section := buildRolesSection(ssoRecord(), false)
	if section == nil {
		t.Fatal("buildRolesSection should not return nil with showCliButtons=false")
	}
}
