package ui_plugins

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const skillsTarballRoot = "ui-plugin-templates-main"

// buildTemplatesArchive builds a gzipped tar archive rooted at skillsTarballRoot
// from files keyed by repo-root-relative path (forward slashes). Parent
// directories are created implicitly by extraction, so only file entries are
// written.
func buildTemplatesArchive(files map[string]string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	root := skillsTarballRoot + "/"
	_ = tw.WriteHeader(&tar.Header{Name: root, Typeflag: tar.TypeDir, Mode: 0755})
	for name, body := range files {
		_ = tw.WriteHeader(&tar.Header{Name: root + name, Typeflag: tar.TypeReg, Size: int64(len(body)), Mode: 0644})
		_, _ = tw.Write([]byte(body))
	}
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}

// archiveFetcher returns a fetchArchive seam that serves a fresh reader over the
// given archive on each call (callers buffer it, but model independent reads).
func archiveFetcher(t *testing.T, archive []byte) func() (io.ReadCloser, error) {
	t.Helper()
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(archive)), nil
	}
}

// starterTarFiles builds the file set for a templates archive: the Angular
// starter fixture under angular/starter/, plus any extra repo-relative files
// (e.g. root or starter skills) supplied by the caller.
func starterTarFiles(extra map[string]string) map[string]string {
	files := map[string]string{}
	for rel, body := range starterFixtureFiles() {
		files["angular/starter/"+rel] = body
	}
	for rel, body := range extra {
		files[rel] = body
	}
	return files
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

func skillMD(name, extraFrontmatter string) string {
	fm := "name: " + name + "\ndescription: test skill " + name
	if extraFrontmatter != "" {
		fm += "\n" + extraFrontmatter
	}
	return "---\n" + fm + "\n---\n\n# " + name + "\n\nbody\n"
}

func TestNewWorkspace_DeliversRootSkills(t *testing.T) {
	workdir := initTestChdir(t)
	archive := buildTemplatesArchive(starterTarFiles(map[string]string{
		".agents/skills/plugin-lifecycle-cli/SKILL.md":           skillMD("plugin-lifecycle-cli", ""),
		"angular/starter/.agents/skills/plugin-routing/SKILL.md": skillMD("plugin-routing", ""),
	}))
	out := &strings.Builder{}
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          out,
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	if err := runInit(context.Background(), deps, initOptions{name: "My Plugin"}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	dest := filepath.Join(workdir, "my-plugin")
	// Root agnostic skill merged in.
	mustExist(t, filepath.Join(dest, ".agents", "skills", "plugin-lifecycle-cli", "SKILL.md"))
	// Starter's own skill came along with the starter subtree.
	mustExist(t, filepath.Join(dest, ".agents", "skills", "plugin-routing", "SKILL.md"))
	if !strings.Contains(out.String(), "Agent skills: added 1") {
		t.Errorf("summary should report one added root skill, got: %q", out.String())
	}
}

func TestNewWorkspace_SkipsPathOnlySkill(t *testing.T) {
	workdir := initTestChdir(t)
	archive := buildTemplatesArchive(starterTarFiles(map[string]string{
		".agents/skills/agnostic/SKILL.md":  skillMD("agnostic", ""),
		".agents/skills/path-only/SKILL.md": skillMD("path-only", "metadata:\n  sailpoint-merge: path-only"),
	}))
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          &strings.Builder{},
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	if err := runInit(context.Background(), deps, initOptions{name: "My Plugin"}); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	dest := filepath.Join(workdir, "my-plugin")
	mustExist(t, filepath.Join(dest, ".agents", "skills", "agnostic", "SKILL.md"))
	mustNotExist(t, filepath.Join(dest, ".agents", "skills", "path-only", "SKILL.md"))
}

func TestNewWorkspace_SkillCollision_FailsLoudly(t *testing.T) {
	workdir := initTestChdir(t)
	// Same folder name at the repo root and in the starter.
	archive := buildTemplatesArchive(starterTarFiles(map[string]string{
		".agents/skills/dup/SKILL.md":                 skillMD("dup", ""),
		"angular/starter/.agents/skills/dup/SKILL.md": skillMD("dup", ""),
	}))
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          &strings.Builder{},
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	err := runInit(context.Background(), deps, initOptions{name: "My Plugin"})
	if err == nil || !strings.Contains(err.Error(), "both") {
		t.Fatalf("expected a loud collision error, got %v", err)
	}
	// All-or-nothing: the destination must be removed on failure.
	mustNotExist(t, filepath.Join(workdir, "my-plugin"))
}

func TestNewWorkspace_PathOnlyCollisionAllowed(t *testing.T) {
	workdir := initTestChdir(t)
	// A path-only root skill may share a name with a starter skill: it is never
	// merged into the framework, so there is no collision.
	archive := buildTemplatesArchive(starterTarFiles(map[string]string{
		".agents/skills/dup/SKILL.md":                 skillMD("dup", "metadata:\n  sailpoint-merge: path-only"),
		"angular/starter/.agents/skills/dup/SKILL.md": skillMD("dup", ""),
	}))
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          &strings.Builder{},
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	if err := runInit(context.Background(), deps, initOptions{name: "My Plugin"}); err != nil {
		t.Fatalf("path-only name reuse must not collide, got %v", err)
	}
	// The starter's version is the one that lands.
	mustExist(t, filepath.Join(workdir, "my-plugin", ".agents", "skills", "dup", "SKILL.md"))
}

func TestNewWorkspace_NoSkillsFlag(t *testing.T) {
	workdir := initTestChdir(t)
	archive := buildTemplatesArchive(starterTarFiles(map[string]string{
		".agents/skills/agnostic/SKILL.md": skillMD("agnostic", ""),
	}))
	out := &strings.Builder{}
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          out,
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	if err := runInit(context.Background(), deps, initOptions{name: "My Plugin", noSkills: true}); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	mustNotExist(t, filepath.Join(workdir, "my-plugin", ".agents", "skills", "agnostic"))
	if strings.Contains(out.String(), "Agent skills:") {
		t.Errorf("no-skills run should not print a skills summary, got: %q", out.String())
	}
}

func TestNewWorkspace_TemplateWithoutSkills_Skips(t *testing.T) {
	workdir := initTestChdir(t)
	// Starter only; no .agents/skills anywhere in the archive.
	archive := buildTemplatesArchive(starterTarFiles(nil))
	out := &strings.Builder{}
	deps := initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(""),
		out:          out,
		errOut:       &strings.Builder{},
		fetchArchive: archiveFetcher(t, archive),
	}

	if err := runInit(context.Background(), deps, initOptions{name: "My Plugin"}); err != nil {
		t.Fatalf("missing .agents/skills must not fail init, got %v", err)
	}
	mustExist(t, filepath.Join(workdir, "my-plugin", "package.json"))
	if strings.Contains(out.String(), "Agent skills:") {
		t.Errorf("no skills in template should print no skills message, got: %q", out.String())
	}
}

// --- Attach flow ---

func attachDeps(t *testing.T, stdin string, out, errOut *strings.Builder, extraSkills map[string]string) initDeps {
	t.Helper()
	return initDeps{
		client:       &fakeClient{getStatus: 200},
		in:           strings.NewReader(stdin),
		out:          out,
		errOut:       errOut,
		fetchArchive: archiveFetcher(t, buildTemplatesArchive(starterTarFiles(extraSkills))),
		fetchGuide:   func(destPath string) error { return os.WriteFile(destPath, []byte("guide"), 0644) },
	}
}

func TestAttach_MergesWithoutClobbering(t *testing.T) {
	target := t.TempDir()
	// A pre-existing unrelated user skill must survive untouched.
	userSkill := filepath.Join(target, ".agents", "skills", "my-own-skill")
	if err := os.MkdirAll(userSkill, 0755); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(userSkill, "SKILL.md"), []byte("user content"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	out := &strings.Builder{}
	deps := attachDeps(t, "", out, &strings.Builder{}, map[string]string{
		".agents/skills/plugin-lifecycle-cli/SKILL.md": skillMD("plugin-lifecycle-cli", ""),
		".agents/skills/path-only/SKILL.md":            skillMD("path-only", "metadata:\n  sailpoint-merge: path-only"),
	})
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Repo-wide skills delivered, including the path-only one (attach is its only path).
	mustExist(t, filepath.Join(target, ".agents", "skills", "plugin-lifecycle-cli", "SKILL.md"))
	mustExist(t, filepath.Join(target, ".agents", "skills", "path-only", "SKILL.md"))
	// Unrelated user skill untouched.
	if data := readFile(t, filepath.Join(userSkill, "SKILL.md")); data != "user content" {
		t.Errorf("unrelated user skill modified: %q", data)
	}
	if !strings.Contains(out.String(), "Agent skills: added 2") {
		t.Errorf("summary should report two added skills, got: %q", out.String())
	}
}

func TestAttach_OverwritePromptDeclined(t *testing.T) {
	target := t.TempDir()
	existing := filepath.Join(target, ".agents", "skills", "plugin-lifecycle-cli")
	if err := os.MkdirAll(existing, 0755); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(existing, "SKILL.md"), []byte("original"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	out := &strings.Builder{}
	// Decline the single overwrite prompt.
	deps := attachDeps(t, "n\n", out, &strings.Builder{}, map[string]string{
		".agents/skills/plugin-lifecycle-cli/SKILL.md": skillMD("plugin-lifecycle-cli", ""),
	})
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if data := readFile(t, filepath.Join(existing, "SKILL.md")); data != "original" {
		t.Errorf("declined overwrite should leave the skill untouched, got %q", data)
	}
	if !strings.Contains(out.String(), "skipped 1") {
		t.Errorf("summary should report a skip, got: %q", out.String())
	}
}

func TestAttach_ForceOverwrites(t *testing.T) {
	target := t.TempDir()
	existing := filepath.Join(target, ".agents", "skills", "plugin-lifecycle-cli")
	if err := os.MkdirAll(existing, 0755); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(existing, "SKILL.md"), []byte("original"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	out := &strings.Builder{}
	deps := attachDeps(t, "", out, &strings.Builder{}, map[string]string{
		".agents/skills/plugin-lifecycle-cli/SKILL.md": skillMD("plugin-lifecycle-cli", ""),
	})
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app", force: true}

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if data := readFile(t, filepath.Join(existing, "SKILL.md")); !strings.Contains(data, "# plugin-lifecycle-cli") {
		t.Errorf("force should overwrite with the delivered skill, got %q", data)
	}
	if !strings.Contains(out.String(), "overwritten 1") {
		t.Errorf("summary should report an overwrite, got: %q", out.String())
	}
}

func TestAttach_IdempotentReRun(t *testing.T) {
	target := t.TempDir()
	extra := map[string]string{
		".agents/skills/plugin-lifecycle-cli/SKILL.md": skillMD("plugin-lifecycle-cli", ""),
	}

	// First run: fresh add.
	deps1 := attachDeps(t, "", &strings.Builder{}, &strings.Builder{}, extra)
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}
	if err := runInit(context.Background(), deps1, opts); err != nil {
		t.Fatalf("first runInit: %v", err)
	}

	// Second run: folder exists; decline the overwrite. No error, no duplicate.
	out := &strings.Builder{}
	deps2 := attachDeps(t, "n\n", out, &strings.Builder{}, extra)
	if err := runInit(context.Background(), deps2, opts); err != nil {
		t.Fatalf("second runInit: %v", err)
	}
	mustExist(t, filepath.Join(target, ".agents", "skills", "plugin-lifecycle-cli", "SKILL.md"))
	if !strings.Contains(out.String(), "skipped 1") {
		t.Errorf("re-run should skip the already-present skill, got: %q", out.String())
	}
}

func TestDeliverSkills_PathOnlyUnit(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(md, []byte(skillMD("x", "metadata:\n  sailpoint-merge: path-only")), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if !isPathOnlySkill(md) {
		t.Error("expected path-only to be detected")
	}
	if err := os.WriteFile(md, []byte(skillMD("x", "")), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if isPathOnlySkill(md) {
		t.Error("unflagged skill must not be path-only")
	}
	// Nesting-agnostic: a top-level key (no metadata parent) still counts, matching the PRB.
	if err := os.WriteFile(md, []byte(skillMD("x", "sailpoint-merge: path-only")), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if !isPathOnlySkill(md) {
		t.Error("expected nesting-agnostic path-only detection")
	}
}
