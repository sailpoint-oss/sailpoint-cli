package ui_plugins

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestCommand_Success(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"validate-manifest"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected validate-manifest to succeed, got: %v", err)
	}
	if !strings.Contains(out.String(), "Manifest structure is valid (offline check only).") {
		t.Fatalf("expected success message, got: %s", out.String())
	}
}

func TestValidateManifestCommand_RejectsExtraArgs(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{"validate-manifest", "./other.json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected extra args to fail")
	}
	if !strings.Contains(err.Error(), "unknown command") && !strings.Contains(err.Error(), "accepts 0 arg") {
		t.Fatalf("expected no-args error, got: %v", err)
	}
}

func TestValidateManifestCommand_AliasSuccess(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{"validate"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected validate alias to succeed, got: %v", err)
	}
}

func TestValidateManifestCommand_Failure(t *testing.T) {
	t.Setenv(experimentalUIPluginsEnvVar, "1")
	cwd := t.TempDir()
	writeManifestAtPath(t, filepath.Join(cwd, manifestFileName), `{
  "version": 1,
  "manifest": {
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	restore := chdirForTest(t, cwd)
	defer restore()

	cmd := NewUIPluginsCommand()
	cmd.SetArgs([]string{"validate-manifest"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected validation failure")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected actionable alias error, got: %v", err)
	}
}

func chdirForTest(t *testing.T, target string) func() {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(target); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	return func() {
		_ = os.Chdir(orig)
	}
}

func writeManifestAtPath(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
}

