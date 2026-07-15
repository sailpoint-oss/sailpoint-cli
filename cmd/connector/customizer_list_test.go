// connector/list_test.go
package connector

import (
	"bytes"
	"testing"
)

func TestNewCustomizerListCmd_rejectsArgs(t *testing.T) {
	cmd := newCustomizerListCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"unexpected"}) // list takes no args

	if err := cmd.Execute(); err == nil {
		t.Error("expected error when unexpected args are passed")
	}
}
