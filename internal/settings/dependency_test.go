package settings

import (
	"testing"

	"aws-profile-manager/internal/test"
)

func TestEvaluateDependency_Equals(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		dep      *FieldDependency
		values   map[string]interface{}
		expected bool
	}{
		{
			name: "nil dependency always visible",
			dep:  nil,
			values: map[string]interface{}{
				"field1": "value1",
			},
			expected: true,
		},
		{
			name: "equals operator - match",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "equals",
				Value:    "s3",
			},
			values: map[string]interface{}{
				"strategy": "s3",
			},
			expected: true,
		},
		{
			name: "equals operator - no match",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "equals",
				Value:    "s3",
			},
			values: map[string]interface{}{
				"strategy": "http",
			},
			expected: false,
		},
		{
			name: "== operator - match",
			dep: &FieldDependency{
				Field:    "enabled",
				Operator: "==",
				Value:    true,
			},
			values: map[string]interface{}{
				"enabled": true,
			},
			expected: true,
		},
		{
			name: "field doesn't exist",
			dep: &FieldDependency{
				Field:    "missing_field",
				Operator: "equals",
				Value:    "value",
			},
			values: map[string]interface{}{
				"other_field": "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateDependency(tt.dep, tt.values)
			if result != tt.expected {
				t.Errorf("EvaluateDependency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluateDependency_NotEquals(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		dep      *FieldDependency
		values   map[string]interface{}
		expected bool
	}{
		{
			name: "not_equals operator - different",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "not_equals",
				Value:    "s3",
			},
			values: map[string]interface{}{
				"strategy": "http",
			},
			expected: true,
		},
		{
			name: "not_equals operator - same",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "not_equals",
				Value:    "s3",
			},
			values: map[string]interface{}{
				"strategy": "s3",
			},
			expected: false,
		},
		{
			name: "!= operator - different",
			dep: &FieldDependency{
				Field:    "enabled",
				Operator: "!=",
				Value:    false,
			},
			values: map[string]interface{}{
				"enabled": true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateDependency(tt.dep, tt.values)
			if result != tt.expected {
				t.Errorf("EvaluateDependency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluateDependency_In(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		dep      *FieldDependency
		values   map[string]interface{}
		expected bool
	}{
		{
			name: "in operator - value in list",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "in",
				Value:    []interface{}{"s3", "http", "git"},
			},
			values: map[string]interface{}{
				"strategy": "http",
			},
			expected: true,
		},
		{
			name: "in operator - value not in list",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "in",
				Value:    []interface{}{"s3", "http"},
			},
			values: map[string]interface{}{
				"strategy": "git",
			},
			expected: false,
		},
		{
			name: "not_in operator - value not in list",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "not_in",
				Value:    []interface{}{"s3", "http"},
			},
			values: map[string]interface{}{
				"strategy": "git",
			},
			expected: true,
		},
		{
			name: "not_in operator - value in list",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "not_in",
				Value:    []interface{}{"s3", "http"},
			},
			values: map[string]interface{}{
				"strategy": "s3",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateDependency(tt.dep, tt.values)
			if result != tt.expected {
				t.Errorf("EvaluateDependency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluateDependency_NumericComparisons(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		dep      *FieldDependency
		values   map[string]interface{}
		expected bool
	}{
		{
			name: "greater_than - true",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: "greater_than",
				Value:    30,
			},
			values: map[string]interface{}{
				"timeout": 60,
			},
			expected: true,
		},
		{
			name: "greater_than - false",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: "greater_than",
				Value:    30,
			},
			values: map[string]interface{}{
				"timeout": 20,
			},
			expected: false,
		},
		{
			name: "> operator - true",
			dep: &FieldDependency{
				Field:    "count",
				Operator: ">",
				Value:    5,
			},
			values: map[string]interface{}{
				"count": 10,
			},
			expected: true,
		},
		{
			name: "less_than - true",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: "less_than",
				Value:    60,
			},
			values: map[string]interface{}{
				"timeout": 30,
			},
			expected: true,
		},
		{
			name: "< operator - false",
			dep: &FieldDependency{
				Field:    "count",
				Operator: "<",
				Value:    5,
			},
			values: map[string]interface{}{
				"count": 10,
			},
			expected: false,
		},
		{
			name: "greater_or_equal - equal",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: "greater_or_equal",
				Value:    30,
			},
			values: map[string]interface{}{
				"timeout": 30,
			},
			expected: true,
		},
		{
			name: ">= operator - greater",
			dep: &FieldDependency{
				Field:    "count",
				Operator: ">=",
				Value:    5,
			},
			values: map[string]interface{}{
				"count": 10,
			},
			expected: true,
		},
		{
			name: "less_or_equal - equal",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: "less_or_equal",
				Value:    60,
			},
			values: map[string]interface{}{
				"timeout": 60,
			},
			expected: true,
		},
		{
			name: "<= operator - less",
			dep: &FieldDependency{
				Field:    "count",
				Operator: "<=",
				Value:    5,
			},
			values: map[string]interface{}{
				"count": 3,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateDependency(tt.dep, tt.values)
			if result != tt.expected {
				t.Errorf("EvaluateDependency() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEvaluateDependency_UnknownOperator(t *testing.T) {
	test.SetupTestEnvironment(t)

	dep := &FieldDependency{
		Field:    "field1",
		Operator: "unknown_op",
		Value:    "value",
	}

	values := map[string]interface{}{
		"field1": "value",
	}

	// Unknown operator should default to true (visible)
	result := EvaluateDependency(dep, values)
	if result != true {
		t.Errorf("Unknown operator should default to true, got %v", result)
	}
}

func TestCompareValues_TypeConversion(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "same string values",
			a:        "test",
			b:        "test",
			expected: true,
		},
		{
			name:     "different string values",
			a:        "test1",
			b:        "test2",
			expected: false,
		},
		{
			name:     "same int values",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "same bool values",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "nil values",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil value",
			a:        "test",
			b:        nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestValueInList(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		value    interface{}
		list     interface{}
		expected bool
	}{
		{
			name:     "value in list",
			value:    "b",
			list:     []interface{}{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "value not in list",
			value:    "d",
			list:     []interface{}{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "numeric value in list",
			value:    2,
			list:     []interface{}{1, 2, 3},
			expected: true,
		},
		{
			name:     "empty list",
			value:    "a",
			list:     []interface{}{},
			expected: false,
		},
		{
			name:     "non-slice list",
			value:    "a",
			list:     "not a slice",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := valueInList(tt.value, tt.list)
			if result != tt.expected {
				t.Errorf("valueInList(%v, %v) = %v, want %v", tt.value, tt.list, result, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		shouldOk bool
	}{
		{
			name:     "float64",
			value:    3.14,
			expected: 3.14,
			shouldOk: true,
		},
		{
			name:     "float32",
			value:    float32(2.5),
			expected: 2.5,
			shouldOk: true,
		},
		{
			name:     "int",
			value:    42,
			expected: 42.0,
			shouldOk: true,
		},
		{
			name:     "int32",
			value:    int32(10),
			expected: 10.0,
			shouldOk: true,
		},
		{
			name:     "int64",
			value:    int64(100),
			expected: 100.0,
			shouldOk: true,
		},
		{
			name:     "uint",
			value:    uint(50),
			expected: 50.0,
			shouldOk: true,
		},
		{
			name:     "uint32",
			value:    uint32(25),
			expected: 25.0,
			shouldOk: true,
		},
		{
			name:     "uint64",
			value:    uint64(75),
			expected: 75.0,
			shouldOk: true,
		},
		{
			name:     "string",
			value:    "not a number",
			expected: 0,
			shouldOk: false,
		},
		{
			name:     "bool",
			value:    true,
			expected: 0,
			shouldOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toFloat64(tt.value)
			if ok != tt.shouldOk {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.value, ok, tt.shouldOk)
			}
			if tt.shouldOk && result != tt.expected {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestValidateDependency(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name      string
		dep       *FieldDependency
		shouldErr bool
	}{
		{
			name:      "nil dependency",
			dep:       nil,
			shouldErr: false,
		},
		{
			name: "valid equals dependency",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "equals",
				Value:    "s3",
			},
			shouldErr: false,
		},
		{
			name: "valid in dependency",
			dep: &FieldDependency{
				Field:    "strategy",
				Operator: "in",
				Value:    []interface{}{"s3", "http"},
			},
			shouldErr: false,
		},
		{
			name: "empty field name",
			dep: &FieldDependency{
				Field:    "",
				Operator: "equals",
				Value:    "value",
			},
			shouldErr: true,
		},
		{
			name: "invalid operator",
			dep: &FieldDependency{
				Field:    "field1",
				Operator: "invalid_op",
				Value:    "value",
			},
			shouldErr: true,
		},
		{
			name: "in operator with non-slice value",
			dep: &FieldDependency{
				Field:    "field1",
				Operator: "in",
				Value:    "not a slice",
			},
			shouldErr: true,
		},
		{
			name: "all comparison operators valid",
			dep: &FieldDependency{
				Field:    "timeout",
				Operator: ">=",
				Value:    30,
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependency(tt.dep)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateDependency() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	test.SetupTestEnvironment(t)

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		compare  func(float64, float64) bool
		expected bool
	}{
		{
			name:     "both numeric - greater",
			a:        10,
			b:        5,
			compare:  func(a, b float64) bool { return a > b },
			expected: true,
		},
		{
			name:     "both numeric - less",
			a:        5,
			b:        10,
			compare:  func(a, b float64) bool { return a < b },
			expected: true,
		},
		{
			name:     "non-numeric a",
			a:        "not a number",
			b:        10,
			compare:  func(a, b float64) bool { return a > b },
			expected: false,
		},
		{
			name:     "non-numeric b",
			a:        10,
			b:        "not a number",
			compare:  func(a, b float64) bool { return a > b },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNumeric(tt.a, tt.b, tt.compare)
			if result != tt.expected {
				t.Errorf("compareNumeric() = %v, want %v", result, tt.expected)
			}
		})
	}
}
