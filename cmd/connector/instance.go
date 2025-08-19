// Copyright (c) 2023, SailPoint Technologies, Inc. All rights reserved.
package connector

import (
	"fmt"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/spf13/cobra"
)

func newConnInstancesCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage connector instances",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.AddCommand(
		newInstanceListCmd(client),
	)

	return cmd
}
