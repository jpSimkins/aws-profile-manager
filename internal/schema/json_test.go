package schema

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSchema_ToJSON(t *testing.T) {
	schema := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Organization",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-east-1"},
							Accounts: []Account{
								{Alias: "dev", Name: "Development", ID: "123456789012"},
							},
							Roles: []string{"Developer"},
						},
					},
				},
			},
		},
	}

	data, err := schema.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Errorf("ToJSON() produced invalid JSON: %v", err)
	}

	// Verify version is present
	if version, ok := result["version"].(string); !ok || version != CurrentSchemaVersion {
		t.Errorf("ToJSON() version = %v, want %s", result["version"], CurrentSchemaVersion)
	}
}

func TestSchemaFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "Valid JSON",
			json: fmt.Sprintf(`{
				"version": "%s",
				"managed": {
					"organizations": {
						"test-org": {
							"name": "Test Organization",
							"partitions": {
								"commercial": {
									"url": "https://test.awsapps.com/start",
									"default_region": "us-east-1",
									"regions": ["us-east-1"],
									"accounts": [
										{"alias": "dev", "name": "Development", "id": "123456789012"}
									],
									"roles": ["Developer"]
								}
							}
						}
					}
				}
			}`, CurrentSchemaVersion),
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name: "Missing version",
			json: `{
				"managed": {
					"organizations": {}
				}
			}`,
			wantErr: true,
		},
		{
			name: "No sections",
			json: fmt.Sprintf(`{
				"version": "%s"
			}`, CurrentSchemaVersion),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := SchemaFromJSON([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("SchemaFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && schema == nil {
				t.Error("SchemaFromJSON() returned nil schema without error")
			}
		})
	}
}

func TestSchema_RoundTrip(t *testing.T) {
	original := &Schema{
		Version: CurrentSchemaVersion,
		Managed: &ProfileCollection{
			Organizations: map[string]*Organization{
				"test-org": {
					Name: "Test Organization",
					Partitions: map[string]Partition{
						"commercial": {
							URL:           "https://test.awsapps.com/start",
							DefaultRegion: "us-east-1",
							Regions:       []string{"us-east-1", "us-west-2"},
							Accounts: []Account{
								{Alias: "prod", Name: "Production", ID: "123456789012"},
								{Alias: "dev", Name: "Development", ID: "210987654321"},
							},
							Roles: []string{"Administrator", "Developer"},
						},
					},
				},
			},
			IamUsers: []*IamUser{
				{ProfileName: "my-iam", Region: "us-west-2", CredentialProcess: "aws-vault exec my-user"},
			},
		},
	}

	// Serialize
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	restored, err := SchemaFromJSON(data)
	if err != nil {
		t.Fatalf("SchemaFromJSON() error = %v", err)
	}

	// Verify round-trip
	if restored.Version != original.Version {
		t.Errorf("Round-trip version = %v, want %v", restored.Version, original.Version)
	}

	if len(restored.Managed.Organizations) != len(original.Managed.Organizations) {
		t.Errorf("Round-trip org count = %v, want %v",
			len(restored.Managed.Organizations), len(original.Managed.Organizations))
	}

	if len(restored.Managed.IamUsers) != len(original.Managed.IamUsers) {
		t.Errorf("Round-trip IAM user count = %v, want %v",
			len(restored.Managed.IamUsers), len(original.Managed.IamUsers))
	}
}
