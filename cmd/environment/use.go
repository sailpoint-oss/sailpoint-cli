// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/spf13/cobra"
)

func newUseCommand() *cobra.Command {
	var env string
	return &cobra.Command{
		Use:     "use",
		Short:   "Set an environment as the active environment in the CLI",
		Long:    "\nSet an environment as the active environment in the CLI\n\n",
		Example: "sail env use environment_name | sail env u environment_name",
		Aliases: []string{"u"},
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			if len(args) > 0 {
				env = args[0]
			} else {
				env = config.GetActiveEnvironment()
			}

			if environments[env] != nil {
				config.SetActiveEnvironment(env)
				log.Info("Active Environment", "env", env)
			} else {
				log.Warn("Environment does not exist \nUse `sail env create " + env + "` to create it.")
			}

			return nil
		},
	}
}
