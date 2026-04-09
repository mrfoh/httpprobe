package reqassert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func graphqlContext(bodyJSON string) *AssertionContext {
	ctx, _ := (&Builder{}).PrepareContext(200, nil, []byte(bodyJSON))
	return ctx
}

func TestGraphQLNoErrorsAssertion_Pass(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}}`)
	a := &GraphQLNoErrorsAssertion{Expected: true}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLNoErrorsAssertion_FailWithErrors(t *testing.T) {
	ctx := graphqlContext(`{"data": null, "errors": [{"message": "Unauthorized"}]}`)
	a := &GraphQLNoErrorsAssertion{Expected: true}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestGraphQLNoErrorsAssertion_EmptyErrorsArray(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}, "errors": []}`)
	a := &GraphQLNoErrorsAssertion{Expected: true}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLNoErrorsAssertion_ExpectErrors(t *testing.T) {
	ctx := graphqlContext(`{"data": null, "errors": [{"message": "Not found"}]}`)
	a := &GraphQLNoErrorsAssertion{Expected: false}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLNoErrorsAssertion_ExpectErrorsButNone(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}}`)
	a := &GraphQLNoErrorsAssertion{Expected: false}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected errors but none found")
}

func TestGraphQLDataAssertion_Pass(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "123", "name": "Alice"}}}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.id", ExpectedValue: "123"}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLDataAssertion_Fail(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "456"}}}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.id", ExpectedValue: "123"}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected '123', got '456'")
}

func TestGraphQLDataAssertion_MissingPath(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.email", ExpectedValue: "test@example.com"}
	err := a.Validate(ctx)
	assert.Error(t, err)
}

func TestGraphQLDataAssertion_MissingDataField(t *testing.T) {
	ctx := graphqlContext(`{"errors": [{"message": "error"}]}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.id", ExpectedValue: "1"}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain 'data' field")
}

func TestGraphQLDataAssertion_Regex(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"email": "alice@example.com"}}}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.email", ExpectedValue: "/.+@.+/"}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLDataAssertion_RegexNoMatch(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"email": "not-an-email"}}}`)
	a := &GraphQLDataAssertion{JSONPath: "$.user.email", ExpectedValue: "/.+@.+/"}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not match pattern")
}

func TestGraphQLErrorAssertion_Pass(t *testing.T) {
	ctx := graphqlContext(`{
		"data": null,
		"errors": [
			{"message": "User not found", "extensions": {"code": "NOT_FOUND"}}
		]
	}`)
	a := &GraphQLErrorAssertion{
		ExpectedMessage:    "User not found",
		ExpectedExtensions: map[string]string{"code": "NOT_FOUND"},
	}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLErrorAssertion_MessageOnly(t *testing.T) {
	ctx := graphqlContext(`{"errors": [{"message": "Unauthorized"}]}`)
	a := &GraphQLErrorAssertion{ExpectedMessage: "Unauthorized"}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLErrorAssertion_NoMatch(t *testing.T) {
	ctx := graphqlContext(`{"errors": [{"message": "Server error"}]}`)
	a := &GraphQLErrorAssertion{ExpectedMessage: "Not found"}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no error matching")
}

func TestGraphQLErrorAssertion_NoErrors(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}}`)
	a := &GraphQLErrorAssertion{ExpectedMessage: "Some error"}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no errors found")
}

func TestGraphQLPartialDataAssertion_NoPartialAllowed(t *testing.T) {
	// Both data and errors present — partial data
	ctx := graphqlContext(`{
		"data": {"user": {"id": "1"}},
		"errors": [{"message": "Partial failure"}]
	}`)
	a := &GraphQLPartialDataAssertion{AllowPartial: false}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial data detected")
}

func TestGraphQLPartialDataAssertion_AllowPartial(t *testing.T) {
	ctx := graphqlContext(`{
		"data": {"user": {"id": "1"}},
		"errors": [{"message": "Partial failure"}]
	}`)
	a := &GraphQLPartialDataAssertion{AllowPartial: true}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLPartialDataAssertion_DataOnly(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1"}}}`)
	a := &GraphQLPartialDataAssertion{AllowPartial: false}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLPartialDataAssertion_NullDataNoErrors(t *testing.T) {
	ctx := graphqlContext(`{"data": null}`)
	a := &GraphQLPartialDataAssertion{AllowPartial: false}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data is null")
}

func TestGraphQLDataSchemaAssertion_Pass(t *testing.T) {
	ctx := graphqlContext(`{"data": {"user": {"id": "1", "name": "Alice"}}}`)
	schema := `{"type": "object", "required": ["user"], "properties": {"user": {"type": "object"}}}`
	a := &GraphQLDataSchemaAssertion{SchemaJSON: schema}
	assert.NoError(t, a.Validate(ctx))
}

func TestGraphQLDataSchemaAssertion_Fail(t *testing.T) {
	ctx := graphqlContext(`{"data": {"product": {"id": "1"}}}`)
	schema := `{"type": "object", "required": ["user"]}`
	a := &GraphQLDataSchemaAssertion{SchemaJSON: schema}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestGraphQLDataSchemaAssertion_MissingData(t *testing.T) {
	ctx := graphqlContext(`{"errors": [{"message": "error"}]}`)
	a := &GraphQLDataSchemaAssertion{SchemaJSON: `{"type": "object"}`}
	err := a.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain 'data' field")
}

func TestGraphQLAssertionFactory_FullConfig(t *testing.T) {
	factory := &GraphQLAssertionFactory{}

	config := map[string]interface{}{
		"no_errors":    true,
		"partial_data": false,
		"data": map[string]interface{}{
			"$.user.id":    "123",
			"$.user.email": "/.+@.+/",
		},
		"data_schema": `{"type": "object", "required": ["user"]}`,
	}

	assertion, err := factory.Create("", config)
	require.NoError(t, err)

	group, ok := assertion.(*GraphQLAssertionGroup)
	require.True(t, ok)
	// no_errors + partial_data + 2 data paths + data_schema = 5
	assert.Equal(t, 5, len(group.Assertions))
}

func TestGraphQLAssertionFactory_WithErrors(t *testing.T) {
	factory := &GraphQLAssertionFactory{}

	config := map[string]interface{}{
		"errors": []interface{}{
			map[string]interface{}{
				"message":         "User not found",
				"extensions.code": "NOT_FOUND",
			},
		},
	}

	assertion, err := factory.Create("", config)
	require.NoError(t, err)

	group, ok := assertion.(*GraphQLAssertionGroup)
	require.True(t, ok)
	assert.Equal(t, 1, len(group.Assertions))
}

func TestGraphQLAssertionFactory_InvalidType(t *testing.T) {
	factory := &GraphQLAssertionFactory{}
	_, err := factory.Create("", "not a map")
	assert.Error(t, err)
}

func TestGraphQLAssertionGroup_Integration(t *testing.T) {
	// End-to-end test: build from factory, validate against a response
	factory := &GraphQLAssertionFactory{}
	config := map[string]interface{}{
		"no_errors": true,
		"data": map[string]interface{}{
			"$.user.id":   "42",
			"$.user.name": "Alice",
		},
	}

	assertion, err := factory.Create("", config)
	require.NoError(t, err)

	ctx := graphqlContext(`{
		"data": {
			"user": {
				"id": "42",
				"name": "Alice"
			}
		}
	}`)

	assert.NoError(t, assertion.Validate(ctx))
}

func TestGraphQLAssertionGroup_IntegrationFail(t *testing.T) {
	factory := &GraphQLAssertionFactory{}
	config := map[string]interface{}{
		"no_errors": true,
		"data": map[string]interface{}{
			"$.user.id": "42",
		},
	}

	assertion, err := factory.Create("", config)
	require.NoError(t, err)

	ctx := graphqlContext(`{
		"data": {"user": {"id": "99"}},
		"errors": [{"message": "Partial error"}]
	}`)

	err = assertion.Validate(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "graphql assertion failures")
}

func TestBuilderRegistersGraphQL(t *testing.T) {
	builder := NewBuilder()

	assertionData := map[string]interface{}{
		"status": 200,
		"graphql": map[string]interface{}{
			"no_errors": true,
		},
	}

	assertions, err := builder.BuildAssertions(assertionData)
	require.NoError(t, err)
	assert.Equal(t, 2, len(assertions)) // status + graphql
}
