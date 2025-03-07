---
layout: default
title: Assertions
nav_order: 4
description: "Learn how to validate API responses with HttpProbe assertions."
---

# Assertions
{: .no_toc }

Assertions are the heart of API testing with HttpProbe. They allow you to validate that responses meet your expectations by checking status codes, headers, response bodies, and schemas.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Introduction

Assertions in HttpProbe are defined in the `assertions` section of a test case. They specify the expected properties of the HTTP response. When a test is executed, the actual response is compared against these assertions to determine if the test passes or fails.

```yaml
assertions:
  status: 200
  headers:
    content-type: "application/json"
  body:
    "$.status": "success"
    "$.user.id": 123
```

## Types of Assertions

HttpProbe supports four main types of assertions:

1. Status code assertions
2. Header assertions
3. Body assertions
4. Schema assertions

### Status Code Assertions

Status code assertions verify that the HTTP response has the expected status code:

```yaml
assertions:
  status: 200  # Expects HTTP 200 OK
```

You can assert any standard HTTP status code:

```yaml
status: 201  # Created
status: 400  # Bad Request
status: 404  # Not Found
status: 500  # Internal Server Error
```

### Header Assertions

Header assertions validate that the response contains specific HTTP headers with expected values:

```yaml
assertions:
  headers:
    content-type: "application/json; charset=utf-8"
    cache-control: "no-cache"
```

Header names are case-insensitive, matching HTTP standards.

### Body Assertions

Body assertions verify the content of the response body using JSONPath expressions:

```yaml
assertions:
  body:
    "$.id": 123
    "$.name": "John Doe"
    "$.enabled": true
```

#### JSONPath Syntax

JSONPath is a query language for JSON, similar to how XPath is for XML. HttpProbe uses JSONPath to extract values from JSON responses for validation.

Common JSONPath expressions:

| Expression | Description | Example |
| ---------- | ----------- | ------- |
| `$` | Root object | `$` (the entire document) |
| `$.property` | Child property | `$.name` |
| `$.property.nested` | Nested property | `$.user.email` |
| `$[index]` | Array index | `$[0]` (first item) |
| `$.array[index]` | Array property with index | `$.users[0]` (first user) |
| `$.array[*]` | All items in array | `$.users[*].name` (all user names) |
| `$.*.property` | Property in all objects | `$.*.id` (all IDs) |

#### Value Comparisons

Body assertions can use various comparison operators:

```yaml
body:
  # Equality (default)
  "$.status": "success"
  
  # Numeric comparisons
  "$.count": "> 5"       # Greater than
  "$.price": ">= 10.5"   # Greater than or equal
  "$.stock": "< 100"     # Less than
  "$.rating": "<= 5"     # Less than or equal
  
  # String contains
  "$.message": "contains error"  # String contains 'error'
  
  # Array length
  "$.items": "length 10"      # Exactly 10 items
  "$.users": "length > 0"     # At least 1 item
  "$.roles": "length <= 5"    # At most 5 items
```

#### Type Validation

HttpProbe automatically checks that the value type matches the expected type:

```yaml
body:
  "$.id": 123              # Expects a number
  "$.name": "John"         # Expects a string
  "$.active": true         # Expects a boolean
  "$.settings": null       # Expects null
  "$.tags": ["tag1", "tag2"] # Expects an array
```

### Schema Assertions

Schema assertions validate the entire response structure using JSON Schema:

```yaml
assertions:
  schema: |
    {
      "type": "object",
      "required": ["id", "name", "email"],
      "properties": {
        "id": { "type": "integer" },
        "name": { "type": "string" },
        "email": { "type": "string", "format": "email" },
        "age": { "type": "integer", "minimum": 18 },
        "tags": {
          "type": "array",
          "items": { "type": "string" }
        }
      }
    }
```

JSON Schema allows you to validate:

- Required properties
- Property types
- String formats (email, date-time, etc.)
- Numeric ranges
- Array contents
- And much more

## Handling Assertion Failures

When assertions fail, HttpProbe provides detailed error messages to help you understand what went wrong:

```
Test case "Get User" FAILED:
  - Expected status code 200, got 404
  - Header 'content-type' not found in response
  - JSONPath '$.user.email' not found in response body
  - Schema validation error: required property 'email' not present
```

The failure messages include:
- The expected and actual values
- The path to the failed property
- Specific details about schema validation errors

## Combining Assertion Types

You can combine multiple assertion types in a single test case:

```yaml
assertions:
  status: 200
  headers:
    content-type: "application/json"
  body:
    "$.success": true
    "$.data.count": "> 0"
  schema: |
    {
      "type": "object",
      "required": ["success", "data"],
      "properties": {
        "success": { "type": "boolean" },
        "data": {
          "type": "object",
          "required": ["count", "items"],
          "properties": {
            "count": { "type": "integer" },
            "items": { "type": "array" }
          }
        }
      }
    }
```

## Best Practices

1. **Start with status code** assertions as a baseline check
2. **Validate critical headers** like `content-type` and authentication tokens
3. **Check key properties** in the response body that are essential to your test case
4. **Use schema validation** for comprehensive structure validation
5. **Focus on business-critical assertions** rather than checking every property
6. **Use specific JSONPath expressions** to target exactly what you want to validate
7. **Include detailed error messages** to make debugging easier when tests fail

## Examples

### Testing a User API

```yaml
cases:
  - title: "Get User Profile"
    request:
      method: GET
      url: "${base_url}/users/123"
      headers:
        - key: Authorization
          value: Bearer ${token}
      body:
        type: json
        data: null
    assertions:
      status: 200
      headers:
        content-type: "application/json; charset=utf-8"
      body:
        "$.id": 123
        "$.name": "John Doe"
        "$.email": "john@example.com"
        "$.subscription.active": true
        "$.roles": "length > 0"
```

### Testing Authentication

```yaml
cases:
  - title: "Login with Valid Credentials"
    request:
      method: POST
      url: "${base_url}/auth/login"
      headers:
        - key: Content-Type
          value: application/json
      body:
        type: json
        data: |
          {
            "username": "validuser",
            "password": "validpassword"
          }
    assertions:
      status: 200
      body:
        "$.success": true
        "$.token": "contains eyJ"  # JWT tokens start with eyJ
      schema: |
        {
          "type": "object",
          "required": ["success", "token", "user"],
          "properties": {
            "success": { "type": "boolean" },
            "token": { "type": "string" },
            "user": {
              "type": "object",
              "required": ["id", "username"],
              "properties": {
                "id": { "type": "integer" },
                "username": { "type": "string" }
              }
            }
          }
        }
        
  - title: "Login with Invalid Credentials"
    request:
      method: POST
      url: "${base_url}/auth/login"
      headers:
        - key: Content-Type
          value: application/json
      body:
        type: json
        data: |
          {
            "username": "validuser",
            "password": "wrongpassword"
          }
    assertions:
      status: 401
      body:
        "$.success": false
        "$.message": "contains Invalid credentials"
```

### Testing Pagination

```yaml
cases:
  - title: "List Items with Pagination"
    request:
      method: GET
      url: "${base_url}/items?page=1&limit=10"
      headers:
        - key: Accept
          value: application/json
      body:
        type: json
        data: null
    assertions:
      status: 200
      body:
        "$.data": "length <= 10"
        "$.meta.page": 1
        "$.meta.limit": 10
        "$.meta.total": "> 0"
```