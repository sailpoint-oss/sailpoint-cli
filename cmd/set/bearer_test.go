package set

import (
	"io"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/viper"
)

func TestAuthCommandSetsBearer(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)

	cmd := newAuthCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"Bearer"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := config.GetAuthType(); got != "bearer" {
		t.Fatalf("GetAuthType() = %q, want bearer", got)
	}
}

func TestBearerCommandRejectsUnmintedToken(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	config.SetAuthType("pat")

	cmd := newBearerCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--token", "not-a-token"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not a signed Identity Security Cloud access token") {
		t.Fatalf("Execute() error = %v, want a rejected token", err)
	}
	if got := config.GetAuthType(); got != "pat" {
		t.Fatalf("GetAuthType() = %q, want pat after a rejected token", got)
	}
}
