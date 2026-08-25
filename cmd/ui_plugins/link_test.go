package ui_plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
)

const linkResolvedBody = `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`

func intPtr(i int) *int { return &i }

// withTenantURL points config at a throwaway active environment with the given
// tenant URL for the duration of the test, restoring the previous active
// environment afterward. An empty url exercises the unconfigured-tenant path.
func withTenantURL(t *testing.T, url string) {
	t.Helper()
	prev := config.GetActiveEnvironment()
	config.SetActiveEnvironment("clitest")
	config.SetTenantUrl(url)
	t.Cleanup(func() {
		config.SetTenantUrl("")
		config.SetActiveEnvironment(prev)
	})
}

func TestResolveLinkPort(t *testing.T) {
	tests := []struct {
		name          string
		flagPort      int
		portSet       bool
		cfg           *uiPluginWorkspaceConfig
		wantPort      int
		wantDefaulted bool
		wantErr       bool
	}{
		{
			name:     "flag port takes precedence over build.port",
			flagPort: 4300,
			portSet:  true,
			cfg:      &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{Port: intPtr(8080)}},
			wantPort: 4300,
		},
		{
			name:     "invalid flag port errors",
			flagPort: 0,
			portSet:  true,
			cfg:      &uiPluginWorkspaceConfig{},
			wantErr:  true,
		},
		{
			name:          "nil build defaults",
			portSet:       false,
			cfg:           &uiPluginWorkspaceConfig{},
			wantPort:      defaultDevServerPort,
			wantDefaulted: true,
		},
		{
			name:          "nil build.port defaults",
			portSet:       false,
			cfg:           &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "dist"}},
			wantPort:      defaultDevServerPort,
			wantDefaulted: true,
		},
		{
			name:     "build.port used when set",
			portSet:  false,
			cfg:      &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{Port: intPtr(8080)}},
			wantPort: 8080,
		},
		{
			name:    "invalid build.port errors",
			portSet: false,
			cfg:     &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{Port: intPtr(0)}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, defaulted, err := resolveLinkPort(tt.flagPort, tt.portSet, tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got port=%d defaulted=%v", port, defaulted)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if port != tt.wantPort {
				t.Fatalf("port = %d, want %d", port, tt.wantPort)
			}
			if defaulted != tt.wantDefaulted {
				t.Fatalf("defaulted = %v, want %v", defaulted, tt.wantDefaulted)
			}
		})
	}
}

func TestLink_SuccessWithFlagPort(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   linkResolvedBody,
		status:    http.StatusOK,
		body:      `{}`,
	}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 4300, true, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fc.getCalls != 1 {
		t.Fatalf("expected exactly one Get (resolve-alias), got %d", fc.getCalls)
	}
	if fc.postCalls != 1 {
		t.Fatalf("expected exactly one Post (link), got %d", fc.postCalls)
	}
	wantURL := pluginInstancesEndpoint + "/pi-123/link"
	if fc.gotURL != wantURL {
		t.Fatalf("POST url = %s, want %s", fc.gotURL, wantURL)
	}
	if fc.gotPostHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on link POST, got: %v", experimentalHeader, fc.gotPostHeaders)
	}
	if got := strings.TrimSpace(string(fc.gotBody)); got != `{"port":4300}` {
		t.Fatalf("link request body = %s, want {\"port\":4300}", got)
	}

	// stdout must carry only the developer URL so terminal link detection and
	// piping work without truncation.
	wantDevURL := "https://tenant.example.com/ui/plugin/pi-123?spPluginDev=access-request-plugin"
	if got := strings.TrimSpace(out.String()); got != wantDevURL {
		t.Fatalf("stdout = %q, want exactly %q", got, wantDevURL)
	}
	if strings.Contains(out.String(), "linked to port") {
		t.Fatalf("stdout must not contain human diagnostics, got: %s", out.String())
	}
	if !strings.Contains(errOut.String(), "linked to port 4300") {
		t.Fatalf("errOut should contain the human confirmation, got: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "defaulting to port") {
		t.Fatalf("did not expect a defaulting notice when --port is set, got: %s", errOut.String())
	}
}

func TestLink_AppliesDevDocumentHeaders(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, manifestFileName)
	writeManifestAtPath(t, manifestPath, testManifestJSON)
	writeAngularFixture(t, dir, "access-request-plugin", true)

	linkBody := `{"devOverrides":{"devUrl":"https://localhost:4300"},"devDocumentHeaders":{"Content-Security-Policy":"` + sampleCSP + `"}}`
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   linkResolvedBody,
		status:    http.StatusOK,
		body:      linkBody,
	}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, manifestPath, 4300, true, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(errOut.String(), "Restart ng serve") {
		t.Fatalf("expected restart note, got: %s", errOut.String())
	}

	raw, err := os.ReadFile(filepath.Join(dir, angularManifestFileName))
	if err != nil {
		t.Fatalf("read angular.json: %v", err)
	}
	if !strings.Contains(string(raw), sampleCSP) {
		t.Fatalf("angular.json missing updated CSP: %s", raw)
	}
	if strings.Contains(string(raw), "old-csp") {
		t.Fatalf("angular.json should replace prior CSP: %s", raw)
	}
}

func TestLink_PrintsDeveloperURLWhenDevDocumentHeadersPatchFails(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, manifestFileName)
	writeManifestAtPath(t, manifestPath, testManifestJSON)
	writeMultiProjectAngularFixture(t, dir)

	linkBody := `{"devOverrides":{"devUrl":"https://localhost:4300"},"devDocumentHeaders":{"Content-Security-Policy":"` + sampleCSP + `"}}`
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   linkResolvedBody,
		status:    http.StatusOK,
		body:      linkBody,
	}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, manifestPath, 4300, true, &out, &errOut)
	if err != nil {
		t.Fatalf("link should succeed when header patching fails, got: %v", err)
	}

	wantDevURL := "https://tenant.example.com/ui/plugin/pi-123?spPluginDev=access-request-plugin"
	if got := strings.TrimSpace(out.String()); got != wantDevURL {
		t.Fatalf("stdout = %q, want exactly %q", got, wantDevURL)
	}
	if !strings.Contains(errOut.String(), "Warning: could not update angular.json dev server headers") {
		t.Fatalf("expected header patch warning on stderr, got: %s", errOut.String())
	}
}

func TestLink_UsesBuildPort(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{getStatus: http.StatusOK, getBody: linkResolvedBody, status: http.StatusOK, body: `{}`}
	var out, errOut bytes.Buffer

	// manifestWithBuildJSON declares build.port = 8080 and --port is not set.
	err := link(context.Background(), fc, tempManifestPath(t, manifestWithBuildJSON), 0, false, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.TrimSpace(string(fc.gotBody)); got != `{"port":8080}` {
		t.Fatalf("link request body = %s, want {\"port\":8080}", got)
	}
	if strings.Contains(errOut.String(), "defaulting to port") {
		t.Fatalf("did not expect a defaulting notice when build.port is set, got: %s", errOut.String())
	}
}

func TestLink_DefaultsPortWhenUnset(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{getStatus: http.StatusOK, getBody: linkResolvedBody, status: http.StatusOK, body: `{}`}
	var out, errOut bytes.Buffer

	// testManifestJSON has no build section, so neither --port nor build.port is present.
	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 0, false, &out, &errOut)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantBody := fmt.Sprintf(`{"port":%d}`, defaultDevServerPort)
	if got := strings.TrimSpace(string(fc.gotBody)); got != wantBody {
		t.Fatalf("link request body = %s, want %s", got, wantBody)
	}
	if !strings.Contains(errOut.String(), "defaulting to port") {
		t.Fatalf("expected a defaulting notice on stderr, got: %s", errOut.String())
	}
}

func TestLink_EmptyTenantURLFailsFast(t *testing.T) {
	withTenantURL(t, "")
	fc := &fakeClient{}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 0, false, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "no tenant URL configured") {
		t.Fatalf("expected a tenant URL error, got: %v", err)
	}
	if fc.getCalls != 0 || fc.postCalls != 0 {
		t.Fatalf("expected no client calls before the tenant URL check, got get=%d post=%d", fc.getCalls, fc.postCalls)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no stdout, got: %s", out.String())
	}
}

func TestLink_InvalidPortFailsBeforeClientCalls(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), -1, true, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "provided port is invalid") {
		t.Fatalf("expected an invalid port error, got: %v", err)
	}
	if fc.getCalls != 0 || fc.postCalls != 0 {
		t.Fatalf("expected no client calls on an invalid port, got get=%d post=%d", fc.getCalls, fc.postCalls)
	}
}

func TestLink_AliasResolutionFailureSkipsPost(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{getStatus: http.StatusNotFound, getBody: `{"message":"not found"}`}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 4300, true, &out, &errOut)
	if err == nil {
		t.Fatal("expected an error when alias resolution fails")
	}
	if fc.postCalls != 0 {
		t.Fatalf("expected no link POST after alias resolution failure, got %d", fc.postCalls)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no stdout on failure, got: %s", out.String())
	}
}

func TestLink_LinkErrorSurfaced(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   linkResolvedBody,
		status:    http.StatusForbidden,
		body:      `{"message":"Forbidden"}`,
	}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 4300, true, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "unable to create link") {
		t.Fatalf("expected a link error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Forbidden") {
		t.Fatalf("expected the backend message surfaced in the error, got: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no stdout on link failure, got: %s", out.String())
	}
}

func TestLink_TransportError(t *testing.T) {
	withTenantURL(t, "https://tenant.example.com")
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   linkResolvedBody,
		postErr:   errors.New("connection refused"),
	}
	var out, errOut bytes.Buffer

	err := link(context.Background(), fc, tempManifestPath(t, testManifestJSON), 4300, true, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "failed to create link") {
		t.Fatalf("expected a wrapped transport error, got: %v", err)
	}
}
