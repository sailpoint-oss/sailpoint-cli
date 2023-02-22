// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "va",
		Short:   "collection of commands to interact with sailpoint virtual appliance",
		Aliases: []string{"va"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCollectCmd(),
		// newTroubleshootCmd(),
		newUpdateCmd(),
		newParseCmd(),
	)

	return cmd
}
