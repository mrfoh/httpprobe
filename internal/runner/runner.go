package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alitto/pond/v2"
	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/tests"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"go.uber.org/zap"
)

type Runner struct {
	Parser       tests.TestDefinitionParser
	Logger       logging.Logger
	HttpClient   easyreq.HttpClient
	Concurrency  int
	ResultWriter tests.TestResultWriter
	// Map to track processed hooks to prevent infinite recursion
	processedHooks map[string]bool
	// Mutex to protect the processed hooks map
	hooksMutex     sync.Mutex
}

func NewRunner(opts *TestRunnerOptions) TestRunner {
	return &Runner{
		Parser:         opts.Parser,
		Logger:         opts.Logger,
		HttpClient:     opts.HttpClient,
		Concurrency:    opts.Concurrency,
		ResultWriter:   opts.Writer,
		processedHooks: make(map[string]bool),
	}
}

func openFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return data, nil
}

func hasMatchingExtension(path string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// Reads test definitions from the specified path
func (r *Runner) GetTestDefinitions(params *GetTestDefinitionsParams) ([]*tests.TestDefinition, error) {
	definitions := make([]*tests.TestDefinition, 0)

	// Ensure search path exists
	searchPathInfo, err := os.Stat(params.SearchPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing search path: %v", err)
	}

	// If search path is a file, process it directly
	if !searchPathInfo.IsDir() {
		if hasMatchingExtension(params.SearchPath, params.FileExtensions) {
			def, err := r.processTestFile(params.SearchPath)
			if err != nil {
				return nil, err
			}
			if def != nil {
				definitions = append(definitions, def)
			}
		}
		return definitions, nil
	}

	// Process all files in directory
	r.Logger.Debug(fmt.Sprintf("searching for test files in %s with extensions: %v", params.SearchPath, params.FileExtensions))

	err = filepath.Walk(params.SearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path: %v", err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has a matching extension
		if hasMatchingExtension(path, params.FileExtensions) {
			r.Logger.Debug(fmt.Sprintf("processing test file: %s", path))

			def, err := r.processTestFile(path)
			if err != nil {
				r.Logger.Error(fmt.Sprintf("error processing file %s: %v", path, err))
				return nil // Continue processing other files
			}

			if def != nil {
				definitions = append(definitions, def)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	if len(definitions) == 0 {
		r.Logger.Info(fmt.Sprintf("no test definitions found in %s", params.SearchPath))
	} else {
		r.Logger.Debug(fmt.Sprintf("found %d test definition(s)", len(definitions)))
	}

	return definitions, nil
}

// Execute runs the specified test definitions
func (r *Runner) Execute(definition []*tests.TestDefinition) (map[string]tests.TestDefinitionExecResult, error) {
	result := make(map[string]tests.TestDefinitionExecResult)

	// Reset processed hooks map for a new execution
	r.processedHooks = make(map[string]bool)

	if r.Concurrency > 1 {
		// Create a worker pool for executing test definitions concurrently
		pool := pond.NewPool(r.Concurrency)
		resultChan := make(chan struct {
			name   string
			result tests.TestDefinitionExecResult
			err    error
		}, len(definition))

		// Submit all test definitions to the pool
		for _, def := range definition {
			// Create a local copy of the test definition to avoid closure issues
			testDef := def
			pool.Submit(func() {
				testResult, err := r.executeTestDefinition(testDef)
				resultChan <- struct {
					name   string
					result tests.TestDefinitionExecResult
					err    error
				}{testDef.Name, testResult, err}
			})
		}

		// Wait for all tasks to complete
		pool.StopAndWait()
		close(resultChan)

		// Collect results
		var execError error
		for res := range resultChan {
			if res.err != nil {
				execError = res.err
				continue
			}
			result[res.name] = res.result
		}

		if execError != nil {
			return nil, execError
		}
	} else {
		// Execute sequentially for single-threaded mode or when hooks require it
		for _, def := range definition {
			testResult, err := r.executeTestDefinition(def)
			if err != nil {
				return nil, err
			}

			// Store result
			result[def.Name] = testResult
		}
	}

	return result, nil
}

// Write writes the test results
func (r *Runner) Write(results map[string]tests.TestDefinitionExecResult) {
	r.ResultWriter.Write(results)
}

// processTestFile reads and parses a single test file
func (r *Runner) processTestFile(path string) (*tests.TestDefinition, error) {
	data, err := openFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", path, err)
	}

	ext := filepath.Ext(path)
	def, err := r.Parser.Parse(data, ext)
	if err != nil {
		return nil, fmt.Errorf("error parsing file %s: %v", path, err)
	}

	def.Path = path

	if err := def.Validate(); err != nil {
		return nil, err
	}

	return def, nil
}

func (r *Runner) executeTestDefinition(def *tests.TestDefinition) (tests.TestDefinitionExecResult, error) {
	// Use mutex to protect access to processedHooks map
	r.hooksMutex.Lock()
	// Avoid recursive processing of the same definition
	if def.Path != "" && r.processedHooks[def.Path] {
		r.hooksMutex.Unlock()
		r.Logger.Warn("Skipping already processed hook to prevent recursion", zap.String("path", def.Path))
		return tests.TestDefinitionExecResult{}, nil
	}

	// Mark this definition as processed if it has a path
	if def.Path != "" {
		r.processedHooks[def.Path] = true
	}
	r.hooksMutex.Unlock()

	result := tests.TestDefinitionExecResult{
		Path:   def.Path,
		Suites: make(map[string]tests.TestSuiteResult, len(def.Suites)),
	}

	r.Logger.Debug(fmt.Sprintf("executing test definition: %s", def.Name))

	// Execute BeforeAll hooks if they exist
	if len(def.BeforeAll) > 0 {
		r.Logger.Debug("Executing BeforeAll hooks", zap.Strings("hooks", def.BeforeAll))
		hookVars, err := r.executeHooks(def.BeforeAll, def.Variables)
		if err != nil {
			r.Logger.Error("Error executing BeforeAll hooks", zap.Error(err))
			// We continue execution despite hook errors
		}

		// Merge hook variables into definition variables
		for k, v := range hookVars {
			if def.Variables == nil {
				def.Variables = make(map[string]tests.Variable)
			}
			def.Variables[k] = v
		}
	}

	// Make a copy of the suites to avoid modifying the original
	for i := range def.Suites {
		// Create a local copy of the suite to avoid issues with the loop variable
		suite := def.Suites[i]

		// Create a copy of definition variables for the suite
		suiteVars := make(map[string]tests.Variable)
		
		// First add definition-level variables
		for k, v := range def.Variables {
			suiteVars[k] = v
		}
		
		// Then add suite-level variables (to override any definition variables with the same name)
		if suite.Variables != nil {
			for k, v := range suite.Variables {
				suiteVars[k] = v
			}
		}

		// Execute BeforeEach hooks if they exist
		if len(def.BeforeEach) > 0 {
			r.Logger.Debug("Executing BeforeEach hooks", zap.Strings("hooks", def.BeforeEach))
			hookVars, err := r.executeHooks(def.BeforeEach, suiteVars)
			if err != nil {
				r.Logger.Error("Error executing BeforeEach hooks", zap.Error(err))
				// We continue execution despite hook errors
			}

			// Merge hook variables into suite variables
			for k, v := range hookVars {
				// Hook variables are added as definition-level variables (no prefix)
				suiteVars[k] = v
			}
		}

		// Set up the suite with variables
		
		// Pass variables to suite
		suite.Variables = suiteVars

		// Execute the test suite
		r.Logger.Debug(fmt.Sprintf("executing test suite: %s", suite.Name))
		suiteResult, err := suite.Run(r.Logger, r.HttpClient)
		if err != nil {
			r.Logger.Error("error executing test suite", zap.Error(err))
		}

		// Store variables from suite execution in the result
		suiteResult.Variables = suite.Variables

		// Execute AfterEach hooks if they exist
		if len(def.AfterEach) > 0 {
			r.Logger.Debug("Executing AfterEach hooks", zap.Strings("hooks", def.AfterEach))
			_, err := r.executeHooks(def.AfterEach, suite.Variables)
			if err != nil {
				r.Logger.Error("Error executing AfterEach hooks", zap.Error(err))
				// We continue execution despite hook errors
			}
		}

		result.Suites[suite.Name] = suiteResult
	}

	// Execute AfterAll hooks if they exist
	if len(def.AfterAll) > 0 {
		r.Logger.Debug("Executing AfterAll hooks", zap.Strings("hooks", def.AfterAll))
		_, err := r.executeHooks(def.AfterAll, def.Variables)
		if err != nil {
			r.Logger.Error("Error executing AfterAll hooks", zap.Error(err))
			// We continue execution despite hook errors
		}
	}

	return result, nil
}
