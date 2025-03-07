package tests

type TestResultWriter interface {
	Write(results map[string]TestDefinitionExecResult)
}

func NewResultWriter(output string) TestResultWriter {
	switch output {
	case "text":
		return NewTextResultWriter()
	case "table":
		return NewTableResultWriter()
	case "json":
		return NewJSONResultWriter()
	default:
		return NewTextResultWriter() // Default to text output
	}
}
