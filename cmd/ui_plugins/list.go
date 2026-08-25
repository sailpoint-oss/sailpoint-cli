package ui_plugins

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed list.md
var listHelp string

func newListCommand() *cobra.Command {
	var jsonOutput bool

	help := util.ParseHelp(listHelp)
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List UI plugin instances in the current tenant",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runList(context.Background(), spClient, cmd.OutOrStdout(), jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the raw plugin instance list as JSON")

	return cmd
}

// runList fetches all plugin instances for the active tenant and renders them as a
// table, or as the raw aggregated JSON array when jsonOutput is set.
func runList(ctx context.Context, c client.Client, out io.Writer, jsonOutput bool) error {
	raw, err := listPluginInstances(ctx, c)
	if err != nil {
		return err
	}

	if jsonOutput {
		encoded, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode plugin instances: %w", err)
		}
		_, _ = fmt.Fprintln(out, string(encoded))
		return nil
	}

	instances := make([]pluginInstance, 0, len(raw))
	for _, item := range raw {
		var inst pluginInstance
		if err := json.Unmarshal(item, &inst); err != nil {
			return fmt.Errorf("failed to parse plugin instance: %w", err)
		}
		instances = append(instances, inst)
	}

	if len(instances) == 0 {
		_, _ = fmt.Fprintln(out, "No plugin instances found.")
		return nil
	}

	renderPluginInstanceTable(out, instances)
	return nil
}
