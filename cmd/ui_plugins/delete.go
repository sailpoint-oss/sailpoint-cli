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

//go:embed delete.md
var deleteHelp string

func newDeleteCommand() *cobra.Command {
	var force bool
	var jsonOutput bool

	help := util.ParseHelp(deleteHelp)
	cmd := &cobra.Command{
		Use:     "delete (alias | plugin-id)",
		Short:   "Delete a UI plugin instance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runDelete(context.Background(), spClient, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], force, jsonOutput)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "F", false, "Bypass confirmation prompts")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the deleted plugin instance as JSON")

	return cmd
}

// runDelete resolves the target (by alias or plugin ID), confirms the deletion
// (unless force), then deletes it. The confirmation is skipped by force, but the
// ambiguous-alias and not-found guards are always enforced. When jsonOutput is set,
// the confirmation is written to errOut so stdout carries only the resulting JSON.
func runDelete(ctx context.Context, c client.Client, in io.Reader, out io.Writer, errOut io.Writer, arg string, force bool, jsonOutput bool) error {
	inst, raw, err := resolvePluginTarget(ctx, c, arg)
	if err != nil {
		return err
	}

	if !force {
		promptOut := out
		if jsonOutput {
			promptOut = errOut
		}
		renderConfirmation(promptOut, "delete", inst)

		confirmed, err := promptYesNo(in, promptOut, fmt.Sprintf("Delete plugin instance %s?", inst.PluginInstanceID))
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(promptOut, "Cancelled.")
			return nil
		}
	}

	url := pluginInstancesEndpoint + "/" + neturl.PathEscape(inst.PluginInstanceID)
	resp, err := c.Delete(ctx, url, nil, uiPluginRequestHeaders())
	if err != nil {
		return fmt.Errorf("failed to delete plugin instance: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mapUMSDeleteError(resp.StatusCode, body, inst.PluginInstanceID)
	}

	if jsonOutput {
		_, _ = fmt.Fprintln(out, strings.TrimSpace(string(raw)))
		return nil
	}

	if inst.Alias != "" {
		_, _ = fmt.Fprintf(out, "Deleted plugin instance %s (alias: %s)\n", inst.PluginInstanceID, inst.Alias)
	} else {
		_, _ = fmt.Fprintf(out, "Deleted plugin instance %s\n", inst.PluginInstanceID)
	}
	return nil
}
