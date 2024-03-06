// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package environment

import (
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "create",
		Short:   "Create a new environment in the CLI",
		Long:    "\nCreate a new environment in the CLI\n\n",
		Example: "sail env create | sail env create environment_name | sail env c environment_name",
		Aliases: []string{"c"},
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, environmentName := range args {
				err := util.CreateOrUpdateEnvironment(environmentName, false)
				if err != nil {
					return err
				}
			}

			if len(args) == 0 {
				err := util.CreateOrUpdateEnvironment("", false)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}
