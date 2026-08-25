package ui_plugins

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	headerContentSecurityPolicy = "Content-Security-Policy"
	headerPermissionsPolicy     = "Permissions-Policy"
	angularManifestFileName     = "angular.json"
)

// devDocumentHeaders holds the plugin-document headers the backend
// returns ephemerally on create/link for local dev CSP parity with
// CDN uploads. A set* flag records whether the backend included that
// key so absent keys are not written to angular.json.
type devDocumentHeaders struct {
	contentSecurityPolicy string
	permissionsPolicy     string
	setCSP                bool
	setPP                 bool
}

func (h devDocumentHeaders) present() bool {
	return h.setCSP || h.setPP
}

// parseDevDocumentHeaders extracts devDocumentHeaders from a successful
// backend create or link response body. The second return value reports
// whether the backend returned at least one known header key.
func parseDevDocumentHeaders(body []byte) (devDocumentHeaders, bool) {
	var parsed struct {
		DevDocumentHeaders map[string]string `json:"devDocumentHeaders"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return devDocumentHeaders{}, false
	}
	return headersFromMap(parsed.DevDocumentHeaders)
}

func headersFromMap(m map[string]string) (devDocumentHeaders, bool) {
	if len(m) == 0 {
		return devDocumentHeaders{}, false
	}

	var h devDocumentHeaders
	if v, ok := m[headerContentSecurityPolicy]; ok {
		h.contentSecurityPolicy = v
		h.setCSP = true
	}
	if v, ok := m[headerPermissionsPolicy]; ok {
		h.permissionsPolicy = v
		h.setPP = true
	}
	if !h.present() {
		return devDocumentHeaders{}, false
	}
	return h, true
}

// applyDevDocumentHeadersBestEffort runs applyDevDocumentHeaders after a successful
// create or link API call. Patching failures are reported to errOut and do not
// fail the command.
func applyDevDocumentHeadersBestEffort(manifestPath, alias string, headers devDocumentHeaders, errOut io.Writer) {
	if err := applyDevDocumentHeaders(manifestPath, alias, headers, errOut); err != nil {
		_, _ = fmt.Fprintf(errOut, "Warning: could not update %s dev server headers: %v\n", angularManifestFileName, err)
	}
}

// applyDevDocumentHeaders writes backend dev document headers into
// the workspace angular.json when present. Missing angular.json or
// empty headers are reported to errOut and treated as non-fatal.
func applyDevDocumentHeaders(manifestPath, alias string, headers devDocumentHeaders, errOut io.Writer) error {
	if !headers.present() {
		_, _ = fmt.Fprintln(errOut, "Note: the backend did not return dev document headers; angular.json was not updated.")
		return nil
	}

	workspaceDir := filepath.Dir(manifestPath)
	if workspaceDir == "." {
		var err error
		workspaceDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot resolve workspace directory: %w", err)
		}
	}

	angularPath := filepath.Join(workspaceDir, angularManifestFileName)
	if _, err := os.Stat(angularPath); err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintf(errOut, "Note: %s not found; skipped writing dev document headers (non-Angular workspace).\n", angularManifestFileName)
			return nil
		}
		return fmt.Errorf("cannot access %s: %w", angularManifestFileName, err)
	}

	raw, err := os.ReadFile(angularPath)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", angularManifestFileName, err)
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("invalid %s: %w", angularManifestFileName, err)
	}

	projectsRaw, ok := root["projects"]
	if !ok {
		return fmt.Errorf("%s has no projects section", angularManifestFileName)
	}

	var projects map[string]json.RawMessage
	if err := json.Unmarshal(projectsRaw, &projects); err != nil {
		return fmt.Errorf("invalid projects in %s: %w", angularManifestFileName, err)
	}

	resolution, err := resolveAngularProjectKey(projects, alias)
	if err != nil {
		return err
	}
	projectKey := resolution.key

	updated, changed, err := patchAngularDevServerHeaders(raw, projectKey, headers)
	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	if err := writeFileAtomic(angularPath, updated); err != nil {
		return fmt.Errorf("failed to write %s: %w", angularManifestFileName, err)
	}

	if resolution.usedSoleProjectFallback {
		_, _ = fmt.Fprintf(errOut, "Note: manifest alias %q does not match the sole Angular project %q in %s; dev server headers were applied to that project anyway.\n", alias, projectKey, angularManifestFileName)
	}

	_, _ = fmt.Fprintln(errOut, "Updated angular.json dev server headers. Restart ng serve / npm start for the changes to take effect.")
	return nil
}

type angularProjectResolution struct {
	key                     string
	usedSoleProjectFallback bool
}

// resolveAngularProjectKey picks the Angular project to patch: alias match first,
// then the sole project when unambiguous.
func resolveAngularProjectKey(projects map[string]json.RawMessage, alias string) (angularProjectResolution, error) {
	if alias != "" {
		if _, ok := projects[alias]; ok {
			return angularProjectResolution{key: alias}, nil
		}
	}

	switch len(projects) {
	case 0:
		return angularProjectResolution{}, fmt.Errorf("%s defines no Angular projects", angularManifestFileName)
	case 1:
		for name := range projects {
			res := angularProjectResolution{key: name}
			if alias != "" && name != alias {
				res.usedSoleProjectFallback = true
			}
			return res, nil
		}
	default:
		if alias != "" {
			return angularProjectResolution{}, fmt.Errorf("manifest alias %q does not match any project in %s; specify a matching alias or use a single-project workspace", alias, angularManifestFileName)
		}
		return angularProjectResolution{}, fmt.Errorf("%s defines multiple projects; cannot choose which project to patch", angularManifestFileName)
	}
	return angularProjectResolution{}, fmt.Errorf("cannot resolve Angular project in %s", angularManifestFileName)
}

// patchAngularDevServerHeaders updates serve.options.headers for projectKey and
// re-encodes the full angular.json document.
func patchAngularDevServerHeaders(raw []byte, projectKey string, headers devDocumentHeaders) ([]byte, bool, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, false, fmt.Errorf("invalid %s: %w", angularManifestFileName, err)
	}

	projects, ok := root["projects"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s has no projects section", angularManifestFileName)
	}

	project, ok := projects[projectKey].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("%s has no project %q", angularManifestFileName, projectKey)
	}

	architect, ok := project["architect"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("project %q in %s has no architect section", projectKey, angularManifestFileName)
	}

	serve, ok := architect["serve"].(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("project %q in %s has no architect.serve target", projectKey, angularManifestFileName)
	}

	options, ok := serve["options"].(map[string]any)
	if !ok {
		options = map[string]any{}
		serve["options"] = options
	}

	headersObj, ok := options["headers"].(map[string]any)
	if !ok {
		headersObj = map[string]any{}
		options["headers"] = headersObj
	}

	changed := false
	if headers.setCSP {
		cur, ok := headersObj[headerContentSecurityPolicy].(string)
		if !ok || cur != headers.contentSecurityPolicy {
			headersObj[headerContentSecurityPolicy] = headers.contentSecurityPolicy
			changed = true
		}
	}
	if headers.setPP {
		cur, ok := headersObj[headerPermissionsPolicy].(string)
		if !ok || cur != headers.permissionsPolicy {
			headersObj[headerPermissionsPolicy] = headers.permissionsPolicy
			changed = true
		}
	}

	if !changed {
		return nil, false, nil
	}

	updated, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("failed to encode %s: %w", angularManifestFileName, err)
	}
	return append(updated, '\n'), true, nil
}
