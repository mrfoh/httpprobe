---
layout: default
title: Variable Interpolation
nav_order: 3
description: "Learn how to use dynamic values in your test definitions."
---

# Variable Interpolation
{: .no_toc }

HttpProbe provides powerful variable interpolation capabilities that make your test definitions reusable and dynamic.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Variable Types

HttpProbe supports several types of variables:

1. **Defined Variables** - Values defined in your test definition
2. **Environment Variables** - Values loaded from your system environment or `.env` file
3. **Function Calls** - Dynamic values generated at runtime
4. **Exported Variables** - Values extracted from response bodies during test execution

## Defining Variables

Variables can be defined at the test definition level or within a test suite:

```yaml
name: Example with Variables
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  api_key:
    type: string
    value: "${env:API_KEY}"

suites:
  - name: User API
    variables:
      user_id:
        type: string
        value: "123"
    cases:
      - title: Get User
        request:
          method: GET
          url: "${base_url}/users/${user_id}"
          headers:
            - key: Authorization
              value: Bearer ${api_key}
```

Variables defined at the test definition level are available to all test suites, while suite-level variables are only available within that suite.

## Using Environment Variables

You can reference environment variables using the `${env:VAR_NAME}` syntax:

```yaml
variables:
  api_key:
    type: string
    value: "${env:API_KEY}"
```

Environment variables can be loaded from:

1. Your system environment
2. A `.env` file in the current directory
3. A specific environment file specified with the `--envfile` option

Example `.env` file:

```
API_KEY=abc123secret
API_URL=https://api.example.com
```

## Dynamic Functions

HttpProbe supports these built-in functions:

### Random String Generation

Generate a random string of a specified length:

```yaml
variables:
  random_id:
    type: string
    value: "${random(10)}"  # Generates a 10-character random string
```

### Timestamp Generation

Generate the current timestamp in a specified format:

```yaml
variables:
  current_date:
    type: string
    value: "${timestamp(2006-01-02)}"  # Format: YYYY-MM-DD
```

The format uses Go's time formatting syntax.

## Exporting Response Values as Variables

You can extract values from response bodies and use them in subsequent test cases. This is particularly useful for authentication flows, where you need to extract a token from a login response and use it in subsequent API calls.

To export a value from a response body, use the `export` property in your request:

```yaml
export:
  body:
    - path: "$.data.token"
      as: "access_token"
    - path: "$.data.refresh_token"
      as: "refresh_token"
```

The `path` property uses JSONPath syntax to locate the value in the response body, and the `as` property defines the variable name that will be created.

Here's a complete example demonstrating variable exports:

```yaml
name: Login API Test with Value Export
description: Tests login API and exports tokens for subsequent requests
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  username:
    type: string
    value: "testuser"
  password:
    type: string
    value: "password123"
suites:
  - name: Authentication
    cases:
      - title: Login to get token
        request:
          method: POST
          url: "${base_url}/login"
          headers:
            - key: content-type
              value: application/json
          body:
            type: json
            data: |
              {
                "username": "${username}",
                "password": "${password}"
              }
          assertions:
            status: 200
            headers:
              content-type: "application/json"
            body:
              $.success: true
          export:
            body:
              - path: "$.data.token"
                as: "access_token"
              - path: "$.data.refresh_token"
                as: "refresh_token"
              - path: "$.data.expires_in"
                as: "token_expires"
      
      - title: Get user profile using exported token
        request:
          method: GET
          url: "${base_url}/profile"
          headers:
            - key: Authorization
              value: "Bearer ${access_token}"
            - key: content-type
              value: application/json
          assertions:
            status: 200
            body:
              $.username: "${username}"
```

In this example, the login response contains a token that is exported as `access_token` and used in the subsequent request to get the user profile.

## Variable Usage Examples

Variables can be used in various parts of your test definition:

### In URLs

```yaml
url: "${base_url}/users/${user_id}"
```

### In Headers

```yaml
headers:
  - key: Authorization
    value: Bearer ${api_key}
  - key: X-Request-ID
    value: ${random(8)}
```

### In Request Bodies

```yaml
body:
  type: json
  data:
    username: "${username}"
    timestamp: "${timestamp(2006-01-02T15:04:05Z)}"
```

### In Assertions

```yaml
assertions:
  body:
    $.username: "${username}"
```

## Variable Resolution Order

When resolving variables, HttpProbe follows this order:

1. Environment variables
2. Test definition variables
3. Test suite variables
4. Exported variables (from previous test cases)
5. Function calls

If a variable with the same name exists at multiple levels, the most specific one takes precedence.

## Variable Best Practices

1. Use variables for all values that might change between environments (URLs, credentials, etc.)
2. Define reusable values at the test definition level
3. Define test-specific values at the suite level
4. Use environment variables for sensitive information like API keys and passwords
5. Use random values for data that should be unique (IDs, email addresses, etc.)
6. Create separate .env files for different environments (.env.dev, .env.prod)
7. Use exported variables for values that need to be extracted from responses and used in subsequent requests