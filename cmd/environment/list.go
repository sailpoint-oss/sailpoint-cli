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

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all configured environments in the CLI",
		Long:    "\nList all configured environments in the CLI\n\n",
		Example: "sail env list | sail env ls",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			environments := config.GetEnvironments()

			if len(environments) != 0 {
				log.Warn("You are about to Print out the list of Environments")
				res := terminal.InputPrompt("Press Enter to continue")
				log.Info("Response", "res", res)
				if res == "" {
					fmt.Println(util.PrettyPrint(environments))
				}
			} else {
				log.Warn("No environments configured")
				return nil
			}
			return nil
		},
	}
}
