package ui_plugins

import (
	"context"
	_ "embed"
	"fmt"
	"io"

	"strings"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed enable.md
var enableHelp string

func newEnableCommand() *cobra.Command {
	var jsonOutput bool

	help := util.ParseHelp(enableHelp)
	cmd := &cobra.Command{
		Use:     "enable (alias | plugin-id)",
		Short:   "Enable a UI plugin instance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runEnable(context.Background(), spClient, cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the enabled plugin instance as JSON")

	return cmd
}

// runEnable resolves the target (by alias or plugin ID), then enables
// it.
func runEnable(ctx context.Context, c client.Client, out io.Writer, errOut io.Writer, arg string, jsonOutput bool) error {
	inst, raw, err := resolvePluginTarget(ctx, c, arg)
	if err != nil {
		return err
	}

	if inst.State == stateEnabled {
		writer := out
		if jsonOutput {
			writer = errOut
		}
		_, _ = fmt.Fprintf(writer, "Plugin instance %s is already enabled\n", pluginInstanceLabel(inst))

		if jsonOutput {
			_, _ = fmt.Fprintln(out, strings.TrimSpace(string(raw)))
		}
		return nil
	}

	response, err := setPluginInstanceState(ctx, c, inst, stateEnabled)
	if err != nil {
		return err
	}

	if jsonOutput {
		_, _ = fmt.Fprintln(out, strings.TrimSpace(string(response)))
		return nil
	}

	_, _ = fmt.Fprintf(out, "Enabled plugin instance %s\n", pluginInstanceLabel(inst))

	return nil
}
