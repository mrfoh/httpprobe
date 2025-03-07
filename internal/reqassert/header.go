package reqassert

import (
	"fmt"
	"strings"
)

// HeaderAssertion validates HTTP headers
type HeaderAssertion struct {
	HeaderName     string
	ExpectedValue  string
}

// Validate checks if the header value matches the expected value
func (a *HeaderAssertion) Validate(ctx *AssertionContext) error {
	actualValue, exists := ctx.Headers[a.HeaderName]
	if !exists {
		return fmt.Errorf("header '%s' not found in response", a.HeaderName)
	}
	
	if strings.TrimSpace(actualValue) != strings.TrimSpace(a.ExpectedValue) {
		return fmt.Errorf("expected header '%s' to be '%s', got '%s'", 
			a.HeaderName, a.ExpectedValue, actualValue)
	}
	
	return nil
}

// HeaderAssertionFactory creates header assertions
type HeaderAssertionFactory struct{}

// Create returns a new HeaderAssertion
func (f *HeaderAssertionFactory) Create(key string, expected interface{}) (Assertion, error) {
	// Convert expected to string
	expectedValue, ok := expected.(string)
	if !ok {
		// Try converting other types to string
		switch v := expected.(type) {
		case int:
			expectedValue = fmt.Sprintf("%d", v)
		case float64:
			expectedValue = fmt.Sprintf("%g", v)
		case bool:
			expectedValue = fmt.Sprintf("%t", v)
		default:
			return nil, fmt.Errorf("header value must be convertible to string, got %T", expected)
		}
	}
	
	return &HeaderAssertion{
		HeaderName:    key,
		ExpectedValue: expectedValue,
	}, nil
}