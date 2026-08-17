package ui_plugins

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleCSP = "default-src 'self'; connect-src 'self' https://tenant.api.cloud.sailpoint.com"
const samplePP = "camera=(), microphone=()"

func devHeaders(csp, pp string, setCSP, setPP bool) devDocumentHeaders {
	return devDocumentHeaders{
		contentSecurityPolicy: csp,
		permissionsPolicy:     pp,
		setCSP:                setCSP,
		setPP:                 setPP,
	}
}

func TestParseDevDocumentHeadersFromCreate(t *testing.T) {
	body := `{"pluginInstanceId":"pi-1","alias":"acme","devDocumentHeaders":{"Content-Security-Policy":"` + sampleCSP + `","Permissions-Policy":"` + samplePP + `"}}`
	got, ok := parseDevDocumentHeaders([]byte(body))
	if !ok {
		t.Fatal("expected headers to be present")
	}
	if got.contentSecurityPolicy != sampleCSP || !got.setCSP {
		t.Fatalf("CSP = %q set=%v", got.contentSecurityPolicy, got.setCSP)
	}
	if got.permissionsPolicy != samplePP || !got.setPP {
		t.Fatalf("PP = %q set=%v", got.permissionsPolicy, got.setPP)
	}
}

func TestParseDevDocumentHeaders_EmptyPermissionsPolicyValue(t *testing.T) {
	body := `{"devDocumentHeaders":{"Content-Security-Policy":"` + sampleCSP + `","Permissions-Policy":""}}`
	got, ok := parseDevDocumentHeaders([]byte(body))
	if !ok {
		t.Fatal("expected headers to be present")
	}
	if !got.setCSP || got.contentSecurityPolicy != sampleCSP {
		t.Fatalf("CSP not parsed: %+v", got)
	}
	if !got.setPP || got.permissionsPolicy != "" {
		t.Fatalf("expected empty PP with setPP=true, got %+v", got)
	}
}

func TestParseDevDocumentHeadersFromLink(t *testing.T) {
	body := `{"devOverrides":{"devUrl":"https://localhost:4200"},"devDocumentHeaders":{"Content-Security-Policy":"` + sampleCSP + `"}}`
	got, ok := parseDevDocumentHeaders([]byte(body))
	if !ok {
		t.Fatal("expected headers to be present")
	}
	if got.contentSecurityPolicy != sampleCSP || !got.setCSP {
		t.Fatalf("CSP = %q set=%v", got.contentSecurityPolicy, got.setCSP)
	}
	if got.setPP {
		t.Fatal("PP should not be set when absent from UMS")
	}
}

func TestParseDevDocumentHeaders_MissingField(t *testing.T) {
	if _, ok := parseDevDocumentHeaders([]byte(`{"pluginInstanceId":"pi-1"}`)); ok {
		t.Fatal("expected missing field to report false")
	}
	if _, ok := parseDevDocumentHeaders([]byte(`{"devOverrides":{}}`)); ok {
		t.Fatal("expected missing headers to report false")
	}
	if _, ok := parseDevDocumentHeaders([]byte(`{"devDocumentHeaders":{}}`)); ok {
		t.Fatal("expected empty devDocumentHeaders object to report false")
	}
}

func TestResolveAngularProjectKey(t *testing.T) {
	projects := map[string]json.RawMessage{
		"acme-plugin": json.RawMessage(`{}`),
		"other":       json.RawMessage(`{}`),
	}

	res, err := resolveAngularProjectKey(projects, "acme-plugin")
	if err != nil || res.key != "acme-plugin" || res.usedSoleProjectFallback {
		t.Fatalf("alias match: res=%+v err=%v", res, err)
	}

	sole := map[string]json.RawMessage{"only": json.RawMessage(`{}`)}
	res, err = resolveAngularProjectKey(sole, "missing-alias")
	if err != nil || res.key != "only" || !res.usedSoleProjectFallback {
		t.Fatalf("sole project: res=%+v err=%v", res, err)
	}

	if _, err := resolveAngularProjectKey(projects, "missing"); err == nil {
		t.Fatal("expected ambiguity error")
	}
}

func TestApplyDevDocumentHeaders_SoleProjectFallbackNotice(t *testing.T) {
	dir := t.TempDir()
	writeManifestAtPath(t, filepath.Join(dir, manifestFileName), testManifestJSON)
	writeAngularFixture(t, dir, "legacy-project-name", false)

	var errOut bytes.Buffer
	headers := devHeaders(sampleCSP, samplePP, true, true)
	if err := applyDevDocumentHeaders(filepath.Join(dir, manifestFileName), "access-request-plugin", headers, &errOut); err != nil {
		t.Fatalf("apply: %v", err)
	}

	msg := errOut.String()
	if !strings.Contains(msg, `manifest alias "access-request-plugin" does not match the sole Angular project "legacy-project-name"`) {
		t.Fatalf("expected sole-project fallback note, got: %s", msg)
	}
	if !strings.Contains(msg, "headers were applied to that project anyway") {
		t.Fatalf("expected fallback confirmation, got: %s", msg)
	}
	if !strings.Contains(msg, "Restart ng serve") {
		t.Fatalf("expected restart note, got: %s", msg)
	}

	raw, err := os.ReadFile(filepath.Join(dir, angularManifestFileName))
	if err != nil {
		t.Fatalf("read angular.json: %v", err)
	}
	if !strings.Contains(string(raw), sampleCSP) {
		t.Fatalf("expected headers on sole project, got: %s", raw)
	}
}

func TestPatchAngularDevServerHeaders(t *testing.T) {
	raw := []byte(`{
  "projects": {
    "acme-plugin": {
      "architect": {
        "serve": {
          "options": {
            "ssl": true
          }
        }
      }
    }
  }
}`)

	updated, changed, err := patchAngularDevServerHeaders(raw, "acme-plugin", devHeaders(sampleCSP, samplePP, true, true))
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true on first patch")
	}

	var root map[string]any
	if err := json.Unmarshal(updated, &root); err != nil {
		t.Fatalf("updated JSON invalid: %v", err)
	}
	projects := root["projects"].(map[string]any)
	project := projects["acme-plugin"].(map[string]any)
	serve := project["architect"].(map[string]any)["serve"].(map[string]any)
	options := serve["options"].(map[string]any)
	if options["ssl"] != true {
		t.Fatal("ssl option should be preserved")
	}
	headers := options["headers"].(map[string]any)
	if headers[headerContentSecurityPolicy] != sampleCSP {
		t.Fatalf("CSP not written: %v", headers)
	}
	if headers[headerPermissionsPolicy] != samplePP {
		t.Fatalf("PP not written: %v", headers)
	}

	_, changed, err = patchAngularDevServerHeaders(updated, "acme-plugin", devHeaders(sampleCSP, samplePP, true, true))
	if err != nil {
		t.Fatalf("second patch: %v", err)
	}
	if changed {
		t.Fatal("expected changed=false when headers unchanged")
	}
}

func TestPatchAngularDevServerHeaders_EmptyPermissionsPolicy(t *testing.T) {
	raw := []byte(`{
  "projects": {
    "acme-plugin": {
      "architect": {
        "serve": {
          "options": {
            "headers": {}
          }
        }
      }
    }
  }
}`)

	updated, changed, err := patchAngularDevServerHeaders(raw, "acme-plugin", devHeaders(sampleCSP, "", true, true))
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}

	var root map[string]any
	if err := json.Unmarshal(updated, &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	headers := root["projects"].(map[string]any)["acme-plugin"].(map[string]any)["architect"].(map[string]any)["serve"].(map[string]any)["options"].(map[string]any)["headers"].(map[string]any)
	if headers[headerContentSecurityPolicy] != sampleCSP {
		t.Fatalf("CSP = %v", headers[headerContentSecurityPolicy])
	}
	pp, ok := headers[headerPermissionsPolicy].(string)
	if !ok {
		t.Fatalf("Permissions-Policy key missing: %v", headers)
	}
	if pp != "" {
		t.Fatalf("Permissions-Policy = %q, want empty string", pp)
	}
}

func TestPatchAngularDevServerHeaders_CSPOnlyLeavesPermissionsPolicyUntouched(t *testing.T) {
	raw := []byte(`{
  "projects": {
    "acme-plugin": {
      "architect": {
        "serve": {
          "options": {
            "headers": {
              "Permissions-Policy": "camera=()"
            }
          }
        }
      }
    }
  }
}`)

	updated, changed, err := patchAngularDevServerHeaders(raw, "acme-plugin", devHeaders(sampleCSP, "", true, false))
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	if !changed {
		t.Fatal("expected CSP change")
	}

	headers := jsonPath(t, updated, "projects", "acme-plugin", "architect", "serve", "options", "headers")
	if headers[headerContentSecurityPolicy] != sampleCSP {
		t.Fatalf("CSP = %v", headers[headerContentSecurityPolicy])
	}
	if headers[headerPermissionsPolicy] != "camera=()" {
		t.Fatalf("existing PP should be preserved, got %v", headers[headerPermissionsPolicy])
	}
}

func TestApplyDevDocumentHeaders_WritesAngularJSON(t *testing.T) {
	dir := t.TempDir()
	writeManifestAtPath(t, filepath.Join(dir, manifestFileName), testManifestJSON)
	writeAngularFixture(t, dir, "access-request-plugin", false)

	var errOut bytes.Buffer
	headers := devHeaders(sampleCSP, samplePP, true, true)
	if err := applyDevDocumentHeaders(filepath.Join(dir, manifestFileName), "access-request-plugin", headers, &errOut); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !strings.Contains(errOut.String(), "Restart ng serve") {
		t.Fatalf("expected restart note, got: %s", errOut.String())
	}

	raw, err := os.ReadFile(filepath.Join(dir, angularManifestFileName))
	if err != nil {
		t.Fatalf("read angular.json: %v", err)
	}
	if !strings.Contains(string(raw), sampleCSP) {
		t.Fatalf("angular.json missing CSP: %s", raw)
	}
}

func TestApplyDevDocumentHeaders_SkipsMissingAngularJSON(t *testing.T) {
	dir := t.TempDir()
	writeManifestAtPath(t, filepath.Join(dir, manifestFileName), testManifestJSON)

	var errOut bytes.Buffer
	err := applyDevDocumentHeaders(filepath.Join(dir, manifestFileName), "access-request-plugin", devHeaders(sampleCSP, "", true, false), &errOut)
	if err != nil {
		t.Fatalf("expected non-fatal skip, got: %v", err)
	}
	if !strings.Contains(errOut.String(), "not found") {
		t.Fatalf("expected skip note, got: %s", errOut.String())
	}
}

func jsonPath(t *testing.T, raw []byte, keys ...string) map[string]any {
	t.Helper()
	var cur any
	if err := json.Unmarshal(raw, &cur); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			t.Fatalf("expected object at %q", key)
		}
		cur = m[key]
	}
	m, ok := cur.(map[string]any)
	if !ok {
		t.Fatalf("expected map at end of path")
	}
	return m
}

func writeAngularFixture(t *testing.T, dir, project string, withHeaders bool) {
	t.Helper()
	options := `"ssl": true`
	if withHeaders {
		options += `, "headers": {"Content-Security-Policy": "old-csp"}`
	}
	content := `{
  "projects": {
    "` + project + `": {
      "architect": {
        "serve": {
          "options": {
            ` + options + `
          }
        }
      }
    }
  }
}`
	path := filepath.Join(dir, angularManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write angular.json: %v", err)
	}
}

func writeMultiProjectAngularFixture(t *testing.T, dir string) {
	t.Helper()
	content := `{
  "projects": {
    "other-plugin": {
      "architect": {
        "serve": {
          "options": {
            "ssl": true
          }
        }
      }
    },
    "another-plugin": {
      "architect": {
        "serve": {
          "options": {
            "ssl": true
          }
        }
      }
    }
  }
}`
	path := filepath.Join(dir, angularManifestFileName)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write angular.json: %v", err)
	}
}
