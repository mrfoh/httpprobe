package reqassert

import (
	"encoding/json"
	"testing"
)

func TestBodyAssertionValidate(t *testing.T) {
	tests := []struct {
		name        string
		assertion   BodyAssertion
		bodyJSON    string
		shouldError bool
	}{
		{
			name: "simple equals - pass",
			assertion: BodyAssertion{
				JSONPath:       "$.name",
				ExpectedValue:  "test",
				ComparisonType: "equals",
			},
			bodyJSON:    `{"name": "test"}`,
			shouldError: false,
		},
		{
			name: "simple equals - fail",
			assertion: BodyAssertion{
				JSONPath:       "$.name",
				ExpectedValue:  "test",
				ComparisonType: "equals",
			},
			bodyJSON:    `{"name": "different"}`,
			shouldError: true,
		},
		{
			name: "default comparison - pass",
			assertion: BodyAssertion{
				JSONPath:      "$.age",
				ExpectedValue: float64(25),
			},
			bodyJSON:    `{"age": 25}`,
			shouldError: false,
		},
		{
			name: "path not found",
			assertion: BodyAssertion{
				JSONPath:      "$.missing",
				ExpectedValue: "test",
			},
			bodyJSON:    `{"name": "test"}`,
			shouldError: true,
		},
		{
			name: "contains comparison - pass",
			assertion: BodyAssertion{
				JSONPath:       "$.description",
				ExpectedValue:  "world",
				ComparisonType: "contains",
			},
			bodyJSON:    `{"description": "hello world"}`,
			shouldError: false,
		},
		{
			name: "contains comparison - fail",
			assertion: BodyAssertion{
				JSONPath:       "$.description",
				ExpectedValue:  "missing",
				ComparisonType: "contains",
			},
			bodyJSON:    `{"description": "hello world"}`,
			shouldError: true,
		},
		{
			name: "contains comparison - non-string actual value",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  "5",
				ComparisonType: "contains",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: true,
		},
		{
			name: "greater than - pass",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  10,
				ComparisonType: "gt",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: false,
		},
		{
			name: "greater than - fail",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  50,
				ComparisonType: "gt",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: true,
		},
		{
			name: "greater than or equal - pass (equal)",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  42,
				ComparisonType: "gte",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: false,
		},
		{
			name: "less than - pass",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  50,
				ComparisonType: "lt",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: false,
		},
		{
			name: "less than or equal - pass (equal)",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  42,
				ComparisonType: "lte",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: false,
		},
		{
			name: "unknown comparison type",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  42,
				ComparisonType: "unknown",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: true,
		},
		{
			name: "numeric comparison with string values - pass",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  "10",
				ComparisonType: "gt",
			},
			bodyJSON:    `{"count": "42"}`,
			shouldError: false,
		},
		{
			name: "numeric comparison with invalid string - fail",
			assertion: BodyAssertion{
				JSONPath:       "$.count",
				ExpectedValue:  "ten",
				ComparisonType: "gt",
			},
			bodyJSON:    `{"count": 42}`,
			shouldError: true,
		},
		{
			name: "nested path - pass",
			assertion: BodyAssertion{
				JSONPath:      "$.user.details.name",
				ExpectedValue: "John",
			},
			bodyJSON:    `{"user": {"details": {"name": "John", "age": 30}}}`,
			shouldError: false,
		},
		{
			name: "array access - pass",
			assertion: BodyAssertion{
				JSONPath:      "$.users[0].name",
				ExpectedValue: "John",
			},
			bodyJSON:    `{"users": [{"name": "John"}, {"name": "Jane"}]}`,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the JSON body
			var bodyMap map[string]interface{}
			if err := json.Unmarshal([]byte(tt.bodyJSON), &bodyMap); err != nil {
				t.Fatalf("Failed to parse JSON: %v", err)
			}

			// Create the context
			ctx := &AssertionContext{
				BodyMap: bodyMap,
				Body:    []byte(tt.bodyJSON),
			}

			// Run the validation
			err := tt.assertion.Validate(ctx)

			// Check if the error state matches what we expect
			if (err != nil) != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v - %v", tt.shouldError, err != nil, err)
			}
		})
	}
}

func TestBodyAssertionFactory(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expected      interface{}
		expectedPath  string
		expectedType  string
		expectedValue interface{}
	}{
		{
			name:          "simple key",
			key:           "$.name",
			expected:      "test",
			expectedPath:  "$.name",
			expectedType:  "",
			expectedValue: "test",
		},
		{
			name:          "comparison in expected string",
			key:           "$.age",
			expected:      "> 18",
			expectedPath:  "$.age",
			expectedType:  ">",
			expectedValue: "18",
		},
		{
			name:          "contains comparison",
			key:           "$.description",
			expected:      "contains hello",
			expectedPath:  "$.description",
			expectedType:  "contains",
			expectedValue: "hello",
		},
		{
			name:          "equals comparison",
			key:           "$.status",
			expected:      "= success",
			expectedPath:  "$.status",
			expectedType:  "=",
			expectedValue: "success",
		},
		{
			name:          "non-string expected",
			key:           "$.count",
			expected:      42,
			expectedPath:  "$.count",
			expectedType:  "",
			expectedValue: 42,
		},
	}

	factory := &BodyAssertionFactory{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion, err := factory.Create(tt.key, tt.expected)
			if err != nil {
				t.Fatalf("Factory.Create() error = %v", err)
			}

			bodyAssertion, ok := assertion.(*BodyAssertion)
			if !ok {
				t.Fatalf("Factory.Create() did not return a *BodyAssertion")
			}

			if bodyAssertion.JSONPath != tt.expectedPath {
				t.Errorf("JSONPath = %v, want %v", bodyAssertion.JSONPath, tt.expectedPath)
			}

			if bodyAssertion.ComparisonType != tt.expectedType {
				t.Errorf("ComparisonType = %v, want %v", bodyAssertion.ComparisonType, tt.expectedType)
			}

			if bodyAssertion.ExpectedValue != tt.expectedValue {
				t.Errorf("ExpectedValue = %v, want %v", bodyAssertion.ExpectedValue, tt.expectedValue)
			}
		})
	}
}

func TestCompareNumeric(t *testing.T) {
	tests := []struct {
		name        string
		actual      interface{}
		expected    interface{}
		comparison  func(float64, float64) bool
		shouldError bool
	}{
		{
			name:        "int comparison - pass",
			actual:      42,
			expected:    40,
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: false,
		},
		{
			name:        "int comparison - fail",
			actual:      30,
			expected:    40,
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: true,
		},
		{
			name:        "float comparison - pass",
			actual:      42.5,
			expected:    42.0,
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: false,
		},
		{
			name:        "string to float - pass",
			actual:      "42.5",
			expected:    "40",
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: false,
		},
		{
			name:        "invalid string - actual",
			actual:      "not-a-number",
			expected:    40,
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: true,
		},
		{
			name:        "invalid string - expected",
			actual:      42,
			expected:    "not-a-number",
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: true,
		},
		{
			name:        "unsupported type - actual",
			actual:      []int{1, 2, 3},
			expected:    40,
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: true,
		},
		{
			name:        "unsupported type - expected",
			actual:      42,
			expected:    []int{1, 2, 3},
			comparison:  func(a, b float64) bool { return a > b },
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := compareNumeric(tt.actual, tt.expected, tt.comparison)

			if (err != nil) != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v - %v", tt.shouldError, err != nil, err)
			}
		})
	}
}