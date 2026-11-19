package config

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// fakeToken builds a compact JWS with the given claims. The signature is not
// valid, which is fine because ParseTokenClaims does not verify it.
func fakeToken(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	enc := base64.RawURLEncoding
	header := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return header + "." + enc.EncodeToString(body) + "." + enc.EncodeToString([]byte("sig"))
}

func TestParseTokenClaims(t *testing.T) {
	exp := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	token := fakeToken(t, map[string]interface{}{
		"user_name":   "jane.doe",
		"identity_id": "2c91808a",
		"client_id":   "abc123",
		"tenant_id":   "tenant-1",
		"org":         "acme",
		"pod":         "stg01-useast1",
		"scope":       []string{"sp:scopes:default", "sp:scopes:all"},
		"authorities": []string{"ORG_ADMIN"},
		"strong_auth": true,
		"exp":         exp.Unix(),
	})

	got, err := ParseTokenClaims(token)
	if err != nil {
		t.Fatalf("ParseTokenClaims: %v", err)
	}

	if got.UserName != "jane.doe" || got.IdentityID != "2c91808a" || got.Org != "acme" || got.Pod != "stg01-useast1" || got.TenantID != "tenant-1" || got.ClientID != "abc123" {
		t.Errorf("unexpected identity claims: %+v", got)
	}
	if !reflect.DeepEqual(got.Scopes, []string{"sp:scopes:default", "sp:scopes:all"}) {
		t.Errorf("scopes = %v", got.Scopes)
	}
	if !reflect.DeepEqual(got.Authorities, []string{"ORG_ADMIN"}) {
		t.Errorf("authorities = %v", got.Authorities)
	}
	if !got.StrongAuth {
		t.Error("expected strong auth")
	}
	if !got.ExpiresAt.Equal(exp) {
		t.Errorf("expiresAt = %v, want %v", got.ExpiresAt, exp)
	}
	if !got.HasUserContext() {
		t.Error("expected user context")
	}
}

func TestParseTokenClaims_noUserContext(t *testing.T) {
	got, err := ParseTokenClaims(fakeToken(t, map[string]interface{}{"client_id": "abc", "scope": "sp:scopes:all"}))
	if err != nil {
		t.Fatal(err)
	}
	if got.HasUserContext() {
		t.Error("expected no user context")
	}
	if !reflect.DeepEqual(got.Scopes, []string{"sp:scopes:all"}) {
		t.Errorf("scopes = %v", got.Scopes)
	}
}

func TestParseTokenClaims_invalid(t *testing.T) {
	if _, err := ParseTokenClaims("not-a-token"); err == nil {
		t.Error("expected error")
	}
}

func TestNormalizeScopes(t *testing.T) {
	tests := []struct {
		name  string
		scope interface{}
		want  []string
	}{
		{"space delimited string", "a b  c", []string{"a", "b", "c"}},
		{"list", []interface{}{"a", "", "b", 3}, []string{"a", "b"}},
		{"missing", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeScopes(tt.scope); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestViperSource_env(t *testing.T) {
	t.Setenv("SAIL_AUTHTYPE", "oauth")
	if got := viperSource("authtype"); got != SourceEnv+" (SAIL_AUTHTYPE)" {
		t.Errorf("got %q", got)
	}
}

func TestCollectEnvVars(t *testing.T) {
	got := collectEnvVars([]string{
		"PATH=/usr/bin",
		"SAIL_CLIENT_SECRET=supersecret",
		"SAIL_CLIENT_ID=0123456789abcdef",
		"SAIL_BASE_URL=https://acme.api.identitynow.com",
		"SAIL_EMPTY=",
	})
	want := []EnvVar{
		{Name: "SAIL_BASE_URL", Value: "https://acme.api.identitynow.com"},
		{Name: "SAIL_CLIENT_ID", Value: "****cdef"},
		{Name: "SAIL_CLIENT_SECRET", Value: "(set, hidden)"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestDescribeOverrides(t *testing.T) {
	got := describeOverrides("prod", []override{
		{"SAIL_CLIENT_ID", "PAT client ID", "old-id", "new-id"},   // differs: warn
		{"SAIL_BASE_URL", "base URL", "https://a", "https://a"},   // same: no warning
		{"SAIL_CLIENT_SECRET", "PAT client secret", "secret", ""}, // nothing saved: no warning
	})
	if len(got) != 1 || !strings.Contains(got[0], "SAIL_CLIENT_ID") || !strings.Contains(got[0], `"prod"`) {
		t.Errorf("unexpected warnings: %v", got)
	}
	if strings.Contains(got[0], "old-id") || strings.Contains(got[0], "new-id") {
		t.Errorf("warning must not include credential values: %s", got[0])
	}
}

func TestPATWarnings(t *testing.T) {
	now := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	soon := now.Add(7 * 24 * time.Hour)
	past := now.Add(-24 * time.Hour)
	far := now.Add(300 * 24 * time.Hour)

	tests := []struct {
		name string
		pat  PATDetails
		want []string
	}{
		{"recent, no expiry", PATDetails{Name: "ci", Created: now.Add(-30 * 24 * time.Hour)}, nil},
		{"two years old", PATDetails{Name: "ci", Created: now.AddDate(-2, 0, 0)}, []string{"over a year ago"}},
		{"expires soon", PATDetails{Name: "ci", Created: now, ExpiresAt: &soon}, []string{"expires on"}},
		{"expired", PATDetails{Name: "ci", Created: now, ExpiresAt: &past}, []string{"expired on"}},
		{"expires later", PATDetails{Name: "ci", Created: now, ExpiresAt: &far}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := patWarnings(tt.pat, now)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %d warnings", got, len(tt.want))
			}
			for i, w := range tt.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("warning %q missing %q", got[i], w)
				}
			}
		})
	}
}

func TestScopesEncoded(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   bool
	}{
		{"encoded", []string{"Bg=="}, true},
		{"readable", []string{"sp:scopes:all"}, false},
		{"mixed", []string{"Bg==", "sp:scopes:all"}, false},
		{"empty", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scopesEncoded(tt.scopes); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCachedTokenMatches(t *testing.T) {
	token := fakeToken(t, map[string]interface{}{"client_id": "8b456d75aaaa"})

	tests := []struct {
		name     string
		token    string
		clientID string
		want     bool
	}{
		{"same client ID", token, "8b456d75aaaa", true},
		{"case differs", token, "8B456D75AAAA", true},
		{"different client ID", token, "00000000", false},
		{"no configured client ID", token, "", true},
		{"undecodable token", "not-a-token", "00000000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := cachedTokenMatches(tt.token, tt.clientID); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
