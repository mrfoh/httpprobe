package reqassert

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/oliveagle/jsonpath"
	"github.com/xeipuuv/gojsonschema"
)

// GraphQLAssertionGroup is a composite assertion that wraps multiple GraphQL sub-assertions
type GraphQLAssertionGroup struct {
	Assertions []Assertion
}

// Validate runs all sub-assertions and collects errors
func (g *GraphQLAssertionGroup) Validate(ctx *AssertionContext) error {
	var errs []string
	for _, a := range g.Assertions {
		if err := a.Validate(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("graphql assertion failures: %s", strings.Join(errs, "; "))
	}
	return nil
}

// GraphQLNoErrorsAssertion checks that response.errors is empty or absent
type GraphQLNoErrorsAssertion struct {
	Expected bool
}

func (a *GraphQLNoErrorsAssertion) Validate(ctx *AssertionContext) error {
	errorsRaw, hasErrors := ctx.BodyMap["errors"]
	if !a.Expected {
		// We expect errors to be present
		if !hasErrors || errorsRaw == nil {
			return fmt.Errorf("graphql: expected errors but none found")
		}
		errorsSlice, ok := errorsRaw.([]interface{})
		if !ok || len(errorsSlice) == 0 {
			return fmt.Errorf("graphql: expected errors but none found")
		}
		return nil
	}

	// We expect no errors
	if !hasErrors || errorsRaw == nil {
		return nil
	}
	errorsSlice, ok := errorsRaw.([]interface{})
	if !ok || len(errorsSlice) == 0 {
		return nil
	}

	// Build error summary from the errors array
	var messages []string
	for _, e := range errorsSlice {
		if errMap, ok := e.(map[string]interface{}); ok {
			if msg, ok := errMap["message"].(string); ok {
				messages = append(messages, msg)
			}
		}
	}
	if len(messages) > 0 {
		return fmt.Errorf("graphql: expected no errors but got: %s", strings.Join(messages, "; "))
	}
	return fmt.Errorf("graphql: expected no errors but got %d error(s)", len(errorsSlice))
}

// GraphQLDataAssertion validates a JSONPath within response.data
type GraphQLDataAssertion struct {
	JSONPath       string
	ExpectedValue  interface{}
	ComparisonType string // equals, contains, gt, lt, etc.
	IsLengthCheck  bool   // true when asserting on the length of the value
}

func (a *GraphQLDataAssertion) Validate(ctx *AssertionContext) error {
	dataRaw, ok := ctx.BodyMap["data"]
	if !ok || dataRaw == nil {
		return fmt.Errorf("graphql data: response does not contain 'data' field")
	}

	dataMap, ok := dataRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("graphql data: 'data' field is not an object")
	}

	path, err := jsonpath.Compile(a.JSONPath)
	if err != nil {
		return fmt.Errorf("graphql data: invalid JSONPath '%s': %v", a.JSONPath, err)
	}

	actualValue, err := path.Lookup(dataMap)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("graphql data: JSONPath '%s' not found in response.data", a.JSONPath)
		}
		return fmt.Errorf("graphql data: error extracting '%s': %v", a.JSONPath, err)
	}

	// Handle length checks (e.g., "length > 0")
	if a.IsLengthCheck {
		var length int
		switch v := actualValue.(type) {
		case string:
			length = len(v)
		case []interface{}:
			length = len(v)
		case map[string]interface{}:
			length = len(v)
		default:
			return fmt.Errorf("graphql data: '%s' cannot check length of %T", a.JSONPath, actualValue)
		}
		bodyAssertion := &BodyAssertion{
			JSONPath:       a.JSONPath,
			ComparisonType: a.ComparisonType,
		}
		if err := bodyAssertion.compareValues(float64(length), a.ExpectedValue); err != nil {
			return fmt.Errorf("graphql data: '%s' length %d: %v", a.JSONPath, length, err)
		}
		return nil
	}

	// Check if expected value is a regex pattern (e.g., "/.+@.+/")
	if expectedStr, ok := a.ExpectedValue.(string); ok {
		if strings.HasPrefix(expectedStr, "/") && strings.HasSuffix(expectedStr, "/") && len(expectedStr) > 2 {
			pattern := expectedStr[1 : len(expectedStr)-1]
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("graphql data: invalid regex pattern '%s': %v", pattern, err)
			}
			actualStr := fmt.Sprintf("%v", actualValue)
			if !re.MatchString(actualStr) {
				return fmt.Errorf("graphql data: '%s' value '%s' does not match pattern '%s'", a.JSONPath, actualStr, pattern)
			}
			return nil
		}
	}

	// Delegate to BodyAssertion for comparison operator support
	bodyAssertion := &BodyAssertion{
		JSONPath:       a.JSONPath,
		ExpectedValue:  a.ExpectedValue,
		ComparisonType: a.ComparisonType,
	}
	if err := bodyAssertion.compareValues(actualValue, a.ExpectedValue); err != nil {
		return fmt.Errorf("graphql data: '%s' %v", a.JSONPath, err)
	}
	return nil
}

// GraphQLErrorAssertion matches a specific error in the response.errors array
type GraphQLErrorAssertion struct {
	ExpectedMessage    string
	ExpectedExtensions map[string]string
}

func (a *GraphQLErrorAssertion) Validate(ctx *AssertionContext) error {
	errorsRaw, ok := ctx.BodyMap["errors"]
	if !ok || errorsRaw == nil {
		return fmt.Errorf("graphql errors: expected error with message '%s' but no errors found", a.ExpectedMessage)
	}

	errorsSlice, ok := errorsRaw.([]interface{})
	if !ok || len(errorsSlice) == 0 {
		return fmt.Errorf("graphql errors: expected error with message '%s' but no errors found", a.ExpectedMessage)
	}

	for _, e := range errorsSlice {
		errMap, ok := e.(map[string]interface{})
		if !ok {
			continue
		}

		// Check message
		msg, _ := errMap["message"].(string)
		if a.ExpectedMessage != "" && msg != a.ExpectedMessage {
			continue
		}

		// Check extensions
		if len(a.ExpectedExtensions) > 0 {
			extensions, _ := errMap["extensions"].(map[string]interface{})
			if extensions == nil {
				continue
			}

			matched := true
			for key, expectedVal := range a.ExpectedExtensions {
				actualVal := fmt.Sprintf("%v", extensions[key])
				if actualVal != expectedVal {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Found a matching error
		return nil
	}

	return fmt.Errorf("graphql errors: no error matching message='%s' found", a.ExpectedMessage)
}

// GraphQLPartialDataAssertion checks whether partial data is allowed
type GraphQLPartialDataAssertion struct {
	AllowPartial bool
}

func (a *GraphQLPartialDataAssertion) Validate(ctx *AssertionContext) error {
	if a.AllowPartial {
		return nil
	}

	// partial_data: false means data must not be null when there are no errors
	dataRaw, hasData := ctx.BodyMap["data"]
	errorsRaw, hasErrors := ctx.BodyMap["errors"]

	dataIsNil := !hasData || dataRaw == nil
	hasActualErrors := false
	if hasErrors && errorsRaw != nil {
		if errSlice, ok := errorsRaw.([]interface{}); ok && len(errSlice) > 0 {
			hasActualErrors = true
		}
	}

	if dataIsNil && !hasActualErrors {
		return fmt.Errorf("graphql: data is null but no errors present (partial_data: false)")
	}

	if hasActualErrors && !dataIsNil {
		return fmt.Errorf("graphql: partial data detected — both data and errors are present (partial_data: false)")
	}

	return nil
}

// GraphQLDataSchemaAssertion validates response.data against a JSON Schema
type GraphQLDataSchemaAssertion struct {
	SchemaJSON string
}

func (a *GraphQLDataSchemaAssertion) Validate(ctx *AssertionContext) error {
	dataRaw, ok := ctx.BodyMap["data"]
	if !ok || dataRaw == nil {
		return fmt.Errorf("graphql data_schema: response does not contain 'data' field")
	}

	// Marshal data back to JSON for schema validation
	dataBytes, err := json.Marshal(dataRaw)
	if err != nil {
		return fmt.Errorf("graphql data_schema: error marshaling data: %v", err)
	}

	schemaLoader := gojsonschema.NewStringLoader(a.SchemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(dataBytes)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("graphql data_schema: schema validation error: %v", err)
	}

	if !result.Valid() {
		var errs []string
		for _, desc := range result.Errors() {
			errs = append(errs, desc.String())
		}
		return fmt.Errorf("graphql data_schema: validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GraphQLAssertionFactory creates GraphQL assertions from the assertion data map
type GraphQLAssertionFactory struct{}

func (f *GraphQLAssertionFactory) Create(key string, expected interface{}) (Assertion, error) {
	graphqlMap, ok := expected.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("graphql assertion value must be a map, got %T", expected)
	}

	group := &GraphQLAssertionGroup{}

	// no_errors
	if noErrors, ok := graphqlMap["no_errors"]; ok {
		if boolVal, ok := noErrors.(bool); ok {
			group.Assertions = append(group.Assertions, &GraphQLNoErrorsAssertion{Expected: boolVal})
		}
	}

	// partial_data
	if partialData, ok := graphqlMap["partial_data"]; ok {
		if boolVal, ok := partialData.(bool); ok {
			group.Assertions = append(group.Assertions, &GraphQLPartialDataAssertion{AllowPartial: boolVal})
		}
	}

	// data — JSONPath assertions relative to response.data
	if data, ok := graphqlMap["data"].(map[string]interface{}); ok {
		for jsonPath, expectedValue := range data {
			comparisonType := ""
			isLengthCheck := false
			if expectedStr, ok := expectedValue.(string); ok {
				// Check for "length" patterns first (e.g., "length > 0", "length 10")
				lengthRe := regexp.MustCompile(`^\s*length\s*(=|==|!=|>|>=|<|<=)?\s*(\d+)\s*$`)
				if matches := lengthRe.FindStringSubmatch(expectedStr); len(matches) > 0 {
					isLengthCheck = true
					comparisonType = matches[1]
					if comparisonType == "" {
						comparisonType = "="
					}
					expectedValue = matches[2]
				} else {
					// Parse comparison operators (same as BodyAssertionFactory)
					re := regexp.MustCompile(`^\s*(=|==|!=|>|>=|<|<=|contains)\s*(.+)$`)
					if matches := re.FindStringSubmatch(expectedStr); len(matches) > 0 {
						comparisonType = matches[1]
						expectedValue = matches[2]
					}
				}
			}
			group.Assertions = append(group.Assertions, &GraphQLDataAssertion{
				JSONPath:       jsonPath,
				ExpectedValue:  expectedValue,
				ComparisonType: comparisonType,
				IsLengthCheck:  isLengthCheck,
			})
		}
	}

	// errors — list of expected error matchers
	if errors, ok := graphqlMap["errors"].([]interface{}); ok {
		for _, errDef := range errors {
			errMap, ok := errDef.(map[string]interface{})
			if !ok {
				continue
			}

			assertion := &GraphQLErrorAssertion{}
			if msg, ok := errMap["message"].(string); ok {
				assertion.ExpectedMessage = msg
			}

			// Collect extensions.* keys
			extensions := make(map[string]string)
			for k, v := range errMap {
				if strings.HasPrefix(k, "extensions.") {
					extKey := strings.TrimPrefix(k, "extensions.")
					extensions[extKey] = fmt.Sprintf("%v", v)
				}
			}
			if len(extensions) > 0 {
				assertion.ExpectedExtensions = extensions
			}

			group.Assertions = append(group.Assertions, assertion)
		}
	}

	// data_schema — JSON Schema validation of response.data
	if schema, ok := graphqlMap["data_schema"].(string); ok {
		group.Assertions = append(group.Assertions, &GraphQLDataSchemaAssertion{
			SchemaJSON: schema,
		})
	}

	return group, nil
}
