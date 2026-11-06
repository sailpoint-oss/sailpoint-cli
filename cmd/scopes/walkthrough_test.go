package scopes

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/ui_plugins"
	"github.com/sailpoint-oss/sailpoint-cli/internal/tui"
)

// fakePrompter answers the walkthrough questions in order and records them.
type fakePrompter struct {
	selects  []string
	inputs   []string
	confirms []bool
	asked    []string
}

func (f *fakePrompter) Select(title string, _ []tui.Choice) (string, error) {
	f.asked = append(f.asked, title)
	if len(f.selects) == 0 {
		return "", errCancelled
	}
	id := f.selects[0]
	f.selects = f.selects[1:]
	return id, nil
}

func (f *fakePrompter) Input(label string) (string, error) {
	f.asked = append(f.asked, label)
	if len(f.inputs) == 0 {
		return "", nil
	}
	v := f.inputs[0]
	f.inputs = f.inputs[1:]
	return v, nil
}

func (f *fakePrompter) Confirm(question string) (bool, error) {
	f.asked = append(f.asked, question)
	if len(f.confirms) == 0 {
		return false, nil
	}
	v := f.confirms[0]
	f.confirms = f.confirms[1:]
	return v, nil
}

const walkthroughManifest = `{
    "version": 1,
    "manifest": {
        "alias": "demo",
        "name": {"en": "Demo"},
        "description": {"en": "Demo"},
        "apiScopes": ["sp:scopes:all"],
        "contentSecurityPolicies": {},
        "permissionPolicy": {},
        "iframeAllow": {},
        "slots": [{"slotId": "full-page"}]
    }
}`

func writePluginManifest(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ui_plugins.ManifestFileName)
	if err := os.WriteFile(path, []byte(walkthroughManifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readAPIScopes(t *testing.T, path string) []string {
	t.Helper()
	scopes, err := ui_plugins.ReadManifestAPIScopes(path)
	if err != nil {
		t.Fatal(err)
	}
	return scopes
}

func TestWalkthroughOtherFolderAndUpdate(t *testing.T) {
	dir, opts := setupProject(t)
	path := writePluginManifest(t, dir)
	fake := &fakePrompter{
		selects:  []string{"application", "other"},
		inputs:   []string{filepath.Join(dir, "missing"), dir},
		confirms: []bool{true},
	}
	opts.path = "."
	opts.interactive = true
	opts.prompt = fake

	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Required scopes (application auth)") {
		t.Errorf("output:\n%s", out.String())
	}
	if !strings.Contains(out.String(), filepath.Join(dir, "index.ts")+":2") {
		t.Errorf("expected an absolute location for another folder:\n%s", out.String())
	}
	if !strings.Contains(errOut.String(), "is not a folder") {
		t.Errorf("stderr = %q", errOut.String())
	}
	if got := readAPIScopes(t, path); !slices.Equal(got, []string{"idn:accounts:read", "sp:scopes:all"}) {
		t.Errorf("apiScopes = %v", got)
	}
}

func TestWalkthroughDeclineUpdate(t *testing.T) {
	dir, opts := setupProject(t)
	path := writePluginManifest(t, dir)
	fake := &fakePrompter{confirms: []bool{false}}
	opts.interactive = true
	opts.pathSet = true
	opts.authTypeSet = true
	opts.prompt = fake

	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if len(fake.asked) != 1 {
		t.Errorf("flags must skip their questions, asked = %v", fake.asked)
	}
	if !strings.Contains(errOut.String(), "Current apiScopes: sp:scopes:all") ||
		!strings.Contains(errOut.String(), "was not changed") {
		t.Errorf("stderr = %q", errOut.String())
	}
	if got := readAPIScopes(t, path); !slices.Equal(got, []string{"sp:scopes:all"}) {
		t.Errorf("apiScopes = %v", got)
	}
}

func TestWalkthroughCancel(t *testing.T) {
	_, opts := setupProject(t)
	opts.interactive = true
	opts.prompt = &fakePrompter{}

	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 || !strings.Contains(errOut.String(), "Cancelled.") {
		t.Errorf("stdout = %q, stderr = %q", out.String(), errOut.String())
	}
}

func TestNonInteractiveManifestHint(t *testing.T) {
	dir, opts := setupProject(t)
	path := writePluginManifest(t, dir)

	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "--update-manifest") {
		t.Errorf("stderr = %q", errOut.String())
	}
	if got := readAPIScopes(t, path); !slices.Equal(got, []string{"sp:scopes:all"}) {
		t.Errorf("apiScopes = %v", got)
	}

	// Once the scopes match, there is nothing to offer.
	if err := ui_plugins.UpdateManifestAPIScopes(path, []string{"idn:accounts:read", "idn:identity:read"}); err != nil {
		t.Fatal(err)
	}
	errOut.Reset()
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errOut.String(), "already match") {
		t.Errorf("stderr = %q", errOut.String())
	}
}
