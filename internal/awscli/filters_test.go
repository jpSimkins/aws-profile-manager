package awscli

import (
	"testing"
)

func TestNewFilter(t *testing.T) {
	filter := NewFilter()
	if filter == nil {
		t.Fatal("NewFilter() returned nil")
	}
}

func TestFilterProfiles_NoFilters(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", AccountID: "123456789012", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "session1", Properties: map[string]string{}},
		{Name: "profile2", AccountID: "123456789013", RoleName: "ReadOnlyRole", Region: "us-west-2", SsoSession: "session2", Properties: map[string]string{}},
		{Name: "profile3", AccountID: "123456789014", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "session1", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{} // No filters

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 3 {
		t.Errorf("Expected 3 profiles with no filters, got %d", len(result))
	}

	// Verify results are sorted by name
	expectedOrder := []string{"profile1", "profile2", "profile3"}
	for i, expected := range expectedOrder {
		if i >= len(result) || result[i].Name != expected {
			t.Errorf("Expected profile %s at index %d, got %s", expected, i, result[i].Name)
		}
	}
}

func TestFilterProfiles_ByAccountID(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", AccountID: "123456789012", Properties: map[string]string{}},
		{Name: "profile2", AccountID: "123456789013", Properties: map[string]string{}},
		{Name: "profile3", AccountID: "123456789012", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{AccountIDs: []string{"123456789012"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles with account ID filter, got %d", len(result))
	}

	for _, profile := range result {
		if profile.AccountID != "123456789012" {
			t.Errorf("Filtered profile should have account ID 123456789012, got %s", profile.AccountID)
		}
	}
}

func TestFilterProfiles_ByMultipleAccountIDs(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", AccountID: "123456789012", Properties: map[string]string{}},
		{Name: "profile2", AccountID: "123456789013", Properties: map[string]string{}},
		{Name: "profile3", AccountID: "123456789014", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{AccountIDs: []string{"123456789012", "123456789014"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles with multiple account ID filter, got %d", len(result))
	}

	validAccountIDs := map[string]bool{"123456789012": true, "123456789014": true}
	for _, profile := range result {
		if !validAccountIDs[profile.AccountID] {
			t.Errorf("Filtered profile has unexpected account ID: %s", profile.AccountID)
		}
	}
}

func TestFilterProfiles_ByRoleName(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", RoleName: "AdminRole", Properties: map[string]string{}},
		{Name: "profile2", RoleName: "ReadOnlyRole", Properties: map[string]string{}},
		{Name: "profile3", RoleName: "AdminRole", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{RoleNames: []string{"AdminRole"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles with role name filter, got %d", len(result))
	}

	for _, profile := range result {
		if profile.RoleName != "AdminRole" {
			t.Errorf("Filtered profile should have role name AdminRole, got %s", profile.RoleName)
		}
	}
}

func TestFilterProfiles_ByRegion(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", Region: "us-east-1", Properties: map[string]string{}},
		{Name: "profile2", Region: "us-west-2", Properties: map[string]string{}},
		{Name: "profile3", Region: "us-east-1", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{Regions: []string{"us-east-1"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles with region filter, got %d", len(result))
	}

	for _, profile := range result {
		if profile.Region != "us-east-1" {
			t.Errorf("Filtered profile should have region us-east-1, got %s", profile.Region)
		}
	}
}

func TestFilterProfiles_BySsoSession(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", SsoSession: "session1", Properties: map[string]string{}},
		{Name: "profile2", SsoSession: "session2", Properties: map[string]string{}},
		{Name: "profile3", SsoSession: "session1", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{SsoSessions: []string{"session1"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 profiles with SSO session filter, got %d", len(result))
	}

	for _, profile := range result {
		if profile.SsoSession != "session1" {
			t.Errorf("Filtered profile should have SSO session session1, got %s", profile.SsoSession)
		}
	}
}

func TestFilterProfiles_ByNamePattern(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "prod-admin-east", Properties: map[string]string{}},
		{Name: "prod-readonly-west", Properties: map[string]string{}},
		{Name: "dev-admin-east", Properties: map[string]string{}},
		{Name: "test-something", Properties: map[string]string{}},
	}

	filter := NewFilter()

	tests := []struct {
		pattern  string
		expected int
		desc     string
	}{
		{"prod-.*", 2, "prod prefix pattern"},
		{".*-admin-.*", 2, "admin in middle pattern"},
		{".*east", 2, "east suffix pattern"},
		{"test.*", 1, "test prefix pattern"},
		{"nonexistent.*", 0, "no match pattern"},
		{"", 4, "empty pattern matches all"},
		{".*", 4, "wildcard pattern matches all"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			criteria := FilterCriteria{NamePattern: tt.pattern}
			result := filter.FilterProfiles(profiles, criteria)
			if len(result) != tt.expected {
				t.Errorf("Pattern %q: expected %d matches, got %d", tt.pattern, tt.expected, len(result))
			}
		})
	}
}

func TestFilterProfiles_InvalidRegexPattern(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{NamePattern: "[invalid"}

	result := filter.FilterProfiles(profiles, criteria)

	// Invalid regex should result in no matches
	if len(result) != 0 {
		t.Errorf("Invalid regex pattern should return no matches, got %d", len(result))
	}
}

func TestFilterProfiles_MultipleFilters(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "prod-admin", AccountID: "123456789012", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "prod-session", Properties: map[string]string{}},
		{Name: "prod-readonly", AccountID: "123456789012", RoleName: "ReadOnlyRole", Region: "us-east-1", SsoSession: "prod-session", Properties: map[string]string{}},
		{Name: "dev-admin", AccountID: "123456789013", RoleName: "AdminRole", Region: "us-west-2", SsoSession: "dev-session", Properties: map[string]string{}},
		{Name: "test-admin", AccountID: "123456789014", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "test-session", Properties: map[string]string{}},
	}

	filter := NewFilter()

	// Filter by account ID AND role name (should use AND logic)
	criteria := FilterCriteria{
		AccountIDs: []string{"123456789012"},
		RoleNames:  []string{"AdminRole"},
	}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 1 {
		t.Errorf("Expected 1 profile with multiple filters (AND logic), got %d", len(result))
	}

	if len(result) > 0 {
		profile := result[0]
		if profile.Name != "prod-admin" {
			t.Errorf("Expected filtered profile to be prod-admin, got %s", profile.Name)
		}
		if profile.AccountID != "123456789012" {
			t.Errorf("Filtered profile should have account ID 123456789012, got %s", profile.AccountID)
		}
		if profile.RoleName != "AdminRole" {
			t.Errorf("Filtered profile should have role name AdminRole, got %s", profile.RoleName)
		}
	}
}

func TestFilterProfiles_NoMatches(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", AccountID: "123456789012", Properties: map[string]string{}},
		{Name: "profile2", AccountID: "123456789013", Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{AccountIDs: []string{"999999999999"}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 0 {
		t.Errorf("Expected 0 profiles with no matching filter, got %d", len(result))
	}
}

func TestFilterProfiles_EmptyInput(t *testing.T) {
	filter := NewFilter()
	criteria := FilterCriteria{AccountIDs: []string{"123456789012"}}

	result := filter.FilterProfiles([]AwsCliProfile{}, criteria)

	if len(result) != 0 {
		t.Errorf("Expected 0 profiles from empty input, got %d", len(result))
	}
}

func TestGetAvailableFilterOptions(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "profile1", AccountID: "123456789012", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "session1", Properties: map[string]string{}},
		{Name: "profile2", AccountID: "123456789013", RoleName: "ReadOnlyRole", Region: "us-west-2", SsoSession: "session2", Properties: map[string]string{}},
		{Name: "profile3", AccountID: "123456789012", RoleName: "AdminRole", Region: "us-east-1", SsoSession: "session1", Properties: map[string]string{}}, // Duplicates
	}

	sessions := []SsoSession{
		{Name: "session1", Properties: map[string]string{}},
		{Name: "session2", Properties: map[string]string{}},
		{Name: "unused-session", Properties: map[string]string{}}, // Should still be included
	}

	filter := NewFilter()
	options := filter.GetAvailableFilterOptions(profiles, sessions)

	// Check account IDs (should be unique and sorted)
	expectedAccountIDs := []string{"123456789012", "123456789013"}
	if len(options.AccountIDs) != len(expectedAccountIDs) {
		t.Errorf("Expected %d unique account IDs, got %d", len(expectedAccountIDs), len(options.AccountIDs))
	}
	for i, expected := range expectedAccountIDs {
		if i >= len(options.AccountIDs) || options.AccountIDs[i] != expected {
			t.Errorf("Expected account ID %s at index %d, got %s", expected, i, options.AccountIDs[i])
		}
	}

	// Check role names (should be unique and sorted)
	expectedRoleNames := []string{"AdminRole", "ReadOnlyRole"}
	if len(options.RoleNames) != len(expectedRoleNames) {
		t.Errorf("Expected %d unique role names, got %d", len(expectedRoleNames), len(options.RoleNames))
	}
	for i, expected := range expectedRoleNames {
		if i >= len(options.RoleNames) || options.RoleNames[i] != expected {
			t.Errorf("Expected role name %s at index %d, got %s", expected, i, options.RoleNames[i])
		}
	}

	// Check regions (should be unique and sorted)
	expectedRegions := []string{"us-east-1", "us-west-2"}
	if len(options.Regions) != len(expectedRegions) {
		t.Errorf("Expected %d unique regions, got %d", len(expectedRegions), len(options.Regions))
	}
	for i, expected := range expectedRegions {
		if i >= len(options.Regions) || options.Regions[i] != expected {
			t.Errorf("Expected region %s at index %d, got %s", expected, i, options.Regions[i])
		}
	}

	// Check SSO sessions (should include all sessions, sorted)
	expectedSessions := []string{"session1", "session2", "unused-session"}
	if len(options.SsoSessions) != len(expectedSessions) {
		t.Errorf("Expected %d SSO sessions, got %d", len(expectedSessions), len(options.SsoSessions))
	}
	for i, expected := range expectedSessions {
		if i >= len(options.SsoSessions) || options.SsoSessions[i] != expected {
			t.Errorf("Expected SSO session %s at index %d, got %s", expected, i, options.SsoSessions[i])
		}
	}
}

func TestGetAvailableFilterOptions_EmptyInput(t *testing.T) {
	filter := NewFilter()
	options := filter.GetAvailableFilterOptions([]AwsCliProfile{}, []SsoSession{})

	if len(options.AccountIDs) != 0 {
		t.Errorf("Expected 0 account IDs from empty input, got %d", len(options.AccountIDs))
	}

	if len(options.RoleNames) != 0 {
		t.Errorf("Expected 0 role names from empty input, got %d", len(options.RoleNames))
	}

	if len(options.Regions) != 0 {
		t.Errorf("Expected 0 regions from empty input, got %d", len(options.Regions))
	}

	if len(options.SsoSessions) != 0 {
		t.Errorf("Expected 0 SSO sessions from empty input, got %d", len(options.SsoSessions))
	}
}

func TestFilterByAccountIDs(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterByAccountIDs("123456789012", "123456789013")

	expected := []string{"123456789012", "123456789013"}
	if len(criteria.AccountIDs) != len(expected) {
		t.Errorf("Expected %d account IDs, got %d", len(expected), len(criteria.AccountIDs))
	}

	for i, exp := range expected {
		if criteria.AccountIDs[i] != exp {
			t.Errorf("Expected account ID %s at index %d, got %s", exp, i, criteria.AccountIDs[i])
		}
	}

	// Other fields should be empty
	if len(criteria.RoleNames) != 0 || len(criteria.Regions) != 0 || len(criteria.SsoSessions) != 0 || criteria.NamePattern != "" {
		t.Error("FilterByAccountIDs should only set AccountIDs field")
	}
}

func TestFilterByRoleNames(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterByRoleNames("AdminRole", "ReadOnlyRole")

	expected := []string{"AdminRole", "ReadOnlyRole"}
	if len(criteria.RoleNames) != len(expected) {
		t.Errorf("Expected %d role names, got %d", len(expected), len(criteria.RoleNames))
	}

	for i, exp := range expected {
		if criteria.RoleNames[i] != exp {
			t.Errorf("Expected role name %s at index %d, got %s", exp, i, criteria.RoleNames[i])
		}
	}

	// Other fields should be empty
	if len(criteria.AccountIDs) != 0 || len(criteria.Regions) != 0 || len(criteria.SsoSessions) != 0 || criteria.NamePattern != "" {
		t.Error("FilterByRoleNames should only set RoleNames field")
	}
}

func TestFilterByRegions(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterByRegions("us-east-1", "us-west-2")

	expected := []string{"us-east-1", "us-west-2"}
	if len(criteria.Regions) != len(expected) {
		t.Errorf("Expected %d regions, got %d", len(expected), len(criteria.Regions))
	}

	for i, exp := range expected {
		if criteria.Regions[i] != exp {
			t.Errorf("Expected region %s at index %d, got %s", exp, i, criteria.Regions[i])
		}
	}

	// Other fields should be empty
	if len(criteria.AccountIDs) != 0 || len(criteria.RoleNames) != 0 || len(criteria.SsoSessions) != 0 || criteria.NamePattern != "" {
		t.Error("FilterByRegions should only set Regions field")
	}
}

func TestFilterBySsoSessions(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterBySsoSessions("session1", "session2")

	expected := []string{"session1", "session2"}
	if len(criteria.SsoSessions) != len(expected) {
		t.Errorf("Expected %d SSO sessions, got %d", len(expected), len(criteria.SsoSessions))
	}

	for i, exp := range expected {
		if criteria.SsoSessions[i] != exp {
			t.Errorf("Expected SSO session %s at index %d, got %s", exp, i, criteria.SsoSessions[i])
		}
	}

	// Other fields should be empty
	if len(criteria.AccountIDs) != 0 || len(criteria.RoleNames) != 0 || len(criteria.Regions) != 0 || criteria.NamePattern != "" {
		t.Error("FilterBySsoSessions should only set SsoSessions field")
	}
}

func TestFilterByNamePattern(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterByNamePattern("prod-.*")

	if criteria.NamePattern != "prod-.*" {
		t.Errorf("Expected name pattern 'prod-.*', got %s", criteria.NamePattern)
	}

	// Other fields should be empty
	if len(criteria.AccountIDs) != 0 || len(criteria.RoleNames) != 0 || len(criteria.Regions) != 0 || len(criteria.SsoSessions) != 0 {
		t.Error("FilterByNamePattern should only set NamePattern field")
	}
}

func TestCombineFilters(t *testing.T) {
	filter := NewFilter()

	criteria1 := FilterCriteria{
		AccountIDs: []string{"123456789012", "123456789013"},
		RoleNames:  []string{"AdminRole"},
	}

	criteria2 := FilterCriteria{
		AccountIDs:  []string{"123456789013", "123456789014"}, // Some overlap
		Regions:     []string{"us-east-1"},
		NamePattern: "first-pattern",
	}

	criteria3 := FilterCriteria{
		SsoSessions: []string{"session1"},
		NamePattern: "second-pattern", // Should override first pattern
	}

	combined := filter.CombineFilters(criteria1, criteria2, criteria3)

	// Check account IDs (should be deduplicated and sorted)
	expectedAccountIDs := []string{"123456789012", "123456789013", "123456789014"}
	if len(combined.AccountIDs) != len(expectedAccountIDs) {
		t.Errorf("Expected %d combined account IDs, got %d", len(expectedAccountIDs), len(combined.AccountIDs))
	}
	for i, expected := range expectedAccountIDs {
		if combined.AccountIDs[i] != expected {
			t.Errorf("Expected account ID %s at index %d, got %s", expected, i, combined.AccountIDs[i])
		}
	}

	// Check role names
	expectedRoleNames := []string{"AdminRole"}
	if len(combined.RoleNames) != len(expectedRoleNames) {
		t.Errorf("Expected %d combined role names, got %d", len(expectedRoleNames), len(combined.RoleNames))
	}

	// Check regions
	expectedRegions := []string{"us-east-1"}
	if len(combined.Regions) != len(expectedRegions) {
		t.Errorf("Expected %d combined regions, got %d", len(expectedRegions), len(combined.Regions))
	}

	// Check SSO sessions
	expectedSessions := []string{"session1"}
	if len(combined.SsoSessions) != len(expectedSessions) {
		t.Errorf("Expected %d combined SSO sessions, got %d", len(expectedSessions), len(combined.SsoSessions))
	}

	// Check name pattern (should use the last non-empty one)
	if combined.NamePattern != "second-pattern" {
		t.Errorf("Expected name pattern 'second-pattern', got %s", combined.NamePattern)
	}
}

func TestCombineFilters_EmptyInput(t *testing.T) {
	filter := NewFilter()
	combined := filter.CombineFilters()

	if len(combined.AccountIDs) != 0 || len(combined.RoleNames) != 0 ||
		len(combined.Regions) != 0 || len(combined.SsoSessions) != 0 ||
		combined.NamePattern != "" {
		t.Error("CombineFilters with no input should return empty criteria")
	}
}

func TestValidateFilterCriteria(t *testing.T) {
	filter := NewFilter()

	tests := []struct {
		name      string
		criteria  FilterCriteria
		expectErr bool
	}{
		{
			name:     "Valid empty criteria",
			criteria: FilterCriteria{},
		},
		{
			name: "Valid criteria with all fields",
			criteria: FilterCriteria{
				AccountIDs:  []string{"123456789012"},
				RoleNames:   []string{"AdminRole"},
				Regions:     []string{"us-east-1"},
				SsoSessions: []string{"session1"},
				NamePattern: "prod-.*",
			},
		},
		{
			name: "Valid regex pattern",
			criteria: FilterCriteria{
				NamePattern: "^prod-[a-z]+-[0-9]+$",
			},
		},
		{
			name: "Invalid regex pattern",
			criteria: FilterCriteria{
				NamePattern: "[invalid",
			},
			expectErr: true,
		},
		{
			name: "Another invalid regex",
			criteria: FilterCriteria{
				NamePattern: "*invalid",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := filter.ValidateFilterCriteria(tt.criteria)

			if tt.expectErr && err == nil {
				t.Error("ValidateFilterCriteria() should have failed")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("ValidateFilterCriteria() should not have failed: %v", err)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	filter := NewFilter()

	slice := []string{"apple", "banana", "cherry"}

	tests := []struct {
		target   string
		expected bool
	}{
		{"apple", true},
		{"banana", true},
		{"cherry", true},
		{"grape", false},
		{"", false},      // Empty string should return false
		{"APPLE", false}, // Case sensitive
		{"app", false},   // Partial match should return false
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := filter.containsString(slice, tt.target)
			if result != tt.expected {
				t.Errorf("containsString(%v, %q): expected %v, got %v", slice, tt.target, tt.expected, result)
			}
		})
	}
}

func TestMatchesFilters(t *testing.T) {
	filter := NewFilter()

	profile := AwsCliProfile{
		Name:       "test-profile",
		AccountID:  "123456789012",
		RoleName:   "AdminRole",
		Region:     "us-east-1",
		SsoSession: "test-session",
		Properties: map[string]string{},
	}

	tests := []struct {
		name     string
		criteria FilterCriteria
		expected bool
	}{
		{
			name:     "Empty criteria (should match)",
			criteria: FilterCriteria{},
			expected: true,
		},
		{
			name: "Matching account ID",
			criteria: FilterCriteria{
				AccountIDs: []string{"123456789012"},
			},
			expected: true,
		},
		{
			name: "Non-matching account ID",
			criteria: FilterCriteria{
				AccountIDs: []string{"999999999999"},
			},
			expected: false,
		},
		{
			name: "Matching role name",
			criteria: FilterCriteria{
				RoleNames: []string{"AdminRole"},
			},
			expected: true,
		},
		{
			name: "Non-matching role name",
			criteria: FilterCriteria{
				RoleNames: []string{"ReadOnlyRole"},
			},
			expected: false,
		},
		{
			name: "Matching region",
			criteria: FilterCriteria{
				Regions: []string{"us-east-1"},
			},
			expected: true,
		},
		{
			name: "Non-matching region",
			criteria: FilterCriteria{
				Regions: []string{"us-west-2"},
			},
			expected: false,
		},
		{
			name: "Matching SSO session",
			criteria: FilterCriteria{
				SsoSessions: []string{"test-session"},
			},
			expected: true,
		},
		{
			name: "Non-matching SSO session",
			criteria: FilterCriteria{
				SsoSessions: []string{"other-session"},
			},
			expected: false,
		},
		{
			name: "Matching name pattern",
			criteria: FilterCriteria{
				NamePattern: "test-.*",
			},
			expected: true,
		},
		{
			name: "Non-matching name pattern",
			criteria: FilterCriteria{
				NamePattern: "prod-.*",
			},
			expected: false,
		},
		{
			name: "Invalid regex pattern",
			criteria: FilterCriteria{
				NamePattern: "[invalid",
			},
			expected: false,
		},
		{
			name: "Multiple criteria (all match)",
			criteria: FilterCriteria{
				AccountIDs: []string{"123456789012"},
				RoleNames:  []string{"AdminRole"},
				Regions:    []string{"us-east-1"},
			},
			expected: true,
		},
		{
			name: "Multiple criteria (one doesn't match)",
			criteria: FilterCriteria{
				AccountIDs: []string{"123456789012"},
				RoleNames:  []string{"ReadOnlyRole"}, // This doesn't match
				Regions:    []string{"us-east-1"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.matchesFilters(profile, tt.criteria)
			if result != tt.expected {
				t.Errorf("matchesFilters: expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRemoveDuplicatesAndSort(t *testing.T) {
	filter := NewFilter()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"apple", "banana", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "With duplicates",
			input:    []string{"banana", "apple", "cherry", "apple", "banana"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "With empty strings (should be filtered out)",
			input:    []string{"apple", "", "banana", "", "cherry"},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "All empty strings",
			input:    []string{"", "", ""},
			expected: []string{},
		},
		{
			name:     "Unsorted input",
			input:    []string{"zebra", "apple", "banana"},
			expected: []string{"apple", "banana", "zebra"},
		},
		{
			name:     "Single item",
			input:    []string{"apple"},
			expected: []string{"apple"},
		},
		{
			name:     "Duplicates with empty strings",
			input:    []string{"apple", "", "apple", "", "banana"},
			expected: []string{"apple", "banana"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.removeDuplicatesAndSort(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestMapKeysToSortedSlice(t *testing.T) {
	filter := NewFilter()

	tests := []struct {
		name     string
		input    map[string]bool
		expected []string
	}{
		{
			name: "Normal map",
			input: map[string]bool{
				"banana": true,
				"apple":  true,
				"cherry": true,
			},
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "Empty map",
			input:    map[string]bool{},
			expected: []string{},
		},
		{
			name: "Single item",
			input: map[string]bool{
				"apple": true,
			},
			expected: []string{"apple"},
		},
		{
			name: "Mixed boolean values (shouldn't matter)",
			input: map[string]bool{
				"banana": false,
				"apple":  true,
				"cherry": false,
			},
			expected: []string{"apple", "banana", "cherry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.mapKeysToSortedSlice(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestFilterProfiles_ByProfileType(t *testing.T) {
	profiles := []AwsCliProfile{
		{Name: "sso-profile1", Type: ProfileTypeSSO, AccountID: "123456789012", Properties: map[string]string{}},
		{Name: "iam-profile1", Type: ProfileTypeIAM, HasAccessKey: true, Properties: map[string]string{}},
		{Name: "sso-profile2", Type: ProfileTypeSSO, AccountID: "123456789013", Properties: map[string]string{}},
		{Name: "assume-role", Type: ProfileTypeAssumeRole, Properties: map[string]string{}},
	}

	filter := NewFilter()
	criteria := FilterCriteria{ProfileTypes: []ProfileType{ProfileTypeSSO}}

	result := filter.FilterProfiles(profiles, criteria)

	if len(result) != 2 {
		t.Errorf("Expected 2 SSO profiles, got %d", len(result))
	}

	for _, profile := range result {
		if profile.Type != ProfileTypeSSO {
			t.Errorf("Expected only SSO profiles, got %s", profile.Type)
		}
	}
}

func TestFilterByProfileTypes(t *testing.T) {
	filter := NewFilter()
	criteria := filter.FilterByProfileTypes(ProfileTypeSSO, ProfileTypeIAM)

	if len(criteria.ProfileTypes) != 2 {
		t.Errorf("Expected 2 profile types, got %d", len(criteria.ProfileTypes))
	}

	if criteria.ProfileTypes[0] != ProfileTypeSSO {
		t.Errorf("Expected first type to be SSO, got %s", criteria.ProfileTypes[0])
	}

	if criteria.ProfileTypes[1] != ProfileTypeIAM {
		t.Errorf("Expected second type to be IAM, got %s", criteria.ProfileTypes[1])
	}
}

func TestContainsProfileType(t *testing.T) {
	filter := NewFilter()

	types := []ProfileType{ProfileTypeSSO, ProfileTypeIAM}

	if !filter.containsProfileType(types, ProfileTypeSSO) {
		t.Error("Expected to find SSO type")
	}

	if !filter.containsProfileType(types, ProfileTypeIAM) {
		t.Error("Expected to find IAM type")
	}

	if filter.containsProfileType(types, ProfileTypeAssumeRole) {
		t.Error("Expected NOT to find AssumeRole type")
	}
}

func TestRemoveDuplicateProfileTypes(t *testing.T) {
	filter := NewFilter()

	input := []ProfileType{ProfileTypeSSO, ProfileTypeIAM, ProfileTypeSSO, ProfileTypeIAM, ProfileTypeSSO}
	result := filter.removeDuplicateProfileTypes(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 unique types, got %d", len(result))
	}

	// Verify both types are present
	foundSSO := false
	foundIAM := false
	for _, pt := range result {
		if pt == ProfileTypeSSO {
			foundSSO = true
		}
		if pt == ProfileTypeIAM {
			foundIAM = true
		}
	}

	if !foundSSO {
		t.Error("Expected SSO type in result")
	}

	if !foundIAM {
		t.Error("Expected IAM type in result")
	}
}
