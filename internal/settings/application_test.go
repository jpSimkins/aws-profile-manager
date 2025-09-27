package settings

import "testing"

func TestGetDefaultApplication(t *testing.T) {
	app := GetDefaultApplication()

	if app.ManagedSectionStart == "" {
		t.Error("ManagedSectionStart should not be empty")
	}
	if app.ManagedSectionEnd == "" {
		t.Error("ManagedSectionEnd should not be empty")
	}
	if !app.IncludeTimestamp {
		t.Error("IncludeTimestamp should be true by default")
	}
	if !app.IncludeVersion {
		t.Error("IncludeVersion should be true by default")
	}
}

func TestGetFormattedStartMarker(t *testing.T) {
	app := ApplicationSettings{
		ManagedSectionStart: "START MARKER",
	}

	formatted := app.GetFormattedStartMarker()
	expected := "# START MARKER"

	if formatted != expected {
		t.Errorf("Expected %q, got %q", expected, formatted)
	}
}

func TestGetFormattedEndMarker(t *testing.T) {
	app := ApplicationSettings{
		ManagedSectionEnd: "END MARKER",
	}

	formatted := app.GetFormattedEndMarker()
	expected := "# END MARKER"

	if formatted != expected {
		t.Errorf("Expected %q, got %q", expected, formatted)
	}
}

func TestApplicationValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings ApplicationSettings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaultApplication(),
			wantErr:  false,
		},
		{
			name: "empty start marker",
			settings: ApplicationSettings{
				ManagedSectionStart: "",
				ManagedSectionEnd:   "END",
			},
			wantErr: true,
		},
		{
			name: "empty end marker",
			settings: ApplicationSettings{
				ManagedSectionStart: "START",
				ManagedSectionEnd:   "",
			},
			wantErr: true,
		},
		{
			name: "identical markers",
			settings: ApplicationSettings{
				ManagedSectionStart: "SAME",
				ManagedSectionEnd:   "SAME",
			},
			wantErr: true,
		},
		{
			name: "valid custom markers",
			settings: ApplicationSettings{
				ManagedSectionStart: "BEGIN MANAGED",
				ManagedSectionEnd:   "END MANAGED",
				IncludeTimestamp:    true,
				IncludeVersion:      true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.settings.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplicationGetSchema(t *testing.T) {
	app := GetDefaultApplication()
	schema := app.GetSchema()

	if schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", schema.Version)
	}

	expectedFields := []string{
		"managed_section_start",
		"managed_section_end",
		"include_timestamp",
		"include_version",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Verify field properties
	startField := schema.Fields["managed_section_start"]
	if startField.Type != "string" {
		t.Errorf("Expected type string, got %s", startField.Type)
	}
	if !startField.Required {
		t.Error("managed_section_start should be required")
	}

	timestampField := schema.Fields["include_timestamp"]
	if timestampField.Type != "bool" {
		t.Errorf("Expected type bool, got %s", timestampField.Type)
	}
}
