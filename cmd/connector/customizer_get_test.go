// connector/get_test.go
package connector

import (
	"bytes"
	"testing"
)

func TestNewCustomizerGetCmd_missingRequiredFlags(t *testing.T) {
	cmd := newCustomizerGetCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{}) // no -c

	if err := cmd.Execute(); err == nil {
		t.Error("expected command to fail when -c is missing")
	}
}
