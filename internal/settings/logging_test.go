package settings

import "testing"

func TestGetDefaultLogging(t *testing.T) {
	logging := GetDefaultLogging()

	if logging.LogLevel != "warn" {
		t.Errorf("Expected default log level warn, got %s", logging.LogLevel)
	}
	if logging.EnableDebug {
		t.Error("Debug should be disabled by default")
	}
}

func TestLoggingValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings LoggingSettings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaultLogging(),
			wantErr:  false,
		},
		{
			name: "valid debug level",
			settings: LoggingSettings{
				LogLevel:    "debug",
				EnableDebug: true,
			},
			wantErr: false,
		},
		{
			name: "valid warn level",
			settings: LoggingSettings{
				LogLevel:    "warn",
				EnableDebug: false,
			},
			wantErr: false,
		},
		{
			name: "valid error level",
			settings: LoggingSettings{
				LogLevel:    "error",
				EnableDebug: false,
			},
			wantErr: false,
		},
		{
			name: "valid silent level",
			settings: LoggingSettings{
				LogLevel:    "silent",
				EnableDebug: false,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			settings: LoggingSettings{
				LogLevel:    "invalid",
				EnableDebug: false,
			},
			wantErr: true,
		},
		{
			name: "empty log level",
			settings: LoggingSettings{
				LogLevel:    "",
				EnableDebug: false,
			},
			wantErr: true,
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

func TestLoggingGetSchema(t *testing.T) {
	logging := GetDefaultLogging()
	schema := logging.GetSchema()

	if schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", schema.Version)
	}

	expectedFields := []string{"log_level", "enable_debug"}
	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Verify log_level field
	logLevelField := schema.Fields["log_level"]
	if logLevelField.Type != "string" {
		t.Errorf("Expected type string, got %s", logLevelField.Type)
	}
	if !logLevelField.Required {
		t.Error("log_level should be required")
	}
	expectedEnum := []string{"debug", "info", "warn", "error", "silent"}
	if len(logLevelField.Enum) != len(expectedEnum) {
		t.Errorf("Expected %d enum values, got %d", len(expectedEnum), len(logLevelField.Enum))
	}

	// Verify enable_debug field
	debugField := schema.Fields["enable_debug"]
	if debugField.Type != "bool" {
		t.Errorf("Expected type bool, got %s", debugField.Type)
	}
}
