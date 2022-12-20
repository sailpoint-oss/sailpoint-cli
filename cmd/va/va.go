// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package va

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/client"
	"github.com/spf13/cobra"
)

func NewVACmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "va",
		Short:   "Virtual Appliance commands",
		Aliases: []string{"va"},
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newCollectCmd(client),
		newTroubleshootCmd(client),
		newUpdateCmd(client),
		newParseCmd(client),
	)

	return cmd
}
