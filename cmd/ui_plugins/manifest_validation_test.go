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
    "iframeAllow": {"camera": ["'self'"]}
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if len(cfg.Manifest.IframeAllow) != 1 {
		t.Fatalf("expected one iframeAllow directive, got: %+v", cfg.Manifest.IframeAllow)
	}
	if got := cfg.Manifest.IframeAllow["camera"]; len(got) != 1 || got[0] != "'self'" {
		t.Fatalf("unexpected camera directive: %+v", got)
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

func TestLoadAndValidateWorkspaceManifest_IframeSandboxAbsent(t *testing.T) {
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
	if cfg.Manifest.IframeSandbox != nil {
		t.Fatalf("expected absent iframeSandbox to decode as a nil pointer, got: %+v", *cfg.Manifest.IframeSandbox)
	}
}

func TestLoadAndValidateWorkspaceManifest_IframeSandboxEmpty(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {},
    "iframeSandbox": []
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if cfg.Manifest.IframeSandbox == nil {
		t.Fatal("expected a present empty iframeSandbox to decode as a non-nil pointer")
	}
	if len(*cfg.Manifest.IframeSandbox) != 0 {
		t.Fatalf("expected empty iframeSandbox, got: %+v", *cfg.Manifest.IframeSandbox)
	}
}

func TestLoadAndValidateWorkspaceManifest_IframeSandboxValidTokens(t *testing.T) {
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {},
    "iframeSandbox": ["allow-popups", "allow-popups-to-escape-sandbox"]
  }
}`)

	cfg, err := loadAndValidateWorkspaceManifest(path)
	if err != nil {
		t.Fatalf("expected valid manifest, got err: %v", err)
	}
	if cfg.Manifest.IframeSandbox == nil || len(*cfg.Manifest.IframeSandbox) != 2 {
		t.Fatalf("expected two iframeSandbox tokens, got: %+v", cfg.Manifest.IframeSandbox)
	}
}

func TestLoadAndValidateWorkspaceManifest_IframeSandboxUnknownTokenAccepted(t *testing.T) {
	// A well-formed but unknown token passes the CLI by design; UMS owns the
	// allowlist and rejects it at request time.
	path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {},
    "iframeSandbox": ["allow-forms"]
  }
}`)

	if _, err := loadAndValidateWorkspaceManifest(path); err != nil {
		t.Fatalf("expected well-formed unknown token to pass local validation, got err: %v", err)
	}
}

func TestLoadAndValidateWorkspaceManifest_IframeSandboxMalformedTokens(t *testing.T) {
	cases := map[string]string{
		"uppercase":       "Allow-Popups",
		"underscore":      "allow_popups",
		"leading hyphen":  "-allow",
		"trailing hyphen": "allow-",
		"double hyphen":   "allow--popups",
		"empty token":     "",
		"space":           "allow popups",
	}

	for name, token := range cases {
		t.Run(name, func(t *testing.T) {
			tokenJSON, err := json.Marshal(token)
			if err != nil {
				t.Fatalf("failed to encode token: %v", err)
			}
			path := writeManifestFixture(t, `{
  "version": 1,
  "manifest": {
    "alias": "access-request-plugin",
    "name": {"en-US": "Access Request"},
    "description": {"en-US": "Plugin description"},
    "slots": [{"slotId": "full-page"}],
    "contentSecurityPolicies": {},
    "permissionPolicy": {},
    "iframeAllow": {},
    "iframeSandbox": [`+string(tokenJSON)+`]
  }
}`)

			_, err = loadAndValidateWorkspaceManifest(path)
			if err == nil {
				t.Fatalf("expected malformed token %q to fail validation", token)
			}
			if !strings.Contains(err.Error(), "manifest.iframeSandbox[0]") {
				t.Fatalf("expected indexed iframeSandbox error, got: %v", err)
			}
		})
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
