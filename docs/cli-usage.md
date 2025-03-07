---
layout: default
title: Command Line Usage
nav_order: 5
description: "Learn how to use the HttpProbe command line interface."
---

# Command Line Usage
{: .no_toc }

HttpProbe provides a powerful command line interface for running tests, configuring options, and customizing output.
{: .fs-6 .fw-300 }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

---

## Basic Command Structure

HttpProbe's command line interface follows this general structure:

```bash
httpprobe [command] [arguments] [options]
```

The main commands are:

- `run` - Run test definitions
- `version` - Display the current version

## Running Tests

The most common command is `run`, which executes test definitions:

```bash
httpprobe run [options]
```

### Running Tests in Current Directory

```bash
httpprobe run
```

### Running Tests in a Specific Directory

```bash
httpprobe run --searchpath ./tests/
```

### Running Only Specific Test Files

```bash
httpprobe run --include *.test.json
```

## Command Line Options

HttpProbe supports several command line options to customize test execution and output. Here's a complete reference of all available flags:

| Flag | Description | Default |
| ---- | ----------- | ------- |
| `-c, --concurrency` | Number of concurrent test definitions to execute | 2 |
| `-e, --envfile` | Environment file to load environment variables from | `.env` |
| `-i, --include` | Include tests with the specified extensions | `.test.yaml, .test.json` |
| `-o, --output` | Output format to use (text, json, table) | `text` |
| `-p, --searchpath` | Path to search for test files | `./` |
| `-v, --verbose` | Enable verbose output | `false` |
| `-h, --help` | Display help information | - |

### Output Format

Control how test results are displayed:

```bash
httpprobe run --output text|table|json
```

- `text` (default) - Human-readable output with colors for pass/fail status
- `table` - Tabular format for more compact display
- `json` - JSON format for programmatic processing

### Concurrency

Control how many test definitions run in parallel:

```bash
httpprobe run --concurrency 5
```

The default concurrency is 2. Increasing this value can significantly speed up test execution, especially for tests with high latency.

### Verbose Output

Enable detailed output for debugging:

```bash
httpprobe run --verbose
```

### Environment File

Load environment variables from a file:

```bash
httpprobe run --envfile .env.test
```

Default is `.env`. This is useful for loading different environment variables for different environments.

### Include Pattern

Specify which file extensions to include as test definitions:

```bash
httpprobe run --include .test.yaml,.test.json
```

Default includes `.test.yaml` and `.test.json` files. This allows you to filter which test files to run.

### Search Path

Specify the path to search for test files:

```bash
httpprobe run --searchpath ./tests/api/
```

Default is the current directory (`./`). This allows you to specify which directory to scan for test files.

## Complete Command Examples

### Basic Test Run

```bash
httpprobe run
```

### Run Tests in a Specific Directory

```bash
httpprobe run --searchpath ./tests/api/
```

### Fast Parallel Execution

```bash
httpprobe run --concurrency 10
```

### CI/CD Integration

```bash
httpprobe run --output json --searchpath ./tests/
```

### Debugging Tests

```bash
httpprobe run --verbose
```

## Environment Variables

HttpProbe can load environment variables from a file specified with `--envfile`. These variables can be accessed in your test definitions using the `${env:VARIABLE_NAME}` syntax.

### Configuration

By default, HttpProbe looks for a file named `.env` in the current directory. You can specify a different file using the `--envfile` flag:

```bash
httpprobe run --envfile .env.production
```

### Environment File Format

The environment file should use the standard format of one variable per line, with `KEY=VALUE` pairs:

```
# Comments are supported
API_KEY=secret-key
BASE_URL=https://api.example.com
TIMEOUT=30
```

Both single and double quotes are supported and will be stripped from the values:

```
SECRET_KEY='this is a quoted string'
API_TOKEN="another quoted string"
```

### Using Environment Variables in Tests

Once loaded, environment variables can be accessed in your test definitions using the `${env:VARIABLE_NAME}` syntax:

```yaml
name: API Tests with Environment Variables
suites:
  - name: Authentication Tests
    cases:
      - title: Test API key authentication
        request:
          url: ${env:BASE_URL}/auth
          headers:
            - key: Authorization
              value: Bearer ${env:API_KEY}
          body:
            type: json
            Data:
              environment: ${env:ENV_NAME}
```

### Environment Switching

This feature is particularly useful for switching between different environments (development, staging, production) without modifying your test files:

```bash
# Run tests against development environment
httpprobe run --envfile .env.development

# Run tests against production environment
httpprobe run --envfile .env.production
```

### Best Practices

1. **Sensitive Information**: Use environment variables for API keys, tokens, and other sensitive data
2. **Environment-specific URLs**: Store base URLs and endpoints in environment variables
3. **Credentials**: Keep usernames, passwords, and other credentials in environment variables
4. **Multiple Environments**: Create separate .env files for different environments (.env.dev, .env.staging, .env.prod)
5. **CI/CD Integration**: In CI/CD pipelines, use the appropriate environment file for each environment

## Exit Codes

HttpProbe returns different exit codes depending on the execution result:

| Exit Code | Description |
| --------- | ----------- |
| 0 | All tests passed |
| 1 | One or more tests failed |
| 2 | Execution error (invalid arguments, file not found, etc.) |

This is useful for integrating with CI/CD systems that use exit codes to determine if a step passed or failed.

## Best Practices

1. **Use meaningful test file names** that indicate what API or functionality they test
2. **Organize tests in directories** by API, feature, or environment
3. **Run tests with high concurrency** in CI/CD pipelines for faster feedback
4. **Use lower concurrency values** when testing rate-limited APIs
5. **Save test results as JSON** for integrating with other tools and dashboards
6. **Use verbose logging** when debugging test failures