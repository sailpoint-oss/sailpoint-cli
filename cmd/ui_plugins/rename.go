package ui_plugins

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed rename.md
var renameHelp string

// renameNoChangesMsg is printed when neither the name nor the alias actually
// changes, so the command is a no-op that never touches the file or UMS.
const renameNoChangesMsg = "No changes: the name and alias are unchanged. Nothing to update."

// renameOptions holds the resolved flag values for the rename command. The *Set
// fields record whether the flag was explicitly passed, which drives both
// whether to prompt and whether an alias rejection re-prompts or fails fast.
type renameOptions struct {
	name     string
	alias    string
	nameSet  bool
	aliasSet bool
}

// renameDeps carries the command's client and I/O so the orchestration is
// testable with a fake client and buffers.
type renameDeps struct {
	client client.Client
	in     io.Reader
	out    io.Writer
	errOut io.Writer
}

func newRenameCommand() *cobra.Command {
	var opts renameOptions

	help := util.ParseHelp(renameHelp)
	cmd := &cobra.Command{
		Use:     "rename",
		Short:   "Rename a UI plugin's display name and/or alias",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.nameSet = cmd.Flags().Changed("name")
			opts.aliasSet = cmd.Flags().Changed("alias")

			spClient, err := newPluginClient()
			if err != nil {
				return err
			}

			deps := renameDeps{
				client: spClient,
				in:     cmd.InOrStdin(),
				out:    cmd.OutOrStdout(),
				errOut: cmd.ErrOrStderr(),
			}
			return runRename(context.Background(), deps, manifestFileName, opts)
		},
	}

	cmd.Flags().StringVar(&opts.name, "name", "", "New plugin display name (updates manifest.name.en)")
	cmd.Flags().StringVar(&opts.alias, "alias", "", "New tenant-unique plugin alias (stored lowercase)")

	return cmd
}

// runRename resolves the new name/alias (from flags or interactive prompts that
// default to the current values), short-circuits when nothing changed, then
// reconciles the workspace manifest with the plugin instance in the tenant.
//
// The current alias is resolved first. If the plugin does not exist yet (404)
// the file is written and the author is pointed at `create`; if it exists (200)
// a changed alias is validated, the instance is PATCHed, and only then is the
// file written — so a failed PATCH leaves the workspace untouched.
func runRename(ctx context.Context, deps renameDeps, manifestPath string, opts renameOptions) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	r := bufio.NewReader(deps.in)
	currentAlias := strings.TrimSpace(cfg.Manifest.Alias)
	currentName := strings.TrimSpace(localizedName(cfg.Manifest.Name))

	newName, err := resolveRenameName(r, deps, opts, currentName)
	if err != nil {
		return err
	}
	newAlias, err := resolveRenameAlias(r, deps, opts, currentAlias)
	if err != nil {
		return err
	}

	nameChanged := newName != currentName
	aliasChanged := newAlias != normalizeAliasInput(currentAlias)

	if !nameChanged && !aliasChanged {
		fmt.Fprintln(deps.out, renameNoChangesMsg)
		return nil
	}

	instance, notFound, err := resolvePluginForRename(ctx, deps.client, currentAlias)
	if err != nil {
		// 409 (ambiguous), any other non-200, or a transport error: stop before
		// writing anything.
		return err
	}

	if aliasChanged {
		finalAlias, stillChanged, verr := validateAliasForRename(ctx, deps, r, currentAlias, newAlias, opts.aliasSet)
		if verr != nil {
			return verr
		}
		newAlias, aliasChanged = finalAlias, stillChanged
	}

	// Re-prompting an alias back to the current value cancels the alias change,
	// which can leave nothing to do.
	if !nameChanged && !aliasChanged {
		fmt.Fprintln(deps.out, renameNoChangesMsg)
		return nil
	}

	if notFound {
		if err := writeRenamedManifest(manifestPath, newName, nameChanged, newAlias, aliasChanged); err != nil {
			return err
		}
		fmt.Fprintf(deps.out, "Updated %s. The plugin has not been created in this tenant yet.\n", manifestFileName)
		fmt.Fprint(deps.out, "\nNext steps:\n  sail ui-plugins create\n")
		return nil
	}

	effectiveAlias := instance.Alias
	if aliasChanged {
		effectiveAlias = newAlias
	}

	if err := patchRename(ctx, deps.client, instance, newName, nameChanged, newAlias, aliasChanged, effectiveAlias); err != nil {
		return err
	}
	if err := writeRenamedManifest(manifestPath, newName, nameChanged, newAlias, aliasChanged); err != nil {
		return fmt.Errorf("the plugin instance was updated in the tenant, but writing %s failed: %w", manifestFileName, err)
	}

	fmt.Fprintf(deps.out, "Renamed plugin instance %s (alias: %s)\nUpdated %s\n", instance.PluginInstanceID, effectiveAlias, manifestFileName)
	if viewURL := pluginViewURL(config.GetTenantUrl(), instance.PluginInstanceID); viewURL != "" {
		fmt.Fprintf(deps.out, "The plugin can be viewed at:\n%s\n", viewURL)
	}
	if aliasChanged {
		fmt.Fprint(deps.out, "\nNext steps:\n  sail ui-plugins link   # run again if you already linked, to get the new dev URL\n")
	}
	return nil
}

// resolveRenameName returns the new display name: the --name value when passed,
// otherwise an interactive prompt defaulting to the current name (Enter keeps
// it). An explicitly empty --name is rejected.
func resolveRenameName(r *bufio.Reader, deps renameDeps, opts renameOptions, currentName string) (string, error) {
	if opts.nameSet {
		name := strings.TrimSpace(opts.name)
		if name == "" {
			return "", fmt.Errorf("plugin name cannot be empty")
		}
		return name, nil
	}
	v, err := promptLine(r, deps.out, "Plugin Name", currentName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(v), nil
}

// resolveRenameAlias returns the new alias (trimmed + lowercased): the --alias
// value when passed, otherwise an interactive prompt defaulting to the current
// alias (Enter keeps it). An explicitly empty --alias is rejected.
func resolveRenameAlias(r *bufio.Reader, deps renameDeps, opts renameOptions, currentAlias string) (string, error) {
	if opts.aliasSet {
		alias := normalizeAliasInput(opts.alias)
		if alias == "" {
			return "", fmt.Errorf("plugin alias cannot be empty")
		}
		return alias, nil
	}
	v, err := promptLine(r, deps.out, "Plugin Alias", currentAlias)
	if err != nil {
		return "", err
	}
	return normalizeAliasInput(v), nil
}

// resolvePluginForRename resolves the current alias to its plugin instance,
// distinguishing the states the rename flow branches on: notFound (404, not yet
// created), a resolved instance (2xx), or an error for 409 (ambiguous alias,
// reported with the conflicting ids) and any other non-2xx or transport failure.
func resolvePluginForRename(ctx context.Context, c client.Client, alias string) (inst *pluginInstance, notFound bool, err error) {
	url := resolveAliasEndpoint + "?alias=" + neturl.QueryEscape(alias)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		return nil, false, fmt.Errorf("failed to resolve alias %q: %w", alias, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response: %w", err)
	}

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, true, nil
	case resp.StatusCode == http.StatusConflict:
		return nil, false, ambiguousAliasError(alias, body)
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		var pi pluginInstance
		if err := json.Unmarshal(body, &pi); err != nil {
			return nil, false, fmt.Errorf("failed to parse plugin instance: %w", err)
		}
		return &pi, false, nil
	default:
		return nil, false, mapUMSLookupError(resp.StatusCode, body, alias)
	}
}

// validateAliasForRename validates a changed alias against UMS. A definitive
// rejection — invalid (400) or already taken (409) — re-prompts for a
// replacement when the alias came from an interactive prompt, up to
// maxAliasPrompts; an alias supplied via --alias fails immediately. Entering the
// current alias at a re-prompt cancels the alias change (returns changed=false).
// Inconclusive checks are surfaced as warnings by checkAlias and accepted.
func validateAliasForRename(ctx context.Context, deps renameDeps, r *bufio.Reader, currentAlias, alias string, aliasFromFlag bool) (string, bool, error) {
	currentNorm := normalizeAliasInput(currentAlias)
	for attempt := 0; attempt < maxAliasPrompts; attempt++ {
		err := checkAlias(ctx, deps.client, deps.errOut, alias)
		if err == nil {
			return alias, true, nil
		}
		if aliasFromFlag {
			return "", false, err
		}
		fmt.Fprintf(deps.out, "%v — please choose a different alias.\n", err)
		v, perr := promptLine(r, deps.out, "Plugin Alias", currentAlias)
		if perr != nil {
			return "", false, perr
		}
		next := normalizeAliasInput(v)
		if next == currentNorm {
			return currentAlias, false, nil
		}
		alias = next
	}
	return "", false, fmt.Errorf("no valid alias provided after %d attempts", maxAliasPrompts)
}

// patchRename sends only the changed fields to UMS. The name payload starts from
// the resolved instance's name map so other locales survive — UMS replaces the
// whole map — with the English entry set to the new name.
func patchRename(ctx context.Context, c client.Client, instance *pluginInstance, newName string, nameChanged bool, newAlias string, aliasChanged bool, aliasForMsg string) error {
	payload := map[string]any{}
	if aliasChanged {
		payload["alias"] = newAlias
	}
	if nameChanged {
		nameMap := make(map[string]string, len(instance.Name)+1)
		for k, v := range instance.Name {
			nameMap[k] = v
		}
		nameMap["en"] = newName
		payload["name"] = nameMap
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode rename payload: %w", err)
	}

	// Patch does not default a Content-Type, so the caller must set it or UMS
	// receives an empty body and silently applies no changes.
	headers := uiPluginRequestHeaders()
	headers["Content-Type"] = "application/json"

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(instance.PluginInstanceID)
	resp, err := c.Patch(ctx, url, bytes.NewReader(data), headers)
	if err != nil {
		return fmt.Errorf("failed to update plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSUpdateError(resp.StatusCode, body, aliasForMsg)
	}
	return nil
}

// writeRenamedManifest edits only manifest.alias and manifest.name.en in place,
// leaving every other field untouched. It operates on the generic JSON tree
// (not the typed model) so fields the typed model would normalize — an omitted
// state, apiScopes ordering — are preserved verbatim in value.
func writeRenamedManifest(path, newName string, nameChanged bool, newAlias string, aliasChanged bool) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("unable to read %s: %w", path, err)
	}

	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("invalid %s: %w", path, err)
	}

	manifest, ok := root["manifest"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid %s: manifest section is missing or malformed", path)
	}

	if aliasChanged {
		manifest["alias"] = newAlias
	}
	if nameChanged {
		nameMap, ok := manifest["name"].(map[string]any)
		if !ok {
			nameMap = map[string]any{}
			manifest["name"] = nameMap
		}
		nameMap["en"] = newName
	}

	data, err := json.MarshalIndent(root, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", manifestFileName, err)
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data)
}

// normalizeAliasInput canonicalizes alias input to the form UMS stores: trimmed
// and lowercased. Comparisons and the persisted value both use this form.
func normalizeAliasInput(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
