package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	keyring "github.com/zalando/go-keyring"
)

func mintedToken(t *testing.T, claims map[string]any) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString([]byte("sig"))
}

func futureToken(t *testing.T) string {
	t.Helper()
	return mintedToken(t, map[string]any{
		"org": "acme",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

func clearRuntimeToken(t *testing.T) {
	t.Helper()
	previous := runtimeAccessToken
	runtimeAccessToken = ""
	t.Cleanup(func() { runtimeAccessToken = previous })
}

func TestParseBearerTokenUsesMintedTokenAsIs(t *testing.T) {
	clearRuntimeToken(t)
	token := futureToken(t)

	for _, raw := range []string{token, "Bearer " + token, "bearer " + token, `"` + token + `"`, "'Bearer " + token + "'"} {
		got, err := parseBearerToken(raw)
		if err != nil {
			t.Fatalf("parseBearerToken(%q) error = %v", raw, err)
		}
		if got != token {
			t.Fatalf("parseBearerToken(%q) = %q, want the minted access token", raw, got)
		}
	}
}

func TestParseBearerTokenRejectsExpiredAndMalformed(t *testing.T) {
	expired := mintedToken(t, map[string]any{"exp": time.Now().Add(-time.Hour).Unix()})
	_, err := parseBearerToken(expired)
	if err == nil || !strings.Contains(err.Error(), "expired at") || !strings.Contains(err.Error(), accessTokenEnvVar) || !strings.Contains(err.Error(), "sail set bearer") {
		t.Fatalf("parseBearerToken(expired) error = %v, want expiry guidance", err)
	}

	_, err = parseBearerToken("not-a-token")
	if err == nil || !strings.Contains(err.Error(), "not a signed Identity Security Cloud access token") || !strings.Contains(err.Error(), "does not exchange a client ID and secret") {
		t.Fatalf("parseBearerToken(malformed) error = %v", err)
	}

	_, err = parseBearerToken("   ")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("parseBearerToken(blank) error = %v", err)
	}
}

func TestParseBearerTokenAllowsTokenWithoutExpiry(t *testing.T) {
	token := mintedToken(t, map[string]any{"org": "acme"})
	got, err := parseBearerToken(token)
	if err != nil {
		t.Fatalf("parseBearerToken() error = %v", err)
	}
	if got != token {
		t.Fatalf("parseBearerToken() = %q, want %q", got, token)
	}
}

func TestGetBearerTokenFromEnvironment(t *testing.T) {
	clearRuntimeToken(t)
	token := futureToken(t)
	t.Setenv(accessTokenEnvVar, "Bearer "+token)
	SetRuntimeAccessToken("ignored")
	t.Cleanup(func() { SetRuntimeAccessToken("") })

	got, err := GetBearerToken()
	if err != nil {
		t.Fatalf("GetBearerToken() error = %v", err)
	}
	if got != token {
		t.Fatalf("GetBearerToken() = %q, want the environment token", got)
	}
}

func TestGetBearerTokenFromFlag(t *testing.T) {
	clearRuntimeToken(t)
	t.Setenv(accessTokenEnvVar, "")
	token := futureToken(t)
	SetRuntimeAccessToken(token)
	t.Cleanup(func() { SetRuntimeAccessToken("") })

	got, err := GetBearerToken()
	if err != nil {
		t.Fatalf("GetBearerToken() error = %v", err)
	}
	if got != token {
		t.Fatalf("GetBearerToken() = %q, want the flag token", got)
	}
}

func TestAuthTypeInUsePrefersExplicitBearerToken(t *testing.T) {
	resetViper(t)
	clearRuntimeToken(t)
	SetAuthType("pat")

	if got := authTypeInUse(); got != "pat" {
		t.Fatalf("authTypeInUse() = %q, want pat", got)
	}

	t.Setenv(accessTokenEnvVar, futureToken(t))
	if got := authTypeInUse(); got != "bearer" {
		t.Fatalf("authTypeInUse() = %q, want bearer when %s is set", got, accessTokenEnvVar)
	}
}

func TestMissingBearerTokenError(t *testing.T) {
	err := missingBearerTokenError("alpha", keyring.ErrNotFound)
	for _, want := range []string{`bearer token is stored for environment "alpha"`, "sail set bearer", accessTokenEnvVar, "--access-token", "OAuth token endpoint"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("missingBearerTokenError() = %q, want it to contain %q", err, want)
		}
	}
}

func TestValidateBearerFromEnvironment(t *testing.T) {
	resetViper(t)
	clearRuntimeToken(t)
	SetAuthType("pat")
	t.Setenv("SAIL_BASE_URL", "https://acme.api.identitynow.com")
	t.Setenv("SAIL_CLIENT_ID", "")
	t.Setenv("SAIL_CLIENT_SECRET", "")
	t.Setenv(accessTokenEnvVar, futureToken(t))

	if err := Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want bearer token from the environment to skip the client ID and secret", err)
	}
}

func TestValidateBearerRejectsExpiredToken(t *testing.T) {
	resetViper(t)
	clearRuntimeToken(t)
	SetAuthType("bearer")
	t.Setenv("SAIL_BASE_URL", "https://acme.api.identitynow.com")
	t.Setenv(accessTokenEnvVar, mintedToken(t, map[string]any{"exp": time.Now().Add(-time.Minute).Unix()}))

	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "expired at") {
		t.Fatalf("Validate() error = %v, want an expired token error", err)
	}
}

func TestCheckEnvironmentMentionsBearerToken(t *testing.T) {
	resetViper(t)
	t.Setenv("SAIL_BASE_URL", "")

	err := CheckEnvironment()
	if err == nil || !strings.Contains(err.Error(), accessTokenEnvVar) || !strings.Contains(err.Error(), "sail env create") {
		t.Fatalf("CheckEnvironment() error = %v, want bearer token guidance", err)
	}
}

func TestValidateBearerRequiresBaseURL(t *testing.T) {
	resetViper(t)
	clearRuntimeToken(t)
	viper.Set("environments.alpha.tenanturl", "https://alpha.identitynow.com")
	SetActiveEnvironment("alpha")
	SetAuthType("bearer")
	t.Setenv("SAIL_BASE_URL", "")
	t.Setenv(accessTokenEnvVar, futureToken(t))

	err := Validate()
	if err == nil || !strings.Contains(err.Error(), "configuration invalid") {
		t.Fatalf("Validate() error = %v, want a missing BaseURL error", err)
	}
}
