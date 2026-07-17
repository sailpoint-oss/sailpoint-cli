package ui_plugins

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed unlink.md
var unlinkHelp string

func newUnlinkCommand() *cobra.Command {
	help := util.ParseHelp(unlinkHelp)
	cmd := &cobra.Command{
		Use:     "unlink",
		Short:   "Remove your local development link for a UI plugin",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runUnlink(context.Background(), spClient, manifestFileName, cmd.OutOrStdout())
		},
	}

	return cmd
}

// runUnlink removes the authenticated developer's local dev server override from
// the plugin instance resolved from the workspace alias. The removal is
// idempotent: UMS returns success whether or not a mapping existed, so this
// prints a confirmation in both cases. The client and output writer are injected
// so the flow is testable without a live tenant.
func runUnlink(ctx context.Context, c client.Client, manifestPath string, out io.Writer) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	instance, _, err := resolvePluginInstanceByAlias(ctx, c, cfg.Manifest.Alias)
	if err != nil {
		return err
	}

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(instance.PluginInstanceID) + "/link"
	resp, err := c.Delete(ctx, url, nil, uiPluginRequestHeaders())
	if err != nil {
		return fmt.Errorf("failed to remove link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// The response body is read only to explain a failure.
		respBody, _ := io.ReadAll(resp.Body)
		return mapUMSUnlinkError(resp.StatusCode, respBody, cfg.Manifest.Alias)
	}

	fmt.Fprintf(out, "Removed the local dev link for plugin %q\n", instance.Alias)

	return nil
}

// mapUMSUnlinkError translates a non-2xx unlink response into an actionable error.
func mapUMSUnlinkError(status int, body []byte, alias string) error {
	message := umsErrorMessage(body)
	switch status {
	case http.StatusForbidden:
		return fmt.Errorf("not authorized to unlink UI plugins (requires the idn:plugins-ui:update right): %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("plugin instance for alias %q not found (or the UI plugins feature is not enabled for this tenant): %s", alias, message)
	default:
		return fmt.Errorf("failed to remove link for plugin %q (status %d): %s", alias, status, message)
	}
}
