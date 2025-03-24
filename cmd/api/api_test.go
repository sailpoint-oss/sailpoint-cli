// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

// TestCommandExecutionFlow tests that the command execution flow works without errors
func TestCommandExecutionFlow(t *testing.T) {
	cmd := NewAPICommand()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Check that help output is generated
	out := b.String()
	if out == "" || len(out) == 0 {
		t.Error("Expected help output to not be empty")
	}
}

// Helper function for executing commands in tests
func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}
