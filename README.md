# HttpProbe [![main](https://github.com/mrfoh/httpprobe/actions/workflows/main.yml/badge.svg?event=push)](https://github.com/mrfoh/httpprobe/actions/workflows/main.yml)

A powerful HTTP API testing tool for defining, running, and validating API tests using YAML or JSON test definitions.

![HttpProbe](https://via.placeholder.com/800x400?text=HttpProbe+API+Testing+Tool)

## Features

- **Test Definitions**: Define API tests in YAML or JSON format
- **Variable Interpolation**: Support for variables, environment variables, and dynamic functions
- **Rich Assertions**: Validate status codes, headers, and response bodies with detailed failure reporting
- **Schema Validation**: Verify response structures with JSON schema
- **Multiple Output Formats**: View results in text, table, or JSON formats
- **Concurrency**: Run test definitions and test cases in parallel for faster execution
- **Test Lifecycle Hooks**: Run setup and teardown operations before/after tests
- **Response Value Export**: Extract and reuse values from responses in subsequent tests
- **Flexible Logging**: Configurable logging levels and formats

## Installation

### Using the Install Script (Linux, macOS, WSL)

```bash
curl -sSL https://raw.githubusercontent.com/mrfoh/httpprobe/main/install.sh | bash
```

### macOS (Homebrew)

```bash
# Add the tap
brew tap mrfoh/tap

# Install httpprobe
brew install httpprobe
```

### Linux (Snap)

```bash
# Install httpprobe
sudo snap install httpprobe
```

### Windows (Scoop)

```powershell
# Add the bucket
scoop bucket add mrfoh https://github.com/mrfoh/scoopbucket

# Install httpprobe
scoop install mrfoh/httpprobe
```

### From Binary Releases

Download the prebuilt binary for your platform from the [releases page](https://github.com/mrfoh/httpprobe/releases).

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

### Release Process

This project uses [GoReleaser](https://goreleaser.com/) for building and publishing releases:

1. Make sure all your changes are committed and pushed
2. Tag the release: `git tag -a v1.2.3 -m "Release v1.2.3"`
3. Push the tag: `git push origin v1.2.3`
4. The GitHub Actions workflow will automatically build binaries and publish the release

#### Testing Releases Locally

To test the release process locally without publishing:

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Test the release process (no publishing)
goreleaser release --snapshot --clean --config .goreleaser.local.yml
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.