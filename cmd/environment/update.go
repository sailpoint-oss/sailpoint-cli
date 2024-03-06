// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "update",
		Short:   "Update an existing environment in the CLI",
		Long:    "\nUpdate an existing environment in the CLI\n\n",
		Example: "sail env update | sail env update environment_name | sail env u environment_name",
		Aliases: []string{"u"},
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			for _, environmentName := range args {
				if environments[environmentName] != nil {
					err := util.CreateOrUpdateEnvironment(environmentName, true)
					if err != nil {
						return err
					}
				} else {
					log.Warn("Environment does not exist to update", "env", environmentName)
					return nil
				}

			}

			if len(args) == 0 {
				env := config.GetActiveEnvironment()

				if env != "" && env != " " {
					log.Warn("You are about to Update the active Environment", "env", env)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {

						err := util.CreateOrUpdateEnvironment(env, true)
						if err != nil {
							return err
						}
					}
				} else {
					log.Warn("No environments configured")
					return nil
				}
			}

			return nil
		},
	}
}
