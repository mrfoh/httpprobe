# HttpProbe

A powerful HTTP API testing tool for defining, running, and validating API tests using YAML or JSON test definitions.

![HttpProbe](https://via.placeholder.com/800x400?text=HttpProbe+API+Testing+Tool)

## Features

- **Test Definitions**: Define API tests in YAML or JSON format
- **Variable Interpolation**: Support for variables, environment variables, and dynamic functions
- **Rich Assertions**: Validate status codes, headers, and response bodies with detailed failure reporting
- **Schema Validation**: Verify response structures with JSON schema
- **Multiple Output Formats**: View results in text, table, or JSON formats
- **Concurrency**: Run tests in parallel for faster execution
- **Flexible Logging**: Configurable logging levels and formats

## Installation

### From Source

```bash
git clone https://github.com/mrfoh/httpprobe.git
cd httpprobe
go build -o httpprobe ./cmd/main.go
```

### Using Go Install

```bash
go install github.com/mrfoh/httpprobe@latest
```

## Quick Start

1. Create a test definition file (`test.yaml`):

```yaml
name: "Simple API Test"
description: "Testing basic API functionality"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
  api_key:
    type: string
    value: "${env:API_KEY}"

suites:
  - name: "User API"
    cases:
      - title: "Get Users"
        request:
          method: GET
          url: "${base_url}/users"
          headers:
            - key: Authorization
              value: Bearer ${api_key}
          body:
            type: json
            data: null
          assertions:
            status: 200
            headers:
              content-type: "application/json"
            body:
              "$.status": "success"
```

2. Run the test:

```bash
httpprobe run test.yaml
```

## Test Definition Structure

### Basic Structure

```yaml
name: "Test Name"
description: "Test Description"
variables:
  variable_name:
    type: string
    value: "variable_value"
suites:
  - name: "Suite Name"
    cases:
      - title: "Test Case Title"
        request:
          method: HTTP_METHOD
          url: "URL_PATH"
          headers:
            - key: "Header-Name"
              value: "Header-Value"
          body:
            type: json|text
            data: {} or "string"
          assertions:
            status: STATUS_CODE
            headers:
              header-name: "expected-value"
            body:
              "$.json.path": "expected-value"
```

### Variable Interpolation

HttpProbe supports several types of variable interpolation:

1. **Simple Variables**: `${variable_name}`
   ```yaml
   variables:
     api_version:
       type: string
       value: "v1"
   # Usage
   url: "https://api.example.com/${api_version}/users"
   ```

2. **Environment Variables**: `${env:ENV_VAR_NAME}`
   ```yaml
   # Usage - reads API_KEY from environment
   headers:
     - key: Authorization
       value: Bearer ${env:API_KEY}
   ```

3. **Function Variables**:
   - Random strings: `${random(10)}` - Generates a 10-character random string
     ```yaml
     variables:
       request_id:
         type: string
         value: ${random(16)}
     ```
   - Timestamps: `${timestamp(2006-01-02)}` - Generates a timestamp in specified format
     ```yaml
     # Current date in ISO-8601 format
     body:
       type: json
       data: |
         {
           "created_at": "${timestamp(2006-01-02T15:04:05Z)}"
         }
     ```

Variables can be used in URLs, headers, and request bodies, and are processed at execution time.

### Assertions

Test assertions can validate:

1. **Status Codes**: Verify HTTP response status
   ```yaml
   assertions:
     status: 200
   ```

2. **Headers**: Validate response headers
   ```yaml
   assertions:
     headers:
       content-type: "application/json; charset=utf-8"
       cache-control: "no-cache"
   ```

3. **Body**: Check response body using JSONPath expressions
   ```yaml
   assertions:
     body:
       # Exact match
       "$.status": "success"
       # Value in array
       "$.users[0].email": "user@example.com"
       # Array length check
       "$.users": "length 5"
       # Numeric comparison
       "$.count": "> 10"
       # Contains value
       "$.message": "contains partial text"
   ```

4. **Schema**: Validate the entire response structure using JSON Schema
   ```yaml
   assertions:
     status: 200
     schema: |
       {
         "type": "object",
         "required": ["status", "data"],
         "properties": {
           "status": { "type": "string" },
           "data": { "type": "array" }
         }
       }
   ```

#### JSONPath Expression Examples

- `$.users[0].name`: First user's name
- `$.meta.total`: Total count in metadata
- `$.items[*].id`: All IDs in the items array
- `$.countries.length`: Number of countries in the array

## Command Line Usage

```bash
# Run a single test file
httpprobe run test.yaml

# Run multiple test files
httpprobe run test1.yaml test2.yaml

# Run all tests in a directory
httpprobe run ./tests/

# Specify output format
httpprobe run test.yaml --output text|table|json

# Set log level
httpprobe run test.yaml --log-level debug|info|warn|error

# Set concurrency
httpprobe run test.yaml --concurrency 5
```

## Test Results and Failure Reporting

HttpProbe provides detailed information about test failures to help diagnose and fix issues quickly. Each failing test case displays:

- The expected and actual values that didn't match
- The specific part of the response that failed validation
- Validation errors from JSON schema checks

Example output for a failing test:

```
API Test: samples/test.yaml
  Suite: Authentication Tests
    Login Request (0.25 ms): FAIL
      Failures:
        - expected status code 200, got 401
        - header 'content-type' not found in response
    Validate Response Schema (0.43 ms): FAIL
      Failures:
        - JSONPath '$.user.id' not found in response body
        - Property 'email' is required but missing
```

### Output Formats

- **Text Format**: Human-readable output with colored status indicators
- **Table Format**: Compact table view showing test results and summarized failures
- **JSON Format**: Structured output for programmatic processing, saved to `test-results.json`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Example Projects

Check out these example projects to see HttpProbe in action:

- [API Test Suite](https://github.com/example/api-test-suite) - Comprehensive test suite for RESTful APIs
- [Microservices Testing](https://github.com/example/microservices-testing) - Testing patterns for microservice architectures

## Acknowledgements

- [Go HTTP Client](https://pkg.go.dev/net/http) - For reliable HTTP communication
- [ZAP Logger](https://github.com/uber-go/zap) - For structured, high-performance logging
- [JSON Schema](https://json-schema.org/) - For response validation