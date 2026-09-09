package config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func buildPasteCode(t *testing.T, payload pastePayload) string {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return pastePrefix + base64.RawURLEncoding.EncodeToString(raw)
}

func TestConfirmationCodeFromState(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  string
	}{
		{name: "first eight characters", state: "abcd1234efgh", want: "abcd-1234"},
		{name: "exactly eight characters", state: "abcd1234", want: "abcd-1234"},
		{name: "short state has no code", state: "abc", want: ""},
		{name: "empty state has no code", state: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := confirmationCodeFromState(tt.state); got != tt.want {
				t.Fatalf("confirmationCodeFromState(%q) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestCodeChallengeMatchesRFC7636(t *testing.T) {
	// Test vector from RFC 7636 appendix B.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	if got := codeChallenge(verifier); got != want {
		t.Fatalf("codeChallenge() = %q, want %q", got, want)
	}
}

func TestParsePasteCodeReturnsCode(t *testing.T) {
	state := "state-value"
	pasted := buildPasteCode(t, pastePayload{Version: pasteVersion, Code: "auth-code", State: state})

	got, err := parsePasteCode("  "+pasted+"\n", state)
	if err != nil {
		t.Fatalf("parsePasteCode() error = %v", err)
	}
	if got != "auth-code" {
		t.Fatalf("parsePasteCode() = %q, want %q", got, "auth-code")
	}
}

func TestParsePasteCodeRejectsBadInput(t *testing.T) {
	state := "state-value"

	tests := []struct {
		name   string
		pasted string
		want   string
	}{
		{
			name:   "empty input",
			pasted: "   ",
			want:   "no code was entered",
		},
		{
			name:   "missing prefix",
			pasted: "not-a-sailpoint-code",
			want:   "did not come from the SailPoint sign-in page",
		},
		{
			name:   "damaged payload",
			pasted: pastePrefix + "!!!not-base64!!!",
			want:   "the code is damaged",
		},
		{
			name:   "unknown version",
			pasted: buildPasteCode(t, pastePayload{Version: 99, Code: "auth-code", State: state}),
			want:   "update the CLI",
		},
		{
			name:   "missing code",
			pasted: buildPasteCode(t, pastePayload{Version: pasteVersion, State: state}),
			want:   "missing the authorization code",
		},
		{
			name:   "state from another attempt",
			pasted: buildPasteCode(t, pastePayload{Version: pasteVersion, Code: "auth-code", State: "other-state"}),
			want:   "different sign-in attempt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePasteCode(tt.pasted, state)
			if err == nil {
				t.Fatalf("parsePasteCode() = %q, want an error", got)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("parsePasteCode() error = %q, want it to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestAssertHTTPSURLAcceptsAnyHost(t *testing.T) {
	// A tenant can use a vanity domain, so the host is not restricted.
	accepted := []string{
		"https://acme.api.identitynow.com",
		"https://acme.api.cloud.sailpoint.com",
		"https://acme.login.sailpoint.com/oauth/authorize",
		"https://iga.acme.com",
		"https://identity.acme.co.uk/oauth/authorize",
	}

	for _, rawURL := range accepted {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := assertHTTPSURL(rawURL, "test URL"); err != nil {
				t.Fatalf("assertHTTPSURL(%q) error = %v", rawURL, err)
			}
		})
	}
}

func TestAssertHTTPSURLRejectsUnsafeURLs(t *testing.T) {
	rejected := []string{
		"http://acme.api.identitynow.com",
		"https://user:pass@acme.api.identitynow.com",
		"https://acme.api.identitynow.com/oauth/token#fragment",
		"https://",
		"not a url at all",
	}

	for _, rawURL := range rejected {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := assertHTTPSURL(rawURL, "test URL"); err == nil {
				t.Fatalf("assertHTTPSURL(%q) = nil error, want an error", rawURL)
			}
		})
	}
}

func TestRedirectURIDefaultsToPortalPage(t *testing.T) {
	t.Setenv(redirectURLEnvVar, "")

	got, err := redirectURI()
	if err != nil {
		t.Fatalf("redirectURI() error = %v", err)
	}
	if got != RedirectURI {
		t.Fatalf("redirectURI() = %q, want %q", got, RedirectURI)
	}
}

func TestRedirectURIAcceptsLoopbackOverride(t *testing.T) {
	for _, override := range []string{
		"http://localhost:4200/sailapps",
		"http://127.0.0.1:4200/sailapps",
	} {
		t.Run(override, func(t *testing.T) {
			t.Setenv(redirectURLEnvVar, override)

			got, err := redirectURI()
			if err != nil {
				t.Fatalf("redirectURI() error = %v", err)
			}
			if got != override {
				t.Fatalf("redirectURI() = %q, want %q", got, override)
			}
		})
	}
}

func TestRedirectURIRejectsNonLoopbackOverride(t *testing.T) {
	for _, override := range []string{
		"https://attacker.example/sailapps",
		"http://attacker.example/sailapps",
		"https://localhost:4200/sailapps",
		"http://localhost.attacker.example/sailapps",
	} {
		t.Run(override, func(t *testing.T) {
			t.Setenv(redirectURLEnvVar, override)

			if got, err := redirectURI(); err == nil {
				t.Fatalf("redirectURI() = %q, want an error", got)
			}
		})
	}
}

func TestRequireInteractiveTerminalFailsWithoutTerminal(t *testing.T) {
	// Test runs redirect standard input, so no terminal is attached.
	err := requireInteractiveTerminal()
	if err == nil {
		t.Fatal("requireInteractiveTerminal() = nil error, want an error")
	}
	if !strings.Contains(err.Error(), "interactive terminal") {
		t.Fatalf("requireInteractiveTerminal() error = %q, want it to mention an interactive terminal", err.Error())
	}
}

func TestPromptForPasteCodeTrimsInput(t *testing.T) {
	got, err := promptForPasteCode(strings.NewReader("  sp1.abc  \n"))
	if err != nil {
		t.Fatalf("promptForPasteCode() error = %v", err)
	}
	if got != "sp1.abc" {
		t.Fatalf("promptForPasteCode() = %q, want %q", got, "sp1.abc")
	}
}
