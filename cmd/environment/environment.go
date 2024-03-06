// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	_ "embed"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

//go:embed environment.md
var environmentHelp string

func NewEnvironmentCommand() *cobra.Command {
	help := util.ParseHelp(environmentHelp)
	var env string
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Environments for the CLI",
		Long:    help.Long,
		Example: help.Example,
		Aliases: []string{"env"},
		Run: func(cmd *cobra.Command, args []string) {
			environments := config.GetEnvironments()

			if len(args) > 0 {
				env = args[0]
			} else {
				env = config.GetActiveEnvironment()
			}

			if environments[env] != nil {
				config.SetActiveEnvironment(env)
				log.Info("Active Environment", "env", env)
			} else if env != "help" {
				log.Warn("Environment does not exist", "env", env)
			} else {
				cmd.Help()
			}

		},
	}

	cmd.AddCommand(
		newListCommand(),
		newShowCommand(),
		newDeleteCommand(),
		newCreateCommand(),
		newUpdateCommand(),
	)

	return cmd
}
