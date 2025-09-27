package settings

import (
	"reflect"

	"aws-profile-manager/internal/logging"
)

// EvaluateDependency checks if a field dependency condition is met.
//
// This function evaluates conditional field visibility/requirement based on
// the values of other fields. Used by the GUI to dynamically show/hide fields
// and enforce conditional validation rules.
//
// Parameters:
//   - dep: Dependency definition (nil means no dependency - always visible)
//   - currentValues: Map of current field values
//
// Returns:
//   - bool: true if the dependency condition is met (field should be visible/enabled)
//
// Example:
//
//	// Show s3.bucket field only when strategy equals "s3"
//	dep := &FieldDependency{
//	    Field: "strategy",
//	    Operator: "equals",
//	    Value: "s3",
//	}
//	shouldShow := EvaluateDependency(dep, values)
func EvaluateDependency(dep *FieldDependency, currentValues map[string]interface{}) bool {
	if dep == nil {
		return true // No dependency = always visible
	}

	fieldValue, exists := currentValues[dep.Field]
	if !exists {
		return false // Dependency field doesn't exist
	}

	switch dep.Operator {
	case "equals", "==":
		return compareValues(fieldValue, dep.Value)

	case "not_equals", "!=":
		return !compareValues(fieldValue, dep.Value)

	case "in":
		return valueInList(fieldValue, dep.Value)

	case "not_in":
		return !valueInList(fieldValue, dep.Value)

	case "greater_than", ">":
		return compareNumeric(fieldValue, dep.Value, func(a, b float64) bool { return a > b })

	case "less_than", "<":
		return compareNumeric(fieldValue, dep.Value, func(a, b float64) bool { return a < b })

	case "greater_or_equal", ">=":
		return compareNumeric(fieldValue, dep.Value, func(a, b float64) bool { return a >= b })

	case "less_or_equal", "<=":
		return compareNumeric(fieldValue, dep.Value, func(a, b float64) bool { return a <= b })

	default:
		// Unknown operator - default to visible
		return true
	}
}

// compareValues compares two values for equality with type conversion handling.
//
// This internal function performs type-safe comparison, attempting to convert
// types when necessary to enable comparison.
//
// Parameters:
//   - a: First value to compare
//   - b: Second value to compare
//
// Returns:
//   - bool: true if values are equal (after type conversion if needed)
func compareValues(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use reflection for type-safe comparison
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// If types are different, try to convert
	if va.Type() != vb.Type() {
		// Try converting b to a's type
		if vb.Type().ConvertibleTo(va.Type()) {
			vb = vb.Convert(va.Type())
		} else {
			return false
		}
	}

	return reflect.DeepEqual(va.Interface(), vb.Interface())
}

// valueInList checks if a value exists in a list.
//
// This internal function checks for membership in a slice or array,
// using compareValues for element comparison.
//
// Parameters:
//   - value: Value to search for
//   - list: Slice or array to search in
//
// Returns:
//   - bool: true if value is found in list
func valueInList(value interface{}, list interface{}) bool {
	// list should be a slice
	listValue := reflect.ValueOf(list)
	if listValue.Kind() != reflect.Slice && listValue.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < listValue.Len(); i++ {
		item := listValue.Index(i).Interface()
		if compareValues(value, item) {
			return true
		}
	}

	return false
}

// compareNumeric compares numeric values using a comparison function.
//
// This internal function converts values to float64 and applies the provided
// comparison function. Supports various numeric types.
//
// Parameters:
//   - a: First numeric value
//   - b: Second numeric value
//   - compare: Comparison function (e.g., func(a, b float64) bool { return a > b })
//
// Returns:
//   - bool: Result of comparison function, false if conversion fails
func compareNumeric(a, b interface{}, compare func(float64, float64) bool) bool {
	numA, okA := toFloat64(a)
	numB, okB := toFloat64(b)

	if !okA || !okB {
		return false
	}

	return compare(numA, numB)
}

// toFloat64 converts various numeric types to float64.
//
// This internal function handles conversion from multiple numeric types
// (int, int32, int64, uint, uint32, uint64, float32, float64) to float64
// for numeric comparisons.
//
// Parameters:
//   - value: Value to convert
//
// Returns:
//   - float64: Converted value
//   - bool: true if conversion succeeded, false otherwise
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

// ValidateDependency validates that a dependency is properly configured.
//
// This function checks that a FieldDependency has all required fields and
// uses valid operators. Used during settings initialization to catch
// configuration errors early.
//
// Parameters:
//   - dep: Dependency definition to validate (nil is valid - means no dependency)
//
// Returns:
//   - error: First validation error encountered, nil if valid
//
// Validation Rules:
//   - Field name must not be empty
//   - Operator must be one of the supported operators
//   - Value must be a slice/array for "in" and "not_in" operators
//
// Supported Operators:
//   - "equals", "==" - Field equals value
//   - "not_equals", "!=" - Field does not equal value
//   - "in" - Field value is in list
//   - "not_in" - Field value is not in list
//   - "greater_than", ">" - Field > value
//   - "less_than", "<" - Field < value
//   - "greater_or_equal", ">=" - Field >= value
//   - "less_or_equal", "<=" - Field <= value
func ValidateDependency(dep *FieldDependency) error {
	if dep == nil {
		return nil
	}

	if dep.Field == "" {
		return logging.Log.Error("dependency field name cannot be empty")
	}

	validOperators := map[string]bool{
		"equals":           true,
		"==":               true,
		"not_equals":       true,
		"!=":               true,
		"in":               true,
		"not_in":           true,
		"greater_than":     true,
		">":                true,
		"less_than":        true,
		"<":                true,
		"greater_or_equal": true,
		">=":               true,
		"less_or_equal":    true,
		"<=":               true,
	}

	if !validOperators[dep.Operator] {
		return logging.Log.Errorf("invalid dependency operator: %s", dep.Operator)
	}

	// Validate value type for "in" and "not_in" operators
	if dep.Operator == "in" || dep.Operator == "not_in" {
		listValue := reflect.ValueOf(dep.Value)
		if listValue.Kind() != reflect.Slice && listValue.Kind() != reflect.Array {
			return logging.Log.Errorf("dependency value for '%s' operator must be a slice or array", dep.Operator)
		}
	}

	return nil
}
