// connector/update_test.go
package connector

import (
	"bytes"
	"testing"
)

func TestNewCustomizerUpdateCmd_missingFlags(t *testing.T) {
	// neither -c nor -n
	cmd := newCustomizerUpdateCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when flags are missing")
	}

	// just -c
	cmd = newCustomizerUpdateCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-c", "cust-1"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -n is missing")
	}

	// just -n
	cmd = newCustomizerUpdateCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"-n", "NewName"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error when -c is missing")
	}
}
