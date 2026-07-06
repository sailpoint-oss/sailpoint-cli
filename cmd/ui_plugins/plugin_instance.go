package ui_plugins

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
)

const (
	// resolveAliasEndpoint returns the full plugin instance for a tenant-unique alias.
	// It is distinct from validateAliasEndpoint (used by create), which only checks availability.
	resolveAliasEndpoint = pluginInstancesEndpoint + "/resolve-alias"
	// pluginInstancesPageSize is the UMS maximum page size for the list endpoint.
	pluginInstancesPageSize = 250
	// tableColumnMaxWidth caps the width of author-entered columns (alias, name) in
	// the list table so long values don't blow out the layout. Full values are always
	// available via --json.
	tableColumnMaxWidth = 40
)

// pluginInstance is the subset of the UMS PluginInstanceDto the list and delete
// commands need for display. Unrecognized fields are ignored on decode; the raw
// body is preserved separately when --json fidelity is required.
type pluginInstance struct {
	PluginInstanceID    string            `json:"pluginInstanceId"`
	Alias               string            `json:"alias"`
	Name                map[string]string `json:"name"`
	Created             string            `json:"created"`
	Slots               []uiPluginSlot    `json:"slots"`
	ActiveAssetBundleID *string           `json:"activeAssetBundleId"`
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// looksLikeUUID reports whether s has the canonical 8-4-4-4-12 UUID shape. A UUID
// is also a syntactically valid alias, so an alias that happens to look like a UUID
// is treated as a plugin ID; this is harmless (it resolves to "not found" rather
// than deleting the wrong instance) and is documented in delete.md.
func looksLikeUUID(s string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(s))
}

// listPluginInstances fetches every plugin instance for the active tenant, paging
// through the UMS list endpoint until a short page is returned. Items are kept as
// raw JSON so --json can emit them without field loss.
func listPluginInstances(ctx context.Context, c client.Client) ([]json.RawMessage, error) {
	all := []json.RawMessage{}
	headers := uiPluginRequestHeaders()

	for offset := 0; ; offset += pluginInstancesPageSize {
		url := fmt.Sprintf("%s?limit=%d&offset=%d", pluginInstancesEndpoint, pluginInstancesPageSize, offset)
		resp, err := c.Get(ctx, url, headers)
		if err != nil {
			return nil, fmt.Errorf("failed to list plugin instances: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			return nil, mapUMSListError(resp.StatusCode, body)
		}

		var page []json.RawMessage
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("failed to parse plugin instance list: %w", err)
		}
		all = append(all, page...)

		if len(page) < pluginInstancesPageSize {
			break
		}
	}

	return all, nil
}

// resolveDeleteTarget resolves a delete argument to a plugin instance, choosing the
// lookup by the argument's shape: a UUID is looked up by ID, anything else by alias.
// It returns the typed instance and its raw response body (for --json).
func resolveDeleteTarget(ctx context.Context, c client.Client, arg string) (*pluginInstance, []byte, error) {
	if looksLikeUUID(arg) {
		return getPluginInstanceByID(ctx, c, arg)
	}
	return resolvePluginInstanceByAlias(ctx, c, arg)
}

// getPluginInstanceByID fetches a single instance by its UUID.
func getPluginInstanceByID(ctx context.Context, c client.Client, id string) (*pluginInstance, []byte, error) {
	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(id)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, mapUMSLookupError(resp.StatusCode, body, id)
	}

	var inst pluginInstance
	if err := json.Unmarshal(body, &inst); err != nil {
		return nil, nil, fmt.Errorf("failed to parse plugin instance: %w", err)
	}
	return &inst, body, nil
}

// resolvePluginInstanceByAlias resolves a tenant-unique alias to its instance. An
// ambiguous alias (409) is reported with the conflicting plugin IDs and never
// auto-resolved — the caller must re-run with a specific plugin ID.
func resolvePluginInstanceByAlias(ctx context.Context, c client.Client, alias string) (*pluginInstance, []byte, error) {
	url := resolveAliasEndpoint + "?alias=" + neturl.QueryEscape(alias)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve alias: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusConflict {
		return nil, nil, ambiguousAliasError(alias, body)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, mapUMSLookupError(resp.StatusCode, body, alias)
	}

	var inst pluginInstance
	if err := json.Unmarshal(body, &inst); err != nil {
		return nil, nil, fmt.Errorf("failed to parse plugin instance: %w", err)
	}
	return &inst, body, nil
}

// ambiguousAliasError builds an actionable error listing the plugin IDs an alias
// resolves to, so the user can re-run delete with a specific ID.
func ambiguousAliasError(alias string, body []byte) error {
	var parsed struct {
		Conflicts []pluginInstance `json:"conflicts"`
	}
	var ids []string
	if json.Unmarshal(body, &parsed) == nil {
		for _, c := range parsed.Conflicts {
			if c.PluginInstanceID != "" {
				ids = append(ids, c.PluginInstanceID)
			}
		}
	}
	if len(ids) > 0 {
		return fmt.Errorf("alias %q resolves to multiple plugin instances (%s); re-run delete with a specific plugin ID", alias, strings.Join(ids, ", "))
	}
	return fmt.Errorf("alias %q resolves to multiple plugin instances; re-run delete with a specific plugin ID: %s", alias, umsErrorMessage(body))
}

// localizedName returns a display name from a localized name map, preferring the
// English locale and falling back to the first locale in sorted order.
func localizedName(name map[string]string) string {
	for _, key := range []string{"en", "en-US"} {
		if v := strings.TrimSpace(name[key]); v != "" {
			return v
		}
	}
	keys := make([]string, 0, len(name))
	for k := range name {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if v := strings.TrimSpace(name[k]); v != "" {
			return v
		}
	}
	return ""
}

// renderPluginInstanceTable prints instances as an Alias/Id/Name/Created table,
// sorted by alias. Author-entered columns (alias, name) are truncated for layout;
// full values remain available via --json.
func renderPluginInstanceTable(w io.Writer, items []pluginInstance) {
	rows := make([][]string, 0, len(items))
	for _, p := range items {
		rows = append(rows, []string{
			truncateForTable(p.Alias, tableColumnMaxWidth),
			p.PluginInstanceID,
			truncateForTable(localizedName(p.Name), tableColumnMaxWidth),
			p.Created,
		})
	}
	output.WriteTable(w, []string{"Alias", "Id", "Name", "Created"}, rows, "Alias")
}

// truncateForTable shortens s to at most maxRunes runes, appending an ellipsis when
// truncated. It counts runes, not bytes, so multibyte author-entered text is never
// split mid-character.
func truncateForTable(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	return string(runes[:maxRunes-1]) + "…"
}

// renderDeleteConfirmation prints the details of the instance about to be deleted,
// warning when it has an active asset bundle (a live deployment).
func renderDeleteConfirmation(w io.Writer, inst *pluginInstance) {
	fmt.Fprintln(w, "You are about to delete the following UI plugin instance:")
	fmt.Fprintf(w, "  Alias:     %s\n", inst.Alias)
	fmt.Fprintf(w, "  Name:      %s\n", localizedName(inst.Name))
	fmt.Fprintf(w, "  Plugin ID: %s\n", inst.PluginInstanceID)
	if inst.Created != "" {
		fmt.Fprintf(w, "  Created:   %s\n", inst.Created)
	}
	if len(inst.Slots) > 0 {
		ids := make([]string, 0, len(inst.Slots))
		for _, s := range inst.Slots {
			ids = append(ids, s.SlotID)
		}
		fmt.Fprintf(w, "  Slots:     %s\n", strings.Join(ids, ", "))
	}
	if inst.ActiveAssetBundleID != nil && strings.TrimSpace(*inst.ActiveAssetBundleID) != "" {
		fmt.Fprintln(w, "  WARNING:   This plugin has an active asset bundle (a live deployment).")
	}
}

// promptYesNo asks question and returns true only for an explicit yes. An empty
// response (or EOF) defaults to No.
func promptYesNo(in io.Reader, w io.Writer, question string) (bool, error) {
	fmt.Fprintf(w, "%s [y/N]: ", question)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

// mapUMSListError translates a non-2xx list response into an actionable error.
func mapUMSListError(status int, body []byte) error {
	message := umsErrorMessage(body)
	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request listing plugin instances: %s", message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to list UI plugins (requires the idn:plugins-ui:read right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("the UI plugins feature is not enabled for this tenant, or the endpoint is unavailable: %s", message)
	default:
		return fmt.Errorf("failed to list plugin instances (status %d): %s", status, message)
	}
}

// mapUMSLookupError translates a non-2xx resolve/get response into an actionable error.
func mapUMSLookupError(status int, body []byte, target string) error {
	message := umsErrorMessage(body)
	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid plugin instance identifier %q: %s", target, message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to read UI plugins (requires the idn:plugins-ui:read right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("plugin instance %q not found (or the UI plugins feature is not enabled for this tenant): %s", target, message)
	default:
		return fmt.Errorf("failed to look up plugin instance %q (status %d): %s", target, status, message)
	}
}

// mapUMSDeleteError translates a non-2xx delete response into an actionable error.
func mapUMSDeleteError(status int, body []byte, target string) error {
	message := umsErrorMessage(body)
	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid plugin instance identifier %q: %s", target, message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to delete UI plugins (requires the idn:plugins-ui:delete right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("plugin instance %q not found: %s", target, message)
	default:
		return fmt.Errorf("failed to delete plugin instance %q (status %d): %s", target, status, message)
	}
}
