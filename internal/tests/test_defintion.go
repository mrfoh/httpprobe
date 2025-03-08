package tests

import "fmt"

type TestDefinition struct {
	// Path is the path to the test definition file
	Path string
	// Name is the identifier for the test definition. It is used to reference the test in logs, reports, and commands.
	Name string `yaml:"name" json:"name"`
	// Description of the test definition
	Description string `yaml:"description" json:"description"`
	// Variables to be used in the test definition
	Variables map[string]Variable `yaml:"variables" json:"variables"`
	// Other test definitions to be executed before the test suites in this definition are run
	BeforeAll []string `yaml:"before_all" json:"before_all"`
	// Other test definitions to be executed after the test suites in this definition are run
	AfterAll []string `yaml:"after_all" json:"after_all"`
	// Test definitions to be executed before each test suite in this definition
	BeforeEach []string `yaml:"before_each" json:"before_each"`
	// Test definitions to be executed after each test suite in this definition
	AfterEach []string `yaml:"after_each"`
	// Test suites to be executed
	Suites []TestSuite `yaml:"suites" json:"suites"`
}

type Variable struct {
	Type  string `yaml:"type" json:"type"`
	Value string `yaml:"value" json:"value"`
}

// TestSuite represent a suite of test cases
type TestSuite struct {
	// Name of the test suite
	Name string `yaml:"name" json:"name"`
	// Cases to test for in the suite
	Cases []TestCase `yaml:"cases" json:"cases"`
	// Variables available for test cases in this suite
	Variables map[string]Variable
	// Configuration options for the test suite
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// TestCase represent a test case to be executed
type TestCase struct {
	// Title of the test case
	Title string `yaml:"title" json:"title"`
	// Request is the HTTP request to be made
	Request Request `yaml:"request" json:"request"`
}

// Request represent an HTTP request to be made
type Request struct {
	// Method is the HTTP method to be used
	Method string `yaml:"method" json:"method"`
	// URL is the URL to be used for the request
	URL string `yaml:"url" json:"url"`
	// Headers are the headers to be used in the request
	Headers []RequestHeader `yaml:"headers" json:"headers"`
	// Body is the body to be sent in the request
	Body RequestBody `yaml:"body" json:"body"`
	// Assertions are the assertions to be made on the response
	Assertions map[string]interface{} `yaml:"assertions" json:"assertions"`
	// Export is the data to be exported from the response
	Export RequestExport `yaml:"export" json:"export"`
}

type RequestBody struct {
	Type string `yaml:"type" json:"type"`
	Data any    `yaml:"data" json:"data"`
}

type RequestHeader struct {
	Key   string `yaml:"key" json:"key"`
	Value string `yaml:"value" json:"value"`
}

// Legacy assertion types - kept for backward compatibility
type ResponseAssertion struct {
	Status  int         `yaml:"status" json:"status"`
	Body    []Assertion `yaml:"body" json:"body"`
	Headers []Assertion `yaml:"headers" json:"headers"`
}

type Assertion struct {
	Path     string `yaml:"path" json:"path"`
	Operator string `yaml:"operator" json:"operator"`
	Expected any    `yaml:"expected" json:"expected"`
}

type RequestExport struct {
	// The data to be exported from the response body
	Body []BodyExport `yaml:"body" json:"body"`
}

type BodyExport struct {
	// Path is the JSON path to the data to be exported
	Path string `yaml:"path" json:"path"`
	// As is the name to be used when exporting the value from the response to the test context variables
	As string `yaml:"as" json:"as"`
}

func (def *TestDefinition) Validate() error {
	if def.Name == "" {
		return fmt.Errorf("test definition name is required")
	}

	if len(def.Suites) == 0 {
		return fmt.Errorf("test definition must have at least one suite")
	}

	for _, suite := range def.Suites {
		if suite.Name == "" {
			return fmt.Errorf("suite name is required")
		}
	}

	return nil
}
