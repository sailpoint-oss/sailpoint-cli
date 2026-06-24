package accessprofile

import (
	"strings"
	"testing"
)

func TestAccessProfileDeleteRequiresForce(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetArgs([]string{"access-profile-id"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected access-profile delete without --force to fail")
	}
	if !strings.Contains(err.Error(), "access-profile delete requires --force") {
		t.Fatalf("unexpected error: %v", err)
	}
}
