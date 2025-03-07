package reqassert

import (
	"fmt"
)

// StatusAssertion validates HTTP status codes
type StatusAssertion struct {
	Expected int
}

// Validate checks if the status code matches the expected value
func (a *StatusAssertion) Validate(ctx *AssertionContext) error {
	if ctx.StatusCode != a.Expected {
		return fmt.Errorf("expected status code %d, got %d", a.Expected, ctx.StatusCode)
	}
	return nil
}

// StatusAssertionFactory creates status assertions
type StatusAssertionFactory struct{}

// Create returns a new StatusAssertion
func (f *StatusAssertionFactory) Create(key string, expected interface{}) (Assertion, error) {
	// For status assertions, key is ignored as there's only one status code
	
	// Convert expected to int
	statusCode, ok := expected.(int)
	if !ok {
		// Try to convert from float64 (common when parsing from JSON)
		if floatVal, isFloat := expected.(float64); isFloat {
			statusCode = int(floatVal)
		} else {
			return nil, fmt.Errorf("status code must be an integer, got %T", expected)
		}
	}
	
	return &StatusAssertion{Expected: statusCode}, nil
}