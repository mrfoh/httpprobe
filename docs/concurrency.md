# Concurrency

HttpProbe supports running tests concurrently at multiple levels:

1. **Test Definitions**: Multiple test definition files can be executed concurrently
2. **Test Suites**: Test cases within a suite can be executed concurrently

## Concurrent Test Definitions

When running multiple test definition files, HttpProbe can execute them in parallel. To enable this, use the `--concurrency` or `-c` flag when running the CLI:

```bash
httpprobe run -p /path/to/tests -c 5
```

This will run up to 5 test definition files concurrently. Note that each file will still process its hooks in the correct sequence.

## Concurrent Test Cases

For test suites where test cases don't depend on each other, you can enable concurrent test case execution by adding a `config` section with `concurrent: true`:

```yaml
suites:
  - name: "API Tests"
    config:
      concurrent: true
    cases:
      - title: "Test 1"
        # Test case definition...
      
      - title: "Test 2"
        # Test case definition...
```

When `concurrent: true` is set, all test cases in that suite will run in parallel, which can significantly improve performance for independent tests.

## Variables and Concurrency

When running test cases concurrently, variables created or modified by one test case are not immediately available to other test cases in the same suite, since they're running in parallel. However:

1. Any variables exported by test cases will still be merged back into the suite's variables after all test cases complete
2. These merged variables will be available to subsequent test suites
3. Hook scripts always run sequentially, ensuring proper variable handling

## Best Practices

1. **Use sequential execution** when test cases depend on each other (e.g., one test case creates a resource that another test case needs)
2. **Use concurrent execution** when test cases are independent and can run in any order
3. **Use hooks** for setup and teardown operations that should happen before or after concurrent test execution
4. Set a reasonable concurrency level based on your system resources and the rate limits of the API you're testing
