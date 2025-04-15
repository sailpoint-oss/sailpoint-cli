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
			expectError:    false,
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"get", "/v2024/transforms/123"},
			expectedOutput: `{"detailCode":"404 Not found"`,
			expectError:    false,
		},
		{
			name:           "File_output",
			args:           []string{"get", "/v2024/transforms/123", "--output", "/dev/null"},
			expectedOutput: "Response saved to /dev/null",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetErr(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			output := b.String()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

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
			expectError:    false,
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"post", "/v2024/transforms", "--body", `{"name":"Test Transform","type":"dateFormat"}`},
			expectedOutput: `{"detailCode":"400.1.0 Required data missing or empty"`,
			expectError:    false,
		},
		{
			name:           "File_output",
			args:           []string{"post", "/v2024/transforms", "--output", "/dev/null", "--body", `{"name":"Test Transform","type":"dateFormat"}`},
			expectedOutput: "Response saved to /dev/null",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetErr(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			output := b.String()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

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
			expectError:    false,
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"put", "/v2024/transforms/123", "--body", `{"name":"Updated Transform","type":"dateFormat"}`},
			expectedOutput: `{"detailCode":"404 Not found"`,
			expectError:    false,
		},
		{
			name:           "File_output",
			args:           []string{"put", "/v2024/transforms/123", "--output", "/dev/null", "--body", `{"name":"Updated Transform","type":"dateFormat"}`},
			expectedOutput: "Response saved to /dev/null",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetErr(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			output := b.String()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

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
			expectError:    false,
		},
		{
			name:           "Full_output_with_status",
			args:           []string{"delete", "/v2024/transforms/123"},
			expectedOutput: `{"detailCode":"404 Not found"`,
			expectError:    false,
		},
		{
			name:           "File_output",
			args:           []string{"delete", "/v2024/transforms/123", "--output", "/dev/null"},
			expectedOutput: "Response saved to /dev/null",
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCmd()
			b := new(bytes.Buffer)
			cmd.SetOut(b)
			cmd.SetErr(b)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			output := b.String()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if !strings.Contains(output, tc.expectedOutput) {
					t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !strings.Contains(output, tc.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tc.expectedOutput, output)
			}
		})
	}
}
