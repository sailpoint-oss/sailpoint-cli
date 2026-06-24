package auth

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/viper"
)

func resetAuthCommandState(t *testing.T) {
	t.Helper()
	viper.Reset()
	config.ClearActiveEnvironmentOverride()
	t.Cleanup(func() {
		viper.Reset()
		config.ClearActiveEnvironmentOverride()
	})
}

func TestAuthStatusNoActiveEnvironment(t *testing.T) {
	resetAuthCommandState(t)
	viper.Set("activeenvironment", "")

	cmd := newStatusCommand()
	out := new(bytes.Buffer)
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected status to succeed without active environment: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "No active environment configured") {
		t.Fatalf("expected no-active-environment message, got %q", got)
	}
}

func TestAuthLogoutNoActiveEnvironment(t *testing.T) {
	resetAuthCommandState(t)
	viper.Set("activeenvironment", "")

	cmd := newLogoutCommand()

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected logout to fail without active environment")
	}
	if !strings.Contains(err.Error(), "no active environment configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewAuthCommandRegistersSubcommands(t *testing.T) {
	cmd := NewAuthCommand()
	for _, name := range []string{"login", "logout", "pat", "status"} {
		found, _, err := cmd.Find([]string{name})
		if err != nil || found == nil || found.Name() != name {
			t.Fatalf("expected auth subcommand %q to exist", name)
		}
	}
}
