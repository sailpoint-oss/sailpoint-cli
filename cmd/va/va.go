// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/cmd/va/logConfig"
	"github.com/spf13/cobra"
)

func NewVACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "va",
		Short:   "Interact with SailPoint Virtual Appliances",
		Long:    "\nInteract with SailPoint Virtual Appliances\n\n",
		Aliases: []string{"va"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCollectCmd(),
		// newTroubleshootCmd(),
		newListCmd(),
		newParseCmd(),
		newUpdateCmd(),
		logConfig.NewLogCmd(),
	)

	return cmd
}
