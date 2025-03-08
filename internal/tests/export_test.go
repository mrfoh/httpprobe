package tests

import (
	"testing"

	"github.com/mrfoh/httpprobe/internal/logging"
	"github.com/mrfoh/httpprobe/pkg/easyreq"
)

func TestProcessBodyExports(t *testing.T) {
	// Create a test logger
	logger := logging.NewMockLogger()

	// Create a mock response body
	responseBody := `{
		"success": true,
		"data": {
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			"refresh_token": "refresh-token-value",
			"expires_in": 3600,
			"user": {
				"id": 123,
				"username": "testuser",
				"profile": {
					"full_name": "Test User",
					"email": "test@example.com"
				}
			},
			"permissions": ["read", "write"]
		}
	}`

	// Create a mock HTTP response
	resp := &easyreq.HttpResponse{
		Status: 200,
		Body:   []byte(responseBody),
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}

	// Create test suite with variables
	suite := &TestSuite{
		Variables: map[string]Variable{
			"existing_var": {Type: "string", Value: "existing-value"},
		},
	}

	// Create a test request with exports
	request := &Request{
		Export: RequestExport{
			Body: []BodyExport{
				{Path: "$.data.token", As: "access_token"},
				{Path: "$.data.refresh_token", As: "refresh_token"},
				{Path: "$.data.expires_in", As: "token_expires"},
				{Path: "$.data.user.profile.email", As: "user_email"},
				{Path: "$.data.permissions", As: "permissions"},
				{Path: "$.data.user.id", As: "user_id"},
				{Path: "$.nonexistent", As: "should_not_exist"},
			},
		},
	}

	// Call the function under test
	err := processBodyExports(request, resp, suite, logger)
	if err != nil {
		t.Fatalf("processBodyExports returned error: %v", err)
	}

	// Check that variables were correctly exported
	expectedExports := map[string]string{
		"access_token":  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"refresh_token": "refresh-token-value",
		"token_expires": "3600",
		"user_email":    "test@example.com",
		"permissions":   `["read","write"]`,
		"user_id":       "123",
	}

	for key, expectedValue := range expectedExports {
		if variable, exists := suite.Variables[key]; !exists {
			t.Errorf("Expected variable %s to be exported, but it wasn't", key)
		} else if variable.Value != expectedValue {
			t.Errorf("Variable %s has value %s, expected %s", key, variable.Value, expectedValue)
		}
	}

	// Check that non-existent paths don't create variables
	if _, exists := suite.Variables["should_not_exist"]; exists {
		t.Errorf("Variable for non-existent path was created but shouldn't have been")
	}

	// Check that existing variables are preserved
	if existingVar, exists := suite.Variables["existing_var"]; !exists || existingVar.Value != "existing-value" {
		t.Errorf("Existing variable was not preserved correctly")
	}

	// Test error case with invalid JSON
	resp.Body = []byte("not a json")
	err = processBodyExports(request, resp, suite, logger)
	if err == nil {
		t.Errorf("Expected error with invalid JSON, but got nil")
	}

	// Test with invalid JSONPath
	resp.Body = []byte(responseBody)
	request.Export.Body = []BodyExport{
		{Path: "$[invalid", As: "invalid_path"},
	}
	err = processBodyExports(request, resp, suite, logger)
	if err == nil {
		t.Errorf("Expected error with invalid JSONPath, but got nil")
	}
}
