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

	// Convert JSONPath syntax to jsonslice syntax
	// jsonslice uses a simpler syntax where:
	// - .name becomes name
	// - .items[0].id becomes items[0].id
	// - .items[*].id becomes items[*].id
	path = strings.TrimPrefix(path, ".")

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

	// Remove quotes if the result is a string
	str := string(result)
	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
		str = str[1 : len(str)-1]
	}

	return str, nil
}
