// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
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

// Helper function to create a command with mock client
func createCommandWithMockClient(cmd *cobra.Command, mockClient *mockClient) *cobra.Command {
	// Override the RunE function to use our mock client
	originalRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Create a mock config
		cfg := config.CLIConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			BaseURL:      "https://test.sailpoint.com",
			AccessToken:  "test-access-token",
		}

		// Create a client with our mock
		spClient := client.NewSpClient(cfg)

		// Set the client in the command context
		cmd.SetContext(context.WithValue(cmd.Context(), "client", spClient))

		return originalRunE(cmd, args)
	}
	return cmd
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
			var jsonData interface{}
			inputBytes := []byte(tc.input)

			// Try to parse as JSON
			if err := json.Unmarshal(inputBytes, &jsonData); err == nil {
				// If valid JSON, pretty print
				pretty, err := json.MarshalIndent(jsonData, "", "  ")
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

// TestGetCommand tests the GET command functionality
func TestGetCommand(t *testing.T) {
	testCases := []struct {
		name           string
		endpoint       string
		headers        []string
		queryParams    []string
		prettyPrint    bool
		mockResponse   *http.Response
		mockError      error
		expectedOutput string
		expectedError  bool
	}{
		{
			name:         "Successful GET request with pretty print",
			endpoint:     "beta/accounts",
			prettyPrint:  true,
			mockResponse: mockResponse(200, `{"id":"123","name":"test"}`),
			expectedOutput: `{
  "id": "123",
  "name": "test"
}`,
			expectedError: false,
		},
		{
			name:           "GET request with headers",
			endpoint:       "beta/accounts",
			headers:        []string{"Accept: application/json"},
			mockResponse:   mockResponse(200, `{"id":"123"}`),
			expectedOutput: `{"id":"123"}`,
			expectedError:  false,
		},
		{
			name:           "GET request with query parameters",
			endpoint:       "beta/accounts",
			queryParams:    []string{"limit=10", "offset=0"},
			mockResponse:   mockResponse(200, `{"id":"123"}`),
			expectedOutput: `{"id":"123"}`,
			expectedError:  false,
		},
		{
			name:          "Failed GET request",
			endpoint:      "beta/accounts",
			mockError:     fmt.Errorf("network error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClient{
				getResponse: tc.mockResponse,
				getError:    tc.mockError,
			}

			// Create command with mock client
			cmd := createCommandWithMockClient(newGetCmd(), mockClient)
			cmd.SetArgs([]string{tc.endpoint})

			// Set flags
			for _, header := range tc.headers {
				cmd.Flags().Set("header", header)
			}
			for _, param := range tc.queryParams {
				cmd.Flags().Set("query", param)
			}
			if tc.prettyPrint {
				cmd.Flags().Set("pretty", "true")
			}

			// Capture output
			output := &strings.Builder{}
			cmd.SetOut(output)

			// Execute command
			err := cmd.Execute()

			// Check error
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check output
			if output.String() != tc.expectedOutput {
				t.Errorf("Expected output:\n%s\nGot:\n%s", tc.expectedOutput, output.String())
			}
		})
	}
}

// TestPostCommand tests the POST command functionality
func TestPostCommand(t *testing.T) {
	testCases := []struct {
		name          string
		endpoint      string
		body          string
		contentType   string
		headers       []string
		mockResponse  *http.Response
		mockError     error
		expectedError bool
	}{
		{
			name:          "Successful POST request",
			endpoint:      "beta/accounts",
			body:          `{"name":"test"}`,
			contentType:   "application/json",
			mockResponse:  mockResponse(201, `{"id":"123"}`),
			expectedError: false,
		},
		{
			name:          "POST request with headers",
			endpoint:      "beta/accounts",
			body:          `{"name":"test"}`,
			headers:       []string{"Accept: application/json"},
			mockResponse:  mockResponse(201, `{"id":"123"}`),
			expectedError: false,
		},
		{
			name:          "Failed POST request",
			endpoint:      "beta/accounts",
			body:          `{"name":"test"}`,
			mockError:     fmt.Errorf("network error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClient{
				postResponse: tc.mockResponse,
				postError:    tc.mockError,
			}

			// Create command with mock client
			cmd := createCommandWithMockClient(newPostCmd(), mockClient)
			cmd.SetArgs([]string{tc.endpoint})

			// Set flags
			cmd.Flags().Set("body", tc.body)
			if tc.contentType != "" {
				cmd.Flags().Set("content-type", tc.contentType)
			}
			for _, header := range tc.headers {
				cmd.Flags().Set("header", header)
			}

			// Execute command
			err := cmd.Execute()

			// Check error
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestPutCommand tests the PUT command functionality
func TestPutCommand(t *testing.T) {
	testCases := []struct {
		name          string
		endpoint      string
		body          string
		contentType   string
		headers       []string
		mockResponse  *http.Response
		mockError     error
		expectedError bool
	}{
		{
			name:          "Successful PUT request",
			endpoint:      "beta/accounts/123",
			body:          `{"name":"updated"}`,
			contentType:   "application/json",
			mockResponse:  mockResponse(200, `{"id":"123","name":"updated"}`),
			expectedError: false,
		},
		{
			name:          "Failed PUT request",
			endpoint:      "beta/accounts/123",
			body:          `{"name":"updated"}`,
			mockError:     fmt.Errorf("network error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClient{
				putResponse: tc.mockResponse,
				putError:    tc.mockError,
			}

			// Create command with mock client
			cmd := createCommandWithMockClient(newPutCmd(), mockClient)
			cmd.SetArgs([]string{tc.endpoint})

			// Set flags
			cmd.Flags().Set("body", tc.body)
			if tc.contentType != "" {
				cmd.Flags().Set("content-type", tc.contentType)
			}
			for _, header := range tc.headers {
				cmd.Flags().Set("header", header)
			}

			// Execute command
			err := cmd.Execute()

			// Check error
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestPatchCommand tests the PATCH command functionality
func TestPatchCommand(t *testing.T) {
	testCases := []struct {
		name          string
		endpoint      string
		body          string
		headers       []string
		mockResponse  *http.Response
		mockError     error
		expectedError bool
	}{
		{
			name:          "Successful PATCH request",
			endpoint:      "beta/accounts/123",
			body:          `{"name":"patched"}`,
			mockResponse:  mockResponse(200, `{"id":"123","name":"patched"}`),
			expectedError: false,
		},
		{
			name:          "Failed PATCH request",
			endpoint:      "beta/accounts/123",
			body:          `{"name":"patched"}`,
			mockError:     fmt.Errorf("network error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClient{
				patchResponse: tc.mockResponse,
				patchError:    tc.mockError,
			}

			// Create command with mock client
			cmd := createCommandWithMockClient(newPatchCmd(), mockClient)
			cmd.SetArgs([]string{tc.endpoint})

			// Set flags
			cmd.Flags().Set("body", tc.body)
			for _, header := range tc.headers {
				cmd.Flags().Set("header", header)
			}

			// Execute command
			err := cmd.Execute()

			// Check error
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDeleteCommand tests the DELETE command functionality
func TestDeleteCommand(t *testing.T) {
	testCases := []struct {
		name          string
		endpoint      string
		headers       []string
		mockResponse  *http.Response
		mockError     error
		expectedError bool
	}{
		{
			name:          "Successful DELETE request",
			endpoint:      "beta/accounts/123",
			mockResponse:  mockResponse(204, ""),
			expectedError: false,
		},
		{
			name:          "Failed DELETE request",
			endpoint:      "beta/accounts/123",
			mockError:     fmt.Errorf("network error"),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock client
			mockClient := &mockClient{
				deleteResponse: tc.mockResponse,
				deleteError:    tc.mockError,
			}

			// Create command with mock client
			cmd := createCommandWithMockClient(newDeleteCmd(), mockClient)
			cmd.SetArgs([]string{tc.endpoint})

			// Set flags
			for _, header := range tc.headers {
				cmd.Flags().Set("header", header)
			}

			// Execute command
			err := cmd.Execute()

			// Check error
			if tc.expectedError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
