package ui_plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, angularManifestFileName)
	original := []byte(`{"version":1}`)
	replacement := []byte(`{"version":1,"updated":true}` + "\n")

	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := writeFileAtomic(path, replacement); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(replacement) {
		t.Fatalf("content = %q, want %q", got, replacement)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	// Windows has no Unix permission model; Mode().Perm() only reflects the
	// read-only bit, so exact-mode assertions cannot hold there.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("mode = %o, want %o", info.Mode().Perm(), os.FileMode(0600))
	}

	matches, err := filepath.Glob(filepath.Join(dir, ".*"+angularManifestFileName+".tmp-*"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no leftover temp files, found: %v", matches)
	}
}

func TestWriteFileAtomicCreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, angularManifestFileName)
	data := []byte("{\n}\n")

	if err := writeFileAtomic(path, data); err != nil {
		t.Fatalf("writeFileAtomic: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("content = %q, want %q", got, data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	// Windows has no Unix permission model; Mode().Perm() only reflects the
	// read-only bit, so exact-mode assertions cannot hold there.
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0644 {
		t.Fatalf("mode = %o, want %o", info.Mode().Perm(), os.FileMode(0644))
	}
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, angularManifestFileName)

	perm, err := filePermissions(path, 0644)
	if err != nil {
		t.Fatalf("missing file: %v", err)
	}
	if perm != 0644 {
		t.Fatalf("default perm = %o, want %o", perm, os.FileMode(0644))
	}

	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	perm, err = filePermissions(path, 0644)
	if err != nil {
		t.Fatalf("existing file: %v", err)
	}
	// Windows has no Unix permission model; Mode().Perm() only reflects the
	// read-only bit, so exact-mode assertions cannot hold there.
	if runtime.GOOS != "windows" && perm != 0600 {
		t.Fatalf("existing perm = %o, want %o", perm, os.FileMode(0600))
	}
}
