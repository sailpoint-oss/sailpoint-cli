package ui_plugins

import (
	"strings"
	"testing"
)

func TestNewUIPluginsCommandStructure(t *testing.T) {
	cmd := NewUIPluginsCommand()

	if cmd.Use != "ui-plugins" {
		t.Fatalf("expected use to be ui-plugins, got %s", cmd.Use)
	}

	if !cmd.Hidden {
		t.Fatal("expected ui-plugins command to be hidden")
	}

	if len(cmd.Commands()) != 9 {
		t.Fatalf("expected 9 subcommands, got %d", len(cmd.Commands()))
	}
}

func TestUIPluginsGateDisabled(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "")
	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected command to fail when experimental gate is disabled")
	}

	errText := err.Error()
	if !strings.Contains(errText, "experimental") || !strings.Contains(errText, experimentalUIPluginsEnvVar) {
		t.Fatalf("expected error to mention experimental gate and env var, got: %s", errText)
	}
}

func TestUIPluginsGateEnabled(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected command to run when gate is enabled, got: %v", err)
	}
}
