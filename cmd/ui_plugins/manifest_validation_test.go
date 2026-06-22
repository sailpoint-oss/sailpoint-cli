// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
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
    "slots": [{"slotId": "full-page"}]
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
    }]
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

