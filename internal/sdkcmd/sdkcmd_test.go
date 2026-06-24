package sdkcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadJSONFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.json")
	if err := os.WriteFile(path, []byte(`{"name":"demo"}`), 0600); err != nil {
		t.Fatalf("failed to write payload: %v", err)
	}

	got, err := ReadJSONFile[map[string]string](path)
	if err != nil {
		t.Fatalf("ReadJSONFile returned error: %v", err)
	}
	if got["name"] != "demo" {
		t.Fatalf("decoded name = %q, want %q", got["name"], "demo")
	}
}

func TestReadJSONFileRequiresPath(t *testing.T) {
	_, err := ReadJSONFile[map[string]any]("")
	if err == nil {
		t.Fatal("expected missing path to fail")
	}
	if !strings.Contains(err.Error(), "JSON payload file is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadJSONFileRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload.json")
	if err := os.WriteFile(path, []byte(`{"name":`), 0600); err != nil {
		t.Fatalf("failed to write payload: %v", err)
	}

	_, err := ReadJSONFile[map[string]any](path)
	if err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON payload file") {
		t.Fatalf("unexpected error: %v", err)
	}
}
