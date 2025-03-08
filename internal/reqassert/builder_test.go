package reqassert

import (
	"fmt"
	"strings"
	"testing"
)

// Helper function to check if an assertion's type contains a string
func typeContains(assertion Assertion, typeName string) bool {
	assertionType := fmt.Sprintf("%T", assertion)
	return strings.Contains(assertionType, typeName)
}

func TestBuilderBuildAssertions(t *testing.T) {
	tests := []struct {
		name           string
		assertionData  map[string]interface{}
		expectedCount  int
		shouldContain  []string
		shouldNotError bool
	}{
		{
			name: "status assertion",
			assertionData: map[string]interface{}{
				"status": 200,
			},
			expectedCount:  1,
			shouldContain:  []string{"StatusAssertion"},
			shouldNotError: true,
		},
		{
			name: "headers assertion",
			assertionData: map[string]interface{}{
				"headers": map[string]interface{}{
					"content-type": "application/json",
					"x-api-key":    "abc123",
				},
			},
			expectedCount:  2,
			shouldContain:  []string{"HeaderAssertion"},
			shouldNotError: true,
		},
		{
			name: "body assertion",
			assertionData: map[string]interface{}{
				"body": map[string]interface{}{
					"$.name":  "John",
					"$.age":   30,
					"$.email": "john@example.com",
				},
			},
			expectedCount:  3,
			shouldContain:  []string{"BodyAssertion"},
			shouldNotError: true,
		},
		{
			name: "combined assertions",
			assertionData: map[string]interface{}{
				"status": 200,
				"headers": map[string]interface{}{
					"content-type": "application/json",
				},
				"body": map[string]interface{}{
					"$.success": true,
				},
			},
			expectedCount:  3,
			shouldContain:  []string{"StatusAssertion", "HeaderAssertion", "BodyAssertion"},
			shouldNotError: true,
		},
		{
			name:           "empty assertions",
			assertionData:  map[string]interface{}{},
			expectedCount:  0,
			shouldContain:  []string{},
			shouldNotError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			assertions, err := builder.BuildAssertions(tt.assertionData)

			if tt.shouldNotError && err != nil {
				t.Fatalf("BuildAssertions() error = %v", err)
			}

			if len(assertions) != tt.expectedCount {
				t.Errorf("Expected %d assertions, got %d", tt.expectedCount, len(assertions))
			}

			// Check that the assertions are of the expected types
			for _, typeName := range tt.shouldContain {
				found := false
				for _, assertion := range assertions {
					if typeContains(assertion, typeName) {
						found = true
						break
					}
				}
				if !found && len(assertions) > 0 {
					t.Errorf("Expected to find assertion of type %s", typeName)
				}
			}
		})
	}
}

func TestBuilderPrepareContext(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		headers    map[string]string
		body       []byte
	}{
		{
			name:       "valid json body",
			statusCode: 200,
			headers: map[string]string{
				"content-type": "application/json",
			},
			body: []byte(`{"name":"John","age":30}`),
		},
		{
			name:       "empty body",
			statusCode: 204,
			headers: map[string]string{
				"content-type": "application/json",
			},
			body: []byte{},
		},
		{
			name:       "invalid json body",
			statusCode: 200,
			headers: map[string]string{
				"content-type": "text/plain",
			},
			body: []byte(`Not a JSON body`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			ctx, err := builder.PrepareContext(tt.statusCode, tt.headers, tt.body)

			if err != nil {
				t.Fatalf("PrepareContext() error = %v", err)
			}

			if ctx.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %v, want %v", ctx.StatusCode, tt.statusCode)
			}

			if len(ctx.Headers) != len(tt.headers) {
				t.Errorf("Headers length = %v, want %v", len(ctx.Headers), len(tt.headers))
			}

			if len(ctx.Body) != len(tt.body) {
				t.Errorf("Body length = %v, want %v", len(ctx.Body), len(tt.body))
			}

			// For valid JSON body, check that it was parsed correctly
			if tt.name == "valid json body" {
				if name, ok := ctx.BodyMap["name"]; !ok || name != "John" {
					t.Errorf("BodyMap[\"name\"] = %v, want %v", name, "John")
				}

				if age, ok := ctx.BodyMap["age"]; !ok || age != float64(30) {
					t.Errorf("BodyMap[\"age\"] = %v, want %v", age, float64(30))
				}
			}

			// For invalid JSON body, ensure an empty map is created
			if tt.name == "invalid json body" && len(ctx.BodyMap) != 0 {
				t.Errorf("Expected empty BodyMap for invalid JSON")
			}
		})
	}
}

func TestBuilderValidateAll(t *testing.T) {
	tests := []struct {
		name           string
		assertions     []Assertion
		context        *AssertionContext
		expectedErrors int
	}{
		{
			name: "all passing",
			assertions: []Assertion{
				&StatusAssertion{Expected: 200},
				&HeaderAssertion{HeaderName: "content-type", ExpectedValue: "application/json"},
				&BodyAssertion{JSONPath: "$.name", ExpectedValue: "John"},
			},
			context: &AssertionContext{
				StatusCode: 200,
				Headers: map[string]string{
					"content-type": "application/json",
				},
				BodyMap: map[string]interface{}{
					"name": "John",
				},
			},
			expectedErrors: 0,
		},
		{
			name: "some failing",
			assertions: []Assertion{
				&StatusAssertion{Expected: 200},
				&HeaderAssertion{HeaderName: "content-type", ExpectedValue: "application/json"},
				&BodyAssertion{JSONPath: "$.name", ExpectedValue: "Jane"},
			},
			context: &AssertionContext{
				StatusCode: 200,
				Headers: map[string]string{
					"content-type": "application/json",
				},
				BodyMap: map[string]interface{}{
					"name": "John",
				},
			},
			expectedErrors: 1,
		},
		{
			name: "all failing",
			assertions: []Assertion{
				&StatusAssertion{Expected: 201},
				&HeaderAssertion{HeaderName: "content-type", ExpectedValue: "text/plain"},
				&BodyAssertion{JSONPath: "$.age", ExpectedValue: 40},
			},
			context: &AssertionContext{
				StatusCode: 200,
				Headers: map[string]string{
					"content-type": "application/json",
				},
				BodyMap: map[string]interface{}{
					"age": float64(30),
				},
			},
			expectedErrors: 3,
		},
		{
			name:           "no assertions",
			assertions:     []Assertion{},
			context:        &AssertionContext{},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			errors := builder.ValidateAll(tt.assertions, tt.context)

			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errors))
				for i, err := range errors {
					t.Logf("Error %d: %v", i, err)
				}
			}
		})
	}
}
