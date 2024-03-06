// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/terminal"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "show",
		Short:   "Show the current active environment in the CLI",
		Long:    "\nShow the current active environment in the CLI\n\n",
		Example: "sail env show | sail env s",
		Aliases: []string{"s"},
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			for _, environmentName := range args {
				if environments[environmentName] != nil {
					log.Warn("You are about to Print out the Environment", "env", environmentName)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {
						fmt.Println(util.PrettyPrint(environments[environmentName]))
					}
				} else {
					log.Warn("Environment does not exist", "env", environmentName)
					return nil
				}

			}

			if len(args) == 0 {
				env := config.GetActiveEnvironment()

				if env != "" && env != " " {
					log.Warn("You are about to Print out the Environment", "env", env)
					res := terminal.InputPrompt("Press Enter to continue")
					if res == "" {
						fmt.Println(util.PrettyPrint(environments[env]))
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
