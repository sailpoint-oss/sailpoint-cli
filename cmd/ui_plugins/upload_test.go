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

func TestPluginViewURL(t *testing.T) {
	const id = "2c918084-8bce-1f2a-0181-abc123def456"

	tests := []struct {
		name       string
		tenantURL  string
		instanceID string
		want       string
	}{
		{
			name:       "builds view url",
			tenantURL:  "https://tenant.example.com",
			instanceID: id,
			want:       "https://tenant.example.com/ui/plugin/" + id,
		},
		{
			name:       "trims trailing slash",
			tenantURL:  "https://tenant.example.com/",
			instanceID: id,
			want:       "https://tenant.example.com/ui/plugin/" + id,
		},
		{
			name:       "trims surrounding whitespace",
			tenantURL:  "  https://tenant.example.com  ",
			instanceID: id,
			want:       "https://tenant.example.com/ui/plugin/" + id,
		},
		{
			name:       "preserves existing base path",
			tenantURL:  "https://tenant.example.com/base",
			instanceID: id,
			want:       "https://tenant.example.com/base/ui/plugin/" + id,
		},
		{
			name:       "empty tenant url yields empty string",
			tenantURL:  "",
			instanceID: id,
			want:       "",
		},
		{
			name:       "whitespace-only tenant url yields empty string",
			tenantURL:  "   ",
			instanceID: id,
			want:       "",
		},
		{
			name:       "empty instance id yields empty string",
			tenantURL:  "https://tenant.example.com",
			instanceID: "",
			want:       "",
		},
		{
			name:       "unparseable tenant url yields empty string",
			tenantURL:  "https://exa\x7fmple.com",
			instanceID: id,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pluginViewURL(tt.tenantURL, tt.instanceID); got != tt.want {
				t.Fatalf("pluginViewURL(%q, %q) = %q, want %q", tt.tenantURL, tt.instanceID, got, tt.want)
			}
		})
	}
}

func TestResolveUploadOutDir(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		cfg     *uiPluginWorkspaceConfig
		want    string
		wantErr bool
	}{
		{
			name: "flag takes precedence over build.outDir",
			flag: "custom-dist",
			cfg:  &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "dist"}},
			want: "custom-dist",
		},
		{
			name: "flag is trimmed",
			flag: "  custom-dist  ",
			cfg:  &uiPluginWorkspaceConfig{},
			want: "custom-dist",
		},
		{
			name: "falls back to build.outDir",
			flag: "",
			cfg:  &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "dist"}},
			want: "dist",
		},
		{
			name: "build.outDir is trimmed",
			flag: "",
			cfg:  &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "  dist  "}},
			want: "dist",
		},
		{
			name: "whitespace-only flag falls back to build.outDir",
			flag: "   ",
			cfg:  &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "dist"}},
			want: "dist",
		},
		{
			name:    "nil build without flag errors",
			flag:    "",
			cfg:     &uiPluginWorkspaceConfig{},
			wantErr: true,
		},
		{
			name:    "empty build.outDir without flag errors",
			flag:    "",
			cfg:     &uiPluginWorkspaceConfig{Build: &uiPluginBuildConfig{OutDir: "  "}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveUploadOutDir(tt.flag, tt.cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got result %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolveUploadOutDir(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestRunUpload_Success(t *testing.T) {
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`,
		status:    http.StatusCreated,
		body:      `{"assetBundleId":"ab-1","pluginInstanceId":"pi-123","assets":[{"path":"index.html"}]}`,
	}
	outDir := newOutDir(t, map[string]string{"index.html": "<html></html>"})
	var out bytes.Buffer

	err := runUpload(context.Background(), fc, tempManifestPath(t, testManifestJSON), outDir, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fc.getCalls != 1 {
		t.Fatalf("expected exactly one Get (resolve-alias), got %d", fc.getCalls)
	}
	if fc.postCalls != 1 {
		t.Fatalf("expected exactly one Post (upload), got %d", fc.postCalls)
	}
	wantURL := pluginInstancesEndpoint + "/pi-123/asset-bundles"
	if fc.gotURL != wantURL {
		t.Fatalf("expected POST to %s, got %s", wantURL, fc.gotURL)
	}
	if fc.gotPostHeaders[experimentalHeader] != "true" {
		t.Fatalf("expected %s header on upload, got: %v", experimentalHeader, fc.gotPostHeaders)
	}
	if !strings.Contains(string(fc.gotBody), "index.html") {
		t.Fatalf("expected multipart body to include the asset filename, got: %s", fc.gotBody)
	}
	if got := out.String(); !strings.Contains(got, "ab-1") || !strings.Contains(got, "access-request-plugin") {
		t.Fatalf("expected success summary with bundle id and alias, got: %s", got)
	}
}

func TestRunUpload_InvalidManifestSkipsClient(t *testing.T) {
	fc := &fakeClient{}

	err := runUpload(context.Background(), fc, tempManifestPath(t, `{"version":1}`), "somewhere", io.Discard)
	if err == nil {
		t.Fatal("expected an error for an invalid manifest")
	}
	if fc.getCalls != 0 || fc.postCalls != 0 {
		t.Fatalf("expected no client calls, got get=%d post=%d", fc.getCalls, fc.postCalls)
	}
}

func TestRunUpload_MissingOutDirSkipsClient(t *testing.T) {
	fc := &fakeClient{}
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	err := runUpload(context.Background(), fc, tempManifestPath(t, testManifestJSON), missing, io.Discard)
	if err == nil {
		t.Fatal("expected an error for a missing output directory")
	}
	if fc.getCalls != 0 || fc.postCalls != 0 {
		t.Fatalf("expected no client calls, got get=%d post=%d", fc.getCalls, fc.postCalls)
	}
}

func TestRunUpload_AliasResolutionErrorSkipsUpload(t *testing.T) {
	fc := &fakeClient{getStatus: http.StatusNotFound, getBody: `{"message":"not found"}`}
	outDir := newOutDir(t, map[string]string{"index.html": "x"})

	err := runUpload(context.Background(), fc, tempManifestPath(t, testManifestJSON), outDir, io.Discard)
	if err == nil {
		t.Fatal("expected an error when alias resolution fails")
	}
	if fc.postCalls != 0 {
		t.Fatalf("expected no upload after alias resolution failure, got %d", fc.postCalls)
	}
}

func TestRunUpload_UploadErrorMapped(t *testing.T) {
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`,
		status:    http.StatusBadRequest,
		body:      `{"message":"Bundle must include index.html at the root"}`,
	}
	outDir := newOutDir(t, map[string]string{"index.html": "x"})
	var out bytes.Buffer

	err := runUpload(context.Background(), fc, tempManifestPath(t, testManifestJSON), outDir, &out)
	if err == nil {
		t.Fatal("expected an error for a 400 upload response")
	}
	if !strings.Contains(err.Error(), "invalid asset bundle") {
		t.Fatalf("expected mapped upload error, got: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no success output on error, got: %s", out.String())
	}
}

func TestRunUpload_TransportError(t *testing.T) {
	fc := &fakeClient{
		getStatus: http.StatusOK,
		getBody:   `{"pluginInstanceId":"pi-123","alias":"access-request-plugin"}`,
		postErr:   errors.New("boom"),
	}
	outDir := newOutDir(t, map[string]string{"index.html": "x"})

	err := runUpload(context.Background(), fc, tempManifestPath(t, testManifestJSON), outDir, io.Discard)
	if err == nil {
		t.Fatal("expected a transport error")
	}
	if !strings.Contains(err.Error(), "failed to upload plugin assets") {
		t.Fatalf("expected upload transport error, got: %v", err)
	}
}
