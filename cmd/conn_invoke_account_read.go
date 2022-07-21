// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.
package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountReadCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-read [id/lookupId] [uniqueId]",
		Short: "Invoke a std:account:read command",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cc, err := connClient(cmd, client)
			if err != nil {
				return err
			}

			uniqueID := ""
			if len(args) > 1 {
				uniqueID = args[1]
			}

			_, rawResponse, err := cc.AccountRead(ctx, args[0], uniqueID)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))
			return nil
		},
	}

	return cmd
}
