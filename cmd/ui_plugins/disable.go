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

//go:embed disable.md
var disableHelp string

func newDisableCommand() *cobra.Command {
	var force bool
	var jsonOutput bool

	help := util.ParseHelp(disableHelp)
	cmd := &cobra.Command{
		Use:     "disable (alias | plugin-id)",
		Short:   "Disable a UI plugin instance",
		Long:    help.Long,
		Example: help.Example,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spClient, err := newPluginClient()
			if err != nil {
				return err
			}
			return runDisable(context.Background(), spClient, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], force, jsonOutput)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "F", false, "Bypass confirmation prompts")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Print the disabled plugin instance as JSON")

	return cmd
}

// runDisable resolves the target (by alias or plugin ID), confirms disabling
// (unless force), then disables it. The confirmation is skipped by force, but the
// ambiguous-alias and not-found guards are always enforced. When jsonOutput is set,
// the confirmation is written to errOut so stdout carries only the resulting JSON.
func runDisable(ctx context.Context, c client.Client, in io.Reader, out io.Writer, errOut io.Writer, arg string, force bool, jsonOutput bool) error {
	inst, raw, err := resolvePluginTarget(ctx, c, arg)
	if err != nil {
		return err
	}

	if inst.State == stateDisabled {
		writer := out
		if jsonOutput {
			writer = errOut
		}

		_, _ = fmt.Fprintf(writer, "Plugin instance %s is already disabled\n", pluginInstanceLabel(inst))

		if jsonOutput {
			_, _ = fmt.Fprintln(out, strings.TrimSpace(string(raw)))
		}
		return nil
	}

	if !force {
		promptOut := out
		if jsonOutput {
			promptOut = errOut
		}
		renderConfirmation(promptOut, "disable", inst)

		confirmed, err := promptYesNo(in, promptOut, fmt.Sprintf("Disable plugin instance %s?", inst.PluginInstanceID))
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(promptOut, "Cancelled.")
			return nil
		}
	}

	response, err := setPluginInstanceState(ctx, c, inst, stateDisabled)
	if err != nil {
		return err
	}

	if jsonOutput {
		_, _ = fmt.Fprintln(out, strings.TrimSpace(string(response)))
		return nil
	}
	_, _ = fmt.Fprintf(out, "Disabled plugin instance %s\n", pluginInstanceLabel(inst))

	return nil
}
