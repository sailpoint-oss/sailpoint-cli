package ui_plugins

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newOutDir creates a temporary directory populated with the given
// forward-slash-relative files and returns its path.
func newOutDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return dir
}

func TestCollectUploadFiles(t *testing.T) {
	t.Run("missing directory", func(t *testing.T) {
		_, err := collectUploadFiles(filepath.Join(t.TempDir(), "nope"))
		if err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("expected does-not-exist error, got: %v", err)
		}
	})

	t.Run("path is not a directory", func(t *testing.T) {
		dir := newOutDir(t, map[string]string{"index.html": "x"})
		_, err := collectUploadFiles(filepath.Join(dir, "index.html"))
		if err == nil || !strings.Contains(err.Error(), "is not a directory") {
			t.Fatalf("expected not-a-directory error, got: %v", err)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		_, err := collectUploadFiles(t.TempDir())
		if err == nil || !strings.Contains(err.Error(), "no files to upload") {
			t.Fatalf("expected empty-directory error, got: %v", err)
		}
	})

	t.Run("only dotfiles", func(t *testing.T) {
		dir := newOutDir(t, map[string]string{".DS_Store": "junk"})
		_, err := collectUploadFiles(dir)
		if err == nil || !strings.Contains(err.Error(), "no files to upload") {
			t.Fatalf("expected empty-after-dotfile-skip error, got: %v", err)
		}
	})

	t.Run("collects nested files and skips dotfiles", func(t *testing.T) {
		dir := newOutDir(t, map[string]string{
			"index.html":          "<html>",
			"assets/app.js":       "console.log(1)",
			"assets/img/logo.png": "PNG",
			".DS_Store":           "junk",
			"assets/.hidden":      "secret",
		})

		files, err := collectUploadFiles(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got := map[string]bool{}
		for _, f := range files {
			got[f.relPath] = true
		}
		want := []string{"index.html", "assets/app.js", "assets/img/logo.png"}
		if len(got) != len(want) {
			t.Fatalf("expected %d files, got %d: %v", len(want), len(got), got)
		}
		for _, w := range want {
			if !got[w] {
				t.Fatalf("expected %q among collected files, got: %v", w, got)
			}
		}
		if got[".DS_Store"] || got["assets/.hidden"] {
			t.Fatalf("dotfiles should be skipped, got: %v", got)
		}
	})

	t.Run("exceeds size guardrail", func(t *testing.T) {
		dir := t.TempDir()
		f, err := os.Create(filepath.Join(dir, "big.js"))
		if err != nil {
			t.Fatal(err)
		}
		// Sparse file: sets the reported size without writing the bytes.
		if err := f.Truncate(maxBundleBytes + 1); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()

		_, err = collectUploadFiles(dir)
		if err == nil || !strings.Contains(err.Error(), "guardrail") {
			t.Fatalf("expected size-guardrail error, got: %v", err)
		}
	})
}

func TestContentTypeForPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"index.html", "text/html"},
		{"main.js", "text/javascript"},
		{"styles.css", "text/css"},
		{"data.json", "application/json"},
		{"main.js.map", "application/json"},
		{"logo.PNG", "image/png"},
		{"font.woff2", "font/woff2"},
		{"favicon.ico", "image/x-icon"},
		{"archive.zip", "application/octet-stream"},
		{"noext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := contentTypeForPath(tt.path); got != tt.want {
				t.Fatalf("contentTypeForPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestBuildAssetBundleBody(t *testing.T) {
	dir := newOutDir(t, map[string]string{
		"index.html":    "<html></html>",
		"assets/app.js": "console.log(1)",
	})
	files := []uploadFile{
		{relPath: "index.html", absPath: filepath.Join(dir, "index.html")},
		{relPath: "assets/app.js", absPath: filepath.Join(dir, "assets", "app.js")},
	}

	body, contentType, err := buildAssetBundleBody(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Fatalf("unexpected content type: %s", contentType)
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("failed to parse content type: %v", err)
	}

	mr := multipart.NewReader(body, params["boundary"])
	foundIndex, foundApp := false, false
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading part: %v", err)
		}

		disposition := p.Header.Get("Content-Disposition")
		if !strings.Contains(disposition, `name="file"`) {
			t.Fatalf("expected form field name 'file', got disposition %q", disposition)
		}
		data, _ := io.ReadAll(p)

		switch {
		// The subdirectory is preserved on the wire (the backend uses
		// preservePath); Go's Part.FileName would strip it, so assert the raw
		// Content-Disposition instead.
		case strings.Contains(disposition, `filename="index.html"`):
			foundIndex = true
			if ct := p.Header.Get("Content-Type"); ct != "text/html" {
				t.Fatalf("index.html content type = %q, want text/html", ct)
			}
			if string(data) != "<html></html>" {
				t.Fatalf("index.html content = %q", data)
			}
		case strings.Contains(disposition, `filename="assets/app.js"`):
			foundApp = true
			if ct := p.Header.Get("Content-Type"); ct != "text/javascript" {
				t.Fatalf("app.js content type = %q, want text/javascript", ct)
			}
			if string(data) != "console.log(1)" {
				t.Fatalf("app.js content = %q", data)
			}
		default:
			t.Fatalf("unexpected part disposition: %q", disposition)
		}
	}

	if !foundIndex || !foundApp {
		t.Fatalf("missing expected parts: index=%v app=%v", foundIndex, foundApp)
	}
}

func TestBuildAssetBundleBody_ReadError(t *testing.T) {
	files := []uploadFile{{relPath: "missing.js", absPath: filepath.Join(t.TempDir(), "missing.js")}}

	_, _, err := buildAssetBundleBody(files)
	if err == nil || !strings.Contains(err.Error(), "failed to read") {
		t.Fatalf("expected a read error, got: %v", err)
	}
}

func TestRenderUploadSuccess(t *testing.T) {
	t.Run("with bundle id and asset count", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderUploadSuccess(&out, []byte(`{"assetBundleId":"ab-9","assets":[{"path":"index.html"},{"path":"app.js"}]}`), "my-plugin", "")
		got := out.String()
		for _, want := range []string{"ab-9", "my-plugin", "2 asset"} {
			if !strings.Contains(got, want) {
				t.Fatalf("expected %q in output, got: %s", want, got)
			}
		}
	})

	t.Run("without bundle id uses fallback line", func(t *testing.T) {
		var out bytes.Buffer
		_ = renderUploadSuccess(&out, []byte(`{}`), "my-plugin", "")
		got := out.String()
		if !strings.Contains(got, "Uploaded assets to plugin") || !strings.Contains(got, "my-plugin") {
			t.Fatalf("expected fallback summary, got: %s", got)
		}
	})
}

func TestMapUMSUploadError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{name: "bad request", status: 400, body: `{"message":"Bundle must include index.html at the root"}`, want: "invalid asset bundle"},
		{name: "forbidden", status: 403, body: `{"message":"Forbidden"}`, want: "not authorized"},
		{name: "not found", status: 404, body: `{"message":"Not Found"}`, want: "not found"},
		{name: "payload too large", status: 413, body: `{"message":"too big"}`, want: "exceeds the maximum allowed size"},
		{name: "server error", status: 500, body: `{"message":"boom"}`, want: "status 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapUMSUploadError(tt.status, []byte(tt.body), "my-plugin")
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
		})
	}
}
