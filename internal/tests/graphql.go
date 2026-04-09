package tests

import (
	"encoding/json"
	"fmt"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
	"github.com/oliveagle/jsonpath"
	"go.uber.org/zap"
)

// GraphQLEnvelope represents the standard GraphQL request envelope
type GraphQLEnvelope struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// BuildGraphQLBody creates a GraphQLEnvelope from a RequestBody with type "graphql"
func BuildGraphQLBody(body *RequestBody) *GraphQLEnvelope {
	envelope := &GraphQLEnvelope{
		Query:     body.Query,
		Variables: body.Variables,
	}
	if body.OperationName != "" {
		envelope.OperationName = body.OperationName
	}
	return envelope
}

// processGraphQLExports extracts values from response.data using JSONPath
// and adds them to the suite variables for use in subsequent test cases.
// Paths are relative to response.data, not the full response body.
func processGraphQLExports(request *Request, resp *easyreq.HttpResponse, suite *TestSuite, logger logging.Logger) error {
	if len(request.Export.GraphQL) == 0 {
		return nil
	}

	// Initialize variables map if it doesn't exist
	if suite.Variables == nil {
		suite.Variables = make(map[string]Variable)
	}

	// Parse full response body
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(resp.Body, &bodyMap); err != nil {
		return fmt.Errorf("error parsing response body for graphql exports: %w", err)
	}

	// Extract the "data" subtree
	dataRaw, ok := bodyMap["data"]
	if !ok || dataRaw == nil {
		return fmt.Errorf("response does not contain 'data' field for graphql exports")
	}

	dataMap, ok := dataRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("response 'data' field is not an object for graphql exports")
	}

	// Process each export using JSONPath against the data subtree
	for _, export := range request.Export.GraphQL {
		logger.Debug("Processing graphql export",
			zap.String("path", export.Path),
			zap.String("as", export.As))

		if export.Path == "" || export.As == "" {
			logger.Warn("Skipping graphql export with empty path or variable name")
			continue
		}

		// Extract value using JSONPath against the data subtree
		path, err := jsonpath.Compile(export.Path)
		if err != nil {
			return fmt.Errorf("invalid JSONPath '%s': %w", export.Path, err)
		}

		value, err := path.Lookup(dataMap)
		if err != nil {
			logger.Warn("Error extracting value using JSONPath",
				zap.String("path", export.Path),
				zap.Error(err))
			continue
		}

		// Convert the value to string for storage as a variable
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case float64:
			strValue = fmt.Sprintf("%g", v)
		case int:
			strValue = fmt.Sprintf("%d", v)
		case bool:
			strValue = fmt.Sprintf("%t", v)
		case nil:
			strValue = ""
		default:
			bytes, err := json.Marshal(v)
			if err != nil {
				logger.Warn("Error marshaling complex value to JSON",
					zap.String("path", export.Path),
					zap.Any("value", v),
					zap.Error(err))
				strValue = fmt.Sprintf("%v", v)
			} else {
				strValue = string(bytes)
			}
		}

		suite.Variables[export.As] = Variable{
			Type:  "string",
			Value: strValue,
		}

		logger.Debug("Exported graphql response value to variable",
			zap.String("variable", export.As),
			zap.String("value", strValue))
	}

	return nil
}
