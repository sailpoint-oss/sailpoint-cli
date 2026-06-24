package account

import "testing"

func TestNewAccountCommandRegistersSubcommands(t *testing.T) {
	cmd := NewAccountCommand()
	for _, name := range []string{"list", "get", "entitlements", "enable", "disable", "unlock", "reload"} {
		found, _, err := cmd.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected account subcommand %q to exist", name)
		}
	}
}
