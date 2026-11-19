package whoami

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

func TestPrintInfo_loggedIn(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	info := config.AuthInfo{
		Environment:        "default",
		EnvironmentSource:  config.SourceDefault,
		AuthType:           "pat",
		AuthTypeSource:     config.SourceConfig,
		BaseURL:            "https://acme.api.identitynow.com",
		BaseURLSource:      config.SourceEnv + " (SAIL_BASE_URL)",
		ClientIDSource:     config.SourceEnv + " (SAIL_CLIENT_ID)",
		ClientSecretSource: config.SourceKeyring,
		LoggedIn:           true,
		Claims: &config.TokenClaims{
			UserName:  "jane.doe",
			Org:       "acme",
			Pod:       "stg01-useast1",
			Scopes:    []string{"sp:scopes:all"},
			ExpiresAt: now.Add(10 * time.Minute),
		},
	}

	var b bytes.Buffer
	printInfo(&b, info, now)
	out := b.String()

	for _, want := range []string{"PAT", "SAIL_BASE_URL", "SAIL_CLIENT_ID", "keychain", "jane.doe", "acme", "sp:scopes:all", "expires in 10m0s"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintInfo_notLoggedIn(t *testing.T) {
	info := config.AuthInfo{AuthType: "oauth", TokenError: config.ErrNotLoggedIn.Error()}

	var b bytes.Buffer
	printInfo(&b, info, time.Now())
	out := b.String()

	if !strings.Contains(out, "not logged in") || !strings.Contains(out, "log in with OAuth") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestPrintInfo_noUserContext(t *testing.T) {
	info := config.AuthInfo{AuthType: "pat", LoggedIn: true, Claims: &config.TokenClaims{ClientID: "abc"}}

	var b bytes.Buffer
	printInfo(&b, info, time.Now())

	if !strings.Contains(b.String(), "without a user context") {
		t.Errorf("unexpected output:\n%s", b.String())
	}
}

func TestPrintInfo_warningsAndPAT(t *testing.T) {
	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	info := config.AuthInfo{
		AuthType: "pat",
		EnvVars:  []config.EnvVar{{Name: "SAIL_CLIENT_ID", Value: "****cdef"}},
		Warnings: []string{"SAIL_CLIENT_ID is overriding the PAT client ID saved for environment \"default\"."},
		LoggedIn: true,
		Claims:   &config.TokenClaims{UserName: "jane.doe"},
		PAT:      &config.PATDetails{Name: "old-ci-token", ID: "****cdef", Created: now.AddDate(-2, 0, 0)},
	}

	var b bytes.Buffer
	printInfo(&b, info, now)
	out := b.String()

	if !strings.HasPrefix(out, "Warnings") {
		t.Errorf("warnings should be printed first:\n%s", out)
	}
	for _, want := range []string{"override the config file and keychain", "SAIL_CLIENT_ID", "old-ci-token", "2 years", "Expires:", "never"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintInfo_encodedScopes(t *testing.T) {
	info := config.AuthInfo{
		AuthType: "pat",
		LoggedIn: true,
		Claims:   &config.TokenClaims{UserName: "jane.doe", Scopes: []string{"Bg=="}, ScopesEncoded: true},
		PAT:      &config.PATDetails{Name: "ci", Scopes: []string{"sp:scopes:all"}, Created: time.Now()},
	}

	var b bytes.Buffer
	printInfo(&b, info, time.Now())
	out := b.String()

	for _, want := range []string{"encoded in the token: Bg==", "see the PAT's scopes below", "sp:scopes:all"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintInfo_multiLineReason(t *testing.T) {
	info := config.AuthInfo{AuthType: "pat", TokenError: "the tenant rejected the PAT\n\nCheck the values of SAIL_CLIENT_ID"}

	var b bytes.Buffer
	printInfo(&b, info, time.Now())

	if !strings.Contains(b.String(), "Reason:  the tenant rejected the PAT Check the values of SAIL_CLIENT_ID") {
		t.Errorf("reason should be on one line:\n%s", b.String())
	}
}
