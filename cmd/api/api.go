// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"github.com/spf13/cobra"
)

func NewAPICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api",
		Short:   "Make API requests to SailPoint endpoints",
		Long:    "\nMake API requests to SailPoint endpoints. Use this command to interact with SailPoint APIs directly.\n\n",
		Example: "sail api get /beta/accounts",
		Aliases: []string{"a"},
		Args:    cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newGetCmd(),
		newPostCmd(),
		newPutCmd(),
		newPatchCmd(),
		newDeleteCmd(),
	)

	return cmd
}
