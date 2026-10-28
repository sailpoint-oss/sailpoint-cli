package ui_plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	manifestFileName  = "sp-ui-plugin.json"
	supportedVersion1 = 1
)

// iframeSandbox tokens are lowercase ASCII words of letters and digits joined by
// single hyphens (e.g. allow-popups, allow-popups-to-escape-sandbox). The pattern
// rejects uppercase, underscores, empty strings, and leading/trailing/double
// hyphens. Membership in the accepted set is enforced by UMS, not here, so a new
// token needs only a UMS change.
var iframeSandboxTokenPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// loadAndValidateWorkspaceManifest is the shared full-validation entrypoint for
// ui-plugins commands. It performs strict parse + semantic validation.
func loadAndValidateWorkspaceManifest(path string) (*uiPluginWorkspaceConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", path, err)
	}

	cfg, err := parseWorkspaceManifestStrict(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}

	if err := validateWorkspaceManifest(cfg); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}

	return cfg, nil
}

func parseWorkspaceManifestStrict(raw []byte) (*uiPluginWorkspaceConfig, error) {
	var cfg uiPluginWorkspaceConfig

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}

	// Enforce a single JSON object in file.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("manifest must contain a single JSON object")
	}

	if cfg.Manifest.State == nil {
		defaultState := stateEnabled
		cfg.Manifest.State = &defaultState
	}

	return &cfg, nil
}

func validateWorkspaceManifest(cfg *uiPluginWorkspaceConfig) error {
	if cfg.Version == 0 {
		return fmt.Errorf("version is required")
	}
	if cfg.Version != supportedVersion1 {
		return fmt.Errorf("unsupported version %d (supported: %d)", cfg.Version, supportedVersion1)
	}

	manifest := cfg.Manifest
	if strings.TrimSpace(manifest.Alias) == "" {
		return fmt.Errorf("manifest.alias is required")
	}
	if len(manifest.Name) == 0 {
		return fmt.Errorf("manifest.name is required and must contain at least one locale entry")
	}
	if len(manifest.Description) == 0 {
		return fmt.Errorf("manifest.description is required and must contain at least one locale entry")
	}
	if len(manifest.Slots) == 0 {
		return fmt.Errorf("manifest.slots is required and must contain at least one slot")
	}
	for i, slot := range manifest.Slots {
		if strings.TrimSpace(slot.SlotID) == "" {
			return fmt.Errorf("manifest.slots[%d].slotId is required", i)
		}
	}

	// The backend requires these policy objects to be present; an empty object
	// {} is accepted. Directive names/values are validated by UMS, not here.
	if manifest.ContentSecurityPolicies == nil {
		return fmt.Errorf("manifest.contentSecurityPolicies is required (use an empty object {} if the plugin declares no CSP directives)")
	}
	if manifest.PermissionPolicy == nil {
		return fmt.Errorf("manifest.permissionPolicy is required (use an empty object {} if the plugin declares no permission-policy directives)")
	}
	if manifest.IframeAllow == nil {
		return fmt.Errorf("manifest.iframeAllow is required (use an empty object {} if the plugin declares no iframe-allow directives)")
	}

	// iframeSandbox is optional: a nil pointer (absent) and an empty array both
	// carry no tokens. When present, every token must be well-formed; the CLI does
	// not check membership (UMS does).
	if manifest.IframeSandbox != nil {
		for i, token := range *manifest.IframeSandbox {
			if !iframeSandboxTokenPattern.MatchString(token) {
				return fmt.Errorf("manifest.iframeSandbox[%d] %q is not a valid sandbox token (lowercase letters, digits, and hyphens, with a hyphen only between characters)", i, token)
			}
		}
	}

	if cfg.Build != nil {
		if cfg.Build.Port != nil && *cfg.Build.Port <= 0 {
			return fmt.Errorf("build.port must be greater than 0")
		}
		if strings.TrimSpace(cfg.Build.OutDir) == "" && cfg.Build.OutDir != "" {
			return fmt.Errorf("build.outDir must not be empty when provided")
		}
	}

	return nil
}
