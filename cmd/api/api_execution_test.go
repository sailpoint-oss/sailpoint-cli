// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestTenantEndpoint(t *testing.T) {
	// Create the GET command
	cmd := newGetCmd()

	// Set the endpoint
	cmd.SetArgs([]string{"v2024/tenant"})

	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Failed to execute command: %v", err)
	}
}

func TestListTransformations(t *testing.T) {
	// Create the GET command
	cmd := newGetCmd()

	// Set the endpoint
	cmd.SetArgs([]string{"/v2024/transforms"})

	cmd.Flags().Set("pretty", "true")
	cmd.Flags().Set("query", "limit=2")

	// Capture stdout
	buffer := new(bytes.Buffer)
	cmd.SetOut(buffer)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("TestNewCreateCmd: Unable to execute the command successfully: %v", err)
	}

	// Read the output
	responseBytes, err := io.ReadAll(buffer)
	if err != nil {
		t.Fatalf("Error reading stdout: %v", err)
	}

	// Count the number of transformations in the response
	transformations := strings.Count(string(responseBytes), "id")

	// Check if the response is as expected
	if transformations != 2 {
		t.Errorf("Expected 2 transformations in the response, but got: %d", transformations)
	}
}
