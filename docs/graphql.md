---
layout: default
title: GraphQL Testing
nav_order: 5
description: "Learn how to test GraphQL APIs with HttpProbe using declarative YAML/JSON definitions."
---

# GraphQL Testing
{: .no_toc }

HttpProbe provides first-class support for testing GraphQL APIs. You can write e2e tests for GraphQL queries and mutations using the same declarative YAML/JSON format used for REST, with dedicated handling for GraphQL's unique request and response patterns.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

GraphQL APIs differ from REST APIs in several important ways:

| Concern | REST | GraphQL |
| ------- | ---- | ------- |
| Endpoint | Many URLs | Single endpoint (e.g. `/graphql`) |
| HTTP method | GET / POST / PUT / DELETE | Always POST |
| Request body | Arbitrary JSON | `{ "query": "...", "variables": {} }` |
| Success indicator | HTTP status code | HTTP 200 + absence of `errors` in body |
| Errors | HTTP 4xx/5xx | HTTP 200 with `{ "errors": [...] }` |

HttpProbe handles these differences automatically, so you can focus on writing your queries and validating the results.

## GraphQL Body Type

To send a GraphQL request, set the body type to `graphql`:

```yaml
request:
  url: "https://api.example.com/graphql"
  body:
    type: graphql
    query: |
      query GetUser($id: ID!) {
        user(id: $id) {
          id
          name
          email
        }
      }
    variables:
      id: "${user_id}"
    operation_name: GetUser
```

### Fields

| Field | Required | Description |
| ----- | -------- | ----------- |
| `type` | Yes | Must be `graphql` |
| `query` | Yes | The GraphQL query or mutation string |
| `variables` | No | A map of variables to pass to the query |
| `operation_name` | No | The name of the operation to execute (useful when the query contains multiple operations) |

### Automatic Behavior

When `body.type` is set to `graphql`, HttpProbe automatically:

- Sets the HTTP method to **POST** (if not explicitly specified)
- Sets the `Content-Type` header to `application/json`
- Serializes the query, variables, and operation name into the standard GraphQL JSON envelope:

```json
{
  "query": "query GetUser($id: ID!) { user(id: $id) { id name email } }",
  "variables": { "id": "123" },
  "operationName": "GetUser"
}
```

You do not need to construct this envelope yourself.

### Variable Interpolation

HttpProbe's standard variable interpolation applies to all GraphQL body fields:

```yaml
variables:
  graphql_url:
    type: string
    value: "https://api.example.com/graphql"
  user_id:
    type: string
    value: "123"

suites:
  - name: GraphQL Tests
    cases:
      - title: Get User
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              query GetUser($id: ID!) {
                user(id: $id) { id name email }
              }
            variables:
              id: "${user_id}"
```

Environment variables, dynamic functions, and exported variables all work the same way as they do with REST requests. See [Variable Interpolation](variable-interpolation) for details.

## GraphQL Assertions

Because GraphQL always returns HTTP 200 even for errors, `status: 200` alone is not sufficient for validating responses. HttpProbe provides a dedicated `graphql` assertion block that understands GraphQL response structure:

```yaml
assertions:
  status: 200
  graphql:
    no_errors: true
    data:
      "$.user.id": "${user_id}"
      "$.user.email": "/.+@.+/"
```

### no_errors

Checks that the response does not contain a `errors` array (or that it is empty). This is the go-to assertion for happy-path tests:

```yaml
graphql:
  no_errors: true   # Fails if response.errors is non-empty
```

Set to `false` for negative tests where you expect errors:

```yaml
graphql:
  no_errors: false   # Fails if response.errors is empty or absent
```

### data

Validates values within `response.data` using JSONPath. All paths are relative to `data`, not the full response body:

```yaml
graphql:
  data:
    "$.user.id": "123"
    "$.user.name": "Alice"
    "$.user.email": "/.+@.+/"    # Regex pattern matching
```

Regex patterns are wrapped in forward slashes (`/pattern/`).

### errors

For negative tests, you can assert that specific errors are present by matching their `message` and `extensions` fields:

```yaml
graphql:
  errors:
    - message: "User not found"
      extensions.code: "NOT_FOUND"
```

Each entry matches against the `errors` array in the response. The test passes if at least one error in the array matches all specified fields.

### partial_data

Controls whether partial responses (both `data` and `errors` present) are allowed:

```yaml
graphql:
  partial_data: false    # Fail if data and errors coexist (default: false)
```

Set to `true` to allow partial success responses where both `data` and `errors` are present.

### data_schema

Validates `response.data` against a JSON Schema:

```yaml
graphql:
  data_schema: |
    {
      "type": "object",
      "required": ["user"],
      "properties": {
        "user": {
          "type": "object",
          "required": ["id", "name", "email"],
          "properties": {
            "id": { "type": "string" },
            "name": { "type": "string" },
            "email": { "type": "string" }
          }
        }
      }
    }
```

The schema is applied to `response.data`, not the full response body.

### Combining Assertions

You can combine multiple GraphQL assertions in a single test case, alongside standard assertions:

```yaml
assertions:
  status: 200
  graphql:
    no_errors: true
    data:
      "$.user.id": "${user_id}"
    data_schema: |
      { "type": "object", "required": ["user"] }
```

## GraphQL Exports

HttpProbe provides a `graphql` export shorthand where JSONPaths are relative to `response.data`:

```yaml
export:
  graphql:
    - path: "$.login.token"
      as: "auth_token"
    - path: "$.login.user.id"
      as: "user_id"
```

This is equivalent to using the standard `body` export with full paths:

```yaml
export:
  body:
    - path: "$.data.login.token"
      as: "auth_token"
```

The `graphql` shorthand is more readable and avoids repeating `$.data.` on every path. Both export styles can be used in the same test case.

Exported variables are available to subsequent test cases in the same suite, just like REST body exports. See [Variable Interpolation](variable-interpolation#exporting-response-values-as-variables) for more details.

## Pre-flight Query Validation

HttpProbe can validate all GraphQL queries in a test file against your API's schema before sending any HTTP requests. If any query is invalid, HttpProbe reports all violations with context and exits without running any tests.

### Configuration

Add a top-level `graphql` block to your test definition:

```yaml
graphql:
  schema:
    source: file
    path: "./schema.graphql"
  validation:
    enabled: true
    on_error: fail
```

### Schema Sources

HttpProbe supports three schema source modes:

| Source | Description | Best for |
| ------ | ----------- | -------- |
| `file` | Load from a local `.graphql` SDL file | CI (deterministic, no network dependency) |
| `introspection` | Fetch by sending a standard introspection query to the API | Local development |
| `url` | Fetch a raw SDL document from an HTTP URL | Shared schema registries |

#### File

```yaml
graphql:
  schema:
    source: file
    path: "./schema.graphql"
```

#### Introspection

```yaml
graphql:
  schema:
    source: introspection
    path: "https://api.example.com/graphql"
```

#### URL

```yaml
graphql:
  schema:
    source: url
    path: "https://schema.example.com/api.graphql"
```

### Validation Behavior

| `on_error` | Behavior |
| ---------- | -------- |
| `fail` | Report all errors and abort without running any tests (default) |
| `warn` | Log warnings but continue execution |

### What Gets Validated

Pre-flight validation checks for:

- Field existence on parent types
- Type compatibility of variables and arguments
- Required arguments without defaults
- Leaf selection correctness (scalars vs objects)
- Fragment validity and variable usage
- Directive placement
- Syntax errors in the query

### Error Output

When validation fails, HttpProbe reports each error with the suite name, case title, and error details:

```
PRE-FLIGHT VALIDATION — 1 error(s) found

  [suite: User Queries / case: Fetch user by ID]
    ERROR: line 4, col 9: Cannot query field "nam" on type "User". Did you mean "name"?

Aborting. Fix the queries above and re-run.
```

## Complete Examples

### Simple Query Test

```yaml
name: "GraphQL User API Tests"
description: "Tests for the GraphQL user API"
variables:
  graphql_url:
    type: string
    value: "https://api.example.com/graphql"
  user_id:
    type: string
    value: "123"
suites:
  - name: "User Queries"
    cases:
      - title: "Get User by ID"
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              query GetUser($id: ID!) {
                user(id: $id) {
                  id
                  name
                  email
                }
              }
            variables:
              id: "${user_id}"
            operation_name: GetUser
          assertions:
            status: 200
            graphql:
              no_errors: true
              data:
                "$.user.id": "${user_id}"
                "$.user.name": "Alice"
```

### Mutation with Token Export

```yaml
name: "GraphQL Authentication"
description: "Test login mutation and use the token in subsequent queries"
variables:
  graphql_url:
    type: string
    value: "https://api.example.com/graphql"
suites:
  - name: "Auth Flow"
    cases:
      - title: "Login"
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              mutation Login($email: String!, $password: String!) {
                login(email: $email, password: $password) {
                  token
                  user {
                    id
                    name
                  }
                }
              }
            variables:
              email: "alice@example.com"
              password: "${env:TEST_PASSWORD}"
          assertions:
            status: 200
            graphql:
              no_errors: true
              data:
                "$.login.token": "/.+/"
          export:
            graphql:
              - path: "$.login.token"
                as: "auth_token"
              - path: "$.login.user.id"
                as: "user_id"

      - title: "Get Profile with Token"
        request:
          url: "${graphql_url}"
          headers:
            - key: Authorization
              value: "Bearer ${auth_token}"
          body:
            type: graphql
            query: |
              query GetProfile {
                me {
                  id
                  name
                  email
                  role
                }
              }
          assertions:
            status: 200
            graphql:
              no_errors: true
              data:
                "$.me.id": "${user_id}"
```

### Error Handling Test

```yaml
name: "GraphQL Error Tests"
description: "Test GraphQL error responses"
variables:
  graphql_url:
    type: string
    value: "https://api.example.com/graphql"
suites:
  - name: "Error Cases"
    cases:
      - title: "Query nonexistent user"
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              query GetUser($id: ID!) {
                user(id: $id) {
                  id
                  name
                }
              }
            variables:
              id: "nonexistent-id"
          assertions:
            status: 200
            graphql:
              no_errors: false
              errors:
                - message: "User not found"
                  extensions.code: "NOT_FOUND"

      - title: "Unauthorized access"
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              query {
                adminDashboard {
                  totalUsers
                }
              }
          assertions:
            status: 200
            graphql:
              no_errors: false
              errors:
                - message: "Unauthorized"
                  extensions.code: "FORBIDDEN"
```

### Pre-flight Validation with File Schema

```yaml
name: "GraphQL Tests with Schema Validation"
description: "Validates queries against a local schema before running"

graphql:
  schema:
    source: file
    path: "./schema.graphql"
  validation:
    enabled: true
    on_error: fail

variables:
  graphql_url:
    type: string
    value: "https://api.example.com/graphql"

suites:
  - name: "Validated Queries"
    cases:
      - title: "List Users"
        request:
          url: "${graphql_url}"
          body:
            type: graphql
            query: |
              query ListUsers($limit: Int) {
                users(limit: $limit) {
                  id
                  name
                  email
                }
              }
            variables:
              limit: 10
          assertions:
            status: 200
            graphql:
              no_errors: true
              data_schema: |
                {
                  "type": "object",
                  "required": ["users"],
                  "properties": {
                    "users": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "required": ["id", "name", "email"]
                      }
                    }
                  }
                }
```

## Best Practices

1. **Always assert `no_errors: true`** for happy-path tests rather than relying on `status: 200` alone
2. **Use `graphql` exports** instead of `body` exports for cleaner paths (`$.login.token` vs `$.data.login.token`)
3. **Enable pre-flight validation** with `source: file` in CI to catch query errors before they hit the network
4. **Use `on_error: warn`** during development to see validation warnings without blocking execution
5. **Test error cases explicitly** using `no_errors: false` and `errors` matchers to verify your API returns the right error codes
6. **Use `data_schema`** for comprehensive structure validation on critical queries
7. **Keep queries in separate test files** grouped by feature area, using hooks for shared authentication
