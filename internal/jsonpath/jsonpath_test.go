// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package jsonpath

import (
	"testing"
)

func TestEvaluateJSONPath(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected string
	}{
		{
			name:     "simple object access",
			json:     `{"name": "test", "value": 123}`,
			path:     "$.name",
			expected: `"test"`,
		},
		{
			name:     "nested object access",
			json:     `{"user": {"name": "test", "age": 30}}`,
			path:     "$.user.name",
			expected: `"test"`,
		},
		{
			name:     "array access",
			json:     `{"items": [{"id": 1}, {"id": 2}]}`,
			path:     "$.items[0].id",
			expected: `1`,
		},
		{
			name:     "array wildcard",
			json:     `{"items": [{"id": 1}, {"id": 2}]}`,
			path:     "$.items[*].id",
			expected: `[1,2]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := EvaluateJSONPath([]byte(tc.json), tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(result) != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, string(result))
			}
		})
	}
}

func TestEvaluateJSONPathToString(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected string
	}{
		{
			name:     "simple string value",
			json:     `{"name": "test", "value": 123}`,
			path:     "$.name",
			expected: "test",
		},
		{
			name:     "number value",
			json:     `{"name": "test", "value": 123}`,
			path:     "$.value",
			expected: "123",
		},
		{
			name:     "nested string value",
			json:     `{"user": {"name": "test", "age": 30}}`,
			path:     "$.user.name",
			expected: "test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := EvaluateJSONPathToString([]byte(tc.json), tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}
