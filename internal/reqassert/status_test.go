package reqassert

import (
	"testing"
)

func TestStatusAssertionValidate(t *testing.T) {
	tests := []struct {
		name        string
		assertion   StatusAssertion
		statusCode  int
		shouldError bool
	}{
		{
			name: "status match - pass",
			assertion: StatusAssertion{
				Expected: 200,
			},
			statusCode:  200,
			shouldError: false,
		},
		{
			name: "status mismatch - fail",
			assertion: StatusAssertion{
				Expected: 200,
			},
			statusCode:  404,
			shouldError: true,
		},
		{
			name: "error status - pass",
			assertion: StatusAssertion{
				Expected: 500,
			},
			statusCode:  500,
			shouldError: false,
		},
		{
			name: "redirect status - pass",
			assertion: StatusAssertion{
				Expected: 302,
			},
			statusCode:  302,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the context
			ctx := &AssertionContext{
				StatusCode: tt.statusCode,
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

func TestStatusAssertionFactory(t *testing.T) {
	tests := []struct {
		name           string
		expected       interface{}
		expectedStatus int
		shouldError    bool
	}{
		{
			name:           "integer value",
			expected:       200,
			expectedStatus: 200,
			shouldError:    false,
		},
		{
			name:           "float value - converted to int",
			expected:       float64(404),
			expectedStatus: 404,
			shouldError:    false,
		},
		{
			name:        "string value - not supported",
			expected:    "200",
			shouldError: true,
		},
		{
			name:        "boolean value - not supported",
			expected:    true,
			shouldError: true,
		},
		{
			name:        "complex type - not supported",
			expected:    []int{200, 201},
			shouldError: true,
		},
	}

	factory := &StatusAssertionFactory{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion, err := factory.Create("", tt.expected)
			
			if tt.shouldError {
				if err == nil {
					t.Fatalf("Factory.Create() should have returned an error for unsupported type")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Factory.Create() error = %v", err)
			}

			statusAssertion, ok := assertion.(*StatusAssertion)
			if !ok {
				t.Fatalf("Factory.Create() did not return a *StatusAssertion")
			}

			if statusAssertion.Expected != tt.expectedStatus {
				t.Errorf("Expected = %v, want %v", statusAssertion.Expected, tt.expectedStatus)
			}
		})
	}
}