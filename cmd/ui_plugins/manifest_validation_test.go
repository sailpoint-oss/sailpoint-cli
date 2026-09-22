package ui_plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateWorkspaceManifest_Valid(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  },
  "build": {
    "outDir": "./dist",
    "port": 4200
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if cfg.Version != 1 {
		t.Fatalf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Manifest.Slots) != 1 || cfg.Manifest.Slots[0].SlotID != "full-page" {
		t.Fatalf("expected one full-page slot, got: %+v", cfg.Manifest.Slots)
	}
}

func TestLoadAndValidateWorkspaceManifest_ValidIframeAllow(t *testing.T) {
	path := writeManifestFixture(t, `{
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

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if cfg.Manifest.IframeAllow == nil {
		t.Fatal("expected iframeAllow map to be present")
	}
}

func TestLoadAndValidateWorkspaceManifest_ValidIframeAllowWithDirectives(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {"camera": ["'src'"]}
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if len(cfg.Manifest.IframeAllow) != 1 {
		t.Fatalf("expected one iframeAllow directive, got: %+v", cfg.Manifest.IframeAllow)
	}
	if got := cfg.Manifest.IframeAllow["camera"]; len(got) != 1 || got[0] != "'src'" {
		t.Fatalf("unexpected camera directive: %+v", got)
	}
}

func TestLoadAndValidateWorkspaceManifest_IframeSandboxRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		wantPresent bool
		wantValues  []string
	}{
		{name: "omitted"},
		{name: "null treated as omitted", field: `, "iframeSandbox": null`},
		{name: "empty array preserved", field: `, "iframeSandbox": []`, wantPresent: true, wantValues: []string{}},
		{
			name:        "opaque strings preserve order and duplicates",
			field:       `, "iframeSandbox": ["allow-popups", "b", "allow-popups", "", " "]`,
			wantPresent: true,
			wantValues:  []string{"allow-popups", "b", "allow-popups", "", " "},
		},
		{
			name:        "null element follows existing string slice behavior",
			field:       `, "iframeSandbox": ["x", null]`,
			wantPresent: true,
			wantValues:  []string{"x", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadAndValidateWorkspaceManifest(writeManifestFixture(t, workspaceManifestWithSandbox(tt.field)))
			if err != nil {
				t.Fatalf("expected valid iframeSandbox, got: %v", err)
			}

			if gotPresent := cfg.Manifest.IframeSandbox != nil; gotPresent != tt.wantPresent {
				t.Fatalf("iframeSandbox presence = %t, want %t", gotPresent, tt.wantPresent)
			}
			if tt.wantPresent && !equalStrings(*cfg.Manifest.IframeSandbox, tt.wantValues) {
				t.Fatalf("iframeSandbox = %#v, want %#v", *cfg.Manifest.IframeSandbox, tt.wantValues)
			}

			payload, err := json.Marshal(cfg.Manifest)
			if err != nil {
				t.Fatalf("marshal manifest: %v", err)
			}
			var decoded map[string]any
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatalf("decode marshaled manifest: %v", err)
			}
			_, gotPresent := decoded["iframeSandbox"]
			if gotPresent != tt.wantPresent {
				t.Fatalf("marshaled iframeSandbox presence = %t, want %t; payload: %s", gotPresent, tt.wantPresent, payload)
			}
		})
	}
}

func TestLoadAndValidateWorkspaceManifest_InvalidIframeSandboxRejected(t *testing.T) {
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{name: "string", field: `, "iframeSandbox": "allow-popups"`, want: "cannot unmarshal"},
		{name: "object", field: `, "iframeSandbox": {"allow-popups": true}`, want: "cannot unmarshal"},
		{name: "number element", field: `, "iframeSandbox": [1]`, want: "cannot unmarshal"},
		{name: "snake case key", field: `, "iframe_sandbox": []`, want: "unknown field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadAndValidateWorkspaceManifest(writeManifestFixture(t, workspaceManifestWithSandbox(tt.field)))
			if err == nil {
				t.Fatal("expected invalid iframeSandbox to fail")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got: %v", tt.want, err)
			}
		})
	}
}

func TestLoadAndValidateWorkspaceManifest_StringIframeAllowRejected(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "iframeAllow": "camera 'self'"
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected string iframeAllow to fail")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected unmarshal type error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_ValidSlotWithOptionalFields(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{
      "slotId": "full-page",
      "requiredCapabilities": ["ORG_ADMIN"],
      "restrictToUsers": ["2c9180827f9b911e017f9b9122340000"]
    }],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	slot := cfg.Manifest.Slots[0]
	if slot.SlotID != "full-page" {
		t.Fatalf("expected slotId full-page, got %q", slot.SlotID)
	}
	if len(slot.RequiredCapabilities) != 1 || slot.RequiredCapabilities[0] != "ORG_ADMIN" {
		t.Fatalf("unexpected requiredCapabilities: %+v", slot.RequiredCapabilities)
	}
	if len(slot.RestrictToUsers) != 1 {
		t.Fatalf("unexpected restrictToUsers: %+v", slot.RestrictToUsers)
	}
}

func TestLoadAndValidateWorkspaceManifest_LegacyStringSlotsRejected(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": ["full-page"]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected legacy string slots to fail")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected unmarshal type error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_MissingSlotID(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected missing slotId to fail")
	}
	if !strings.Contains(err.Error(), "manifest.slots[0].slotId is required") {
		t.Fatalf("expected slotId required error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_UnknownSlotField(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page", "unexpectedField": true}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected unknown slot field to fail")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_MissingVersion(t *testing.T) {
	path := writeManifestFixture(t, `{
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected missing version to fail")
	}
	if !strings.Contains(err.Error(), "version is required") {
		t.Fatalf("expected version error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_UnknownField(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "unexpectedField": true
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected unknown field to fail")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_TypeMismatch(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": 123,
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected type mismatch to fail")
	}
	if !strings.Contains(err.Error(), "cannot unmarshal") {
		t.Fatalf("expected unmarshal type error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_MissingRequiredManifestField(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected missing alias to fail")
	}
	if !strings.Contains(err.Error(), "manifest.alias is required") {
		t.Fatalf("expected alias required error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_MissingContentSecurityPolicies(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "permissionPolicy": {},
    "iframeAllow": {}
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected missing contentSecurityPolicies to fail")
	}
	if !strings.Contains(err.Error(), "manifest.contentSecurityPolicies is required") {
		t.Fatalf("expected contentSecurityPolicies required error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_MissingIframeAllow(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {}
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected missing iframeAllow to fail")
	}
	if !strings.Contains(err.Error(), "manifest.iframeAllow is required") {
		t.Fatalf("expected iframeAllow required error, got: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_UnsupportedVersion(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 2,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}]
  }
}`)

	_, err := loadAndValidateWorkspaceManifest(path)
	if err == nil {
		t.Fatal("expected unsupported version to fail")
	}
	if !strings.Contains(err.Error(), "unsupported version") {
		t.Fatalf("expected unsupported version error, got: %v", err)
	}
}

func writeManifestFixture(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, manifestFileName)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write manifest fixture: %v", err)
	}
	return path
}

func workspaceManifestWithSandbox(field string) string {
	return `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {"camera": ["'self'"]},
    "iframeAllow": {"camera": ["'src'"]}` + field + `
  }
}`
}
