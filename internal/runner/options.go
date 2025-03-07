package runner

import (
	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/internal/tests"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
)

type TestRunnerOptions struct {
	// Logger is the logger to use
	Logger logging.Logger
	// TestDefinitionParser is the parser used to parse test definitions
	Parser     tests.TestDefinitionParser
	HttpClient easyreq.HttpClient
	// Number of concurrent test definitions to execute
	Concurrency int
	// ResultWriter is the writer to use for writing test results
	Writer tests.TestResultWriter
}

func NewOptions() *TestRunnerOptions {
	return &TestRunnerOptions{}
}

func (o *TestRunnerOptions) Validate() error {
	return nil
}

func (o *TestRunnerOptions) SetLogger(logger logging.Logger) *TestRunnerOptions {
	o.Logger = logger
	return o
}

func (o *TestRunnerOptions) SetParser(parser tests.TestDefinitionParser) *TestRunnerOptions {
	o.Parser = parser
	return o
}

func (o *TestRunnerOptions) SetConcurrency(concurrency int) *TestRunnerOptions {
	o.Concurrency = concurrency
	return o
}

func (o *TestRunnerOptions) SetHttpClient(httpClient easyreq.HttpClient) *TestRunnerOptions {
	o.HttpClient = httpClient
	return o
}

func (o *TestRunnerOptions) SetResultWriter(writer tests.TestResultWriter) *TestRunnerOptions {
	o.Writer = writer
	return o
}
