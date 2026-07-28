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

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed push_manifest.md
var updateHelp string

func newUpdateCommand() *cobra.Command {
	var private bool
	var restrictToUsers []string
	var dryRun bool
	var jsonOutput bool

	help := util.ParseHelp(updateHelp)
	cmd := &cobra.Command{
		Use:     "push-manifest",
		Aliases: []string{"update"},
		Short:   "Push your local manifest configuration to the UI plugin instance in ISC",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := updateConfig{
				private:         private,
				restrictToUsers: restrictToUsers,
				dryRun:          dryRun,
				jsonOutput:      jsonOutput,
			}

			// A dry run still calls the read-only resolve-alias endpoint, so it
			// needs a client too. Tolerate a missing client on a dry run so the
			// payload preview still prints when config/auth is unavailable.
			spClient, clientErr := newPluginClient()
			if !dryRun && clientErr != nil {
				return clientErr
			}

			return runUpdate(context.Background(), spClient, manifestFileName, cmd.OutOrStdout(), cmd.ErrOrStderr(), config.GetCurrentIdentityID, opts)
		},
	}

	cmd.Flags().BoolVar(&private, "private", false, "Restrict the plugin to the current user on every slot")
	cmd.Flags().StringSliceVar(&restrictToUsers, "restrict-to-users", nil, "Restrict the plugin to the given user identity GUIDs on every slot (comma-separated)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate the manifest and print the payload that would be sent, without updating the plugin instance")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the raw UMS response on success")

	return cmd
}

// updateConfig captures the update command's flag values.
type updateConfig struct {
	private         bool
	restrictToUsers []string
	dryRun          bool
	jsonOutput      bool
}

// runUpdate loads and validates the workspace manifest at manifestPath, applies
// visibility overrides, then either previews the payload (dry run, with a
// best-effort check that the target instance exists) or resolves the workspace
// alias to its instance and PATCHes the manifest via the provided client. Only the
// manifest section is sent — static assets are handled by `upload`. currentUser
// resolves the GUID for --private and is injected for testability. The payload
// preview is written to out; advisory notes from the dry-run existence check are
// written to errOut.
func runUpdate(ctx context.Context, c client.Client, manifestPath string, out io.Writer, errOut io.Writer, currentUser func() (string, error), opts updateConfig) error {
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
		return checkInstanceExists(ctx, c, errOut, cfg.Manifest.Alias)
	}

	instance, _, err := resolvePluginInstanceByAlias(ctx, c, cfg.Manifest.Alias)
	if err != nil {
		return err
	}

	// Patch does not default a Content-Type, so the caller must set it or UMS
	// receives an empty body and silently applies no changes.
	headers := uiPluginRequestHeaders()
	headers["Content-Type"] = "application/json"

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(instance.PluginInstanceID)
	resp, err := c.Patch(ctx, url, bytes.NewReader(payload), headers)
	if err != nil {
		return fmt.Errorf("failed to update plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSUpdateError(resp.StatusCode, body, cfg.Manifest.Alias)
	}

	if opts.jsonOutput {
		_, _ = fmt.Fprintln(out, string(body))
		return nil
	}

	return renderUpdateSuccess(out, body, instance)
}

// checkInstanceExists performs a best-effort dry-run pre-check that the workspace
// alias resolves to an existing plugin instance in the tenant, using the read-only
// resolve-alias endpoint (GET). It returns a non-nil error only when the backend
// gives a definitive negative answer — the alias resolves to nothing (404) or to
// more than one instance (409) — so a dry run surfaces an update that would fail.
// Every other outcome (no client, transport error, or an inconclusive status such
// as 401/403 while the route is not yet externally accessible) is reported to
// errOut and treated as non-fatal, leaving the printed payload as the dry-run
// result. This mirrors create's checkAliasAvailability with inverted polarity:
// create wants the alias to be free, update wants it to already exist.
func checkInstanceExists(ctx context.Context, c client.Client, errOut io.Writer, alias string) error {
	if c == nil {
		_, _ = fmt.Fprintln(errOut, "Skipped instance existence check: no authenticated client available.")
		return nil
	}

	url := resolveAliasEndpoint + "?alias=" + neturl.QueryEscape(alias)
	resp, err := c.Get(ctx, url, uiPluginRequestHeaders())
	if err != nil {
		_, _ = fmt.Fprintf(errOut, "Could not verify the plugin instance exists: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		return nil
	case resp.StatusCode == http.StatusConflict:
		return ambiguousAliasError(alias, body)
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("no plugin instance found for alias %q; run `sail ui-plugins create` first: %s", alias, umsErrorMessage(body))
	default:
		_, _ = fmt.Fprintf(errOut, "Could not verify the plugin instance exists (status %d): %s\n", resp.StatusCode, umsErrorMessage(body))
		return nil
	}
}

// mapUMSUpdateError translates a non-2xx UMS update (PATCH) response into an
// actionable error, surfacing the message returned by the backend so authors can
// correct manifest issues without guessing.
func mapUMSUpdateError(status int, body []byte, alias string) error {
	message := umsErrorMessage(body)

	switch status {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request updating plugin instance for alias %q: %s", alias, message)
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to update UI plugins (requires the idn:plugins-ui:update right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("plugin instance for alias %q not found, or the UI plugins feature is not enabled for this tenant: %s", alias, message)
	case http.StatusConflict:
		return fmt.Errorf("alias conflict updating plugin instance for alias %q: %s", alias, message)
	default:
		return fmt.Errorf("failed to update plugin instance for alias %q (status %d): %s", alias, status, message)
	}
}

// renderUpdateSuccess prints a human-readable confirmation of the updated instance,
// preferring the identifiers returned by the PATCH response and falling back to the
// instance resolved from the workspace alias.
func renderUpdateSuccess(w io.Writer, body []byte, resolved *pluginInstance) error {
	var updated struct {
		PluginInstanceID string `json:"pluginInstanceId"`
		Alias            string `json:"alias"`
	}
	_ = json.Unmarshal(body, &updated)

	id := updated.PluginInstanceID
	if id == "" {
		id = resolved.PluginInstanceID
	}
	alias := updated.Alias
	if alias == "" {
		alias = resolved.Alias
	}

	if id != "" {
		_, _ = fmt.Fprintf(w, "Updated plugin instance %s (alias: %s)\n", id, alias)
	} else {
		_, _ = fmt.Fprintf(w, "Updated plugin instance (alias: %s)\n", alias)
	}

	return nil
}
