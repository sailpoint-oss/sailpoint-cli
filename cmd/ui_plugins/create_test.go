package ui_plugins

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveRestrictToUsers(t *testing.T) {
	currentUser := func() (string, error) { return "current-user-guid", nil }

	tests := []struct {
		name      string
		private   bool
		flagUsers []string
		want      []string
	}{
		{name: "no flags", want: nil},
		{name: "flag users only", flagUsers: []string{"user-a", "user-b"}, want: []string{"user-a", "user-b"}},
		{name: "private only", private: true, want: []string{"current-user-guid"}},
		{
			name:      "union of both",
			private:   true,
			flagUsers: []string{"user-a", "user-b"},
			want:      []string{"user-a", "user-b", "current-user-guid"},
		},
		{
			name:      "dedup current user already in list",
			private:   true,
			flagUsers: []string{"current-user-guid", "user-b"},
			want:      []string{"current-user-guid", "user-b"},
		},
		{
			name:      "dedup and trim flag users",
			flagUsers: []string{" user-a ", "user-a", ""},
			want:      []string{"user-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRestrictToUsers(tt.private, tt.flagUsers, currentUser)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !equalStrings(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestResolveRestrictToUsers_CurrentUserError(t *testing.T) {
	wantErr := errors.New("no user context")
	currentUser := func() (string, error) { return "", wantErr }

	_, err := resolveRestrictToUsers(true, []string{"user-a"}, currentUser)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected current-user error to propagate, got: %v", err)
	}
}

func TestResolveRestrictToUsers_CurrentUserNotCalledWithoutPrivate(t *testing.T) {
	called := false
	currentUser := func() (string, error) { called = true; return "x", nil }

	if _, err := resolveRestrictToUsers(false, []string{"user-a"}, currentUser); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("current user resolver should not be called when --private is unset")
	}
}

func TestApplyVisibilityOverride(t *testing.T) {
	manifest := uiPluginManifest{
		Slots: []uiPluginSlot{
			{SlotID: "full-page"},
			{SlotID: "side-panel", RestrictToUsers: []string{"existing"}},
		},
	}

	users := []string{"user-a", "user-b"}
	applyVisibilityOverride(&manifest, users)

	for i, slot := range manifest.Slots {
		if !equalStrings(slot.RestrictToUsers, users) {
			t.Fatalf("slot %d: expected %v, got %v", i, users, slot.RestrictToUsers)
		}
	}

	// Each slot must own an independent copy.
	manifest.Slots[0].RestrictToUsers[0] = "mutated"
	if manifest.Slots[1].RestrictToUsers[0] == "mutated" {
		t.Fatal("slots share the same restrictToUsers backing array")
	}
}

func TestMapUMSCreateError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "bad request", status: 400, body: `{"message":"alias must be 3-63 chars"}`, want: "invalid request"},
		{name: "forbidden", status: 403, body: `{"message":"Forbidden"}`, want: "not authorized"},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`, want: "not enabled"},
		{name: "conflict", status: 409, body: `{"message":"alias 'my-plugin' already exists for this tenant"}`, want: "already exists"},
		{name: "server error", status: 500, body: `{"message":"boom"}`, want: "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapUMSCreateError(tt.status, []byte(tt.body), "my-plugin")
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
		})
	}
}

func TestUMSErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "string message", body: `{"statusCode":400,"message":"bad alias","error":"Bad Request"}`, want: "bad alias"},
		{name: "array message", body: `{"message":["alias too short","name required"]}`, want: "alias too short; name required"},
		{name: "error fallback", body: `{"error":"Conflict"}`, want: "Conflict"},
		{name: "plain body", body: `something failed`, want: "something failed"},
		{name: "empty body", body: ``, want: "no response body"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := umsErrorMessage([]byte(tt.body)); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestRenderCreateSuccess(t *testing.T) {
	t.Run("with id and alias", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderCreateSuccess(&out, []byte(`{"pluginInstanceId":"abc123","alias":"my-plugin"}`), "fallback")
		got := out.String()
		if !strings.Contains(got, "abc123") || !strings.Contains(got, "my-plugin") {
			t.Fatalf("expected id and alias in output, got: %s", got)
		}
	})

	t.Run("falls back to manifest alias", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderCreateSuccess(&out, []byte(`{}`), "fallback-alias")
		if !strings.Contains(out.String(), "fallback-alias") {
			t.Fatalf("expected fallback alias, got: %s", out.String())
		}
	})
}

func TestCreateCommand_DryRunPrintsPayloadWithoutBuildSection(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  },
  "build": {"outDir": "dist", "port": 8080}
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"create", "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected dry-run to succeed, got: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"alias": "access-request-plugin"`) {
		t.Fatalf("expected manifest alias in payload, got: %s", output)
	}
	if strings.Contains(output, "outDir") || strings.Contains(output, "8080") {
		t.Fatalf("local build section must not be in the payload, got: %s", output)
	}
}

func TestCreateCommand_DryRunAppliesRestrictToUsers(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}, {"slotId": "side-panel"}]
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"create", "--dry-run", "--restrict-to-users", "user-a,user-b"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected dry-run to succeed, got: %v", err)
	}

	output := out.String()
	if strings.Count(output, "restrictToUsers") != 2 {
		t.Fatalf("expected restrictToUsers on both slots, got: %s", output)
	}
	if !strings.Contains(output, "user-a") || !strings.Contains(output, "user-b") {
		t.Fatalf("expected override users in payload, got: %s", output)
	}
}

func TestCreateCommand_InvalidManifestFailsBeforeNetwork(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{"create"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to fail on an invalid manifest")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error before any network call, got: %v", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
