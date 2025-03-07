package runner

import "github.com/mrfoh/httpprobe/internal/tests"

type TestRunner interface {
	GetTestDefinitions(params *GetTestDefinitionsParams) ([]*tests.TestDefinition, error)
	Execute(definition []*tests.TestDefinition) (map[string]tests.TestDefinitionExecResult, error)
	Write(results map[string]tests.TestDefinitionExecResult)
}

type GetTestDefinitionsParams struct {
	// SearchPath is the path to search for test files
	SearchPath string
	// FileExtensions is a list of file extensions to include in the search
	FileExtensions []string
}
