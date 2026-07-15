package ui_plugins

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed upload.md
var uploadHelp string

func newUploadCommand() *cobra.Command {
	var outDir string

	help := util.ParseHelp(uploadHelp)

	cmd := &cobra.Command{
		Use:     "upload",
		Short:   "Upload the already compiled UI plugin assets to the plugin instance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runUpload(context.Background(), spClient, manifestFileName, outDir, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&outDir, "out-dir", "", "Custom out directory to use instead of sp-ui-plugin.json's build.outDir value")

	return cmd
}

// runUpload loads the workspace manifest, resolves the build output directory,
// collects the compiled assets, resolves the plugin instance from the workspace
// alias, and uploads the assets as a new asset bundle. The client and output
// writer are injected so the flow is testable without a live tenant.
func runUpload(ctx context.Context, c client.Client, manifestPath string, flagOutDir string, out io.Writer) error {
	cfg, err := loadAndValidateWorkspaceManifest(manifestPath)
	if err != nil {
		return err
	}

	outDir, err := resolveUploadOutDir(flagOutDir, cfg)
	if err != nil {
		return err
	}

	files, err := collectUploadFiles(outDir)
	if err != nil {
		return err
	}

	instance, _, err := resolvePluginInstanceByAlias(ctx, c, cfg.Manifest.Alias)
	if err != nil {
		return err
	}

	body, contentType, err := buildAssetBundleBody(files)
	if err != nil {
		return err
	}

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(instance.PluginInstanceID) + "/asset-bundles"
	resp, err := c.Post(ctx, url, contentType, body, uiPluginRequestHeaders())
	if err != nil {
		return fmt.Errorf("failed to upload plugin assets: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSUploadError(resp.StatusCode, respBody, cfg.Manifest.Alias)
	}

	return renderUploadSuccess(out, respBody, cfg.Manifest.Alias)
}

// resolveUploadOutDir determines which build output directory to upload. An
// explicit --out-dir flag takes precedence; otherwise the build.outDir value
// from the already-loaded workspace manifest is used.
func resolveUploadOutDir(flagOutDir string, cfg *uiPluginWorkspaceConfig) (string, error) {
	if trimmed := strings.TrimSpace(flagOutDir); trimmed != "" {
		return trimmed, nil
	}

	if cfg.Build == nil {
		return "", fmt.Errorf("no output directory to upload: pass --out-dir or set build.outDir in %s", manifestFileName)
	}

	outDir := strings.TrimSpace(cfg.Build.OutDir)

	if outDir == "" {
		return "", fmt.Errorf("no output directory to upload: pass --out-dir or set build.outDir in %s", manifestFileName)
	}

	return outDir, nil
}
