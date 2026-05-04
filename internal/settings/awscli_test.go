package settings

import "testing"

func TestGetDefaultAwsCLI(t *testing.T) {
	awscli := GetDefaultAwsCLI()

	if !awscli.AutoRefresh {
		t.Error("AutoRefresh should be enabled by default")
	}
	if awscli.RefreshIntervalMins != 5 {
		t.Errorf("Expected refresh interval 5 mins, got %d", awscli.RefreshIntervalMins)
	}
	if !awscli.ShowSsoSessions {
		t.Error("ShowSsoSessions should be enabled by default")
	}
}

func TestAwsCliValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings AwsCliSettings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaultAwsCLI(),
			wantErr:  false,
		},
		{
			name: "valid custom interval",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: 30,
			},
			wantErr: false,
		},
		{
			name: "minimum interval (1 min)",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: 1,
			},
			wantErr: false,
		},
		{
			name: "maximum interval (1440 mins / 24 hours)",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: 1440,
			},
			wantErr: false,
		},
		{
			name: "interval too low",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: 0,
			},
			wantErr: true,
		},
		{
			name: "interval too high",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: 1441,
			},
			wantErr: true,
		},
		{
			name: "negative interval",
			settings: AwsCliSettings{
				AutoRefresh:         true,
				RefreshIntervalMins: -5,
			},
			wantErr: true,
		},
		{
			name: "auto refresh disabled with any interval",
			settings: AwsCliSettings{
				AutoRefresh:         false,
				RefreshIntervalMins: 15,
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

func TestAwsCliGetSchema(t *testing.T) {
	awscli := GetDefaultAwsCLI()
	schema := awscli.GetSchema()

	if schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", schema.Version)
	}

	expectedFields := []string{
		"auto_refresh",
		"refresh_interval_mins",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Verify auto_refresh field
	autoRefreshField := schema.Fields["auto_refresh"]
	if autoRefreshField.Type != "bool" {
		t.Errorf("Expected type bool, got %s", autoRefreshField.Type)
	}
	if !autoRefreshField.Required {
		t.Error("auto_refresh should be required")
	}

	// Verify refresh_interval_mins has constraints
	intervalField := schema.Fields["refresh_interval_mins"]
	if intervalField.Type != "int" {
		t.Errorf("Expected type int, got %s", intervalField.Type)
	}
	if intervalField.Min == nil {
		t.Error("refresh_interval_mins should have min constraint")
	}
	if intervalField.Max == nil {
		t.Error("refresh_interval_mins should have max constraint")
	}
	if *intervalField.Min != 1.0 {
		t.Errorf("Expected min 1, got %f", *intervalField.Min)
	}
	if *intervalField.Max != 1440.0 {
		t.Errorf("Expected max 1440, got %f", *intervalField.Max)
	}
}
