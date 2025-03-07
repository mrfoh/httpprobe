package reqassert

import "github.com/pkg/errors"

// Assertion defines the interface for validating HTTP responses
type Assertion interface {
	// Validate checks if the assertion passes against the given context
	Validate(ctx *AssertionContext) error
}

// AssertionContext contains all the data needed for assertions
type AssertionContext struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
	BodyMap    map[string]interface{}
}

// AssertionFactory creates assertions from data
type AssertionFactory interface {
	// Create returns an Assertion from the given data
	Create(key string, expected interface{}) (Assertion, error)
}

// Registry holds all available assertion factories
type Registry struct {
	factories map[string]AssertionFactory
}

// NewRegistry creates a new assertion registry
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]AssertionFactory),
	}
}

// Register adds a new assertion factory to the registry
func (r *Registry) Register(name string, factory AssertionFactory) {
	r.factories[name] = factory
}

// Create builds an assertion based on the type and expected value
func (r *Registry) Create(assertionType string, key string, expected interface{}) (Assertion, error) {
	factory, exists := r.factories[assertionType]
	if !exists {
		return nil, errors.Errorf("unknown assertion type: %s", assertionType)
	}
	
	return factory.Create(key, expected)
}