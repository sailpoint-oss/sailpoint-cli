package source

import (
	"strings"
	"testing"
)

func TestSourceDeleteRequiresForce(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetArgs([]string{"source-id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected source delete without --force to fail")
	}
	if !strings.Contains(err.Error(), "source delete requires --force") {
		t.Fatalf("unexpected error: %v", err)
	}
}
