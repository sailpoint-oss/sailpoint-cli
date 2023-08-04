// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package logConfig

import (
	"github.com/spf13/cobra"
)

func NewLogCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "Interact with a SailPoint Virtual Appliances log configuration",
		Long:    "\nInteract with SailPoint Virtual Appliances log configuration\n\n",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newGetCommand(),
		newSetCommand(),
	)

	return cmd
}
