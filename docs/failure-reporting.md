---
layout: default
title: Failure Reporting
nav_order: 6
description: "Understanding and troubleshooting test failures in HttpProbe."
---

# Failure Reporting
{: .no_toc }

HttpProbe provides detailed information about test failures to help you quickly identify and fix issues.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Overview

When tests fail, HttpProbe generates detailed failure reports that help you understand what went wrong. These reports include:

- The specific assertions that failed
- Expected vs. actual values
- Helpful context about the failure
- Suggestions for fixing common issues

## Reading Failure Reports

### Text Output Format

In the default text output format, failures are displayed directly under the test case:

```
User API Tests: tests/user-api.yaml
  Suite: Authentication
    Login with Valid Credentials (124.56 ms): PASS
    Login with Invalid Credentials (85.23 ms): FAIL
      Failures:
        - expected status code 401, got 500
        - JSONPath '$.message' not found in response body
```

Each failure includes:
- The test definition and suite name
- The failing test case name
- The execution time
- A list of specific failures with details

### Table Output Format

In the table output format, failures are displayed in a more compact form:

```
+----------------+----------------+---------------------------+--------+-------------------------+
| Test Definition | Test Suite     | Test Case                 | Result | Failures               |
+----------------+----------------+---------------------------+--------+-------------------------+
| User API Tests  | Authentication | Login with Valid Creds    | PASS   |                        |
|                 |                | Login with Invalid Creds  | FAIL   | expected status co...  |
+----------------+----------------+---------------------------+--------+-------------------------+
```

The table format shows truncated failure messages due to space constraints, but you can switch to text format for more details.

### JSON Output Format

JSON output provides the most comprehensive failure details for programmatic processing:

```json
{
  "testDefinitions": [
    {
      "name": "User API Tests",
      "path": "tests/user-api.yaml",
      "suites": [
        {
          "name": "Authentication",
          "cases": [
            {
              "name": "Login with Valid Credentials",
              "passed": true,
              "timingMs": 124.56
            },
            {
              "name": "Login with Invalid Credentials",
              "passed": false,
              "timingMs": 85.23,
              "failureReasons": [
                "expected status code 401, got 500",
                "JSONPath '$.message' not found in response body"
              ]
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "totalTestDefinitions": 1,
    "totalSuites": 1,
    "passedSuites": 0,
    "totalCases": 2,
    "passedCases": 1,
    "totalTimeMs": 209.79
  }
}
```

JSON output is saved to a file named `test-results.json` by default.

## Common Failure Types

### Status Code Failures

```
expected status code 200, got 404
```

This indicates that the API returned a different status code than expected. Common causes include:

- Incorrect URL or endpoint
- Missing or invalid authentication
- Server-side issues
- Resource not found

### Header Failures

```
header 'content-type' not found in response
```

or

```
expected header 'content-type' to be 'application/json', got 'text/plain'
```

These failures indicate issues with response headers. Common causes include:

- API returning a different content type than expected
- Missing required headers in the response
- Case sensitivity issues in header names

### Body Failures

```
JSONPath '$.user.id' not found in response body
```

or

```
expected '123', got '456' at path '$.user.id'
```

These failures indicate issues with the response body. Common causes include:

- API returning a different data structure than expected
- Missing fields in the response
- Different data types or values than expected
- Incorrect JSONPath expression

### Schema Failures

```
Schema validation error: required property 'email' not present
```

or

```
Schema validation error: property 'age' must be an integer
```

These failures indicate that the response doesn't match the expected schema. Common causes include:

- Missing required fields
- Wrong data types
- Values outside of expected ranges
- Structural differences in the response

## Troubleshooting Strategies

When tests fail, follow these steps to diagnose and fix the issues:

1. **Examine the failure messages** carefully to understand what went wrong
2. **Run the test with debug logging** to see the actual request and response:
   ```bash
   httpprobe run failing-test.yaml --log-level debug
   ```
3. **Check the API documentation** to confirm the expected behavior
4. **Try the request manually** using a tool like cURL or Postman
5. **Verify authentication credentials** and other request parameters
6. **Update assertions** to match the actual API behavior if it has changed

## Failure Aggregation

HttpProbe aggregates failures to help you understand the overall health of your API:

```
Test Suites: 3 passed, 5 total
Test Cases: 12 passed, 18 total
Total time: 3.56 s
```

This summary helps you quickly see how many tests passed and failed.

## Integrating with CI/CD

HttpProbe's exit codes make it easy to integrate with CI/CD systems:

- Exit code 0: All tests passed
- Exit code 1: One or more tests failed
- Exit code 2: Execution error (e.g., invalid configuration)

You can use JSON output to generate custom reports:

```bash
httpprobe run tests/*.yaml --output json
# Then process test-results.json with a custom script
```

## Best Practices

1. **Start with simple assertions** and add more specific ones as needed
2. **Use descriptive test case names** that clearly indicate the purpose of the test
3. **Group related tests** in the same suite for better organization
4. **Review failure messages** carefully before making changes
5. **Update tests** when API behavior changes intentionally
6. **Add comments** to explain complex assertions or edge cases