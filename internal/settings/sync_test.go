package settings

import "testing"

func TestGetDefaultSync(t *testing.T) {
	sync := GetDefaultSync()

	if sync.Enabled {
		t.Error("Sync should be disabled by default")
	}
	if sync.AutoUpdate {
		t.Error("AutoUpdate should be disabled by default")
	}
	if sync.UpdateOnRead {
		t.Error("UpdateOnRead should be disabled by default")
	}
	if sync.Strategy != "local" {
		t.Errorf("Expected default strategy local, got %s", sync.Strategy)
	}
}

func TestSyncValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings SyncSettings
		wantErr  bool
	}{
		{
			name:     "valid defaults",
			settings: GetDefaultSync(),
			wantErr:  false,
		},
		{
			name: "valid local config",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "local",
				Local: LocalSettings{
					Path: "/path/to/config.json",
				},
			},
			wantErr: false,
		},
		{
			name: "valid S3 config",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "s3",
				S3: S3Settings{
					Bucket: "my-bucket",
					Key:    "config.json",
					Region: "us-west-2",
				},
			},
			wantErr: false,
		},
		{
			name: "valid HTTP config",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "http",
				HTTP: HTTPSettings{
					URL: "https://example.com/config.json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid strategy",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "invalid",
			},
			wantErr: true,
		},
		{
			name: "local strategy missing path",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "local",
				Local:    LocalSettings{}, // Empty path
			},
			wantErr: true,
		},
		{
			name: "S3 strategy missing bucket",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "s3",
				S3: S3Settings{
					Key:    "config.json",
					Region: "us-west-2",
				},
			},
			wantErr: true,
		},
		{
			name: "S3 strategy missing key",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "s3",
				S3: S3Settings{
					Bucket: "my-bucket",
					Region: "us-west-2",
				},
			},
			wantErr: true,
		},
		{
			name: "HTTP strategy missing URL",
			settings: SyncSettings{
				Enabled:  true,
				Strategy: "http",
				HTTP:     HTTPSettings{}, // Empty URL
			},
			wantErr: true,
		},
		{
			name: "disabled sync with invalid config",
			settings: SyncSettings{
				Enabled:  false, // Disabled, so validation should pass
				Strategy: "local",
				Local:    LocalSettings{}, // Missing path, but sync is disabled
			},
			wantErr: false, // Should not validate strategy-specific settings when disabled
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

func TestSyncGetSchema(t *testing.T) {
	sync := GetDefaultSync()
	schema := sync.GetSchema()

	if schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", schema.Version)
	}

	expectedFields := []string{
		"enabled",
		"auto_update",
		"update_on_read",
		"strategy",
		"local",
		"s3",
		"http",
		"git",
	}

	for _, field := range expectedFields {
		if _, exists := schema.Fields[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Verify strategy field has enum
	strategyField := schema.Fields["strategy"]
	if len(strategyField.Enum) == 0 {
		t.Error("Strategy field should have enum values")
	}

	expectedStrategies := []string{"local", "s3", "http", "https", "git"}
	enumMap := make(map[string]bool)
	for _, strategy := range strategyField.Enum {
		enumMap[strategy] = true
	}
	for _, expected := range expectedStrategies {
		if !enumMap[expected] {
			t.Errorf("Expected strategy %s not found in enum", expected)
		}
	}

	// Verify nested schemas exist and have dependencies
	testCases := []struct {
		fieldName string
		depValue  interface{}
	}{
		{"local", "local"},
		{"s3", "s3"},
		{"http", []interface{}{"http", "https"}},
		{"git", "git"},
	}

	for _, tc := range testCases {
		field := schema.Fields[tc.fieldName]

		if field.Type != "object" {
			t.Errorf("Field %s should be type object, got %s", tc.fieldName, field.Type)
		}

		if field.DependsOn == nil {
			t.Errorf("Field %s should have DependsOn", tc.fieldName)
			continue
		}

		if field.DependsOn.Field != "strategy" {
			t.Errorf("Field %s DependsOn should be 'strategy', got %s", tc.fieldName, field.DependsOn.Field)
		}

		if field.Nested == nil {
			t.Errorf("Field %s should have nested schema", tc.fieldName)
		}
	}

	// Verify nested fields exist
	localField := schema.Fields["local"]
	if localField.Nested != nil {
		if _, exists := localField.Nested.Fields["path"]; !exists {
			t.Error("Local nested schema should have 'path' field")
		}
	}

	s3Field := schema.Fields["s3"]
	if s3Field.Nested != nil {
		expectedS3Fields := []string{"bucket", "key", "region", "profile", "use_sso", "use_iam", "public_read"}
		for _, field := range expectedS3Fields {
			if _, exists := s3Field.Nested.Fields[field]; !exists {
				t.Errorf("S3 nested schema should have '%s' field", field)
			}
		}
	}

	httpField := schema.Fields["http"]
	if httpField.Nested != nil {
		expectedHTTPFields := []string{"url", "basic_auth", "username"}
		for _, field := range expectedHTTPFields {
			if _, exists := httpField.Nested.Fields[field]; !exists {
				t.Errorf("HTTP nested schema should have '%s' field", field)
			}
		}
	}

	gitField := schema.Fields["git"]
	if gitField.Nested != nil {
		expectedGitFields := []string{"repo", "branch", "config_path", "use_ssh", "private"}
		for _, field := range expectedGitFields {
			if _, exists := gitField.Nested.Fields[field]; !exists {
				t.Errorf("Git nested schema should have '%s' field", field)
			}
		}
	}
}
