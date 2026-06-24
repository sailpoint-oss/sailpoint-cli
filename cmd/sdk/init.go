// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize SDK projects",
		Long:    "\nInitialize a new SailPoint SDK project.\n\nChoose a language subcommand to scaffold a project with\nthe necessary dependencies and configuration.\n",
		Example: "  sail sdk init golang\n  sail sdk init typescript my-project\n  sail sdk init config",
		Aliases: []string{"i"},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newTypescriptCommand(),
		newGolangCommand(),
		newPowerShellCommand(),
		newPythonCommand(),
		newConfigCommand(),
	)

	return cmd
}
