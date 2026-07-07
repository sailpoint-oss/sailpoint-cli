package ui_plugins

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed create.md
var createHelp string

const (
	pluginInstancesEndpoint = "/v2026/ui-plugins"
	validateAliasEndpoint   = pluginInstancesEndpoint + "/validate-alias"

	// experimentalHeader must be sent on ui-plugins requests while the routes are
	// in a preview API lifecycle state (limited-preview / public-preview).
	experimentalHeader = "X-SailPoint-Experimental"
)

// uiPluginRequestHeaders returns the headers required on every external
// ui-plugins request: JSON accept plus the experimental opt-in header the
// preview-state routes require.
func uiPluginRequestHeaders() map[string]string {
	return map[string]string{
		"Accept":           "application/json",
		experimentalHeader: "true",
	}
}

func newCreateCommand() *cobra.Command {
	var private bool
	var restrictToUsers []string
	var dryRun bool
	var jsonOutput bool

	help := util.ParseHelp(createHelp)
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a UI plugin instance in the current tenant",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := createConfig{
				private:         private,
				restrictToUsers: restrictToUsers,
				dryRun:          dryRun,
				jsonOutput:      jsonOutput,
			}

			// A dry run still calls the read-only validate-alias endpoint, so it
			// needs a client too. Tolerate a missing client on a dry run so the
			// payload preview still prints when config/auth is unavailable.
			spClient, clientErr := newPluginClient()
			if !dryRun && clientErr != nil {
				return clientErr
			}

			return runCreate(context.Background(), spClient, manifestFileName, cmd.OutOrStdout(), cmd.ErrOrStderr(), config.GetCurrentIdentityID, opts)
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Restrict the plugin to the current user on every slot")
	cmd.Flags().StringSliceVar(&restrictToUsers, "restrict-to-users", nil, "Restrict the plugin to the given user identity GUIDs on every slot (comma-separated)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate the manifest and print the payload that would be sent, without creating the plugin instance")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the raw UMS response on success")

	return cmd
}

// createConfig captures the create command's flag values.
type createConfig struct {
	private         bool
	restrictToUsers []string
	dryRun          bool
	jsonOutput      bool
}

// newPluginClient builds an authenticated client from the active CLI config.
func newPluginClient() (client.Client, error) {
	if err := config.InitConfig(); err != nil {
		return nil, err
	}
	cliConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return client.NewSpClient(cliConfig), nil
}

// runCreate loads and validates the workspace manifest at manifestPath, applies
// visibility overrides, then either previews the payload (dry run, with a
// best-effort alias availability check) or POSTs it via the provided client and
// renders the result. currentUser resolves the GUID for --private and is injected
// for testability. The payload preview is written to out; advisory notes from the
// dry-run alias check are written to errOut.
func runCreate(ctx context.Context, c client.Client, manifestPath string, out io.Writer, errOut io.Writer, currentUser func() (string, error), opts createConfig) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	users, err := resolveRestrictToUsers(opts.private, opts.restrictToUsers, currentUser)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		applyVisibilityOverride(&cfg.Manifest, users)
	}

	payload, err := json.MarshalIndent(cfg.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode plugin manifest: %w", err)
	}

	if opts.dryRun {
		_, _ = fmt.Fprintln(out, string(payload))
		return checkAliasAvailability(ctx, c, errOut, cfg.Manifest.Alias)
	}

	resp, err := c.Post(ctx, pluginInstancesEndpoint, "application/json", bytes.NewReader(payload), uiPluginRequestHeaders())
	if err != nil {
		return fmt.Errorf("failed to create plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSCreateError(resp.StatusCode, body, cfg.Manifest.Alias)
	}

	if opts.jsonOutput {
		_, _ = fmt.Fprintln(out, string(body))
		return nil
	}

	return renderCreateSuccess(out, body, cfg.Manifest.Alias)
}

// checkAliasAvailability performs a best-effort dry-run pre-check against the UMS
// validate-alias endpoint (GET, read-only). It returns a non-nil error only when the
// backend gives a definitive negative answer — the alias is already taken (409) or
// invalid (400) — so a dry run surfaces a create that would fail. Every other outcome
// (no client, transport error, or an inconclusive status such as 401/403/404 while the
// route is not yet externally accessible) is reported to errOut and treated as
// non-fatal, leaving the printed payload as the dry-run result.
func checkAliasAvailability(ctx context.Context, c client.Client, errOut io.Writer, alias string) error {
	if c == nil {
		_, _ = fmt.Fprintln(errOut, "Skipped alias availability check: no authenticated client available.")
		return nil
	}

	url := validateAliasEndpoint + "?alias=" + neturl.QueryEscape(alias)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "Could not verify alias availability: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		return nil
	case resp.StatusCode == http.StatusConflict:
		return fmt.Errorf("alias %q is already in use in this tenant; create would be rejected: %s", alias, umsErrorMessage(body))
	case resp.StatusCode == http.StatusBadRequest:
		return fmt.Errorf("alias %q is invalid: %s", alias, umsErrorMessage(body))
	default:
		_, _ = fmt.Fprintf(errOut, "Could not verify alias availability (status %d): %s\n", resp.StatusCode, umsErrorMessage(body))
		return nil
	}
}

// resolveRestrictToUsers builds the de-duplicated, order-preserving union of the
// explicitly provided user GUIDs and, when private is set, the current user's
// identity GUID. currentUser is injected so the resolution path is testable.
func resolveRestrictToUsers(private bool, flagUsers []string, currentUser func() (string, error)) ([]string, error) {
	seen := make(map[string]struct{})
	var users []string

	add := func(user string) {
		user = strings.TrimSpace(user)
		if user == "" {
			return
		}
		if _, ok := seen[user]; ok {
			return
		}
		seen[user] = struct{}{}
		users = append(users, user)
	}

	for _, user := range flagUsers {
		add(user)
	}

	if private {
		id, err := currentUser()
		if err != nil {
			return nil, err
		}
		add(id)
	}

	return users, nil
}

// applyVisibilityOverride sets restrictToUsers on every slot in the manifest,
// giving each slot its own copy of the user list.
func applyVisibilityOverride(manifest *uiPluginManifest, users []string) {
	for i := range manifest.Slots {
		slotUsers := make([]string, len(users))
		copy(slotUsers, users)
		manifest.Slots[i].RestrictToUsers = slotUsers
	}
}

// mapUMSCreateError translates a non-2xx UMS response into an actionable error,
// surfacing the message returned by the backend.
func mapUMSCreateError(status int, body []byte, alias string) error {
	message := umsErrorMessage(body)

	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request creating plugin instance: %s", message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to create UI plugins (requires the idn:plugins-ui:create right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("the UI plugins feature is not enabled for this tenant, or the endpoint is unavailable: %s", message)
	case http.StatusConflict:
		return fmt.Errorf("a plugin instance with alias %q already exists for this tenant: %s", alias, message)
	default:
		return fmt.Errorf("failed to create plugin instance (status %d): %s", status, message)
	}
}

// umsErrorMessage extracts a human-readable message from a UMS error body,
// which follows the NestJS shape {"statusCode":N,"message":string|[]string,"error":string}.
// It falls back to the raw body when the shape is unrecognized.
func umsErrorMessage(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "no response body"
	}

	var parsed struct {
		Message json.RawMessage `json:"message"`
		Error   string          `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if len(parsed.Message) > 0 {
			var single string
			if json.Unmarshal(parsed.Message, &single) == nil && strings.TrimSpace(single) != "" {
				return single
			}
			var many []string
			if json.Unmarshal(parsed.Message, &many) == nil && len(many) > 0 {
				return strings.Join(many, "; ")
			}
		}
		if parsed.Error != "" {
			return parsed.Error
		}
	}

	return trimmed
}

// renderCreateSuccess prints a human-readable confirmation of the created instance.
func renderCreateSuccess(w io.Writer, body []byte, fallbackAlias string) error {
	var created struct {
		PluginInstanceID string `json:"pluginInstanceId"`
		Alias            string `json:"alias"`
	}
	_ = json.Unmarshal(body, &created)

	alias := created.Alias
	if alias == "" {
		alias = fallbackAlias
	}

	if created.PluginInstanceID != "" {
		_, _ = fmt.Fprintf(w, "Created plugin instance %s (alias: %s)\n", created.PluginInstanceID, alias)
	} else {
		_, _ = fmt.Fprintf(w, "Created plugin instance (alias: %s)\n", alias)
	}

	return nil
}
