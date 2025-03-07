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
httpprobe run [test-files...] [options]
```

### Running a Single Test File

```bash
httpprobe run test.yaml
```

### Running Multiple Test Files

```bash
httpprobe run test1.yaml test2.yaml test3.yaml
```

### Running All Tests in a Directory

```bash
httpprobe run ./tests/
```

### Running Tests with Glob Pattern

```bash
httpprobe run ./tests/*.yaml
```

## Command Line Options

HttpProbe supports several command line options to customize test execution and output.

### Output Format

Control how test results are displayed:

```bash
httpprobe run test.yaml --output text|table|json
```

- `text` (default) - Human-readable output with colors for pass/fail status
- `table` - Tabular format for more compact display
- `json` - JSON format for programmatic processing (saved to `test-results.json`)

### Concurrency

Control how many tests run in parallel:

```bash
httpprobe run test.yaml --concurrency 5
```

The default concurrency is 1 (sequential execution). Increasing this value can significantly speed up test execution, especially for tests with high latency.

### Log Level

Control the verbosity of logging:

```bash
httpprobe run test.yaml --log-level debug|info|warn|error
```

- `debug` - Most verbose, shows all details including request/response information
- `info` (default) - Shows general execution information
- `warn` - Shows only warnings and errors
- `error` - Shows only errors

### Log File

Save logs to a file instead of displaying them in the console:

```bash
httpprobe run test.yaml --log-file ./logs/test-run.log
```

### Timeout

Set a timeout for HTTP requests:

```bash
httpprobe run test.yaml --timeout 30s
```

Default timeout is 30 seconds. The value can be specified in seconds (`30s`), milliseconds (`5000ms`), or minutes (`2m`).

### Variables

Override or add variables at runtime:

```bash
httpprobe run test.yaml --set base_url=https://api.staging.example.com
```

You can specify multiple variables:

```bash
httpprobe run test.yaml --set base_url=https://api.staging.example.com --set api_key=test-key
```

This is useful for:
- Testing against different environments
- CI/CD pipelines where values are passed in at runtime
- Quick testing without modifying test files

## Complete Command Examples

### Basic Test Run

```bash
httpprobe run ./tests/api-tests.yaml
```

### Production Test Run

```bash
httpprobe run ./tests/api-tests.yaml --set base_url=https://api.production.example.com --set api_key=${PROD_API_KEY}
```

### Fast Parallel Execution

```bash
httpprobe run ./tests/*.yaml --concurrency 10 --timeout 60s
```

### CI/CD Integration

```bash
httpprobe run ./tests/*.yaml --output json --log-level error --log-file ./logs/test-run.log
```

### Debugging a Specific Test

```bash
httpprobe run ./tests/failing-test.yaml --log-level debug
```

## Environment Variables

HttpProbe also supports configuration via environment variables:

| Environment Variable | Description | Default |
| -------------------- | ----------- | ------- |
| `HTTPPROBE_LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `HTTPPROBE_TIMEOUT` | Default request timeout | 30s |
| `HTTPPROBE_CONCURRENCY` | Default concurrency level | 1 |
| `HTTPPROBE_OUTPUT` | Default output format | text |

Environment variables are overridden by command line options.

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