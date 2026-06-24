package role

import (
	"strings"
	"testing"
)

func TestRoleDeleteRequiresForce(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetArgs([]string{"role-id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected role delete without --force to fail")
	}
	if !strings.Contains(err.Error(), "role delete requires --force") {
		t.Fatalf("unexpected error: %v", err)
	}
}
