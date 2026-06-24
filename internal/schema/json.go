package schema

import (
	"encoding/json"
	"fmt"

	"aws-profile-manager/internal/logging"
)

// ToJSON serializes the Schema to JSON format.
//
// This method converts the entire schema structure to formatted JSON with
// 2-space indentation for readability.
//
// Returns:
//   - []byte: JSON-encoded schema data
//   - error: Any error encountered during serialization
//
// Example:
//
//	jsonData, err := schema.ToJSON()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("config.json", jsonData, 0600)
func (s *Schema) ToJSON() ([]byte, error) {
	logging.Debug.Log("Serializing Schema to JSON")

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize schema: %w", err)
	}

	logging.Debug.Log("Serialization completed", "bytes", len(data))
	return data, nil
}

// SchemaFromJSON deserializes JSON data to a Schema.
//
// This function parses JSON data into a Schema structure and validates it
// automatically. The validation ensures the schema is well-formed and ready
// to use.
//
// Parameters:
//   - data: JSON-encoded schema data
//
// Returns:
//   - *Schema: Parsed and validated schema
//   - error: Any error encountered during parsing or validation
//
// Example:
//
//	jsonData, _ := os.ReadFile("config.json")
//	schema, err := schema.SchemaFromJSON(jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
func SchemaFromJSON(data []byte) (*Schema, error) {
	logging.Debug.Log("Deserializing JSON to Schema", "bytes", len(data))

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to deserialize schema: %w", err)
	}

	// Validate after deserialization
	if err := schema.Validate(); err != nil {
		return nil, err
	}

	logging.Debug.Log("Schema deserialization completed")
	return &schema, nil
}
