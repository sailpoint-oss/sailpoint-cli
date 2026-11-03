package ui_plugins

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed ui_plugins.md
var uiPluginsHelp string

func NewUIPluginsCommand() *cobra.Command {
	help := util.ParseHelp(uiPluginsHelp)
	cmd := &cobra.Command{
		Use:     "ui-plugins",
		Aliases: []string{"ui-plugin"},
		Short:   "Manage UI plugin workflows in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newInitCommand(),
		newCreateCommand(),
		newLinkCommand(),
		newUnlinkCommand(),
		newUpdateCommand(),
		newUploadCommand(),
		newListCommand(),
		newDeleteCommand(),
		newDisableCommand(),
		newEnableCommand(),
		newValidateManifestCommand(),
	)

	return cmd
}
