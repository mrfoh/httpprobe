package reqassert

import (
	"encoding/json"
)

// Builder creates assertions from test definition structures
type Builder struct {
	registry *Registry
}

// NewBuilder creates a new assertion builder with default assertion types
func NewBuilder() *Builder {
	registry := NewRegistry()
	
	// Register built-in assertion factories
	registry.Register("status", &StatusAssertionFactory{})
	registry.Register("headers", &HeaderAssertionFactory{})
	registry.Register("body", &BodyAssertionFactory{})
	
	return &Builder{
		registry: registry,
	}
}

// RegisterType adds a new assertion type
func (b *Builder) RegisterType(name string, factory AssertionFactory) {
	b.registry.Register(name, factory)
}

// BuildAssertions creates a list of assertions from the assertion data
func (b *Builder) BuildAssertions(assertionData map[string]interface{}) ([]Assertion, error) {
	var assertions []Assertion
	
	// Process status assertion
	if status, ok := assertionData["status"]; ok {
		assertion, err := b.registry.Create("status", "", status)
		if err != nil {
			return nil, err
		}
		assertions = append(assertions, assertion)
	}
	
	// Process header assertions
	if headers, ok := assertionData["headers"].(map[string]interface{}); ok {
		for headerName, expectedValue := range headers {
			assertion, err := b.registry.Create("headers", headerName, expectedValue)
			if err != nil {
				return nil, err
			}
			assertions = append(assertions, assertion)
		}
	}
	
	// Process body assertions
	if body, ok := assertionData["body"].(map[string]interface{}); ok {
		for jsonPath, expectedValue := range body {
			assertion, err := b.registry.Create("body", jsonPath, expectedValue)
			if err != nil {
				return nil, err
			}
			assertions = append(assertions, assertion)
		}
	}
	
	return assertions, nil
}

// PrepareContext creates an AssertionContext from response data
func (b *Builder) PrepareContext(statusCode int, headers map[string]string, body []byte) (*AssertionContext, error) {
	// Create body map from JSON response
	var bodyMap map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &bodyMap); err != nil {
			// If JSON parsing fails, create an empty map
			bodyMap = make(map[string]interface{})
		}
	} else {
		bodyMap = make(map[string]interface{})
	}
	
	return &AssertionContext{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
		BodyMap:    bodyMap,
	}, nil
}

// ValidateAll validates all assertions against the given context
func (b *Builder) ValidateAll(assertions []Assertion, ctx *AssertionContext) []error {
	var errors []error
	
	for _, assertion := range assertions {
		if err := assertion.Validate(ctx); err != nil {
			errors = append(errors, err)
		}
	}
	
	return errors
}