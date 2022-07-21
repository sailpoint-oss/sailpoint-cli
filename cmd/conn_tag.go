// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnTagCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage tags",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), cmd.UsageString())
		},
	}

	cmd.PersistentFlags().StringP("id", "c", "", "Connector ID or Alias")
	_ = cmd.MarkPersistentFlagRequired("id")

	cmd.AddCommand(
		newConnTagListCmd(client),
		newConnTagCreateCmd(client),
		newConnTagUpdateCmd(client),
	)

	return cmd
}
