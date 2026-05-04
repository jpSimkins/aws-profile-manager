package settings

// Schema defines the structure and validation rules for settings sections.
//
// Schema is used by the GUI to dynamically build settings forms, ensuring
// consistency between validation logic and UI generation.
type Schema struct {
	Version     string                 `json:"version"`               // Schema version
	Description string                 `json:"description,omitempty"` // Human-readable description shown below the section heading
	Fields      map[string]FieldSchema `json:"fields"`                // Field definitions keyed by field name
}

// FieldSchema defines validation rules and UI metadata for a single field.
//
// This struct contains all information needed to validate a field value and
// generate appropriate UI widgets automatically.
type FieldSchema struct {
	Type        string      `json:"type"`              // Data type: "string", "bool", "int", "float", "object", "array"
	Description string      `json:"description"`       // Human-readable description for users
	Label       string      `json:"label,omitempty"`   // Display label (overrides humanized field name)
	Required    bool        `json:"required"`          // Whether field is required
	Default     interface{} `json:"default"`           // Default value
	Enum        []string    `json:"enum,omitempty"`    // Valid values for string fields (creates dropdown)
	Min         *float64    `json:"min,omitempty"`     // Minimum value for numeric fields
	Max         *float64    `json:"max,omitempty"`     // Maximum value for numeric fields
	Pattern     string      `json:"pattern,omitempty"` // Regex pattern for string validation
	Nested      *Schema     `json:"nested,omitempty"`  // Schema for nested object fields

	// Conditional visibility
	DependsOn *FieldDependency `json:"depends_on,omitempty"` // Field shown only if dependency is met

	// UI hints
	Group           string `json:"group,omitempty"`            // Group related fields together in UI
	Order           int    `json:"order,omitempty"`            // Display order (lower = first)
	Placeholder     string `json:"placeholder,omitempty"`      // Placeholder text for input widgets
	HelpText        string `json:"help_text,omitempty"`        // Additional contextual help text
	RequiresRestart bool   `json:"requires_restart,omitempty"` // If true, a restart prompt is shown after save
}

// FieldDependency defines a conditional dependency for field visibility.
//
// Used to implement conditional UI logic where certain fields are only shown
// or required based on the values of other fields.
type FieldDependency struct {
	Field    string      `json:"field"`    // Field name to check
	Operator string      `json:"operator"` // Comparison operator: "equals", "not_equals", "in", "not_in", "greater_than", "less_than"
	Value    interface{} `json:"value"`    // Expected value(s) for comparison
}

// ValidationError represents a settings validation error.
type ValidationError struct {
	Field   string      `json:"field"`           // Field name that failed validation
	Message string      `json:"message"`         // Error message
	Value   interface{} `json:"value,omitempty"` // The invalid value
}

// Error implements the error interface.
func (v ValidationError) Error() string {
	return v.Message
}

// SchemaProvider interface for types that provide schemas
// Each settings type (GUI, Sync, etc.) implements this
type SchemaProvider interface {
	GetSchema() Schema
}
