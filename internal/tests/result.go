package tests

// ExecutionResult is the result of executing a test definition
type TestDefinitionExecResult struct {
	// Path is the path to the test definition file
	Path string
	// Suites is a map of test suite names to their results
	Suites map[string]TestSuiteResult
}

// TestSuiteResult is the result of executing a test suite
type TestSuiteResult struct {
	// Cases is a map of test case titles to their results
	Cases map[string]TestCaseResult
	// Variables contains any variables defined or exported during suite execution
	Variables map[string]Variable
}

// TestCaseResult is the result of executing a test case
type TestCaseResult struct {
	// Passed indicates if the test case passed
	Passed bool
	// Timing is the time taken to execute the test case
	Timing float64
	// FailureReasons contains the detailed reasons for failure (validation errors)
	FailureReasons []string
}

func (t *TestSuiteResult) Passed() bool {
	for _, result := range t.Cases {
		if !result.Passed {
			return false
		}
	}
	return true
}
