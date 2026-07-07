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
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`

const manifestWithBuildJSON = `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  },
  "build": {"outDir": "dist", "port": 8080}
}`

// fakeClient is a test double for client.Client. Post backs the create path and Get
// backs the dry-run validate-alias check; each records its call and returns a canned
// response. The remaining methods are unused.
type fakeClient struct {
	// Post (create)
	status         int
	body           string
	postErr        error
	postCalls      int
	gotURL         string
	gotBody        []byte
	gotPostHeaders map[string]string

	// Get (validate-alias)
	getStatus     int
	getBody       string
	getErr        error
	getCalls      int
	gotGetURL     string
	gotGetHeaders map[string]string
}

var _ client.Client = (*fakeClient)(nil)

func (f *fakeClient) Post(ctx context.Context, url string, contentType string, body io.Reader, headers map[string]string) (*http.Response, error) {
	f.postCalls++
	f.gotURL = url
	f.gotPostHeaders = headers
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
	f.getCalls++
	f.gotGetURL = url
	f.gotGetHeaders = headers
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &http.Response{
		StatusCode: f.getStatus,
		Body:       io.NopCloser(strings.NewReader(f.getBody)),
	}, nil
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

// --- create path (POST) via fake client ---

func TestRunCreate_SuccessHumanOutput(t *testing.T) {
	fc := &fakeClient{status: http.StatusCreated, body: `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, createConfig{})
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
	if fc.gotPostHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on create, got: %v", experimentalHeader, fc.gotPostHeaders)
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

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, createConfig{jsonOutput: true})
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

			err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, createConfig{})
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

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, createConfig{})
	if err == nil {
		t.Fatal("expected a transport error")
	}
	if !strings.Contains(err.Error(), "failed to create plugin instance") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
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

	err := runCreate(context.Background(), fc, tempManifestPath(t, invalid), &out, io.Discard, stubCurrentUser, createConfig{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
	if fc.postCalls != 0 || fc.getCalls != 0 {
		t.Fatal("invalid manifest must fail before any backend call")
	}
}

func TestRunCreate_PrivateResolverErrorSkipsPost(t *testing.T) {
	wantErr := errors.New("no user context")
	fc := &fakeClient{status: http.StatusCreated, body: `{}`}
	var out bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, io.Discard,
		func() (string, error) { return "", wantErr }, createConfig{private: true})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected current-user error to propagate, got: %v", err)
	}
	if fc.postCalls != 0 {
		t.Fatal("resolver failure must fail before any backend call")
	}
}

// --- dry-run path: payload preview + best-effort alias availability check ---

func TestRunCreate_DryRunPrintsPayloadAndChecksAlias(t *testing.T) {
	fc := &fakeClient{getStatus: http.StatusOK}
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, manifestWithBuildJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fc.postCalls != 0 {
		t.Fatal("dry run must not POST")
	}
	if fc.getCalls != 1 {
		t.Fatalf("expected one alias check GET, got %d", fc.getCalls)
	}
	if !strings.Contains(fc.gotGetURL, validateAliasEndpoint) || !strings.Contains(fc.gotGetURL, "alias=access-request-plugin") {
		t.Fatalf("unexpected validate-alias URL: %s", fc.gotGetURL)
	}
	if fc.gotGetHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on alias check, got: %v", experimentalHeader, fc.gotGetHeaders)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("expected payload on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "outDir") || strings.Contains(out.String(), "8080") {
		t.Fatalf("local build section must not be in the payload, got: %s", out.String())
	}
	// Available => clean stdout (payload only), nothing on stderr.
	if errOut.Len() != 0 {
		t.Fatalf("expected no advisory output when alias is available, got: %s", errOut.String())
	}
}

func TestRunCreate_DryRunAliasTaken(t *testing.T) {
	fc := &fakeClient{getStatus: http.StatusConflict, getBody: `{"message":"alias 'access-request-plugin' already exists for this tenant"}`}
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err == nil {
		t.Fatal("expected a definitive error when the alias is taken")
	}
	if !strings.Contains(err.Error(), "already in use") {
		t.Fatalf("expected alias-taken error, got: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print before the alias error, got: %s", out.String())
	}
	if fc.postCalls != 0 {
		t.Fatal("dry run must not POST")
	}
}

func TestRunCreate_DryRunAliasInvalid(t *testing.T) {
	fc := &fakeClient{getStatus: http.StatusBadRequest, getBody: `{"message":"alias must be 3-63 chars"}`}
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err == nil || !strings.Contains(err.Error(), "is invalid") {
		t.Fatalf("expected alias-invalid error, got: %v", err)
	}
}

func TestRunCreate_DryRunAliasUnverifiableIsNonFatal(t *testing.T) {
	// 401 (private route / external token) is the current real-world case; treat as advisory.
	fc := &fakeClient{getStatus: http.StatusUnauthorized, getBody: `{"message":"private route cannot be accessed using external token"}`}
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err != nil {
		t.Fatalf("inconclusive check must be non-fatal, got: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "Could not verify alias availability") {
		t.Fatalf("expected advisory note on stderr, got: %s", errOut.String())
	}
}

func TestRunCreate_DryRunAliasTransportErrorIsNonFatal(t *testing.T) {
	fc := &fakeClient{getErr: errors.New("connection refused")}
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), fc, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err != nil {
		t.Fatalf("transport error during alias check must be non-fatal, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "Could not verify alias availability") {
		t.Fatalf("expected advisory note on stderr, got: %s", errOut.String())
	}
}

func TestRunCreate_DryRunNoClientSkipsCheck(t *testing.T) {
	var out, errOut bytes.Buffer

	err := runCreate(context.Background(), nil, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, createConfig{dryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print without a client, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "Skipped alias availability check") {
		t.Fatalf("expected skip note on stderr, got: %s", errOut.String())
	}
}

func TestRunCreate_DryRunAppliesOverrides(t *testing.T) {
	fc := &fakeClient{getStatus: http.StatusOK}
	var out, errOut bytes.Buffer

	manifest := `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}, {"slotId": "side-panel"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`
	err := runCreate(context.Background(), fc, tempManifestPath(t, manifest), &out, &errOut, stubCurrentUser,
		createConfig{dryRun: true, private: true, restrictToUsers: []string{"user-a"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if strings.Count(got, "restrictToUsers") != 2 {
		t.Fatalf("expected restrictToUsers on both slots, got: %s", got)
	}
	if !strings.Contains(got, "user-a") || !strings.Contains(got, "current-user-guid") {
		t.Fatalf("expected union of override users in payload, got: %s", got)
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

// --- command wiring (cobra + experimental gate), hermetic: fails at validation
// before the dry-run alias check, so no client/network is exercised ---

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
	cmd.SetArgs([]string{"create", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to fail on an invalid manifest")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

func TestUIPluginRequestHeaders(t *testing.T) {
	h := uiPluginRequestHeaders()
	if h[experimentalHeader] != "true" {
		t.Fatalf("expected %s: true, got: %v", experimentalHeader, h)
	}
	if h["Accept"] != "application/json" {
		t.Fatalf("expected Accept: application/json, got: %v", h)
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
