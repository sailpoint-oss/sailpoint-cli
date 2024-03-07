// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	_ "embed"

	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed environment.md
var environmentHelp string

func NewEnvironmentCommand() *cobra.Command {
	help := util.ParseHelp(environmentHelp)
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Environments for the CLI",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"env"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newListCommand(),
		newShowCommand(),
		newDeleteCommand(),
		newCreateCommand(),
		newUpdateCommand(),
		newUseCommand(),
	)

	return cmd
}
