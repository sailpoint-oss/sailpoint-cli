// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Mock client for testing
type mockClient struct {
	getResponse    *http.Response
	postResponse   *http.Response
	putResponse    *http.Response
	patchResponse  *http.Response
	deleteResponse *http.Response
	getError       error
	postError      error
	putError       error
	patchError     error
	deleteError    error
}

func (m *mockClient) Get(ctx context.Context, url string) (*http.Response, error) {
	return m.getResponse, m.getError
}

func (m *mockClient) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	return m.postResponse, m.postError
}

func (m *mockClient) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	return m.putResponse, m.putError
}

func (m *mockClient) Patch(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return m.patchResponse, m.patchError
}

func (m *mockClient) Delete(ctx context.Context, url string, params map[string]string) (*http.Response, error) {
	return m.deleteResponse, m.deleteError
}

// Helper function to create a mock response
func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// TestQueryParameterParsing tests the parsing of query parameters
func TestQueryParameterParsing(t *testing.T) {
	// Test cases
	testCases := []struct {
		name       string
		queryParam string
		expectKey  string
		expectVal  string
		expectErr  bool
	}{
		{
			name:       "Valid query parameter",
			queryParam: "key=value",
			expectKey:  "key",
			expectVal:  "value",
			expectErr:  false,
		},
		{
			name:       "Query parameter with equals in value",
			queryParam: "key=value=with=equals",
			expectKey:  "key",
			expectVal:  "value=with=equals",
			expectErr:  false,
		},
		{
			name:       "Invalid query parameter - no equals",
			queryParam: "invalid-format",
			expectErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parts := strings.SplitN(tc.queryParam, "=", 2)
			if len(parts) != 2 {
				if !tc.expectErr {
					t.Errorf("Expected parsing to succeed but failed for: %s", tc.queryParam)
				}
				return
			}

			if tc.expectErr {
				t.Errorf("Expected parsing to fail but succeeded for: %s", tc.queryParam)
				return
			}

			if parts[0] != tc.expectKey {
				t.Errorf("Expected key %s, got %s", tc.expectKey, parts[0])
			}

			if parts[1] != tc.expectVal {
				t.Errorf("Expected value %s, got %s", tc.expectVal, parts[1])
			}
		})
	}
}

// TestHeaderParsing tests the parsing of header values
func TestHeaderParsing(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		header    string
		expectKey string
		expectVal string
		expectErr bool
	}{
		{
			name:      "Valid header",
			header:    "Content-Type: application/json",
			expectKey: "Content-Type",
			expectVal: "application/json",
			expectErr: false,
		},
		{
			name:      "Header with colon in value",
			header:    "Authorization: Bearer: token",
			expectKey: "Authorization",
			expectVal: "Bearer: token",
			expectErr: false,
		},
		{
			name:      "Invalid header - no colon",
			header:    "invalid-format",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parts := strings.SplitN(tc.header, ":", 2)
			if len(parts) != 2 {
				if !tc.expectErr {
					t.Errorf("Expected parsing to succeed but failed for: %s", tc.header)
				}
				return
			}

			if tc.expectErr {
				t.Errorf("Expected parsing to fail but succeeded for: %s", tc.header)
				return
			}

			key := strings.TrimSpace(parts[0])
			if key != tc.expectKey {
				t.Errorf("Expected key %s, got %s", tc.expectKey, key)
			}

			val := strings.TrimSpace(parts[1])
			if val != tc.expectVal {
				t.Errorf("Expected value %s, got %s", tc.expectVal, val)
			}
		})
	}
}

// TestPrettyPrintJSON tests the JSON pretty-printing feature
func TestPrettyPrintJSON(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple JSON object",
			input:    `{"key":"value"}`,
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:     "Nested JSON object",
			input:    `{"outer":{"inner":"value"}}`,
			expected: "{\n  \"outer\": {\n    \"inner\": \"value\"\n  }\n}",
		},
		{
			name:     "JSON array",
			input:    `[1,2,3]`,
			expected: "[\n  1,\n  2,\n  3\n]",
		},
		{
			name:     "Invalid JSON",
			input:    `not json`,
			expected: "not json", // Should return the original if not valid JSON
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data interface{}
			inputBytes := []byte(tc.input)

			// Try to parse as JSON
			if err := json.Unmarshal(inputBytes, &data); err == nil {
				// If valid JSON, pretty print
				pretty, err := json.MarshalIndent(data, "", "  ")
				if err == nil {
					inputBytes = pretty
				}
			}

			// Compare with expected output
			if string(inputBytes) != tc.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tc.expected, string(inputBytes))
			}
		})
	}
}

// TestEndpointNormalization tests that endpoints are normalized correctly
func TestEndpointNormalization(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Already has leading slash",
			input:    "/beta/accounts",
			expected: "/beta/accounts",
		},
		{
			name:     "Missing leading slash",
			input:    "beta/accounts",
			expected: "/beta/accounts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			endpoint := tc.input
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

			if endpoint != tc.expected {
				t.Errorf("Expected endpoint %s, got %s", tc.expected, endpoint)
			}
		})
	}
}
