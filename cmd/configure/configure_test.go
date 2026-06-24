package configure

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetConfigureTestState(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestConfigureSetAndGetDebug(t *testing.T) {
	resetConfigureTestState(t)

	if err := setConfig("debug", "true"); err != nil {
		t.Fatalf("setConfig returned error: %v", err)
	}
	if got := viper.GetBool("debug"); !got {
		t.Fatal("expected debug to be true")
	}

	out := new(bytes.Buffer)
	if err := getConfig(out, "debug"); err != nil {
		t.Fatalf("getConfig returned error: %v", err)
	}
	if !strings.Contains(out.String(), "debug = true") {
		t.Fatalf("unexpected get output: %q", out.String())
	}
}

func TestConfigureRejectsInvalidDebugValue(t *testing.T) {
	resetConfigureTestState(t)

	err := setConfig("debug", "sometimes")
	if err == nil {
		t.Fatal("expected invalid debug value to fail")
	}
	if !strings.Contains(err.Error(), "invalid value for debug") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureUnknownKey(t *testing.T) {
	resetConfigureTestState(t)

	out := new(bytes.Buffer)
	err := getConfig(out, "unknown")
	if err == nil {
		t.Fatal("expected unknown key to fail")
	}
	if !strings.Contains(out.String(), "Unknown config key") {
		t.Fatalf("expected unknown-key help, got %q", out.String())
	}
}
