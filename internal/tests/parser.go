package tests

import (
	"encoding/json"
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type TestDefinitionParser interface {
	// Parses an input file and returns a TestDefinition
	Parse(data []byte, ext string) (*TestDefinition, error)
}

type TestDefinitionParserImpl struct{}

func NewTestDefinitionParser() TestDefinitionParser {
	return &TestDefinitionParserImpl{}
}

func (p *TestDefinitionParserImpl) ParseYaml(data []byte, ext string) (*TestDefinition, error) {
	var def TestDefinition
	err := yaml.Unmarshal(data, &def)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}
	return &def, nil
}

func (p *TestDefinitionParserImpl) ParseJson(data []byte, ext string) (*TestDefinition, error) {
	var def TestDefinition
	err := json.Unmarshal(data, &def)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	return &def, nil
}

func (p *TestDefinitionParserImpl) Parse(data []byte, ext string) (*TestDefinition, error) {
	supportedExts := []string{".yaml", ".json"}
	if !slices.Contains(supportedExts, ext) {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	switch ext {
	case ".yaml":
		return p.ParseYaml(data, ext)
	case ".json":
		return p.ParseJson(data, ext)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}
