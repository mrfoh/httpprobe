---
layout: default
title: Test Definitions
nav_order: 2
description: "Learn how to structure HttpProbe test definition files."
---

# Test Definitions
{: .no_toc }

Test definitions are the core of HttpProbe. They define your API tests in a declarative way, allowing you to focus on what you want to test rather than how to test it. Test definitions support hooks to run other tests before and after test execution, creating powerful, reusable test workflows.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Basic Structure

A test definition file contains the following main sections:

```yaml
name: "Test Definition Name"
description: "Description of the test definition"
variables:
  # Variables to use in the test cases
  variable_name:
    type: string
    value: "variable value"

# Hooks to run before and after test execution
before_all:
  - "path/to/setup.test.yaml"
after_all:
  - "path/to/cleanup.test.yaml"
before_each:
  - "path/to/auth.test.yaml"
after_each:
  - "path/to/logging.test.yaml"

suites:
  # Test suites
  - name: "Suite Name"
    cases:
      # Test cases
      - title: "Test Case Title"
        request:
          # Request details
          method: GET
          url: "https://api.example.com/endpoint"
          headers:
            - key: "Header-Name"
              value: "Header Value"
          body:
            type: json
            data: null  # or JSON object/array
          assertions:
            # Assertions to validate the response
            status: 200
            # Other assertions...
          export:
            # Export values from the response for use in subsequent tests
            body:
              - path: "$.token"
                as: "auth_token"
```

## File Formats

HttpProbe supports test definitions in two formats:

- **YAML** (.yaml, .yml) - Recommended for readability
- **JSON** (.json) - Useful for programmatic generation

## Top-Level Elements

### Name and Description

```yaml
name: "User API Tests"
description: "Tests for the user management API endpoints"
```

### Hooks

Hooks allow you to run other test definitions at specific points in the test lifecycle:

```yaml
# Run before any test suites in this definition
before_all:
  - "setup/initialize.test.yaml"

# Run after all test suites in this definition have completed
after_all:
  - "cleanup/teardown.test.yaml"

# Run before each test suite in this definition
before_each:
  - "common/auth.test.yaml"

# Run after each test suite in this definition
after_each:
  - "common/logging.test.yaml"
```

Hooks are useful for:
- Setting up test data before tests run
- Cleaning up resources after tests complete
- Obtaining authentication tokens before each test suite
- Logging test results after each test suite

Variables exported by hooks are available to subsequent tests. For example, if a `before_each` hook exports an authentication token, that token is available in the test suite that follows.

The `name` is used in test reports and logs to identify the test definition. The `description` provides additional information about the purpose of the tests.

### Variables

Variables allow you to define reusable values that can be referenced in your test cases.

```yaml
variables:
  base_url:
    type: string
    value: "https://api.example.com/v1"
  admin_token:
    type: string
    value: "${env:ADMIN_TOKEN}"  # From environment variable
  request_id:
    type: string
    value: "${random(16)}"  # Generated random string
```

Variables can be referenced in URLs, headers, and request bodies using the syntax `${variable_name}`.

### Suites

Test suites group related test cases together.

```yaml
suites:
  - name: "User Management"
    config:
      concurrent: true  # Run test cases in this suite concurrently
    cases:
      # Test cases for user management
  - name: "Authentication"
    cases:
      # Test cases for authentication
```

#### Suite Configuration

The optional `config` section allows you to configure suite-specific behavior:

```yaml
config:
  concurrent: true  # Run test cases in this suite concurrently
```

Available configuration options:

- `concurrent`: When set to `true`, test cases in the suite will run concurrently instead of sequentially. This can significantly improve performance when test cases are independent, but should be used carefully if test cases depend on each other or export variables that other test cases need. See the [Concurrency](concurrency) documentation for more details.

## Test Cases

Each test case represents a single API request with its assertions.

```yaml
- title: "Get User Profile"
  request:
    method: GET
    url: "${base_url}/users/123"
    headers:
      - key: Authorization
        value: Bearer ${admin_token}
      - key: Accept
        value: application/json
    body:
      type: json
      data: null
    assertions:
      status: 200
      headers:
        content-type: "application/json; charset=utf-8"
      body:
        "$.id": 123
        "$.email": "user@example.com"
```

### Request

The `request` section defines the HTTP request to be made:

- `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `url`: The endpoint URL (can include variables)
- `headers`: List of HTTP headers to include
- `body`: Request body (if applicable)

#### Request Body

For request bodies, you must specify:

- `type`: The content type (currently supported: `json`, `text`)
- `data`: The actual body content

For JSON bodies, you can specify the data in several ways:

```yaml
# Inline JSON object
body:
  type: json
  data:
    name: "John Doe"
    email: "john@example.com"

# JSON string with variables
body:
  type: json
  data: |
    {
      "name": "John Doe",
      "email": "john@example.com",
      "token": "${api_token}"
    }

# No body
body:
  type: json
  data: null
```

### Assertions

The `assertions` section defines the expected response:

```yaml
assertions:
  # Status code assertion
  status: 200
  
  # Header assertions
  headers:
    content-type: "application/json; charset=utf-8"
    cache-control: "no-cache"
  
  # Body assertions using JSONPath
  body:
    "$.id": 123
    "$.name": "John Doe"
    "$.roles[0]": "admin"

  # Schema assertion
  schema: |
    {
      "type": "object",
      "required": ["id", "name", "email"],
      "properties": {
        "id": { "type": "integer" },
        "name": { "type": "string" },
        "email": { "type": "string", "format": "email" }
      }
    }
```

See the [Assertions](assertions) page for detailed information on all available assertion types.

## Complete Example

Here's a complete example of a test definition file:

```yaml
name: "User API Tests"
description: "Tests for user management API endpoints"
variables:
  base_url:
    type: string
    value: "https://api.example.com/v1"
  admin_token:
    type: string
    value: "${env:ADMIN_TOKEN}"

suites:
  - name: "User Management"
    cases:
      - title: "List Users"
        request:
          method: GET
          url: "${base_url}/users"
          headers:
            - key: Authorization
              value: Bearer ${admin_token}
            - key: Accept
              value: application/json
          body:
            type: json
            data: null
          assertions:
            status: 200
            headers:
              content-type: "application/json; charset=utf-8"
            body:
              "$.users": "length > 0"

      - title: "Get Specific User"
        request:
          method: GET
          url: "${base_url}/users/123"
          headers:
            - key: Authorization
              value: Bearer ${admin_token}
            - key: Accept
              value: application/json
          body:
            type: json
            data: null
          assertions:
            status: 200
            body:
              "$.id": 123
              "$.name": "John Doe"
              "$.email": "john@example.com"
```