package profiles

import (
	"strings"
	"testing"

	"aws-profile-manager/internal/core"
	"aws-profile-manager/internal/schema"
	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/test"
)

// TestDetectMarkers tests marker detection in various scenarios
func TestDetectMarkers(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		startMarker string
		endMarker   string
		wantFound   bool
		wantStart   int
		wantEnd     int
	}{
		{
			name: "markers found",
			lines: []string{
				"[profile personal]",
				"# START - Managed",
				"[profile work]",
				"# END - Managed",
				"[profile other]",
			},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   true,
			wantStart:   1,
			wantEnd:     3,
		},
		{
			name: "no markers",
			lines: []string{
				"[profile personal]",
				"[profile other]",
			},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   false,
			wantStart:   -1,
			wantEnd:     -1,
		},
		{
			name: "only start marker",
			lines: []string{
				"[profile personal]",
				"# START - Managed",
				"[profile work]",
			},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   false,
			wantStart:   1,
			wantEnd:     -1,
		},
		{
			name: "only end marker",
			lines: []string{
				"[profile personal]",
				"[profile work]",
				"# END - Managed",
			},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   false,
			wantStart:   -1,
			wantEnd:     2,
		},
		{
			name: "markers with extra whitespace",
			lines: []string{
				"[profile personal]",
				"  # START - Managed  ",
				"[profile work]",
				"	# END - Managed	",
				"[profile other]",
			},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   true,
			wantStart:   1,
			wantEnd:     3,
		},
		{
			name:        "empty file",
			lines:       []string{},
			startMarker: "START - Managed",
			endMarker:   "END - Managed",
			wantFound:   false,
			wantStart:   -1,
			wantEnd:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectMarkers(tt.lines, tt.startMarker, tt.endMarker)

			if result.Found != tt.wantFound {
				t.Errorf("Found = %v, want %v", result.Found, tt.wantFound)
			}
			if result.StartLine != tt.wantStart {
				t.Errorf("StartLine = %d, want %d", result.StartLine, tt.wantStart)
			}
			if result.EndLine != tt.wantEnd {
				t.Errorf("EndLine = %d, want %d", result.EndLine, tt.wantEnd)
			}
		})
	}
}

// TestCalculateProfileCounts tests profile counting
func TestCalculateProfileCounts(t *testing.T) {
	tests := []struct {
		name            string
		schema          *schema.Schema
		wantProfiles    int
		wantSsoSessions int
	}{
		{
			name:            "single SSO org",
			schema:          schematest.NewManagedSsoSingle(),
			wantProfiles:    1, // Generator creates 1 profile (1 account * 1 role, default region only)
			wantSsoSessions: 1,
		},
		{
			name:            "multi-account SSO",
			schema:          schematest.NewManagedSsoMultiAccount(),
			wantProfiles:    12, // Generator creates 12 profiles (actual generated count)
			wantSsoSessions: 1,
		},
		{
			name:            "all profile types",
			schema:          schematest.NewManagedAll(),
			wantProfiles:    4, // Generator creates 4 profiles (SSO + IAM + AssumeRole + Generic)
			wantSsoSessions: 1,
		},
		{
			name:            "IAM only",
			schema:          schematest.NewManagedIamSingle(),
			wantProfiles:    1,
			wantSsoSessions: 0,
		},
		{
			name:            "empty schema",
			schema:          schematest.NewEmpty(),
			wantProfiles:    0,
			wantSsoSessions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiles, sessions := calculateProfileCounts(tt.schema)

			if profiles != tt.wantProfiles {
				t.Errorf("profiles = %d, want %d", profiles, tt.wantProfiles)
			}
			if sessions != tt.wantSsoSessions {
				t.Errorf("sessions = %d, want %d", sessions, tt.wantSsoSessions)
			}
		})
	}
}

// TestCountProfiles tests counting profiles in ProfileCollection
func TestCountProfiles(t *testing.T) {
	tests := []struct {
		name       string
		collection *schema.ProfileCollection
		want       int
	}{
		{
			name:       "single SSO org",
			collection: schematest.NewManagedSsoSingle().Managed,
			want:       1, // Generator creates 1 profile
		},
		{
			name:       "mixed types",
			collection: schematest.NewManagedAll().Managed,
			want:       4, // Generator creates 4 profiles (SSO + IAM + AssumeRole + Generic)
		},
		{
			name:       "nil collection",
			collection: nil,
			want:       0,
		},
		{
			name:       "empty collection",
			collection: &schema.ProfileCollection{},
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countProfiles(tt.collection)
			if got != tt.want {
				t.Errorf("countProfiles() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCountSsoSessions tests SSO session counting
func TestCountSsoSessions(t *testing.T) {
	tests := []struct {
		name       string
		collection *schema.ProfileCollection
		want       int
	}{
		{
			name:       "single SSO org",
			collection: schematest.NewManagedSsoSingle().Managed,
			want:       1,
		},
		{
			name:       "multi-org SSO",
			collection: schematest.NewManagedSsoMultiOrg().Managed,
			want:       3, // 2 orgs with multiple partitions
		},
		{
			name:       "no SSO profiles",
			collection: schematest.NewManagedIamSingle().Managed,
			want:       0,
		},
		{
			name:       "nil collection",
			collection: nil,
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSsoSessions(tt.collection)
			if got != tt.want {
				t.Errorf("countSsoSessions() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestHasProfiles tests profile existence check
func TestHasProfiles(t *testing.T) {
	tests := []struct {
		name       string
		collection *schema.ProfileCollection
		want       bool
	}{
		{
			name:       "has SSO profiles",
			collection: schematest.NewManagedSsoSingle().Managed,
			want:       true,
		},
		{
			name:       "has IAM profiles",
			collection: schematest.NewManagedIamSingle().Managed,
			want:       true,
		},
		{
			name:       "empty collection",
			collection: &schema.ProfileCollection{},
			want:       false,
		},
		{
			name:       "nil collection",
			collection: nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasProfiles(tt.collection)
			if got != tt.want {
				t.Errorf("hasProfiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidateSchema tests schema validation
func TestValidateSchema(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name    string
		schema  *schema.Schema
		wantErr bool
	}{
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name:    "empty schema",
			schema:  &schema.Schema{},
			wantErr: true,
		},
		{
			name:    "valid managed only",
			schema:  schematest.NewManagedSsoSingle(),
			wantErr: false,
		},
		{
			name:    "unmanaged only - should fail",
			schema:  schematest.NewUnmanagedSsoSingle(),
			wantErr: true, // validateSchema requires managed section
		},
		{
			name:    "valid mixed",
			schema:  schematest.NewMixedSimple(),
			wantErr: false,
		},
		{
			name:    "empty managed section",
			schema:  schematest.NewManagedEmpty(),
			wantErr: true, // Empty managed section is invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchema(tt.schema)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAddManagedSection tests managed section building
func TestAddManagedSection(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name           string
		existingLines  []string
		managedContent string
		startMarker    string
		endMarker      string
		wantContains   []string
		wantNotEmpty   bool
	}{
		{
			name:           "empty config",
			existingLines:  []string{},
			managedContent: "[profile work-dev]\nregion = us-west-2",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantContains: []string{
				"# START - Test",
				"[profile work-dev]",
				"# END - Test",
			},
			wantNotEmpty: true,
		},
		{
			name: "existing personal profiles",
			existingLines: []string{
				"[profile personal-above]",
				"region = us-east-1",
			},
			managedContent: "[profile work-dev]\nregion = us-west-2",
			startMarker:    "# START - Test",
			endMarker:      "# END - Test",
			wantContains: []string{
				"[profile personal-above]",
				"# START - Test",
				"[profile work-dev]",
				"# END - Test",
			},
			wantNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addManagedSection(tt.existingLines, tt.managedContent, tt.startMarker, tt.endMarker, Config{})

			if tt.wantNotEmpty && len(result) == 0 {
				t.Error("Expected non-empty result")
			}

			// Check for expected strings
			for _, want := range tt.wantContains {
				found := false
				for _, line := range result {
					if line == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected line %q not found in result", want)
				}
			}
		})
	}
}

// TestBuildFinalConfig tests final config assembly
func TestBuildFinalConfig(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name           string
		existing       []string
		managedContent string
		markers        markerPosition
		startMarker    string
		endMarker      string
		wantContains   []string
	}{
		{
			name: "replace existing managed section",
			existing: []string{
				"[profile above]",
				"# START - Old",
				"[profile old-work]",
				"# END - Old",
				"[profile below]",
			},
			managedContent: "[profile new-work]",
			markers: markerPosition{
				Found:     true,
				StartLine: 1,
				EndLine:   3,
			},
			startMarker: "# START - New",
			endMarker:   "# END - New",
			wantContains: []string{
				"[profile above]",
				"# START - New",
				"[profile new-work]",
				"# END - New",
				"[profile below]",
			},
		},
		{
			name: "no existing markers",
			existing: []string{
				"[profile above]",
				"[profile personal]",
			},
			managedContent: "[profile work]",
			markers: markerPosition{
				Found:     false,
				StartLine: -1,
				EndLine:   -1,
			},
			startMarker: "# START - Test",
			endMarker:   "# END - Test",
			wantContains: []string{
				"[profile above]",
				"[profile personal]",
				"# START - Test",
				"[profile work]",
				"# END - Test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFinalConfig(tt.existing, tt.managedContent, tt.markers, tt.startMarker, tt.endMarker, Config{})

			for _, want := range tt.wantContains {
				found := false
				for _, line := range result {
					if line == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected line %q not found in result", want)
				}
			}
		})
	}
}

// TestAddManagedSection_IncludeVersion verifies that the version line is written
// into the managed-section comment block when IncludeVersion is true, and is
// absent when the flag is false.
func TestAddManagedSection_IncludeVersion(t *testing.T) {
	test.SetupTestEnvironment(t)

	const startMarker = "# START - Test"
	const endMarker = "# END - Test"
	const managedContent = "[profile work-dev]\nregion = us-east-1"
	wantVersion := "# Generated by AWS Profile Manager v" + core.AppVersion

	t.Run("version line present when IncludeVersion=true", func(t *testing.T) {
		cfg := Config{IncludeVersion: true}
		result := addManagedSection([]string{}, managedContent, startMarker, endMarker, cfg)
		output := strings.Join(result, "\n")
		if !strings.Contains(output, wantVersion) {
			t.Errorf("expected version line %q in output:\n%s", wantVersion, output)
		}
	})

	t.Run("version line absent when IncludeVersion=false", func(t *testing.T) {
		cfg := Config{IncludeVersion: false}
		result := addManagedSection([]string{}, managedContent, startMarker, endMarker, cfg)
		output := strings.Join(result, "\n")
		if strings.Contains(output, wantVersion) {
			t.Errorf("did not expect version line in output when disabled:\n%s", output)
		}
	})
}

// TestAddManagedSection_IncludeTimestamp verifies that a timestamp comment line
// is written when IncludeTimestamp is true and omitted when false.
func TestAddManagedSection_IncludeTimestamp(t *testing.T) {
	test.SetupTestEnvironment(t)

	const startMarker = "# START - Test"
	const endMarker = "# END - Test"
	const managedContent = "[profile work-dev]\nregion = us-east-1"
	const timestampPrefix = "# Generated at "

	t.Run("timestamp line present when IncludeTimestamp=true", func(t *testing.T) {
		cfg := Config{IncludeTimestamp: true}
		result := addManagedSection([]string{}, managedContent, startMarker, endMarker, cfg)
		output := strings.Join(result, "\n")
		if !strings.Contains(output, timestampPrefix) {
			t.Errorf("expected timestamp line starting with %q in output:\n%s", timestampPrefix, output)
		}
	})

	t.Run("timestamp line absent when IncludeTimestamp=false", func(t *testing.T) {
		cfg := Config{IncludeTimestamp: false}
		result := addManagedSection([]string{}, managedContent, startMarker, endMarker, cfg)
		output := strings.Join(result, "\n")
		if strings.Contains(output, timestampPrefix) {
			t.Errorf("did not expect timestamp line in output when disabled:\n%s", output)
		}
	})
}

// TestAddManagedSection_BothMetadataFlags exercises all four combinations of
// the IncludeVersion and IncludeTimestamp flags to prevent regression.
func TestAddManagedSection_BothMetadataFlags(t *testing.T) {
	test.SetupTestEnvironment(t)

	const startMarker = "# START - Test"
	const endMarker = "# END - Test"
	const managedContent = "[profile work]\nregion = us-east-1"
	versionPrefix := "# Generated by AWS Profile Manager v"
	timestampPrefix := "# Generated at "

	tests := []struct {
		name             string
		includeVersion   bool
		includeTimestamp bool
		wantVersion      bool
		wantTimestamp    bool
	}{
		{"both off", false, false, false, false},
		{"version only", true, false, true, false},
		{"timestamp only", false, true, false, true},
		{"both on", true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{IncludeVersion: tt.includeVersion, IncludeTimestamp: tt.includeTimestamp}
			result := addManagedSection([]string{}, managedContent, startMarker, endMarker, cfg)
			output := strings.Join(result, "\n")

			hasVersion := strings.Contains(output, versionPrefix)
			hasTimestamp := strings.Contains(output, timestampPrefix)

			if hasVersion != tt.wantVersion {
				t.Errorf("version line present=%v, want %v\noutput:\n%s", hasVersion, tt.wantVersion, output)
			}
			if hasTimestamp != tt.wantTimestamp {
				t.Errorf("timestamp line present=%v, want %v\noutput:\n%s", hasTimestamp, tt.wantTimestamp, output)
			}
		})
	}
}
