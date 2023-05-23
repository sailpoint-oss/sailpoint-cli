// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package logConfig

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Short:   "Interact with a SailPoint Virtual Appliances log configuration",
		Long:    "\nInteract with SailPoint Virtual Appliances log configuration\n\n",
		Aliases: []string{"l"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newGetCmd(),
		newSetCmd(),
	)

	return cmd
}
