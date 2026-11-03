package scopes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	scopeanalysis "github.com/sailpoint-oss/sailpoint-cli/internal/scopes"
)

const testManifestJSON = `{
  "schemaVersion": 1,
  "sdk": "typescript",
  "sdkVersion": "9.9.9",
  "operations": {
    "listAccountsV1": {"sdkName": "listAccountsV1", "httpMethod": "GET", "path": "/accounts/v1",
      "userAuth": ["idn:accounts:read", "idn:accounts:manage"], "applicationAuth": ["idn:accounts:read"]},
    "listIdentitiesV1": {"sdkName": "listIdentitiesV1", "httpMethod": "GET", "path": "/identities/v1",
      "userAuth": ["idn:identity:read"], "applicationAuth": null}
  }
}`

func setupProject(t *testing.T) (string, analyzeOptions) {
	t.Helper()
	dir := t.TempDir()
	src := "import { AccountsApi, IdentitiesApi } from '@sailpoint/api-client';\n" +
		"await accounts.listAccountsV1({ limit: 10 });\n" +
		"await identities.listIdentitiesV1();\n"
	if err := os.WriteFile(filepath.Join(dir, "index.ts"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "sailpoint-scopes.json")
	if err := os.WriteFile(manifestPath, []byte(testManifestJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	opts := analyzeOptions{
		path:     dir,
		authType: "user",
		locate: func(lang scopeanalysis.Language, _ []string) (string, bool) {
			return manifestPath, lang == scopeanalysis.LangTypeScript
		},
	}
	return dir, opts
}

func TestRunAnalyzeTable(t *testing.T) {
	_, opts := setupProject(t)
	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"LOCATION", "index.ts:2", "listAccountsV1", "idn:accounts:read | idn:accounts:manage", "Required scopes (user auth)", "SDK 9.9.9"} {
		if !strings.Contains(got, want) {
			t.Errorf("output does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Warnings:") {
		t.Errorf("unexpected warnings:\n%s", got)
	}
}

func TestRunAnalyzeScopesOnly(t *testing.T) {
	_, opts := setupProject(t)
	opts.scopesOnly = true
	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "idn:accounts:read,idn:identity:read\n" {
		t.Errorf("output = %q", got)
	}

	opts.authType = "application"
	out.Reset()
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "idn:accounts:read\n" {
		t.Errorf("application output = %q", got)
	}
	if !strings.Contains(errOut.String(), "listIdentitiesV1 cannot be called with application auth") {
		t.Errorf("stderr = %q", errOut.String())
	}
}

func TestRunAnalyzeJSON(t *testing.T) {
	_, opts := setupProject(t)
	opts.jsonOutput = true
	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	var res scopeanalysis.Result
	if err := json.Unmarshal(out.Bytes(), &res); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	if len(res.Calls) != 2 || len(res.Operations) != 2 || len(res.RequiredScopes) != 2 {
		t.Errorf("result = %+v", res)
	}
	if len(res.Sources) != 1 || res.Sources[0].Embedded {
		t.Errorf("sources = %+v", res.Sources)
	}
}

func TestRunAnalyzeUpdateManifest(t *testing.T) {
	dir, opts := setupProject(t)
	manifest := `{
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
	path := filepath.Join(dir, "sp-ui-plugin.json")
	if err := os.WriteFile(path, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	opts.scopesOnly = true
	opts.updateManifest = "sp-ui-plugin.json"

	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var written struct {
		Manifest struct {
			APIScopes []string `json:"apiScopes"`
		} `json:"manifest"`
	}
	if err := json.Unmarshal(raw, &written); err != nil {
		t.Fatal(err)
	}
	if strings.Join(written.Manifest.APIScopes, ",") != "idn:accounts:read,idn:identity:read" {
		t.Errorf("apiScopes = %v", written.Manifest.APIScopes)
	}
}

func TestRunAnalyzeNoCalls(t *testing.T) {
	opts := analyzeOptions{path: t.TempDir(), authType: "user"}
	var out, errOut bytes.Buffer
	if err := runAnalyze(opts, &out, &errOut); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No SailPoint SDK calls found.") {
		t.Errorf("output = %q", out.String())
	}
}

func TestRunAnalyzeBadAuthType(t *testing.T) {
	opts := analyzeOptions{path: t.TempDir(), authType: "service"}
	if err := runAnalyze(opts, &bytes.Buffer{}, &bytes.Buffer{}); err == nil {
		t.Error("expected an error for an invalid auth type")
	}
}
