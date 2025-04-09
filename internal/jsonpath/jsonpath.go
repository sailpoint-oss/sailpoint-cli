// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package jsonpath

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sailpoint-oss/jsonslice"
)

// EvaluateJSONPath evaluates a JSONPath expression against a JSON document
func EvaluateJSONPath(jsonData []byte, path string) ([]byte, error) {
	// Ensure path starts with $
	if !strings.HasPrefix(path, "$") {
		path = "$" + path
	}

	// Parse the JSON data
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Evaluate the JSONPath expression
	result, err := jsonslice.Get(jsonData, path)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate JSONPath: %w", err)
	}

	return result, nil
}

// EvaluateJSONPathToString evaluates a JSONPath expression and returns the result as a string
func EvaluateJSONPathToString(jsonData []byte, path string) (string, error) {
	result, err := EvaluateJSONPath(jsonData, path)
	if err != nil {
		return "", err
	}

	// Check if result is a primitive value (string, number, boolean)
	var data interface{}
	if err := json.Unmarshal(result, &data); err != nil {
		return "", fmt.Errorf("failed to parse JSONPath result: %w", err)
	}

	// If it's a string, clean it up
	if str, ok := data.(string); ok {
		// Remove any newlines
		str = strings.TrimSpace(str)
		return str, nil
	}

	// For arrays, objects, or other types, return the JSON as is
	return string(result), nil
}
