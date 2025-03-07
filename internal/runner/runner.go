package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
}

func NewRunner(opts *TestRunnerOptions) TestRunner {
	return &Runner{
		Parser:       opts.Parser,
		Logger:       opts.Logger,
		HttpClient:   opts.HttpClient,
		Concurrency:  opts.Concurrency,
		ResultWriter: opts.Writer,
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
	// Execute test definitions concurrently
	pool := pond.NewResultPool[tests.TestDefinitionExecResult](r.Concurrency, pond.WithQueueSize(len(definition)))

	for _, def := range definition {
		test := pool.SubmitErr(func(def *tests.TestDefinition) func() (tests.TestDefinitionExecResult, error) {
			return func() (tests.TestDefinitionExecResult, error) {
				return r.executeTestDefinition(def)
			}
		}(def))

		testResult, err := test.Wait()
		if err != nil {
			return nil, err
		}

		// Store result
		result[def.Name] = testResult
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
	result := tests.TestDefinitionExecResult{
		Path:   def.Path,
		Suites: make(map[string]tests.TestSuiteResult, len(def.Suites)),
	}

	r.Logger.Debug(fmt.Sprintf("executing test definition: %s", def.Name))

	// Make a copy of the suites to avoid modifying the original
	for i := range def.Suites {
		// Create a local copy of the suite to avoid issues with the loop variable
		suite := def.Suites[i]
		
		// Pass variables from test definition to suite
		suite.Variables = def.Variables
		
		r.Logger.Debug(fmt.Sprintf("executing test suite: %s", suite.Name))
		suiteResult, err := suite.Run(r.Logger, r.HttpClient)
		if err != nil {
			r.Logger.Error("error executing test suite", zap.Error(err))
		}

		result.Suites[suite.Name] = suiteResult
	}

	return result, nil
}
