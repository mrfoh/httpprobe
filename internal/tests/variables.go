package tests

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"math/rand"
)

// init initializes the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

// LoadEnvFile loads environment variables from a file
// Environment variables in the file should be in the format KEY=VALUE
func LoadEnvFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		// If file doesn't exist, just return nil (not an error)
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error opening env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first equals sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format in env file at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) > 1 && (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Set environment variable
		err := os.Setenv(key, value)
		if err != nil {
			return fmt.Errorf("error setting environment variable %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}

	return nil
}

// InterpolateVariables replaces variable references in the input string
// with their corresponding values from the variables map
func InterpolateVariables(input string, variables map[string]Variable) (string, error) {
	if input == "" {
		return input, nil
	}
	
	result := input
	
	// First, handle environment variables: ${env:VAR_NAME}
	envVarPattern := regexp.MustCompile(`\${env:([^}]+)}`)
	result = envVarPattern.ReplaceAllStringFunc(result, func(match string) string {
		submatches := envVarPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		
		envName := submatches[1]
		envValue := os.Getenv(envName)
		if envValue == "" {
			// If environment variable doesn't exist, leave the original reference
			return match
		}
		
		return envValue
	})
	
	// Then, handle regular variables: ${variable_name}
	for name, variable := range variables {
		pattern := "${" + name + "}"
		result = strings.Replace(result, pattern, variable.Value, -1)
	}
	
	// Finally, handle function calls: ${functionName(arg1,arg2,...)}
	funcPattern := regexp.MustCompile(`\${([a-zA-Z]+)\(([^)]*)\)}`)
	result = funcPattern.ReplaceAllStringFunc(result, func(match string) string {
		submatches := funcPattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		
		funcName := submatches[1]
		argsStr := submatches[2]
		args := strings.Split(argsStr, ",")
		
		// Process different functions
		switch funcName {
		case "random":
			return processRandomFunction(args)
		case "timestamp":
			return processTimestampFunction(args)
		default:
			return match // Unknown function, no change
		}
	})
	
	return result, nil
}

// processRandomFunction generates a random string of the specified length
func processRandomFunction(args []string) string {
	// Default length
	length := 10
	
	// Parse length if provided
	if len(args) > 0 && args[0] != "" {
		parsedLen, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err == nil && parsedLen > 0 {
			length = parsedLen
		}
	}
	
	// Generate random string
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	
	return string(result)
}

// processTimestampFunction generates a timestamp in the specified format
func processTimestampFunction(args []string) string {
	// Default format
	format := "2006-01-02T15:04:05Z"
	
	// Parse format if provided
	if len(args) > 0 && args[0] != "" {
		format = strings.TrimSpace(args[0])
	}
	
	return time.Now().Format(format)
}

// InterpolateObject recursively interpolates variables in an object (map, slice, or scalar value)
func InterpolateObject(obj interface{}, variables map[string]Variable) (interface{}, error) {
	switch v := obj.(type) {
	case string:
		return InterpolateVariables(v, variables)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			interpolatedKey, err := InterpolateVariables(key, variables)
			if err != nil {
				return nil, fmt.Errorf("error interpolating map key: %w", err)
			}
			
			interpolatedVal, err := InterpolateObject(val, variables)
			if err != nil {
				return nil, fmt.Errorf("error interpolating map value: %w", err)
			}
			
			result[interpolatedKey] = interpolatedVal
		}
		return result, nil
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			interpolatedVal, err := InterpolateObject(val, variables)
			if err != nil {
				return nil, fmt.Errorf("error interpolating slice element: %w", err)
			}
			result[i] = interpolatedVal
		}
		return result, nil
	default:
		// Return other types as-is
		return v, nil
	}
}

// InterpolateRequest applies variable interpolation to all fields in a Request
func InterpolateRequest(request *Request, variables map[string]Variable) error {
	var err error
	
	// Interpolate URL
	request.URL, err = InterpolateVariables(request.URL, variables)
	if err != nil {
		return fmt.Errorf("error interpolating URL: %w", err)
	}
	
	// Interpolate headers
	for i := range request.Headers {
		request.Headers[i].Key, err = InterpolateVariables(request.Headers[i].Key, variables)
		if err != nil {
			return fmt.Errorf("error interpolating header key: %w", err)
		}
		
		request.Headers[i].Value, err = InterpolateVariables(request.Headers[i].Value, variables)
		if err != nil {
			return fmt.Errorf("error interpolating header value: %w", err)
		}
	}
	
	// Interpolate body
	if request.Body.Type == "json" && request.Body.Data != nil {
		// Handle string JSON body
		if strData, ok := request.Body.Data.(string); ok {
			interpolated, err := InterpolateVariables(strData, variables)
			if err != nil {
				return fmt.Errorf("error interpolating body string: %w", err)
			}
			request.Body.Data = interpolated
		} else {
			// Handle structured JSON body
			interpolated, err := InterpolateObject(request.Body.Data, variables)
			if err != nil {
				return fmt.Errorf("error interpolating body object: %w", err)
			}
			request.Body.Data = interpolated
		}
	}
	
	return nil
}