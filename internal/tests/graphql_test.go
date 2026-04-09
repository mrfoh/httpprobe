package tests

import (
	"encoding/json"
	"testing"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphQLBody(t *testing.T) {
	tests := []struct {
		name     string
		body     RequestBody
		expected GraphQLEnvelope
	}{
		{
			name: "query only",
			body: RequestBody{
				Type:  "graphql",
				Query: "{ users { id name } }",
			},
			expected: GraphQLEnvelope{
				Query: "{ users { id name } }",
			},
		},
		{
			name: "query with variables",
			body: RequestBody{
				Type:  "graphql",
				Query: "query GetUser($id: ID!) { user(id: $id) { id name } }",
				Variables: map[string]interface{}{
					"id": "123",
				},
			},
			expected: GraphQLEnvelope{
				Query: "query GetUser($id: ID!) { user(id: $id) { id name } }",
				Variables: map[string]interface{}{
					"id": "123",
				},
			},
		},
		{
			name: "query with variables and operation name",
			body: RequestBody{
				Type:          "graphql",
				Query:         "query GetUser($id: ID!) { user(id: $id) { id name } }",
				Variables:     map[string]interface{}{"id": "123"},
				OperationName: "GetUser",
			},
			expected: GraphQLEnvelope{
				Query:         "query GetUser($id: ID!) { user(id: $id) { id name } }",
				Variables:     map[string]interface{}{"id": "123"},
				OperationName: "GetUser",
			},
		},
		{
			name: "empty operation name is omitted",
			body: RequestBody{
				Type:          "graphql",
				Query:         "{ users { id } }",
				OperationName: "",
			},
			expected: GraphQLEnvelope{
				Query: "{ users { id } }",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := BuildGraphQLBody(&tt.body)
			assert.Equal(t, tt.expected.Query, envelope.Query)
			assert.Equal(t, tt.expected.Variables, envelope.Variables)
			assert.Equal(t, tt.expected.OperationName, envelope.OperationName)
		})
	}
}

func TestBuildGraphQLBody_JSONSerialization(t *testing.T) {
	body := RequestBody{
		Type:  "graphql",
		Query: "query GetUser($id: ID!) { user(id: $id) { id name } }",
		Variables: map[string]interface{}{
			"id": "123",
		},
		OperationName: "GetUser",
	}

	envelope := BuildGraphQLBody(&body)
	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, body.Query, result["query"])
	assert.Equal(t, "GetUser", result["operationName"])
	assert.NotNil(t, result["variables"])
}

func TestBuildGraphQLBody_OmitsEmptyFields(t *testing.T) {
	body := RequestBody{
		Type:  "graphql",
		Query: "{ users { id } }",
	}

	envelope := BuildGraphQLBody(&body)
	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	_, hasVars := result["variables"]
	_, hasOpName := result["operationName"]
	assert.False(t, hasVars, "variables should be omitted when nil")
	assert.False(t, hasOpName, "operationName should be omitted when empty")
}

func TestInterpolateGraphQLRequest(t *testing.T) {
	variables := map[string]Variable{
		"user_id":     {Type: "string", Value: "abc-123"},
		"graphql_url": {Type: "string", Value: "https://api.example.com/graphql"},
		"op_name":     {Type: "string", Value: "FetchUser"},
	}

	request := Request{
		URL: "${graphql_url}",
		Body: RequestBody{
			Type:  "graphql",
			Query: "query FetchUser($id: ID!) { user(id: $id) { id name } }",
			Variables: map[string]interface{}{
				"id": "${user_id}",
			},
			OperationName: "${op_name}",
		},
	}

	err := InterpolateRequest(&request, variables)
	require.NoError(t, err)

	assert.Equal(t, "https://api.example.com/graphql", request.URL)
	assert.Equal(t, "query FetchUser($id: ID!) { user(id: $id) { id name } }", request.Body.Query)
	assert.Equal(t, "abc-123", request.Body.Variables["id"])
	assert.Equal(t, "FetchUser", request.Body.OperationName)
}

func TestInterpolateGraphQLRequest_QueryInterpolation(t *testing.T) {
	variables := map[string]Variable{
		"type_name": {Type: "string", Value: "User"},
	}

	request := Request{
		URL: "http://localhost/graphql",
		Body: RequestBody{
			Type:  "graphql",
			Query: "{ __type(name: \"${type_name}\") { name fields { name } } }",
		},
	}

	err := InterpolateRequest(&request, variables)
	require.NoError(t, err)

	assert.Contains(t, request.Body.Query, "\"User\"")
}

func TestInterpolateGraphQLRequest_NilVariables(t *testing.T) {
	variables := map[string]Variable{}

	request := Request{
		URL: "http://localhost/graphql",
		Body: RequestBody{
			Type:  "graphql",
			Query: "{ users { id } }",
		},
	}

	err := InterpolateRequest(&request, variables)
	require.NoError(t, err)
	assert.Equal(t, "{ users { id } }", request.Body.Query)
}

func TestProcessGraphQLExports(t *testing.T) {
	responseBody := `{
		"data": {
			"login": {
				"token": "jwt-abc-123",
				"user": {
					"id": "42",
					"name": "Alice"
				}
			}
		}
	}`

	request := Request{
		Export: RequestExport{
			GraphQL: []GraphQLExport{
				{Path: "$.login.token", As: "auth_token"},
				{Path: "$.login.user.id", As: "user_id"},
			},
		},
	}

	suite := &TestSuite{
		Variables: make(map[string]Variable),
	}

	resp := &easyreq.HttpResponse{
		Body: []byte(responseBody),
	}

	logger, _ := logging.NewLogger(&logging.LoggerOptions{LogLevel: "info"})
	err := processGraphQLExports(&request, resp, suite, logger)
	require.NoError(t, err)

	assert.Equal(t, "jwt-abc-123", suite.Variables["auth_token"].Value)
	assert.Equal(t, "42", suite.Variables["user_id"].Value)
}

func TestProcessGraphQLExports_MissingDataField(t *testing.T) {
	responseBody := `{"errors": [{"message": "Unauthorized"}]}`

	request := Request{
		Export: RequestExport{
			GraphQL: []GraphQLExport{
				{Path: "$.token", As: "auth_token"},
			},
		},
	}

	suite := &TestSuite{
		Variables: make(map[string]Variable),
	}

	resp := &easyreq.HttpResponse{
		Body: []byte(responseBody),
	}

	logger, _ := logging.NewLogger(&logging.LoggerOptions{LogLevel: "info"})
	err := processGraphQLExports(&request, resp, suite, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain 'data' field")
}

func TestProcessGraphQLExports_NullDataField(t *testing.T) {
	responseBody := `{"data": null, "errors": [{"message": "Not found"}]}`

	request := Request{
		Export: RequestExport{
			GraphQL: []GraphQLExport{
				{Path: "$.user.id", As: "user_id"},
			},
		},
	}

	suite := &TestSuite{
		Variables: make(map[string]Variable),
	}

	resp := &easyreq.HttpResponse{
		Body: []byte(responseBody),
	}

	logger, _ := logging.NewLogger(&logging.LoggerOptions{LogLevel: "info"})
	err := processGraphQLExports(&request, resp, suite, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain 'data' field")
}

func TestProcessGraphQLExports_EmptyExports(t *testing.T) {
	request := Request{
		Export: RequestExport{
			GraphQL: []GraphQLExport{},
		},
	}

	suite := &TestSuite{}
	resp := &easyreq.HttpResponse{Body: []byte(`{"data": {"id": "1"}}`)}

	logger, _ := logging.NewLogger(&logging.LoggerOptions{LogLevel: "info"})
	err := processGraphQLExports(&request, resp, suite, logger)
	assert.NoError(t, err)
}

func TestProcessGraphQLExports_NumericValue(t *testing.T) {
	responseBody := `{"data": {"user": {"age": 30, "active": true}}}`

	request := Request{
		Export: RequestExport{
			GraphQL: []GraphQLExport{
				{Path: "$.user.age", As: "user_age"},
				{Path: "$.user.active", As: "user_active"},
			},
		},
	}

	suite := &TestSuite{
		Variables: make(map[string]Variable),
	}

	resp := &easyreq.HttpResponse{
		Body: []byte(responseBody),
	}

	logger, _ := logging.NewLogger(&logging.LoggerOptions{LogLevel: "info"})
	err := processGraphQLExports(&request, resp, suite, logger)
	require.NoError(t, err)

	assert.Equal(t, "30", suite.Variables["user_age"].Value)
	assert.Equal(t, "true", suite.Variables["user_active"].Value)
}
