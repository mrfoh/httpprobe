---
layout: default
title: Variable Interpolation
nav_order: 3
description: "Learn how to use dynamic values in your test definitions."
---

# Variable Interpolation
{: .no_toc }

Variable interpolation allows you to make your tests dynamic by substituting values at runtime. This feature is essential for creating reusable and flexible test definitions.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Introduction

Variables in HttpProbe are placeholders that get replaced with actual values when tests are executed. They use the syntax `${variable_name}` and can be used in:

- URLs
- Header values
- Request bodies
- And more

Variables help you:

- **Keep tests DRY** (Don't Repeat Yourself) by defining values once
- **Handle environment-specific values** like API keys and base URLs
- **Create dynamic data** for each test run
- **Make tests more readable** by using descriptive variable names

## Variable Types

HttpProbe supports several types of variables:

### 1. Simple Variables

These are defined in the `variables` section of your test definition:

```yaml
variables:
  base_url:
    type: string
    value: "https://api.example.com/v1"
  max_items:
    type: string
    value: "100"
```

You can then reference these variables in your tests:

```yaml
url: "${base_url}/items?limit=${max_items}"
```

### 2. Environment Variables

Environment variables let you access values from the host environment:

```yaml
variables:
  api_key:
    type: string
    value: "${env:API_KEY}"
```

This will substitute the value of the `API_KEY` environment variable at runtime. This is particularly useful for:

- **Sensitive information** like API keys and tokens
- **Environment-specific settings** that change between development, staging, and production
- **CI/CD pipelines** where environment variables are commonly used for configuration

### 3. Function Variables

HttpProbe also supports function-like syntax for generating dynamic values:

#### Random Strings

```yaml
variables:
  request_id:
    type: string
    value: "${random(16)}"  # Generates a 16-character random string
```

The `random()` function generates a random alphanumeric string of the specified length.

#### Timestamps

```yaml
variables:
  current_date:
    type: string
    value: "${timestamp(2006-01-02)}"  # Current date in YYYY-MM-DD format
```

The `timestamp()` function generates the current date/time in the specified format. It uses Go's time formatting syntax.

Common timestamp formats:

| Format | Description | Example |
| ------ | ----------- | ------- |
| `2006-01-02` | ISO date | "2023-11-05" |
| `2006-01-02T15:04:05Z` | ISO datetime | "2023-11-05T13:45:30Z" |
| `15:04:05` | Time only | "13:45:30" |
| `Mon, 02 Jan 2006` | RFC format | "Sun, 05 Nov 2023" |

## Variable Scope

Variables defined at the test definition level are available to all test suites and cases within that definition. This allows you to define common values once and reuse them throughout your tests.

## Where Variables Can Be Used

You can use variables in many places within your test definitions:

### In URLs

```yaml
url: "${base_url}/users/${user_id}"
```

### In Headers

```yaml
headers:
  - key: Authorization
    value: Bearer ${api_token}
  - key: X-Request-ID
    value: ${request_id}
```

### In JSON Bodies

```yaml
body:
  type: json
  data: |
    {
      "name": "John Doe",
      "email": "john@example.com",
      "token": "${api_token}",
      "created_at": "${timestamp(2006-01-02T15:04:05Z)}",
      "request_id": "${random(10)}"
    }
```

## Variable Resolution Process

Variables are resolved in the following order:

1. **Environment variables** (`${env:VAR_NAME}`) are resolved first
2. **Simple variables** (`${variable_name}`) are resolved next
3. **Function variables** (`${random()}`, `${timestamp()}`, etc.) are resolved last

This means you can have a variable that references another variable, and it will be resolved correctly.

## Examples

### Basic Variable Usage

```yaml
variables:
  api_version:
    type: string
    value: "v1"
  base_url:
    type: string
    value: "https://api.example.com/${api_version}"
```

### Environment Variable for Authentication

```yaml
variables:
  jwt_token:
    type: string
    value: "${env:JWT_TOKEN}"

suites:
  - name: "Authenticated Tests"
    cases:
      - title: "Get Protected Resource"
        request:
          method: GET
          url: "https://api.example.com/protected"
          headers:
            - key: Authorization
              value: Bearer ${jwt_token}
```

### Dynamic Data for Each Request

```yaml
variables:
  trace_id:
    type: string
    value: "${random(16)}"
  current_time:
    type: string
    value: "${timestamp(2006-01-02T15:04:05Z)}"

suites:
  - name: "Order API"
    cases:
      - title: "Create Order"
        request:
          method: POST
          url: "https://api.example.com/orders"
          headers:
            - key: X-Trace-ID
              value: ${trace_id}
          body:
            type: json
            data: |
              {
                "product_id": "prod-123",
                "quantity": 1,
                "order_date": "${current_time}"
              }
```

## Best Practices

1. **Use descriptive variable names** that clearly indicate their purpose
2. **Define common values as variables** to avoid duplication
3. **Use environment variables for credentials** and sensitive information
4. **Use dynamic functions for timestamps and IDs** to make tests more realistic
5. **Keep variable definitions at the top** of your test file for better visibility