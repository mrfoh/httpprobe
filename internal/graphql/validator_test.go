package graphql

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSchema = `
type Query {
  user(id: ID!): User
  users(limit: Int): [User!]!
  searchUsers(name: String!): [User!]!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
  login(email: String!, password: String!): AuthPayload!
}

type User {
  id: ID!
  name: String!
  email: String!
  age: Int
  posts: [Post!]!
}

type Post {
  id: ID!
  title: String!
  body: String!
  author: User!
}

type AuthPayload {
  token: String!
  user: User!
}

input CreateUserInput {
  name: String!
  email: String!
  age: Int
}
`

func writeTestSchemaFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.graphql")
	err := os.WriteFile(path, []byte(testSchema), 0644)
	require.NoError(t, err)
	return path
}

func TestFileSchemaLoader(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, schema)
}

func TestFileSchemaLoader_NotFound(t *testing.T) {
	loader := &FileSchemaLoader{Path: "/nonexistent/schema.graphql"}
	_, err := loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error reading schema file")
}

func TestFileSchemaLoader_InvalidSchema(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.graphql")
	err := os.WriteFile(path, []byte("this is not valid graphql {{{"), 0644)
	require.NoError(t, err)

	loader := &FileSchemaLoader{Path: path}
	_, err = loader.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing schema file")
}

func TestValidateQuery_Valid(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		query GetUser($id: ID!) {
			user(id: $id) {
				id
				name
				email
			}
		}
	`)
	assert.Empty(t, errors)
}

func TestValidateQuery_ValidMutation(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		mutation Login($email: String!, $password: String!) {
			login(email: $email, password: $password) {
				token
				user {
					id
					name
				}
			}
		}
	`)
	assert.Empty(t, errors)
}

func TestValidateQuery_InvalidField(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		query GetUser($id: ID!) {
			user(id: $id) {
				id
				nam
			}
		}
	`)
	require.NotEmpty(t, errors)
	assert.Contains(t, errors[0].Message, "Cannot query field")
}

func TestValidateQuery_MissingRequiredArgument(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		query {
			user {
				id
				name
			}
		}
	`)
	require.NotEmpty(t, errors)
}

func TestValidateQuery_WrongArgumentType(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		query {
			users(limit: "not a number") {
				id
			}
		}
	`)
	require.NotEmpty(t, errors)
}

func TestValidateQuery_ErrorLocation(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`query { user(id: "1") { nonexistent } }`)
	require.NotEmpty(t, errors)
	assert.Greater(t, errors[0].Line, 0)
	assert.Greater(t, errors[0].Column, 0)
}

func TestValidateQuery_MultipleErrors(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`
		query GetUser($id: ID!) {
			user(id: $id) {
				nam
				emai
			}
		}
	`)
	assert.GreaterOrEqual(t, len(errors), 2)
}

func TestValidateQuery_SyntaxError(t *testing.T) {
	path := writeTestSchemaFile(t)
	loader := &FileSchemaLoader{Path: path}
	schema, err := loader.Load()
	require.NoError(t, err)

	v := NewValidator(schema)
	errors := v.ValidateQuery(`query { user(id: "1") { id name `)
	require.NotEmpty(t, errors)
}

func TestNewSchemaLoader(t *testing.T) {
	tests := []struct {
		name   string
		source string
		valid  bool
	}{
		{"file", "file", true},
		{"url", "url", true},
		{"introspection", "introspection", true},
		{"unknown", "ftp", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, err := NewSchemaLoader(tt.source, "/some/path")
			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, loader)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Message: "Cannot query field \"nam\" on type \"User\"",
		Line:    4,
		Column:  9,
	}
	assert.Equal(t, "line 4, col 9: Cannot query field \"nam\" on type \"User\"", ve.Error())
}

func TestIntrospectionJSONToSDL(t *testing.T) {
	// Minimal introspection response
	jsonData := `{
		"data": {
			"__schema": {
				"queryType": {"name": "Query"},
				"mutationType": null,
				"subscriptionType": null,
				"types": [
					{
						"kind": "OBJECT",
						"name": "Query",
						"description": "",
						"fields": [
							{
								"name": "hello",
								"description": "",
								"args": [],
								"type": {"kind": "SCALAR", "name": "String", "ofType": null},
								"isDeprecated": false,
								"deprecationReason": null
							}
						],
						"inputFields": null,
						"interfaces": [],
						"enumValues": null,
						"possibleTypes": null
					},
					{
						"kind": "SCALAR",
						"name": "String",
						"description": "",
						"fields": null,
						"inputFields": null,
						"interfaces": null,
						"enumValues": null,
						"possibleTypes": null
					}
				],
				"directives": []
			}
		}
	}`

	sdl, err := introspectionJSONToSDL([]byte(jsonData))
	require.NoError(t, err)
	assert.Contains(t, sdl, "type Query")
	assert.Contains(t, sdl, "hello")
	assert.Contains(t, sdl, "String")
}

func TestIntrospectionJSONToSDL_InvalidJSON(t *testing.T) {
	_, err := introspectionJSONToSDL([]byte("not json"))
	assert.Error(t, err)
}
