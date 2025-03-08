package tests

type TestResultWriter interface {
	Write(results map[string]TestDefinitionExecResult)
}

func NewResultWriter(outputType string, outputFile string) TestResultWriter {
	switch outputType {
	case "text":
		return NewTextResultWriter()
	case "table":
		return NewTableResultWriter()
	case "json":
		return NewJSONResultWriter(outputFile)
	default:
		return NewTextResultWriter() // Default to text output
	}
}
