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
)

// resolveOKBody is the resolve-alias response for the workspace alias used by the
// shared testManifestJSON fixture (alias: access-request-plugin).
const resolveOKBody = `{"pluginInstanceId":"pi-123","alias":"access-request-plugin","slots":[{"slotId":"full-page"}]}`

// --- update path (resolve GET -> PATCH) via sequenced client ---

func TestRunUpdate_SuccessHumanOutput(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
		patchResp: stubResp{status: http.StatusOK, body: `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`},
	}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.patchCalls != 1 {
		t.Fatalf("expected exactly one Patch, got %d", c.patchCalls)
	}
	if !strings.HasSuffix(c.patchURL, "/pi-123") {
		t.Fatalf("expected PATCH on resolved id, got: %s", c.patchURL)
	}
	if !strings.Contains(string(c.patchBody), "access-request-plugin") {
		t.Fatalf("expected manifest alias in request body, got: %s", c.patchBody)
	}
	if c.patchHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on update, got: %v", experimentalHeader, c.patchHeaders)
	}
	if c.patchHeaders["Content-Type"] != "application/json" {
		t.Fatalf("expected Content-Type application/json on update so UMS parses the body, got: %v", c.patchHeaders)
	}
	got := out.String()
	if !strings.Contains(got, "Updated plugin instance pi-123") || !strings.Contains(got, "access-request-plugin") {
		t.Fatalf("expected success summary with id and alias, got: %s", got)
	}
}

func TestRunUpdate_SendsFullManifestNotBuildSection(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
		patchResp: stubResp{status: http.StatusOK, body: `{}`},
	}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, manifestWithBuildJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(c.patchBody)
	if !strings.Contains(body, "access-request-plugin") {
		t.Fatalf("expected manifest in PATCH body, got: %s", body)
	}
	if strings.Contains(body, "outDir") || strings.Contains(body, "8080") {
		t.Fatalf("local build section must never be sent, got: %s", body)
	}
}

func TestRunUpdate_JSONPassthrough(t *testing.T) {
	respBody := `{"pluginInstanceId":"pi-123","alias":"access-request-plugin","slots":[]}`
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
		patchResp: stubResp{status: http.StatusOK, body: respBody},
	}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{jsonOutput: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out.String()) != respBody {
		t.Fatalf("expected raw response body passthrough, got: %s", out.String())
	}
	if strings.Contains(out.String(), "Updated plugin instance") {
		t.Fatalf("--json output should not include the human summary, got: %s", out.String())
	}
}

func TestRunUpdate_AppliesVisibilityOverrides(t *testing.T) {
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
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
		patchResp: stubResp{status: http.StatusOK, body: `{}`},
	}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, manifest), &out, io.Discard, stubCurrentUser,
		updateConfig{private: true, restrictToUsers: []string{"user-a"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(c.patchBody)
	if strings.Count(body, "restrictToUsers") != 2 {
		t.Fatalf("expected restrictToUsers on both slots, got: %s", body)
	}
	if !strings.Contains(body, "user-a") || !strings.Contains(body, "current-user-guid") {
		t.Fatalf("expected union of override users in payload, got: %s", body)
	}
}

func TestRunUpdate_PatchErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "bad request surfaces backend message", status: 400, body: `{"message":"contentSecurityPolicies has an invalid directive"}`, want: "invalid directive"},
		{name: "forbidden", status: 403, body: `{"message":"Forbidden"}`, want: "not authorized to update"},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`, want: "not found"},
		{name: "server error", status: 500, body: `{"message":"boom"}`, want: "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &seqClient{
				getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
				patchResp: stubResp{status: tt.status, body: tt.body},
			}
			var out bytes.Buffer

			err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
			if c.patchCalls != 1 {
				t.Fatalf("expected the PATCH to be sent, got %d Patch calls", c.patchCalls)
			}
			if out.Len() != 0 {
				t.Fatalf("expected no stdout on error, got: %s", out.String())
			}
		})
	}
}

func TestRunUpdate_TransportError(t *testing.T) {
	c := &seqClient{
		getQueue:  []stubResp{{status: http.StatusOK, body: resolveOKBody}},
		patchResp: stubResp{err: errors.New("connection refused")},
	}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err == nil || !strings.Contains(err.Error(), "failed to update plugin instance") {
		t.Fatalf("expected wrapped transport error, got: %v", err)
	}
}

func TestRunUpdate_ResolveNotFound(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: http.StatusNotFound, body: `{"message":"Not Found"}`}}}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("must not PATCH when the alias does not resolve")
	}
}

func TestRunUpdate_ResolveAmbiguous(t *testing.T) {
	body := `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: http.StatusConflict, body: body}}}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err == nil {
		t.Fatal("expected an ambiguity error")
	}
	if !strings.Contains(err.Error(), "pi-1") || !strings.Contains(err.Error(), "pi-2") {
		t.Fatalf("expected conflicting IDs in error, got: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("ambiguous alias must never be updated")
	}
}

func TestRunUpdate_InvalidManifestSkipsCalls(t *testing.T) {
	invalid := `{
  "version": 1,
  "manifest": {
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`
	c := &seqClient{}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, invalid), &out, io.Discard, stubCurrentUser, updateConfig{})
	if err == nil || !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
	if len(c.getURLs) != 0 || c.patchCalls != 0 {
		t.Fatal("invalid manifest must fail before any backend call")
	}
}

func TestRunUpdate_PrivateResolverErrorSkipsCalls(t *testing.T) {
	wantErr := errors.New("no user context")
	c := &seqClient{}
	var out bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, io.Discard,
		func() (string, error) { return "", wantErr }, updateConfig{private: true})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected current-user error to propagate, got: %v", err)
	}
	if len(c.getURLs) != 0 || c.patchCalls != 0 {
		t.Fatal("resolver failure must fail before any backend call")
	}
}

// --- dry-run path: payload preview + best-effort existence check ---

func TestRunUpdate_DryRunPrintsPayloadAndChecksExistence(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: http.StatusOK, body: resolveOKBody}}}
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, manifestWithBuildJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.patchCalls != 0 {
		t.Fatal("dry run must not PATCH")
	}
	if len(c.getURLs) != 1 {
		t.Fatalf("expected one resolve GET, got %d", len(c.getURLs))
	}
	if !strings.Contains(c.getURLs[0], resolveAliasEndpoint) || !strings.Contains(c.getURLs[0], "alias=access-request-plugin") {
		t.Fatalf("unexpected resolve-alias URL: %s", c.getURLs[0])
	}
	if c.getHeaders[0][experimentalHeader] != "true" {
		t.Fatalf("expected %s header on existence check, got: %v", experimentalHeader, c.getHeaders[0])
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("expected payload on stdout, got: %s", out.String())
	}
	if strings.Contains(out.String(), "outDir") || strings.Contains(out.String(), "8080") {
		t.Fatalf("local build section must not be in the payload, got: %s", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("expected no advisory output when the instance exists, got: %s", errOut.String())
	}
}

func TestRunUpdate_DryRunNotFound(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{status: http.StatusNotFound, body: `{"message":"Not Found"}`}}}
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err == nil || !strings.Contains(err.Error(), "run `sail ui-plugins create` first") {
		t.Fatalf("expected not-found dry-run error, got: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print before the existence error, got: %s", out.String())
	}
	if c.patchCalls != 0 {
		t.Fatal("dry run must not PATCH")
	}
}

func TestRunUpdate_DryRunAmbiguous(t *testing.T) {
	body := `{"conflicts":[{"pluginInstanceId":"pi-1"},{"pluginInstanceId":"pi-2"}]}`
	c := &seqClient{getQueue: []stubResp{{status: http.StatusConflict, body: body}}}
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err == nil || !strings.Contains(err.Error(), "pi-1") {
		t.Fatalf("expected ambiguity error listing IDs, got: %v", err)
	}
}

func TestRunUpdate_DryRunInconclusiveIsNonFatal(t *testing.T) {
	// 403 (private route / external token) is inconclusive; treat as advisory.
	c := &seqClient{getQueue: []stubResp{{status: http.StatusForbidden, body: `{"message":"private route cannot be accessed using external token"}`}}}
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err != nil {
		t.Fatalf("inconclusive check must be non-fatal, got: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "Could not verify the plugin instance exists") {
		t.Fatalf("expected advisory note on stderr, got: %s", errOut.String())
	}
}

func TestRunUpdate_DryRunTransportErrorIsNonFatal(t *testing.T) {
	c := &seqClient{getQueue: []stubResp{{err: errors.New("connection refused")}}}
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), c, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err != nil {
		t.Fatalf("transport error during existence check must be non-fatal, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "Could not verify the plugin instance exists") {
		t.Fatalf("expected advisory note on stderr, got: %s", errOut.String())
	}
}

func TestRunUpdate_DryRunNoClientSkipsCheck(t *testing.T) {
	var out, errOut bytes.Buffer

	err := runUpdate(context.Background(), nil, tempManifestPath(t, testManifestJSON), &out, &errOut, stubCurrentUser, updateConfig{dryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), `"alias": "access-request-plugin"`) {
		t.Fatalf("payload should still print without a client, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "Skipped instance existence check") {
		t.Fatalf("expected skip note on stderr, got: %s", errOut.String())
	}
}

// --- pure helpers ---

func TestMapUMSUpdateError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "bad request", status: 400, body: `{"message":"iframeAllow has an invalid directive"}`, want: "invalid request updating"},
		{name: "bad request passthrough", status: 400, body: `{"message":"iframeAllow has an invalid directive"}`, want: "invalid directive"},
		{name: "forbidden", status: 403, body: `{"message":"Forbidden"}`, want: "not authorized to update"},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`, want: "not found"},
		{name: "conflict", status: 409, body: `{"message":"conflict"}`, want: "alias conflict"},
		{name: "server error", status: 500, body: `{"message":"boom"}`, want: "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapUMSUpdateError(tt.status, []byte(tt.body), "my-plugin")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
		})
	}
}

func TestRenderUpdateSuccess(t *testing.T) {
	resolved := &pluginInstance{PluginInstanceID: "pi-resolved", Alias: "resolved-alias"}

	t.Run("prefers response id and alias", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderUpdateSuccess(&out, []byte(`{"pluginInstanceId":"pi-999","alias":"resp-alias"}`), resolved)
		got := out.String()
		if !strings.Contains(got, "pi-999") || !strings.Contains(got, "resp-alias") {
			t.Fatalf("expected response id and alias, got: %s", got)
		}
	})

	t.Run("falls back to resolved instance", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderUpdateSuccess(&out, []byte(`{}`), resolved)
		got := out.String()
		if !strings.Contains(got, "pi-resolved") || !strings.Contains(got, "resolved-alias") {
			t.Fatalf("expected fallback to resolved instance, got: %s", got)
		}
	})
}

// --- command wiring (cobra + experimental gate), hermetic: fails at validation
// before any backend call, so no client/network is exercised ---

func TestUpdateCommand_InvalidManifestSurfacesValidationError(t *testing.T) {
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
	cmd.SetArgs([]string{"update", "--dry-run"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected update to fail on an invalid manifest")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected validation error, got: %v", err)
	}
}
