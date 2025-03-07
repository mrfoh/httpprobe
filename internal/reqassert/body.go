package reqassert

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/oliveagle/jsonpath"
	"github.com/pkg/errors"
)

// BodyAssertion validates response body content using JSONPath
type BodyAssertion struct {
	JSONPath       string
	ExpectedValue  interface{}
	ComparisonType string // equals, contains, gt, lt, etc.
}

// Validate checks if the body value at JSONPath matches the expected value
func (a *BodyAssertion) Validate(ctx *AssertionContext) error {
	// Parse JSONPath
	path, err := jsonpath.Compile(a.JSONPath)
	if err != nil {
		return errors.Wrap(err, "invalid JSONPath")
	}

	// Extract value from response body
	actualValue, err := path.Lookup(ctx.BodyMap)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("JSONPath '%s' not found in response body", a.JSONPath)
		}
		return errors.Wrap(err, "error extracting value from JSONPath")
	}

	// Perform the appropriate comparison
	return a.compareValues(actualValue, a.ExpectedValue)
}

// compareValues compares the actual and expected values based on the comparison type
func (a *BodyAssertion) compareValues(actual, expected interface{}) error {
	if a.ComparisonType == "" {
		// Default to equals comparison
		if actual != expected {
			return fmt.Errorf("expected '%v', got '%v'", expected, actual)
		}
		return nil
	}

	// Handle different comparison types
	switch a.ComparisonType {
	case "equals", "=":
		if actual != expected {
			return fmt.Errorf("expected '%v', got '%v'", expected, actual)
		}
	case "contains":
		actualStr, ok := actual.(string)
		if !ok {
			return fmt.Errorf("'contains' comparison requires string value, got %T", actual)
		}
		expectedStr, ok := expected.(string)
		if !ok {
			return fmt.Errorf("'contains' comparison requires string value, got %T", expected)
		}
		if !strings.Contains(actualStr, expectedStr) {
			return fmt.Errorf("expected '%s' to contain '%s'", actualStr, expectedStr)
		}
	case "gt", ">":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a > b })
	case "gte", ">=":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a >= b })
	case "lt", "<":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a < b })
	case "lte", "<=":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a <= b })
	default:
		return fmt.Errorf("unknown comparison type: %s", a.ComparisonType)
	}

	return nil
}

// Helper function to compare numeric values
func compareNumeric(actual, expected interface{}, compare func(float64, float64) bool) error {
	// Convert actual to float64
	var actualNum float64
	switch v := actual.(type) {
	case int:
		actualNum = float64(v)
	case int64:
		actualNum = float64(v)
	case float64:
		actualNum = v
	case string:
		var err error
		actualNum, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot convert actual value '%s' to number", v)
		}
	default:
		return fmt.Errorf("expected numeric value, got %T", actual)
	}

	// Convert expected to float64
	var expectedNum float64
	switch v := expected.(type) {
	case int:
		expectedNum = float64(v)
	case int64:
		expectedNum = float64(v)
	case float64:
		expectedNum = v
	case string:
		var err error
		expectedNum, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("cannot convert expected value '%s' to number", v)
		}
	default:
		return fmt.Errorf("expected numeric value, got %T", expected)
	}

	if !compare(actualNum, expectedNum) {
		return fmt.Errorf("comparison failed: %v and %v", actual, expected)
	}

	return nil
}

// BodyAssertionFactory creates body assertions
type BodyAssertionFactory struct{}

// Create returns a new BodyAssertion
func (f *BodyAssertionFactory) Create(key string, expected interface{}) (Assertion, error) {
	// Check if key contains a comparison operator
	comparisonType := ""
	jsonPath := key

	// Check for comparison operators in the expected value if it's a string
	if expectedStr, ok := expected.(string); ok {
		// Regex to match comparison operators at the beginning of the string
		re := regexp.MustCompile(`^\s*(=|==|!=|>|>=|<|<=|contains)\s*(.+)$`)
		if matches := re.FindStringSubmatch(expectedStr); len(matches) > 0 {
			comparisonType = matches[1]
			expected = matches[2]
		}
	}

	return &BodyAssertion{
		JSONPath:       jsonPath,
		ExpectedValue:  expected,
		ComparisonType: comparisonType,
	}, nil
}
