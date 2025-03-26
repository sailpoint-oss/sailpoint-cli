// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
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
