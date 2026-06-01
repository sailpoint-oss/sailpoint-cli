package config

import (
	"testing"

	"github.com/spf13/viper"
)

func resetViperForTest(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestGetActiveEnvironmentDefaultsWithoutInitConfig(t *testing.T) {
	resetViperForTest(t)

	if got := GetActiveEnvironment(); got != "default" {
		t.Fatalf("GetActiveEnvironment() = %q, want %q", got, "default")
	}
}

func TestGetAuthTypeDefaultsWithoutInitConfig(t *testing.T) {
	resetViperForTest(t)

	if got := GetAuthType(); got != "pat" {
		t.Fatalf("GetAuthType() = %q, want %q", got, "pat")
	}
}

func TestGetAuthTypeReadsEnvironmentWithoutInitConfig(t *testing.T) {
	resetViperForTest(t)
	t.Setenv("SAIL_AUTHTYPE", "OAUTH")

	if got := GetAuthType(); got != "oauth" {
		t.Fatalf("GetAuthType() = %q, want %q", got, "oauth")
	}
}

func TestGetEnvAuthTypeDefaultsWithoutInitConfig(t *testing.T) {
	resetViperForTest(t)

	if got := GetEnvAuthType("default"); got != "pat" {
		t.Fatalf("GetEnvAuthType() = %q, want %q", got, "pat")
	}
}
