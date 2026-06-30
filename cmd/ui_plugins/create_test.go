package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
)

const testManifestJSON = `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`

// fakeClient is a test double for client.Client that returns a canned response
// from Post and records the request it received. Other methods are unused.
type fakeClient struct {
	status    int
	body      string
	postErr   error
	postCalls int
	gotURL    string
	gotBody   []byte
}

var _ client.Client = (*fakeClient)(nil)

func (f *fakeClient) Post(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	f.postCalls++
	f.gotURL = url
	if body != nil {
		f.gotBody, _ = io.ReadAll(body)
	}
	if f.postErr != nil {
		return nil, f.postErr
	}
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(strings.NewReader(f.body)),
	}, nil
}

func (f *fakeClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}

func (f *fakeClient) Delete(ctx context.Context, url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}

func (f *fakeClient) Put(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}

func (f *fakeClient) Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	return nil, errors.New("unused")
}

func tempManifestPath(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), manifestFileName)
	writeManifestAtPath(t, path, content)
	return path
}

func stubCurrentUser() (string, error) { return "current-user-guid", nil }

// --- runCreate: end-to-end HTTP path via fake client ---

func TestRunCreate_SuccessHumanOutput(t *testing.T) {
	fc := &fakeClient{status: http.StatusCreated, body: `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser, createConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fc.postCalls != 1 {
		t.Fatalf("expected exactly one Post, got %d", fc.postCalls)
	}
	if fc.gotURL != pluginInstancesEndpoint {
		t.Fatalf("expected POST to %s, got %s", pluginInstancesEndpoint, fc.gotURL)
	}
	if !strings.Contains(string(fc.gotBody), "access-request-plugin") {
		t.Fatalf("expected manifest alias in request body, got: %s", fc.gotBody)
	}
	got := out.String()
	if !strings.Contains(got, "pi-123") || !strings.Contains(got, "access-request-plugin") {
		t.Fatalf("expected success summary with id and alias, got: %s", got)
	}
}

func TestRunCreate_JSONPassthrough(t *testing.T) {
	respBody := `{"pluginInstanceId":"pi-123","alias":"access-request-plugin","slots":[]}`
	fc := &fakeClient{status: http.StatusCreated, body: respBody}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser, createConfig{jsonOutput: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(out.String()) != respBody {
		t.Fatalf("expected raw response body passthrough, got: %s", out.String())
	}
	if strings.Contains(out.String(), "Created plugin instance") {
		t.Fatalf("--json output should not include the human summary, got: %s", out.String())
	}
}

func TestRunCreate_ErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "bad request", status: 400, body: `{"message":"alias must be 3-63 chars"}`, want: "invalid request"},
		{name: "forbidden", status: 403, body: `{"message":"Forbidden"}`, want: "not authorized"},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`, want: "not enabled"},
		{name: "conflict", status: 409, body: `{"message":"alias 'access-request-plugin' already exists for this tenant"}`, want: "already exists"},
		{name: "server error", status: 500, body: `{"message":"boom"}`, want: "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &fakeClient{status: tt.status, body: tt.body}
			var out bytes.Buffer

			err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser, createConfig{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
			if fc.postCalls != 1 {
				t.Fatalf("expected the request to be sent, got %d Post calls", fc.postCalls)
			}
			if out.Len() != 0 {
				t.Fatalf("expected no stdout on error, got: %s", out.String())
			}
		})
	}
}

func TestRunCreate_TransportError(t *testing.T) {
	fc := &fakeClient{postErr: errors.New("connection refused")}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser, createConfig{})
	if err == nil {
		t.Fatal("expected a transport error")
	}
	if !strings.Contains(err.Error(), "failed to create plugin instance") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}

func TestRunCreate_DryRunSkipsClient(t *testing.T) {
	fc := &fakeClient{status: http.StatusCreated, body: `{}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser, createConfig{dryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fc.postCalls != 0 {
		t.Fatal("dry run must not call the client")
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("expected payload printed on dry run, got: %s", out.String())
	}
}

func TestRunCreate_DryRunAppliesOverrides(t *testing.T) {
	fc := &fakeClient{}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, stubCurrentUser,
		createConfig{dryRun: true, private: true, restrictToUsers: []string{"user-a"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "restrictToUsers") || !strings.Contains(got, "user-a") || !strings.Contains(got, "current-user-guid") {
		t.Fatalf("expected union of override users in payload, got: %s", got)
	}
}

func TestRunCreate_InvalidManifestSkipsPost(t *testing.T) {
	invalid := `{
  "version": 1,
  "manifest": {
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`
	fc := &fakeClient{status: http.StatusCreated, body: `{}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, invalid), &out, stubCurrentUser, createConfig{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
	if fc.postCalls != 0 {
		t.Fatal("invalid manifest must fail before any backend call")
	}
}

func TestRunCreate_PrivateResolverErrorSkipsPost(t *testing.T) {
	wantErr := errors.New("no user context")
	fc := &fakeClient{status: http.StatusCreated, body: `{}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out,
		func() (string, error) { return "", wantErr }, createConfig{private: true})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected current-user error to propagate, got: %v", err)
	}
	if fc.postCalls != 0 {
		t.Fatal("resolver failure must fail before any backend call")
	}
}

// --- pure helpers ---

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

// --- command wiring (cobra + experimental gate + flags), hermetic via dry run ---

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

func TestCreateCommand_InvalidManifestSurfacesValidationError(t *testing.T) {
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
	// --dry-run keeps this hermetic (no config/auth needed); validation runs first regardless.
	cmd.SetArgs([]string{"create", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to fail on an invalid manifest")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
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
