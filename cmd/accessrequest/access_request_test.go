package accessrequest

import "testing"

func TestNewAccessRequestCommandRegistersSubcommands(t *testing.T) {
	cmd := NewAccessRequestCommand()
	for _, name := range []string{"list", "get", "create", "cancel", "approve", "close", "work-items"} {
		found, _, err := cmd.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected access-request subcommand %q to exist", name)
		}
	}

	workItems, _, err := cmd.Find([]string{"work-items"})
	if err != nil {
		t.Fatalf("expected work-items command: %v", err)
	}
	for _, name := range []string{"list", "get", "approve", "reject", "forward", "complete"} {
		found, _, err := workItems.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected work-items subcommand %q to exist", name)
		}
	}
}
