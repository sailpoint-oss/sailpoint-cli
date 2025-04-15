// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Make API requests to SailPoint",
	}
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newPostCmd())
	cmd.AddCommand(newPutCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

func TestGetOutputFormat(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "JSONPath_output_only",
			args:           []string{"get", "/v2024/transforms/123", "--jsonpath", "$.name"},
			expectedOutput: "Test Transform",
			expectError:    true, // Expect error due to 404
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"get", "/v2024/transforms/123"},
			expectedOutput: "Status: 404 Not Found",
			expectError:    true,
		},
		{
			name:           "File_output",
			args:           []string{"get", "/v2024/transforms/123", "--output", "output.json"},
			expectedOutput: "Response saved to output.json",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output := b.String()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
			}
		})
	}
}

func TestPostOutputFormat(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "JSONPath_output_only",
			args:           []string{"post", "/v2024/transforms", "--jsonpath", "$.id", "--body", `{"name":"Test Transform","type":"dateFormat"}`},
			expectedOutput: "123",
			expectError:    true, // Expect error due to 400
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"post", "/v2024/transforms", "--body", `{"name":"Test Transform","type":"dateFormat"}`},
			expectedOutput: "Status: 400 Bad Request",
			expectError:    true,
		},
		{
			name:           "File_output",
			args:           []string{"post", "/v2024/transforms", "--output", "post_output.json", "--body", `{"name":"Test Transform","type":"dateFormat"}`},
			expectedOutput: "Response saved to post_output.json",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output := b.String()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
			}
		})
	}
}

func TestPutOutputFormat(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "JSONPath_output_only",
			args:           []string{"put", "/v2024/transforms/123", "--jsonpath", "$.name", "--body", `{"name":"Updated Transform","type":"dateFormat"}`},
			expectedOutput: "Updated Transform",
			expectError:    true, // Expect error due to 404
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"put", "/v2024/transforms/123", "--body", `{"name":"Updated Transform","type":"dateFormat"}`},
			expectedOutput: "Status: 404 Not Found",
			expectError:    true,
		},
		{
			name:           "File_output",
			args:           []string{"put", "/v2024/transforms/123", "--output", "put_output.json", "--body", `{"name":"Updated Transform","type":"dateFormat"}`},
			expectedOutput: "Response saved to put_output.json",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output := b.String()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
			}
		})
	}
}

func TestDeleteOutputFormat(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "JSONPath_output_only",
			args:           []string{"delete", "/v2024/transforms/123", "--jsonpath", "$.message"},
			expectedOutput: "Transform deleted",
			expectError:    true, // Expect error due to 404
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"delete", "/v2024/transforms/123"},
			expectedOutput: "Status: 404 Not Found",
			expectError:    true,
		},
		{
			name:           "File_output",
			args:           []string{"delete", "/v2024/transforms/123", "--output", "delete_output.json"},
			expectedOutput: "Response saved to delete_output.json",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			output := b.String()
			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
			}
		})
	}
}
