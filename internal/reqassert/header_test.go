package reqassert

import (
	"testing"
)

func TestHeaderAssertionValidate(t *testing.T) {
	tests := []struct {
		name        string
		assertion   HeaderAssertion
		headers     map[string]string
		shouldError bool
	}{
		{
			name: "exact match - pass",
			assertion: HeaderAssertion{
				HeaderName:    "content-type",
				ExpectedValue: "application/json",
			},
			headers: map[string]string{
				"content-type": "application/json",
			},
			shouldError: false,
		},
		{
			name: "case sensitive - fail",
			assertion: HeaderAssertion{
				HeaderName:    "content-type",
				ExpectedValue: "application/json",
			},
			headers: map[string]string{
				"content-type": "Application/JSON",
			},
			shouldError: true,
		},
		{
			name: "header not found",
			assertion: HeaderAssertion{
				HeaderName:    "x-custom-header",
				ExpectedValue: "custom-value",
			},
			headers: map[string]string{
				"content-type": "application/json",
			},
			shouldError: true,
		},
		{
			name: "whitespace trimming - pass",
			assertion: HeaderAssertion{
				HeaderName:    "content-type",
				ExpectedValue: "application/json ",
			},
			headers: map[string]string{
				"content-type": " application/json",
			},
			shouldError: false,
		},
		{
			name: "numeric header value - pass",
			assertion: HeaderAssertion{
				HeaderName:    "content-length",
				ExpectedValue: "42",
			},
			headers: map[string]string{
				"content-length": "42",
			},
			shouldError: false,
		},
		{
			name: "multiple headers - pass",
			assertion: HeaderAssertion{
				HeaderName:    "x-rate-limit",
				ExpectedValue: "100",
			},
			headers: map[string]string{
				"content-type": "application/json",
				"x-rate-limit": "100",
				"x-api-key":    "abc123",
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the context
			ctx := &AssertionContext{
				Headers: tt.headers,
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

func TestHeaderAssertionFactory(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expected      interface{}
		expectedName  string
		expectedValue string
		shouldError   bool
	}{
		{
			name:          "string value",
			key:           "content-type",
			expected:      "application/json",
			expectedName:  "content-type",
			expectedValue: "application/json",
			shouldError:   false,
		},
		{
			name:          "integer value",
			key:           "content-length",
			expected:      42,
			expectedName:  "content-length",
			expectedValue: "42",
			shouldError:   false,
		},
		{
			name:          "float value",
			key:           "x-rate-limit",
			expected:      99.9,
			expectedName:  "x-rate-limit",
			expectedValue: "99.9",
			shouldError:   false,
		},
		{
			name:          "boolean value",
			key:           "x-cache-hit",
			expected:      true,
			expectedName:  "x-cache-hit",
			expectedValue: "true",
			shouldError:   false,
		},
		{
			name:        "unsupported type",
			key:         "x-invalid",
			expected:    []string{"not", "supported"},
			shouldError: true,
		},
	}

	factory := &HeaderAssertionFactory{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertion, err := factory.Create(tt.key, tt.expected)
			
			if tt.shouldError {
				if err == nil {
					t.Fatalf("Factory.Create() should have returned an error for unsupported type")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Factory.Create() error = %v", err)
			}

			headerAssertion, ok := assertion.(*HeaderAssertion)
			if !ok {
				t.Fatalf("Factory.Create() did not return a *HeaderAssertion")
			}

			if headerAssertion.HeaderName != tt.expectedName {
				t.Errorf("HeaderName = %v, want %v", headerAssertion.HeaderName, tt.expectedName)
			}

			if headerAssertion.ExpectedValue != tt.expectedValue {
				t.Errorf("ExpectedValue = %v, want %v", headerAssertion.ExpectedValue, tt.expectedValue)
			}
		})
	}
}