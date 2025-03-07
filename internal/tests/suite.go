package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/reqassert"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"go.uber.org/zap"
)

func (suite *TestSuite) Run(logger logging.Logger, client easyreq.HttpClient) (TestSuiteResult, error) {
	result := TestSuiteResult{
		Cases: make(map[string]TestCaseResult, len(suite.Cases)),
	}

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
	return result, nil
}

func (suite *TestSuite) ExecCase(testcase *TestCase, logger logging.Logger, client easyreq.HttpClient) (TestCaseResult, error) {
	startTime := time.Now()

	// Create a copy of the request to apply variable interpolation
	request := testcase.Request

	// Get variables from the test definition
	// We'll need to pass these variables to the suite when executing tests
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

