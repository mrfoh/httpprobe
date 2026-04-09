package graphql

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// ValidationError represents a query validation error with location info
type ValidationError struct {
	Message string
	Line    int
	Column  int
	// SuiteName is the name of the suite containing the query
	SuiteName string
	// CaseTitle is the title of the test case containing the query
	CaseTitle string
	// Query is the query string that failed validation
	Query string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Column, e.Message)
}

// SchemaLoader loads a GraphQL schema from various sources
type SchemaLoader interface {
	Load() (*ast.Schema, error)
}

// FileSchemaLoader loads a GraphQL schema from a local SDL file
type FileSchemaLoader struct {
	Path string
}

func (l *FileSchemaLoader) Load() (*ast.Schema, error) {
	data, err := os.ReadFile(l.Path)
	if err != nil {
		return nil, fmt.Errorf("error reading schema file '%s': %w", l.Path, err)
	}

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  l.Path,
		Input: string(data),
	})
	if gqlErr != nil {
		return nil, fmt.Errorf("error parsing schema file '%s': %v", l.Path, gqlErr)
	}

	return schema, nil
}

// URLSchemaLoader loads a GraphQL schema by fetching SDL from an HTTP URL
type URLSchemaLoader struct {
	URL string
}

func (l *URLSchemaLoader) Load() (*ast.Schema, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(l.URL)
	if err != nil {
		return nil, fmt.Errorf("error fetching schema from URL '%s': %w", l.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching schema from URL '%s': status %d", l.URL, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading schema response from '%s': %w", l.URL, err)
	}

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  l.URL,
		Input: string(data),
	})
	if gqlErr != nil {
		return nil, fmt.Errorf("error parsing schema from URL '%s': %v", l.URL, gqlErr)
	}

	return schema, nil
}

// IntrospectionSchemaLoader loads a GraphQL schema via introspection query
type IntrospectionSchemaLoader struct {
	Endpoint string
	Headers  map[string]string
}

func (l *IntrospectionSchemaLoader) Load() (*ast.Schema, error) {
	// Build introspection request
	reqBody, err := json.Marshal(map[string]string{
		"query": IntrospectionQuery,
	})
	if err != nil {
		return nil, fmt.Errorf("error marshaling introspection query: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", l.Endpoint, io.NopCloser(
		io.Reader(
			&readCloserWrapper{data: reqBody},
		),
	))
	if err != nil {
		return nil, fmt.Errorf("error creating introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range l.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing introspection query against '%s': %w", l.Endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("introspection query returned status %d from '%s'", resp.StatusCode, l.Endpoint)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading introspection response: %w", err)
	}

	// Parse the introspection response and convert to SDL
	sdl, err := introspectionJSONToSDL(respBody)
	if err != nil {
		return nil, fmt.Errorf("error converting introspection response to SDL: %w", err)
	}

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  l.Endpoint,
		Input: sdl,
	})
	if gqlErr != nil {
		return nil, fmt.Errorf("error parsing introspection schema from '%s': %v", l.Endpoint, gqlErr)
	}

	return schema, nil
}

// readCloserWrapper wraps a byte slice as an io.Reader
type readCloserWrapper struct {
	data   []byte
	offset int
}

func (r *readCloserWrapper) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

// Validator validates GraphQL queries against a loaded schema
type Validator struct {
	Schema *ast.Schema
}

// NewValidator creates a Validator with the given schema
func NewValidator(schema *ast.Schema) *Validator {
	return &Validator{Schema: schema}
}

// ValidateQuery validates a single GraphQL query string against the schema
func (v *Validator) ValidateQuery(query string) []ValidationError {
	doc, gqlErr := gqlparser.LoadQuery(v.Schema, query)
	if gqlErr != nil {
		var errors []ValidationError
		for _, e := range gqlErr {
			ve := ValidationError{
				Message: e.Message,
				Query:   query,
			}
			if len(e.Locations) > 0 {
				ve.Line = e.Locations[0].Line
				ve.Column = e.Locations[0].Column
			}
			errors = append(errors, ve)
		}
		return errors
	}

	// doc is non-nil if parsing succeeded — no additional validation needed
	_ = doc
	return nil
}

// NewSchemaLoader creates a SchemaLoader based on the source type and path
func NewSchemaLoader(source, path string) (SchemaLoader, error) {
	switch source {
	case "file":
		return &FileSchemaLoader{Path: path}, nil
	case "url":
		return &URLSchemaLoader{URL: path}, nil
	case "introspection":
		return &IntrospectionSchemaLoader{Endpoint: path}, nil
	default:
		return nil, fmt.Errorf("unknown schema source type: '%s' (expected: file, url, or introspection)", source)
	}
}

// introspectionJSONToSDL converts a GraphQL introspection JSON response to SDL format.
// This is a simplified conversion that handles common types.
func introspectionJSONToSDL(data []byte) (string, error) {
	var response struct {
		Data struct {
			Schema struct {
				QueryType        *typeName   `json:"queryType"`
				MutationType     *typeName   `json:"mutationType"`
				SubscriptionType *typeName   `json:"subscriptionType"`
				Types            []sdlType   `json:"types"`
				Directives       []directive `json:"directives"`
			} `json:"__schema"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("error parsing introspection JSON: %w", err)
	}

	schema := response.Data.Schema
	var sdl string

	// Build schema block
	sdl += "schema {\n"
	if schema.QueryType != nil {
		sdl += fmt.Sprintf("  query: %s\n", schema.QueryType.Name)
	}
	if schema.MutationType != nil {
		sdl += fmt.Sprintf("  mutation: %s\n", schema.MutationType.Name)
	}
	if schema.SubscriptionType != nil {
		sdl += fmt.Sprintf("  subscription: %s\n", schema.SubscriptionType.Name)
	}
	sdl += "}\n\n"

	// Build type definitions
	for _, t := range schema.Types {
		// Skip built-in types
		if len(t.Name) > 0 && t.Name[0] == '_' && len(t.Name) > 1 && t.Name[1] == '_' {
			continue
		}

		switch t.Kind {
		case "OBJECT":
			sdl += fmt.Sprintf("type %s", t.Name)
			if len(t.Interfaces) > 0 {
				var ifaces []string
				for _, iface := range t.Interfaces {
					ifaces = append(ifaces, renderTypeRef(iface))
				}
				sdl += " implements " + joinStrings(ifaces, " & ")
			}
			sdl += " {\n"
			for _, f := range t.Fields {
				sdl += fmt.Sprintf("  %s", f.Name)
				if len(f.Args) > 0 {
					sdl += "("
					for i, arg := range f.Args {
						if i > 0 {
							sdl += ", "
						}
						sdl += fmt.Sprintf("%s: %s", arg.Name, renderTypeRef(arg.Type))
						if arg.DefaultValue != nil {
							sdl += fmt.Sprintf(" = %v", *arg.DefaultValue)
						}
					}
					sdl += ")"
				}
				sdl += fmt.Sprintf(": %s\n", renderTypeRef(f.Type))
			}
			sdl += "}\n\n"

		case "INPUT_OBJECT":
			sdl += fmt.Sprintf("input %s {\n", t.Name)
			for _, f := range t.InputFields {
				sdl += fmt.Sprintf("  %s: %s", f.Name, renderTypeRef(f.Type))
				if f.DefaultValue != nil {
					sdl += fmt.Sprintf(" = %v", *f.DefaultValue)
				}
				sdl += "\n"
			}
			sdl += "}\n\n"

		case "INTERFACE":
			sdl += fmt.Sprintf("interface %s {\n", t.Name)
			for _, f := range t.Fields {
				sdl += fmt.Sprintf("  %s: %s\n", f.Name, renderTypeRef(f.Type))
			}
			sdl += "}\n\n"

		case "UNION":
			sdl += fmt.Sprintf("union %s = ", t.Name)
			var types []string
			for _, pt := range t.PossibleTypes {
				types = append(types, renderTypeRef(pt))
			}
			sdl += joinStrings(types, " | ")
			sdl += "\n\n"

		case "ENUM":
			sdl += fmt.Sprintf("enum %s {\n", t.Name)
			for _, ev := range t.EnumValues {
				sdl += fmt.Sprintf("  %s\n", ev.Name)
			}
			sdl += "}\n\n"

		case "SCALAR":
			// Skip built-in scalars
			if t.Name == "String" || t.Name == "Int" || t.Name == "Float" || t.Name == "Boolean" || t.Name == "ID" {
				continue
			}
			sdl += fmt.Sprintf("scalar %s\n\n", t.Name)
		}
	}

	return sdl, nil
}

type typeName struct {
	Name string `json:"name"`
}

type sdlType struct {
	Kind          string      `json:"kind"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Fields        []sdlField  `json:"fields"`
	InputFields   []sdlInput  `json:"inputFields"`
	Interfaces    []typeRef   `json:"interfaces"`
	EnumValues    []enumValue `json:"enumValues"`
	PossibleTypes []typeRef   `json:"possibleTypes"`
}

type sdlField struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Args              []sdlInput `json:"args"`
	Type              typeRef   `json:"type"`
	IsDeprecated      bool      `json:"isDeprecated"`
	DeprecationReason *string   `json:"deprecationReason"`
}

type sdlInput struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         typeRef `json:"type"`
	DefaultValue *string `json:"defaultValue"`
}

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   *string  `json:"name"`
	OfType *typeRef `json:"ofType"`
}

type enumValue struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	IsDeprecated      bool    `json:"isDeprecated"`
	DeprecationReason *string `json:"deprecationReason"`
}

type directive struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Locations   []string   `json:"locations"`
	Args        []sdlInput `json:"args"`
}

func renderTypeRef(ref typeRef) string {
	switch ref.Kind {
	case "NON_NULL":
		if ref.OfType != nil {
			return renderTypeRef(*ref.OfType) + "!"
		}
	case "LIST":
		if ref.OfType != nil {
			return "[" + renderTypeRef(*ref.OfType) + "]"
		}
	default:
		if ref.Name != nil {
			return *ref.Name
		}
	}
	return "Unknown"
}

func joinStrings(s []string, sep string) string {
	result := ""
	for i, str := range s {
		if i > 0 {
			result += sep
		}
		result += str
	}
	return result
}
