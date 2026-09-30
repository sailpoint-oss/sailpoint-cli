package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/viper"
	keyring "github.com/zalando/go-keyring"
)

// resetViper clears all configuration for the test and again when it ends.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestCheckEnvironmentNoEnvironments(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "")

	err := CheckEnvironment()
	if err == nil || !strings.Contains(err.Error(), "no environment is configured") || !strings.Contains(err.Error(), "sail env create") {
		t.Fatalf("CheckEnvironment() error = %v, want no-environment guidance", err)
	}
}

func TestCheckEnvironmentUnknownActive(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "")
	viper.Set("environments.beta.baseurl", "https://beta.api.identitynow.com")
	viper.Set("environments.alpha.baseurl", "https://alpha.api.identitynow.com")
	SetActiveEnvironment("gone")

	err := CheckEnvironment()
	if err == nil {
		t.Fatal("CheckEnvironment() error = nil, want error")
	}
	for _, want := range []string{`"gone" does not exist`, "alpha, beta", "sail env use"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("CheckEnvironment() error = %q, want it to contain %q", err, want)
		}
	}
}

func TestCheckEnvironmentNoActive(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "")
	viper.Set("environments.alpha.baseurl", "https://alpha.api.identitynow.com")
	SetActiveEnvironment("")

	err := CheckEnvironment()
	if err == nil || !strings.Contains(err.Error(), "no active environment is selected") {
		t.Fatalf("CheckEnvironment() error = %v, want no-active guidance", err)
	}
}

func TestCheckEnvironmentValid(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "")
	viper.Set("environments.alpha.baseurl", "https://alpha.api.identitynow.com")
	SetActiveEnvironment("Alpha")

	if err := CheckEnvironment(); err != nil {
		t.Fatalf("CheckEnvironment() error = %v", err)
	}
}

func TestCheckEnvironmentEnvVars(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "https://acme.api.identitynow.com")

	if err := CheckEnvironment(); err != nil {
		t.Fatalf("CheckEnvironment() error = %v, want nil when SAIL_BASE_URL is set", err)
	}
}

func TestMissingSecretError(t *testing.T) {
	err := missingSecretError("client ID", "SAIL_CLIENT_ID", "alpha", keyring.ErrNotFound)
	for _, want := range []string{`client ID is stored for environment "alpha"`, "sail set pat", "SAIL_CLIENT_ID"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missingSecretError() = %q, want it to contain %q", err, want)
		}
	}

	other := errors.New("keychain locked")
	if err := missingSecretError("client ID", "SAIL_CLIENT_ID", "alpha", other); !errors.Is(err, other) {
		t.Errorf("missingSecretError() = %v, want it to wrap the keyring error", err)
	}
}

func TestPatTokenErrorInvalidClient(t *testing.T) {
	body := []byte(`{"error_description":"Full authentication is required to access this resource","error":"invalid_client"}`)

	err := patTokenError(400, body, "devrel", false)
	for _, want := range []string{"rejected the PAT client ID or client secret", `"devrel"`, "sail set pat"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("patTokenError() = %q, want it to contain %q", err, want)
		}
	}

	err = patTokenError(400, body, "devrel", true)
	if !strings.Contains(err.Error(), "SAIL_CLIENT_ID and SAIL_CLIENT_SECRET") {
		t.Errorf("patTokenError() = %q, want environment variable guidance", err)
	}
}

func TestPatTokenErrorOther(t *testing.T) {
	err := patTokenError(500, []byte(`{"error":"server_error","error_description":"try again later"}`), "devrel", false)
	if !strings.Contains(err.Error(), "status 500") || !strings.Contains(err.Error(), "try again later") {
		t.Errorf("patTokenError() = %q, want status and description", err)
	}

	err = patTokenError(502, []byte("bad gateway"), "devrel", false)
	if !strings.Contains(err.Error(), "bad gateway") {
		t.Errorf("patTokenError() = %q, want raw body", err)
	}
}
