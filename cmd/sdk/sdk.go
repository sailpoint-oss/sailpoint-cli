// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package sdk

import (
	"github.com/spf13/cobra"
)

func NewSDKCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sdk",
		Short:   "Initialize or configure SDK projects",
		Long:    "\nInitialize or configure SDK projects\n\n",
		Example: "sail sdk",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newInitCommand(),
	)

	return cmd

}
