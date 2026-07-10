package ui_plugins

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugifyAlias(t *testing.T) {
	cases := map[string]string{
		"My Plugin":       "my-plugin",
		"My Plugin!":      "my-plugin",
		"  Access  Req  ": "access-req",
		"UPPER_case":      "upper-case",
		"a--b":            "a-b",
		"!!!":             "",
		"café99":          "caf-99",
	}
	for in, want := range cases {
		if got := slugifyAlias(in); got != want {
			t.Errorf("slugifyAlias(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGenerateWorkspaceManifest_Valid(t *testing.T) {
	cfg := generateWorkspaceManifest("acme-plugin", "Acme Plugin", "./dist/app", 4200)
	if err := validateWorkspaceManifest(cfg); err != nil {
		t.Fatalf("generated manifest failed validation: %v", err)
	}
	if cfg.Manifest.Alias != "acme-plugin" {
		t.Errorf("alias = %q", cfg.Manifest.Alias)
	}
	if cfg.Manifest.Name["en"] != "Acme Plugin" || cfg.Manifest.Description["en"] != "Acme Plugin" {
		t.Errorf("name/description not set from plugin name: %+v", cfg.Manifest)
	}
	if cfg.Build == nil || cfg.Build.OutDir != "./dist/app" || cfg.Build.Port == nil || *cfg.Build.Port != 4200 {
		t.Errorf("build config wrong: %+v", cfg.Build)
	}
	if len(cfg.Manifest.Slots) != 1 || cfg.Manifest.Slots[0].SlotID != "full-page" {
		t.Errorf("slots wrong: %+v", cfg.Manifest.Slots)
	}
	// Non-nil empty policy maps satisfy the validator's presence requirement.
	if cfg.Manifest.ContentSecurityPolicies == nil || cfg.Manifest.PermissionPolicy == nil || cfg.Manifest.IframeAllow == nil {
		t.Error("policy maps must be non-nil")
	}
}

// writeStarterFixture creates a minimal copy of the Angular starter's
// personalization-relevant files (with the sentinel token) under dir.
func writeStarterFixture(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"angular.json": `{
  "projects": {
    "starter": {
      "architect": {
        "serve": {
          "configurations": {
            "production": { "buildTarget": "starter:build:production" },
            "development": { "buildTarget": "starter:build:development" }
          }
        }
      }
    }
  }
}`,
		"package.json":        `{` + "\n" + `  "name": "starter",` + "\n" + `  "version": "0.0.0"` + "\n" + `}`,
		"sp-ui-plugin.json":   `{"version":1,"manifest":{"alias":"","name":{"en":""},"description":{"en":""},"apiScopes":["sp:scopes:all"],"permissionPolicy":{},"iframeAllow":{},"contentSecurityPolicies":{},"slots":[{"slotId":"full-page"}]},"build":{"outDir":"./dist/starter/browser","port":4200}}`,
		"src/app/app.ts":      "protected readonly title = signal('starter');",
		"src/app/app.spec.ts": "expect(x).toContain('Hello, starter');",
		"src/index.html":      "<head><title>Starter</title></head>",
	}
	for rel, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

func TestPersonalizeScaffold(t *testing.T) {
	dir := t.TempDir()
	writeStarterFixture(t, dir)

	if err := personalizeScaffold(dir, "acme-plugin", "O'Brien & Co"); err != nil {
		t.Fatalf("personalizeScaffold: %v", err)
	}

	angular := readFile(t, filepath.Join(dir, "angular.json"))
	if !strings.Contains(angular, `"acme-plugin": {`) {
		t.Error("angular.json project not renamed")
	}
	if !strings.Contains(angular, `"acme-plugin:build:production"`) || !strings.Contains(angular, `"acme-plugin:build:development"`) {
		t.Error("angular.json buildTargets not rewritten")
	}
	if strings.Contains(angular, "starter") {
		t.Error("angular.json still contains sentinel")
	}

	pkg := readFile(t, filepath.Join(dir, "package.json"))
	if !strings.Contains(pkg, `"name": "acme-plugin"`) {
		t.Error("package.json name not set")
	}

	// Manifest is re-generated and valid, with identity + alias-derived outDir.
	raw := readFile(t, filepath.Join(dir, manifestFileName))
	var cfg uiPluginWorkspaceConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("manifest not valid JSON: %v", err)
	}
	if cfg.Manifest.Alias != "acme-plugin" {
		t.Errorf("manifest alias = %q", cfg.Manifest.Alias)
	}
	if cfg.Manifest.Name["en"] != "O'Brien & Co" {
		t.Errorf("manifest name = %q", cfg.Manifest.Name["en"])
	}
	if cfg.Build.OutDir != "./dist/acme-plugin/browser" {
		t.Errorf("outDir = %q", cfg.Build.OutDir)
	}

	// Display content uses the (escaped) name and stays in lockstep.
	appTs := readFile(t, filepath.Join(dir, "src", "app", "app.ts"))
	appSpec := readFile(t, filepath.Join(dir, "src", "app", "app.spec.ts"))
	if !strings.Contains(appTs, `signal('O\'Brien & Co')`) {
		t.Errorf("app.ts not personalized/escaped: %s", appTs)
	}
	if !strings.Contains(appSpec, `Hello, O\'Brien & Co`) {
		t.Errorf("app.spec.ts not personalized/escaped: %s", appSpec)
	}
	index := readFile(t, filepath.Join(dir, "src", "index.html"))
	if !strings.Contains(index, "<title>O&#39;Brien &amp; Co</title>") {
		t.Errorf("index.html title not HTML-escaped: %s", index)
	}
}

func TestPersonalizeScaffold_AliasEqualsSentinel(t *testing.T) {
	dir := t.TempDir()
	writeStarterFixture(t, dir)
	// alias == "starter" is legitimate; the backstop must not trip.
	if err := personalizeScaffold(dir, "starter", "Starter"); err != nil {
		t.Fatalf("personalizeScaffold with alias=sentinel: %v", err)
	}
}

func TestCheckAlias(t *testing.T) {
	ctx := context.Background()

	// Definitive negatives fail.
	conflict := &fakeClient{getStatus: 409}
	if err := checkAlias(ctx, conflict, &strings.Builder{}, "taken"); err == nil || !strings.Contains(err.Error(), "already in use") {
		t.Errorf("409 should report conflict, got %v", err)
	}
	bad := &fakeClient{getStatus: 400, getBody: `{"message":"too short"}`}
	if err := checkAlias(ctx, bad, &strings.Builder{}, "ab"); err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("400 should report invalid, got %v", err)
	}

	// Available passes silently.
	ok := &fakeClient{getStatus: 200}
	var okWarn strings.Builder
	if err := checkAlias(ctx, ok, &okWarn, "free-alias"); err != nil {
		t.Errorf("200 should pass, got %v", err)
	}
	if okWarn.Len() != 0 {
		t.Errorf("200 should not warn, got %q", okWarn.String())
	}

	// Inconclusive results warn and continue (no error).
	// nil client:
	var w1 strings.Builder
	if err := checkAlias(ctx, nil, &w1, "x"); err != nil {
		t.Errorf("nil client should not error (best-effort), got %v", err)
	}
	if !strings.Contains(w1.String(), "Warning") {
		t.Errorf("nil client should emit a warning, got %q", w1.String())
	}
	// transport error:
	var w2 strings.Builder
	if err := checkAlias(ctx, &fakeClient{getErr: errors.New("boom")}, &w2, "x"); err != nil {
		t.Errorf("transport error should not error (best-effort), got %v", err)
	}
	if !strings.Contains(w2.String(), "Warning") {
		t.Errorf("transport error should emit a warning, got %q", w2.String())
	}
	// non-2xx (e.g. forbidden):
	var w3 strings.Builder
	if err := checkAlias(ctx, &fakeClient{getStatus: 403}, &w3, "x"); err != nil {
		t.Errorf("403 should not error (best-effort), got %v", err)
	}
	if !strings.Contains(w3.String(), "Warning") {
		t.Errorf("403 should emit a warning, got %q", w3.String())
	}
}

func TestRunInit_PathAttach_Headless(t *testing.T) {
	target := t.TempDir()
	// A pre-existing user file must be preserved untouched.
	userFile := filepath.Join(target, "keep.txt")
	if err := os.WriteFile(userFile, []byte("original"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	guideFetched := false
	deps := initDeps{
		client: &fakeClient{getStatus: 200},
		in:     strings.NewReader(""),
		out:    &strings.Builder{},
		errOut: &strings.Builder{},
		fetchGuide: func(destPath string) error {
			guideFetched = true
			return os.WriteFile(destPath, []byte("guide"), 0644)
		},
	}
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Manifest generated and valid.
	cfg, err := loadAndValidateWorkspaceManifest(filepath.Join(target, manifestFileName))
	if err != nil {
		t.Fatalf("generated manifest invalid: %v", err)
	}
	if cfg.Manifest.Alias != "my-plugin" {
		t.Errorf("alias = %q, want my-plugin (slug default)", cfg.Manifest.Alias)
	}
	if cfg.Build.OutDir != "./dist/app" || cfg.Build.Port == nil || *cfg.Build.Port != defaultDevServerPort {
		t.Errorf("build config = %+v (want outDir ./dist/app, port %d)", cfg.Build, defaultDevServerPort)
	}
	if !guideFetched {
		t.Error("guide was not fetched")
	}
	if data, _ := os.ReadFile(userFile); string(data) != "original" {
		t.Error("pre-existing user file was modified")
	}
}

func TestRunInit_AliasConflict_NoMutation(t *testing.T) {
	target := t.TempDir()
	deps := initDeps{
		client:     &fakeClient{getStatus: 409},
		in:         strings.NewReader(""),
		out:        &strings.Builder{},
		errOut:     &strings.Builder{},
		fetchGuide: func(string) error { return nil },
	}
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	if err := runInit(context.Background(), deps, opts); err == nil {
		t.Fatal("expected alias-conflict error")
	}
	if _, err := os.Stat(filepath.Join(target, manifestFileName)); err == nil {
		t.Error("manifest must not be written when alias validation fails")
	}
}

func TestRunInit_PathMissingOutDir_Headless(t *testing.T) {
	target := t.TempDir()
	deps := initDeps{
		client:     &fakeClient{getStatus: 200},
		in:         strings.NewReader(""),
		out:        &strings.Builder{},
		errOut:     &strings.Builder{},
		fetchGuide: func(string) error { return nil },
	}
	opts := initOptions{name: "My Plugin", path: target} // no out-dir

	if err := runInit(context.Background(), deps, opts); err == nil || !strings.Contains(err.Error(), "out-dir") {
		t.Fatalf("expected out-dir required error, got %v", err)
	}
}

func TestResolveAndCheckAlias_RepromptsOnConflict(t *testing.T) {
	// Interactive: first alias taken (409), second available (200) → succeeds
	// with the second, after telling the user why.
	client := &seqClient{getQueue: []stubResp{{status: 409}, {status: 200}}}
	r := bufio.NewReader(strings.NewReader("taken\ngood-alias\n"))
	var out strings.Builder
	deps := initDeps{client: client, out: &out, errOut: &strings.Builder{}}

	alias, err := resolveAndCheckAlias(context.Background(), r, deps, initOptions{}, "My Plugin", true)
	if err != nil {
		t.Fatalf("expected success after re-prompt, got %v", err)
	}
	if alias != "good-alias" {
		t.Errorf("alias = %q, want good-alias", alias)
	}
	if !strings.Contains(out.String(), "please choose a different alias") {
		t.Errorf("expected a re-prompt notice, got %q", out.String())
	}
}

func TestResolveAndCheckAlias_HeadlessConflictFailsFast(t *testing.T) {
	// Headless: a definitive conflict fails immediately (no prompting).
	client := &seqClient{getQueue: []stubResp{{status: 409}}}
	deps := initDeps{client: client, out: &strings.Builder{}, errOut: &strings.Builder{}}

	_, err := resolveAndCheckAlias(context.Background(), bufio.NewReader(strings.NewReader("")), deps, initOptions{alias: "taken"}, "My Plugin", false)
	if err == nil || !strings.Contains(err.Error(), "already in use") {
		t.Fatalf("headless conflict should fail fast, got %v", err)
	}
}

func TestResolveAndCheckAlias_EOFBoundsReprompt(t *testing.T) {
	// Interactive with exhausted stdin: promptLine keeps returning the default,
	// which stays taken (409), so the loop must be bounded and then error.
	queue := make([]stubResp, maxAliasPrompts)
	for i := range queue {
		queue[i] = stubResp{status: 409}
	}
	client := &seqClient{getQueue: queue}
	deps := initDeps{client: client, out: &strings.Builder{}, errOut: &strings.Builder{}}

	_, err := resolveAndCheckAlias(context.Background(), bufio.NewReader(strings.NewReader("")), deps, initOptions{}, "My Plugin", true)
	if err == nil || !strings.Contains(err.Error(), "no valid alias") {
		t.Fatalf("expected bounded re-prompt to error, got %v", err)
	}
}

func TestRunInit_PathInconclusiveAlias_Proceeds(t *testing.T) {
	target := t.TempDir()
	var errOut strings.Builder
	deps := initDeps{
		client:     &fakeClient{getStatus: 403}, // cannot validate (e.g. bad PAT)
		in:         strings.NewReader(""),
		out:        &strings.Builder{},
		errOut:     &errOut,
		fetchGuide: func(destPath string) error { return os.WriteFile(destPath, []byte("guide"), 0644) },
	}
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("inconclusive alias validation should not block init, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "Warning") {
		t.Errorf("expected a validation warning, got %q", errOut.String())
	}
	// It proceeded: the manifest was generated.
	if _, err := loadAndValidateWorkspaceManifest(filepath.Join(target, manifestFileName)); err != nil {
		t.Errorf("manifest should have been written despite inconclusive validation: %v", err)
	}
}

func TestRunInit_PathControlCharsOutDir(t *testing.T) {
	target := t.TempDir()
	deps := initDeps{
		client:     &fakeClient{getStatus: 200},
		in:         strings.NewReader(""),
		out:        &strings.Builder{},
		errOut:     &strings.Builder{},
		fetchGuide: func(string) error { return nil },
	}
	opts := initOptions{name: "My Plugin", path: target, outDir: "dist/\x01app"}

	if err := runInit(context.Background(), deps, opts); err == nil || !strings.Contains(err.Error(), "control characters") {
		t.Fatalf("expected control-character error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, manifestFileName)); err == nil {
		t.Error("manifest must not be written when out-dir is invalid")
	}
}

func TestRunInit_PathConflict_SkipWithoutForce(t *testing.T) {
	target := t.TempDir()
	manifestPath := filepath.Join(target, manifestFileName)
	if err := os.WriteFile(manifestPath, []byte("preexisting"), 0644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	deps := initDeps{
		client:     &fakeClient{getStatus: 200},
		in:         strings.NewReader("n\nn\n"), // decline both overwrite prompts
		out:        &strings.Builder{},
		errOut:     &strings.Builder{},
		fetchGuide: func(destPath string) error { return os.WriteFile(destPath, []byte("guide"), 0644) },
	}
	opts := initOptions{name: "My Plugin", path: target, outDir: "./dist/app"}

	// Interactive (name empty would prompt); here name is set so headless, but
	// conflict prompts still read from stdin. Provide "n" to decline.
	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}
	if data, _ := os.ReadFile(manifestPath); string(data) != "preexisting" {
		t.Error("manifest should be left untouched when overwrite declined")
	}
}

func TestScaffoldFlow_PersonalizesFromStub(t *testing.T) {
	workdir := initTestChdir(t)
	deps := initDeps{
		client: &fakeClient{getStatus: 200},
		in:     strings.NewReader(""),
		out:    &strings.Builder{},
		errOut: &strings.Builder{},
		scaffold: func(destDir string) error {
			writeStarterFixture(t, destDir)
			return nil
		},
	}
	opts := initOptions{name: "My Plugin"} // new-workspace, headless; alias slug = my-plugin

	if err := runInit(context.Background(), deps, opts); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	dest := filepath.Join(workdir, "my-plugin")
	pkg := readFile(t, filepath.Join(dest, "package.json"))
	if !strings.Contains(pkg, `"name": "my-plugin"`) {
		t.Errorf("scaffold not personalized: %s", pkg)
	}
}

func TestScaffoldFlow_DestExists_HardFail(t *testing.T) {
	workdir := initTestChdir(t)
	if err := os.Mkdir(filepath.Join(workdir, "my-plugin"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	scaffoldCalled := false
	deps := initDeps{
		client:   &fakeClient{getStatus: 200},
		in:       strings.NewReader(""),
		out:      &strings.Builder{},
		errOut:   &strings.Builder{},
		scaffold: func(string) error { scaffoldCalled = true; return nil },
	}
	opts := initOptions{name: "My Plugin"}

	if err := runInit(context.Background(), deps, opts); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got %v", err)
	}
	if scaffoldCalled {
		t.Error("scaffold must not run when destination already exists")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func initTestChdir(t *testing.T) string {
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
