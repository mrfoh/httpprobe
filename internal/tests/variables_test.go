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

func TestInterpolateVariableValues_CrossVariableReferences(t *testing.T) {
	variables := map[string]Variable{
		"base_url":       {Type: "string", Value: "https://api.example.com"},
		"users_endpoint": {Type: "string", Value: "${base_url}/users"},
	}

	err := InterpolateVariableValues(variables)
	if err != nil {
		t.Fatalf("InterpolateVariableValues returned an error: %v", err)
	}

	expected := "https://api.example.com/users"
	if variables["users_endpoint"].Value != expected {
		t.Errorf("users_endpoint = %q, want %q", variables["users_endpoint"].Value, expected)
	}
}

func TestInterpolateVariableValues_EnvAndCrossVariable(t *testing.T) {
	os.Setenv("TEST_HOST", "api.example.com")
	defer os.Unsetenv("TEST_HOST")

	variables := map[string]Variable{
		"base_url": {Type: "string", Value: "https://${env:TEST_HOST}"},
		"endpoint": {Type: "string", Value: "${base_url}/users"},
	}

	err := InterpolateVariableValues(variables)
	if err != nil {
		t.Fatalf("InterpolateVariableValues returned an error: %v", err)
	}

	if variables["base_url"].Value != "https://api.example.com" {
		t.Errorf("base_url = %q, want %q", variables["base_url"].Value, "https://api.example.com")
	}
}

func TestInterpolateVariableValues_NilMap(t *testing.T) {
	err := InterpolateVariableValues(nil)
	if err != nil {
		t.Fatalf("InterpolateVariableValues(nil) returned an error: %v", err)
	}
}

func TestInterpolateVariableValues_NoReferences(t *testing.T) {
	variables := map[string]Variable{
		"host": {Type: "string", Value: "localhost"},
		"port": {Type: "string", Value: "8080"},
	}

	err := InterpolateVariableValues(variables)
	if err != nil {
		t.Fatalf("InterpolateVariableValues returned an error: %v", err)
	}

	if variables["host"].Value != "localhost" {
		t.Errorf("host = %q, want %q", variables["host"].Value, "localhost")
	}
	if variables["port"].Value != "8080" {
		t.Errorf("port = %q, want %q", variables["port"].Value, "8080")
	}
}

func TestInterpolateVariableValues_UnresolvableReference(t *testing.T) {
	variables := map[string]Variable{
		"endpoint": {Type: "string", Value: "${nonexistent}/users"},
	}

	err := InterpolateVariableValues(variables)
	if err != nil {
		t.Fatalf("InterpolateVariableValues returned an error: %v", err)
	}

	// Unresolvable references should be left as-is
	if variables["endpoint"].Value != "${nonexistent}/users" {
		t.Errorf("endpoint = %q, want %q", variables["endpoint"].Value, "${nonexistent}/users")
	}
}

func TestCoerceVariableValue(t *testing.T) {
	tests := []struct {
		name     string
		variable Variable
		want     interface{}
		wantErr  bool
	}{
		{"int", Variable{Type: "int", Value: "42"}, 42, false},
		{"float", Variable{Type: "float", Value: "3.14"}, 3.14, false},
		{"bool true", Variable{Type: "bool", Value: "true"}, true, false},
		{"bool false", Variable{Type: "bool", Value: "false"}, false, false},
		{"string", Variable{Type: "string", Value: "hello"}, "hello", false},
		{"empty type", Variable{Type: "", Value: "hello"}, "hello", false},
		{"invalid int", Variable{Type: "int", Value: "abc"}, 0, true},
		{"invalid float", Variable{Type: "float", Value: "abc"}, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CoerceVariableValue(tt.variable)
			if (err != nil) != tt.wantErr {
				t.Errorf("CoerceVariableValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("CoerceVariableValue() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestInterpolateObject_TypedVariables(t *testing.T) {
	variables := map[string]Variable{
		"count":   {Type: "int", Value: "42"},
		"rate":    {Type: "float", Value: "3.14"},
		"active":  {Type: "bool", Value: "true"},
		"name":    {Type: "string", Value: "John"},
	}

	obj := map[string]interface{}{
		"count":  "${count}",
		"rate":   "${rate}",
		"active": "${active}",
		"name":   "${name}",
	}

	result, err := InterpolateObject(obj, variables)
	if err != nil {
		t.Fatalf("InterpolateObject returned an error: %v", err)
	}

	m := result.(map[string]interface{})

	if v, ok := m["count"].(int); !ok || v != 42 {
		t.Errorf("count = %v (%T), want 42 (int)", m["count"], m["count"])
	}
	if v, ok := m["rate"].(float64); !ok || v != 3.14 {
		t.Errorf("rate = %v (%T), want 3.14 (float64)", m["rate"], m["rate"])
	}
	if v, ok := m["active"].(bool); !ok || v != true {
		t.Errorf("active = %v (%T), want true (bool)", m["active"], m["active"])
	}
	if v, ok := m["name"].(string); !ok || v != "John" {
		t.Errorf("name = %v (%T), want John (string)", m["name"], m["name"])
	}
}

func TestInterpolateObject_MixedStringNotCoerced(t *testing.T) {
	variables := map[string]Variable{
		"count": {Type: "int", Value: "42"},
	}

	obj := map[string]interface{}{
		"label": "items_${count}",
	}

	result, err := InterpolateObject(obj, variables)
	if err != nil {
		t.Fatalf("InterpolateObject returned an error: %v", err)
	}

	m := result.(map[string]interface{})
	if v, ok := m["label"].(string); !ok || v != "items_42" {
		t.Errorf("label = %v (%T), want items_42 (string)", m["label"], m["label"])
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

