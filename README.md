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
httpprobe run -p sample.test.yaml
```

## Docs

See docs [here](https://mrfoh.github.io/httpprobe/)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.