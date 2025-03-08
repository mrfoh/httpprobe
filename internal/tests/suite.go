package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/reqassert"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"github.com/oliveagle/jsonpath"
	"go.uber.org/zap"
)

func (suite *TestSuite) Run(logger logging.Logger, client easyreq.HttpClient) (TestSuiteResult, error) {
	result := TestSuiteResult{
		Cases: make(map[string]TestCaseResult, len(suite.Cases)),
	}

	// Safety check for nil logger or client
	if logger == nil {
		// Create a default logger if none is provided
		defaultLogger, _ := logging.NewLogger(&logging.LoggerOptions{
			LogLevel: "info",
		})
		logger = defaultLogger
	}

	if client == nil {
		return result, fmt.Errorf("HTTP client is required to run test suite")
	}

	// Initialize variables map if nil
	if suite.Variables == nil {
		suite.Variables = make(map[string]Variable)
	}

	// Check if concurrency is enabled from the suite configuration
	// Default to sequential execution if not specified
	concurrent := false
	if val, ok := suite.Config["concurrent"]; ok {
		if boolVal, ok := val.(bool); ok {
			concurrent = boolVal
		}
	}

	if concurrent {
		// Use a mutex to protect access to the result map and variables
		var mutex sync.Mutex
		var wg sync.WaitGroup

		// For variable tracking across concurrent test cases
		variablesChan := make(chan map[string]Variable, len(suite.Cases))

		for _, c := range suite.Cases {
			wg.Add(1)
			// Create a local copy to avoid issues with the loop variable
			testCase := c

			go func() {
				defer wg.Done()

				logger.Debug("Running test case concurrently", zap.String("title", testCase.Title))
				// Create a copy of variables for this test case
				testVars := make(map[string]Variable)
				mutex.Lock()
				for k, v := range suite.Variables {
					testVars[k] = v
				}
				mutex.Unlock()

				// Create a local suite copy with copied variables
				localSuite := *suite
				localSuite.Variables = testVars

				// Run the test case
				testCaseResult, err := localSuite.ExecCase(&testCase, logger, client)

				mutex.Lock()
				if err != nil {
					logger.Error("Error executing test case", zap.String("title", testCase.Title), zap.Error(err))
					result.Cases[testCase.Title] = TestCaseResult{Passed: false}
				} else {
					result.Cases[testCase.Title] = testCaseResult
				}
				mutex.Unlock()

				// Send back any variables that were created/modified
				variablesChan <- localSuite.Variables
			}()
		}

		// Wait for all test cases to complete
		wg.Wait()
		close(variablesChan)

		// Merge all exported variables from the test cases
		for vars := range variablesChan {
			for k, v := range vars {
				suite.Variables[k] = v
			}
		}
	} else {
		// Sequential execution (original behavior)
		for _, c := range suite.Cases {
			logger.Debug("Running test case", zap.String("title", c.Title))
			// Run the test case
			testCaseResult, err := suite.ExecCase(&c, logger, client)
			if err != nil {
				logger.Error("Error executing test case", zap.String("title", c.Title), zap.Error(err))
				result.Cases[c.Title] = TestCaseResult{Passed: false}
			} else {
				result.Cases[c.Title] = testCaseResult
			}
		}
	}

	return result, nil
}

func (suite *TestSuite) ExecCase(testcase *TestCase, logger logging.Logger, client easyreq.HttpClient) (TestCaseResult, error) {
	startTime := time.Now()

	// Create a copy of the request to apply variable interpolation
	request := testcase.Request

	// Get variables from the test definition
	// We need to pass these variables to the suite when executing tests
	variables := suite.Variables

	// Apply variable interpolation to the request
	if err := InterpolateRequest(&request, variables); err != nil {
		logger.Debug("Error interpolating variables in request", zap.Error(err))
		return TestCaseResult{}, fmt.Errorf("error interpolating variables: %w", err)
	}

	// Convert request headers to map format
	headers := make(map[string]interface{})
	for _, h := range request.Headers {
		headers[h.Key] = h.Value
	}

	// Prepare request params
	params := easyreq.RequestParams{
		Headers: headers,
	}

	var resp *easyreq.HttpResponse
	var err error

	logger.Debug("Executing request", zap.String("method", request.Method), zap.String("url", request.URL))

	// Execute request based on method
	switch strings.ToUpper(request.Method) {
	case "GET":
		resp, err = client.Get(request.URL, params)
	case "POST":
		var body interface{}
		if request.Body.Data != nil {
			if request.Body.Type == "json" {
				// Handle JSON body string data
				if strData, ok := request.Body.Data.(string); ok {
					var jsonBody interface{}
					if err := json.Unmarshal([]byte(strData), &jsonBody); err == nil {
						body = jsonBody
					} else {
						body = strData // Use as string if not valid JSON
					}
				} else {
					body = request.Body.Data
				}
			} else {
				body = request.Body.Data
			}
		}
		resp, err = client.Post(request.URL, body, params)
	case "PUT":
		resp, err = client.Put(request.URL, request.Body.Data, params)
	case "DELETE":
		resp, err = client.Delete(request.URL, params)
	case "OPTIONS":
		resp, err = client.Options(request.URL, params)
	case "HEAD":
		resp, err = client.Head(request.URL, params)
	case "PATCH":
		resp, err = client.Patch(request.URL, request.Body.Data, params)
	default:
		return TestCaseResult{}, fmt.Errorf("unsupported HTTP method: %s", request.Method)
	}

	if err != nil {
		return TestCaseResult{}, fmt.Errorf("error executing request: %v", err)
	}

	// Process response body exports if they exist and we have exports defined
	if len(request.Export.Body) > 0 && resp != nil && resp.Body != nil {
		if err := processBodyExports(&request, resp, suite, logger); err != nil {
			logger.Warn("Error processing response body exports", zap.Error(err))
			// We continue execution even if export fails
		}
	}

	// Validate response using the new assertion framework
	passed, validationErrors, err := validateWithAssertions(resp, testcase.Request.Assertions, logger)

	elapsedTime := time.Since(startTime).Seconds()

	// Convert validation errors to strings for the result
	var failureReasons []string
	if !passed && len(validationErrors) > 0 {
		for _, valErr := range validationErrors {
			failureReasons = append(failureReasons, valErr.Error())
		}
	}

	return TestCaseResult{
		Passed:         passed,
		Timing:         elapsedTime,
		FailureReasons: failureReasons,
	}, err
}

// validateWithAssertions checks if the response matches the assertions using the reqassert package
// It returns:
// - a boolean indicating if all assertions passed
// - a slice of validation errors when assertions fail
// - an error if there was a problem with the validation process itself
func validateWithAssertions(resp *easyreq.HttpResponse, assertionsData map[string]interface{}, logger logging.Logger) (bool, []error, error) {
	// Handle nil response or nil assertions gracefully
	if resp == nil {
		return false, []error{fmt.Errorf("nil response cannot be validated")}, nil
	}

	if len(assertionsData) == 0 {
		// No assertions to validate, consider it a pass
		return true, nil, nil
	}

	// Create assertion builder
	builder := reqassert.NewBuilder()

	// Build assertions from data
	assertions, err := builder.BuildAssertions(assertionsData)
	if err != nil {
		logger.Error("Failed to build assertions", zap.Error(err))
		return false, nil, err
	}

	// Convert response headers from http.Header to map[string]string
	respHeaders := make(map[string]string)
	for name, values := range resp.Headers {
		if len(values) > 0 {
			respHeaders[name] = values[0]
		}
	}

	// Prepare assertion context
	ctx, err := builder.PrepareContext(resp.Status, respHeaders, resp.Body)
	if err != nil {
		logger.Error("Failed to prepare assertion context", zap.Error(err))
		return false, nil, err
	}

	// Validate all assertions
	validationErrors := builder.ValidateAll(assertions, ctx)

	// Log any validation errors
	for _, err := range validationErrors {
		logger.Debug("Assertion failed", zap.Error(err))
	}

	// Test passes if there are no validation errors
	return len(validationErrors) == 0, validationErrors, nil
}

// processBodyExports extracts values from the response body based on JSONPath expressions
// and adds them to the suite variables for use in subsequent test cases
func processBodyExports(request *Request, resp *easyreq.HttpResponse, suite *TestSuite, logger logging.Logger) error {
	// If there are no exports, return early
	if len(request.Export.Body) == 0 {
		return nil
	}

	// Initialize variables map if it doesn't exist
	if suite.Variables == nil {
		suite.Variables = make(map[string]Variable)
	}

	// Parse response body as JSON
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(resp.Body, &bodyMap); err != nil {
		return fmt.Errorf("error parsing response body as JSON for exports: %w", err)
	}

	// Process each body export definition
	for _, export := range request.Export.Body {
		logger.Debug("Processing body export",
			zap.String("path", export.Path),
			zap.String("as", export.As))

		// Skip empty paths or variable names
		if export.Path == "" || export.As == "" {
			logger.Warn("Skipping body export with empty path or variable name")
			continue
		}

		// Extract value using JSONPath
		path, err := jsonpath.Compile(export.Path)
		if err != nil {
			return fmt.Errorf("invalid JSONPath '%s': %w", export.Path, err)
		}

		// Lookup the value in the response body
		value, err := path.Lookup(bodyMap)
		if err != nil {
			logger.Warn("Error extracting value using JSONPath",
				zap.String("path", export.Path),
				zap.Error(err))
			continue
		}

		// Convert the value to string for storage as a variable
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%g", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		case nil:
			strValue = ""
		default:
			// Try to JSON marshal complex objects
			bytes, err := json.Marshal(v)
			if err != nil {
				logger.Warn("Error marshaling complex value to JSON",
					zap.String("path", export.Path),
					zap.Any("value", v),
					zap.Error(err))
				strValue = fmt.Sprintf("%v", v)
			} else {
				strValue = string(bytes)
			}
		}

		// Store the extracted value as a variable
		suite.Variables[export.As] = Variable{
			Type:  "string",
			Value: strValue,
		}

		logger.Debug("Exported response value to variable",
			zap.String("variable", export.As),
			zap.String("value", strValue))
	}

	return nil
}
