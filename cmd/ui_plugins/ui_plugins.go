// Copyright (c) 2026, SailPoint Technologies, Inc. All rights reserved.
package ui_plugins

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

const experimentalUIPluginsEnvVar = "SAIL_EXPERIMENTAL_UI_PLUGINS"

//go:embed ui_plugins.md
var uiPluginsHelp string

func NewUIPluginsCommand() *cobra.Command {
	help := util.ParseHelp(uiPluginsHelp)
	cmd := &cobra.Command{
		Use:     "ui-plugins",
		Short:   "Manage UI plugin workflows in Identity Security Cloud",
		Long:    help.Long,
		Example: help.Example,
		Hidden:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !isUIPluginsEnabled() {
				return experimentalDisabledError()
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newInitCommand(),
		newCreateCommand(),
		newLinkCommand(),
		newUpdateCommand(),
		newUploadCommand(),
		newListCommand(),
		newDeleteCommand(),
		newValidateManifestCommand(),
	)

	return cmd
}

func isUIPluginsEnabled() bool {
	return os.Getenv(experimentalUIPluginsEnvVar) == "1"
}

func experimentalDisabledError() error {
	return fmt.Errorf(
		"the `sail ui-plugins` command group is experimental and currently disabled. Enable it with `%s=1`",
		experimentalUIPluginsEnvVar,
	)
}
