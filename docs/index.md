---
layout: default
title: Home
nav_order: 1
description: "HttpProbe is a powerful HTTP API testing tool for defining, running, and validating API tests."
permalink: /
---

# HttpProbe Documentation
{: .fs-9 }

A powerful HTTP API testing tool for defining, running, and validating API tests using YAML or JSON test definitions.
{: .fs-6 .fw-300 }

[Get started now](#getting-started){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View it on GitHub](https://github.com/mrfoh/httpprobe){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## Overview

HttpProbe makes API testing simple and powerful by allowing you to define tests using YAML or JSON files. It supports:

- **Variable interpolation** for dynamic test data
- **Comprehensive assertions** for status codes, headers, and response bodies
- **Schema validation** for verifying response structures
- **Multiple output formats** for results visualization
- **Parallel execution** for faster test runs

## Getting Started

### Installation

```bash
# From source
git clone https://github.com/mrfoh/httpprobe.git
cd httpprobe
go build -o httpprobe ./cmd/main.go

# Using Go Install
go install github.com/mrfoh/httpprobe@latest
```

### Basic Usage

1. Create a test definition file named `simple.test.yaml`:

```yaml
name: "Simple API Test"
description: "Testing a simple API endpoint"
variables:
  base_url:
    type: string
    value: "https://api.example.com"
suites:
  - name: "Basic Tests"
    cases:
      - title: "Get Users"
        request:
          method: GET
          url: "${base_url}/users"
          headers:
            - key: Accept
              value: application/json
          assertions:
            status: 200
```

2. Run your test:

```bash
httpprobe run simple.test.yaml
```

## Example Test Output

```
Simple API Test: simple.test.yaml
  Suite: Basic Tests
    Get Users (124.56 ms): PASS

Test Suites: 1 passed, 1 total
Test Cases: 1 passed, 1 total
Total time: 124.56 ms
```

## Why HttpProbe?

- **Declarative tests** - Focus on what to test, not how to test it
- **Reusable variables** - Define once, use across multiple test cases
- **Powerful assertions** - Validate complex responses with ease
- **Detailed failure reporting** - Quickly understand what went wrong
- **Flexible execution** - Run tests in sequence or in parallel

## Next Steps

Explore the detailed documentation:

- [Test Definitions](test-definitions) - Learn the structure of test definition files
- [Variable Interpolation](variable-interpolation) - Dynamic values in your tests
- [Assertions](assertions) - Validating API responses
- [Command Line Usage](cli-usage) - Command line options and arguments