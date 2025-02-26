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
)

type Runner[R any] struct {
	Parser     tests.TestDefinitionParser
	Logger     logging.Logger
	HttpClient interface{}
	Pool       pond.ResultPool[R]
}

func NewRunner[R any](parser tests.TestDefinitionParser, logger logging.Logger, testPool pond.ResultPool[R]) TestRunner {
	return &Runner[R]{
		Parser: parser,
		Logger: logger,
		Pool:   testPool,
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

func (r *Runner[R]) Execute(definition []*tests.TestDefinition) (ExecutionResult, error) {
	for _, def := range definition {
		r.Pool.SubmitErr(func() (R, error) {
			return r.testDefinitionWorker(def)
		})
	}

	return ExecutionResult{}, nil
}

// Reads test definitions from the specified path
func (r *Runner[R]) GetTestDefinitions(params *GetTestDefinitionsParams) ([]*tests.TestDefinition, error) {
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
		r.Logger.Info(fmt.Sprintf("found %d test definition(s)", len(definitions)))
	}

	return definitions, nil
}

// processTestFile reads and parses a single test file
func (r *Runner[R]) processTestFile(path string) (*tests.TestDefinition, error) {
	data, err := openFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", path, err)
	}

	ext := filepath.Ext(path)
	def, err := r.Parser.Parse(data, ext)
	if err != nil {
		return nil, fmt.Errorf("error parsing file %s: %v", path, err)
	}

	return def, nil
}

func (r *Runner[R]) testDefinitionWorker(def *tests.TestDefinition) (R, error) {
	r.Logger.Debug(fmt.Sprintf("executing test: %s", def.Name))
	return {}, nil
}
