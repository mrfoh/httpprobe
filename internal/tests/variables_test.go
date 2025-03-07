package tests

import (
	"os"
	"testing"
)

func TestInterpolateVariables(t *testing.T) {
	variables := map[string]Variable{
		"name": {Type: "string", Value: "John"},
		"api_url": {Type: "string", Value: "https://api.example.com"},
	}

	// Test simple variable replacement
	input := "Hello, ${name}!"
	expected := "Hello, John!"
	result, err := InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result != expected {
		t.Errorf("InterpolateVariables(%q) = %q, want %q", input, result, expected)
	}

	// Test URL with variable
	input = "${api_url}/users"
	expected = "https://api.example.com/users"
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result != expected {
		t.Errorf("InterpolateVariables(%q) = %q, want %q", input, result, expected)
	}

	// Test environment variable
	os.Setenv("TEST_API_KEY", "test-key-123")
	input = "${env:TEST_API_KEY}"
	expected = "test-key-123"
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result != expected {
		t.Errorf("InterpolateVariables(%q) = %q, want %q", input, result, expected)
	}

	// Test non-existent environment variable
	input = "${env:NONEXISTENT_VAR}"
	expected = "${env:NONEXISTENT_VAR}"  // Should keep the original reference
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result != expected {
		t.Errorf("InterpolateVariables(%q) = %q, want %q", input, result, expected)
	}

	// Test random function
	input = "${random(5)}"
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("InterpolateVariables(%q) result length = %d, want %d", input, len(result), 5)
	}
	if result == input {
		t.Errorf("Random function was not evaluated: %q", result)
	}

	// Test timestamp function
	input = "${timestamp(2006-01-02)}"
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result == input {
		t.Errorf("Timestamp function was not evaluated: %q", result)
	}

	// Test multiple variables in one string
	input = "${api_url}/users/${name}?key=${env:TEST_API_KEY}"
	expected = "https://api.example.com/users/John?key=test-key-123"
	result, err = InterpolateVariables(input, variables)
	if err != nil {
		t.Errorf("InterpolateVariables returned an error: %v", err)
	}
	if result != expected {
		t.Errorf("InterpolateVariables(%q) = %q, want %q", input, result, expected)
	}
}

func TestInterpolateRequest(t *testing.T) {
	variables := map[string]Variable{
		"api_url": {Type: "string", Value: "https://api.example.com"},
		"user_id": {Type: "string", Value: "123"},
		"token": {Type: "string", Value: "abc-token"},
	}

	request := Request{
		Method: "GET",
		URL: "${api_url}/users/${user_id}",
		Headers: []RequestHeader{
			{Key: "Authorization", Value: "Bearer ${token}"},
			{Key: "X-User-ID", Value: "${user_id}"},
		},
		Body: RequestBody{
			Type: "json",
			Data: `{"id": "${user_id}", "token": "${token}"}`,
		},
	}

	err := InterpolateRequest(&request, variables)
	if err != nil {
		t.Errorf("InterpolateRequest returned an error: %v", err)
	}

	// Check URL interpolation
	expectedURL := "https://api.example.com/users/123"
	if request.URL != expectedURL {
		t.Errorf("InterpolateRequest URL = %q, want %q", request.URL, expectedURL)
	}

	// Check header interpolation
	if len(request.Headers) != 2 {
		t.Errorf("InterpolateRequest header count = %d, want 2", len(request.Headers))
	} else {
		if request.Headers[0].Value != "Bearer abc-token" {
			t.Errorf("Header[0] value = %q, want %q", request.Headers[0].Value, "Bearer abc-token")
		}
		if request.Headers[1].Value != "123" {
			t.Errorf("Header[1] value = %q, want %q", request.Headers[1].Value, "123")
		}
	}

	// Check body interpolation
	expectedBody := `{"id": "123", "token": "abc-token"}`
	if bodyStr, ok := request.Body.Data.(string); ok {
		if bodyStr != expectedBody {
			t.Errorf("Body = %q, want %q", bodyStr, expectedBody)
		}
	} else {
		t.Errorf("Body is not a string: %v", request.Body.Data)
	}
}