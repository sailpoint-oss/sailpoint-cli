package ui_plugins

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewUIPluginsCommandStructure(t *testing.T) {
	cmd := NewUIPluginsCommand()

	if cmd.Use != "ui-plugins" {
		t.Fatalf("expected use to be ui-plugins, got %s", cmd.Use)
	}

	if cmd.Hidden {
		t.Fatal("expected ui-plugins command to be visible")
	}

	if len(cmd.Commands()) != 11 {
		t.Fatalf("expected 11 subcommands, got %d", len(cmd.Commands()))
	}

	for _, name := range []string{"disable", "enable"} {
		if !hasSubcommand(cmd, name) {
			t.Fatalf("expected %q subcommand to be registered", name)
		}
	}
}

// hasSubcommand reports whether cmd has a direct subcommand with the given name.
func hasSubcommand(cmd *cobra.Command, name string) bool {
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return true
		}
	}
	return false
}

func TestUIPluginsCommandRuns(t *testing.T) {
	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to run, got: %v", err)
	}
}
