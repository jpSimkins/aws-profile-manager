package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestPreset_Validate tests preset validation.
func TestPreset_Validate(t *testing.T) {
	tests := []struct {
		name      string
		schema    *Schema
		presetKey string
		preset    *Preset
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid preset with organization filter",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			presetKey: "developer",
			preset: &Preset{
				Label:         "Developer",
				Organizations: []string{"test-org"},
				Roles:         []string{"Developer"},
			},
			wantError: false,
		},
		{
			name: "Valid preset with all regions",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2", "us-east-1"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			presetKey: "all-regions",
			preset: &Preset{
				Label:      "All Regions",
				AllRegions: true,
			},
			wantError: false,
		},
		{
			name:      "Invalid - empty key",
			schema:    &Schema{Version: "1.0", Managed: &ProfileCollection{}},
			presetKey: "",
			preset: &Preset{
				Label: "Test",
			},
			wantError: true,
			errorMsg:  "preset key cannot be empty",
		},
		{
			name:      "Invalid - missing label",
			schema:    &Schema{Version: "1.0", Managed: &ProfileCollection{}},
			presetKey: "test",
			preset: &Preset{
				Label: "",
			},
			wantError: true,
			errorMsg:  "label is required",
		},
		{
			name: "Invalid - non-existent organization",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			presetKey: "bad-org",
			preset: &Preset{
				Label:         "Bad Org",
				Organizations: []string{"nonexistent"},
			},
			wantError: true,
			errorMsg:  "non-existent organization",
		},
		{
			name: "Invalid - non-existent role",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			presetKey: "bad-role",
			preset: &Preset{
				Label: "Bad Role",
				Roles: []string{"NonExistentRole"},
			},
			wantError: true,
			errorMsg:  "non-existent role",
		},
		{
			name: "Valid - no managed section (skip reference validation)",
			schema: &Schema{
				Version: "1.0",
			},
			presetKey: "test",
			preset: &Preset{
				Label:         "Test",
				Organizations: []string{"any"},
				Roles:         []string{"any"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.preset.Validate(tt.presetKey, tt.schema)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSchema_ValidatePresets tests schema-level preset validation.
func TestSchema_ValidatePresets(t *testing.T) {
	tests := []struct {
		name      string
		schema    *Schema
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid schema with presets",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer", "Admin"},
								},
							},
						},
					},
				},
				Presets: map[string]*Preset{
					"developer": {
						Label:         "Developer",
						Organizations: []string{"test-org"},
						Roles:         []string{"Developer"},
					},
					"admin": {
						Label: "Admin",
						Roles: []string{"Admin"},
					},
				},
			},
			wantError: false,
		},
		{
			name: "Schema without presets (valid)",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "Invalid preset in schema",
			schema: &Schema{
				Version: "1.0",
				Managed: &ProfileCollection{
					Organizations: map[string]*Organization{
						"test-org": {
							Name: "Test Org",
							Partitions: map[string]Partition{
								"commercial": {
									URL:           "https://test.awsapps.com/start",
									DefaultRegion: "us-west-2",
									Regions:       []string{"us-west-2"},
									Accounts: []Account{
										{Alias: "dev", Name: "Dev", ID: "123456789012"},
									},
									Roles: []string{"Developer"},
								},
							},
						},
					},
				},
				Presets: map[string]*Preset{
					"bad": {
						Label: "", // Missing label
					},
				},
			},
			wantError: true,
			errorMsg:  "label is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Validate()

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestPresets_JSONRoundTrip tests JSON serialization/deserialization of presets.
func TestPresets_JSONRoundTrip(t *testing.T) {
	original := &Schema{
		Version: "1.0",
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Org",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-west-2",
							Regions:       []string{"us-west-2"},
							Accounts: []Account{
								{Alias: "dev", Name: "Dev", ID: "123456789012"},
							},
							Roles: []string{"Developer", "Admin"},
						},
					},
				},
			},
		},
		Presets: map[string]*Preset{
			"developer": {
				Label:         "Developer",
				Description:   "Standard developer access",
				Organizations: []string{"test-org"},
				Roles:         []string{"Developer"},
			},
			"all-regions": {
				Label:      "All Regions",
				AllRegions: true,
			},
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	// Deserialize from JSON
	var restored Schema
	if err := json.Unmarshal(jsonData, &restored); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify presets were preserved
	if len(restored.Presets) != 2 {
		t.Errorf("Expected 2 presets, got %d", len(restored.Presets))
	}

	// Verify developer preset
	dev, exists := restored.Presets["developer"]
	if !exists {
		t.Fatal("Developer preset not found after round trip")
	}
	if dev.Label != "Developer" {
		t.Errorf("Expected label 'Developer', got '%s'", dev.Label)
	}
	if dev.Description != "Standard developer access" {
		t.Errorf("Expected description 'Standard developer access', got '%s'", dev.Description)
	}
	if len(dev.Organizations) != 1 || dev.Organizations[0] != "test-org" {
		t.Errorf("Expected organizations ['test-org'], got %v", dev.Organizations)
	}
	if len(dev.Roles) != 1 || dev.Roles[0] != "Developer" {
		t.Errorf("Expected roles ['Developer'], got %v", dev.Roles)
	}

	// Verify all-regions preset
	allRegions, exists := restored.Presets["all-regions"]
	if !exists {
		t.Fatal("All-regions preset not found after round trip")
	}
	if !allRegions.AllRegions {
		t.Error("Expected AllRegions to be true")
	}

	// Validate restored schema
	if err := restored.Validate(); err != nil {
		t.Errorf("Restored schema validation failed: %v", err)
	}
}

// TestPresets_BackwardCompatibility tests that schemas without presets still work.
func TestPresets_BackwardCompatibility(t *testing.T) {
	jsonWithoutPresets := `{
		"version": "1.0",
		"managed": {
			"organizations": {
				"test-org": {
					"name": "Test Org",
					"partitions": {
						"commercial": {
							"url": "https://test.awsapps.com/start",
							"default_region": "us-west-2",
							"regions": ["us-west-2"],
							"accounts": [
								{"alias": "dev", "name": "Dev", "id": "123456789012"}
							],
							"roles": ["Developer"]
						}
					}
				}
			}
		}
	}`

	var schema Schema
	if err := json.Unmarshal([]byte(jsonWithoutPresets), &schema); err != nil {
		t.Fatalf("Failed to unmarshal schema without presets: %v", err)
	}

	// Presets should be nil (not present)
	if schema.Presets != nil {
		t.Error("Expected Presets to be nil for schema without presets")
	}

	// Validation should pass
	if err := schema.Validate(); err != nil {
		t.Errorf("Schema without presets validation failed: %v", err)
	}
}
