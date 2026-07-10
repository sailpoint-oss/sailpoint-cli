// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package initialize

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

const uiTemplatesTarballRoot = "ui-plugin-templates-main"

// tarEntry is a single file or directory to place in a test archive.
type tarEntry struct {
	name string // path below the archive root (forward slashes), no trailing slash
	dir  bool
	body string
}

// buildTemplatesTarball builds a gzipped tar archive rooted at
// uiTemplatesTarballRoot containing the given entries.
func buildTemplatesTarball(entries []tarEntry) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	root := uiTemplatesTarballRoot + "/"
	_ = tw.WriteHeader(&tar.Header{Name: root, Typeflag: tar.TypeDir, Mode: 0755})
	for _, e := range entries {
		if e.dir {
			_ = tw.WriteHeader(&tar.Header{Name: root + e.name + "/", Typeflag: tar.TypeDir, Mode: 0755})
			continue
		}
		_ = tw.WriteHeader(&tar.Header{Name: root + e.name, Typeflag: tar.TypeReg, Size: int64(len(e.body)), Mode: 0644})
		_, _ = tw.Write([]byte(e.body))
	}
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

// starterFixtureEntries mirrors the real repo shape: an angular/starter subtree
// plus siblings that must be ignored (repo-root files, and a starter-extra dir
// that shares the "starter" prefix but is not under starter/).
func starterFixtureEntries() []tarEntry {
	return []tarEntry{
		{name: "README.md", body: "# root readme"},
		{name: "SAILPOINT_PLUGIN_GUIDE.md", body: "generic guide"},
		{name: "angular", dir: true},
		{name: "angular/starter", dir: true},
		{name: "angular/starter/package.json", body: `{"name":"starter"}`},
		{name: "angular/starter/src", dir: true},
		{name: "angular/starter/src/app", dir: true},
		{name: "angular/starter/src/app/app.ts", body: "export class App {}"},
		{name: "angular/starter-extra", dir: true},
		{name: "angular/starter-extra/notes.md", body: "sibling, must be ignored"},
	}
}

func chdirTemp(t *testing.T) string {
	t.Helper()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })
	workdir := t.TempDir()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return workdir
}

func TestExtractSubtree_StripsPrefixAndIgnoresSiblings(t *testing.T) {
	workdir := chdirTemp(t)
	dest := "acme-plugin"

	if err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(starterFixtureEntries())), dest, "angular/starter"); err != nil {
		t.Fatalf("ExtractSubtree: %v", err)
	}

	root := filepath.Join(workdir, dest)

	// Subtree contents land at the destination root, prefix stripped.
	mustExist(t, filepath.Join(root, "package.json"))
	mustExist(t, filepath.Join(root, "src", "app", "app.ts"))

	// Repo-root siblings are not extracted.
	mustNotExist(t, filepath.Join(root, "README.md"))
	mustNotExist(t, filepath.Join(root, "SAILPOINT_PLUGIN_GUIDE.md"))

	// A sibling that only shares the "starter" prefix is not pulled in.
	mustNotExist(t, filepath.Join(root, "notes.md"))
	mustNotExist(t, filepath.Join(root, "starter-extra"))

	// No leftover "angular/" nesting inside the destination.
	mustNotExist(t, filepath.Join(root, "angular"))
}

func TestExtractSubtree_MissingSubdir(t *testing.T) {
	chdirTemp(t)
	err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(starterFixtureEntries())), "dest", "react/starter")
	if err == nil {
		t.Fatal("expected error for absent subdir")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestExtractSubtree_AlreadyExists(t *testing.T) {
	workdir := chdirTemp(t)
	dest := "exists"
	if err := os.Mkdir(filepath.Join(workdir, dest), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(starterFixtureEntries())), dest, "angular/starter")
	if err == nil {
		t.Fatal("expected error when destination already exists")
	}
	if !contains(err.Error(), "already exists") {
		t.Errorf("error should mention already exists, got: %v", err)
	}
}

func TestExtractSubtree_EmptyArgs(t *testing.T) {
	if err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(nil)), "", "angular/starter"); err == nil {
		t.Error("expected error for empty destDir")
	}
	if err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(nil)), "dest", ""); err == nil {
		t.Error("expected error for empty subdir")
	}
}

func TestExtractSubtree_RejectsZipSlip(t *testing.T) {
	workdir := chdirTemp(t)

	entries := []tarEntry{
		{name: "angular", dir: true},
		{name: "angular/starter", dir: true},
		{name: "angular/starter/../../evil.txt", body: "must not be written"},
	}
	err := ExtractSubtree(bytes.NewReader(buildTemplatesTarball(entries)), "dest", "angular/starter")
	if err == nil {
		t.Fatal("expected error for path traversal inside subtree")
	}
	if !contains(err.Error(), "unsafe path") {
		t.Errorf("error should mention unsafe path, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(workdir, "evil.txt")); statErr == nil {
		t.Fatalf("Zip Slip must not create file outside destination")
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}

func mustNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected %s NOT to exist", path)
	}
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
