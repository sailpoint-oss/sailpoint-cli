package ui_plugins

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

// sentinelName is the placeholder token the Angular starter template ships with.
// init substitutes it for the plugin's alias (in identifier files) or display
// name (in human-facing content) during personalization.
const sentinelName = "starter"

// slugifyAlias derives a filesystem/URL-safe alias suggestion from the plugin
// name: lowercased, with runs of non-alphanumeric characters collapsed to a
// single dash and leading/trailing dashes trimmed. It is only a suggestion —
// UMS is the authority on alias validity (length, uniqueness, etc.).
func slugifyAlias(name string) string {
	var b strings.Builder
	prevDash := false
	for _, ch := range strings.ToLower(name) {
		switch {
		case (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9'):
			b.WriteRune(ch)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// generateWorkspaceManifest builds a workspace config with the documented
// defaults for the --path flow.
//
// Defaults are built from the typed model here rather than fetched from
// ui-plugin-templates on purpose: the manifest schema and validator must live
// in this package anyway (for local validation and the UMS payload), so
// generating from the model keeps the written file in lockstep with what the
// CLI accepts. Additionally, besides minimal defaults, this is constructed
// using user inputs
func generateWorkspaceManifest(alias, name, outDir string, port int) *uiPluginWorkspaceConfig {
	p := port
	return &uiPluginWorkspaceConfig{
		Version: supportedVersion1,
		Manifest: uiPluginManifest{
			Alias:                   alias,
			Name:                    map[string]string{"en": name},
			Description:             map[string]string{"en": name},
			APIScopes:               []string{"sp:scopes:all"},
			ContentSecurityPolicies: map[string][]string{},
			PermissionPolicy:        map[string][]string{},
			IframeAllow:             map[string][]string{},
			Slots:                   []uiPluginSlot{{SlotID: "full-page"}},
		},
		Build: &uiPluginBuildConfig{OutDir: outDir, Port: &p},
	}
}

// writeWorkspaceManifest validates cfg with the shared validator, then writes it
// as indented JSON to path.
func writeWorkspaceManifest(path string, cfg *uiPluginWorkspaceConfig) error {
	if err := validateWorkspaceManifest(cfg); err != nil {
		return fmt.Errorf("generated %s is invalid: %w", manifestFileName, err)
	}
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", manifestFileName, err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// personalizeScaffold rewrites the freshly extracted starter in place: the
// build/identity files adopt the alias, the human-facing display strings adopt
// the plugin name. Identifier substitution is scoped to the three files below —
// a blanket text replace across the tree is intentionally avoided.
func personalizeScaffold(destDir, alias, name string) error {
	if err := renameAngularProject(destDir, alias); err != nil {
		return err
	}
	if err := setPackageName(destDir, alias); err != nil {
		return err
	}
	if err := personalizeManifestFile(destDir, alias, name); err != nil {
		return err
	}
	if err := personalizeDisplayContent(destDir, name); err != nil {
		return err
	}
	return assertNoSentinel(destDir, alias)
}

// renameAngularProject renames the Angular project in angular.json and rewrites
// its build target references to the new alias.
func renameAngularProject(destDir, alias string) error {
	path := filepath.Join(destDir, "angular.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// The Angular project name is the sole quoted "starter" token (the project
	// key under "projects"); buildTarget values keep the sentinel followed by a
	// colon and are handled separately below.
	projectKey := `"` + sentinelName + `"`
	if n := strings.Count(content, projectKey); n != 1 {
		return fmt.Errorf("expected exactly one Angular project key %s in angular.json, found %d", projectKey, n)
	}
	content = strings.Replace(content, projectKey, `"`+alias+`"`, 1)

	// buildTarget values use Angular's stable "project:target:configuration"
	// specifier grammar (the same form `ng run` consumes). Split on ':' and the
	// leading segment is always the project name, so swapping the "starter:"
	// prefix renames the reference regardless of the target/configuration names
	// — robust across Angular versions and independent of angular.json layout.
	content = strings.ReplaceAll(content, `"`+sentinelName+`:`, `"`+alias+`:`)

	return os.WriteFile(path, []byte(content), 0644)
}

func setPackageName(destDir, alias string) error {
	return replaceExactlyOnce(
		filepath.Join(destDir, "package.json"),
		`"name": "`+sentinelName+`"`,
		`"name": "`+alias+`"`,
	)
}

// personalizeManifestFile rewrites the scaffolded sp-ui-plugin.json with the
// plugin's identity and alias-derived output path, re-validating before writing.
func personalizeManifestFile(destDir, alias, name string) error {
	path := filepath.Join(destDir, manifestFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	cfg, err := parseWorkspaceManifestStrict(raw)
	if err != nil {
		return fmt.Errorf("invalid %s in template: %w", manifestFileName, err)
	}
	cfg.Manifest.Alias = alias
	cfg.Manifest.Name = map[string]string{"en": name}
	cfg.Manifest.Description = map[string]string{"en": name}
	if cfg.Build == nil {
		cfg.Build = &uiPluginBuildConfig{}
	}
	cfg.Build.OutDir = "./dist/" + alias + "/browser"
	return writeWorkspaceManifest(path, cfg)
}

// personalizeDisplayContent updates the human-facing title strings to the plugin
// name. app.ts and app.spec.ts are kept in lockstep so the scaffold's unit test
// still passes; the name is escaped for each file's syntax.
func personalizeDisplayContent(destDir, name string) error {
	tsName := escapeSingleQuoted(name)
	if err := replaceExactlyOnce(
		filepath.Join(destDir, "src", "app", "app.ts"),
		"signal('"+sentinelName+"')",
		"signal('"+tsName+"')",
	); err != nil {
		return err
	}
	if err := replaceExactlyOnce(
		filepath.Join(destDir, "src", "app", "app.spec.ts"),
		"Hello, "+sentinelName,
		"Hello, "+tsName,
	); err != nil {
		return err
	}
	// index.html carries the capitalized "Starter" literal in its <title>.
	return replaceExactlyOnce(
		filepath.Join(destDir, "src", "index.html"),
		"<title>Starter</title>",
		"<title>"+html.EscapeString(name)+"</title>",
	)
}

// assertNoSentinel guards against a missed identifier substitution by checking
// the specific sentinel patterns that must have been rewritten. It intentionally
// does not scan whole files — the display name legitimately may contain the
// word — and is a no-op when the alias itself equals the sentinel.
func assertNoSentinel(destDir, alias string) error {
	if alias == sentinelName {
		return nil
	}
	checks := []struct{ file, pattern string }{
		{"angular.json", `"` + sentinelName + `"`},
		{"angular.json", `"` + sentinelName + `:`},
		{"package.json", `"name": "` + sentinelName + `"`},
		{manifestFileName, "dist/" + sentinelName + "/"},
	}
	for _, c := range checks {
		data, err := os.ReadFile(filepath.Join(destDir, c.file))
		if err != nil {
			return err
		}
		if strings.Contains(string(data), c.pattern) {
			return fmt.Errorf("unexpected %q remaining in %s after personalization", c.pattern, c.file)
		}
	}
	return nil
}

// replaceExactlyOnce replaces the single expected occurrence of old with new in
// the file at path, erroring if it is not found exactly once (a guard against
// template drift).
func replaceExactlyOnce(path, old, new string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if n := strings.Count(content, old); n != 1 {
		return fmt.Errorf("expected exactly one %q in %s, found %d", old, filepath.Base(path), n)
	}
	content = strings.Replace(content, old, new, 1)
	return os.WriteFile(path, []byte(content), 0644)
}

// escapeSingleQuoted escapes a value for embedding inside a single-quoted
// TypeScript/JavaScript string literal.
func escapeSingleQuoted(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}
