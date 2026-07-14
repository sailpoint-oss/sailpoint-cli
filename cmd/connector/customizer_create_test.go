// connector/create_test.go
package connector

import (
	"bytes"
	"testing"
)

func TestNewCustomizerCreateCmd_missingArg(t *testing.T) {
	cmd := newCustomizerCreateCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{}) // no <customizer-name>

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when customizer name arg is missing")
	}
}
