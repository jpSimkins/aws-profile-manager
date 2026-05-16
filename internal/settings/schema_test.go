package settings

import (
	"encoding/json"
	"testing"

	"aws-profile-manager/internal/test"
)

func TestFieldSchema_JSONSerialization(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name   string
		schema FieldSchema
	}{
		{
			name: "simple string field",
			schema: FieldSchema{
				Type:        "string",
				Description: "Test field",
				Required:    true,
				Default:     "default_value",
			},
		},
		{
			name: "enum field",
			schema: FieldSchema{
				Type:        "string",
				Description: "Test enum",
				Required:    true,
				Enum:        []string{"option1", "option2", "option3"},
			},
		},
		{
			name: "numeric field with bounds",
			schema: FieldSchema{
				Type:        "int",
				Description: "Test numeric",
				Required:    false,
				Min:         ptrFloat64(0),
				Max:         ptrFloat64(100),
			},
		},
		{
			name: "field with dependency",
			schema: FieldSchema{
				Type:        "string",
				Description: "Conditional field",
				DependsOn: &FieldDependency{
					Field:    "other_field",
					Operator: "equals",
					Value:    "enabled",
				},
			},
		},
		{
			name: "nested object field",
			schema: FieldSchema{
				Type:        "object",
				Description: "Nested object",
				Nested: &Schema{
					Version: "1.0",
					Fields: map[string]FieldSchema{
						"inner_field": {
							Type:        "string",
							Description: "Inner field",
							Required:    true,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			data, err := json.Marshal(tt.schema)
			if err != nil {
				t.Fatalf("Failed to marshal schema: %v", err)
			}

			// Deserialize from JSON
			var decoded FieldSchema
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal schema: %v", err)
			}

			// Verify type
			if decoded.Type != tt.schema.Type {
				t.Errorf("Type mismatch: got %v, want %v", decoded.Type, tt.schema.Type)
			}

			// Verify description
			if decoded.Description != tt.schema.Description {
				t.Errorf("Description mismatch: got %v, want %v", decoded.Description, tt.schema.Description)
			}

			// Verify required flag
			if decoded.Required != tt.schema.Required {
				t.Errorf("Required mismatch: got %v, want %v", decoded.Required, tt.schema.Required)
			}
		})
	}
}

func TestSchema_JSONSerialization(t *testing.T) {
	test.SetupTestEnvironment(t)

	schema := Schema{
		Version: "1.0.0",
		Fields: map[string]FieldSchema{
			"field1": {
				Type:        "string",
				Description: "First field",
				Required:    true,
			},
			"field2": {
				Type:        "bool",
				Description: "Second field",
				Required:    false,
				Default:     false,
			},
			"field3": {
				Type:        "int",
				Description: "Third field",
				Min:         ptrFloat64(0),
				Max:         ptrFloat64(100),
			},
		},
	}

	// Serialize to JSON
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}

	// Deserialize from JSON
	var decoded Schema
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Verify version
	if decoded.Version != schema.Version {
		t.Errorf("Version mismatch: got %v, want %v", decoded.Version, schema.Version)
	}

	// Verify field count
	if len(decoded.Fields) != len(schema.Fields) {
		t.Errorf("Field count mismatch: got %v, want %v", len(decoded.Fields), len(schema.Fields))
	}

	// Verify individual fields
	for name, field := range schema.Fields {
		decoded, exists := decoded.Fields[name]
		if !exists {
			t.Errorf("Field %s not found in decoded schema", name)
			continue
		}

		if decoded.Type != field.Type {
			t.Errorf("Field %s type mismatch: got %v, want %v", name, decoded.Type, field.Type)
		}
	}
}

func TestFieldDependency_JSONSerialization(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name string
		dep  FieldDependency
	}{
		{
			name: "equals operator",
			dep: FieldDependency{
				Field:    "strategy",
				Operator: "equals",
				Value:    "s3",
			},
		},
		{
			name: "in operator with slice",
			dep: FieldDependency{
				Field:    "strategy",
				Operator: "in",
				Value:    []string{"s3", "http"},
			},
		},
		{
			name: "greater_than operator",
			dep: FieldDependency{
				Field:    "timeout",
				Operator: "greater_than",
				Value:    30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			data, err := json.Marshal(tt.dep)
			if err != nil {
				t.Fatalf("Failed to marshal dependency: %v", err)
			}

			// Deserialize from JSON
			var decoded FieldDependency
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal dependency: %v", err)
			}

			// Verify field
			if decoded.Field != tt.dep.Field {
				t.Errorf("Field mismatch: got %v, want %v", decoded.Field, tt.dep.Field)
			}

			// Verify operator
			if decoded.Operator != tt.dep.Operator {
				t.Errorf("Operator mismatch: got %v, want %v", decoded.Operator, tt.dep.Operator)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	test.SetupTestEnvironment(t)

	err := ValidationError{
		Field:   "test_field",
		Message: "field is required",
		Value:   nil,
	}

	errorString := err.Error()
	if errorString != "field is required" {
		t.Errorf("Error() returned wrong message: got %v, want %v", errorString, "field is required")
	}
}

func TestValidationError_JSONSerialization(t *testing.T) {
	test.SetupTestEnvironment(t)

	err := ValidationError{
		Field:   "test_field",
		Message: "invalid value",
		Value:   "bad_value",
	}

	// Serialize to JSON
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal validation error: %v", jsonErr)
	}

	// Deserialize from JSON
	var decoded ValidationError
	if jsonErr := json.Unmarshal(data, &decoded); jsonErr != nil {
		t.Fatalf("Failed to unmarshal validation error: %v", jsonErr)
	}

	// Verify field
	if decoded.Field != err.Field {
		t.Errorf("Field mismatch: got %v, want %v", decoded.Field, err.Field)
	}

	// Verify message
	if decoded.Message != err.Message {
		t.Errorf("Message mismatch: got %v, want %v", decoded.Message, err.Message)
	}
}

func TestFieldSchema_UIHints(t *testing.T) {
	test.SetupTestEnvironment(t)

	schema := FieldSchema{
		Type:        "string",
		Description: "Test field",
		Group:       "authentication",
		Order:       5,
		Placeholder: "Enter value here",
		HelpText:    "Additional help text",
	}

	// Verify all UI hints are set
	if schema.Group != "authentication" {
		t.Errorf("Group mismatch: got %v, want %v", schema.Group, "authentication")
	}

	if schema.Order != 5 {
		t.Errorf("Order mismatch: got %v, want %v", schema.Order, 5)
	}

	if schema.Placeholder != "Enter value here" {
		t.Errorf("Placeholder mismatch: got %v, want %v", schema.Placeholder, "Enter value here")
	}

	if schema.HelpText != "Additional help text" {
		t.Errorf("HelpText mismatch: got %v, want %v", schema.HelpText, "Additional help text")
	}
}

func TestFieldSchema_PatternValidation(t *testing.T) {
	test.SetupTestEnvironment(t)

	schema := FieldSchema{
		Type:        "string",
		Description: "URL field",
		Pattern:     `^https?://.*`,
	}

	// Verify pattern is set
	if schema.Pattern != `^https?://.*` {
		t.Errorf("Pattern mismatch: got %v, want %v", schema.Pattern, `^https?://.*`)
	}
}

func TestSchema_EmptyFields(t *testing.T) {
	test.SetupTestEnvironment(t)

	schema := Schema{
		Version: "1.0",
		Fields:  map[string]FieldSchema{},
	}

	if len(schema.Fields) != 0 {
		t.Errorf("Expected empty fields map, got %v fields", len(schema.Fields))
	}
}

// Helper function for tests
func ptrFloat64(v float64) *float64 {
	return &v
}
