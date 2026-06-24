package identity

import (
	"strings"
	"testing"
)

func TestNewIdentityCommandRegistersSubcommands(t *testing.T) {
	cmd := NewIdentityCommand()
	for _, name := range []string{"list", "get", "entitlements", "sync", "reset", "process"} {
		found, _, err := cmd.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected identity subcommand %q to exist", name)
		}
	}
}

func TestIdentityResetRequiresForce(t *testing.T) {
	cmd := newResetCommand()
	cmd.SetArgs([]string{"identity-id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected reset without --force to fail")
	}
	if !strings.Contains(err.Error(), "reset requires --force") {
		t.Fatalf("unexpected error: %v", err)
	}
}
