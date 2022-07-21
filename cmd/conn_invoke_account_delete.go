// Copyright (c) 2021, SailPoint Technologies, Inc. All rights reserved.

package cmd

import (
	"fmt"

	"github.com/sailpoint/sp-cli/client"
	"github.com/spf13/cobra"
)

func newConnInvokeAccountDeleteCmd(client client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-delete <identity>",
		Short:   "Invoke a std:account:delete command",
		Example: `sp connectors invoke account-delete john.doe`,
		Args:    cobra.RangeArgs(1, 2),
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

			rawResponse, err := cc.AccountDelete(ctx, args[0], uniqueID)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(rawResponse))
			return nil
		},
	}

	return cmd
}
