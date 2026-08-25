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

//go:embed link.md
var linkHelp string

type linkRequest struct {
	Port int `json:"port"`
}

func newLinkCommand() *cobra.Command {
	var port int

	help := util.ParseHelp(linkHelp)
	cmd := &cobra.Command{
		Use:     "link",
		Short:   "Link local development URL for a UI plugin",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return link(context.Background(), spClient, manifestFileName, port, cmd.Flags().Changed("port"), cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
	cmd.Flags().IntVar(&port, "port", 0, "Custom port to use instead of sp-ui-plugin.json's build.port value")

	return cmd
}

func link(ctx context.Context, c client.Client, manifestPath string, flagPort int, portSet bool, out io.Writer, errOut io.Writer) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	tenantUrl := config.GetTenantUrl()
	if tenantUrl == "" {
		return fmt.Errorf("no tenant URL configured; run `sail env` to set one before linking")
	}

	resolvedPort, defaulted, err := resolveLinkPort(flagPort, portSet, cfg)
	if err != nil {
		return err
	}

	if defaulted {
		fmt.Fprintf(errOut, "no port provided via --port or build.port in %s; defaulting to port %d\n", manifestFileName, defaultDevServerPort)
	}

	portData := linkRequest{Port: resolvedPort}
	jsonPortData, err := json.Marshal(portData)
	if err != nil {
		return fmt.Errorf("failed to encode port data: %w", err)
	}

	instance, _, err := resolvePluginInstanceByAlias(ctx, c, cfg.Manifest.Alias)
	if err != nil {
		return err
	}

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(instance.PluginInstanceID) + "/link"
	resp, err := c.Post(ctx, url, "application/json", bytes.NewBuffer(jsonPortData), uiPluginRequestHeaders())
	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		// The response body is read only to explain a failure. On success it
		// carries devDocumentHeaders for angular.json patching.
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unable to create link: %s", umsErrorMessage(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read link response: %w", err)
	}

	headers, _ := parseDevDocumentHeaders(respBody)
	applyDevDocumentHeadersBestEffort(manifestPath, cfg.Manifest.Alias, headers, errOut)

	viewURL := pluginViewURL(tenantUrl, instance.PluginInstanceID)
	if viewURL == "" {
		return fmt.Errorf("could not construct developer URL for alias %q", instance.Alias)
	}

	fmt.Fprintf(errOut, "Plugin %s linked to port %d\nTo load your local plugin in ISC navigate to:\n", instance.Alias, resolvedPort)
	fmt.Fprintf(out, "%s?spPluginDev=%s\n", viewURL, neturl.QueryEscape(instance.Alias))
	fmt.Fprintf(errOut, "\nNext: start your local dev server on port %d, then open the URL above in ISC.\n(Angular template: `npm start`. Other setups: see the plugin guide in your workspace.)\n", resolvedPort)

	return nil
}

func resolveLinkPort(flagPort int, portSet bool, cfg *uiPluginWorkspaceConfig) (port int, defaulted bool, err error) {
	if portSet {
		if err := validatePortNumber(flagPort); err != nil {
			return 0, false, fmt.Errorf("provided port is invalid: %w", err)
		}

		return flagPort, false, nil
	}

	if cfg.Build == nil || cfg.Build.Port == nil {
		return defaultDevServerPort, true, nil
	}

	if err := validatePortNumber(*cfg.Build.Port); err != nil {
		return 0, false, fmt.Errorf("build.port in %s is invalid: %w", manifestFileName, err)
	}

	return *cfg.Build.Port, false, nil
}
