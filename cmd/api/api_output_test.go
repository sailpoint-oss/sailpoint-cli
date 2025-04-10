// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGetOutputFormat(t *testing.T) {
	// Create a test file with JSON content
	testFile := "test_output.json"
	testContent := `{"name": "Test Transform", "id": "123", "type": "test"}`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	tests := []struct {
		name           string
		jsonPath       string
		outputFile     string
		expectedOutput string
	}{
		{
			name:           "JSONPath output only",
			jsonPath:       "$.name",
			expectedOutput: "Test Transform",
		},
		{
			name:           "Full output with status",
			jsonPath:       "",
			expectedOutput: testContent + "\nStatus: 200 OK",
		},
		{
			name:           "File output",
			outputFile:     "output.json",
			expectedOutput: "Response saved to output.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the GET command
			cmd := newGetCmd()
			buffer := new(bytes.Buffer)
			cmd.SetOut(buffer)

			// Set up command arguments
			cmd.SetArgs([]string{"/v2024/transforms/123"})
			if tt.jsonPath != "" {
				cmd.Flags().Set("jsonpath", tt.jsonPath)
			}
			if tt.outputFile != "" {
				cmd.Flags().Set("output", tt.outputFile)
				defer os.Remove(tt.outputFile)
			}

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Read the output
			output, err := io.ReadAll(buffer)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			// Clean up the output by removing any log messages
			outputStr := string(output)
			outputStr = strings.TrimSpace(outputStr)

			// Verify the output matches expected
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, outputStr)
			}

			// Verify no log messages are in the output
			if strings.Contains(outputStr, "Making GET request") {
				t.Error("Output contains log message when it shouldn't")
			}
		})
	}
}

func TestPostOutputFormat(t *testing.T) {
	// Create a test file with JSON content
	testFile := "test_post.json"
	testContent := `{"name": "Test Transform", "type": "test"}`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	tests := []struct {
		name           string
		jsonPath       string
		outputFile     string
		expectedOutput string
	}{
		{
			name:           "JSONPath output only",
			jsonPath:       "$.id",
			expectedOutput: "123",
		},
		{
			name:           "Full output with status",
			jsonPath:       "",
			expectedOutput: `{"id": "123", "name": "Test Transform"}` + "\nStatus: 201 Created",
		},
		{
			name:           "File output",
			outputFile:     "post_output.json",
			expectedOutput: "Response saved to post_output.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the POST command
			cmd := newPostCmd()
			buffer := new(bytes.Buffer)
			cmd.SetOut(buffer)

			// Set up command arguments
			cmd.SetArgs([]string{"/v2024/transforms"})
			cmd.Flags().Set("body-file", testFile)
			if tt.jsonPath != "" {
				cmd.Flags().Set("jsonpath", tt.jsonPath)
			}
			if tt.outputFile != "" {
				cmd.Flags().Set("output", tt.outputFile)
				defer os.Remove(tt.outputFile)
			}

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Read the output
			output, err := io.ReadAll(buffer)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			// Clean up the output by removing any log messages
			outputStr := string(output)
			outputStr = strings.TrimSpace(outputStr)

			// Verify the output matches expected
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, outputStr)
			}

			// Verify no log messages are in the output
			if strings.Contains(outputStr, "Making POST request") {
				t.Error("Output contains log message when it shouldn't")
			}
		})
	}
}

func TestPutOutputFormat(t *testing.T) {
	// Create a test file with JSON content
	testFile := "test_put.json"
	testContent := `{"name": "Updated Transform", "type": "test"}`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	tests := []struct {
		name           string
		jsonPath       string
		outputFile     string
		expectedOutput string
	}{
		{
			name:           "JSONPath output only",
			jsonPath:       "$.name",
			expectedOutput: "Updated Transform",
		},
		{
			name:           "Full output with status",
			jsonPath:       "",
			expectedOutput: testContent + "\nStatus: 200 OK",
		},
		{
			name:           "File output",
			outputFile:     "put_output.json",
			expectedOutput: "Response saved to put_output.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the PUT command
			cmd := newPutCmd()
			buffer := new(bytes.Buffer)
			cmd.SetOut(buffer)

			// Set up command arguments
			cmd.SetArgs([]string{"/v2024/transforms/123"})
			cmd.Flags().Set("body-file", testFile)
			if tt.jsonPath != "" {
				cmd.Flags().Set("jsonpath", tt.jsonPath)
			}
			if tt.outputFile != "" {
				cmd.Flags().Set("output", tt.outputFile)
				defer os.Remove(tt.outputFile)
			}

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Read the output
			output, err := io.ReadAll(buffer)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			// Clean up the output by removing any log messages
			outputStr := string(output)
			outputStr = strings.TrimSpace(outputStr)

			// Verify the output matches expected
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, outputStr)
			}

			// Verify no log messages are in the output
			if strings.Contains(outputStr, "Making PUT request") {
				t.Error("Output contains log message when it shouldn't")
			}
		})
	}
}

func TestDeleteOutputFormat(t *testing.T) {
	tests := []struct {
		name           string
		jsonPath       string
		outputFile     string
		expectedOutput string
	}{
		{
			name:           "JSONPath output only",
			jsonPath:       "$.message",
			expectedOutput: "Transform deleted",
		},
		{
			name:           "Full output with status",
			jsonPath:       "",
			expectedOutput: `{"message": "Transform deleted"}` + "\nStatus: 204 No Content",
		},
		{
			name:           "File output",
			outputFile:     "delete_output.json",
			expectedOutput: "Response saved to delete_output.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the DELETE command
			cmd := newDeleteCmd()
			buffer := new(bytes.Buffer)
			cmd.SetOut(buffer)

			// Set up command arguments
			cmd.SetArgs([]string{"/v2024/transforms/123"})
			if tt.jsonPath != "" {
				cmd.Flags().Set("jsonpath", tt.jsonPath)
			}
			if tt.outputFile != "" {
				cmd.Flags().Set("output", tt.outputFile)
				defer os.Remove(tt.outputFile)
			}

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Read the output
			output, err := io.ReadAll(buffer)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			// Clean up the output by removing any log messages
			outputStr := string(output)
			outputStr = strings.TrimSpace(outputStr)

			// Verify the output matches expected
			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, outputStr)
			}

			// Verify no log messages are in the output
			if strings.Contains(outputStr, "Making DELETE request") {
				t.Error("Output contains log message when it shouldn't")
			}
		})
	}
}
